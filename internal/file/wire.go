//go:build wireinject

package file

import (
	"github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/internal/file/handler"
	"github.com/crazyfrankie/cloud/internal/file/service"
	"github.com/crazyfrankie/cloud/internal/storage"
	"github.com/google/wire"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Handler = handler.FileHandler

type Module struct {
	Handler *Handler
}

func InitFileModule(db *gorm.DB, st *storage.Module, rdb redis.Cmdable, minio *minio.Client) *Module {
	wire.Build(
		dao.NewFileDao,
		service.NewUploadService,
		service.NewDownloadService,
		handler.NewFileHandler,

		wire.FieldsOf(new(*storage.Module), "Service"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
