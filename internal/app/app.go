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

	slog.Info("Running migrations...", "dsn", cfg.DB.ConnString())
	if err := database.Migrate(cfg.DB.ConnString()); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		return errors.Wrap(err, "run migrations")
	}
	slog.Info("Migrations completed")

	slog.Info("Connecting to database", "db_host", cfg.DB.Host, "db_port", cfg.DB.Port)
	db, _ := storage.GetConnect(ctx, cfg.DB.ConnString())
	defer db.Close()
	slog.Info("Connected to database")

	expirationDuration := 10 * time.Minute
	cleanupInterval := 5 * time.Minute

	userRepo := repository.NewUserRepo(db)
	userCachedRepo := cache.NewDecorator(userRepo, expirationDuration, cleanupInterval)
	userUC := usecase.NewUserUsecase(userCachedRepo)
	userHandler := handler.NewHandler(userUC)
	app := getRouter(userHandler)

	shutdownCtx, shutdownCancel := context.WithCancel(ctx)
	defer shutdownCancel()

	go userCachedRepo.CleanupExpiredLoop(shutdownCtx)

	metricsServer := metrics.NewMetrics()
	go metrics.RunMetricsServer(shutdownCtx, cfg.Metrics.Port, metricsServer.Registry())

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		shutdownCancel()
	}()

	if err := app.Listen(":" + cfg.App.Port); err != nil {
		slog.Error("Failed to start server", "port", cfg.App.Port, "error", err)
		return errors.Wrapf(err, "failed to start server on port %s:", cfg.App.Port)
	}

	<-shutdownCtx.Done()

	slog.Info("Server shutdown gracefully")
	return nil
}
