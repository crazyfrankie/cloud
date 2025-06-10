package handler

import (
	"net/http"
	"strconv"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/file/model"
	"github.com/crazyfrankie/cloud/internal/file/service"
	"github.com/crazyfrankie/cloud/pkg/consts"
	"github.com/crazyfrankie/cloud/pkg/response"
)

type FileHandler struct {
	file     *service.FileService
	upload   *service.UploadService
	download *service.DownloadService
}

func NewFileHandler(u *service.UploadService, f *service.FileService, d *service.DownloadService) *FileHandler {
	return &FileHandler{
		upload:   u,
		file:     f,
		download: d,
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
	}

	downloadGroup := fileGroup.Group("download")
	{
		downloadGroup.POST("", h.DownloadFile())
		downloadGroup.GET("/:fileId/stream", h.DownloadLargeFile())     // 流式下载大文件
		downloadGroup.GET("/:fileId/progress", h.GetDownloadProgress()) // 获取下载进度信息
		downloadGroup.GET("/queue")
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

		uuid := c.MustGet("uuid").(int64)

		contents, err := h.file.ListPathContents(c.Request.Context(), uuid, path)
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

		uuid := c.MustGet("uuid").(int64)

		err := h.file.CreateFile(c.Request.Context(), req, uuid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
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

		uuid := c.MustGet("uuid").(int64)

		fileInfo, err := h.file.GetFileById(c.Request.Context(), fileId, uuid)
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

		uuid := c.MustGet("uuid").(int64)

		err := h.file.DeleteByPath(c.Request.Context(), uuid, path)
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

		uuid := c.MustGet("uuid").(int64)

		err := h.file.BatchDeleteByPaths(c.Request.Context(), uuid, req.Paths)
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

		uuid := c.MustGet("uuid").(int64)

		fileInfo, err := h.file.UpdateFile(c.Request.Context(), fileId, uuid, req)
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

		uuid := c.MustGet("uuid").(int64)

		err := h.file.MovePath(c.Request.Context(), uuid, req.SourcePath, req.TargetPath)
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

		uuid := c.MustGet("uuid").(int64)

		err := h.file.CopyPath(c.Request.Context(), uuid, req.SourcePath, req.TargetPath)
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

		uuid := c.MustGet("uuid").(int64)

		resp, err := h.upload.PreUploadCheck(c.Request.Context(), req, uuid)
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

		uuid := c.MustGet("uuid").(int64)

		err := h.upload.ConfirmUpload(c.Request.Context(), req, uuid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
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

		uuid := c.MustGet("uuid").(int64)

		resp, err := h.upload.InitUpload(c.Request.Context(), uuid, req)
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

		uuid := c.MustGet("uuid").(int64)

		resp, err := h.upload.CompleteUpload(c.Request.Context(), uuid, uploadId, req)
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
		uuid := c.MustGet("uuid").(int64)

		stats, err := h.file.GetUserFileStats(c.Request.Context(), uuid)
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

		uuid := c.MustGet("uuid").(int64)

		fileInfo, err := h.file.GetFileById(c.Request.Context(), fileId, uuid)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		versions, err := h.file.GetFileVersionsByHash(c.Request.Context(), uuid, fileInfo.Hash)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, versions)
	}
}

// DownloadFile 下载文件接口
// @Summary 下载文件
// @Description 智能下载文件：小文件直接返回链接或ZIP
// @Tags 文件下载
// @Accept json
// @Produce json
// @Param req body model.DownloadFileReq true "下载请求"
// @Success 200 {object} response.Response{data=model.DownloadFileResp} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/download [post]
func (h *FileHandler) DownloadFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.DownloadFileReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error: "+err.Error()))
			return
		}

		uid := c.MustGet("uuid").(int64)

		// 首先尝试小文件下载
		smallFileResp, err := h.download.DownloadSmallFiles(c.Request.Context(), uid, req.FileIDs, req.ZipName)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(30000, err.Error()))
			return
		}
		// 如果是ZIP文件，需要特殊处理
		if smallFileResp.Type == "zip" && smallFileResp.ZipData != nil {
			// 设置ZIP文件下载的响应头
			c.Header("Content-Disposition", "attachment; filename=\""+smallFileResp.ZipName+"\"")
			c.Header("Content-Type", "application/zip")
			c.Header("Content-Length", strconv.Itoa(len(smallFileResp.ZipData)))

			// 直接返回ZIP数据
			c.Data(http.StatusOK, "application/zip", smallFileResp.ZipData)
			return
		}

		response.SuccessWithData(c, smallFileResp)
	}
}

// DownloadLargeFile 大文件下载接口
// @Summary 大文件流式下载
// @Description 支持 HTTP Range 请求的大文件下载，实现断点续传功能
// @Tags 文件下载
// @Accept json
// @Produce octet-stream
// @Param fileId path int true "文件ID"
// @Param downloadId query string false "下载ID，来自队列授权"
// @Param vip query string false "VIP等级"
// @Param Range header string false "HTTP Range请求头，格式: bytes=start-end"
// @Success 200 {file} binary "完整文件内容"
// @Success 206 {file} binary "部分文件内容（Range请求）"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 404 {object} response.Response "文件不存在(code=40004)"
// @Failure 416 {object} response.Response "Range请求无效(code=41600)"
// @Failure 429 {object} response.Response "并发限制(code=42900)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/download/{fileId}/stream [get]
func (h *FileHandler) DownloadLargeFile() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		vipTypStr := c.Query("vip")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file ID"))
			return
		}
		vipTyp, _ := strconv.Atoi(vipTypStr)

		uid := c.MustGet("uuid").(int64)

		err = h.download.StreamDownload(c.Request.Context(), c, uid, fileId, consts.VIPType(vipTyp))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, "下载失败: "+err.Error()))
			return
		}
	}
}

// GetDownloadProgress 获取下载进度信息
// @Summary 获取下载进度信息
// @Description 获取文件的下载状态和断点续传信息
// @Tags 文件下载
// @Accept json
// @Produce json
// @Param fileId path int true "文件ID"
// @Success 200 {object} response.Response{data=service.DownloadProgressInfo} "操作成功"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 404 {object} response.Response "文件不存在(code=40004)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /files/download/v2/{fileId}/progress [get]
func (h *FileHandler) GetDownloadProgress() gin.HandlerFunc {
	return func(c *gin.Context) {
		fileIdStr := c.Param("fileId")
		fileId, err := strconv.ParseInt(fileIdStr, 10, 64)
		if err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "invalid file ID"))
			return
		}

		uid := c.MustGet("uuid").(int64)

		progress, err := h.download.GetDownloadProgress(c.Request.Context(), uid, fileId)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, progress)
	}
}
