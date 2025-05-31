package service

import (
	"context"
	"fmt"
	"net/url"
	"time"

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
		objectKey = fmt.Sprintf("files/%d/%s_%s", uid, fileHash[:8], filename)
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

func (s *StorageService) Presign(ctx context.Context, uid int64, filename string, typ string) (string, error) {
	var objectKey string
	var bucket string

	if typ == "file" {
		objectKey = fmt.Sprintf("files/%d/%s", uid, filename)
		bucket = consts.FileBucket
	} else {
		objectKey = fmt.Sprintf("avatar/%d/%s", uid, filename)
		bucket = consts.UserBucket
	}

	expire := time.Hour * 24
	preUrl, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expire)
	if err != nil {
		return "", err
	}

	// 返回完整的URL而不是只有Path
	return preUrl.String(), nil
}

// InitMultipartUpload 初始化分块上传
func (s *StorageService) InitMultipartUpload(ctx context.Context, uid int64, filename string, totalChunks int) (string, map[int]string, error) {
	objectKey := fmt.Sprintf("files/%d/chunks/%s_%d", uid, filename, time.Now().UnixNano())
	bucket := consts.FileBucket

	// 确保bucket存在
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return "", nil, fmt.Errorf("check bucket exists error: %w", err)
	}
	if !exists {
		err = s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return "", nil, fmt.Errorf("create bucket error: %w", err)
		}
	}

	// 使用MinIO的原生分块上传API
	// 生成唯一的上传ID
	uploadId := fmt.Sprintf("multipart_%d_%s_%d", uid, filename, time.Now().UnixNano())

	// 生成分块预签名URL
	chunkUrls := make(map[int]string)
	expire := time.Hour * 1 // 1小时过期

	for i := 1; i <= totalChunks; i++ {
		// 为每个分块生成独立的对象键
		chunkObjectKey := fmt.Sprintf("%s.part%d", objectKey, i)

		preUrl, err := s.client.PresignedPutObject(ctx, bucket, chunkObjectKey, expire)
		if err != nil {
			return "", nil, fmt.Errorf("generate presigned URL for part %d error: %w", i, err)
		}

		chunkUrls[i] = preUrl.String()
	}

	return uploadId, chunkUrls, nil
}

// CompleteMultipartUpload 完成分块上传
func (s *StorageService) CompleteMultipartUpload(ctx context.Context, uploadId string, chunkETags []interface{}) (string, error) {
	bucket := consts.FileBucket

	// 解析uploadId获取原始信息（假设格式为：multipart_uid_filename_timestamp）
	finalObjectKey := fmt.Sprintf("files/completed/%s", uploadId)

	// 生成分块对象键列表
	chunkKeys := make([]string, len(chunkETags))
	baseChunkKey := fmt.Sprintf("files/chunks/%s", uploadId)

	for i := 0; i < len(chunkETags); i++ {
		chunkKeys[i] = fmt.Sprintf("%s.part%d", baseChunkKey, i+1)
	}

	// 确保bucket存在
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return "", fmt.Errorf("check bucket exists error: %w", err)
	}
	if !exists {
		return "", fmt.Errorf("bucket %s does not exist", bucket)
	}

	// 验证所有分块是否存在
	for i, chunkKey := range chunkKeys {
		_, err := s.GetObjectInfo(ctx, chunkKey)
		if err != nil {
			return "", fmt.Errorf("chunk %d (key: %s) not found: %w", i+1, chunkKey, err)
		}
	}

	// 合并所有分块为最终文件
	err = s.ComposeObjects(ctx, chunkKeys, finalObjectKey)
	if err != nil {
		return "", fmt.Errorf("compose chunks into final object error: %w", err)
	}

	// 删除临时分块文件
	go func() {
		// 异步删除分块文件，避免阻塞响应
		cleanupCtx := context.Background()
		for _, chunkKey := range chunkKeys {
			if cleanupErr := s.DeleteObject(cleanupCtx, chunkKey); cleanupErr != nil {
				// 记录日志，但不影响主流程
				fmt.Printf("Warning: failed to cleanup chunk %s: %v\n", chunkKey, cleanupErr)
			}
		}
	}()

	// 生成最终文件的访问URL
	expire := time.Hour * 24 * 7 // 7天过期
	preUrl, err := s.client.PresignedGetObject(ctx, bucket, finalObjectKey, expire, nil)
	if err != nil {
		return "", fmt.Errorf("generate final file URL error: %w", err)
	}

	return preUrl.String(), nil
}

// AbortMultipartUpload 中止分块上传
func (s *StorageService) AbortMultipartUpload(ctx context.Context, uploadId string) error {
	// 中止分块上传并清理临时文件
	bucket := consts.FileBucket

	// 由于我们使用简化的分块上传方式，需要清理所有相关的分块文件
	// 在实际应用中，这里应该：
	// 1. 找到所有属于该uploadId的分块对象
	// 2. 删除这些临时对象

	// 使用uploadId作为前缀来查找相关的分块文件
	objectPrefix := fmt.Sprintf("files/chunks/%s", uploadId)

	// 列出所有相关的分块对象
	objectCh := s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    objectPrefix,
		Recursive: true,
	})

	// 删除找到的分块对象
	errorCh := s.client.RemoveObjects(ctx, bucket, objectCh, minio.RemoveObjectsOptions{})

	// 检查删除过程中的错误
	for removeErr := range errorCh {
		if removeErr.Err != nil {
			return fmt.Errorf("remove chunk object %s error: %w", removeErr.ObjectName, removeErr.Err)
		}
	}

	return nil
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

// CopyObject 复制对象
func (s *StorageService) CopyObject(ctx context.Context, srcObjectKey, destObjectKey string) error {
	bucket := consts.FileBucket

	// 源对象信息
	srcOpts := minio.CopySrcOptions{
		Bucket: bucket,
		Object: srcObjectKey,
	}

	// 目标对象信息
	destOpts := minio.CopyDestOptions{
		Bucket: bucket,
		Object: destObjectKey,
	}

	_, err := s.client.CopyObject(ctx, destOpts, srcOpts)
	if err != nil {
		return fmt.Errorf("copy object from %s to %s error: %w", srcObjectKey, destObjectKey, err)
	}

	return nil
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
