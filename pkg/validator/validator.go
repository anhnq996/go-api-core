package validator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"anhnq/api-core/pkg/i18n"
	"anhnq/api-core/pkg/response"

	"github.com/go-playground/validator/v10"
)

// Validator global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()

	// Register custom tag name function để sử dụng json tag thay vì field name
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	// Register custom validators
	registerCustomValidators()
}

// Validate validates a struct
func Validate(data interface{}) error {
	return validate.Struct(data)
}

// ValidateRequest validates request body và tự động parse JSON
func ValidateRequest(r *http.Request, data interface{}) error {
	// Parse JSON body
	if err := json.NewDecoder(r.Body).Decode(data); err != nil {
		if err == io.EOF {
			return fmt.Errorf("request body is empty")
		}
		return fmt.Errorf("invalid JSON format")
	}

	// Validate struct
	return validate.Struct(data)
}

// ValidateAndRespond validates request và tự động response errors
// Trả về true nếu validation pass, false nếu fail (đã response error)
func ValidateAndRespond(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	lang := i18n.GetLanguageFromContext(r.Context())

	// Parse và validate
	if err := ValidateRequest(r, data); err != nil {
		// Empty body error
		if strings.Contains(err.Error(), "empty") {
			emptyBodyErrors := ValidationErrorsMap{
				"body": []string{"Request body là bắt buộc"},
			}
			response.ValidationError(w, lang, response.CodeValidationFailed, emptyBodyErrors)
			return false
		}

		// Invalid JSON error
		if strings.Contains(err.Error(), "JSON") {
			invalidJSONErrors := ValidationErrorsMap{
				"body": []string{"Dữ liệu đầu vào không hợp lệ"},
			}
			response.ValidationError(w, lang, response.CodeValidationFailed, invalidJSONErrors)
			return false
		}

		// Validation errors
		validationErrors := ParseValidationErrors(err)
		if len(validationErrors) > 0 {
			response.ValidationError(w, lang, response.CodeValidationFailed, validationErrors)
			return false
		}

		// Unknown error - cũng trả về validation errors format
		unknownErrors := ValidationErrorsMap{
			"body": []string{"Dữ liệu đầu vào không hợp lệ"},
		}
		response.ValidationError(w, lang, response.CodeValidationFailed, unknownErrors)
		return false
	}

	return true
}

// ValidationErrorsMap format errors theo dạng map[field][]messages
type ValidationErrorsMap map[string][]string

// ParseValidationErrors chuyển validator errors thành ValidationErrorsMap
func ParseValidationErrors(err error) ValidationErrorsMap {
	errorsMap := make(ValidationErrorsMap)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			message := GetErrorMessage(e)

			// Nếu field đã tồn tại, append message vào slice
			if messages, exists := errorsMap[field]; exists {
				errorsMap[field] = append(messages, message)
			} else {
				errorsMap[field] = []string{message}
			}
		}
	}

	return errorsMap
}

// GetErrorMessage trả về error message tiếng Việt dựa trên validation tag
func GetErrorMessage(e validator.FieldError) string {
	field := e.Field()

	// Map field names to Vietnamese
	fieldNames := map[string]string{
		"email":    "Email",
		"password": "Mật khẩu",
		"name":     "Tên",
		"phone":    "Số điện thoại",
		"avatar":   "Ảnh đại diện",
	}

	// Get Vietnamese field name
	viField := fieldNames[field]
	if viField == "" {
		viField = field
	}

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s là bắt buộc", viField)
	case "email":
		return fmt.Sprintf("%s không đúng định dạng", viField)
	case "min":
		return fmt.Sprintf("%s phải có ít nhất %s ký tự", viField, e.Param())
	case "max":
		return fmt.Sprintf("%s không được vượt quá %s ký tự", viField, e.Param())
	case "len":
		return fmt.Sprintf("%s phải có đúng %s ký tự", viField, e.Param())
	case "gte":
		return fmt.Sprintf("%s phải lớn hơn hoặc bằng %s", viField, e.Param())
	case "lte":
		return fmt.Sprintf("%s phải nhỏ hơn hoặc bằng %s", viField, e.Param())
	case "gt":
		return fmt.Sprintf("%s phải lớn hơn %s", viField, e.Param())
	case "lt":
		return fmt.Sprintf("%s phải nhỏ hơn %s", viField, e.Param())
	case "eqfield":
		return fmt.Sprintf("%s phải bằng %s", viField, e.Param())
	case "nefield":
		return fmt.Sprintf("%s không được bằng %s", viField, e.Param())
	case "alpha":
		return fmt.Sprintf("%s chỉ được chứa chữ cái", viField)
	case "alphanum":
		return fmt.Sprintf("%s chỉ được chứa chữ cái và số", viField)
	case "numeric":
		return fmt.Sprintf("%s phải là số", viField)
	case "url":
		return fmt.Sprintf("%s phải là URL hợp lệ", viField)
	case "uri":
		return fmt.Sprintf("%s phải là URI hợp lệ", viField)
	case "uuid":
		return fmt.Sprintf("%s phải là UUID hợp lệ", viField)
	case "oneof":
		return fmt.Sprintf("%s phải là một trong: %s", viField, e.Param())
	case "unique":
		return fmt.Sprintf("%s phải là duy nhất", viField)
	case "phone":
		return fmt.Sprintf("%s phải là số điện thoại hợp lệ", viField)
	case "strongpassword":
		return fmt.Sprintf("%s phải chứa chữ hoa, chữ thường, số và ký tự đặc biệt", viField)
	default:
		return fmt.Sprintf("%s không hợp lệ", viField)
	}
}

// registerCustomValidators đăng ký custom validators
func registerCustomValidators() {
	// Phone number validator (Vietnamese format)
	validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		phone := fl.Field().String()
		// Remove spaces
		phone = strings.ReplaceAll(phone, " ", "")
		// Check format: 10 digits starting with 0
		if len(phone) != 10 {
			return false
		}
		if phone[0] != '0' {
			return false
		}
		for _, c := range phone {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	})

	// Strong password validator
	validate.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()

		if len(password) < 8 {
			return false
		}

		var (
			hasUpper   = false
			hasLower   = false
			hasNumber  = false
			hasSpecial = false
		)

		for _, c := range password {
			switch {
			case c >= 'A' && c <= 'Z':
				hasUpper = true
			case c >= 'a' && c <= 'z':
				hasLower = true
			case c >= '0' && c <= '9':
				hasNumber = true
			case c == '!' || c == '@' || c == '#' || c == '$' || c == '%' || c == '^' || c == '&' || c == '*':
				hasSpecial = true
			}
		}

		return hasUpper && hasLower && hasNumber && hasSpecial
	})
}

// GetValidator trả về validator instance
func GetValidator() *validator.Validate {
	return validate
}
