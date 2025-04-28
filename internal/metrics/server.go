package metrics

import (
	"context"
	"errors"
	"net/http"

	"app/internal/logger"

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
		logger.Logger.Info("Starting metrics server", "port", port)
		err := server.ListenAndServe()

		if err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				logger.Logger.Info("Metrics server closed gracefully")
			} else {
				logger.Logger.Error("Metrics server error", "error", err)
			}
		}
	}()

	go func() {
		<-ctx.Done()
		logger.Logger.Info("Shutting down metrics server")
		_ = server.Shutdown(context.Background())
	}()
}
