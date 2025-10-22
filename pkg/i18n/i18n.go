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

	// Load các file translation
	for _, lang := range cfg.Languages {
		filePath := filepath.Join(cfg.TranslationsDir, lang+".json")
		if err := t.loadTranslationFile(lang, filePath); err != nil {
			// Log warning nhưng không return error, vì có thể file không tồn tại
			fmt.Printf("Warning: Failed to load translation file %s: %v\n", filePath, err)
		}
	}

	// Kiểm tra xem có ít nhất 1 language được load không
	if len(t.translations) == 0 {
		return nil, fmt.Errorf("no translation files loaded")
	}

	return t, nil
}

// loadTranslationFile load một file translation
func (t *Translator) loadTranslationFile(lang, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse translation file %s: %w", filePath, err)
	}

	t.mu.Lock()
	t.translations[lang] = translations
	t.mu.Unlock()

	return nil
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
