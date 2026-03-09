package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gaskaj/OpenAgentFramework/pkg/apitypes"
)

// ConfigFetcher polls the control plane for configuration updates.
type ConfigFetcher struct {
	cfg      Config
	client   *http.Client
	lastETag string
	logger   *slog.Logger
	mu       sync.Mutex
}

// NewConfigFetcher creates a new ConfigFetcher.
func NewConfigFetcher(cfg Config) *ConfigFetcher {
	cfg.defaults()
	return &ConfigFetcher{
		cfg: cfg,
		client: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: slog.Default().With("component", "config_fetcher"),
	}
}

// FetchConfig fetches the merged configuration from the control plane.
// Returns nil if the config hasn't changed since the last fetch (ETag match).
func (f *ConfigFetcher) FetchConfig(ctx context.Context) (*apitypes.ConfigResponse, error) {
	f.mu.Lock()
	lastETag := f.lastETag
	f.mu.Unlock()

	url := f.cfg.ControlPlaneURL + "/api/v1/ingest/config"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating config request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+f.cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	if lastETag != "" {
		req.Header.Set("If-None-Match", lastETag)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching config: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("control plane returned status %d", resp.StatusCode)
	}

	var configResp apitypes.ConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&configResp); err != nil {
		return nil, fmt.Errorf("decoding config response: %w", err)
	}

	// Update ETag
	if etag := resp.Header.Get("ETag"); etag != "" {
		f.mu.Lock()
		f.lastETag = etag
		f.mu.Unlock()
	}

	return &configResp, nil
}

// PollLoop starts a polling loop that calls onChange when configuration changes.
// It blocks until ctx is cancelled.
func (f *ConfigFetcher) PollLoop(ctx context.Context, interval time.Duration, onChange func(*apitypes.ConfigResponse)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			resp, err := f.FetchConfig(ctx)
			if err != nil {
				f.logger.Warn("config poll failed", "error", err)
				continue
			}
			if resp != nil {
				f.logger.Info("configuration updated from control plane", "version", resp.Version)
				onChange(resp)
			}
		}
	}
}
