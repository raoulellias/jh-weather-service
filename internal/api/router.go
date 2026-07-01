package api

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func NewRouter(logger *slog.Logger) *gin.Engine {
	if logger == nil {
		logger = slog.Default()
	}

	router := gin.New()
	router.Use(requestLogger(logger), gin.Recovery())

	GetWeather(router.Group(""))

	return router
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		logger.InfoContext(
			c.Request.Context(),
			"request completed",
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", c.Writer.Status()),
			slog.Int("bytes", c.Writer.Size()),
			slog.Duration("duration", time.Since(start)),
		)
	}
}
