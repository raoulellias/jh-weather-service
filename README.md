# Jack Henry Weather Service

This project is a small Go weather API for the Jack Henry Weather Service assignment.

## Architecture

The service is intentionally split into three small layers:

- `internal/api` — Gin API layer that validates HTTP input and returns application JSON.
- `internal/weather` — application service layer that interprets forecast data.
- `internal/nws` — thin National Weather Service client that handles raw calls to `api.weather.gov`.

```text
GET /weather?lat=39.0997&lon=-94.5786
        |
        v
internal/api
        |
        v
internal/weather
        |
        v
internal/nws
        |
        v
api.weather.gov
```

## Build

```powershell
go build ./cmd/server
```

## Run

```powershell
go run ./cmd/server
```

The server listens on port `8080`.

## Test

```powershell
go test ./...
```

## API

### Get weather

```text
GET /weather?lat=39.0997&lon=-94.5786
```

### Successful response

```json
{
  "forecast": "Partly Cloudy",
  "temperatureType": "moderate"
}
```

### Error responses

The API returns simple client-facing error responses for invalid requests or upstream weather-service failures. Internal wrapped errors are not exposed in response bodies.

## Shortcuts and trade-offs

This implementation is intentionally scoped for the assignment rather than built as a production weather platform. The goal is to keep the service small, readable, testable, and explicit about where production concerns would normally be added.

### Forecast period selection

The service prefers the NWS forecast period named `Today` when one is present. If NWS does not return a `Today` period, the service falls back to the first forecast period in the response.

This is a deliberate shortcut. A production version would select the relevant forecast period using current time, forecast period start/end metadata, and the timezone of the requested location.

### Minimal configuration

This exercise version does not use config files or a config library. Runtime configuration is intentionally minimal.

In a production service, configuration would likely be pulled from environment variables or a validated configuration layer for values such as:

- server port
- NWS base URL
- HTTP client timeout
- User-Agent value
- logging mode
- cache duration

### No caching

Each request calls the National Weather Service API.

A production version would likely cache responses for a short period, probably by NWS gridpoint or rounded coordinates, to reduce upstream traffic and improve latency.

### No retries or circuit breaker

Upstream NWS failures are returned as service failures rather than retried.

A production version would likely add bounded retries, backoff, timeout policy, and possibly a circuit breaker depending on traffic and reliability requirements.

### Simple temperature characterization

Temperature is characterized using fixed Fahrenheit thresholds:

- `cold`: 50°F or below
- `moderate`: 51°F through 84°F
- `hot`: 85°F or above

A production version might make these thresholds configurable or adjust them by region, season, or product requirements.

### Partial NWS response models

The NWS DTOs only model the fields required for this exercise.

A broader integration would likely preserve more response metadata for diagnostics, period selection, caching, and troubleshooting.

### Minimal observability

The service uses structured request logging, but does not expose metrics or tracing.

A production version would add request metrics, upstream latency metrics, error counters, and trace correlation.

### No authentication or rate limiting

The endpoint is unauthenticated and does not perform rate limiting.

For a public production API, rate limiting and possibly authentication would be added depending on the intended consumers.
