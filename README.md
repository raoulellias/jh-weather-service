# Jack Henry Weather Service

This project is a small Go weather API for the Jack Henry Weather Service assignment.

## Architecture

- Gin API layer in `internal/api` validates HTTP input and returns application JSON.
- Weather service layer in `internal/weather` interprets forecast data for application use.
- Thin NWS client layer in `internal/nws` handles raw calls to `api.weather.gov`.

## Build

```powershell
go build ./cmd/server
```

## Run

```powershell
go run ./cmd/server
```

The server listens on port `8080`.

## API

```text
GET /weather?lat=39.0997&lon=-94.5786
```

Successful responses use this shape:

```json
{
  "forecast": "Partly Cloudy",
  "temperatureType": "moderate"
}
```

## Shortcuts and tradeoffs

The service currently uses the first forecast period returned by NWS as the relevant forecast period. This is intentional for the assignment scope.

A production version would choose the period more carefully based on current time, location timezone, and forecast period metadata.
