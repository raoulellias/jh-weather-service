package nws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultBaseURL = "https://api.weather.gov"
	userAgent      = "jh-weather-service/0.1"
	acceptHeader   = "application/geo+json"
)

var (
	errMissingForecastURL = errors.New("missing forecast URL")
	errNilContext         = errors.New("nil context")
	errUnexpectedStatus   = errors.New("unexpected NWS response status")
)

type unexpectedStatusError struct {
	StatusCode int
	Status     string
}

func (e *unexpectedStatusError) Error() string {
	return fmt.Sprintf("%s: %s", errUnexpectedStatus, e.Status)
}

func (e *unexpectedStatusError) Unwrap() error {
	return errUnexpectedStatus
}

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(httpClient *http.Client) *Client {
	return NewClientWithBaseURL(httpClient, defaultBaseURL)
}

func NewClientWithBaseURL(httpClient *http.Client, baseURL string) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultBaseURL
	}

	return &Client{
		httpClient: httpClient,
		baseURL:    strings.TrimRight(baseURL, "/"),
	}
}

func (c *Client) GetPoints(ctx context.Context, latitude float64, longitude float64) (*PointsResponse, error) {
	endpoint := fmt.Sprintf("%s/points/%s,%s", c.baseURL, formatCoordinate(latitude), formatCoordinate(longitude))

	var points PointsResponse
	if err := c.getJSON(ctx, endpoint, &points); err != nil {
		return nil, fmt.Errorf("get points: %w", err)
	}

	return &points, nil
}

func (c *Client) GetForecast(ctx context.Context, forecastURL string) (*ForecastResponse, error) {
	if strings.TrimSpace(forecastURL) == "" {
		return nil, fmt.Errorf("get forecast: %w", errMissingForecastURL)
	}

	endpoint, err := c.resolveURL(forecastURL)
	if err != nil {
		return nil, fmt.Errorf("get forecast: %w", err)
	}

	var forecast ForecastResponse
	if err := c.getJSON(ctx, endpoint, &forecast); err != nil {
		return nil, fmt.Errorf("get forecast: %w", err)
	}

	return &forecast, nil
}

func (c *Client) GetForecastForCoordinates(ctx context.Context, latitude float64, longitude float64) (*ForecastResponse, error) {
	points, err := c.GetPoints(ctx, latitude, longitude)
	if err != nil {
		return nil, fmt.Errorf("get forecast for coordinates: %w", err)
	}

	forecast, err := c.GetForecast(ctx, points.Properties.ForecastURL)
	if err != nil {
		return nil, fmt.Errorf("get forecast for coordinates: %w", err)
	}

	return forecast, nil
}

func (c *Client) getJSON(ctx context.Context, endpoint string, target any) error {
	if ctx == nil {
		return fmt.Errorf("create request: %w", errNilContext)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", acceptHeader)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return &unexpectedStatusError{
			StatusCode: response.StatusCode,
			Status:     response.Status,
		}
	}

	if err := json.NewDecoder(response.Body).Decode(target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

func (c *Client) resolveURL(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse forecast URL: %w", err)
	}
	if parsedURL.IsAbs() {
		return parsedURL.String(), nil
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", fmt.Errorf("parse base URL: %w", err)
	}

	return baseURL.ResolveReference(parsedURL).String(), nil
}

func formatCoordinate(coordinate float64) string {
	return strconv.FormatFloat(coordinate, 'f', -1, 64)
}
