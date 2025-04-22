package repository

import (
	"context"
	"fmt"

	"app/internal/logger"
	"app/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

type UserProvider interface {
	Create(user models.User) (string, error)
	Update(user models.User) error
	Get(id string) (models.User, error)
	Delete(cxt context.Context, id string) error
	GetAll(limit, offset int) ([]models.User, error)
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetAll(limit, offset int) ([]models.User, error) {
	var users []models.User
	query := "SELECT id, name, age FROM users ORDER BY id LIMIT $1 OFFSET $2"
	rows, err := r.db.Query(context.Background(), query, limit, offset)
	if err != nil {
		logger.Logger.Error("GetAll: Database query failed", "error", err)
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Age); err != nil {
			logger.Logger.Error("GetAll: Row scan failed", "error", err)
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepo) Create(user models.User) (string, error) {
	id := uuid.New().String()
	user.ID = id

	_, err := r.db.Exec(context.Background(),
		"INSERT INTO users (id, name, age) VALUES ($1, $2, $3)",
		user.ID, user.Name, user.Age)

	if err != nil {
		logger.Logger.Error("Create: Database insert failed", "error", err)
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return user.ID, nil
}

func (r *UserRepo) Update(user models.User) error {
	cmd, err := r.db.Exec(context.Background(),
		"UPDATE users SET name=$1, age=$2 WHERE id=$3", user.Name, user.Age, user.ID)
	if err != nil || cmd.RowsAffected() == 0 {
		err := fmt.Errorf("not found or update failed: %w", err)
		logger.Logger.Error("Update: User update failed", "error", err)
		return fmt.Errorf("update failed: %w", err)
	}
	return nil
}

func (r *UserRepo) Get(id string) (models.User, error) {
	var user models.User
	err := r.db.QueryRow(context.Background(),
		"SELECT id, name, age FROM users WHERE id=$1", id).
		Scan(&user.ID, &user.Name, &user.Age)
	if err != nil {
		logger.Logger.Error("Get: User not found", "error", err)
		return user, fmt.Errorf("user not found: %w", err)
	}
	return user, nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	cmd, err := r.db.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
	if err != nil || cmd.RowsAffected() == 0 {
		err := fmt.Errorf("user not found: %w", err)
		logger.Logger.Error("Delete: User not found", "error", err)
		return fmt.Errorf("user not found: %w", err)
	}
	return nil
}
