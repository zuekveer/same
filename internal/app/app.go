package app

import (
	"context"
	"log/slog"
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

	tracerShutdown := tracing.InitTracer(ctx, cfg.Tracing)
	defer func() {
		if err := tracerShutdown(ctx); err != nil {
			slog.Error("error shutting down tracer provider", "error", err)
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

	// Контекст завершения по сигналу
	sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Запуск метрик
	go metrics.RunMetricsServer(sigCtx, cfg.Metrics.Port, reg)

	// Очистка кеша
	go func() {
		ticker := time.NewTicker(cleanupDuration * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				userCachedRepo.CleanupExpired()
			case <-sigCtx.Done():
				return
			}
		}
	}()

	// Запуск Fiber-сервера в горутине
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Starting HTTP server", "port", cfg.App.Port)
		if err := app.Listen(":" + cfg.App.Port); err != nil {
			serverErr <- errors.Wrap(err, "HTTP server failed")
		}
	}()

	// Ожидаем либо завершения, либо ошибки сервера
	select {
	case <-sigCtx.Done():
		slog.Info("Shutdown signal received")
		if err := app.Shutdown(); err != nil {
			slog.Error("Failed to shutdown server gracefully", "error", err)
		}
		return nil
	case err := <-serverErr:
		return err
	}
}
