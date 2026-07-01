package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetWeather(t *testing.T) {
	previousMode := gin.Mode()
	gin.SetMode(gin.TestMode)
	t.Cleanup(func() {
		gin.SetMode(previousMode)
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	router := NewRouter(logger)

	request := httptest.NewRequest(http.MethodGet, "/weather", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	contentType := response.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected JSON response, got content type %q", contentType)
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON body: %v", err)
	}

	if body.Status != "not implemented" {
		t.Fatalf("expected status %q, got %q", "not implemented", body.Status)
	}
}
