package handler

import (
	"net/http"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/auth/model"
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

func (h *AuthHandler) RegisterRoute(r *gin.RouterGroup) {
	authGroup := r.Group("auth")
	{
		authGroup.POST("login", h.Auth())
		authGroup.GET("logout", h.Logout())
	}
}

// Auth
// @Summary 用户登录/注册
// @Description 用户登录/注册接口
// @Tags Auth 管理
// @Accept json
// @Produce json
// @Param login body model.LoginReq true "用户验证信息"
// @Success 200 {object} response.Response "操作成功，返回成功消息"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /auth/login [post]
func (h *AuthHandler) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.LoginReq
		if err := c.ShouldBind(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error "+err.Error()))
			return
		}

		tokens, err := h.auth.LoginOrRegister(c.Request.Context(), req, c.Request.UserAgent())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}
		utils.SetTokens(c, tokens)

		response.Success(c)
	}
}

// Logout
// @Summary 用户登录
// @Description 用户登录接口
// @Tags Auth 管理
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "操作成功，返回成功消息"
// @Failure 400 {object} response.Response "参数错误(code=20001)"
// @Failure 500 {object} response.Response "系统错误(code=50000)"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.MustGet("uuid")
		err := h.auth.Logout(c.Request.Context(), uuid.(int64), c.Request.UserAgent())
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		c.SetSameSite(http.SameSiteLaxMode)
		c.SetCookie("cloud_refresh", "", -1, "/", "", false, false)

		response.Success(c)
	}
}
