package nws

type PointsResponse struct {
	Properties PointLinks `json:"properties"`
}

type PointLinks struct {
	ForecastURL string `json:"forecast"`
}

type ForecastResponse struct {
	Properties ForecastData `json:"properties"`
}

type ForecastData struct {
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
