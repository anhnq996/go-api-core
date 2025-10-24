package socket

import (
	"encoding/json"
	"net/http"
	"time"
)

// Handler handles WebSocket related HTTP requests
type Handler struct {
	hub *Hub
}

// NewHandler creates a new socket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// ServeWebSocket handles WebSocket connections
func (h *Handler) ServeWebSocket(w http.ResponseWriter, r *http.Request) {
	ServeWS(h.hub, w, r)
}

// GetStats returns WebSocket hub statistics
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"total_clients": h.hub.GetClientCount(),
		"total_rooms":   h.hub.GetRoomCount(),
		"timestamp":     time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"code":    "SUCCESS",
		"message": "SUCCESS",
		"data":    stats,
	})
}
