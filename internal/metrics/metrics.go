package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HttpRequestDuration *prometheus.HistogramVec
	HttpRequestCount    *prometheus.CounterVec

	cacheCounter *prometheus.CounterVec
)

func Register(ctx context.Context, port string) *prometheus.Registry {
	registry := prometheus.NewRegistry()

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	HttpRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	cacheCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_events_total",
			Help: "Total number of cache events",
		},
		[]string{"status"}, // status: hit | miss | expired
	)

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		HttpRequestDuration,
		HttpRequestCount,
		cacheCounter,
	)

	go runServer(ctx, port, registry)

	return registry
}

func ObserveHttpRequest(method, path string, status int, duration float64) {
	HttpRequestDuration.WithLabelValues(method, path, strconv.Itoa(status)).Observe(duration)
	HttpRequestCount.WithLabelValues(method, path, strconv.Itoa(status)).Inc()
}

func IncCacheHits() {
	cacheCounter.WithLabelValues("hit").Inc()
}

func IncCacheMisses() {
	cacheCounter.WithLabelValues("miss").Inc()
}

func IncCacheExpired() {
	cacheCounter.WithLabelValues("expired").Inc()
}

func runServer(ctx context.Context, port string, reg *prometheus.Registry) {
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
		if err := server.ListenAndServe(); err != nil {
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
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Error during metrics shutdown", "error", err)
		}
	}()
}
