package app

import (
	"context"

	"app/internal/config"
	"app/internal/handler"
	"app/internal/migration"
	"app/internal/repository"
	"app/internal/storage"
	"app/internal/usecase"
)

func Run(ctx context.Context) error {
	cfg := config.LoadDBConfig()
	db := storage.GetConnect(ctx, cfg.ConnString())
	defer db.Close()

	migration.RunMigrations(db)

	userRepo := repository.NewUserRepo(db)
	userUC := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewHandler(userUC)

	app := getRouter(userHandler)

	return app.Listen(":8088")
}
