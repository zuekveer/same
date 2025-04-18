package mapper

import (
	"app/internal/dto"
	"app/internal/entity"
)

func ToEntityFromCreate(req dto.CreateUserRequest) dto.CreateUserRequest {
	return dto.CreateUserRequest{
		Name: req.Name,
		Age:  req.Age,
	}
}

func ToEntityFromUpdate(req dto.UpdateUserRequest) entity.User {
	return entity.User{
		ID:   req.ID,
		Name: req.Name,
		Age:  req.Age,
	}
}

func ToResponse(user entity.User) dto.UserResponse {
	return dto.UserResponse{
		ID:   user.ID,
		Name: user.Name,
		Age:  user.Age,
	}
}

func ToResponseList(users []entity.User) []dto.UserResponse {
	var res []dto.UserResponse
	for _, u := range users {
		res = append(res, ToResponse(u))
	}
	return res
}
