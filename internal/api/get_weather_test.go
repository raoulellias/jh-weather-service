package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/raoulellias/jh-weather-service/internal/model"
)

type fakeForecastService struct {
	forecast    *model.Forecast
	err         error
	called      bool
	coordinates model.Coordinates
}

func (f *fakeForecastService) GetForecast(ctx context.Context, coordinates model.Coordinates) (*model.Forecast, error) {
	f.called = true
	f.coordinates = coordinates
	return f.forecast, f.err
}

func TestGetWeatherMissingLatReturns400(t *testing.T) {
	service := &fakeForecastService{}
	router := newTestRouter(t, service)

	response := performWeatherRequest(router, "/weather?lon=-94.5786")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.called {
		t.Fatal("expected service not to be called")
	}
}

func TestGetWeatherMissingLonReturns400(t *testing.T) {
	service := &fakeForecastService{}
	router := newTestRouter(t, service)

	response := performWeatherRequest(router, "/weather?lat=39.0997")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.called {
		t.Fatal("expected service not to be called")
	}
}

func TestGetWeatherInvalidLatReturns400(t *testing.T) {
	service := &fakeForecastService{}
	router := newTestRouter(t, service)

	response := performWeatherRequest(router, "/weather?lat=not-a-number&lon=-94.5786")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.called {
		t.Fatal("expected service not to be called")
	}
}

func TestGetWeatherInvalidLonReturns400(t *testing.T) {
	service := &fakeForecastService{}
	router := newTestRouter(t, service)

	response := performWeatherRequest(router, "/weather?lat=39.0997&lon=not-a-number")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
	}
	if service.called {
		t.Fatal("expected service not to be called")
	}
}

func TestGetWeatherOutOfRangeCoordinatesReturn400(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "lat less than minimum",
			path: "/weather?lat=-90.1&lon=-94.5786",
		},
		{
			name: "lat greater than maximum",
			path: "/weather?lat=90.1&lon=-94.5786",
		},
		{
			name: "lon less than minimum",
			path: "/weather?lat=39.0997&lon=-180.1",
		},
		{
			name: "lon greater than maximum",
			path: "/weather?lat=39.0997&lon=180.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeForecastService{}
			router := newTestRouter(t, service)

			response := performWeatherRequest(router, tt.path)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.Code)
			}
			if service.called {
				t.Fatal("expected service not to be called")
			}
		})
	}
}

func TestGetWeatherSuccessReturnsForecast(t *testing.T) {
	service := &fakeForecastService{
		forecast: &model.Forecast{
			Forecast:        "Partly Cloudy",
			TemperatureType: model.TemperatureTypeModerate,
		},
	}
	router := newTestRouter(t, service)

	response := performWeatherRequest(router, "/weather?lat=39.0997&lon=-94.5786")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if !service.called {
		t.Fatal("expected service to be called")
	}
	if service.coordinates.Latitude != 39.0997 {
		t.Fatalf("expected latitude %v, got %v", 39.0997, service.coordinates.Latitude)
	}
	if service.coordinates.Longitude != -94.5786 {
		t.Fatalf("expected longitude %v, got %v", -94.5786, service.coordinates.Longitude)
	}

	var body model.Forecast
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON body: %v", err)
	}
	if body.Forecast != "Partly Cloudy" {
		t.Fatalf("expected forecast %q, got %q", "Partly Cloudy", body.Forecast)
	}
	if body.TemperatureType != model.TemperatureTypeModerate {
		t.Fatalf("expected temperature type %q, got %q", model.TemperatureTypeModerate, body.TemperatureType)
	}
}

func TestGetWeatherServiceErrorReturns502(t *testing.T) {
	serviceErr := errors.New("fetch forecast for coordinates: upstream failed")
	service := &fakeForecastService{err: serviceErr}
	router := newTestRouter(t, service)

	response := performWeatherRequest(router, "/weather?lat=39.0997&lon=-94.5786")

	if response.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, response.Code)
	}
	if !service.called {
		t.Fatal("expected service to be called")
	}

	rawBody := response.Body.String()
	if strings.Contains(rawBody, serviceErr.Error()) || strings.Contains(rawBody, "upstream failed") {
		t.Fatalf("expected response body not to expose service error, got %s", rawBody)
	}

	var body map[string]string
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("expected JSON error body: %v", err)
	}
	if len(body) != 1 || body["error"] != "weather service unavailable" {
		t.Fatalf("expected public error body, got %#v", body)
	}
}

func newTestRouter(t *testing.T, forecastService ForecastService) *gin.Engine {
	t.Helper()

	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		gin.SetMode(previousMode)
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewRouter(logger, forecastService)
}

func performWeatherRequest(router http.Handler, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodGet, path, nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	return response
}
