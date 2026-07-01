package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/raoulellias/jh-weather-service/internal/api"
	"github.com/raoulellias/jh-weather-service/internal/nws"
	"github.com/raoulellias/jh-weather-service/internal/weather"
)

const serverAddr = ":8080"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	nwsClient := nws.NewClient(httpClient)
	weatherService := weather.NewService(nwsClient)

	server := &http.Server{
		Addr:         serverAddr,
		Handler:      api.NewRouter(logger, weatherService),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server listening", slog.String("addr", serverAddr))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-shutdownCtx.Done():
	case err := <-serverErr:
		logger.Error("server error", slog.Any("error", err))
		os.Exit(1)
	}

	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("shutting down server")
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("server shutdown error", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("server stopped")
}
