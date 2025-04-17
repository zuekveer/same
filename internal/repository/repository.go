package repository

import (
	"context"
	"errors"

	"app/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetAll(limit, offset int) ([]models.ResponseUser, error) {
	var users []models.ResponseUser
	query := "SELECT id, name, age FROM users ORDER BY id LIMIT $1 OFFSET $2"
	rows, err := r.db.Query(context.Background(), query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.ResponseUser
		if err := rows.Scan(&user.ID, &user.Name, &user.Age); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepo) Create(user models.RequestUser) (string, error) {
	id := uuid.New().String()
	_, err := r.db.Exec(context.Background(), "INSERT INTO users (id, name, age) VALUES ($1, $2, $3)", id, user.Name, user.Age)
	return id, err
}

func (r *UserRepo) Update(user models.RequestUser) error {
	cmd, err := r.db.Exec(context.Background(), "UPDATE users SET name=$1, age=$2 WHERE id=$3", user.Name, user.Age, user.ID)
	if err != nil || cmd.RowsAffected() == 0 {
		return errors.New("not found or update failed")
	}
	return nil
}

func (r *UserRepo) Get(id string) (models.ResponseUser, error) {
	var user models.ResponseUser
	err := r.db.QueryRow(context.Background(), "SELECT id, name, age FROM users WHERE id=$1", id).Scan(&user.ID, &user.Name, &user.Age)
	return user, err
}

func (r *UserRepo) Delete(id string) error {
	cmd, err := r.db.Exec(context.Background(), "DELETE FROM users WHERE id=$1", id)
	if err != nil || cmd.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}
