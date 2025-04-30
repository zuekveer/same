package main

import (
	"context"
	"log/slog"
	"os"

	"app/internal/app"
	"app/internal/logger"
)

func main() {
	logger.Init() //should be removed?

	if err := app.Run(context.Background()); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
