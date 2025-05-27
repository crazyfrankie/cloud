package response

import (
	"net/http"
	
	"github.com/crazyfrankie/gem/gerrors"
	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func Error(c *gin.Context, code int, err error) {
	if bizErr, ok := gerrors.FromBizStatusError(err); ok {
		c.JSON(code, Response{
			Code:    int(bizErr.BizStatusCode()),
			Message: bizErr.BizMessage(),
		})
		return
	}

	c.JSON(code, Response{
		Code:    50000,
		Message: err.Error(),
	})
}

func Success(c *gin.Context) {
	c.JSON(http.StatusOK, Response{
		Code:    20000,
		Message: "ok",
	})
}

func SuccessWithData(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    20000,
		Message: "ok",
		Data:    data,
	})
}
