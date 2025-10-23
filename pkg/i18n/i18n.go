package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Translator quản lý các translations
type Translator struct {
	translations map[string]map[string]string // map[language][code]message
	fallback     string                       // fallback language
	mu           sync.RWMutex
}

var (
	defaultTranslator *Translator
	once              sync.Once
)

// Config cấu hình cho i18n
type Config struct {
	TranslationsDir string   // Thư mục chứa các file translation
	Languages       []string // Danh sách ngôn ngữ hỗ trợ
	FallbackLang    string   // Ngôn ngữ mặc định khi không tìm thấy
}

// Init khởi tạo translator
func Init(cfg Config) error {
	var err error
	once.Do(func() {
		defaultTranslator, err = NewTranslator(cfg)
	})
	return err
}

// NewTranslator tạo translator mới
func NewTranslator(cfg Config) (*Translator, error) {
	if cfg.TranslationsDir == "" {
		cfg.TranslationsDir = "translations"
	}

	if cfg.FallbackLang == "" {
		cfg.FallbackLang = "en"
	}

	if len(cfg.Languages) == 0 {
		cfg.Languages = []string{"en", "vi"}
	}

	t := &Translator{
		translations: make(map[string]map[string]string),
		fallback:     cfg.FallbackLang,
	}

	// Load các file translation từ thư mục con
	for _, lang := range cfg.Languages {
		langDir := filepath.Join(cfg.TranslationsDir, lang)
		if err := t.loadTranslationDirectory(lang, langDir); err != nil {
			// Log warning nhưng không return error, vì có thể thư mục không tồn tại
			fmt.Printf("Warning: Failed to load translation directory %s: %v\n", langDir, err)
		}
	}

	// Kiểm tra xem có ít nhất 1 language được load không
	if len(t.translations) == 0 {
		return nil, fmt.Errorf("no translation files loaded")
	}

	return t, nil
}

// loadTranslationDirectory load tất cả file translation trong một thư mục
func (t *Translator) loadTranslationDirectory(lang, dirPath string) error {
	// Kiểm tra thư mục có tồn tại không
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return fmt.Errorf("translation directory does not exist: %s", dirPath)
	}

	// Đọc tất cả file .json trong thư mục
	files, err := filepath.Glob(filepath.Join(dirPath, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read translation directory %s: %w", dirPath, err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no translation files found in directory: %s", dirPath)
	}

	// Khởi tạo map cho ngôn ngữ này
	t.mu.Lock()
	if t.translations[lang] == nil {
		t.translations[lang] = make(map[string]string)
	}
	t.mu.Unlock()

	// Load từng file
	for _, filePath := range files {
		if err := t.loadTranslationFile(lang, filePath); err != nil {
			fmt.Printf("Warning: Failed to load translation file %s: %v\n", filePath, err)
			continue
		}
	}

	return nil
}

// loadTranslationFile load một file translation
func (t *Translator) loadTranslationFile(lang, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var rawTranslations map[string]interface{}
	if err := json.Unmarshal(data, &rawTranslations); err != nil {
		return fmt.Errorf("failed to parse translation file %s: %w", filePath, err)
	}

	// Lấy prefix từ tên file (bỏ extension)
	filePrefix := t.getFilePrefix(filePath)

	// Flatten nested JSON thành flat map với prefix từ tên file
	translations := t.flattenTranslations(rawTranslations, filePrefix)

	t.mu.Lock()
	if t.translations[lang] == nil {
		t.translations[lang] = make(map[string]string)
	}

	// Merge translations vào map hiện tại
	for k, v := range translations {
		t.translations[lang][k] = v
	}
	t.mu.Unlock()

	return nil
}

// flattenTranslations chuyển nested JSON thành flat map với prefix từ tên file
func (t *Translator) flattenTranslations(data map[string]interface{}, prefix string) map[string]string {
	result := make(map[string]string)

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			result[fullKey] = v
		case map[string]interface{}:
			// Recursively flatten nested objects
			nested := t.flattenTranslations(v, fullKey)
			for k, val := range nested {
				result[k] = val
			}
		default:
			// Convert other types to string
			result[fullKey] = fmt.Sprintf("%v", v)
		}
	}

	return result
}

// getFilePrefix lấy prefix từ tên file (bỏ extension)
func (t *Translator) getFilePrefix(filePath string) string {
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	if ext != "" {
		return fileName[:len(fileName)-len(ext)]
	}
	return fileName
}

// AddTranslations thêm translations động
func (t *Translator) AddTranslations(lang string, translations map[string]string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.translations[lang] == nil {
		t.translations[lang] = make(map[string]string)
	}

	for code, message := range translations {
		t.translations[lang][code] = message
	}
}

// Translate dịch một code sang ngôn ngữ tương ứng
func (t *Translator) Translate(lang, code string, args ...interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Normalize language code
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = t.fallback
	}

	// Tìm translation trong ngôn ngữ được yêu cầu
	if langTranslations, ok := t.translations[lang]; ok {
		if message, ok := langTranslations[code]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(message, args...)
			}
			return message
		}
	}

	// Fallback sang ngôn ngữ mặc định
	if lang != t.fallback {
		if langTranslations, ok := t.translations[t.fallback]; ok {
			if message, ok := langTranslations[code]; ok {
				if len(args) > 0 {
					return fmt.Sprintf(message, args...)
				}
				return message
			}
		}
	}

	// Nếu không tìm thấy, trả về code
	return code
}

// GetSupportedLanguages trả về danh sách ngôn ngữ được hỗ trợ
func (t *Translator) GetSupportedLanguages() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	languages := make([]string, 0, len(t.translations))
	for lang := range t.translations {
		languages = append(languages, lang)
	}
	return languages
}

// HasLanguage kiểm tra xem ngôn ngữ có được hỗ trợ không
func (t *Translator) HasLanguage(lang string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	_, ok := t.translations[lang]
	return ok
}

// Helper functions sử dụng default translator

// T (Translate) dịch một code sang ngôn ngữ tương ứng
func T(lang, code string, args ...interface{}) string {
	if defaultTranslator == nil {
		return code
	}
	return defaultTranslator.Translate(lang, code, args...)
}

// GetSupportedLanguages trả về danh sách ngôn ngữ được hỗ trợ
func GetSupportedLanguages() []string {
	if defaultTranslator == nil {
		return []string{}
	}
	return defaultTranslator.GetSupportedLanguages()
}

// HasLanguage kiểm tra xem ngôn ngữ có được hỗ trợ không
func HasLanguage(lang string) bool {
	if defaultTranslator == nil {
		return false
	}
	return defaultTranslator.HasLanguage(lang)
}

// AddTranslations thêm translations động
func AddTranslations(lang string, translations map[string]string) {
	if defaultTranslator != nil {
		defaultTranslator.AddTranslations(lang, translations)
	}
}

// GetTranslator trả về default translator instance
func GetTranslator() *Translator {
	return defaultTranslator
}

// TranslateNested trả về translation cho nested key (ví dụ: "validation.required")
func (t *Translator) TranslateNested(lang, key string, args ...interface{}) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	// Normalize language code
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "" {
		lang = t.fallback
	}

	// Tìm translation trong ngôn ngữ được yêu cầu
	if langTranslations, ok := t.translations[lang]; ok {
		if message, ok := langTranslations[key]; ok {
			if len(args) > 0 {
				return fmt.Sprintf(message, args...)
			}
			return message
		}
	}

	// Fallback sang ngôn ngữ mặc định
	if lang != t.fallback {
		if langTranslations, ok := t.translations[t.fallback]; ok {
			if message, ok := langTranslations[key]; ok {
				if len(args) > 0 {
					return fmt.Sprintf(message, args...)
				}
				return message
			}
		}
	}

	// Không tìm thấy translation, trả về key
	return key
}

// ParseAcceptLanguage parse Accept-Language header
// Ví dụ: "en-US,en;q=0.9,vi;q=0.8" -> "en"
func ParseAcceptLanguage(acceptLang string) string {
	if acceptLang == "" {
		return ""
	}

	// Parse Accept-Language header
	parts := strings.Split(acceptLang, ",")
	if len(parts) == 0 {
		return ""
	}

	// Lấy language đầu tiên (có priority cao nhất)
	firstLang := strings.TrimSpace(parts[0])

	// Remove quality value nếu có (;q=0.9)
	if idx := strings.Index(firstLang, ";"); idx != -1 {
		firstLang = firstLang[:idx]
	}

	// Chỉ lấy language code, bỏ region (en-US -> en)
	if idx := strings.Index(firstLang, "-"); idx != -1 {
		firstLang = firstLang[:idx]
	}

	return strings.ToLower(strings.TrimSpace(firstLang))
}
