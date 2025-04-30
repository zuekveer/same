package repository

import (
	"context"
	"log/slog"

	"app/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

var (
	ErrNotFound = errors.New("user not found")
)

type UserRepo struct {
	db *pgxpool.Pool
}

type UserProvider interface {
	Create(user *models.User) (string, error)
	Update(user *models.User) error
	Get(id string) (*models.User, error)
	Delete(cxt context.Context, id string) error
	GetAll(limit, offset int) ([]*models.User, error)
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetAll(limit, offset int) ([]*models.User, error) {
	var users []*models.User
	query := "SELECT id, name, age FROM users ORDER BY id LIMIT $1 OFFSET $2"
	rows, err := r.db.Query(context.Background(), query, limit, offset)
	if err != nil {
		slog.Error("GetAll: Failed to query users", "limit", limit, "offset", offset, "error", err)
		return nil, errors.Wrap(err, "failed to fetch users")
	}
	defer rows.Close()

	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Name, &user.Age); err != nil {
			slog.Error("GetAll: Failed to scan row", "error", err)
			return nil, errors.Wrap(err, "failed to scan row")
		}
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		slog.Error("GetAll: Rows iteration error", "error", err)
		return nil, errors.Wrap(err, "rows iteration error")
	}

	slog.Info("GetAll: Users retrieved", "count", len(users))
	return users, nil
}

func (r *UserRepo) Create(user *models.User) (string, error) {
	id := uuid.New().String()
	user.ID = id

	_, err := r.db.Exec(context.Background(),
		"INSERT INTO users (id, name, age) VALUES ($1, $2, $3)",
		user.ID, user.Name, user.Age)

	if err != nil {
		slog.Error("Create: Failed to insert user", "user", user, "error", err)
		return "", errors.Wrap(err, "failed to create user")
	}

	slog.Info("Create: User created", "user", user)
	return user.ID, nil
}

func (r *UserRepo) Get(id string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(context.Background(),
		"SELECT id, name, age FROM users WHERE id=$1", id).
		Scan(&user.ID, &user.Name, &user.Age)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Error("Get: User not found", "id", id, "error", err)
			return &user, errors.Wrapf(ErrNotFound, "Get user %s:", id)
		}
		slog.Error("Get: Database query failed", "id", id, "error", err)
		return &user, errors.Wrapf(err, "failed to query user %s:", id)
	}
	return &user, nil
}

func (r *UserRepo) Update(user *models.User) error {
	cmd, err := r.db.Exec(context.Background(),
		"UPDATE users SET name=$1, age=$2 WHERE id=$3", user.Name, user.Age, user.ID)
	if err != nil {
		slog.Error("Update: DB error", "error", err)
		return errors.Wrap(err, "update query failed")
	}
	if cmd.RowsAffected() == 0 {
		slog.Warn("Update: User not found", "userID", user.ID)
		return ErrNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id string) error {
	cmd, err := r.db.Exec(ctx, "DELETE FROM users WHERE id=$1", id)
	if err != nil {
		slog.Error("Delete: DB error", "error", err)
		return errors.Wrap(err, "delete query failed")
	}
	if cmd.RowsAffected() == 0 {
		slog.Warn("Delete: User not found", "userID", id)
		return ErrNotFound
	}
	return nil
}
