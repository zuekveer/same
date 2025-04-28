package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"app/database"
	"app/internal/cache"
	"app/internal/config"
	"app/internal/handler"
	"app/internal/logger"
	"app/internal/repository"
	"app/internal/storage"
	"app/internal/usecase"
)

func Run(ctx context.Context) error {
	logger.Logger.Info("Starting application")

	cfg := config.LoadConfig()

	logger.Logger.Info("Running migrations...", "dsn", cfg.DB.ConnString())
	if err := database.Migrate(cfg.DB.ConnString()); err != nil {
		logger.Logger.Error("Failed to run migrations", "error", err)
		return fmt.Errorf("run migrations: %w", err)
	}
	logger.Logger.Info("Migrations completed")

	logger.Logger.Info("Connecting to database", "db_host", cfg.DB.Host, "db_port", cfg.DB.Port)
	db := storage.GetConnect(ctx, cfg.DB.ConnString())
	defer db.Close()
	logger.Logger.Info("Connected to database")

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

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		shutdownCancel()
	}()

	logger.Logger.Info("Starting server", "port", cfg.App.Port)
	if err := app.Listen(":" + cfg.App.Port); err != nil {
		logger.Logger.Error("Failed to start server", "port", cfg.App.Port, "error", err)
		return fmt.Errorf("failed to start server on port %s: %w", cfg.App.Port, err)
	}

	<-shutdownCtx.Done()

	logger.Logger.Info("Server shutdown gracefully")
	return nil
}
