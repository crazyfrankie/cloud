package handler

import (
	"net/http"
	"strconv"

	"github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/pkg/response"
	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/file/service"
)

type FileHandler struct {
	upload *service.UploadService
}

func NewFileHandler(u *service.UploadService) *FileHandler {
	return &FileHandler{
		upload: u,
	}
}

func (h *FileHandler) RegisterRoute(r *gin.Engine) {
	fileGroup := r.Group("file")
	{
		fileGroup.POST("upload", h.Upload())
		fileGroup.POST("folder", h.CreateFolder())
		fileGroup.GET("list/:folderId", h.ListFolderContents())
	}
}

// Upload 上传文件元数据接口
func (h *FileHandler) Upload() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name     string `json:"name" binding:"required"`
			Size     int64  `json:"size" binding:"required"`
			FolderID int64  `json:"folderId"`
			URL      string `json:"url" binding:"required"`
			DeviceId string `json:"deviceId"`
		}
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		file := &dao.File{
			Name:           req.Name,
			Size:           req.Size,
			URL:            req.URL,
			FolderID:       req.FolderID,
			UID:            uid,
			DeviceId:       req.DeviceId,
			LastModifiedBy: strconv.FormatInt(uid, 10),
		}

		if file.Version == 0 {
			file.Version = 1
		}

		err := h.upload.Upload(c.Request.Context(), file)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, map[string]interface{}{
			"fileId": file.ID,
			"msg":    "upload success",
		})
	}
}

// CreateFolder 创建文件夹接口
func (h *FileHandler) CreateFolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			Name     string `json:"name" binding:"required"`
			ParentId int64  `json:"parentId"`
		}
		var req Req
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		folder := &dao.Folder{
			Name:     req.Name,
			ParentId: req.ParentId,
			UserId:   uid,
		}

		err := h.upload.CreateFolder(c.Request.Context(), folder)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, map[string]interface{}{
			"folderId": folder.ID,
			"path":     folder.Path,
			"msg":      "folder created successfully",
		})
	}
}

// ListFolderContents 列出文件夹内容接口
func (h *FileHandler) ListFolderContents() gin.HandlerFunc {
	return func(c *gin.Context) {
		folderIdStr := c.Param("folderId")
		folderId, err := strconv.ParseInt(folderIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid folder id"))
			return
		}

		uid := c.MustGet("uid").(int64)

		contents, err := h.upload.ListFolderContents(c.Request.Context(), uid, folderId)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, contents)
	}
}
