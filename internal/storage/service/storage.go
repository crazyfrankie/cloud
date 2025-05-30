package service

import (
	"context"
	"fmt"
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
