package auth

import (
	"net/http"

	"anhnq/api-core/pkg/i18n"
	"anhnq/api-core/pkg/jwt"
	"anhnq/api-core/pkg/response"
	"anhnq/api-core/pkg/validator"

	"github.com/google/uuid"
)

// Handler xử lý HTTP requests cho auth
type Handler struct {
	service *Service
}

// NewHandler tạo auth handler mới
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Login - POST /auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())

	var input LoginRequest

	// Validate request - tự động parse JSON và validate
	if !validator.ValidateAndRespond(w, r, &input) {
		return // Validation failed, response đã được gửi
	}

	// Login
	result, err := h.service.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		if err == ErrInvalidCredentials {
			response.Unauthorized(w, lang, response.CodeInvalidCredentials)
			return
		}
		if err == ErrUserInactive {
			response.Forbidden(w, lang, response.CodeAccountDisabled)
			return
		}
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeLoginSuccess, result)
}

// RefreshToken - POST /auth/refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())

	var input RefreshTokenRequest

	// Validate request
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	// Refresh token
	result, err := h.service.RefreshToken(r.Context(), input.RefreshToken)
	if err != nil {
		if err == jwt.ErrExpiredToken {
			response.Unauthorized(w, lang, response.CodeTokenExpired)
			return
		}
		if err == jwt.ErrInvalidToken {
			response.Unauthorized(w, lang, response.CodeTokenInvalid)
			return
		}
		if err == ErrUserNotFound {
			response.NotFound(w, lang, response.CodeUserNotFound)
			return
		}
		if err == ErrUserInactive {
			response.Forbidden(w, lang, response.CodeAccountDisabled)
			return
		}
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeTokenRefreshed, result)
}

// Logout - POST /auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())

	// Get token from header
	token := jwt.ExtractTokenFromHeader(r)
	if token == "" {
		response.Unauthorized(w, lang, response.CodeTokenMissing)
		return
	}

	// Logout
	if err := h.service.Logout(r.Context(), token); err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeLogoutSuccess, nil)
}

// LogoutAll - POST /auth/logout-all
func (h *Handler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	userID := jwt.GetUserIDFromContext(r.Context())

	if userID == "" {
		response.Unauthorized(w, lang, response.CodeTokenMissing)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeInvalidInput, nil)
		return
	}

	// Logout all
	if err := h.service.LogoutAll(r.Context(), id); err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeLogoutSuccess, nil)
}

// GetMe - GET /auth/me
func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	userID := jwt.GetUserIDFromContext(r.Context())

	if userID == "" {
		response.Unauthorized(w, lang, response.CodeUnauthorized)
		return
	}

	id, err := uuid.Parse(userID)
	if err != nil {
		response.BadRequest(w, lang, response.CodeInvalidInput, nil)
		return
	}

	// Get user info
	user, err := h.service.GetUserInfo(r.Context(), id)
	if err != nil {
		if err == ErrUserNotFound {
			response.NotFound(w, lang, response.CodeUserNotFound)
			return
		}
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeSuccess, user)
}

// Register - POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())

	var input RegisterRequest

	// Validate request
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	// Register
	user, err := h.service.Register(r.Context(), input.Name, input.Email, input.Password, nil)
	if err != nil {
		if err.Error() == "email already exists" {
			response.Conflict(w, lang, response.CodeEmailAlreadyExists)
			return
		}
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Created(w, lang, response.CodeCreated, user)
}
