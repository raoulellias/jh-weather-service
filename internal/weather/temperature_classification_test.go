package weather

import (
	"errors"
	"strings"
	"testing"

	"github.com/raoulellias/jh-weather-service/internal/model"
)

func TestCharacterizeTemperatureThresholds(t *testing.T) {
	tests := []struct {
		name        string
		temperature int
		unit        string
		want        model.TemperatureType
	}{
		{
			name:        "cold at 50 fahrenheit",
			temperature: 50,
			unit:        "F",
			want:        model.TemperatureTypeCold,
		},
		{
			name:        "moderate at 51 fahrenheit",
			temperature: 51,
			unit:        "F",
			want:        model.TemperatureTypeModerate,
		},
		{
			name:        "moderate at 84 fahrenheit",
			temperature: 84,
			unit:        "F",
			want:        model.TemperatureTypeModerate,
		},
		{
			name:        "hot at 85 fahrenheit",
			temperature: 85,
			unit:        "F",
			want:        model.TemperatureTypeHot,
		},
		{
			name:        "celsius converts before classification",
			temperature: 10,
			unit:        "C",
			want:        model.TemperatureTypeCold,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CharacterizeTemperature(tt.temperature, tt.unit)
			if err != nil {
				t.Fatalf("CharacterizeTemperature returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

func TestCharacterizeTemperatureUnsupportedUnitReturnsError(t *testing.T) {
	_, err := CharacterizeTemperature(72, "K")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errUnsupportedTemperatureUnit) {
		t.Fatalf("expected unsupported unit error, got %v", err)
	}
	if !strings.Contains(err.Error(), `"K"`) {
		t.Fatalf("expected unsupported unit in error, got %v", err)
	}
}
