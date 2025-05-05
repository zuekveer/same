package logger

import (
	"log/slog"
	"os"
)

func Init() {
	Logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(Logger)
}
