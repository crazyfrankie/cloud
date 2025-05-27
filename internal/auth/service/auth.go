package service

import (
	"context"

	"golang.org/x/crypto/bcrypt"

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

func (s *AuthService) Login(ctx context.Context, name string, password string, ua string) ([]string, error) {
	var tokens []string
	id, pwd, err := s.user.GetUserInfoByName(ctx, name)
	if err != nil {
		return tokens, err
	}

	err = bcrypt.CompareHashAndPassword(pwd, []byte(password))
	if err != nil {
		return tokens, err
	}

	tokens, err = s.token.GenerateToken(id, ua)

	return tokens, err
}

func (s *AuthService) Logout(ctx context.Context, uid int64, ua string) error {
	return s.token.CleanToken(ctx, uid, ua)
}
