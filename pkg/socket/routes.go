package socket

import (
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes registers WebSocket routes
func RegisterRoutes(r chi.Router, hub *Hub) {
	handler := NewHandler(hub)

	// WebSocket connection endpoint
	r.Get("/ws", handler.ServeWebSocket)

	// WebSocket management endpoints
	r.Route("/socket", func(r chi.Router) {
		// Get WebSocket statistics
		r.Get("/stats", handler.GetStats)
	})
}
