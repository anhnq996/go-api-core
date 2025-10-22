package routes

import (
	"anhnq/api-core/internal/app/auth"
	"anhnq/api-core/internal/app/user"
	"anhnq/api-core/pkg/jwt"

	"github.com/go-chi/chi/v5"
)

// Controllers chứa tất cả các handler của các module
type Controllers struct {
	UserHandler  *user.Handler
	AuthHandler  *auth.Handler
	JWTManager   *jwt.Manager
	JWTBlacklist *jwt.Blacklist
}

// NewControllers tạo Controllers với tất cả handlers (dùng cho Wire DI)
func NewControllers(
	userHandler *user.Handler,
	authHandler *auth.Handler,
	jwtManager *jwt.Manager,
	jwtBlacklist *jwt.Blacklist,
) *Controllers {
	return &Controllers{
		UserHandler:  userHandler,
		AuthHandler:  authHandler,
		JWTManager:   jwtManager,
		JWTBlacklist: jwtBlacklist,
	}
}

// RegisterRoutes đăng ký tất cả routes cho ứng dụng
// Mỗi module sẽ có prefix riêng và quản lý routes của chính nó
func RegisterRoutes(r chi.Router, c *Controllers) {
	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes - /api/v1/auth/*
		auth.RegisterRoutes(r, c.AuthHandler, c.JWTManager, c.JWTBlacklist)

		// User routes - /api/v1/users/* (Protected)
		r.Group(func(r chi.Router) {
			// Apply JWT middleware for user routes
			r.Use(c.JWTManager.MiddlewareWithBlacklist(c.JWTBlacklist))
			user.RegisterRoutes(r, c.UserHandler)
		})

		// Thêm các module khác ở đây
		// order.RegisterRoutes(r, c.OrderHandler)
		// product.RegisterRoutes(r, c.ProductHandler)
	})
}
