package models

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type CreateUserRequest struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type UpdateUserRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
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
