package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/crazyfrankie/cloud/internal/storage/model"

	"github.com/minio/minio-go/v7"

	"github.com/crazyfrankie/cloud/pkg/consts"
)

type StorageService struct {
	client *minio.Client
}

func NewStorageService(cli *minio.Client) *StorageService {
	return &StorageService{client: cli}
}

// PresignWithPolicy 生成带严格策略的预签名URL
func (s *StorageService) PresignWithPolicy(ctx context.Context, uid int64, filename string, fileSize int64, fileHash string, typ string) (string, string, error) {
	var objectKey string
	var bucket string

	if typ == "file" {
		// 使用哈希值作为文件名前缀，避免重复上传
		if fileHash != "" && len(fileHash) >= 8 {
			objectKey = fmt.Sprintf("files/%d/%s_%s", uid, fileHash[:8], filename)
		} else {
			// 如果没有哈希值或哈希值太短，使用时间戳
			objectKey = fmt.Sprintf("files/%d/%d_%s", uid, time.Now().Unix(), filename)
		}
		bucket = consts.FileBucket
	} else {
		objectKey = fmt.Sprintf("avatar/%d/%s", uid, filename)
		bucket = consts.UserBucket
	}

	// 设置较短的过期时间（1小时）
	expire := time.Hour * 1

	// 创建预签名URL，添加文件大小限制
	reqParams := make(url.Values)
	if fileSize > 0 {
		// 可以在这里添加更多的策略限制
		reqParams.Set("X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD")
	}

	preUrl, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expire)
	if err != nil {
		return "", "", err
	}

	return preUrl.String(), objectKey, nil
}

// PresignForChunk 为分块上传生成预签名URL，保持对象键不变
func (s *StorageService) PresignForChunk(ctx context.Context, chunkKey string) (string, error) {
	bucket := consts.FileBucket
	expire := time.Hour * 1

	preUrl, err := s.client.PresignedPutObject(ctx, bucket, chunkKey, expire)
	if err != nil {
		return "", fmt.Errorf("generate chunk presigned URL error: %w", err)
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

// ComposeObjects 合并多个对象为一个对象（真正的分块合并）
func (s *StorageService) ComposeObjects(ctx context.Context, chunkKeys []string, finalObjectKey string) error {
	bucket := consts.FileBucket

	// 创建源对象列表
	sources := make([]minio.CopySrcOptions, len(chunkKeys))
	for i, chunkKey := range chunkKeys {
		sources[i] = minio.CopySrcOptions{
			Bucket: bucket,
			Object: chunkKey,
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
		return fmt.Errorf("compose objects error: %w", err)
	}

	return nil
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
func (s *StorageService) GetObject(ctx context.Context, objectKey string) (*minio.Object, error) {
	bucket := consts.FileBucket

	object, err := s.client.GetObject(ctx, bucket, objectKey, minio.GetObjectOptions{})
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
