package usecase

import (
	"app/internal/models"
	"app/internal/repository"
)

type UserUsecase struct {
	userRepo *repository.UserRepo
}

func NewUserUsecase(repo *repository.UserRepo) *UserUsecase {
	return &UserUsecase{userRepo: repo}
}

func (uc *UserUsecase) GetAllUsers(limit, offset int) ([]models.ResponseUser, error) {
	return uc.userRepo.GetAll(limit, offset)
}

func (uc *UserUsecase) CreateUser(user models.RequestUser) (string, error) {
	return uc.userRepo.Create(user)
}

func (uc *UserUsecase) UpdateUser(user models.RequestUser) error {
	return uc.userRepo.Update(user)
}

func (uc *UserUsecase) GetUser(id string) (models.ResponseUser, error) {
	return uc.userRepo.Get(id)
}

func (uc *UserUsecase) DeleteUser(id string) error {
	return uc.userRepo.Delete(id)
}
