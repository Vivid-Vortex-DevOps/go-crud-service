package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/config"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/handler"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/repository"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/service"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/internal/telemetry"
	"github.com/Vivid-Vortex-DevOps/go-crud-service/migrations"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	// ─── Logger (structured JSON) ─────────────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	// ─── Config ───────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// ─── OTel ─────────────────────────────────────────────────────────────────
	ctx := context.Background()
	shutdownOTel, err := telemetry.Setup(ctx,
		cfg.OTELServiceName, "0.1.0", cfg.Environment, cfg.OTELEndpoint,
	)
	if err != nil {
		return fmt.Errorf("setting up OTel: %w", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := shutdownOTel(shutdownCtx); err != nil {
			logger.Error("OTel shutdown error", "error", err)
		}
	}()

	// ─── Database ─────────────────────────────────────────────────────────────
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("creating database pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}
	logger.Info("database connection established")

	// ─── Migrations ───────────────────────────────────────────────────────────
	migrationSource, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("loading migrations: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", migrationSource, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("initialising migrations: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}
	logger.Info("database migrations applied")

	// ─── Wire Dependencies ────────────────────────────────────────────────────
	productRepo := repository.NewPostgresProductRepository(pool)
	productSvc := service.NewProductService(productRepo)
	productHandler := handler.NewProductHandler(productSvc, logger)
	healthHandler := handler.NewHealthHandler(pool)

	// ─── Router ───────────────────────────────────────────────────────────────
	r := chi.NewRouter()
	r.Use(handler.RequestID)
	r.Use(handler.RequestLogger(logger))
	r.Use(handler.Recoverer(logger))

	// Wrap router with OTel HTTP instrumentation
	r.Mount("/api/v1/products", otelhttp.NewHandler(productHandler.Routes(), "products"))
	r.Get("/health/live", healthHandler.Live)
	r.Get("/health/ready", healthHandler.Ready)
	r.Handle("/metrics", promhttp.Handler())

	// ─── Server ───────────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown on SIGINT / SIGTERM
	stopCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", srv.Addr, "environment", cfg.Environment)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		return err
	case <-stopCtx.Done():
		logger.Info("shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
