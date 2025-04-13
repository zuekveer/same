package main

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

var (
	users = make(map[string]User)
	mu    sync.RWMutex
)

func main() {
	app := fiber.New()

	app.Post("/user", createUser)
	app.Put("/user", updateUser)
	app.Get("/user/:id", getUser)
	app.Delete("/user/:id", deleteUser)

	log.Fatal(app.Listen(":8088"))
}

func createUser(c *fiber.Ctx) error {
	var input User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if input.Name == "" || input.Age <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid name or age"})
	}

	input.ID = uuid.New().String()
	mu.Lock()
	users[input.ID] = input
	mu.Unlock()

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": input.ID})
}

func updateUser(c *fiber.Ctx) error {
	var input User
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if input.ID == "" || input.Name == "" || input.Age <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := users[input.ID]; !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	users[input.ID] = input

	return c.JSON(fiber.Map{"id": input.ID})
}

func getUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing ID"})
	}
	mu.RLock()
	defer mu.RUnlock()

	user, ok := users[id]
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	return c.JSON(user)
}

func deleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing ID"})
	}

	mu.Lock()
	defer mu.Unlock()

	if _, ok := users[id]; !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	delete(users, id)

	return c.SendStatus(fiber.StatusOK)
}
