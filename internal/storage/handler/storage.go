package handler

import (
	"net/http"
	"strconv"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/storage/service"
	"github.com/crazyfrankie/cloud/pkg/response"
)

type StorageHandler struct {
	svc *service.StorageService
}

func NewStorageHandler(s *service.StorageService) *StorageHandler {
	return &StorageHandler{svc: s}
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
		filesize := c.PostForm("filesize")
		size, _ := strconv.ParseInt(filesize, 10, 64)
		filehash := c.PostForm("filehash")

		typ := c.Param("type")

		uid := c.MustGet("uid").(int64)

		url, err := h.svc.Presign(c.Request.Context(), uid, filename, size, filehash, typ)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(30000, err.Error()))
			return
		}

		response.SuccessWithData(c, map[string]any{
			"presignUrl": url,
		})
	}
}
