package chat

import "github.com/go-chi/chi/v5"

// RegisterRoutes đăng ký tất cả routes cho module chat
// Prefix: /api/v1/chats
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/chats", func(r chi.Router) {
		// Conversations
		r.Route("/conversations", func(r chi.Router) {
			r.Get("/", h.GetConversations)         // GET /api/v1/chats/conversations - Danh sách conversations
			r.Post("/", h.GetOrCreateConversation) // POST /api/v1/chats/conversations - Lấy/tạo conversation
			r.Get("/{id}/messages", h.GetMessages) // GET /api/v1/chats/conversations/{id}/messages - Lấy tin nhắn
		})

		// Messages
		r.Post("/messages", h.SendMessage) // POST /api/v1/chats/messages - Gửi tin nhắn
	})
}
