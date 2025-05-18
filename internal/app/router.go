package app

import (
	"app/internal/handler"
	"app/internal/middleware"

	fiber "github.com/gofiber/fiber/v2"
)

func getRouter(h handler.UserHandler) *fiber.App {
	app := fiber.New()

	app.Use(middleware.RequestMetrics())
	app.Post("/user", h.CreateUser)
	app.Put("/user", h.UpdateUser)
	app.Get("/user/:id", h.GetUser)
	app.Delete("/user/:id", h.DeleteUser)
	app.Get("/users", h.GetAllUsers)

	return app
}
