package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/cloud/internal/storage/model"
	"github.com/crazyfrankie/cloud/pkg/consts"
)

type StorageService struct {
	client *minio.Client
}

func NewStorageService(cli *minio.Client) *StorageService {
	return &StorageService{client: cli}
}

// Presign 生成带严格策略的预签名 URL
func (s *StorageService) Presign(ctx context.Context, uid int64, filename string, fileSize int64, typ string) (string, error) {
	var objectKey string
	var bucket string

	objectKey = fmt.Sprintf("%d/%s", uid, filename)
	if typ == "file" {
		bucket = consts.FileBucket
	} else {
		bucket = consts.UserBucket
	}

	expire := time.Hour * 1

	reqParams := make(url.Values)
	if fileSize > 0 {
		reqParams.Set("X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD")
	}

	preUrl, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expire)
	if err != nil {
		return "", err
	}

	return preUrl.String(), nil
}

// PresignForChunk 为分块上传生成预签名 URL
func (s *StorageService) PresignForChunk(ctx context.Context, chunkKey string) (string, error) {
	bucket := consts.FileBucket
	expire := time.Hour * 1

	preUrl, err := s.client.PresignedPutObject(ctx, bucket, chunkKey, expire)
	if err != nil {
		return "", fmt.Errorf("generate chunk presigned URL error: %w", err)
	}

	return preUrl.String(), nil
}

// PresignDownload 为小文件下载生成预签名 URL
func (s *StorageService) PresignDownload(ctx context.Context, objectKey string, filename string, expire time.Duration) (string, error) {
	bucket := consts.FileBucket

	// Set request parameters for download
	reqParams := make(url.Values)
	if filename != "" {
		reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}

	preUrl, err := s.client.PresignedGetObject(ctx, bucket, objectKey, expire, reqParams)
	if err != nil {
		return "", fmt.Errorf("generate download presigned URL error: %w", err)
	}

	return preUrl.String(), nil
}

// DeleteObject 删除对象
func (s *StorageService) DeleteObject(ctx context.Context, objectKey string) error {
	bucket := consts.FileBucket

	err := s.client.RemoveObject(ctx, bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete object %s error: %w", objectKey, err)
	}

	return nil
}

// GetObjectInfo 获取对象信息
func (s *StorageService) GetObjectInfo(ctx context.Context, objectKey string) (minio.ObjectInfo, error) {
	bucket := consts.FileBucket

	objInfo, err := s.client.StatObject(ctx, bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return minio.ObjectInfo{}, fmt.Errorf("get object info %s error: %w", objectKey, err)
	}

	return objInfo, nil
}

// ComposeObjectsWithETag 使用ETag验证合并多个对象为一个对象
func (s *StorageService) ComposeObjectsWithETag(ctx context.Context, chunks []model.ChunkInfo, finalObjectKey string) error {
	bucket := consts.FileBucket

	// 创建源对象列表，包含ETag验证
	sources := make([]minio.CopySrcOptions, len(chunks))
	for i, chunk := range chunks {
		sources[i] = minio.CopySrcOptions{
			Bucket: bucket,
			Object: chunk.Key,
		}

		// 如果提供了ETag，添加条件匹配
		if chunk.ETag != "" {
			// 清理ETag（移除引号）
			etag := strings.Trim(chunk.ETag, "\"")
			sources[i].MatchETag = etag
		}
	}

	// 目标对象配置
	dest := minio.CopyDestOptions{
		Bucket: bucket,
		Object: finalObjectKey,
	}

	// 合并对象
	_, err := s.client.ComposeObject(ctx, dest, sources...)
	if err != nil {
		return fmt.Errorf("compose objects with etag validation error: %w", err)
	}

	return nil
}

// GetUploadChunkObjects 获取断点续传所需的已上传的分块状态
func (s *StorageService) GetUploadChunkObjects(ctx context.Context, uid int64, uploadId string) []minio.ObjectInfo {
	bucket := consts.FileBucket

	res := make([]minio.ObjectInfo, 0, 100)

	info := s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    fmt.Sprintf("%d/chunks/%s", uid, uploadId),
		Recursive: true,
	})
	for i := range info {
		res = append(res, i)
	}

	return res
}

// ExtractObjectKey 从URL中提取对象键
func (s *StorageService) ExtractObjectKey(fileURL string) string {
	if fileURL == "" {
		return ""
	}

	// 解析URL
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return ""
	}

	// 提取路径，去掉开头的"/"
	path := parsedURL.Path
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}

	// MinIO URL格式通常是：/[bucket]/[object-key]
	// 我们需要去掉bucket部分，只保留object-key
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return path // 如果格式不符合预期，返回整个路径
	}

	// 跳过bucket部分，返回剩余的object key
	bucket := parts[0]
	if bucket == consts.FileBucket || bucket == consts.UserBucket {
		return parts[1]
	}

	// 如果不是预期的bucket，返回整个路径
	return path
}

// GetObject 获取对象内容，用于文件下载
func (s *StorageService) GetObject(ctx context.Context, objectKey string, opt minio.GetObjectOptions) (*minio.Object, error) {
	bucket := consts.FileBucket

	object, err := s.client.GetObject(ctx, bucket, objectKey, opt)
	if err != nil {
		return nil, fmt.Errorf("get object %s error: %w", objectKey, err)
	}

	return object, nil
}

// FGetObject 下载对象到本地文件，用于临时文件下载
func (s *StorageService) FGetObject(ctx context.Context, objectKey, filePath string) error {
	bucket := consts.FileBucket

	err := s.client.FGetObject(ctx, bucket, objectKey, filePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("download object %s to %s error: %w", objectKey, filePath, err)
	}

	return nil
}
