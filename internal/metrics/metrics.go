package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestDuration *prometheus.HistogramVec
	httpRequestCount    *prometheus.CounterVec

	cacheHits      prometheus.Counter
	cacheMisses    prometheus.Counter
	cacheEvictions prometheus.Counter
)

func Register() *prometheus.Registry {
	registry := prometheus.NewRegistry()

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

	cacheHits = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits",
	})

	cacheMisses = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses",
	})

	cacheEvictions = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cache_evictions_total",
		Help: "Total number of evicted entries",
	})

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		httpRequestDuration,
		httpRequestCount,
		cacheHits,
		cacheMisses,
		cacheEvictions,
	)

	return registry
}

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

		httpRequestDuration.WithLabelValues(c.Method(), path, strconv.Itoa(status)).Observe(duration)
		httpRequestCount.WithLabelValues(c.Method(), path, strconv.Itoa(status)).Inc()

		return err
	}
}

func IncCacheHits() {
	cacheHits.Inc()
}

func IncCacheMisses() {
	cacheMisses.Inc()
}

func IncCacheEvictions() {
	cacheEvictions.Inc()
}

func RunServer(ctx context.Context, port string, reg *prometheus.Registry) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
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
		_ = server.Shutdown(ctx)
	}()
}
