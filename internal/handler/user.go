package handler

import (
	"log/slog"
	"strconv"

	"app/internal/apperr"
	"app/internal/models"
	"app/internal/usecase"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type UserHandler interface {
	CreateUser(ctx *fiber.Ctx) error
	UpdateUser(ctx *fiber.Ctx) error
	GetUser(ctx *fiber.Ctx) error
	DeleteUser(ctx *fiber.Ctx) error
	GetAllUsers(ctx *fiber.Ctx) error
}

type Handler struct {
	userUC usecase.UserProvider
}

func NewHandler(userUC *usecase.UserUsecase) *Handler {
	return &Handler{
		userUC: userUC,
	}
}

func (h *Handler) CreateUser(ctx *fiber.Ctx) error {
	span := tracing(ctx, "CreateUser")
	defer span.End()

	var req models.CreateUserRequest
	if err := ctx.BodyParser(&req); err != nil || req.Name == "" || req.Age <= 0 {
		slog.Info("CreateUser: Invalid input", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	user := models.ToEntityFromCreate(req)
	id, err := h.userUC.CreateUser(ctx.UserContext(), &user)
	if err != nil {
		slog.Info("CreateUser: Failed to create user", "user", user, "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	slog.Info("CreateUser: User created", "id", id)
	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateUser(ctx *fiber.Ctx) error {
	span := tracing(ctx, "UpdateUser")
	defer span.End()

	var req models.UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil || req.ID == "" || req.Name == "" || req.Age <= 0 {
		slog.Info("UpdateUser: Invalid input", "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if _, err := uuid.Parse(req.ID); err != nil {
		slog.Info("UpdateUser: Invalid UUID", "id", req.ID, "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	user := models.ToEntityFromUpdate(req)
	if err := h.userUC.UpdateUser(ctx.UserContext(), &user); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			slog.Info("UpdateUser: User not found", "id", req.ID)
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		slog.Info("UpdateUser: Failed to update user", "id", req.ID, "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	slog.Info("UpdateUser: User updated", "id", req.ID)
	return ctx.JSON(fiber.Map{"id": req.ID})
}

func (h *Handler) GetUser(ctx *fiber.Ctx) error {
	span := tracing(ctx, "GetUser")
	defer span.End()

	id := ctx.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		slog.Info("GetUser: Invalid UUID", "id", id, "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	user, err := h.userUC.GetUser(ctx.UserContext(), id)
	if err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			slog.Info("GetUser: User not found", "id", id)
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		slog.Info("GetUser: Failed to get user", "id", id, "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	slog.Info("GetUser: User found", "id", id)
	return ctx.JSON(user.ToResponse())
}

func (h *Handler) DeleteUser(ctx *fiber.Ctx) error {
	span := tracing(ctx, "DeleteUser")
	defer span.End()

	id := ctx.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		slog.Info("DeleteUser: Invalid UUID", "id", id, "error", err)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	if err := h.userUC.DeleteUser(ctx.UserContext(), id); err != nil {
		if errors.Is(err, apperr.ErrNotFound) {
			slog.Info("DeleteUser: User not found", "id", id)
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		slog.Info("DeleteUser: Failed to delete user", "id", id, "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	slog.Info("DeleteUser: User deleted", "id", id)
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetAllUsers(ctx *fiber.Ctx) error {
	span := tracing(ctx, "GetAllUsers")
	defer span.End()

	limit, err1 := strconv.Atoi(ctx.Query("limit", "10"))
	offset, err2 := strconv.Atoi(ctx.Query("offset", "0"))
	if err1 != nil || err2 != nil || limit <= 0 || offset < 0 {
		slog.Info("GetAllUsers: Invalid pagination parameters", "limit", limit, "offset", offset)
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid pagination params"})
	}

	users, err := h.userUC.GetAllUsers(ctx.UserContext(), limit, offset)
	if err != nil {
		slog.Info("GetAllUsers: Failed to retrieve users", "limit", limit, "offset", offset, "error", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	slog.Info("GetAllUsers: Users retrieved", "count", len(users))
	return ctx.JSON(models.ToResponseList(users))
}
