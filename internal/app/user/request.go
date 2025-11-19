package user

// CreateUserRequest request cho tạo user
type CreateUserRequest struct {
	Name     string  `json:"name" validate:"required,min=2,max=100"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"omitempty,strongpassword"`
	RoleID   *string `json:"role_id" validate:"omitempty,uuid"`
	FCMToken *string `json:"fcm_token" validate:"omitempty"` // Optional: FCM token để gửi notification chào mừng
}

// UpdateUserRequest request cho update user
type UpdateUserRequest struct {
	Name   string  `json:"name" validate:"omitempty,min=2,max=100"`
	Email  string  `json:"email" validate:"omitempty,email"`
	Avatar *string `json:"avatar" validate:"omitempty,url"`
}

// ListUserRequest request cho list users với pagination và sort
type ListUserRequest struct {
	Page    int    `json:"page" validate:"omitempty,min=1"`
	PerPage int    `json:"per_page" validate:"omitempty,min=1,max=100"`
	Sort    string `json:"sort" validate:"omitempty,oneof=name email created_at updated_at"`
	Order   string `json:"order" validate:"omitempty,oneof=asc desc"`
	Search  string `json:"search" validate:"omitempty,max=100"`
}
