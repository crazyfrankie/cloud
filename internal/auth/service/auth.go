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

func (s *AuthService) Login(ctx context.Context, req model.LoginReq, ua string) ([]string, error) {
	var tokens []string
	id, pwd, err := s.user.GetUserInfoByName(ctx, req.NickName)
	if err != nil {
		return tokens, err
	}

	err = bcrypt.CompareHashAndPassword(pwd, []byte(req.Password))
	if err != nil {
		return tokens, err
	}

	tokens, err = s.token.GenerateToken(id, ua)

	return tokens, err
}

func (s *AuthService) Logout(ctx context.Context, uid int64, ua string) error {
	return s.token.CleanToken(ctx, uid, ua)
}
