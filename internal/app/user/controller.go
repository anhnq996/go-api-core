package user

import (
	model "anhnq/api-core/internal/models"
	"anhnq/api-core/pkg/i18n"
	"anhnq/api-core/pkg/response"
	"encoding/json"
	"net/http"

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

	users, err := h.service.GetAll()
	if err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	response.Success(w, lang, response.CodeSuccess, users)
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

	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		response.BadRequest(w, lang, response.CodeInvalidInput, nil)
		return
	}

	created, err := h.service.Create(u)
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

	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		response.BadRequest(w, lang, response.CodeInvalidInput, nil)
		return
	}

	updated, err := h.service.Update(id, u)
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
