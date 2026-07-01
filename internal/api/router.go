package api

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/raoulellias/jh-weather-service/internal/api/middleware"
	"github.com/raoulellias/jh-weather-service/internal/model"
)

type ForecastService interface {
	GetForecast(ctx context.Context, coordinates model.Coordinates) (*model.Forecast, error)
}

func NewRouter(logger *slog.Logger, forecastService ForecastService) *gin.Engine {
	if logger == nil {
		logger = slog.Default()
	}

	router := gin.New()
	router.Use(middleware.RequestLogger(logger), gin.Recovery())

	GetWeather(router.Group(""), forecastService)

	return router
}
