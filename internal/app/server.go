package app

import (
	"app/internal/config"
	"app/internal/handler"
	"app/internal/migration"
	"app/internal/repository"
	"app/internal/storage"
	"app/internal/usecase"
)

func Run() error {
	cfg := config.LoadDBConfig()
	db := storage.GetConnect(cfg.ConnString())
	defer db.Close()

	migration.RunMigrations(db)

	userRepo := repository.NewUserRepo(db)
	userUC := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewHandler(userUC)

	app := GetRouter(
		userHandler.CreateUser,
		userHandler.UpdateUser,
		userHandler.GetUser,
		userHandler.DeleteUser,
		userHandler.GetAllUsers,
	)

	return app.Listen(":8088")
}
