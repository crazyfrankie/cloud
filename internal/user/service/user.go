package service

import (
	"context"
	"time"

	"github.com/crazyfrankie/snow-flake"
	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/bcrypt"

	"github.com/crazyfrankie/cloud/internal/user/dao"
	"github.com/crazyfrankie/cloud/internal/user/model"
	"github.com/crazyfrankie/cloud/pkg/consts"
)

type UserService struct {
	dao   *dao.UserDao
	node  *snowflake.Node
	minio *minio.Client
}

func NewUserService(d *dao.UserDao, node *snowflake.Node, minio *minio.Client) *UserService {
	return &UserService{dao: d, node: node, minio: minio}
}

func (s *UserService) CreateUser(ctx context.Context, name string, password string) (*dao.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	uuid := s.node.GenerateCode()
	user := &dao.User{
		UUID:     uuid,
		Nickname: name,
		Avatar:   consts.DefaultAvatar,
		Password: hash,
	}
	if err := s.dao.Insert(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserInfo(ctx context.Context, uid int64) (model.UserResp, error) {
	user, err := s.dao.FindByID(ctx, uid)
	if err != nil {
		return model.UserResp{}, err
	}

	var birthday string
	if user.Birthday.Valid {
		birthday = user.Birthday.Time.Format(time.DateTime)
	}

	return model.UserResp{
		ID:       user.ID,
		NickName: user.Nickname,
		Avatar:   user.Avatar,
		Birthday: birthday,
		Ctime:    user.Ctime,
		Utime:    user.Utime,
	}, nil
}

func (s *UserService) GetUserInfoByName(ctx context.Context, name string) (dao.User, error) {
	user, err := s.dao.FindByName(ctx, name)
	if err != nil {
		return dao.User{}, err
	}

	return user, nil
}

func (s *UserService) UpdateUserInfo(ctx context.Context, id int64, nickname string, birthday string) (model.UserResp, error) {
	update := make(map[string]any)
	if nickname != "" {
		update["nickname"] = nickname
	}
	if birthday != "" {
		var err error
		update["birthday"], err = time.Parse(time.DateOnly, birthday)
		if err != nil {
			return model.UserResp{}, err
		}
	}

	newUser, err := s.dao.UpdateUser(ctx, id, update)
	if err != nil {
		return model.UserResp{}, err
	}

	var newBirth string
	if newUser.Birthday.Valid {
		newBirth = newUser.Birthday.Time.Format(time.DateOnly)
	}

	return model.UserResp{
		ID:       newUser.ID,
		NickName: newUser.Nickname,
		Avatar:   newUser.Avatar,
		Birthday: newBirth,
		Utime:    newUser.Utime,
	}, nil
}

func (s *UserService) UpdateUserAvatar(ctx context.Context, uid int64, objectKey string) error {
	return s.dao.UpdateAvatar(ctx, uid, objectKey)
}
