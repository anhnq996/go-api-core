package user

import (
	model "anhnq/api-core/internal/models"
	"anhnq/api-core/pkg/logger"
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
	logger.RequestLog(r, "Fetching all users")

	users, err := h.service.GetAll()
	if err != nil {
		logger.ErrorLog(r, err, "Failed to fetch users")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.RequestLogWithFields(r, "Users fetched successfully", map[string]interface{}{
		"count": len(users),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// Show - GET /users/{id}
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.RequestLogWithFields(r, "Fetching user by ID", map[string]interface{}{
		"user_id": id,
	})

	user, err := h.service.GetByID(id)
	if err != nil {
		logger.ErrorLog(r, err, "User not found")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	logger.RequestLog(r, "User fetched successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Store - POST /users
func (h *Handler) Store(w http.ResponseWriter, r *http.Request) {
	logger.RequestLog(r, "Creating new user")

	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		logger.ErrorLog(r, err, "Invalid request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	created, err := h.service.Create(u)
	if err != nil {
		logger.ErrorLog(r, err, "Failed to create user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.RequestLogWithFields(r, "User created successfully", map[string]interface{}{
		"user_id": created.ID,
		"email":   created.Email,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

// Update - PUT /users/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.RequestLogWithFields(r, "Updating user", map[string]interface{}{
		"user_id": id,
	})

	var u model.User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		logger.ErrorLog(r, err, "Invalid request body")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updated, err := h.service.Update(id, u)
	if err != nil {
		logger.ErrorLog(r, err, "Failed to update user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.RequestLog(r, "User updated successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updated)
}

// Destroy - DELETE /users/{id}
func (h *Handler) Destroy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	logger.RequestLogWithFields(r, "Deleting user", map[string]interface{}{
		"user_id": id,
	})

	if err := h.service.Delete(id); err != nil {
		logger.ErrorLog(r, err, "Failed to delete user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	logger.RequestLog(r, "User deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}

// Options - OPTIONS /users
func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET,POST,PUT,DELETE,OPTIONS")
	w.WriteHeader(http.StatusOK)
}
