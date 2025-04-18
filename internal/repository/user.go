package repository

import (
	"context"
	"errors"

	"app/internal/dto"
	"app/internal/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetAll(limit, offset int) ([]entity.User, error) {
	var users []entity.User
	query := "SELECT id, name, age FROM users ORDER BY id LIMIT $1 OFFSET $2"
	rows, err := r.db.Query(context.Background(), query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user entity.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Age); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *UserRepo) Create(req dto.CreateUserRequest) (string, error) {
	id := uuid.New().String()

	// data from DTO to Entity
	user := entity.User{
		ID:   id,
		Name: req.Name,
		Age:  req.Age,
	}

	// add data in db
	cmdTag, err := r.db.Exec(context.Background(), "INSERT INTO users (id, name, age) VALUES ($1, $2, $3)", user.ID, user.Name, user.Age)
	if err != nil {
		return "", err
	}
	if cmdTag.RowsAffected() == 0 {
		return "", errors.New("user was not created")
	}

	return user.ID, nil
}

func (r *UserRepo) Update(user entity.User) error {
	cmd, err := r.db.Exec(context.Background(), "UPDATE users SET name=$1, age=$2 WHERE id=$3", user.Name, user.Age, user.ID)
	if err != nil || cmd.RowsAffected() == 0 {
		return errors.New("not found or update failed")
	}
	return nil
}

func (r *UserRepo) Get(id string) (entity.User, error) {
	var user entity.User
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
