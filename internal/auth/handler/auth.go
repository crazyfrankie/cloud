package handler

import (
	"net/http"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/auth/service"
	"github.com/crazyfrankie/cloud/pkg/response"
	"github.com/crazyfrankie/cloud/pkg/utils"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{
		auth: auth,
	}
}

func (h *AuthHandler) RegisterRoute(r *gin.Engine) {
	authGroup := r.Group("auth")
	{
		authGroup.POST("login", h.Login())
		authGroup.GET("logout", h.Logout())
	}
}

func (h *AuthHandler) Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		type Req struct {
			NickName string `json:"nickname"`
			Password string `yaml:"password"`
		}
		var req Req
		if err := c.ShouldBind(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error "+err.Error()))
			return
		}

		tokens, err := h.auth.Login(c.Request.Context(), req.NickName, req.Password, c.Request.UserAgent())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}
		utils.SetCookies(c, tokens)

		response.Success(c)
	}
}

func (h *AuthHandler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.MustGet("uid")
		err := h.auth.Logout(c.Request.Context(), id.(int64), c.Request.UserAgent())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		// 清除cookies
		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("cloud_access", "", -1, "/", "", false, false)
		c.SetCookie("cloud_refresh", "", -1, "/", "", false, false)

		response.Success(c)
	}
}
