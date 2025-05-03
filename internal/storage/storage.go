package storage

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func GetConnect(ctx context.Context, connStr string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		slog.Error("Unable to connect to database: %v\n", err)
	}
	if err := pool.Ping(ctx); err != nil {
		slog.Error("Unable to ping database: %v\n", err)
	}
	slog.Info("Connected to database")
	return pool, nil
}
