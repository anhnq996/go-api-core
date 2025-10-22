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
		// Parse JSON error
		if strings.Contains(err.Error(), "JSON") || strings.Contains(err.Error(), "empty") {
			response.BadRequest(w, lang, response.CodeInvalidInput, nil)
			return false
		}

		// Validation errors
		validationErrors := ParseValidationErrors(err)
		if len(validationErrors) > 0 {
			response.ValidationError(w, lang, response.CodeValidationFailed, validationErrors)
			return false
		}

		// Unknown error
		response.BadRequest(w, lang, response.CodeInvalidInput, nil)
		return false
	}

	return true
}

// ParseValidationErrors chuyển validator errors thành response.ErrorDetail
func ParseValidationErrors(err error) []response.ErrorDetail {
	var errors []response.ErrorDetail

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, response.ErrorDetail{
				Field:   e.Field(),
				Message: GetErrorMessage(e),
			})
		}
	}

	return errors
}

// GetErrorMessage trả về error message dựa trên validation tag
func GetErrorMessage(e validator.FieldError) string {
	field := e.Field()

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", field, e.Param())
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, e.Param())
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", field, e.Param())
	case "nefield":
		return fmt.Sprintf("%s must not be equal to %s", field, e.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "numeric":
		return fmt.Sprintf("%s must be a number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, e.Param())
	case "unique":
		return fmt.Sprintf("%s must be unique", field)
	case "phone":
		return fmt.Sprintf("%s must be a valid phone number", field)
	case "strongpassword":
		return fmt.Sprintf("%s must contain uppercase, lowercase, number and special character", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
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
