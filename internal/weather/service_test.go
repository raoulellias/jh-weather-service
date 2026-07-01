package weather

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/raoulellias/jh-weather-service/internal/model"
	"github.com/raoulellias/jh-weather-service/internal/nws"
)

type fakeNWSProvider struct {
	forecast *nws.ForecastResponse
	err      error
}

func (f fakeNWSProvider) GetForecastForCoordinates(ctx context.Context, latitude float64, longitude float64) (*nws.ForecastResponse, error) {
	return f.forecast, f.err
}

func TestServiceUsesForecastPeriodNamedTodayWhenPresent(t *testing.T) {
	service := NewService(fakeNWSProvider{
		forecast: forecastResponse(
			nws.ForecastPeriod{
				Name:            "Tonight",
				ShortForecast:   "Mostly Clear",
				Temperature:     42,
				TemperatureUnit: "F",
			},
			nws.ForecastPeriod{
				Name:            "Today",
				ShortForecast:   "Partly Cloudy",
				Temperature:     72,
				TemperatureUnit: "F",
			},
		),
	})

	forecast, err := service.GetForecast(context.Background(), model.Coordinates{Latitude: 39.0997, Longitude: -94.5786})
	if err != nil {
		t.Fatalf("GetForecast returned error: %v", err)
	}
	if forecast.Forecast != "Partly Cloudy" {
		t.Fatalf("expected Today short forecast %q, got %q", "Partly Cloudy", forecast.Forecast)
	}
	if forecast.TemperatureType != model.TemperatureTypeModerate {
		t.Fatalf("expected temperature type %q, got %q", model.TemperatureTypeModerate, forecast.TemperatureType)
	}
}

func TestServiceFallsBackToFirstForecastPeriodWhenTodayIsNotPresent(t *testing.T) {
	service := NewService(fakeNWSProvider{
		forecast: forecastResponse(
			nws.ForecastPeriod{
				Name:            "Tonight",
				ShortForecast:   "Mostly Clear",
				Temperature:     42,
				TemperatureUnit: "F",
			},
			nws.ForecastPeriod{
				Name:            "Tomorrow",
				ShortForecast:   "Sunny Later",
				Temperature:     88,
				TemperatureUnit: "F",
			},
		),
	})

	forecast, err := service.GetForecast(context.Background(), model.Coordinates{Latitude: 39.0997, Longitude: -94.5786})
	if err != nil {
		t.Fatalf("GetForecast returned error: %v", err)
	}
	if forecast.Forecast != "Mostly Clear" {
		t.Fatalf("expected first short forecast %q, got %q", "Mostly Clear", forecast.Forecast)
	}
}

func TestServiceTemperatureClassifications(t *testing.T) {
	tests := []struct {
		name        string
		temperature int
		unit        string
		want        model.TemperatureType
	}{
		{
			name:        "cold",
			temperature: 50,
			unit:        "F",
			want:        model.TemperatureTypeCold,
		},
		{
			name:        "moderate",
			temperature: 72,
			unit:        "F",
			want:        model.TemperatureTypeModerate,
		},
		{
			name:        "hot",
			temperature: 85,
			unit:        "F",
			want:        model.TemperatureTypeHot,
		},
		{
			name:        "celsius conversion",
			temperature: 30,
			unit:        "C",
			want:        model.TemperatureTypeHot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(fakeNWSProvider{
				forecast: forecastResponse(nws.ForecastPeriod{
					ShortForecast:   "Forecast",
					Temperature:     tt.temperature,
					TemperatureUnit: tt.unit,
				}),
			})

			forecast, err := service.GetForecast(context.Background(), model.Coordinates{Latitude: 1.25, Longitude: -2.5})
			if err != nil {
				t.Fatalf("GetForecast returned error: %v", err)
			}
			if forecast.TemperatureType != tt.want {
				t.Fatalf("expected temperature type %q, got %q", tt.want, forecast.TemperatureType)
			}
		})
	}
}

func TestServiceWrapsProviderError(t *testing.T) {
	providerErr := errors.New("nws unavailable")
	service := NewService(fakeNWSProvider{err: providerErr})

	_, err := service.GetForecast(context.Background(), model.Coordinates{Latitude: 1.25, Longitude: -2.5})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected provider error to be wrapped, got %v", err)
	}
}

func TestServiceNoForecastPeriodsReturnsError(t *testing.T) {
	service := NewService(fakeNWSProvider{forecast: forecastResponse()})

	_, err := service.GetForecast(context.Background(), model.Coordinates{Latitude: 1.25, Longitude: -2.5})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "contains no periods") {
		t.Fatalf("expected no periods error, got %v", err)
	}
}

func TestServiceUnsupportedTemperatureUnitReturnsError(t *testing.T) {
	service := NewService(fakeNWSProvider{
		forecast: forecastResponse(nws.ForecastPeriod{
			ShortForecast:   "Forecast",
			Temperature:     72,
			TemperatureUnit: "K",
		}),
	})

	_, err := service.GetForecast(context.Background(), model.Coordinates{Latitude: 1.25, Longitude: -2.5})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `unsupported temperature unit "K"`) {
		t.Fatalf("expected unsupported unit error, got %v", err)
	}
}

func forecastResponse(periods ...nws.ForecastPeriod) *nws.ForecastResponse {
	return &nws.ForecastResponse{
		Properties: nws.ForecastData{
			Periods: periods,
		},
	}
}
