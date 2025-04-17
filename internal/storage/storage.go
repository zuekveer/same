package storage

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func GetConnect(connStr string) *pgxpool.Pool {
	{
		ctx := context.Background()
		pool, err := pgxpool.Connect(ctx, connStr)
		if err != nil {
			log.Fatalf("Unable to connect to database: %v\n", err)
		}
		if err = pool.Ping(ctx); err != nil {
			log.Fatalf("Unable to ping database: %v\n", err)
		}
		log.Println("Connected to database")
		return pool
	}

}
