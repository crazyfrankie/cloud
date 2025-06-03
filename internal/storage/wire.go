//go:build wireinject

package storage

import (
	"github.com/crazyfrankie/cloud/internal/storage/service"
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
)

type Service = service.StorageService

type Module struct {
	Service *Service
}

func InitStorageModule(minio *minio.Client) *Module {
	wire.Build(
		service.NewStorageService,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
