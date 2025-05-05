package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	registry *prometheus.Registry
}

func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	InitHTTPMetrics(reg)

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &Metrics{registry: reg}
}

func (m *Metrics) Registry() *prometheus.Registry {
	return m.registry
}

// http metrics
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

// cache metrics
type CacheMetrics struct {
	Hits      prometheus.Counter
	Misses    prometheus.Counter
	Evictions prometheus.Counter
}

func NewCacheMetrics(reg prometheus.Registerer) *CacheMetrics {
	hits := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	})
	misses := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	})
	evictions := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_evictions_total",
		Help: "Total number of evicted entries",
	})

	reg.MustRegister(hits, misses, evictions)

	return &CacheMetrics{
		Hits:      hits,
		Misses:    misses,
		Evictions: evictions,
	}
}

//run

func RunMetricsServer(ctx context.Context, port string, registry *prometheus.Registry) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		slog.Info("Starting metrics server", "port", port)
		err := server.ListenAndServe()

		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				slog.Info("Metrics server closed gracefully")
			} else {
				slog.Error("Metrics server error", "error", err)
			}
		}
	}()

	go func() {
		<-ctx.Done()
		slog.Info("Shutting down metrics server")
		_ = server.Shutdown(context.Background())
	}()
}
