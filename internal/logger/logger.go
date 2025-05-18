package logger

import (
	"log/slog"
	"os"
)

func Init(level string) {
	var slogLevel slog.Level
	if err := slogLevel.UnmarshalText([]byte(level)); err != nil {
		slogLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	}))
	slog.SetDefault(logger)
}
