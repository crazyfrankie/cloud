package handler

import (
	"net/http"
	"strconv"

	"github.com/crazyfrankie/cloud/pkg/response"
	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/file/model"
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
	}
	folderGroup := r.Group("folder")
	{
		folderGroup.POST("", h.CreateFolder())
		folderGroup.GET("/:folderId", h.ListFolderContents())
	}
}

// Upload
// @Summary 上传文件元数据接口
// @Description 创建新的文件的元数据
// @Tags File 管理
// @Accept json
// @Produce json
// @Param file body model.CreateFileReq true "File 元数据"
// @Success 200 {object} response.Response "操作成功，返回成功消息"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/upload [post]
func (h *FileHandler) Upload() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateFileReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		res, err := h.upload.Upload(c.Request.Context(), req, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, res)
	}
}

// CreateFolder
// @Summary 创建文件夹接口
// @Description 创建新的文件夹
// @Tags Folder 管理
// @Accept json
// @Produce json
// @Param folder body model.CreateFolderReq true "Folder 元数据"
// @Success 200 {object} response.Response "操作成功，返回成功消息"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /folder [post]
func (h *FileHandler) CreateFolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateFolderReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		res, err := h.upload.CreateFolder(c.Request.Context(), req, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, res)
	}
}

// ListFolderContents
// @Summary 列出文件夹内容接口
// @Description 列出文件夹内容包括文件和文件夹
// @Tags Folder 管理
// @Accept json
// @Produce json
// @Param folderId param string true "Folder ID"
// @Success 200 {object} response.Response "操作成功，返回成功消息"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /folder/list/:folderId [post]
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
