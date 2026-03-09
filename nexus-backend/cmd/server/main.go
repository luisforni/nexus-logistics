package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/luisforni/nexus-logistics/backend/internal/api"
	"github.com/luisforni/nexus-logistics/backend/internal/config"
	"github.com/luisforni/nexus-logistics/backend/internal/repository/postgres"
	"github.com/luisforni/nexus-logistics/backend/internal/service"
	"github.com/luisforni/nexus-logistics/backend/pkg/blockchain"
	"github.com/luisforni/nexus-logistics/backend/pkg/cache"
	"github.com/luisforni/nexus-logistics/backend/pkg/telemetry"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("APP_ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load configuration")
	}

	ctx := context.Background()
	tp, err := telemetry.InitTracer(ctx, cfg.OtelEndpoint)
	if err != nil {
		log.Warn().Err(err).Msg("telemetry init failed – continuing without tracing")
	} else {
		defer func() { _ = tp.Shutdown(ctx) }()
	}

	db, err := postgres.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("database connection failed")
	}

	_, filename, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	if envPath := os.Getenv("MIGRATIONS_PATH"); envPath != "" {
		migrationsPath = envPath
	}
	if err := postgres.RunMigrations(cfg.DatabaseURL, migrationsPath); err != nil {
		log.Fatal().Err(err).Msg("database migration failed")
	}

	if err := postgres.SeedAdminUser(ctx, db); err != nil {
		log.Fatal().Err(err).Msg("admin seed failed")
	}

	redisClient, err := cache.NewRedisClient(cfg.RedisURL, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatal().Err(err).Msg("redis connection failed")
	}

	chainClient, err := blockchain.NewClient(cfg.EthereumRPCURL, cfg.ContractAddress)
	if err != nil {
		log.Warn().Err(err).Msg("blockchain client init failed – immutable trace disabled")
	}

	shipmentRepo := postgres.NewShipmentRepository(db)
	userRepo := postgres.NewUserRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiry, cfg.JWTRefreshExpiry)
	shipmentSvc := service.NewShipmentService(shipmentRepo, chainClient, redisClient)

	router := api.NewRouter(cfg, authSvc, shipmentSvc)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info().Str("port", cfg.Port).Msg("Nexus Backend listening")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-quit
	log.Info().Msg("shutting down gracefully…")

	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Error().Err(err).Msg("forced shutdown")
	}
	log.Info().Msg("server stopped")
}
