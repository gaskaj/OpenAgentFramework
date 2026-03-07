package reporter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gaskaj/OpenAgentFramework/pkg/apitypes"
)

func TestConfigDefaults(t *testing.T) {
	cfg := Config{}
	cfg.defaults()

	if cfg.BufferSize != 100 {
		t.Errorf("expected BufferSize 100, got %d", cfg.BufferSize)
	}
	if cfg.FlushInterval != 5*time.Second {
		t.Errorf("expected FlushInterval 5s, got %v", cfg.FlushInterval)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("expected Timeout 10s, got %v", cfg.Timeout)
	}
	if cfg.Hostname == "" {
		t.Error("expected Hostname to be set")
	}
}

func TestNewReporterValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{"missing URL", Config{APIKey: "key", AgentName: "agent"}, true},
		{"missing API key", Config{ControlPlaneURL: "http://localhost", AgentName: "agent"}, true},
		{"missing agent name", Config{ControlPlaneURL: "http://localhost", APIKey: "key"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
			if r != nil {
				r.Close()
			}
		})
	}
}

func TestReportBuffering(t *testing.T) {
	var received atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/ingest/events/batch" {
			var req apitypes.BatchEventRequest
			json.NewDecoder(r.Body).Decode(&req)
			received.Add(int64(len(req.Events)))
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	r, err := New(Config{
		ControlPlaneURL: server.URL,
		APIKey:          "test-key",
		AgentName:       "test-agent",
		AgentType:       "developer",
		FlushInterval:   100 * time.Millisecond,
		BufferSize:      10,
	})
	if err != nil {
		t.Fatalf("failed to create reporter: %v", err)
	}

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		r.Report(ctx, apitypes.AgentEvent{
			EventType: apitypes.EventIssueClaimed,
			Severity:  apitypes.SeverityInfo,
			Timestamp: time.Now(),
		})
	}

	// Wait for flush
	time.Sleep(300 * time.Millisecond)

	if err := r.Close(); err != nil {
		t.Fatalf("failed to close reporter: %v", err)
	}

	// Should have received events (started + 5 reported + stopped)
	if received.Load() < 5 {
		t.Errorf("expected at least 5 events received, got %d", received.Load())
	}
}

func TestReportNonBlocking(t *testing.T) {
	r, err := New(Config{
		ControlPlaneURL: "http://localhost:99999", // unreachable
		APIKey:          "test-key",
		AgentName:       "test-agent",
		BufferSize:      2,
		FlushInterval:   1 * time.Hour, // don't auto-flush
	})
	if err != nil {
		t.Fatalf("failed to create reporter: %v", err)
	}
	defer r.Close()

	// Fill buffer
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		err := r.Report(ctx, apitypes.AgentEvent{
			EventType: apitypes.EventAgentHeartbeat,
			Severity:  apitypes.SeverityInfo,
			Timestamp: time.Now(),
		})
		if err != nil {
			t.Fatalf("Report should not return error even when buffer is full: %v", err)
		}
	}
}
