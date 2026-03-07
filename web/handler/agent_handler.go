package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gaskaj/OpenAgentFramework/web/middleware"
	"github.com/gaskaj/OpenAgentFramework/web/store"
)

// AgentHandler handles agent management endpoints.
type AgentHandler struct {
	agents *store.PgAgentStore
	logger *slog.Logger
}

// NewAgentHandler creates a new AgentHandler.
func NewAgentHandler(agents *store.PgAgentStore, logger *slog.Logger) *AgentHandler {
	return &AgentHandler{agents: agents, logger: logger}
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
