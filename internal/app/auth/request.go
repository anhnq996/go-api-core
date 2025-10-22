package auth

// LoginRequest request cho login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest request cho register
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,strongpassword"`
}

// RefreshTokenRequest request cho refresh token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// UpdateProfileRequest request cho update profile
type UpdateProfileRequest struct {
	Name   string  `json:"name" validate:"omitempty,min=2,max=100"`
	Avatar *string `json:"avatar" validate:"omitempty,url"`
}

// ChangePasswordRequest request cho đổi password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,strongpassword"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}
