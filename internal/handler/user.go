package handler

import (
	"context"
	"fmt"
	"strconv"

	"app/internal/logger"
	"app/internal/models"
	"app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	userUC usecase.UserProvider
}

type UserHandler interface {
	CreateUser(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
	GetAllUsers(c *fiber.Ctx) error
}

func NewHandler(userUC *usecase.UserUsecase) *Handler {
	return &Handler{userUC: userUC}
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" || req.Age <= 0 {
		err := fmt.Errorf("invalid input: %v", err)
		logger.Logger.Error("CreateUser: Invalid input", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	user := models.ToEntityFromCreate(req)
	id, err := h.userUC.CreateUser(user)
	if err != nil {
		logger.Logger.Error("CreateUser: Failed to create user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil || req.ID == "" || req.Name == "" || req.Age <= 0 {
		err := fmt.Errorf("invalid input: %v", err)
		logger.Logger.Error("UpdateUser: Invalid input", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if _, err := uuid.Parse(req.ID); err != nil {
		err := fmt.Errorf("invalid UUID: %v", err)
		logger.Logger.Error("UpdateUser: Invalid UUID", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	user := models.ToEntityFromUpdate(req)
	if err := h.userUC.UpdateUser(user); err != nil {
		logger.Logger.Error("UpdateUser: Failed to update user", "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"id": req.ID})
}

func (h *Handler) GetAllUsers(c *fiber.Ctx) error {
	limit, err1 := strconv.Atoi(c.Query("limit", "10"))
	offset, err2 := strconv.Atoi(c.Query("offset", "0"))
	if err1 != nil || err2 != nil || limit <= 0 || offset < 0 {
		err := fmt.Errorf("invalid pagination params: limit=%d, offset=%d", limit, offset)
		logger.Logger.Error("GetAllUsers: Invalid pagination parameters", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid pagination params"})
	}

	users, err := h.userUC.GetAllUsers(limit, offset)
	if err != nil {
		logger.Logger.Error("GetAllUsers: Failed to get all users", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(users)
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		err := fmt.Errorf("invalid UUID: %v", err)
		logger.Logger.Error("GetUser: Invalid UUID", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	user, err := h.userUC.GetUser(id)
	if err != nil {
		err := fmt.Errorf("user not found: %v", err)
		logger.Logger.Error("GetUser: User not found", "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.JSON(models.ToResponse(user))
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		err := fmt.Errorf("invalid UUID: %v", err)
		logger.Logger.Error("DeleteUser: Invalid UUID", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	if err := h.userUC.DeleteUser(context.Background(), id); err != nil {
		err := fmt.Errorf("user not found: %v", err)
		logger.Logger.Error("DeleteUser: User not found", "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	return c.SendStatus(fiber.StatusOK)
}
