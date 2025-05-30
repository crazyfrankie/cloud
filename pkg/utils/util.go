package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func SetCookies(c *gin.Context, tokens []string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("cloud_access", tokens[0], int(time.Hour*24), "/", "", false, false)
	c.SetCookie("cloud_refresh", tokens[1], int(time.Hour*24), "/", "", false, false)
}
