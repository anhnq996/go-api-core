package user

import "github.com/go-chi/chi/v5"

// RegisterRoutes đăng ký tất cả routes cho module user
// Prefix: /api/v1/users
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/", h.Index)          // GET /api/v1/users - Lấy danh sách users
		r.Post("/", h.Store)         // POST /api/v1/users - Tạo user mới (có thể kèm avatar)
		r.Get("/{id}", h.Show)       // GET /api/v1/users/{id} - Lấy user theo ID
		r.Put("/{id}", h.Update)     // PUT /api/v1/users/{id} - Cập nhật user (có thể kèm avatar)
		r.Delete("/{id}", h.Destroy) // DELETE /api/v1/users/{id} - Xóa user
	})
}
