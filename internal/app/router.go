package app

import (
	"app/internal/handler"
	"app/internal/metrics"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getRouter(h handler.UserHandler) *fiber.App {
	app := fiber.New()

	app.Post("/user", h.CreateUser)
	app.Put("/user", h.UpdateUser)
	app.Get("/user/:id", h.GetUser)
	app.Delete("/user/:id", h.DeleteUser)
	app.Get("/users", h.GetAllUsers)

	app.Use(metrics.Middleware())
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	return app
}
