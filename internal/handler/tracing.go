package handler

import (
	fiber "github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("handlers")

func tracing(c *fiber.Ctx, spanName string) trace.Span {
	ctx, span := tracer.Start(c.UserContext(), spanName)
	c.SetUserContext(ctx)
	return span
}
