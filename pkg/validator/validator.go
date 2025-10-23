package validator

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
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
	// Check content type
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart form data
		return ValidateMultipartAndRespond(w, r, data)
	} else {
		// Handle JSON
		return ValidateJSONAndRespond(w, r, data)
	}
}

// ValidateJSONAndRespond validates JSON request và tự động response errors
func ValidateJSONAndRespond(w http.ResponseWriter, r *http.Request, data interface{}) bool {
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

// ValidateMultipartAndRespond validates multipart form data và tự động response errors
func ValidateMultipartAndRespond(w http.ResponseWriter, r *http.Request, data interface{}) bool {
	lang := i18n.GetLanguageFromContext(r.Context())

	// Parse multipart form (should already be parsed by controller)
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return false
	}

	// Populate struct from form values
	if err := populateStructFromForm(r, data); err != nil {
		response.BadRequest(w, lang, response.CodeBadRequest, nil)
		return false
	}

	// Validate struct
	if err := Validate(data); err != nil {
		validationErrors := ParseValidationErrors(lang, err)
		response.ValidationError(w, lang, response.CodeValidationFailed, validationErrors)
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

// populateStructFromForm populates struct from form values
func populateStructFromForm(r *http.Request, data interface{}) error {
	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("data must be a pointer to struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get json tag
		jsonTag := fieldType.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		// Remove omitempty and other options
		fieldName := strings.Split(jsonTag, ",")[0]
		if fieldName == "" {
			continue
		}

		// Get form value
		formValue := r.FormValue(fieldName)
		if formValue == "" {
			continue
		}

		// Set field value based on type
		if err := setFieldValue(field, formValue); err != nil {
			return fmt.Errorf("failed to set field %s: %w", fieldName, err)
		}
	}

	return nil
}

// setFieldValue sets field value from string
func setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return fmt.Errorf("field cannot be set")
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	case reflect.Ptr:
		// Handle pointer types (like *string for optional fields)
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setFieldValue(field.Elem(), value)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}
