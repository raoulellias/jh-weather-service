package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/raoulellias/jh-weather-service/internal/model"
)

func GetWeather(router *gin.RouterGroup, forecastService ForecastService) {
	router.GET("/weather", func(c *gin.Context) {
		latRaw, ok := c.GetQuery("lat")
		if !ok || latRaw == "" {
			writeError(c, http.StatusBadRequest, "lat is required")
			return
		}

		lonRaw, ok := c.GetQuery("lon")
		if !ok || lonRaw == "" {
			writeError(c, http.StatusBadRequest, "lon is required")
			return
		}

		lat, err := strconv.ParseFloat(latRaw, 64)
		if err != nil {
			writeError(c, http.StatusBadRequest, "lat must be a number")
			return
		}
		if lat < -90 || lat > 90 {
			writeError(c, http.StatusBadRequest, "lat must be between -90 and 90")
			return
		}

		lon, err := strconv.ParseFloat(lonRaw, 64)
		if err != nil {
			writeError(c, http.StatusBadRequest, "lon must be a number")
			return
		}
		if lon < -180 || lon > 180 {
			writeError(c, http.StatusBadRequest, "lon must be between -180 and 180")
			return
		}

		if forecastService == nil {
			writeError(c, http.StatusBadGateway, "weather service unavailable")
			return
		}

		forecast, err := forecastService.GetForecast(c.Request.Context(), model.Coordinates{
			Latitude:  lat,
			Longitude: lon,
		})
		if err != nil {
			writeError(c, http.StatusBadGateway, "weather service unavailable")
			return
		}

		c.JSON(http.StatusOK, forecast)
	})
}

func writeError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}
