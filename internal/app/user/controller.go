package user

import (
	"mime/multipart"
	"net/http"
	"time"

	model "anhnq/api-core/internal/models"
	"anhnq/api-core/pkg/excel"
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

// UserExportData struct cho export users
type UserExportData struct {
	ID              string `json:"id" excel:"ID"`
	Name            string `json:"name" excel:"Name"`
	Email           string `json:"email" excel:"Email"`
	Avatar          string `json:"avatar" excel:"Avatar"`
	RoleName        string `json:"role_name" excel:"Role"`
	EmailVerifiedAt string `json:"email_verified_at" excel:"Email Verified"`
	IsActive        bool   `json:"is_active" excel:"Active"`
	LastLoginAt     string `json:"last_login_at" excel:"Last Login"`
	CreatedAt       string `json:"created_at" excel:"Created At"`
	UpdatedAt       string `json:"updated_at" excel:"Updated At"`
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

// ExportUsers - GET /users/export
func (h *Handler) ExportUsers(w http.ResponseWriter, r *http.Request) {
	lang := i18n.GetLanguageFromContext(r.Context())

	// Get query parameters
	format := r.URL.Query().Get("format") // excel, csv
	if format == "" {
		format = "excel" // default to excel
	}

	// Get all users (without pagination for export)
	users, _, err := h.service.GetListWithPagination(1, 1000, "", "", "") // Get up to 1000 users
	if err != nil {
		response.InternalServerError(w, lang, response.CodeInternalServerError)
		return
	}

	// Prepare export data
	exportData := make([]UserExportData, len(users))
	for i, user := range users {
		exportData[i] = UserExportData{
			ID:              user.ID.String(),
			Name:            user.Name,
			Email:           user.Email,
			Avatar:          getAvatarURL(user.Avatar),
			RoleName:        getRoleName(user.Role),
			EmailVerifiedAt: formatTime(user.EmailVerifiedAt),
			IsActive:        user.IsActive,
			LastLoginAt:     formatTime(user.LastLoginAt),
			CreatedAt:       user.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:       user.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	// Set headers based on format
	filename := "users_" + time.Now().Format("20060102_150405")

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename+".csv")

		// Export to CSV
		excelManager := excel.NewExcelManager()
		headers := []string{"ID", "Name", "Email", "Avatar", "Role", "Email Verified", "Active", "Last Login", "Created At", "Updated At"}

		if err := excelManager.ExportToCSV(exportData, headers, w); err != nil {
			response.InternalServerError(w, lang, response.CodeInternalServerError)
			return
		}
	} else {
		// Default to Excel
		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename="+filename+".xlsx")

		// Export to Excel
		excelManager := excel.NewExcelManager()
		headers := []string{"ID", "Name", "Email", "Avatar", "Role", "Email Verified", "Active", "Last Login", "Created At", "Updated At"}

		if err := excelManager.ExportToExcel(exportData, "Users", headers); err != nil {
			response.InternalServerError(w, lang, response.CodeInternalServerError)
			return
		}

		// Write Excel file to response
		if err := excelManager.WriteToWriter(w); err != nil {
			response.InternalServerError(w, lang, response.CodeInternalServerError)
			return
		}
	}
}

// Options - OPTIONS /users
func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "GET,POST,PUT,DELETE,OPTIONS")
	w.WriteHeader(http.StatusOK)
}

// Helper functions for export
func getAvatarURL(avatar *string) string {
	if avatar == nil || *avatar == "" {
		return ""
	}
	return *avatar
}

func getRoleName(role *model.Role) string {
	if role == nil {
		return ""
	}
	return role.DisplayName
}

func formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}
