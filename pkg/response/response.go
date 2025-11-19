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

// SuccessResponse tạo success response struct (dùng trong service)
func SuccessResponse(lang, code string, data interface{}) *Response {
	return &Response{
		Success: true,
		Code:    code,
		Message: i18n.T(lang, "response_codes."+code),
		Data:    data,
	}
}

// Success gửi success response
// statusCode optional: nếu không truyền sẽ dùng 200
func Success(w http.ResponseWriter, lang, code string, data interface{}, statusCode ...int) {
	status := http.StatusOK
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	JSON(w, status, *SuccessResponse(lang, code, data))
}

// SuccessResponseWithMeta tạo success response với meta (dùng trong service)
func SuccessResponseWithMeta(lang, code string, data interface{}, meta *Meta) *Response {
	resp := SuccessResponse(lang, code, data)
	resp.Meta = meta
	return resp
}

// SuccessWithMeta gửi success response với metadata
// statusCode optional: nếu không truyền sẽ dùng 200
func SuccessWithMeta(w http.ResponseWriter, lang, code string, data interface{}, meta *Meta, statusCode ...int) {
	status := http.StatusOK
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	JSON(w, status, *SuccessResponseWithMeta(lang, code, data, meta))
}

// Created gửi response cho tạo mới thành công (201)
// statusCode optional: nếu không truyền sẽ dùng 201
func Created(w http.ResponseWriter, lang, code string, data interface{}, statusCode ...int) {
	status := http.StatusCreated
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	JSON(w, status, *SuccessResponse(lang, code, data))
}

// NoContent gửi response không có nội dung (204)
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// ErrorResponse tạo error response struct (dùng trong service)
func ErrorResponse(lang, code string, errors interface{}) *Response {
	return &Response{
		Success: false,
		Code:    code,
		Message: i18n.T(lang, "response_codes."+code),
		Errors:  errors,
	}
}

// Error gửi error response
// statusCode bắt buộc phải truyền
func Error(w http.ResponseWriter, lang, code string, errors interface{}, statusCode ...int) {
	status := http.StatusInternalServerError
	if len(statusCode) > 0 {
		status = statusCode[0]
	}

	JSON(w, status, *ErrorResponse(lang, code, errors))
}

// BadRequestResponse tạo bad request response (dùng trong service)
func BadRequestResponse(lang, code string, errors interface{}) *Response {
	return ErrorResponse(lang, code, errors)
}

// BadRequest gửi bad request error (400)
// statusCode optional: nếu không truyền sẽ dùng 400
func BadRequest(w http.ResponseWriter, lang, code string, errors interface{}, statusCode ...int) {
	status := http.StatusBadRequest
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	JSON(w, status, *BadRequestResponse(lang, code, errors))
}

// UnauthorizedResponse tạo unauthorized response (dùng trong service)
func UnauthorizedResponse(lang, code string) *Response {
	return ErrorResponse(lang, code, nil)
}

// Unauthorized gửi unauthorized error (401)
// statusCode optional: nếu không truyền sẽ dùng 401
func Unauthorized(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusUnauthorized
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	JSON(w, status, *UnauthorizedResponse(lang, code))
}

// ForbiddenResponse tạo forbidden response (dùng trong service)
func ForbiddenResponse(lang, code string) *Response {
	return ErrorResponse(lang, code, nil)
}

// Forbidden gửi forbidden error (403)
// statusCode optional: nếu không truyền sẽ dùng 403
func Forbidden(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusForbidden
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	JSON(w, status, *ForbiddenResponse(lang, code))
}

// NotFoundResponse tạo not found response (dùng trong service)
func NotFoundResponse(lang, code string) *Response {
	return ErrorResponse(lang, code, nil)
}

// NotFound gửi not found error (404)
// statusCode optional: nếu không truyền sẽ dùng 404
func NotFound(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusNotFound
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	JSON(w, status, *NotFoundResponse(lang, code))
}

// ConflictResponse tạo conflict response (dùng trong service)
func ConflictResponse(lang, code string) *Response {
	return ErrorResponse(lang, code, nil)
}

// Conflict gửi conflict error (409)
// statusCode optional: nếu không truyền sẽ dùng 409
func Conflict(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusConflict
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	JSON(w, status, *ConflictResponse(lang, code))
}

// ValidationErrorResponse tạo validation error response (dùng trong service)
func ValidationErrorResponse(lang, code string, errors interface{}) *Response {
	return ErrorResponse(lang, code, errors)
}

// ValidationError gửi validation error (422)
// statusCode optional: nếu không truyền sẽ dùng 422
func ValidationError(w http.ResponseWriter, lang, code string, errors interface{}, statusCode ...int) {
	status := http.StatusUnprocessableEntity
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	JSON(w, status, *ValidationErrorResponse(lang, code, errors))
}

// InternalServerErrorResponse tạo internal server error response (dùng trong service)
func InternalServerErrorResponse(lang, code string) *Response {
	return ErrorResponse(lang, code, nil)
}

// InternalServerError gửi internal server error (500)
// statusCode optional: nếu không truyền sẽ dùng 500
func InternalServerError(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusInternalServerError
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	JSON(w, status, *InternalServerErrorResponse(lang, code))
}

// ServiceUnavailableResponse tạo service unavailable response (dùng trong service)
func ServiceUnavailableResponse(lang, code string) *Response {
	return ErrorResponse(lang, code, nil)
}

// ServiceUnavailable gửi service unavailable error (503)
// statusCode optional: nếu không truyền sẽ dùng 503
func ServiceUnavailable(w http.ResponseWriter, lang, code string, statusCode ...int) {
	status := http.StatusServiceUnavailable
	if len(statusCode) > 0 {
		status = statusCode[0]
	}
	JSON(w, status, *ServiceUnavailableResponse(lang, code))
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
