package main

import (
	"context"
	"os"

	"app/internal/app"
	"app/internal/logger"
)

func main() {
	logger.Init()

	if err := app.Run(context.Background()); err != nil {
		logger.Logger.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
