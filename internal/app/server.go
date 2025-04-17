package app

import (
	"app/internal/config"
	"app/internal/handler"
	"app/internal/repository"
	"app/internal/storage"
	"app/internal/usecase"
)

func Run() error {
	cfg := config.LoadDBConfig()
	connStr := cfg.ConnString()
	db := storage.GetConnect(connStr)
	defer db.Close()

	userRepo := repository.NewUserRepo(db)
	userUC := usecase.NewUserUsecase(userRepo)
	userHandler := handler.NewHandler(userUC)

	app := GetRouter(
		userHandler.GetAllUsers,
		userHandler.CreateUser,
		userHandler.UpdateUser,
		userHandler.GetUser,
		userHandler.DeleteUser,
	)

	return app.Listen(":8088")
}
