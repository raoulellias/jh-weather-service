package nws

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestGetPointsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/points/39.7456,-97.0892" {
			t.Fatalf("expected points path, got %q", r.URL.Path)
		}

		w.Header().Set("Content-Type", acceptHeader)
		fmt.Fprint(w, `{"properties":{"forecast":"https://api.weather.gov/gridpoints/TOP/32,81/forecast"}}`)
	}))
	t.Cleanup(server.Close)

	client := NewClientWithBaseURL(server.Client(), server.URL)

	points, err := client.GetPoints(context.Background(), 39.7456, -97.0892)
	if err != nil {
		t.Fatalf("GetPoints returned error: %v", err)
	}

	expectedForecastURL := "https://api.weather.gov/gridpoints/TOP/32,81/forecast"
	if points.Properties.Forecast != expectedForecastURL {
		t.Fatalf("expected forecast URL %q, got %q", expectedForecastURL, points.Properties.Forecast)
	}
}

func TestGetPointsNon2xxReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	t.Cleanup(server.Close)

	client := NewClientWithBaseURL(server.Client(), server.URL)

	_, err := client.GetPoints(context.Background(), 39.7456, -97.0892)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "get points") || !strings.Contains(err.Error(), "502 Bad Gateway") {
		t.Fatalf("expected wrapped non-2xx error, got %v", err)
	}
}

func TestGetPointsMalformedJSONReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"properties":`)
	}))
	t.Cleanup(server.Close)

	client := NewClientWithBaseURL(server.Client(), server.URL)

	_, err := client.GetPoints(context.Background(), 39.7456, -97.0892)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "get points") || !strings.Contains(err.Error(), "decode response") {
		t.Fatalf("expected wrapped decode error, got %v", err)
	}
}

func TestGetForecastSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/gridpoints/TOP/32,81/forecast" {
			t.Fatalf("expected forecast path, got %q", r.URL.Path)
		}

		w.Header().Set("Content-Type", acceptHeader)
		fmt.Fprint(w, `{"properties":{"periods":[{"number":1,"name":"Today","temperature":72,"temperatureUnit":"F","shortForecast":"Sunny","detailedForecast":"Sunny with light wind."}]}}`)
	}))
	t.Cleanup(server.Close)

	client := NewClientWithBaseURL(server.Client(), server.URL)

	forecast, err := client.GetForecast(context.Background(), server.URL+"/gridpoints/TOP/32,81/forecast")
	if err != nil {
		t.Fatalf("GetForecast returned error: %v", err)
	}

	if len(forecast.Properties.Periods) != 1 {
		t.Fatalf("expected 1 period, got %d", len(forecast.Properties.Periods))
	}
	if forecast.Properties.Periods[0].Name != "Today" {
		t.Fatalf("expected period name %q, got %q", "Today", forecast.Properties.Periods[0].Name)
	}
}

func TestGetForecastNon2xxReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)

	client := NewClientWithBaseURL(server.Client(), server.URL)

	_, err := client.GetForecast(context.Background(), server.URL+"/gridpoints/TOP/32,81/forecast")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "get forecast") || !strings.Contains(err.Error(), "503 Service Unavailable") {
		t.Fatalf("expected wrapped non-2xx error, got %v", err)
	}
}

func TestGetForecastMalformedJSONReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"properties":`)
	}))
	t.Cleanup(server.Close)

	client := NewClientWithBaseURL(server.Client(), server.URL)

	_, err := client.GetForecast(context.Background(), server.URL+"/gridpoints/TOP/32,81/forecast")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "get forecast") || !strings.Contains(err.Error(), "decode response") {
		t.Fatalf("expected wrapped decode error, got %v", err)
	}
}

func TestGetForecastForCoordinates(t *testing.T) {
	var server *httptest.Server
	calls := make([]string, 0, 2)
	headers := make([]http.Header, 0, 2)

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.URL.Path)
		headers = append(headers, r.Header.Clone())

		w.Header().Set("Content-Type", acceptHeader)
		switch r.URL.Path {
		case "/points/1.25,-2.5":
			fmt.Fprintf(w, `{"properties":{"forecast":"%s/gridpoints/TEST/1,2/forecast"}}`, server.URL)
		case "/gridpoints/TEST/1,2/forecast":
			fmt.Fprint(w, `{"properties":{"periods":[{"number":1,"name":"Tonight","temperature":42,"temperatureUnit":"F","shortForecast":"Mostly Clear","detailedForecast":"Mostly clear, with a low around 42."},{"number":2,"name":"Tomorrow","temperature":63,"temperatureUnit":"F","shortForecast":"Partly Sunny","detailedForecast":"Partly sunny, with a high near 63."}]}}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client := NewClientWithBaseURL(server.Client(), server.URL)

	forecast, err := client.GetForecastForCoordinates(context.Background(), 1.25, -2.5)
	if err != nil {
		t.Fatalf("GetForecastForCoordinates returned error: %v", err)
	}

	expectedCalls := []string{"/points/1.25,-2.5", "/gridpoints/TEST/1,2/forecast"}
	if !reflect.DeepEqual(calls, expectedCalls) {
		t.Fatalf("expected calls %#v, got %#v", expectedCalls, calls)
	}

	for _, header := range headers {
		if got := header.Get("Accept"); got != acceptHeader {
			t.Fatalf("expected Accept header %q, got %q", acceptHeader, got)
		}
		if got := header.Get("User-Agent"); got != userAgent {
			t.Fatalf("expected User-Agent header %q, got %q", userAgent, got)
		}
	}

	expectedPeriods := []ForecastPeriod{
		{
			Number:           1,
			Name:             "Tonight",
			Temperature:      42,
			TemperatureUnit:  "F",
			ShortForecast:    "Mostly Clear",
			DetailedForecast: "Mostly clear, with a low around 42.",
		},
		{
			Number:           2,
			Name:             "Tomorrow",
			Temperature:      63,
			TemperatureUnit:  "F",
			ShortForecast:    "Partly Sunny",
			DetailedForecast: "Partly sunny, with a high near 63.",
		},
	}
	if !reflect.DeepEqual(forecast.Properties.Periods, expectedPeriods) {
		t.Fatalf("expected periods %#v, got %#v", expectedPeriods, forecast.Properties.Periods)
	}
}
