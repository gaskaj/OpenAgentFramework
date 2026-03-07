package handler

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gaskaj/OpenAgentFramework/web/middleware"
	"github.com/gaskaj/OpenAgentFramework/web/store"
)

// InvitationHandler handles invitation endpoints.
type InvitationHandler struct {
	invitations *store.PgInvitationStore
	orgs        *store.PgOrgStore
	users       *store.PgUserStore
	logger      *slog.Logger
}

// NewInvitationHandler creates a new InvitationHandler.
func NewInvitationHandler(invitations *store.PgInvitationStore, orgs *store.PgOrgStore, users *store.PgUserStore, logger *slog.Logger) *InvitationHandler {
	return &InvitationHandler{invitations: invitations, orgs: orgs, users: users, logger: logger}
}

func (h *InvitationHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	authCtx := middleware.GetUserFromContext(r.Context())
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if authCtx == nil || orgCtx == nil {
		respondError(w, http.StatusForbidden, "unauthorized")
		return
	}

	var req struct {
		Email string `json:"email"`
		Role  string `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Role == "" {
		req.Role = "member"
	}

	// Generate invitation token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	inv := &store.Invitation{
		OrgID:     orgCtx.OrgID,
		InvitedBy: authCtx.UserID,
		Email:     req.Email,
		Role:      req.Role,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := h.invitations.Create(r.Context(), inv); err != nil {
		h.logger.Error("creating invitation", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to create invitation")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]any{
		"data":             inv,
		"invitation_token": token,
	})
}

func (h *InvitationHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	invitations, err := h.invitations.ListByOrg(r.Context(), orgCtx.OrgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list invitations")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"data": invitations})
}

func (h *InvitationHandler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	invID, err := uuid.Parse(chi.URLParam(r, "invId"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid invitation ID")
		return
	}

	if err := h.invitations.Delete(r.Context(), invID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to cancel invitation")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "cancelled"})
}

func (h *InvitationHandler) HandleAccept(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token string `json:"token"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Token == "" {
		respondError(w, http.StatusBadRequest, "token is required")
		return
	}

	inv, err := h.invitations.GetByToken(r.Context(), req.Token)
	if err != nil || inv == nil {
		respondError(w, http.StatusNotFound, "invitation not found")
		return
	}

	if inv.Accepted {
		respondError(w, http.StatusBadRequest, "invitation already accepted")
		return
	}

	if time.Now().After(inv.ExpiresAt) {
		respondError(w, http.StatusBadRequest, "invitation has expired")
		return
	}

	// Find or check user by email
	user, err := h.users.GetByEmail(r.Context(), inv.Email)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}
	if user == nil {
		respondError(w, http.StatusBadRequest, "please register first, then accept the invitation")
		return
	}

	// Add to org
	if err := h.orgs.AddMember(r.Context(), inv.OrgID, user.ID, inv.Role); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to add member")
		return
	}

	// Mark accepted
	h.invitations.MarkAccepted(r.Context(), inv.ID)

	respondJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}
