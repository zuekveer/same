package handler

import (
	"context"
	"strconv"

	"app/internal/logger"
	"app/internal/models"
	"app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler interface {
	CreateUser(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
	GetAllUsers(c *fiber.Ctx) error
}

type Handler struct {
	userUC usecase.UserProvider
}

func NewHandler(userUC usecase.UserProvider) UserHandler {
	return &Handler{userUC: userUC}
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" || req.Age <= 0 {
		logger.Logger.Warn("CreateUser: Invalid input", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	user := models.ToEntityFromCreate(req)
	id, err := h.userUC.CreateUser(user)
	if err != nil {
		logger.Logger.Error("CreateUser: Failed to create user", "user", user, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Logger.Info("CreateUser: User created", "id", id)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil || req.ID == "" || req.Name == "" || req.Age <= 0 {
		logger.Logger.Warn("UpdateUser: Invalid input", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if _, err := uuid.Parse(req.ID); err != nil {
		logger.Logger.Warn("UpdateUser: Invalid UUID", "id", req.ID, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	user := models.ToEntityFromUpdate(req)
	if err := h.userUC.UpdateUser(user); err != nil {
		logger.Logger.Error("UpdateUser: Failed to update user", "id", req.ID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Logger.Info("UpdateUser: User updated", "id", req.ID)
	return c.JSON(fiber.Map{"id": req.ID})
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		logger.Logger.Warn("GetUser: Invalid UUID", "id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	user, err := h.userUC.GetUser(id)
	if err != nil {
		logger.Logger.Error("GetUser: User not found", "id", id, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	logger.Logger.Info("GetUser: User found", "id", id)
	return c.JSON(models.ToResponse(user))
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		logger.Logger.Warn("DeleteUser: Invalid UUID", "id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	if err := h.userUC.DeleteUser(context.Background(), id); err != nil {
		logger.Logger.Error("DeleteUser: Failed to delete user", "id", id, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	logger.Logger.Info("DeleteUser: User deleted", "id", id)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetAllUsers(c *fiber.Ctx) error {
	limit, err1 := strconv.Atoi(c.Query("limit", "10"))
	offset, err2 := strconv.Atoi(c.Query("offset", "0"))
	if err1 != nil || err2 != nil || limit <= 0 || offset < 0 {
		logger.Logger.Warn("GetAllUsers: Invalid pagination parameters", "limit", limit, "offset", offset)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid pagination params"})
	}

	users, err := h.userUC.GetAllUsers(limit, offset)
	if err != nil {
		logger.Logger.Error("GetAllUsers: Failed to retrieve users", "limit", limit, "offset", offset, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	logger.Logger.Info("GetAllUsers: Users retrieved", "count", len(users))
	return c.JSON(users)
}
