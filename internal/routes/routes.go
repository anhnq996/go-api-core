package routes

import (
	"anhnq/api-core/internal/app/user"

	"github.com/go-chi/chi/v5"
)

// Controllers chứa tất cả các handler của các module
type Controllers struct {
	UserHandler *user.Handler
	// Thêm các handler khác ở đây
	// OrderHandler *order.Handler
	// ProductHandler *product.Handler
}

// NewControllers tạo Controllers với tất cả handlers (dùng cho Wire DI)
func NewControllers(userHandler *user.Handler) *Controllers {
	return &Controllers{
		UserHandler: userHandler,
		// Thêm các handler khác khi có module mới
		// OrderHandler: orderHandler,
	}
}

// RegisterRoutes đăng ký tất cả routes cho ứng dụng
// Mỗi module sẽ có prefix riêng và quản lý routes của chính nó
func RegisterRoutes(r chi.Router, c *Controllers) {
	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// User routes - prefix: /api/v1/users
		user.RegisterRoutes(r, c.UserHandler)

		// Thêm các module khác ở đây
		// order.RegisterRoutes(r, c.OrderHandler)
		// product.RegisterRoutes(r, c.ProductHandler)
	})
}
