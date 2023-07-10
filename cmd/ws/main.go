package main

import (
	"go-outpost/internal/config"
	"go-outpost/internal/lib/logger/handler/slogpretty"
	"go-outpost/internal/lib/logger/sl"
	"go-outpost/internal/ws/handler"
	"golang.org/x/exp/slog"
	"net/http"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("Starting ws server...", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	hub := handler.NewHub(log)

	hub.RunServer()

	http.HandleFunc("/ws", hub.HandleConnection)

	log.Info("Server started", slog.String("address", cfg.WSServer.Address))

	srv := &http.Server{
		Addr:         cfg.WSServer.Address,
		ReadTimeout:  cfg.WSServer.Timeout,
		WriteTimeout: cfg.WSServer.Timeout,
		IdleTimeout:  cfg.WSServer.IdleTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Error("Server error", sl.Err(err))
	}

	log.Error("WS server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlogLogger()
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlogLogger() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
