package ws

import (
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Hub manages WebSocket connections organized by organization.
type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*Client]bool // orgID -> clients
	logger  *slog.Logger
}

// NewHub creates a new WebSocket hub.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[*Client]bool),
		logger:  logger,
	}
}

// Register adds a client to the hub for a given org.
func (h *Hub) Register(orgID uuid.UUID, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[orgID] == nil {
		h.clients[orgID] = make(map[*Client]bool)
	}
	h.clients[orgID][client] = true
	h.logger.Debug("websocket client registered", "org", orgID, "clients", len(h.clients[orgID]))
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(orgID uuid.UUID, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[orgID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, orgID)
		}
	}
	h.logger.Debug("websocket client unregistered", "org", orgID)
}

// Broadcast sends a message to all clients in an organization.
func (h *Hub) Broadcast(orgID uuid.UUID, message []byte) {
	h.mu.RLock()
	clients := h.clients[orgID]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.send <- message:
		default:
			// Client send buffer full, remove it
			go func(c *Client) {
				h.Unregister(orgID, c)
				c.Close()
			}(client)
		}
	}
}

// ClientCount returns the number of connected clients for an org.
func (h *Hub) ClientCount(orgID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[orgID])
}
