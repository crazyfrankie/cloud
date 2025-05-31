//go:build wireinject

package file

import (
	"github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/internal/file/handler"
	"github.com/crazyfrankie/cloud/internal/file/service"
	"github.com/crazyfrankie/cloud/internal/storage"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.FileHandler

type Module struct {
	Handler *Handler
}

func InitFileModule(db *gorm.DB, st *storage.Module) *Module {
	wire.Build(
		dao.NewUploadDao,
		service.NewUploadService,
		handler.NewFileHandler,

		wire.FieldsOf(new(*storage.Module), "Service"),
		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
