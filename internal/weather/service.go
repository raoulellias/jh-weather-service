package weather

import (
	"context"
	"errors"
	"fmt"

	"github.com/raoulellias/jh-weather-service/internal/model"
	"github.com/raoulellias/jh-weather-service/internal/nws"
)

type NWSProvider interface {
	GetForecastForCoordinates(ctx context.Context, latitude float64, longitude float64) (*nws.ForecastResponse, error)
}

type Service struct {
	provider NWSProvider
}

func NewService(provider NWSProvider) *Service {
	return &Service{provider: provider}
}

func (s *Service) GetForecast(ctx context.Context, coordinates model.Coordinates) (*model.Forecast, error) {
	if s.provider == nil {
		return nil, errors.New("nws provider is required")
	}

	forecast, err := s.provider.GetForecastForCoordinates(ctx, coordinates.Latitude, coordinates.Longitude)
	if err != nil {
		return nil, fmt.Errorf("fetch forecast for coordinates: %w", err)
	}
	if forecast == nil {
		return nil, errors.New("nws forecast response is nil")
	}
	if len(forecast.Properties.Periods) == 0 {
		return nil, errors.New("nws forecast response contains no periods")
	}

	period := forecast.Properties.Periods[0]
	for _, candidate := range forecast.Properties.Periods {
		if candidate.Name == "Today" {
			period = candidate
			break
		}
	}

	temperatureType, err := CharacterizeTemperature(period.Temperature, period.TemperatureUnit)
	if err != nil {
		return nil, fmt.Errorf("classify temperature: %w", err)
	}

	return &model.Forecast{
		Forecast:        period.ShortForecast,
		TemperatureType: temperatureType,
	}, nil
}
