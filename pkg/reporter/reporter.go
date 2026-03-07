package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gaskaj/OpenAgentFramework/pkg/apitypes"
)

// Config holds configuration for the control plane reporter.
type Config struct {
	ControlPlaneURL string
	APIKey          string
	AgentName       string
	AgentType       string
	Version         string
	Hostname        string
	GitHubOwner     string
	GitHubRepo      string
	Tags            []string
	APIVersion      string        // API version to use for communication
	BufferSize      int
	FlushInterval   time.Duration
	Timeout         time.Duration
}

func (c *Config) defaults() {
	if c.BufferSize <= 0 {
		c.BufferSize = 100
	}
	if c.FlushInterval <= 0 {
		c.FlushInterval = 5 * time.Second
	}
	if c.Timeout <= 0 {
		c.Timeout = 10 * time.Second
	}
	if c.Hostname == "" {
		c.Hostname, _ = os.Hostname()
	}
	if c.APIVersion == "" {
		c.APIVersion = "v1"
	}
}

// Reporter sends agent events to the control plane API.
// It buffers events internally and flushes them periodically.
// All methods are safe for concurrent use.
type Reporter struct {
	cfg    Config
	client *http.Client
	events chan apitypes.AgentEvent
	logger *slog.Logger

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new Reporter and starts the background flush loop.
// It also registers the agent with the control plane.
func New(cfg Config) (*Reporter, error) {
	cfg.defaults()

	if cfg.ControlPlaneURL == "" {
		return nil, fmt.Errorf("control plane URL is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if cfg.AgentName == "" {
		return nil, fmt.Errorf("agent name is required")
	}

	ctx, cancel := context.WithCancel(context.Background())

	r := &Reporter{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		events: make(chan apitypes.AgentEvent, cfg.BufferSize),
		logger: slog.Default().With("component", "reporter"),
		cancel: cancel,
	}

	// Start background flush loop
	r.wg.Add(1)
	go r.flushLoop(ctx)

	// Register agent (best-effort, don't fail if it doesn't work)
	go r.register(context.Background())

	return r, nil
}

// Report queues an event for sending to the control plane.
// It is non-blocking; if the buffer is full, the event is dropped.
func (r *Reporter) Report(ctx context.Context, event apitypes.AgentEvent) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.APIVersion == "" {
		event.APIVersion = r.cfg.APIVersion
	}
	select {
	case r.events <- event:
		return nil
	default:
		r.logger.Warn("event buffer full, dropping event", "event_type", event.EventType)
		return nil
	}
}

// ReportSimple is a convenience method for reporting a simple event.
func (r *Reporter) ReportSimple(ctx context.Context, eventType apitypes.EventType, severity apitypes.Severity, payload map[string]any) {
	_ = r.Report(ctx, apitypes.AgentEvent{
		EventType: eventType,
		Severity:  severity,
		Payload:   payload,
		Timestamp: time.Now(),
	})
}

// ReportWorkflow reports a workflow state transition event.
func (r *Reporter) ReportWorkflow(ctx context.Context, issueNumber int, fromState, toState string, reason string) {
	_ = r.Report(ctx, apitypes.AgentEvent{
		EventType:     apitypes.EventWorkflowTransition,
		Severity:      apitypes.SeverityInfo,
		IssueNumber:   issueNumber,
		WorkflowState: toState,
		Payload: map[string]any{
			"from_state": fromState,
			"to_state":   toState,
			"reason":     reason,
		},
		Timestamp: time.Now(),
	})
}

// Heartbeat starts a background goroutine that sends periodic heartbeat events.
func (r *Reporter) Heartbeat(ctx context.Context, interval time.Duration) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.sendHeartbeat(ctx)
			}
		}
	}()
}

// Flush sends all buffered events immediately.
func (r *Reporter) Flush(ctx context.Context) error {
	var events []apitypes.AgentEvent
	for {
		select {
		case e := <-r.events:
			events = append(events, e)
		default:
			if len(events) > 0 {
				return r.sendBatch(ctx, events)
			}
			return nil
		}
	}
}

// Close stops the reporter and flushes remaining events.
func (r *Reporter) Close() error {
	r.cancel()

	// Send stopped event
	_ = r.Report(context.Background(), apitypes.AgentEvent{
		EventType: apitypes.EventAgentStopped,
		Severity:  apitypes.SeverityInfo,
		Timestamp: time.Now(),
	})

	// Flush remaining events with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := r.Flush(ctx); err != nil {
		r.logger.Error("failed to flush events on close", "error", err)
	}

	r.wg.Wait()
	return nil
}

func (r *Reporter) flushLoop(ctx context.Context) {
	defer r.wg.Done()
	ticker := time.NewTicker(r.cfg.FlushInterval)
	defer ticker.Stop()

	var buffer []apitypes.AgentEvent

	for {
		select {
		case <-ctx.Done():
			// Drain remaining events
			for {
				select {
				case e := <-r.events:
					buffer = append(buffer, e)
				default:
					if len(buffer) > 0 {
						flushCtx, cancel := context.WithTimeout(context.Background(), r.cfg.Timeout)
						_ = r.sendBatch(flushCtx, buffer)
						cancel()
					}
					return
				}
			}
		case e := <-r.events:
			buffer = append(buffer, e)
			if len(buffer) >= r.cfg.BufferSize {
				_ = r.sendBatch(ctx, buffer)
				buffer = nil
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				_ = r.sendBatch(ctx, buffer)
				buffer = nil
			}
		}
	}
}

func (r *Reporter) sendBatch(ctx context.Context, events []apitypes.AgentEvent) error {
	req := apitypes.BatchEventRequest{
		AgentName: r.cfg.AgentName,
		Events:    events,
	}
	body, err := json.Marshal(req)
	if err != nil {
		r.logger.Error("failed to marshal events", "error", err)
		return fmt.Errorf("marshaling events: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		r.cfg.ControlPlaneURL+"/api/v1/ingest/events/batch", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+r.cfg.APIKey)

	resp, err := r.client.Do(httpReq)
	if err != nil {
		r.logger.Warn("failed to send events", "error", err, "count", len(events))
		return fmt.Errorf("sending events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		r.logger.Warn("control plane rejected events", "status", resp.StatusCode, "count", len(events))
		return fmt.Errorf("control plane returned status %d", resp.StatusCode)
	}

	r.logger.Debug("sent events to control plane", "count", len(events))
	return nil
}

func (r *Reporter) sendHeartbeat(ctx context.Context) {
	req := apitypes.HeartbeatRequest{
		AgentName: r.cfg.AgentName,
		Status:    "online",
	}
	body, err := json.Marshal(req)
	if err != nil {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		r.cfg.ControlPlaneURL+"/api/v1/ingest/heartbeat", bytes.NewReader(body))
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+r.cfg.APIKey)
	httpReq.Header.Set("Accept", fmt.Sprintf("application/vnd.openagent.%s+json", r.cfg.APIVersion))

	resp, err := r.client.Do(httpReq)
	if err != nil {
		r.logger.Debug("heartbeat failed", "error", err)
		return
	}
	defer resp.Body.Close()
}

func (r *Reporter) register(ctx context.Context) {
	reg := apitypes.AgentRegistration{
		Name:        r.cfg.AgentName,
		AgentType:   r.cfg.AgentType,
		Version:     r.cfg.Version,
		Hostname:    r.cfg.Hostname,
		GitHubOwner: r.cfg.GitHubOwner,
		GitHubRepo:  r.cfg.GitHubRepo,
		Tags:        r.cfg.Tags,
		APIVersion:  r.cfg.APIVersion,
	}
	body, err := json.Marshal(reg)
	if err != nil {
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		r.cfg.ControlPlaneURL+"/api/v1/ingest/register", bytes.NewReader(body))
	if err != nil {
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+r.cfg.APIKey)
	httpReq.Header.Set("Accept", fmt.Sprintf("application/vnd.openagent.%s+json", r.cfg.APIVersion))

	resp, err := r.client.Do(httpReq)
	if err != nil {
		r.logger.Warn("agent registration failed", "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 {
		r.logger.Info("agent registered with control plane", "name", r.cfg.AgentName)

		// Report started event
		_ = r.Report(ctx, apitypes.AgentEvent{
			EventType: apitypes.EventAgentStarted,
			Severity:  apitypes.SeverityInfo,
			Payload: map[string]any{
				"version":  r.cfg.Version,
				"hostname": r.cfg.Hostname,
			},
			Timestamp: time.Now(),
		})
	}
}
