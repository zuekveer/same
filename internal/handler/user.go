package handler

import (
	"strconv"

	"app/internal/dto"
	"app/internal/usecase"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	userUC *usecase.UserUsecase
}

func NewHandler(userUC *usecase.UserUsecase) *Handler {
	return &Handler{userUC: userUC}
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	id, err := h.userUC.CreateUser(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	err := h.userUC.UpdateUser(req)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"id": req.ID})
}

func (h *Handler) GetAllUsers(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	users, err := h.userUC.GetAllUsers(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(users)
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	user, err := h.userUC.GetUser(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(user)
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	err := h.userUC.DeleteUser(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.SendStatus(fiber.StatusOK)
}
