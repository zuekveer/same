package models

type UserDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type RequestUser struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type ResponseUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}
