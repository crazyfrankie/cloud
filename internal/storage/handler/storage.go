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

func NewStorageHandler(svc *service.StorageService) *StorageHandler {
	return &StorageHandler{svc: svc}
}

func (h *StorageHandler) RegisterRoute(r *gin.Engine) {
	storageGroup := r.Group("storage")
	{
		storageGroup.POST("presign/:type", h.Presign())
		storageGroup.POST("presign-with-policy/:type", h.PresignWithPolicy()) // 新增严格策略的预签名接口
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

// PresignWithPolicy 生成带严格策略的预签名URL
func (h *StorageHandler) PresignWithPolicy() gin.HandlerFunc {
	return func(c *gin.Context) {
		filename := c.PostForm("filename")
		fileSizeStr := c.PostForm("fileSize")
		fileHash := c.PostForm("fileHash")
		typ := c.Param("type")

		if filename == "" || fileHash == "" {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "filename and fileHash are required"))
			return
		}

		var fileSize int64 = 0
		if fileSizeStr != "" {
			if size, err := strconv.ParseInt(fileSizeStr, 10, 64); err == nil {
				fileSize = size
			}
		}

		id := c.MustGet("uid")
		presignedUrl, objectKey, err := h.svc.PresignWithPolicy(c.Request.Context(), id.(int64), filename, fileSize, fileHash, typ)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, map[string]string{
			"presignedUrl": presignedUrl,
			"objectKey":    objectKey,
			"filename":     filename,
		})
	}
}
