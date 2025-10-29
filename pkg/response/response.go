package response

import (
	"encoding/json"
	"net/http"

	"api-core/pkg/i18n"
)

// Response là cấu trúc chuẩn cho API response
type Response struct {
	Success bool        `json:"success"`          // true nếu thành công, false nếu lỗi
	Code    string      `json:"code"`             // Mã response (SUCCESS, INVALID_INPUT, etc.)
	Message string      `json:"message"`          // Thông điệp đã được dịch
	Data    interface{} `json:"data,omitempty"`   // Dữ liệu trả về (nếu có)
	Errors  interface{} `json:"errors,omitempty"` // Chi tiết lỗi (nếu có)
	Meta    *Meta       `json:"meta,omitempty"`   // Metadata như pagination
}

// Meta chứa metadata như pagination
type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"per_page,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// ErrorDetail chi tiết lỗi validation
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// JSON gửi JSON response
func JSON(w http.ResponseWriter, statusCode int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Success gửi success response
// statusCode optional: nếu không truyền sẽ dùng 200
func Success(w http.ResponseWriter, lang, code string, data interface{}, statusCode ...int) {
	status := http.StatusOK
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	message := i18n.T(lang, code)

	response := Response{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}

	JSON(w, status, response)
}

// SuccessWithMeta gửi success response với metadata
// statusCode optional: nếu không truyền sẽ dùng 200
func SuccessWithMeta(w http.ResponseWriter, lang, code string, data interface{}, meta *Meta, statusCode ...int) {
	status := http.StatusOK
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	message := i18n.T(lang, code)

	response := Response{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
		Meta:    meta,
	}

	JSON(w, status, response)
}

// Created gửi response cho tạo mới thành công (201)
// statusCode optional: nếu không truyền sẽ dùng 201
func Created(w http.ResponseWriter, lang, code string, data interface{}, statusCode ...int) {
	status := http.StatusCreated
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	message := i18n.T(lang, code)

	response := Response{
		Success: true,
		Code:    code,
		Message: message,
		Data:    data,
	}

	JSON(w, status, response)
}

// NoContent gửi response không có nội dung (204)
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error gửi error response
// statusCode bắt buộc phải truyền
func Error(w http.ResponseWriter, lang, code string, errors interface{}, statusCode ...int) {
	status := http.StatusInternalServerError
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	message := i18n.T(lang, "response_codes."+code)

	response := Response{
		Success: false,
		Code:    code,
		Message: message,
		Errors:  errors,
	}

	JSON(w, status, response)
}

// BadRequest gửi bad request error (400)
// statusCode optional: nếu không truyền sẽ dùng 400
func BadRequest(w http.ResponseWriter, lang, code string, errors interface{}, statusCode ...int) {
	status := http.StatusBadRequest
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	Error(w, lang, code, errors, status)
}

// Unauthorized gửi unauthorized error (401)
// statusCode optional: nếu không truyền sẽ dùng 401
func Unauthorized(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusUnauthorized
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	Error(w, lang, code, nil, status)
}

// Forbidden gửi forbidden error (403)
// statusCode optional: nếu không truyền sẽ dùng 403
func Forbidden(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusForbidden
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	Error(w, lang, code, nil, status)
}

// NotFound gửi not found error (404)
// statusCode optional: nếu không truyền sẽ dùng 404
func NotFound(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusNotFound
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	Error(w, lang, code, nil, status)
}

// Conflict gửi conflict error (409)
// statusCode optional: nếu không truyền sẽ dùng 409
func Conflict(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusConflict
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	Error(w, lang, code, nil, status)
}

// ValidationError gửi validation error (422)
// statusCode optional: nếu không truyền sẽ dùng 422
func ValidationError(w http.ResponseWriter, lang, code string, errors interface{}, statusCode ...int) {
	status := http.StatusUnprocessableEntity
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	Error(w, lang, code, errors, status)
}

// InternalServerError gửi internal server error (500)
// statusCode optional: nếu không truyền sẽ dùng 500
func InternalServerError(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusInternalServerError
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	Error(w, lang, code, nil, status)
}

// ServiceUnavailable gửi service unavailable error (503)
// statusCode optional: nếu không truyền sẽ dùng 503
func ServiceUnavailable(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusServiceUnavailable
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	Error(w, lang, code, nil, status)
}

// GetLanguageFromRequest lấy ngôn ngữ từ request
// Thứ tự ưu tiên: query param "lang" -> header "Accept-Language" -> default "en"
func GetLanguageFromRequest(r *http.Request) string {
	// 1. Kiểm tra query parameter
	if lang := r.URL.Query().Get("lang"); lang != "" {
		if i18n.HasLanguage(lang) {
			return lang
		}
	}

	// 2. Kiểm tra Accept-Language header
	if acceptLang := r.Header.Get("Accept-Language"); acceptLang != "" {
		lang := i18n.ParseAcceptLanguage(acceptLang)
		if lang != "" && i18n.HasLanguage(lang) {
			return lang
		}
	}

	// 3. Default
	return "en"
}
