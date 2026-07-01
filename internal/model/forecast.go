package model

type Forecast struct {
	Forecast        string          `json:"forecast"`
	TemperatureType TemperatureType `json:"temperatureType"`
}
