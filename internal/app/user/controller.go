package user

import (
	"mime/multipart"
	"net/http"

	model "anhnq/api-core/internal/models"
	"anhnq/api-core/pkg/i18n"
	"anhnq/api-core/pkg/response"
	"anhnq/api-core/pkg/utils"
	"anhnq/api-core/pkg/validator"

	"github.com/go-chi/chi/v5"
)

// Handler chứa service của user
type Handler struct {
	service *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// Index - GET /users
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())

	// Parse query parameters using common function
	params := utils.ParseQueryParams(r)

	// Get users with pagination
	users, pagination, err := h.service.GetListWithPagination(params.Page, params.PerPage, params.Sort, params.Order, params.Search)
	if err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	// Create response data using common helper
	responseData := utils.PaginatedResponse(users, pagination)

	response.Success(w, lang, response.CodeSuccess, responseData)
}

// Show - GET /users/{id}
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	id := chi.URLParam(r, "id")

	user, err := h.service.GetByID(id)
	if err != nil {
		response.NotFound(w, lang, response.CodeUserNotFound)
		return
	}

	response.Success(w, lang, response.CodeSuccess, user)
}

// Store - POST /users
func (h *Handler) Store(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())

	var input CreateUserRequest

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

	// Convert to model
	u := model.User{
		Name:  input.Name,
		Email: input.Email,
	}

	created, err := h.service.Create(u, avatarFile)
	if err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Created(w, lang, response.CodeCreated, created)
}

// Update - PUT /users/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	id := chi.URLParam(r, "id")

	var input UpdateUserRequest

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

	// Convert to model
	u := model.User{
		Name:   input.Name,
		Email:  input.Email,
		Avatar: input.Avatar,
	}

	updated, err := h.service.Update(id, u, avatarFile)
	if err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeUpdated, updated)
}

// Destroy - DELETE /users/{id}
func (h *Handler) Destroy(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(id); err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeDeleted, nil)
}

// Options - OPTIONS /users
func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET,POST,PUT,DELETE,OPTIONS")
	w.WriteHeader(http.StatusOK)
}
