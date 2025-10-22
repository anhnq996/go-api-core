package i18n

import (
	"testing"
)

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		name       string
		acceptLang string
		want       string
	}{
		{
			name:       "simple language",
			acceptLang: "vi",
			want:       "vi",
		},
		{
			name:       "language with region",
			acceptLang: "en-US",
			want:       "en",
		},
		{
			name:       "multiple languages with quality",
			acceptLang: "vi,en-US;q=0.9,en;q=0.8",
			want:       "vi",
		},
		{
			name:       "multiple languages with quality",
			acceptLang: "en-US,en;q=0.9,vi;q=0.8",
			want:       "en",
		},
		{
			name:       "empty string",
			acceptLang: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAcceptLanguage(tt.acceptLang)
			if got != tt.want {
				t.Errorf("ParseAcceptLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslator_Translate(t *testing.T) {
	// Create a test translator
	translator := &Translator{
		translations: map[string]map[string]string{
			"en": {
				"SUCCESS": "Success",
				"HELLO":   "Hello, %s!",
			},
			"vi": {
				"SUCCESS": "Thành công",
				"HELLO":   "Xin chào, %s!",
			},
		},
		fallback: "en",
	}

	tests := []struct {
		name string
		lang string
		code string
		args []interface{}
		want string
	}{
		{
			name: "translate to English",
			lang: "en",
			code: "SUCCESS",
			want: "Success",
		},
		{
			name: "translate to Vietnamese",
			lang: "vi",
			code: "SUCCESS",
			want: "Thành công",
		},
		{
			name: "translate with parameter",
			lang: "en",
			code: "HELLO",
			args: []interface{}{"John"},
			want: "Hello, John!",
		},
		{
			name: "fallback to English when code not found",
			lang: "vi",
			code: "NOT_EXISTS",
			want: "NOT_EXISTS",
		},
		{
			name: "fallback language when lang not found",
			lang: "ja",
			code: "SUCCESS",
			want: "Success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := translator.Translate(tt.lang, tt.code, tt.args...)
			if got != tt.want {
				t.Errorf("Translate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTranslator_AddTranslations(t *testing.T) {
	translator := &Translator{
		translations: make(map[string]map[string]string),
		fallback:     "en",
	}

	// Add translations
	translator.AddTranslations("en", map[string]string{
		"KEY1": "Value 1",
		"KEY2": "Value 2",
	})

	// Verify
	if msg := translator.Translate("en", "KEY1"); msg != "Value 1" {
		t.Errorf("Expected 'Value 1', got '%s'", msg)
	}

	// Add more to same language
	translator.AddTranslations("en", map[string]string{
		"KEY3": "Value 3",
	})

	if msg := translator.Translate("en", "KEY3"); msg != "Value 3" {
		t.Errorf("Expected 'Value 3', got '%s'", msg)
	}
}

func TestTranslator_GetSupportedLanguages(t *testing.T) {
	translator := &Translator{
		translations: map[string]map[string]string{
			"en": {"SUCCESS": "Success"},
			"vi": {"SUCCESS": "Thành công"},
			"ja": {"SUCCESS": "成功"},
		},
		fallback: "en",
	}

	languages := translator.GetSupportedLanguages()

	if len(languages) != 3 {
		t.Errorf("Expected 3 languages, got %d", len(languages))
	}
}

func TestTranslator_HasLanguage(t *testing.T) {
	translator := &Translator{
		translations: map[string]map[string]string{
			"en": {"SUCCESS": "Success"},
			"vi": {"SUCCESS": "Thành công"},
		},
		fallback: "en",
	}

	if !translator.HasLanguage("en") {
		t.Error("Expected English to be supported")
	}

	if !translator.HasLanguage("vi") {
		t.Error("Expected Vietnamese to be supported")
	}

	if translator.HasLanguage("ja") {
		t.Error("Expected Japanese to NOT be supported")
	}
}
