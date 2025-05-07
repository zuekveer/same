package usecase

import (
	"context"

	"app/internal/models"
	"app/internal/repository"
)

type UserUsecase struct {
	userRepo repository.UserProvider
}

type UserProvider interface {
	CreateUser(ctx context.Context, user *models.User) (string, error)
	UpdateUser(ctx context.Context, user *models.User) error
	GetUser(ctx context.Context, id string) (*models.User, error)
	DeleteUser(ctx context.Context, id string) error
	GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
}

func NewUserUsecase(repo repository.UserProvider) *UserUsecase {
	return &UserUsecase{userRepo: repo}
}

func (uc *UserUsecase) GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	return uc.userRepo.GetAll(ctx, limit, offset)
}

func (uc *UserUsecase) CreateUser(ctx context.Context, user *models.User) (string, error) {
	return uc.userRepo.Create(ctx, user)
}

func (uc *UserUsecase) UpdateUser(ctx context.Context, user *models.User) error {
	return uc.userRepo.Update(ctx, user)
}

func (uc *UserUsecase) GetUser(ctx context.Context, id string) (*models.User, error) {
	return uc.userRepo.Get(ctx, id)
}

func (uc *UserUsecase) DeleteUser(ctx context.Context, id string) error {
	return uc.userRepo.Delete(ctx, id)
}
