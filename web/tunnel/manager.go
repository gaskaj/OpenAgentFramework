package tunnel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Status represents the current state of the ngrok tunnel.
type Status struct {
	Enabled      bool   `json:"enabled"`
	PublicURL    string `json:"public_url,omitempty"`
	Error        string `json:"error,omitempty"`
	HasAuthToken bool   `json:"has_auth_token"`
}

// Manager controls the lifecycle of an ngrok tunnel process.
type Manager struct {
	targetAddr string // address to tunnel to (e.g. "frontend:80")
	logger     *slog.Logger

	mu        sync.Mutex
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	enabled   bool
	url       string
	lastErr   string
	authToken string
}

// NewManager creates a tunnel manager that will forward to the given address.
// The targetAddr should be the address of the service to expose publicly
// (typically the frontend, which proxies API requests to the control plane).
func NewManager(targetAddr string, logger *slog.Logger) *Manager {
	return &Manager{
		targetAddr: targetAddr,
		logger:     logger,
	}
}

// SetAuthToken sets the ngrok authtoken to use for tunnel connections.
func (m *Manager) SetAuthToken(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authToken = token
}

// Start launches the ngrok tunnel. It is safe to call multiple times.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.enabled && m.cmd != nil && m.cmd.Process != nil {
		return nil // already running
	}

	// Check ngrok is installed
	if _, err := exec.LookPath("ngrok"); err != nil {
		m.lastErr = "ngrok is not installed"
		return fmt.Errorf("ngrok not found in PATH: %w", err)
	}

	if m.authToken == "" {
		m.lastErr = "ngrok authtoken not configured — enter it in Settings"
		return fmt.Errorf("ngrok authtoken not configured")
	}

	tunnelCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel

	var stderrBuf bytes.Buffer
	m.cmd = exec.CommandContext(tunnelCtx, "ngrok", "http", m.targetAddr, "--log=stdout", "--log-format=json")
	m.cmd.Env = append(m.cmd.Environ(), "NGROK_AUTHTOKEN="+m.authToken)
	m.cmd.Stderr = &stderrBuf
	if err := m.cmd.Start(); err != nil {
		m.lastErr = err.Error()
		cancel()
		return fmt.Errorf("starting ngrok: %w", err)
	}

	m.enabled = true
	m.lastErr = ""
	m.logger.Info("ngrok tunnel starting", "target", m.targetAddr, "pid", m.cmd.Process.Pid)

	// Wait for the process to exit in the background
	go func() {
		if err := m.cmd.Wait(); err != nil {
			m.mu.Lock()
			if m.enabled {
				errMsg := strings.TrimSpace(stderrBuf.String())
				if errMsg == "" {
					errMsg = err.Error()
				}
				m.lastErr = errMsg
				m.logger.Error("ngrok process exited", "error", errMsg)
			}
			m.enabled = false
			m.url = ""
			m.mu.Unlock()
		}
	}()

	// Poll the ngrok API for the tunnel URL
	go m.pollURL()

	return nil
}

// Stop tears down the ngrok tunnel.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.enabled = false
	m.url = ""
	m.cmd = nil
	m.lastErr = ""
	m.logger.Info("ngrok tunnel stopped")
}

// GetStatus returns the current tunnel status.
func (m *Manager) GetStatus() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return Status{
		Enabled:      m.enabled,
		PublicURL:    m.url,
		Error:        m.lastErr,
		HasAuthToken: m.authToken != "",
	}
}

// ngrokTunnelsResponse is the response from ngrok's local API.
type ngrokTunnelsResponse struct {
	Tunnels []struct {
		PublicURL string `json:"public_url"`
		Proto     string `json:"proto"`
	} `json:"tunnels"`
}

func (m *Manager) pollURL() {
	client := &http.Client{Timeout: 2 * time.Second}

	for i := 0; i < 30; i++ {
		time.Sleep(500 * time.Millisecond)

		m.mu.Lock()
		if !m.enabled {
			m.mu.Unlock()
			return
		}
		m.mu.Unlock()

		resp, err := client.Get("http://127.0.0.1:4040/api/tunnels")
		if err != nil {
			continue
		}

		var tunnels ngrokTunnelsResponse
		if err := json.NewDecoder(resp.Body).Decode(&tunnels); err != nil {
			resp.Body.Close()
			continue
		}
		resp.Body.Close()

		for _, t := range tunnels.Tunnels {
			if t.Proto == "https" || t.PublicURL != "" {
				m.mu.Lock()
				m.url = t.PublicURL
				m.mu.Unlock()
				m.logger.Info("ngrok tunnel established", "url", t.PublicURL)
				return
			}
		}
	}

	m.mu.Lock()
	if m.url == "" {
		m.lastErr = "timed out waiting for ngrok tunnel URL"
		m.logger.Warn("timed out polling ngrok for tunnel URL")
	}
	m.mu.Unlock()
}
