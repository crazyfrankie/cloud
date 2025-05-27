//go:build wireinject

package user

import (
	"github.com/crazyfrankie/cloud/internal/user/dao"
	"github.com/crazyfrankie/cloud/internal/user/service"
	snowflake "github.com/crazyfrankie/snow-flake"
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"

	"github.com/crazyfrankie/cloud/internal/user/handler"
)

type Handler = handler.UserHandler
type Service = service.UserService

type Module struct {
	Handler *Handler
	Service *Service
}

func InitUserModule(db *gorm.DB, minio *minio.Client, snow *snowflake.Node) *Module {
	wire.Build(
		dao.NewUserDao,
		service.NewUserService,
		handler.NewUserHandler,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
