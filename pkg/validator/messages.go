package validator

import (
	"fmt"
	"strings"

	"anhnq/api-core/pkg/i18n"

	"github.com/go-playground/validator/v10"
)

// ValidationMessageManager quản lý validation messages đa ngôn ngữ
type ValidationMessageManager struct {
	i18n *i18n.Translator
}

// NewValidationMessageManager tạo instance mới
func NewValidationMessageManager(i18nInstance *i18n.Translator) *ValidationMessageManager {
	return &ValidationMessageManager{
		i18n: i18nInstance,
	}
}

// GetValidationMessage trả về validation message đa ngôn ngữ
func (vmm *ValidationMessageManager) GetValidationMessage(lang string, fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()

	// Lấy tên field đã dịch
	fieldName := vmm.getFieldName(lang, field)

	// Lấy template message
	messageTemplate := vmm.getValidationTemplate(lang, tag)
	if messageTemplate == "" {
		// Fallback về template mặc định
		messageTemplate = vmm.getValidationTemplate(lang, "invalid")
	}

	// Thay thế placeholders
	message := strings.ReplaceAll(messageTemplate, "{field}", fieldName)
	message = strings.ReplaceAll(message, "{param}", param)

	return message
}

// getFieldName trả về tên field đã dịch
func (vmm *ValidationMessageManager) getFieldName(lang string, field string) string {
	// Lấy field name từ translations với prefix từ file name
	fieldKey := fmt.Sprintf("fields.%s", field)
	fieldName := vmm.i18n.TranslateNested(lang, fieldKey)

	// Nếu không tìm thấy translation, trả về field name gốc
	if fieldName == fieldKey {
		return field
	}

	return fieldName
}

// getValidationTemplate trả về template message cho validation tag
func (vmm *ValidationMessageManager) getValidationTemplate(lang string, tag string) string {
	// Lấy validation template từ translations với prefix từ file name
	templateKey := fmt.Sprintf("validations.%s", tag)
	template := vmm.i18n.TranslateNested(lang, templateKey)

	// Nếu không tìm thấy template, trả về empty string
	if template == templateKey {
		return ""
	}

	return template
}

// GetEmptyBodyMessage trả về message cho empty body
func (vmm *ValidationMessageManager) GetEmptyBodyMessage(lang string) string {
	return vmm.i18n.TranslateNested(lang, "validations.empty_body")
}

// GetInvalidJSONMessage trả về message cho invalid JSON
func (vmm *ValidationMessageManager) GetInvalidJSONMessage(lang string) string {
	return vmm.i18n.TranslateNested(lang, "validations.invalid_json")
}

// Global instance
var messageManager *ValidationMessageManager

// InitValidationMessages khởi tạo validation message manager
func InitValidationMessages(i18nInstance *i18n.Translator) {
	messageManager = NewValidationMessageManager(i18nInstance)
}

// GetValidationMessage trả về validation message (global function)
func GetValidationMessage(lang string, fieldError validator.FieldError) string {
	if messageManager == nil {
		// Fallback về hardcoded messages nếu chưa khởi tạo
		return getFallbackMessage(fieldError)
	}
	return messageManager.GetValidationMessage(lang, fieldError)
}

// GetEmptyBodyMessage trả về empty body message (global function)
func GetEmptyBodyMessage(lang string) string {
	if messageManager == nil {
		return "Request body is required"
	}
	return messageManager.GetEmptyBodyMessage(lang)
}

// GetInvalidJSONMessage trả về invalid JSON message (global function)
func GetInvalidJSONMessage(lang string) string {
	if messageManager == nil {
		return "Invalid input data"
	}
	return messageManager.GetInvalidJSONMessage(lang)
}

// getFallbackMessage trả về fallback message khi chưa có i18n
func getFallbackMessage(fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()

	// Map field names to Vietnamese (fallback)
	fieldNames := map[string]string{
		"email":    "Email",
		"password": "Mật khẩu",
		"name":     "Tên",
		"phone":    "Số điện thoại",
		"avatar":   "Ảnh đại diện",
	}

	viField := fieldNames[field]
	if viField == "" {
		viField = field
	}

	switch tag {
	case "required":
		return fmt.Sprintf("%s là bắt buộc", viField)
	case "email":
		return fmt.Sprintf("%s không đúng định dạng", viField)
	case "min":
		return fmt.Sprintf("%s phải có ít nhất %s ký tự", viField, param)
	case "max":
		return fmt.Sprintf("%s không được vượt quá %s ký tự", viField, param)
	case "len":
		return fmt.Sprintf("%s phải có đúng %s ký tự", viField, param)
	case "gte":
		return fmt.Sprintf("%s phải lớn hơn hoặc bằng %s", viField, param)
	case "lte":
		return fmt.Sprintf("%s phải nhỏ hơn hoặc bằng %s", viField, param)
	case "gt":
		return fmt.Sprintf("%s phải lớn hơn %s", viField, param)
	case "lt":
		return fmt.Sprintf("%s phải nhỏ hơn %s", viField, param)
	case "eqfield":
		return fmt.Sprintf("%s phải bằng %s", viField, param)
	case "nefield":
		return fmt.Sprintf("%s không được bằng %s", viField, param)
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
		return fmt.Sprintf("%s phải là một trong: %s", viField, param)
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
