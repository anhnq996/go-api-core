package response

import (
	"net/http"

	"api-core/pkg/i18n"
)

// Helper functions cho các use cases phổ biến

// RespondWithCode gửi response với code tự động map sang HTTP status
func RespondWithCode(w http.ResponseWriter, lang, code string, data interface{}, errors interface{}) {
	statusCode := GetHTTPStatusCode(code)

	// Xác định success boolean dựa trên HTTP status code
	success := statusCode < 400

	message := i18n.T(lang, code)

	response := Response{
		Success: success,
		Code:    code,
		Message: message,
		Data:    data,
		Errors:  errors,
	}

	JSON(w, statusCode, response)
}

// PaginationFromRequest tạo Meta từ request query params
func PaginationFromRequest(r *http.Request, total int64) *Meta {
	page := 1
	perPage := 10

	// Parse page từ query
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p := parseInt(pageStr); p > 0 {
			page = p
		}
	}

	// Parse per_page từ query
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp := parseInt(perPageStr); pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// parseInt helper để parse string sang int
func parseInt(s string) int {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// NewMeta tạo Meta mới
func NewMeta(page, perPage int, total int64) *Meta {
	totalPages := int(total) / perPage
	if int(total)%perPage != 0 {
		totalPages++
	}

	return &Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// NewErrorDetail tạo ErrorDetail mới
func NewErrorDetail(field, message string) ErrorDetail {
	return ErrorDetail{
		Field:   field,
		Message: message,
	}
}

// ErrorDetailsFromMap tạo slice ErrorDetail từ map
func ErrorDetailsFromMap(errors map[string]string) []ErrorDetail {
	details := make([]ErrorDetail, 0, len(errors))
	for field, message := range errors {
		details = append(details, ErrorDetail{
			Field:   field,
			Message: message,
		})
	}
	return details
}
