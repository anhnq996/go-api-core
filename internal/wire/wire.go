//go:build wireinject
// +build wireinject

package wire

import (
	"api-core/internal/app/auth"
	"api-core/internal/app/user"
	repository "api-core/internal/repositories"
	"api-core/internal/routes"
	"api-core/pkg/cache"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// InitializeApp khởi tạo toàn bộ ứng dụng với database và cache
func InitializeApp(db *gorm.DB, cacheClient cache.Cache) (*routes.Controllers, error) {
	wire.Build(
		// JWT
		ProvideJWTManager,
		ProvideJWTBlacklist,

		// Storage
		ProvideStorageManager,

		// Repositories (cần DB)
		repository.NewUserRepository,

		// Services (cần Repo + Cache + Storage)
		user.NewService,
		auth.NewService,

		// Handlers
		user.NewHandler,
		auth.NewHandler,

		// Controllers
		routes.NewControllers,

		// Cache interface
		ProvideCacheInterface,
	)

	return nil, nil // Wire sẽ thay thế dòng này
}

// ProvideCacheInterface provides cache interface for rate limiting
func ProvideCacheInterface(cacheClient cache.Cache) routes.CacheInterface {
	return cacheClient
}
