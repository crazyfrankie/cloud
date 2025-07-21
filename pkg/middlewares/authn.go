package middlewares

import (
	"net/http"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/auth/service"
	"github.com/crazyfrankie/cloud/pkg/response"
	"github.com/crazyfrankie/cloud/pkg/utils"
)

type AuthnHandler struct {
	ignore map[string]struct{}
	token  *service.TokenService
}

func NewAuthnHandler(t *service.TokenService) *AuthnHandler {
	return &AuthnHandler{token: t, ignore: make(map[string]struct{})}
}

func (h *AuthnHandler) IgnorePath(path string) *AuthnHandler {
	h.ignore[path] = struct{}{}
	return h
}

func (h *AuthnHandler) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := h.ignore[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		access, err := h.token.GetAccessToken(c)
		if err == nil {
			if claims, err := h.token.ParseToken(access); err == nil {
				c.Set("uid", claims.UID)
				c.Set("uuid", claims.UUID)
				c.Next()
				return
			}
		}

		refresh, err := c.Cookie("cloud_refresh")
		if err != nil {
			response.Error(c, http.StatusUnauthorized, gerrors.NewBizError(40001, err.Error()))
			return
		}
		tokens, uuid, err := h.token.TryRefresh(refresh, c.Request.UserAgent())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}
		c.Set("uuid", uuid)
		utils.SetTokens(c, tokens)

		c.Next()
	}
}
