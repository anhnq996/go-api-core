package routes

import (
	"api-core/internal/app/auth"
	"api-core/internal/app/user"
	"api-core/pkg/jwt"
	middlewarePkg "api-core/pkg/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
)

// Controllers chứa tất cả các handler của các module
type Controllers struct {
	UserHandler  *user.Handler
	AuthHandler  *auth.Handler
	JWTManager   *jwt.Manager
	JWTBlacklist *jwt.Blacklist
	Cache        CacheInterface
}

// CacheInterface defines cache interface for rate limiting
type CacheInterface interface {
	GetRedisClient() *redis.Client
}

// NewControllers tạo Controllers với tất cả handlers (dùng cho Wire DI)
func NewControllers(
	userHandler *user.Handler,
	authHandler *auth.Handler,
	jwtManager *jwt.Manager,
	jwtBlacklist *jwt.Blacklist,
	cache CacheInterface,
) *Controllers {
	return &Controllers{
		UserHandler:  userHandler,
		AuthHandler:  authHandler,
		JWTManager:   jwtManager,
		JWTBlacklist: jwtBlacklist,
		Cache:        cache,
	}
}

// RegisterRoutes đăng ký tất cả routes cho ứng dụng
// Mỗi module sẽ có prefix riêng và quản lý routes của chính nó
func RegisterRoutes(r chi.Router, c *Controllers) {
	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes - /api/v1/auth/* (with rate limiting)
		r.Group(func(r chi.Router) {
			// Rate limiting cho auth routes: 5 requests per 15 minutes by IP
			r.Use(middlewarePkg.RateLimitByIP(c.Cache.GetRedisClient(), 150, 60))
			auth.RegisterRoutes(r, c.AuthHandler, c.JWTManager, c.JWTBlacklist)
		})

		// User routes - /api/v1/users/* (Protected with rate limiting)
		r.Group(func(r chi.Router) {
			// Apply JWT middleware for user routes
			r.Use(c.JWTManager.MiddlewareWithBlacklist(c.JWTBlacklist))
			// Rate limiting cho user routes: 100 requests per minute by user or IP
			r.Use(middlewarePkg.RateLimitByUserOrIP(c.Cache.GetRedisClient(), 150, 60))
			user.RegisterRoutes(r, c.UserHandler)
		})

		// Thêm các module khác ở đây
		// order.RegisterRoutes(r, c.OrderHandler)
		// product.RegisterRoutes(r, c.ProductHandler)
	})
}
