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
				"body": []string{GetEmptyBodyMessage(lang)},
			}
			response.ValidationError(w, lang, response.CodeValidationFailed, emptyBodyErrors)
			return false
		}

		// Invalid JSON error
		if strings.Contains(err.Error(), "JSON") {
			invalidJSONErrors := ValidationErrorsMap{
				"body": []string{GetInvalidJSONMessage(lang)},
			}
			response.ValidationError(w, lang, response.CodeValidationFailed, invalidJSONErrors)
			return false
		}

		// Validation errors
		validationErrors := ParseValidationErrors(lang, err)
		if len(validationErrors) > 0 {
			response.ValidationError(w, lang, response.CodeValidationFailed, validationErrors)
			return false
		}

		// Unknown error - cũng trả về validation errors format
		unknownErrors := ValidationErrorsMap{
			"body": []string{GetInvalidJSONMessage(lang)},
		}
		response.ValidationError(w, lang, response.CodeValidationFailed, unknownErrors)
		return false
	}

	return true
}

// ValidationErrorsMap format errors theo dạng map[field][]messages
type ValidationErrorsMap map[string][]string

// ParseValidationErrors chuyển validator errors thành ValidationErrorsMap
func ParseValidationErrors(lang string, err error) ValidationErrorsMap {
	errorsMap := make(ValidationErrorsMap)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			message := GetErrorMessage(lang, e)

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

// GetErrorMessage trả về error message đa ngôn ngữ dựa trên validation tag
func GetErrorMessage(lang string, e validator.FieldError) string {
	return GetValidationMessage(lang, e)
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
