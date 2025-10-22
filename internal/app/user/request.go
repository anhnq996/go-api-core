package user

// CreateUserRequest request cho tạo user
type CreateUserRequest struct {
	Name     string  `json:"name" validate:"required,min=2,max=100"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"omitempty,strongpassword"`
	RoleID   *string `json:"role_id" validate:"omitempty,uuid"`
}

// UpdateUserRequest request cho update user
type UpdateUserRequest struct {
	Name   string  `json:"name" validate:"omitempty,min=2,max=100"`
	Email  string  `json:"email" validate:"omitempty,email"`
	Avatar *string `json:"avatar" validate:"omitempty,url"`
}
