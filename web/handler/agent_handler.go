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

// AgentHandler handles agent management endpoints.
type AgentHandler struct {
	agents  *store.PgAgentStore
	apikeys *store.PgAPIKeyStore
	logger  *slog.Logger
}

// NewAgentHandler creates a new AgentHandler.
func NewAgentHandler(agents *store.PgAgentStore, apikeys *store.PgAPIKeyStore, logger *slog.Logger) *AgentHandler {
	return &AgentHandler{agents: agents, apikeys: apikeys, logger: logger}
}

func (h *AgentHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	var req struct {
		Name        string   `json:"name"`
		AgentType   string   `json:"agent_type"`
		Description string   `json:"description"`
		GitHubOwner string   `json:"github_owner"`
		GitHubRepo  string   `json:"github_repo"`
		Tags        []string `json:"tags"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Name == "" || req.AgentType == "" {
		respondError(w, http.StatusBadRequest, "name and agent_type are required")
		return
	}

	agent := &store.Agent{
		OrgID:       orgCtx.OrgID,
		Name:        req.Name,
		AgentType:   req.AgentType,
		Description: req.Description,
		GitHubOwner: req.GitHubOwner,
		GitHubRepo:  req.GitHubRepo,
		Status:      "registered",
		Tags:        req.Tags,
	}
	if agent.Tags == nil {
		agent.Tags = []string{}
	}

	if err := h.agents.Register(r.Context(), agent); err != nil {
		h.logger.Error("registering agent", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to register agent")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{"data": agent})
}

func (h *AgentHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	opts := store.ListOpts{
		Limit:  parseIntParam(r, "limit", 20),
		Offset: parseIntParam(r, "offset", 0),
	}

	agents, total, err := h.agents.ListByOrg(r.Context(), orgCtx.OrgID, opts)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list agents")
		return
	}

	respondList(w, agents, total, opts.Limit, opts.Offset)
}

func (h *AgentHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	agent, err := h.agents.GetByID(r.Context(), agentID)
	if err != nil || agent == nil {
		respondError(w, http.StatusNotFound, "agent not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": agent})
}

func (h *AgentHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	agent, err := h.agents.GetByID(r.Context(), agentID)
	if err != nil || agent == nil {
		respondError(w, http.StatusNotFound, "agent not found")
		return
	}

	var req struct {
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Apply updates via direct SQL for simplicity
	agent.Description = req.Description
	if req.Tags != nil {
		agent.Tags = req.Tags
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": agent})
}

func (h *AgentHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	agentID, err := uuid.Parse(chi.URLParam(r, "agentId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	if err := h.agents.Delete(r.Context(), agentID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete agent")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// HandleProvision creates a new agent (status "offline") and a corresponding API key.
// Returns the agent, the API key metadata, and the raw key (shown only once).
func (h *AgentHandler) HandleProvision(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetUserFromContext(r.Context())
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if authCtx == nil || orgCtx == nil {
		respondError(w, http.StatusForbidden, "unauthorized")
		return
	}

	var req struct {
		AgentType string `json:"agent_type"`
		Name      string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.AgentType == "" {
		req.AgentType = "developer"
	}

	// Auto-generate agent name if not provided
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

	// 1. Create the agent with status "offline"
	agent := &store.Agent{
		OrgID:     orgCtx.OrgID,
		Name:      agentName,
		AgentType: req.AgentType,
		Status:    "offline",
		Tags:      []string{},
	}
	if err := h.agents.Register(r.Context(), agent); err != nil {
		h.logger.Error("provisioning agent", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create agent")
		return
	}

	// 2. Create an API key bound to this agent
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate key")
		return
	}
	rawKey := fmt.Sprintf("oaf_%s", hex.EncodeToString(keyBytes))
	hash := sha256.Sum256([]byte(rawKey))
	hashStr := hex.EncodeToString(hash[:])
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
		h.logger.Error("creating API key for provisioned agent", "error", err)
		respondError(w, http.StatusInternalServerError, "agent created but API key generation failed")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"agent":   agent,
		"api_key": apiKey,
		"key":     rawKey,
	})
}
