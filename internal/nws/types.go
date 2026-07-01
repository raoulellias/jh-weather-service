package nws

type PointsResponse struct {
	Properties PointsProperties `json:"properties"`
}

type PointsProperties struct {
	Forecast string `json:"forecast"`
}

type ForecastResponse struct {
	Properties ForecastProperties `json:"properties"`
}

type ForecastProperties struct {
	Periods []ForecastPeriod `json:"periods"`
}

type ForecastPeriod struct {
	Number           int    `json:"number"`
	Name             string `json:"name"`
	Temperature      int    `json:"temperature"`
	TemperatureUnit  string `json:"temperatureUnit"`
	ShortForecast    string `json:"shortForecast"`
	DetailedForecast string `json:"detailedForecast"`
}
