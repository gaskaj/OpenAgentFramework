package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gaskaj/OpenAgentFramework/web/store"
	"github.com/gaskaj/OpenAgentFramework/web/tunnel"
)

const settingsKeyNgrokToken = "ngrok_authtoken"

// TunnelHandler exposes endpoints for managing the ngrok tunnel.
type TunnelHandler struct {
	mgr      *tunnel.Manager
	settings *store.PgSettingsStore
	logger   *slog.Logger
}

// NewTunnelHandler creates a new TunnelHandler.
func NewTunnelHandler(mgr *tunnel.Manager, settings *store.PgSettingsStore, logger *slog.Logger) *TunnelHandler {
	return &TunnelHandler{mgr: mgr, settings: settings, logger: logger}
}

// HandleStatus returns the current tunnel status.
func (h *TunnelHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, h.mgr.GetStatus())
}

// HandleToggle enables or disables the tunnel.
func (h *TunnelHandler) HandleToggle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Enabled {
		if err := h.mgr.Start(context.Background()); err != nil {
			h.logger.Error("starting tunnel", "error", err)
			respondJSON(w, http.StatusOK, h.mgr.GetStatus())
			return
		}
	} else {
		h.mgr.Stop()
	}

	respondJSON(w, http.StatusOK, h.mgr.GetStatus())
}

// HandleSaveToken saves the ngrok authtoken and optionally starts the tunnel.
func (h *TunnelHandler) HandleSaveToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AuthToken string `json:"auth_token"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Persist to database
	if req.AuthToken != "" {
		if err := h.settings.Set(r.Context(), settingsKeyNgrokToken, req.AuthToken); err != nil {
			h.logger.Error("saving ngrok authtoken", "error", err)
			respondError(w, http.StatusInternalServerError, "failed to save authtoken")
			return
		}
	} else {
		if err := h.settings.Delete(r.Context(), settingsKeyNgrokToken); err != nil {
			h.logger.Error("deleting ngrok authtoken", "error", err)
			respondError(w, http.StatusInternalServerError, "failed to clear authtoken")
			return
		}
	}

	// Stop existing tunnel if running
	h.mgr.Stop()

	// Update the manager's token
	h.mgr.SetAuthToken(req.AuthToken)

	// Auto-start the tunnel if a token was provided
	if req.AuthToken != "" {
		if err := h.mgr.Start(context.Background()); err != nil {
			h.logger.Warn("tunnel failed to start after saving token", "error", err)
		}
	}

	respondJSON(w, http.StatusOK, h.mgr.GetStatus())
}

// LoadAndStart loads the authtoken from the database and starts the tunnel if set.
// Call this during server boot.
func (h *TunnelHandler) LoadAndStart(ctx context.Context) {
	token, err := h.settings.Get(ctx, settingsKeyNgrokToken)
	if err != nil {
		h.logger.Error("loading ngrok authtoken from database", "error", err)
		return
	}
	if token == "" {
		h.logger.Info("ngrok tunnel disabled (no authtoken configured in Settings)")
		return
	}

	h.mgr.SetAuthToken(token)
	if err := h.mgr.Start(ctx); err != nil {
		h.logger.Warn("ngrok tunnel failed to start on boot", "error", err)
	}
}
