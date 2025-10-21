//go:build wireinject
// +build wireinject

package wire

import (
	"anhnq/api-core/internal/app/user"
	repository "anhnq/api-core/internal/repositories"
	"anhnq/api-core/internal/routes"

	"github.com/google/wire"
)

// InitializeApp khởi tạo toàn bộ ứng dụng
func InitializeApp() *routes.Controllers {
	wire.Build(
		// Repositories
		repository.NewUserRepository,

		// Services
		user.NewService,

		// Handlers
		user.NewHandler,

		// Controllers
		routes.NewControllers,
	)

	return nil // Wire sẽ thay thế dòng này
}
