//go:build wireinject

package storage

import (
	"github.com/crazyfrankie/cloud/internal/storage/handler"
	"github.com/crazyfrankie/cloud/internal/storage/service"
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
)

type Handler = handler.StorageHandler

type Module struct {
	Handler *Handler
}

func InitStorageModule(minio *minio.Client) *Module {
	wire.Build(
		service.NewStorageService,
		handler.NewStorageHandler,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
