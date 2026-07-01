package weather

import (
	"fmt"
	"strings"

	"github.com/raoulellias/jh-weather-service/internal/model"
)

func CharacterizeTemperature(temperature int, unit string) (model.TemperatureType, error) {
	fahrenheit, err := toFahrenheit(temperature, unit)
	if err != nil {
		return "", err
	}

	switch {
	case fahrenheit <= 50:
		return model.TemperatureTypeCold, nil
	case fahrenheit >= 85:
		return model.TemperatureTypeHot, nil
	default:
		return model.TemperatureTypeModerate, nil
	}
}

func toFahrenheit(temperature int, unit string) (float64, error) {
	switch strings.ToUpper(strings.TrimSpace(unit)) {
	case "F", "FAHRENHEIT":
		return float64(temperature), nil
	case "C", "CELSIUS":
		return float64(temperature)*9/5 + 32, nil
	default:
		return 0, fmt.Errorf("unsupported temperature unit %q", unit)
	}
}
