package models

import (
	validator "github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(data interface{}) error {
	return validate.Struct(data)
}

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
	// return User{
	//	ID:   req.ID,
	//	Name: req.Name,
	//	Age:  req.Age,
	//}
	return User(req)
}

func (u User) ToResponse() UserResponse {
	// return UserResponse{
	//	ID:   u.ID,
	//	Name: u.Name,
	//	Age:  u.Age,
	//}
	return UserResponse(u)
}

func ToResponseList(users []*User) []UserResponse {
	res := make([]UserResponse, len(users))
	for i, u := range users {
		res[i] = u.ToResponse()
	}
	return res
}

// Could I use init() instead of new validator in all new methods?
//
//	func init() {
//		validate = validator.New()
//	}
func (r *CreateUserRequest) Validate() error {
	return Validate(r)
}

func (r *UpdateUserRequest) Validate() error {
	return Validate(r)
}
