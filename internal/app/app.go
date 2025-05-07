package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/database"
	"app/internal/cache"
	"app/internal/config"
	"app/internal/handler"
	"app/internal/logger"
	"app/internal/metrics"
	"app/internal/repository"
	"app/internal/storage"
	"app/internal/tracing"
	"app/internal/usecase"

	"github.com/pkg/errors"
)

func Run(ctx context.Context) error {
	slog.Info("Starting application")
	logger.Init()

	cfg, err := config.LoadConfig()
	if err != nil {
		return errors.Wrap(err, "failed to load config")
	}

	if err := database.Migrate(cfg.DB.ConnString()); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		return errors.Wrap(err, "run migrations")
	}

	slog.Info("Connecting to database", "db_host", cfg.DB.Host, "db_port", cfg.DB.Port)
	db, err := storage.GetConnect(ctx, cfg.DB.ConnString())
	if err != nil {
		return errors.Wrap(err, "failed to connect to database")
	}
	defer db.Close()

	tracer := tracing.InitTracer(cfg.Tracing)
	defer func() {
		if err := tracer(context.Background()); err != nil {
			slog.Error("error shutting down tracer provider: %v", err)
		}
	}()

	expirationDuration := time.Duration(cfg.Cache.ExpirationMinutes) * time.Minute
	cleanupDuration := time.Duration(cfg.Cache.CleanupMinutes)

	userRepo := repository.NewUserRepo(db)
	userCachedRepo := cache.NewDecorator(userRepo, expirationDuration)

	userUC := usecase.NewUserUsecase(userCachedRepo)
	userHandler := handler.NewHandler(userUC)
	app := getRouter(userHandler)

	reg := metrics.Register()

	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	go metrics.RunMetricsServer(shutdownCtx, cfg.Metrics.Port, reg)

	// Periodic cleanup loop
	go func() {
		ticker := time.NewTicker(cleanupDuration * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				userCachedRepo.CleanupExpired()
			case <-shutdownCtx.Done():
				return
			}
		}
	}()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		shutdownCancel()
	}()

	if err := app.Listen(":" + cfg.App.Port); err != nil {
		return errors.Wrapf(err, "failed to start server on port %s:", cfg.App.Port)
	}

	<-shutdownCtx.Done()

	return nil
}
