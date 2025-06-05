package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SetTokens(c *gin.Context, tokens []string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.Header("x-access-token", tokens[0])
	c.SetCookie("cloud_refresh", tokens[1], int(time.Hour*24), "/", "", false, true)
}
