package handler

import (
	"net/http"

	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"

	"github.com/crazyfrankie/cloud/internal/user/model"
	"github.com/crazyfrankie/cloud/internal/user/service"
	"github.com/crazyfrankie/cloud/pkg/response"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) RegisterRoute(r *gin.RouterGroup) {
	userGroup := r.Group("user")
	{
		userGroup.GET("", h.GetUserInfo())
		userGroup.PATCH("update/info", h.UpdateUserInfo())
		userGroup.PATCH("update/avatar", h.UpdateAvatar())
	}
}

func (h *UserHandler) GetUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.MustGet("uid")
		user, err := h.svc.GetUserInfo(c.Request.Context(), id.(int64))
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, user)
	}
}

func (h *UserHandler) UpdateUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req model.UpdateInfoReq
		if err := c.ShouldBind(&req); err != nil {
			response.Error(c, http.StatusBadRequest, gerrors.NewBizError(20001, "bind error "+err.Error()))
			return
		}

		id := c.MustGet("uid")
		updated, err := h.svc.UpdateUserInfo(c.Request.Context(), id.(int64), req.Nickname, req.Birthday)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.SuccessWithData(c, updated)
	}
}

func (h *UserHandler) UpdateAvatar() gin.HandlerFunc {
	return func(c *gin.Context) {
		objectKey := c.PostForm("object")

		id := c.MustGet("uid")
		err := h.svc.UpdateUserAvatar(c.Request.Context(), id.(int64), objectKey)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, gerrors.NewBizError(50000, err.Error()))
			return
		}

		response.Success(c)
	}
}
