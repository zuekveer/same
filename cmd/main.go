package main

import (
	"context"
	"log/slog"
	"os"

	"app/internal/app"
)

func main() {

	if err := app.Run(context.Background()); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}
