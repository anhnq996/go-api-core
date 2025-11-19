package friend

import "github.com/go-chi/chi/v5"

// RegisterRoutes đăng ký tất cả routes cho module friend
// Prefix: /api/v1/friends
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/friends", func(r chi.Router) {
		// Danh sách bạn bè
		r.Get("/", h.GetFriendsList) // GET /api/v1/friends

		// Friend requests
		r.Route("/requests", func(r chi.Router) {
			r.Post("/", h.SendFriendRequest)         // POST /api/v1/friends/requests - Gửi lời mời
			r.Post("/accept", h.AcceptFriendRequest) // POST /api/v1/friends/requests/accept - Chấp nhận
			r.Post("/reject", h.RejectFriendRequest) // POST /api/v1/friends/requests/reject - Từ chối
			r.Post("/cancel", h.CancelFriendRequest) // POST /api/v1/friends/requests/cancel - Hủy
			r.Get("/pending", h.GetPendingRequests)  // GET /api/v1/friends/requests/pending - Lời mời nhận được
			r.Get("/sent", h.GetSentRequests)        // GET /api/v1/friends/requests/sent - Lời mời đã gửi
		})
	})
}
