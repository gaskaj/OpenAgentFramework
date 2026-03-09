package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gaskaj/OpenAgentFramework/web/middleware"
	"github.com/gaskaj/OpenAgentFramework/web/store"
)

// APIKeyHandler handles API key management endpoints.
type APIKeyHandler struct {
	apikeys *store.PgAPIKeyStore
	logger  *slog.Logger
}

// NewAPIKeyHandler creates a new APIKeyHandler.
func NewAPIKeyHandler(apikeys *store.PgAPIKeyStore, logger *slog.Logger) *APIKeyHandler {
	return &APIKeyHandler{apikeys: apikeys, logger: logger}
}

func (h *APIKeyHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetUserFromContext(r.Context())
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if authCtx == nil || orgCtx == nil {
		respondError(w, http.StatusForbidden, "unauthorized")
		return
	}

	var req struct {
		AgentType string `json:"agent_type"`
		Name      string `json:"name"` // optional override for agent_name
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AgentType == "" {
		req.AgentType = "developer"
	}

	// Auto-generate agent name: {agent_type}-{XX}
	agentName := req.Name
	if agentName == "" {
		count, err := h.apikeys.CountByAgentType(r.Context(), orgCtx.OrgID, req.AgentType)
		if err != nil {
			h.logger.Error("counting agents by type", "error", err)
			respondError(w, http.StatusInternalServerError, "failed to generate agent name")
			return
		}
		agentName = fmt.Sprintf("%s-%02d", req.AgentType, count+1)
	}

	// Generate random key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate key")
		return
	}
	rawKey := fmt.Sprintf("oaf_%s", hex.EncodeToString(keyBytes))

	// Hash the key
	hash := sha256.Sum256([]byte(rawKey))
	hashStr := hex.EncodeToString(hash[:])

	// Prefix is first 8 chars after "oaf_"
	prefix := hex.EncodeToString(keyBytes)[:8]

	apiKey := &store.APIKey{
		OrgID:     orgCtx.OrgID,
		CreatedBy: authCtx.UserID,
		Name:      agentName,
		KeyHash:   hashStr,
		KeyPrefix: prefix,
		Scopes:    []string{"agent.report"},
		AgentType: req.AgentType,
		AgentName: agentName,
	}

	if err := h.apikeys.Create(r.Context(), apiKey); err != nil {
		h.logger.Error("creating API key", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create API key")
		return
	}

	// Return the raw key only this once
	respondJSON(w, http.StatusCreated, map[string]any{
		"data": apiKey,
		"key":  rawKey,
	})
}

func (h *APIKeyHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	keys, err := h.apikeys.ListByOrg(r.Context(), orgCtx.OrgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": keys})
}

func (h *APIKeyHandler) HandleRevoke(w http.ResponseWriter, r *http.Request) {
	keyID, err := uuid.Parse(chi.URLParam(r, "keyId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid key ID")
		return
	}

	if err := h.apikeys.Revoke(r.Context(), keyID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to revoke API key")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}
