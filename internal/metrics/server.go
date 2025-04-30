package metrics

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
