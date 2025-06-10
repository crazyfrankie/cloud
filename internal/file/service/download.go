package service

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/cloud/internal/file/dao"
	"github.com/crazyfrankie/cloud/internal/file/model"
	"github.com/crazyfrankie/cloud/internal/file/service/ratelimit"
	"github.com/crazyfrankie/cloud/internal/storage"
	"github.com/crazyfrankie/cloud/pkg/consts"
)

type DownloadService struct {
	fileDao     *dao.FileDao
	storage     *storage.Service
	minioClient *minio.Client
}

func NewDownloadService(d *dao.FileDao, storage *storage.Service, minio *minio.Client) *DownloadService {
	return &DownloadService{
		fileDao:     d,
		storage:     storage,
		minioClient: minio,
	}
}

// DownloadSmallFiles 下载小文件（直接返回预签名URL或ZIP）
func (s *DownloadService) DownloadSmallFiles(ctx context.Context, uid int64, fileIDs []int64, zipName string) (*model.DownloadFileResp, error) {
	files, err := s.fileDao.FindByIds(ctx, uid, fileIDs)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("未找到指定文件")
	}

	// 验证都是小文件且不是文件夹
	var totalSize int64
	for _, file := range files {
		if file.IsDir {
			return nil, fmt.Errorf("不能下载文件夹: %s", file.Name)
		}
		totalSize += file.Size
	}

	// 单个文件：返回直接下载链接
	if len(files) == 1 {
		file := files[0]
		objectKey := s.storage.ExtractObjectKey(file.URL)
		presignedURL, err := s.storage.PresignDownload(ctx, objectKey, file.Name, 24*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("生成预签名URL失败: %w", err)
		}

		return &model.DownloadFileResp{
			Type:      "single",
			TotalSize: file.Size,
			DLink:     presignedURL,
		}, nil
	}

	// 多个文件：打包成ZIP
	zipData, err := s.createZipArchive(ctx, files)
	if err != nil {
		return nil, fmt.Errorf("创建ZIP文件失败: %w", err)
	}

	// 生成ZIP文件名
	if zipName == "" {
		zipName = fmt.Sprintf("download_%d_files.zip", len(files))
	}
	if !strings.HasSuffix(zipName, ".zip") {
		zipName += ".zip"
	}

	return &model.DownloadFileResp{
		Type:      "zip",
		TotalSize: totalSize,
		ZipName:   zipName,
		ZipData:   zipData,
	}, nil
}

// createZipArchive 创建ZIP压缩包
func (s *DownloadService) createZipArchive(ctx context.Context, files []dao.File) ([]byte, error) {
	// 创建内存缓冲区
	buf := &bytes.Buffer{}
	zipWriter := zip.NewWriter(buf)
	defer zipWriter.Close()

	for _, file := range files {
		// 从MinIO获取文件对象
		objectKey := s.storage.ExtractObjectKey(file.URL)
		object, err := s.storage.GetObject(ctx, objectKey, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("获取文件 %s 失败: %w", file.Name, err)
		}

		// 在ZIP中创建文件
		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			object.Close()
			return nil, fmt.Errorf("在ZIP中创建文件 %s 失败: %w", file.Name, err)
		}

		// 复制文件内容到ZIP
		_, err = io.Copy(zipFile, object)
		object.Close()
		if err != nil {
			return nil, fmt.Errorf("复制文件 %s 到ZIP失败: %w", file.Name, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("关闭ZIP写入器失败: %w", err)
	}

	return buf.Bytes(), nil
}

// StreamDownload 支持断点续传的大文件下载
// 支持 HTTP Range 请求，实现断点续传功能
func (s *DownloadService) StreamDownload(ctx context.Context, c *gin.Context, uid int64, fileID int64, vipTyp consts.VIPType) error {
	// 获取文件信息
	file, err := s.fileDao.FindByID(ctx, uid, fileID)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	if file.IsDir {
		return fmt.Errorf("不能下载文件夹")
	}

	objectKey := s.storage.ExtractObjectKey(file.URL)
	objInfo, err := s.storage.GetObjectInfo(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("获取对象信息失败: %w", err)
	}

	fileSize := objInfo.Size
	contentType := objInfo.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 解析Range请求头
	rangeHeader := c.GetHeader("Range")
	var rangeInfo *ratelimit.RangeInfo

	if rangeHeader != "" {
		rangeInfo, err = s.parseRangeHeader(rangeHeader, fileSize)
		if err != nil {
			c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
			c.Status(http.StatusRequestedRangeNotSatisfiable)
			return fmt.Errorf("无效的Range请求: %w", err)
		}
	} else {
		// 没有Range请求，下载整个文件
		rangeInfo = &ratelimit.RangeInfo{
			Start:  0,
			End:    fileSize - 1,
			Length: fileSize,
			Total:  fileSize,
		}
	}

	// 设置响应头
	s.setDownloadHeaders(c, file, objInfo, rangeInfo, rangeHeader != "")

	// 获取MinIO对象并流式传输
	opts := minio.GetObjectOptions{}
	if rangeHeader != "" {
		// 设置Range选项
		err = opts.SetRange(rangeInfo.Start, rangeInfo.End)
		if err != nil {
			return fmt.Errorf("设置Range选项失败: %w", err)
		}
	}

	object, err := s.storage.GetObject(ctx, objectKey, opts)
	if err != nil {
		return fmt.Errorf("获取对象失败: %w", err)
	}
	defer object.Close()

	// 流式传输数据
	_, err = s.streamDataWithProgress(c, object, fileSize, rangeInfo, vipTyp)
	if err != nil {
		return fmt.Errorf("流式传输失败: %w", err)
	}

	return nil
}

// parseRangeHeader 解析HTTP Range请求头
func (s *DownloadService) parseRangeHeader(rangeHeader string, fileSize int64) (*ratelimit.RangeInfo, error) {
	// 支持的格式：bytes=start-end, bytes=start-, bytes=-suffix
	re := regexp.MustCompile(`bytes=(\d*)-(\d*)`)
	matches := re.FindStringSubmatch(rangeHeader)

	if len(matches) != 3 {
		return nil, fmt.Errorf("无效的Range头格式")
	}

	var start, end int64
	var err error

	startStr := matches[1]
	endStr := matches[2]

	if startStr == "" && endStr == "" {
		return nil, fmt.Errorf("Range头不能为空")
	}

	if startStr == "" {
		// bytes=-suffix 格式，获取文件末尾的suffix字节
		suffix, err := strconv.ParseInt(endStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的suffix值")
		}
		start = fileSize - suffix
		if start < 0 {
			start = 0
		}
		end = fileSize - 1
	} else {
		start, err = strconv.ParseInt(startStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的start值")
		}

		if endStr == "" {
			// bytes=start- 格式，从start到文件末尾
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(endStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("无效的end值")
			}
		}
	}

	// 验证范围
	if start < 0 || end >= fileSize || start > end {
		return nil, fmt.Errorf("Range范围无效")
	}

	return &ratelimit.RangeInfo{
		Start:  start,
		End:    end,
		Length: end - start + 1,
		Total:  fileSize,
	}, nil
}

// setDownloadHeaders 设置下载响应头
func (s *DownloadService) setDownloadHeaders(c *gin.Context, file *dao.File, objInfo minio.ObjectInfo, rangeInfo *ratelimit.RangeInfo, isRange bool) {
	c.Header("Content-Type", objInfo.ContentType)
	c.Header("Content-Length", strconv.FormatInt(rangeInfo.Length, 10))
	c.Header("Accept-Ranges", "bytes")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Name))

	// ETag 和 Last-Modified
	if objInfo.ETag != "" {
		c.Header("ETag", objInfo.ETag)
	}
	if !objInfo.LastModified.IsZero() {
		c.Header("Last-Modified", objInfo.LastModified.UTC().Format(http.TimeFormat))
	}

	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	if isRange {
		// Range请求的特殊头
		c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", rangeInfo.Start, rangeInfo.End, rangeInfo.Total))
		c.Status(http.StatusPartialContent)
	} else {
		c.Status(http.StatusOK)
	}
}

// streamDataWithProgress 流式传输数据并支持进度监控
func (s *DownloadService) streamDataWithProgress(c *gin.Context, src io.Reader, fileSize int64, rangInfo *ratelimit.RangeInfo, vipTyp consts.VIPType) (int64, error) {
	// 创建速率限制配置
	var rateLimitConfig ratelimit.RateLimitConfig
	if vipTyp == consts.NVIP {
		rateLimitConfig = ratelimit.NewRateLimitConfig()
	} else {
		rateLimitConfig = ratelimit.NewVipRateLimitConfig(vipTyp)
	}

	// 创建速率限制写入器
	rateLimitedWriter := ratelimit.NewRateLimitedWriter(
		c.Request.Context(),
		c.Writer,
		rateLimitConfig,
		fileSize,
		rangInfo,
	)

	// 根据限流速率动态调整缓冲区大小
	bufferSize := calculateOptimalBufferSize(fileSize, rangInfo)
	buffer := make([]byte, bufferSize)
	var written int64

	for {
		// 检查上下文是否被取消
		select {
		case <-c.Request.Context().Done():
			return written, c.Request.Context().Err()
		default:
		}

		// 读取数据
		n, err := src.Read(buffer)
		if n > 0 {
			// 使用速率限制写入器写入响应
			w, writeErr := rateLimitedWriter.Write(buffer[:n])
			written += int64(w)

			if writeErr != nil {
				return written, writeErr
			}

			// 刷新缓冲区，确保数据立即发送
			if flusher, ok := c.Writer.(http.Flusher); ok {
				flusher.Flush()
			}
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			return written, err
		}
	}

	return written, nil
}

// GetDownloadProgress 获取下载进度信息
// 用于前端检查文件状态和断点续传信息
func (s *DownloadService) GetDownloadProgress(ctx context.Context, uid int64, fileID int64) (*model.DownloadProgressInfo, error) {
	// 获取文件信息
	file, err := s.fileDao.FindByID(ctx, uid, fileID)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	if file.IsDir {
		return nil, fmt.Errorf("不能下载文件夹")
	}

	objectKey := s.storage.ExtractObjectKey(file.URL)
	objInfo, err := s.storage.GetObjectInfo(ctx, objectKey)
	if err != nil {
		return nil, fmt.Errorf("获取对象信息失败: %w", err)
	}

	contentType := objInfo.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	progressInfo := &model.DownloadProgressInfo{
		FileID:       file.ID,
		FileName:     file.Name,
		ContentType:  contentType,
		TotalSize:    objInfo.Size,
		AcceptRanges: true,
		ETag:         objInfo.ETag,
	}

	if !objInfo.LastModified.IsZero() {
		progressInfo.LastModified = objInfo.LastModified.UTC().Format(http.TimeFormat)
	}

	return progressInfo, nil
}

// calculateOptimalBufferSize 根据文件大小和限流速率计算最优缓冲区大小
// 针对大文件下载场景优化，缓冲区大小在1MB-20MB区间动态调整
func calculateOptimalBufferSize(fileSize int64, rangeInfo *ratelimit.RangeInfo) int {
	const (
		MB = 1024 * 1024
		GB = 1024 * MB
	)

	// 计算实际请求大小
	requestSize := fileSize
	if rangeInfo != nil {
		requestSize = rangeInfo.End - rangeInfo.Start + 1
	}

	// 据文件大小和限流策略选择缓冲区大小
	switch {
	case requestSize <= 100*MB:
		return 1 * MB
	case requestSize <= 500*MB:
		return 4 * MB
	case requestSize <= 2*GB:
		return 8 * MB
	case requestSize <= 10*GB:
		return 12 * MB
	default:
		return 20 * MB
	}
}
