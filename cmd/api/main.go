package main

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/pusher/pusher-http-go/v5"
	"go-outpost/internal/config"
	"go-outpost/internal/http-server/handlers/event"
	"go-outpost/internal/http-server/handlers/mysql"
	"go-outpost/internal/http-server/handlers/provably_fair"
	"go-outpost/internal/http-server/handlers/roulette/bet/save"
	"go-outpost/internal/http-server/handlers/roulette/start"
	"go-outpost/internal/http-server/handlers/user/balance"
	"go-outpost/internal/http-server/middleware/logger"
	"go-outpost/internal/lib/logger/handler/slogpretty"
	"go-outpost/internal/lib/logger/sl"
	"go-outpost/internal/repository"
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

	log.Info("Starting server...", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4,utf8&parseTime=True&loc=Local", "root", "123", "localhost:3309", "api")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Error("Failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	// Verify the connection
	if err = db.Ping(); err != nil {
		log.Error("Failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	handler := mysql.New(db)

	pusherClient := pusher.Client{
		AppID:   "1602770",
		Key:     "703a70efb23d918d9db0",
		Secret:  "f91c9c17105b0e1676c6",
		Cluster: "eu",
		Secure:  true,
	}

	pusherEvent := event.NewPusherEvent(log, &pusherClient)

	rouletteBetRepo := repository.NewBetRepository(*handler)
	rouletteRepo := repository.NewRouletteRepository(*handler)
	rouletteWinnerRepo := repository.NewRouletteWinnerRepository(*handler)
	userRepo := repository.NewUserRepository(*handler)
	provablyFairRepo := repository.NewProvablyFairRepository(*handler)

	provablyFair := provably_fair.NewProvablyFair(*provablyFairRepo, log)
	roll := start.NewRouletteRoller(*rouletteWinnerRepo, provablyFair, log)
	startRoulette := start.NewRouletteStart(log, *rouletteRepo, *rouletteBetRepo, pusherEvent, roll)
	userBalance := balance.NewBalance(*userRepo, log, pusherEvent)
	betSave := place_bet.NewBet(log, *rouletteRepo, rouletteBetRepo, *userRepo, userBalance)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(logger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	router.Post("/roulette/start", startRoulette.New())
	router.Post("/roulette/{uuid}/place-bet", betSave.New())

	log.Info("Server started", slog.String("address", cfg.Address))

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err = srv.ListenAndServe(); err != nil {
		log.Error("Server failed", sl.Err(err))
	}

	log.Error("Server stopped")
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
