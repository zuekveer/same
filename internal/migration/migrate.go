package migration

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id UUID PRIMARY KEY,
	name TEXT NOT NULL,
	age INT NOT NULL
);
`

func RunMigrations(db *pgxpool.Pool) error {
	fmt.Println("Running migrations...")

	_, err := db.Exec(context.Background(), createUsersTable)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	fmt.Println("Migrations completed.")
	return nil
}
