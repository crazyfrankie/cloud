package handler

import (
	"net/http"
	"strconv"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/file/model"
	"github.com/crazyfrankie/cloud/internal/file/service"
	"github.com/crazyfrankie/cloud/pkg/response"
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
		fileGroup.POST("pre-upload-check", h.PreUploadCheck())
		fileGroup.POST("upload", h.Upload())
		fileGroup.POST("confirm-upload", h.ConfirmUpload())
		fileGroup.DELETE("/:fileId", h.DeleteFile())
		fileGroup.POST("batch-delete", h.BatchDeleteFiles())
		fileGroup.GET("verify/:fileId", h.VerifyFile())
		fileGroup.GET("stats", h.GetFileStats())
		fileGroup.PUT("/:fileId", h.UpdateFile())
		fileGroup.GET("versions/:fileId", h.GetFileVersions())
		fileGroup.POST("chunked-upload", h.InitChunkedUpload())
		fileGroup.POST("chunked-upload/:uploadId/chunk", h.UploadChunk())
		fileGroup.POST("chunked-upload/:uploadId/complete", h.CompleteChunkedUpload())
		fileGroup.DELETE("chunked-upload/:uploadId", h.AbortChunkedUpload())
	}
	folderGroup := r.Group("folder")
	{
		folderGroup.POST("", h.CreateFolder())
		folderGroup.GET("/:folderId", h.ListFolderContents())
		folderGroup.DELETE("/:folderId", h.DeleteFolder())
	}
}

// PreUploadCheck 预上传检查接口
// @Summary 预上传检查
// @Description 检查文件是否已存在，如果不存在则返回预签名URL
// @Tags File 管理
// @Accept json
// @Produce json
// @Param req body model.PreUploadCheckReq true "预上传检查请求"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/pre-upload-check [post]
func (h *FileHandler) PreUploadCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.PreUploadCheckReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		resp, err := h.upload.PreUploadCheck(c.Request.Context(), req, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, resp)
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

// ConfirmUpload 确认上传完成接口
// @Summary 确认上传完成
// @Description 上传完成后保存文件元数据
// @Tags File 管理
// @Accept json
// @Produce json
// @Param req body model.CreateFileReq true "文件元数据"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/confirm-upload [post]
func (h *FileHandler) ConfirmUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateFileReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		res, err := h.upload.ConfirmUpload(c.Request.Context(), req, uid)
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

// DeleteFile 删除文件
// @Summary 删除文件
// @Description 删除指定的文件
// @Tags File 管理
// @Accept json
// @Produce json
// @Param fileId path string true "文件ID"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/{fileId} [delete]
func (h *FileHandler) DeleteFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file id"))
			return
		}

		uid := c.MustGet("uid").(int64)

		err = h.upload.DeleteFile(c.Request.Context(), fileId, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
	}
}

// BatchDeleteFiles 批量删除文件
// @Summary 批量删除文件
// @Description 批量删除多个文件
// @Tags File 管理
// @Accept json
// @Produce json
// @Param req body model.BatchDeleteReq true "批量删除请求"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/batch-delete [post]
func (h *FileHandler) BatchDeleteFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.BatchDeleteReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		err := h.upload.BatchDelete(c.Request.Context(), req.FileIds, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
	}
}

// DeleteFolder 删除文件夹
// @Summary 删除文件夹
// @Description 删除指定的文件夹及其内容
// @Tags Folder 管理
// @Accept json
// @Produce json
// @Param folderId path string true "文件夹ID"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /folder/{folderId} [delete]
func (h *FileHandler) DeleteFolder() gin.HandlerFunc {
	return func(c *gin.Context) {
		folderIdStr := c.Param("folderId")
		folderId, err := strconv.ParseInt(folderIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid folder id"))
			return
		}

		uid := c.MustGet("uid").(int64)

		err = h.upload.DeleteFolder(c.Request.Context(), folderId, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
	}
}

// VerifyFile 验证文件
// @Summary 验证文件
// @Description 验证文件完整性
// @Tags File 管理
// @Accept json
// @Produce json
// @Param fileId path string true "文件ID"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/verify/{fileId} [get]
func (h *FileHandler) VerifyFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file id"))
			return
		}

		uid := c.MustGet("uid").(int64)

		isValid, err := h.upload.VerifyFile(c.Request.Context(), uid, fileId)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, map[string]bool{"valid": isValid})
	}
}

// GetFileStats 获取文件统计
// @Summary 获取文件统计
// @Description 获取用户的文件统计信息
// @Tags File 管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "操作成功"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/stats [get]
func (h *FileHandler) GetFileStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.MustGet("uid").(int64)

		stats, err := h.upload.GetUserFileStats(c.Request.Context(), uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, stats)
	}
}

// UpdateFile 更新文件
// @Summary 更新文件
// @Description 更新文件信息或替换文件内容
// @Tags File 管理
// @Accept json
// @Produce json
// @Param fileId path string true "文件ID"
// @Param req body model.UpdateFileReq true "更新文件请求"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/{fileId} [put]
func (h *FileHandler) UpdateFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file id"))
			return
		}

		var req model.UpdateFileReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		resp, err := h.upload.UpdateFile(c.Request.Context(), uid, fileId, req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, resp)
	}
}

// GetFileVersions 获取文件版本
// @Summary 获取文件版本
// @Description 获取文件的所有版本信息
// @Tags File 管理
// @Accept json
// @Produce json
// @Param fileId path string true "文件ID"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/versions/{fileId} [get]
func (h *FileHandler) GetFileVersions() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file id"))
			return
		}

		uid := c.MustGet("uid").(int64)

		versions, err := h.upload.GetFileVersions(c.Request.Context(), uid, fileId)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, versions)
	}
}

// InitChunkedUpload 初始化分块上传
// @Summary 初始化分块上传
// @Description 为大文件初始化分块上传会话
// @Tags File 管理
// @Accept json
// @Produce json
// @Param req body model.InitChunkedUploadReq true "分块上传初始化请求"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/chunked-upload [post]
func (h *FileHandler) InitChunkedUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.InitChunkedUploadReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		resp, err := h.upload.InitChunkedUpload(c.Request.Context(), uid, req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, resp)
	}
}

// UploadChunk 上传分块
// @Summary 上传分块
// @Description 上传单个文件分块
// @Tags File 管理
// @Accept json
// @Produce json
// @Param uploadId path string true "上传会话ID"
// @Param req body model.UploadChunkReq true "分块上传请求"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/chunked-upload/{uploadId}/chunk [post]
func (h *FileHandler) UploadChunk() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadId := c.Param("uploadId")

		var req model.UploadChunkReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		resp, err := h.upload.UploadChunk(c.Request.Context(), uid, uploadId, req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, resp)
	}
}

// CompleteChunkedUpload 完成分块上传
// @Summary 完成分块上传
// @Description 完成所有分块上传并合并文件
// @Tags File 管理
// @Accept json
// @Produce json
// @Param uploadId path string true "上传会话ID"
// @Param req body model.CompleteChunkedUploadReq true "完成分块上传请求"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/chunked-upload/{uploadId}/complete [post]
func (h *FileHandler) CompleteChunkedUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadId := c.Param("uploadId")

		var req model.CompleteChunkedUploadReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		resp, err := h.upload.CompleteChunkedUpload(c.Request.Context(), uid, uploadId, req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, resp)
	}
}

// AbortChunkedUpload 中止分块上传
// @Summary 中止分块上传
// @Description 中止分块上传会话并清理已上传的分块
// @Tags File 管理
// @Accept json
// @Produce json
// @Param uploadId path string true "上传会话ID"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /file/chunked-upload/{uploadId} [delete]
func (h *FileHandler) AbortChunkedUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadId := c.Param("uploadId")

		uid := c.MustGet("uid").(int64)

		err := h.upload.AbortChunkedUpload(c.Request.Context(), uid, uploadId)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
	}
}
