package middleware

import (
	"time"

	"app/internal/metrics"

	"github.com/gofiber/fiber/v2"
)

func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/metrics" {
			return c.Next()
		}

		start := time.Now()
		err := c.Next()
		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()

		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}

		metrics.ObserveHttpRequest(c.Method(), path, status, duration)

		return err
	}
}
