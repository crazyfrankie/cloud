package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/file/model"
	"github.com/crazyfrankie/cloud/internal/file/service"
	"github.com/crazyfrankie/cloud/pkg/response"
)

type FileHandler struct {
	file *service.FileService
}

func NewFileHandler(f *service.FileService) *FileHandler {
	return &FileHandler{
		file: f,
	}
}

func (h *FileHandler) RegisterRoute(r *gin.Engine) {
	fileGroup := r.Group("files")
	{
		fileGroup.GET("", h.ListContents())              // 列出目录内容
		fileGroup.POST("", h.CreateItem())               // 创建文件/文件夹 (手动创建文件夹或直接文件记录)
		fileGroup.GET("/:fileId", h.GetFileInfo())       // 获取文件详情
		fileGroup.DELETE("", h.DeleteByPath())           // 根据路径删除
		fileGroup.PUT("/:fileId", h.UpdateFileInfo())    // 更新文件信息
		fileGroup.POST("/batch-delete", h.BatchDelete()) // 批量删除
		fileGroup.POST("/move", h.MoveItem())            // 移动文件/文件夹
		fileGroup.POST("/copy", h.CopyItem())            // 复制文件/文件夹

		fileGroup.POST("/precreate", h.PreCreateCheck())       // 小文件上传：预检查和生成预签名URL
		fileGroup.POST("/create", h.ConfirmCreate())           // 小文件上传：确认上传完成并创建文件记录
		fileGroup.POST("/preupload", h.InitUpload())           // 大文件上传：初始化分块上传
		fileGroup.POST("/upload/complete", h.CompleteUpload()) // 大文件上传：合并分块并创建文件记录

		fileGroup.GET("/stats", h.GetUserFileStats())           // 获取用户文件统计
		fileGroup.GET("/:fileId/versions", h.GetFileVersions()) // 获取文件版本

		// 新增预览和下载接口
		fileGroup.GET("/:fileId/preview", h.PreviewFile())    // 统一文件预览接口
		fileGroup.GET("/:fileId/download", h.DownloadFile())  // 统一文件下载接口
		fileGroup.GET("/:fileId/text", h.PreviewTextFile())   // 文本文件预览接口
		fileGroup.GET("/:fileId/thumbnail", h.GetThumbnail()) // 获取文件缩略图
	}
}

// ListContents 列出目录内容
// @Summary 列出目录内容
// @Description 列出指定路径下的所有文件和文件夹
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param path query string false "目录路径，默认为根目录 /"
// @Success 200 {object} response.Response{data=model.ListContentsResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files [get]
func (h *FileHandler) ListContents() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Query("path")
		if path == "" {
			path = "/"
		}

		uid := c.MustGet("uid").(int64)

		contents, err := h.file.ListPathContents(c.Request.Context(), uid, path)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, contents)
	}
}

// CreateItem 创建文件/文件夹
// @Summary 创建文件或文件夹
// @Description 在指定路径创建文件或文件夹
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param req body model.CreateFileReq true "创建文件/文件夹请求"
// @Success 200 {object} response.Response{data=model.CreateItemResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files [post]
func (h *FileHandler) CreateItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateFileReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		fileResp, err := h.file.CreateFile(c.Request.Context(), req, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		// 转换为 CreateItemResp
		createResp := &model.CreateItemResp{
			ID:    fileResp.ID,
			Name:  fileResp.Name,
			Path:  fileResp.Path,
			IsDir: fileResp.IsDir,
			Ctime: fileResp.Ctime,
		}

		response.SuccessWithData(c, createResp)
	}
}

// GetFileInfo 获取文件/文件夹详情
// @Summary 获取文件或文件夹详情
// @Description 获取指定ID的文件或文件夹的详细信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param fileId path string true "文件/文件夹ID"
// @Success 200 {object} response.Response{data=model.FileResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/{fileId} [get]
func (h *FileHandler) GetFileInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file id"))
			return
		}

		uid := c.MustGet("uid").(int64)

		fileInfo, err := h.file.GetFileById(c.Request.Context(), fileId, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, fileInfo)
	}
}

// DeleteByPath 根据路径删除文件/文件夹
// @Summary 根据路径删除文件或文件夹
// @Description 删除指定路径的文件或文件夹
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param path query string true "文件/文件夹路径"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files [delete]
func (h *FileHandler) DeleteByPath() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Query("path")
		if path == "" {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "path is required"))
			return
		}

		uid := c.MustGet("uid").(int64)

		err := h.file.DeleteByPath(c.Request.Context(), uid, path)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
	}
}

// BatchDelete 批量删除文件/文件夹
// @Summary 批量删除文件或文件夹
// @Description 批量删除多个文件或文件夹
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param req body model.BatchDeleteReq true "批量删除请求"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/batch-delete [post]
func (h *FileHandler) BatchDelete() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.BatchDeleteReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		err := h.file.BatchDeleteByPaths(c.Request.Context(), uid, req.Paths)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
	}
}

// UpdateFileInfo 更新文件/文件夹信息
// @Summary 更新文件或文件夹信息
// @Description 更新指定ID的文件或文件夹的信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param fileId path string true "文件/文件夹ID"
// @Param req body model.UpdateFileReq true "更新请求"
// @Success 200 {object} response.Response{data=model.FileResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/{fileId} [put]
func (h *FileHandler) UpdateFileInfo() gin.HandlerFunc {
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

		fileInfo, err := h.file.UpdateFile(c.Request.Context(), fileId, uid, req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, fileInfo)
	}
}

// MoveItem 移动文件/文件夹
// @Summary 移动文件或文件夹
// @Description 将文件或文件夹移动到新位置
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param req body struct{SourcePath string "json:\"sourcePath\""; TargetPath string "json:\"targetPath\""} true "移动请求"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/move [post]
func (h *FileHandler) MoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SourcePath string `json:"sourcePath" binding:"required"`
			TargetPath string `json:"targetPath" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		err := h.file.MovePath(c.Request.Context(), uid, req.SourcePath, req.TargetPath)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
	}
}

// CopyItem 复制文件/文件夹
// @Summary 复制文件或文件夹
// @Description 将文件或文件夹复制到新位置
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param req body struct{SourcePath string "json:\"sourcePath\""; TargetPath string "json:\"targetPath\""} true "复制请求"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/copy [post]
func (h *FileHandler) CopyItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			SourcePath string `json:"sourcePath" binding:"required"`
			TargetPath string `json:"targetPath" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		err := h.file.CopyPath(c.Request.Context(), uid, req.SourcePath, req.TargetPath)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
	}
}

// PreCreateCheck 预上传检查
// @Summary 预上传检查
// @Description 检查文件是否已存在，避免重复上传
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param req body model.PreUploadCheckReq true "预上传检查请求"
// @Success 200 {object} response.Response{data=model.PreUploadCheckResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/pre-upload [post]
func (h *FileHandler) PreCreateCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.PreUploadCheckReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		resp, err := h.file.PreUploadCheck(c.Request.Context(), req, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, resp)
	}
}

// ConfirmCreate 确认上传完成
// @Summary 确认上传完成
// @Description 上传完成后创建文件记录
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param req body model.CreateFileReq true "文件信息"
// @Success 200 {object} response.Response{data=model.FileResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/uploads [post]
func (h *FileHandler) ConfirmCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.CreateFileReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		fileResp, err := h.file.ConfirmUpload(c.Request.Context(), req, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, fileResp)
	}
}

// InitUpload 初始化优化的分块上传
// @Summary 初始化优化的分块上传
// @Description 初始化分块上传流程，获取所有分块的上传URL
// @Tags 文件上传
// @Accept json
// @Produce json
// @Param req body model.InitUploadReq true "初始化请求"
// @Success 200 {object} response.Response{data=model.InitUploadResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/upload [post]
func (h *FileHandler) InitUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.InitUploadReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		resp, err := h.file.InitUpload(c.Request.Context(), uid, req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, resp)
	}
}

// CompleteUpload 完成优化的分块上传
// @Summary 完成优化的分块上传
// @Description 完成分块上传，合并所有分块为完整文件
// @Tags 文件上传
// @Accept json
// @Produce json
// @Query uploadId path string true "上传ID"
// @Param req body model.CompleteUploadReq true "完成请求"
// @Success 200 {object} response.Response{data=model.FileResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/upload/complete [post]
func (h *FileHandler) CompleteUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		uploadId := c.Query("uploadId")

		var req model.CompleteUploadReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uid").(int64)

		resp, err := h.file.CompleteUpload(c.Request.Context(), uid, uploadId, req)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, resp)
	}
}

// GetUserFileStats 获取用户文件统计
// @Summary 获取用户文件统计
// @Description 获取用户的文件和存储统计信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=model.FileStatsResp} "操作成功"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/stats [get]
func (h *FileHandler) GetUserFileStats() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.MustGet("uid").(int64)

		stats, err := h.file.GetUserFileStats(c.Request.Context(), uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, stats)
	}
}

// GetFileVersions 获取文件版本
// @Summary 获取文件版本
// @Description 获取文件的所有历史版本
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param fileId path string true "文件ID"
// @Success 200 {object} response.Response{data=[]model.FileResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/{fileId}/versions [get]
func (h *FileHandler) GetFileVersions() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file id"))
			return
		}

		uid := c.MustGet("uid").(int64)

		// 获取文件信息以获取哈希值
		fileInfo, err := h.file.GetFileById(c.Request.Context(), fileId, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		versions, err := h.file.GetFileVersionsByHash(c.Request.Context(), uid, fileInfo.Hash)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, versions)
	}
}

// PreviewFile 统一文件预览接口
// @Summary 统一文件预览接口
// @Description 根据文件类型智能决定预览方式：可预览文件返回预览页面，不可预览文件自动下载
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param fileId path int true "文件ID"
// @Success 200 {object} response.Response "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 404 {object} response.Response "文件不存在(code=40004)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/{fileId}/preview [get]
func (h *FileHandler) PreviewFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file ID"))
			return
		}

		uid := c.MustGet("uid").(int64)

		// 获取文件信息
		fileInfo, err := h.file.GetFileById(c.Request.Context(), fileId, uid)
		if err != nil {
			response.Error(c, http.StatusNotFound, gerrors.NewBizError(40004, "file not found"))
			return
		}

		// 判断文件操作类型
		actionInfo := h.file.GetFileActionInfo(fileId, fileInfo.Name, fileInfo.URL)

		switch actionInfo.Action {
		case "preview":
			// 可预览文件，重定向到MinIO原生URL
			c.Redirect(http.StatusFound, fileInfo.URL)
		case "text":
			// 文本文件，返回文本内容预览
			c.Redirect(http.StatusFound, fmt.Sprintf("/api/files/%d/text", fileId))
		case "download":
			// 不可预览文件，直接下载
			c.Redirect(http.StatusFound, fmt.Sprintf("/api/files/%d/download", fileId))
		default:
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, "unknown file action"))
		}
	}
}

// DownloadFile 统一文件下载接口
// @Summary 统一文件下载接口
// @Description 提供统一的文件下载功能，设置正确的文件名和Content-Disposition
// @Tags 文件管理
// @Accept json
// @Produce octet-stream
// @Param fileId path int true "文件ID"
// @Success 200 {file} binary "文件内容"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 404 {object} response.Response "文件不存在(code=40004)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/{fileId}/download [get]
func (h *FileHandler) DownloadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file ID"))
			return
		}

		uid := c.MustGet("uid").(int64)

		// 获取文件信息
		fileInfo, err := h.file.GetFileById(c.Request.Context(), fileId, uid)
		if err != nil {
			response.Error(c, http.StatusNotFound, gerrors.NewBizError(40004, "file not found"))
			return
		}

		// 设置下载相关的响应头 - 强制下载
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileInfo.Name))
		c.Header("Content-Type", "application/octet-stream")

		// 从URL中提取对象键
		objectKey := h.file.GetObjectKeyFromURL(fileInfo.URL)
		if objectKey == "" {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, "invalid file URL"))
			return
		}

		// 获取文件下载信息
		downloadInfo, err := h.file.DownloadFile(c.Request.Context(), fileId, uid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}
		defer downloadInfo.Object.Close()

		// 直接将文件流传输给客户端，强制作为下载处理
		c.Header("Content-Length", fmt.Sprintf("%d", downloadInfo.Size))
		c.DataFromReader(http.StatusOK, downloadInfo.Size, "application/octet-stream", downloadInfo.Object, nil)
	}
}

// PreviewTextFile 文本文件预览接口
// @Summary 文本文件预览接口
// @Description 提供文本文件的在线预览功能
// @Tags 文件管理
// @Accept json
// @Produce plain
// @Param fileId path int true "文件ID"
// @Success 200 {string} string "文本内容"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 404 {object} response.Response "文件不存在(code=40004)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/{fileId}/text [get]
func (h *FileHandler) PreviewTextFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file ID"))
			return
		}

		uid := c.MustGet("uid").(int64)

		// 获取文件信息
		fileInfo, err := h.file.GetFileById(c.Request.Context(), fileId, uid)
		if err != nil {
			response.Error(c, http.StatusNotFound, gerrors.NewBizError(40004, "file not found"))
			return
		}

		// 验证是否为文本文件
		actionInfo := h.file.GetFileActionInfo(fileId, fileInfo.Name, fileInfo.URL)
		if actionInfo.Action != "text" {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "file is not a text file"))
			return
		}

		// 重定向到前端的文本预览页面
		textPreviewURL := fmt.Sprintf("/text-preview/%d", fileId)
		c.Redirect(http.StatusFound, textPreviewURL)
	}
}

// GetThumbnail 获取文件缩略图
// @Summary 获取文件缩略图
// @Description 获取支持缩略图的文件的缩略图
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param fileId path int true "文件ID"
// @Success 200 {object} response.Response{data=map[string]string} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 404 {object} response.Response "文件不存在(code=40004)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/{fileId}/thumbnail [get]
func (h *FileHandler) GetThumbnail() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file ID"))
			return
		}

		uid := c.MustGet("uid").(int64)

		// 获取文件信息
		fileInfo, err := h.file.GetFileById(c.Request.Context(), fileId, uid)
		if err != nil {
			response.Error(c, http.StatusNotFound, gerrors.NewBizError(40004, "file not found"))
			return
		}

		// 检查是否支持缩略图
		actionInfo := h.file.GetFileActionInfo(fileId, fileInfo.Name, fileInfo.URL)
		if !actionInfo.HasThumbnail {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "file does not support thumbnail"))
			return
		}

		// 返回缩略图URL（这里可以实现实际的缩略图生成逻辑）
		response.SuccessWithData(c, map[string]string{
			"thumbnailUrl": fileInfo.URL, // 临时返回原始URL，后续可以实现真正的缩略图服务
		})
	}
}
