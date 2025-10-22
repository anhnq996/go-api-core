//go:build wireinject
// +build wireinject

package wire

import (
	"anhnq/api-core/internal/app/user"
	repository "anhnq/api-core/internal/repositories"
	"anhnq/api-core/internal/routes"
	"anhnq/api-core/pkg/cache"

	"github.com/google/wire"
	"gorm.io/gorm"
)

// InitializeApp khởi tạo toàn bộ ứng dụng với database và cache
func InitializeApp(db *gorm.DB, cacheClient cache.Cache) *routes.Controllers {
	wire.Build(
		// Repositories (cần DB)
		repository.NewUserRepository,

		// Services (cần Repo + Cache)
		user.NewService,

		// Handlers
		user.NewHandler,

		// Controllers
		routes.NewControllers,
	)

	return nil // Wire sẽ thay thế dòng này
}
