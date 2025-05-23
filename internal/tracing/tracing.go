package tracing

import (
	"context"
	"log"

	"app/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("app")

func Start(ctx context.Context, name string) (context.Context, trace.Span) {
	return tracer.Start(ctx, name)
}

func Init(ctx context.Context, cfg config.TracingConfig) func(context.Context) error {
	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(cfg.JaegerEndpoint),
		otlptracehttp.WithInsecure(),
		otlptracehttp.WithURLPath("/v1/traces"),
	)
	if err != nil {
		log.Fatalf("failed to initialize OTLP exporter: %v", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("fiber-app"),
		)),
	)

	otel.SetTracerProvider(tp)

	return tp.Shutdown
}
