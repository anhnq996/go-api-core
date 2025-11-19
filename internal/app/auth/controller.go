package auth

import (
	"mime/multipart"
	"net/http"

	"api-core/pkg/i18n"
	"api-core/pkg/jwt"
	"api-core/pkg/response"
	"api-core/pkg/validator"

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
	var input LoginRequest

	// Validate request - tự động parse JSON và validate
	if !validator.ValidateAndRespond(w, r, &input) {
		return // Validation failed, response đã được gửi
	}

	resp := h.service.Login(r.Context(), input.Email, input.Password)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
}

// RefreshToken - POST /auth/refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var input RefreshTokenRequest

	// Validate request
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	resp := h.service.RefreshToken(r.Context(), input.RefreshToken)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
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

	resp := h.service.Logout(r.Context(), token)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
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

	resp := h.service.LogoutAll(r.Context(), id)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
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

	resp := h.service.GetUserInfo(r.Context(), id)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
}

// Register - POST /auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterRequest

	// Validate request (will parse multipart form if needed)
	if !validator.ValidateAndRespond(w, r, &input) {
		return
	}

	// Get avatar file if exists
	var avatarFile *multipart.FileHeader
	if file, fileHeader, err := r.FormFile("avatar"); err == nil {
		file.Close() // Close the file handle
		avatarFile = fileHeader
	}

	resp := h.service.Register(r.Context(), input.Name, input.Email, input.Password, nil, avatarFile)
	statusCode := response.GetHTTPStatusCode(resp.Code)
	response.JSON(w, statusCode, *resp)
}
