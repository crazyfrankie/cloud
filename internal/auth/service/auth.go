package service

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/crazyfrankie/cloud/internal/auth/model"
	"github.com/crazyfrankie/cloud/internal/user/service"
)

type AuthService struct {
	user  *service.UserService
	token *TokenService
}

func NewAuthService(user *service.UserService, token *TokenService) *AuthService {
	return &AuthService{
		user:  user,
		token: token,
	}
}

func (s *AuthService) LoginOrRegister(ctx context.Context, req model.LoginReq, ua string) ([]string, error) {
	var tokens []string
	var uid, uuid int64
	user, err := s.user.GetUserInfoByName(ctx, req.NickName)
	if err != nil {
		return tokens, err
	}
	// create user
	if user.ID == 0 {
		newUser, err := s.user.CreateUser(ctx, req.NickName, req.Password)
		if err != nil {
			return tokens, err
		}
		uid = newUser.ID
		uuid = newUser.UUID
	} else {
		err = bcrypt.CompareHashAndPassword(user.Password, []byte(req.Password))
		if err != nil {
			return tokens, err
		}
		uid = user.ID
		uuid = user.UUID
	}

	tokens, err = s.token.GenerateToken(uid, uuid, ua)

	return tokens, err
}

func (s *AuthService) Logout(ctx context.Context, uuid int64, ua string) error {
	return s.token.CleanToken(ctx, uuid, ua)
}
