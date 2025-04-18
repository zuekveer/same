package handler

import (
	"log"
	"strconv"

	"app/internal/dto"
	"app/internal/mapper"
	"app/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	userUC *usecase.UserUsecase
}

func NewHandler(userUC *usecase.UserUsecase) *Handler {
	return &Handler{userUC: userUC}
}

func (h *Handler) GetAllUsers(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "10")
	offsetStr := c.Query("offset", "0")

	limit, err1 := strconv.Atoi(limitStr)
	offset, err2 := strconv.Atoi(offsetStr)

	if err1 != nil || err2 != nil || limit <= 0 || offset < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid pagination parameters",
		})
	}

	users, err := h.userUC.GetAllUsers(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(mapper.ToResponseList(users))
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	log.Println(">>> CreateUser handler called")

	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		log.Println("Body parse error:", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	log.Printf("Parsed request: %+v\n", req)

	user := mapper.ToEntityFromCreate(req)
	log.Printf("Mapped user entity: %+v\n", user)

	id, err := h.userUC.CreateUser(user)
	log.Println("CreateUser usecase returned:", id, err)

	if err != nil || id == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil || req.ID == "" || req.Name == "" || req.Age <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	user := mapper.ToEntityFromUpdate(req)

	err := h.userUC.UpdateUser(user)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"id": user.ID})
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.userUC.GetUser(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(mapper.ToResponse(user))
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.userUC.DeleteUser(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.SendStatus(fiber.StatusOK)
}
