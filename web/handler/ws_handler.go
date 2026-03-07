package handler

import (
	"log/slog"
	"net/http"

	"nhooyr.io/websocket"

	"github.com/gaskaj/OpenAgentFramework/web/auth"
	"github.com/gaskaj/OpenAgentFramework/web/middleware"
	"github.com/gaskaj/OpenAgentFramework/web/ws"
)

// WSHandler handles WebSocket connections.
type WSHandler struct {
	hub    *ws.Hub
	jwtMgr *auth.JWTManager
	logger *slog.Logger
}

// NewWSHandler creates a new WSHandler.
func NewWSHandler(hub *ws.Hub, jwtMgr *auth.JWTManager, logger *slog.Logger) *WSHandler {
	return &WSHandler{hub: hub, jwtMgr: jwtMgr, logger: logger}
}

func (h *WSHandler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	orgCtx := middleware.GetOrgFromContext(r.Context())
	if orgCtx == nil {
		// Try JWT from query param for WebSocket connections
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		_, err := h.jwtMgr.ValidateToken(token)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
	}

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		h.logger.Error("websocket accept failed", "error", err)
		return
	}

	client := ws.NewClient(conn)
	orgID := orgCtx.OrgID

	h.hub.Register(orgID, client)

	ctx := r.Context()
	go client.WritePump(ctx)
	client.ReadPump(ctx)

	h.hub.Unregister(orgID, client)
}
