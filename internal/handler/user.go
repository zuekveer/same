package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"app/internal/models"
	"app/internal/repository"
	"app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type UserHandler interface {
	CreateUser(c *fiber.Ctx) error
	UpdateUser(c *fiber.Ctx) error
	GetUser(c *fiber.Ctx) error
	DeleteUser(c *fiber.Ctx) error
	GetAllUsers(c *fiber.Ctx) error
}

type Handler struct {
	userUC              usecase.UserProvider
	httpRequestDuration *prometheus.HistogramVec
	requestCount        *prometheus.CounterVec
}

func NewHandler(userUC *usecase.UserUsecase) *Handler {
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "handler", "status_code"},
	)
	prometheus.MustRegister(histogram)

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_request_count",
			Help: "Total number of HTTP requests processed",
		},
		[]string{"method", "handler", "status_code"},
	)
	prometheus.MustRegister(counter)

	return &Handler{
		userUC:              userUC,
		httpRequestDuration: histogram,
		requestCount:        counter,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	statusCode := http.StatusOK

	w.WriteHeader(statusCode)

	duration := time.Since(start).Seconds()

	h.httpRequestDuration.WithLabelValues(r.Method, r.URL.Path, http.StatusText(statusCode)).Observe(duration)
	h.requestCount.WithLabelValues(r.Method, r.URL.Path, http.StatusText(statusCode)).Inc()
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

func (h *Handler) CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil || req.Name == "" || req.Age <= 0 {
		slog.Info("CreateUser: Invalid input", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}

	user := models.ToEntityFromCreate(req)
	id, err := h.userUC.CreateUser(&user)
	if err != nil {
		slog.Info("CreateUser: Failed to create user", "user", user, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	slog.Info("CreateUser: User created", "id", id)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": id})
}

func (h *Handler) UpdateUser(c *fiber.Ctx) error {
	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil || req.ID == "" || req.Name == "" || req.Age <= 0 {
		slog.Info("UpdateUser: Invalid input", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid input"})
	}
	if _, err := uuid.Parse(req.ID); err != nil {
		slog.Info("UpdateUser: Invalid UUID", "id", req.ID, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	user := models.ToEntityFromUpdate(req)
	if err := h.userUC.UpdateUser(&user); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			slog.Info("UpdateUser: User not found", "id", req.ID)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		slog.Info("UpdateUser: Failed to update user", "id", req.ID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	slog.Info("UpdateUser: User updated", "id", req.ID)
	return c.JSON(fiber.Map{"id": req.ID})
}

func (h *Handler) GetUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		slog.Info("GetUser: Invalid UUID", "id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	user, err := h.userUC.GetUser(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			slog.Info("GetUser: User not found", "id", id)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		slog.Info("GetUser: Failed to get user", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	slog.Info("GetUser: User found", "id", id)
	return c.JSON(user.ToResponse())
}

func (h *Handler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if _, err := uuid.Parse(id); err != nil {
		slog.Info("DeleteUser: Invalid UUID", "id", id, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid UUID"})
	}

	if err := h.userUC.DeleteUser(context.Background(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			slog.Info("DeleteUser: User not found", "id", id)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
		}
		slog.Info("DeleteUser: Failed to delete user", "id", id, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	slog.Info("DeleteUser: User deleted", "id", id)
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) GetAllUsers(c *fiber.Ctx) error {
	limit, err1 := strconv.Atoi(c.Query("limit", "10"))
	offset, err2 := strconv.Atoi(c.Query("offset", "0"))
	if err1 != nil || err2 != nil || limit <= 0 || offset < 0 {
		slog.Info("GetAllUsers: Invalid pagination parameters", "limit", limit, "offset", offset)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid pagination params"})
	}

	users, err := h.userUC.GetAllUsers(limit, offset)
	if err != nil {
		slog.Info("GetAllUsers: Failed to retrieve users", "limit", limit, "offset", offset, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	slog.Info("GetAllUsers: Users retrieved", "count", len(users))
	return c.JSON(users)
}
