package migration

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL,
	age INT NOT NULL
);
`

func RunMigrations(db *pgxpool.Pool) {
	log.Println("Running migrations...")

	_, err := db.Exec(context.Background(), createUsersTable)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations completed.")
}
