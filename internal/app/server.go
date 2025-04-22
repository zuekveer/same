package app

import (
	"context"
	"fmt"

	"app/internal/config"
	"app/internal/handler"
	"app/internal/logger"
	"app/internal/migration"
	"app/internal/repository"
	"app/internal/storage"
	"app/internal/usecase"
)

func Run(ctx context.Context) error {
	logger.Logger.Info("Starting application")

	cfg := config.LoadConfig()

	logger.Logger.Info("Connecting to database", "db_host", cfg.DB.Host, "db_port", cfg.DB.Port)

	db := storage.GetConnect(ctx, cfg.DB.ConnString())
	defer db.Close()

	logger.Logger.Info("Connected to database")

	logger.Logger.Info("Running migrations...")
	if err := migration.RunMigrations(db); err != nil {
		logger.Logger.Error("Failed to run migrations", "error", err)
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	logger.Logger.Info("Migrations completed")

	userRepo := repository.NewUserRepo(db)
	userUC := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewHandler(userUC)

	app := getRouter(userHandler)

	logger.Logger.Info("Starting server", "port", cfg.App.Port)

	if err := app.Listen(":" + cfg.App.Port); err != nil {
		logger.Logger.Error("Failed to start server", "port", cfg.App.Port, "error", err)
		return fmt.Errorf("failed to start server on port %s: %w", cfg.App.Port, err)
	}

	logger.Logger.Info("Server started successfully", "port", cfg.App.Port)

	return nil
}
