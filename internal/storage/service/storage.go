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
	objectKey := fmt.Sprintf("avatar/%d/%s", uid, filename)
	expire := time.Hour * 24
	var bucket string
	if typ == "file" {
		bucket = consts.FileBucket
	} else {
		bucket = consts.UserBucket
	}
	preUrl, err := s.client.PresignedPutObject(ctx, bucket, objectKey, expire)
	if err != nil {
		return "", err
	}

	return preUrl.Path, nil
}
