package handler

import (
	"net/http"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/storage/service"
	"github.com/crazyfrankie/cloud/pkg/response"
)

type StorageHandler struct {
	svc *service.StorageService
}

func NewStorageHandler(svc *service.StorageService) *StorageHandler {
	return &StorageHandler{svc: svc}
}

func (h *StorageHandler) RegisterRoute(r *gin.Engine) {
	storageGroup := r.Group("storage")
	{
		storageGroup.POST("presign/:type", h.Presign())
	}
}

func (h *StorageHandler) Presign() gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.PostForm("filename")
		typ := c.Param("type")

		if filename == "" {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "filename is required"))
			return
		}

		id := c.MustGet("uid")
		path, err := h.svc.Presign(c.Request.Context(), id.(int64), filename, typ)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, map[string]string{
			"presignedUrl": path,
			"filename":     filename,
		})
	}
}
