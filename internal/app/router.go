package app

import (
	"github.com/gofiber/fiber/v2"
)

func GetRouter(
	createUser fiber.Handler,
	updateUser fiber.Handler,
	getUser fiber.Handler,
	deleteUser fiber.Handler,
) *fiber.App {
	app := fiber.New()

	app.Post("/user", createUser)
	app.Put("/user", updateUser)
	app.Get("/user/:id", getUser)
	app.Delete("/user/:id", deleteUser)

	return app
}
