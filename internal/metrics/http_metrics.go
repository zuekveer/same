package metrics

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	httpRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
)

func InitHTTPMetrics(reg prometheus.Registerer) {
	reg.MustRegister(httpRequestDuration, httpRequestCount)
}

func Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := c.Response().StatusCode()

		httpRequestDuration.WithLabelValues(c.Method(), c.Route().Path, http.StatusText(status)).Observe(duration)
		httpRequestCount.WithLabelValues(c.Method(), c.Route().Path, http.StatusText(status)).Inc()

		return err
	}
}
