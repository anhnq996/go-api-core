package auth

import (
	"api-core/pkg/jwt"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes đăng ký auth routes
func RegisterRoutes(r chi.Router, handler *Handler, jwtManager *jwt.Manager, blacklist *jwt.Blacklist) {
	// Public routes
	r.Post("/auth/login", handler.Login)
	r.Post("/auth/register", handler.Register)
	r.Post("/auth/refresh", handler.RefreshToken)

	// Protected routes
	r.Group(func(r chi.Router) {
		// Apply JWT middleware với blacklist
		r.Use(jwtManager.MiddlewareWithBlacklist(blacklist))

		r.Get("/auth/me", handler.GetMe)
		r.Post("/auth/logout", handler.Logout)
		r.Post("/auth/logout-all", handler.LogoutAll)
	})
}
