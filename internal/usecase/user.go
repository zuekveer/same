package usecase

import (
	"app/internal/dto"
	"app/internal/repository"
)

type UserUsecase struct {
	userRepo *repository.UserRepo
}

func NewUserUsecase(repo *repository.UserRepo) *UserUsecase {
	return &UserUsecase{userRepo: repo}
}

func (uc *UserUsecase) GetAllUsers(limit, offset int) ([]dto.UserResponse, error) {
	return uc.userRepo.GetAll(limit, offset)
}

func (uc *UserUsecase) CreateUser(user dto.CreateUserRequest) (string, error) {
	return uc.userRepo.Create(user)
}

func (uc *UserUsecase) UpdateUser(user dto.UpdateUserRequest) error {
	return uc.userRepo.Update(user)
}

func (uc *UserUsecase) GetUser(id string) (dto.UserResponse, error) {
	return uc.userRepo.Get(id)
}

func (uc *UserUsecase) DeleteUser(id string) error {
	return uc.userRepo.Delete(id)
}
