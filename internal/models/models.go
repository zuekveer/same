package models

type User struct {
	ID   string
	Name string
	Age  int
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

func ToResponse(user User) UserResponse {
	return UserResponse{
		ID:   user.ID,
		Name: user.Name,
		Age:  user.Age,
	}
}

func ToResponseList(users []User) []UserResponse {
	var res []UserResponse
	for _, u := range users {
		res = append(res, ToResponse(u))
	}
	return res
}
