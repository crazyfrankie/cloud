//go:build wireinject

package file

import (
	"github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/internal/file/handler"
	"github.com/crazyfrankie/cloud/internal/file/service"
	"github.com/google/wire"
	"gorm.io/gorm"
)

type Handler = handler.FileHandler

type Module struct {
	Handler *Handler
}

func InitFileModule(db *gorm.DB) *Module {
	wire.Build(
		dao.NewUploadDao,
		service.NewUploadService,
		handler.NewFileHandler,

		wire.Struct(new(Module), "*"),
	)
	return new(Module)
}
