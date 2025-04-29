package app

import (
	"app/internal/handler"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
)

func getRouter(h handler.UserHandler) *fiber.App {
	app := fiber.New()

	app.Post("/user", h.CreateUser)
	app.Put("/user", h.UpdateUser)
	app.Get("/user/:id", h.GetUser)
	app.Delete("/user/:id", h.DeleteUser)
	app.Get("/users", h.GetAllUsers)

	app.Get("/metrics", adaptor.HTTPHandler(handler.MetricsHandler()))

	return app
}
