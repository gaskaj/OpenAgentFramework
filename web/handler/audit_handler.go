package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/gaskaj/OpenAgentFramework/web/middleware"
	"github.com/gaskaj/OpenAgentFramework/web/store"
)

// AuditHandler handles audit log endpoints.
type AuditHandler struct {
	audit  *store.PgAuditStore
	logger *slog.Logger
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(audit *store.PgAuditStore, logger *slog.Logger) *AuditHandler {
	return &AuditHandler{audit: audit, logger: logger}
}

func (h *AuditHandler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		respondError(w, http.StatusForbidden, "no org context")
		return
	}

	filter := store.AuditFilter{
		OrgID:        orgCtx.OrgID,
		Action:       r.URL.Query().Get("action"),
		ResourceType: r.URL.Query().Get("resource_type"),
		ListOpts: store.ListOpts{
			Limit:  parseIntParam(r, "limit", 20),
			Offset: parseIntParam(r, "offset", 0),
		},
	}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &userID
		}
	}

	if sinceStr := r.URL.Query().Get("since"); sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			filter.Since = &t
		}
	}

	logs, total, err := h.audit.Query(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to query audit logs")
		return
	}

	respondList(w, logs, total, filter.Limit, filter.Offset)
}
