package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gaskaj/OpenAgentFramework/web/middleware"
	"github.com/gaskaj/OpenAgentFramework/web/store"
)

// OrgHandler handles organization management endpoints.
type OrgHandler struct {
	orgs   *store.PgOrgStore
	logger *slog.Logger
}

// NewOrgHandler creates a new OrgHandler.
func NewOrgHandler(orgs *store.PgOrgStore, logger *slog.Logger) *OrgHandler {
	return &OrgHandler{orgs: orgs, logger: logger}
}

func (h *OrgHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetUserFromContext(r.Context())
	if authCtx == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	if req.Slug == "" {
		req.Slug = slugify(req.Name)
	}

	org := &store.Organization{
		Name:     req.Name,
		Slug:     req.Slug,
		Plan:     "free",
		Settings: json.RawMessage("{}"),
	}
	if err := h.orgs.Create(r.Context(), org); err != nil {
		h.logger.Error("creating org", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create organization")
		return
	}

	h.orgs.AddMember(r.Context(), org.ID, authCtx.UserID, "owner")

	respondJSON(w, http.StatusCreated, map[string]any{"data": org})
}

func (h *OrgHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetUserFromContext(r.Context())
	if authCtx == nil {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orgs, err := h.orgs.ListForUser(r.Context(), authCtx.UserID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list organizations")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": orgs})
}

func (h *OrgHandler) HandleGet(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	org, err := h.orgs.GetByID(r.Context(), orgCtx.OrgID)
	if err != nil || org == nil {
		respondError(w, http.StatusNotFound, "organization not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": org})
}

func (h *OrgHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	org, err := h.orgs.GetByID(r.Context(), orgCtx.OrgID)
	if err != nil || org == nil {
		respondError(w, http.StatusNotFound, "organization not found")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name != "" {
		org.Name = req.Name
	}

	if err := h.orgs.Update(r.Context(), org); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update organization")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": org})
}

func (h *OrgHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil || orgCtx.Role != "owner" {
		respondError(w, http.StatusForbidden, "only owners can delete organizations")
		return
	}

	if err := h.orgs.Delete(r.Context(), orgCtx.OrgID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete organization")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *OrgHandler) HandleListMembers(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	members, err := h.orgs.ListMembers(r.Context(), orgCtx.OrgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list members")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": members})
}

func (h *OrgHandler) HandleUpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	var req struct {
		Role string `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Role == "" {
		respondError(w, http.StatusBadRequest, "role is required")
		return
	}

	if err := h.orgs.UpdateMemberRole(r.Context(), orgCtx.OrgID, userID, req.Role); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update role")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *OrgHandler) HandleRemoveMember(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	userID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.orgs.RemoveMember(r.Context(), orgCtx.OrgID, userID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to remove member")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}
