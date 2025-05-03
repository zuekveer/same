package models

import (
	"github.com/go-playground/validator/v10"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type CreateUserRequest struct {
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"required,gte=0,lte=150"`
}

type UpdateUserRequest struct {
	ID   string `json:"id" validate:"required,uuid4"`
	Name string `json:"name" validate:"required"`
	Age  int    `json:"age" validate:"required,gte=0,lte=150"`
}

type UserResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func ToEntityFromCreate(req CreateUserRequest) User {
	return User{
		Name: req.Name,
		Age:  req.Age,
	}
}

func ToEntityFromUpdate(req UpdateUserRequest) User {
	return User{
		ID:   req.ID,
		Name: req.Name,
		Age:  req.Age,
	}
}

func (u User) ToResponse() UserResponse {
	return UserResponse{
		ID:   u.ID,
		Name: u.Name,
		Age:  u.Age,
	}
}

func ToResponseList(users []*User) []UserResponse {
	res := make([]UserResponse, len(users))
	for i, u := range users {
		res[i] = u.ToResponse()
	}
	return res
}

var validate = validator.New()

func (r *CreateUserRequest) Validate() error {
	return validate.Struct(r)
}

func (r *UpdateUserRequest) Validate() error {
	return validate.Struct(r)
}
