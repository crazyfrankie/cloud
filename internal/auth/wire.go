//go:build wireinject

package auth

import (
	"github.com/crazyfrankie/cloud/internal/auth/handler"
	"github.com/crazyfrankie/cloud/internal/auth/service"
	"github.com/crazyfrankie/cloud/internal/user"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
)

type Handler = handler.AuthHandler
type TokenService = service.TokenService

type Module struct {
	Handler *Handler
	Token   *TokenService
}

func InitAuthModule(u *user.Module, cmd redis.Cmdable) *Module {
	wire.Build(
		service.NewAuthService,
		service.NewTokenService,
		handler.NewAuthHandler,

		wire.FieldsOf(new(*user.Module), "Service"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
