package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetWeather(router *gin.RouterGroup) {
	router.GET("/weather", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "not implemented"})
	})
}
