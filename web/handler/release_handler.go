package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// ReleaseAsset represents a downloadable asset from a GitHub release.
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// LatestRelease represents the latest GitHub release info.
type LatestRelease struct {
	TagName     string         `json:"tag_name"`
	PublishedAt string         `json:"published_at"`
	HTMLURL     string         `json:"html_url"`
	Assets      []ReleaseAsset `json:"assets"`
}

// ReleaseHandler serves cached GitHub release information.
type ReleaseHandler struct {
	owner  string
	repo   string
	logger *slog.Logger

	mu        sync.RWMutex
	cached    *LatestRelease
	cachedAt  time.Time
	cacheTTL  time.Duration
}

// NewReleaseHandler creates a handler that proxies GitHub release information.
func NewReleaseHandler(owner, repo string, logger *slog.Logger) *ReleaseHandler {
	return &ReleaseHandler{
		owner:    owner,
		repo:     repo,
		logger:   logger,
		cacheTTL: 5 * time.Minute,
	}
}

// HandleLatestRelease returns the latest GitHub release with download URLs.
func (h *ReleaseHandler) HandleLatestRelease(w http.ResponseWriter, r *http.Request) {
	release, err := h.getLatestRelease()
	if err != nil {
		h.logger.Error("failed to fetch latest release", "error", err)
		http.Error(w, `{"error":"failed to fetch release information"}`, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(release)
}

func (h *ReleaseHandler) getLatestRelease() (*LatestRelease, error) {
	h.mu.RLock()
	if h.cached != nil && time.Since(h.cachedAt) < h.cacheTTL {
		defer h.mu.RUnlock()
		return h.cached, nil
	}
	h.mu.RUnlock()

	h.mu.Lock()
	defer h.mu.Unlock()

	// Double-check after acquiring write lock
	if h.cached != nil && time.Since(h.cachedAt) < h.cacheTTL {
		return h.cached, nil
	}

	release, err := h.fetchFromGitHub()
	if err != nil {
		// Return stale cache if available
		if h.cached != nil {
			h.logger.Warn("returning stale release cache", "error", err)
			return h.cached, nil
		}
		return nil, err
	}

	h.cached = release
	h.cachedAt = time.Now()
	return release, nil
}

func (h *ReleaseHandler) fetchFromGitHub() (*LatestRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", h.owner, h.repo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "OpenAgentFramework-ControlPlane")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// No releases yet — return empty release
		return &LatestRelease{
			TagName: "",
			Assets:  []ReleaseAsset{},
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(body))
	}

	// GitHub API returns more fields; we parse only what we need.
	var ghRelease struct {
		TagName     string `json:"tag_name"`
		PublishedAt string `json:"published_at"`
		HTMLURL     string `json:"html_url"`
		Assets      []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ghRelease); err != nil {
		return nil, fmt.Errorf("decoding release: %w", err)
	}

	release := &LatestRelease{
		TagName:     ghRelease.TagName,
		PublishedAt: ghRelease.PublishedAt,
		HTMLURL:     ghRelease.HTMLURL,
		Assets:      make([]ReleaseAsset, len(ghRelease.Assets)),
	}
	for i, a := range ghRelease.Assets {
		release.Assets[i] = ReleaseAsset{
			Name:               a.Name,
			BrowserDownloadURL: a.BrowserDownloadURL,
			Size:               a.Size,
		}
	}

	return release, nil
}
