# I18n Package

Package đa ngôn ngữ (internationalization) hỗ trợ translate messages dựa trên language code.

## Features

- ✅ Hỗ trợ nhiều ngôn ngữ (hiện tại: en, vi)
- ✅ Load translations từ JSON files
- ✅ Fallback language khi không tìm thấy translation
- ✅ Parse Accept-Language header
- ✅ Middleware tự động detect language
- ✅ Thread-safe với sync.RWMutex
- ✅ Dynamic translations (add translations at runtime)
- ✅ Message formatting với parameters

## Cài đặt

### 1. Cấu trúc thư mục

```
translations/
  ├── en.json  # English
  ├── vi.json  # Tiếng Việt
  └── ja.json  # Japanese (optional)
```

### 2. Khởi tạo trong main.go

```go
package main

import (
    "api-core/pkg/i18n"
    "log"
)

func main() {
    // Khởi tạo i18n
    err := i18n.Init(i18n.Config{
        TranslationsDir: "translations",
        Languages:       []string{"en", "vi"},
        FallbackLang:    "en",
    })
    if err != nil {
        log.Fatal("Failed to initialize i18n:", err)
    }

    // ... rest of your application
}
```

## Sử dụng

### Basic Translation

```go
import "api-core/pkg/i18n"

// Translate một code
message := i18n.T("en", "SUCCESS")
// Output: "Operation successful"

message = i18n.T("vi", "SUCCESS")
// Output: "Thao tác thành công"
```

### Translation với Parameters

```go
// translations/en.json
{
  "WELCOME_USER": "Welcome, %s!",
  "ITEMS_FOUND": "Found %d items"
}

// translations/vi.json
{
  "WELCOME_USER": "Chào mừng, %s!",
  "ITEMS_FOUND": "Tìm thấy %d mục"
}

// Code
message := i18n.T("en", "WELCOME_USER", "John")
// Output: "Welcome, John!"

message = i18n.T("vi", "ITEMS_FOUND", 25)
// Output: "Tìm thấy 25 mục"
```

### Sử dụng Middleware

Middleware tự động detect language và lưu vào context:

```go
import (
    "github.com/go-chi/chi/v5"
    "api-core/pkg/i18n"
)

func main() {
    r := chi.NewRouter()

    // Thêm i18n middleware
    r.Use(i18n.Middleware)

    r.Get("/users", func(w http.ResponseWriter, r *http.Request) {
        // Lấy language từ context
        lang := i18n.GetLanguageFromContext(r.Context())

        message := i18n.T(lang, "SUCCESS")
        // Response với message đã được dịch
    })
}
```

### Parse Accept-Language Header

```go
// Request header: Accept-Language: vi,en-US;q=0.9,en;q=0.8
lang := i18n.ParseAcceptLanguage("vi,en-US;q=0.9,en;q=0.8")
// Output: "vi"

lang = i18n.ParseAcceptLanguage("en-US,en;q=0.9")
// Output: "en"
```

### Check Language Support

```go
// Kiểm tra ngôn ngữ có được hỗ trợ không
if i18n.HasLanguage("vi") {
    // Vietnamese is supported
}

// Lấy danh sách ngôn ngữ được hỗ trợ
languages := i18n.GetSupportedLanguages()
// Output: ["en", "vi"]
```

### Dynamic Translations

Thêm translations động lúc runtime:

```go
// Thêm translations mới
i18n.AddTranslations("en", map[string]string{
    "CUSTOM_MESSAGE": "This is a custom message",
})

i18n.AddTranslations("vi", map[string]string{
    "CUSTOM_MESSAGE": "Đây là thông điệp tùy chỉnh",
})

// Sử dụng
message := i18n.T("vi", "CUSTOM_MESSAGE")
```

### Custom Translator Instance

Nếu không muốn dùng global translator:

```go
// Tạo translator riêng
translator, err := i18n.NewTranslator(i18n.Config{
    TranslationsDir: "custom_translations",
    Languages:       []string{"en", "vi", "ja"},
    FallbackLang:    "en",
})

// Sử dụng
message := translator.Translate("ja", "SUCCESS")
```

## Format Translation Files

### en.json

```json
{
  "SUCCESS": "Operation successful",
  "WELCOME_USER": "Welcome, %s!",
  "ERROR_NOT_FOUND": "Resource not found",
  "ITEMS_COUNT": "You have %d items"
}
```

### vi.json

```json
{
  "SUCCESS": "Thao tác thành công",
  "WELCOME_USER": "Chào mừng, %s!",
  "ERROR_NOT_FOUND": "Không tìm thấy tài nguyên",
  "ITEMS_COUNT": "Bạn có %d mục"
}
```

## Language Detection Flow

Thứ tự ưu tiên detect language:

1. **Query parameter** `?lang=vi`
2. **Accept-Language header**
3. **Fallback language** (default: en)

```go
// URL: /api/users?lang=vi
// -> Language: "vi"

// Header: Accept-Language: vi,en-US;q=0.9
// -> Language: "vi"

// No lang parameter and no header
// -> Language: "en" (fallback)
```

## Thêm Ngôn Ngữ Mới

### 1. Tạo file translation

Tạo file `translations/ja.json`:

```json
{
  "SUCCESS": "操作が成功しました",
  "ERROR_NOT_FOUND": "リソースが見つかりません"
}
```

### 2. Cập nhật config

```go
i18n.Init(i18n.Config{
    TranslationsDir: "translations",
    Languages:       []string{"en", "vi", "ja"},
    FallbackLang:    "en",
})
```

### 3. Sử dụng

```go
message := i18n.T("ja", "SUCCESS")
// Output: "操作が成功しました"
```

## Best Practices

### 1. Sử dụng constants cho translation keys

```go
// constants.go
const (
    KeySuccess      = "SUCCESS"
    KeyError        = "ERROR"
    KeyNotFound     = "NOT_FOUND"
)

// Sử dụng
message := i18n.T(lang, KeySuccess)
```

### 2. Luôn cung cấp fallback

```go
// Config với fallback language
i18n.Init(i18n.Config{
    FallbackLang: "en", // Luôn set fallback
})
```

### 3. Validate language trước khi sử dụng

```go
func getUserLanguage(r *http.Request) string {
    lang := r.URL.Query().Get("lang")

    if lang == "" || !i18n.HasLanguage(lang) {
        return "en" // default
    }

    return lang
}
```

### 4. Centralize translation keys

Tập trung các translation keys vào một package:

```go
// pkg/messages/codes.go
package messages

const (
    Success        = "SUCCESS"
    Created        = "CREATED"
    NotFound       = "NOT_FOUND"
    Unauthorized   = "UNAUTHORIZED"
)
```

### 5. Log missing translations

```go
message := i18n.T(lang, code)
if message == code {
    log.Warnf("Missing translation: lang=%s, code=%s", lang, code)
}
```

## Thread Safety

Package được thiết kế thread-safe với `sync.RWMutex`:

```go
// Safe to use concurrently
go i18n.T("en", "SUCCESS")
go i18n.T("vi", "SUCCESS")
go i18n.AddTranslations("en", translations)
```

## Testing

```go
func TestTranslation(t *testing.T) {
    // Setup
    i18n.Init(i18n.Config{
        TranslationsDir: "testdata/translations",
        Languages:       []string{"en"},
        FallbackLang:    "en",
    })

    // Test
    message := i18n.T("en", "SUCCESS")
    assert.Equal(t, "Operation successful", message)
}
```

## Performance

- ✅ Translations được load vào memory khi khởi động
- ✅ Không có I/O operations trong runtime
- ✅ Fast lookup với Go maps
- ✅ Read-optimized với RWMutex

## Xem thêm

- [pkg/response](../response/README.md) - Response package tích hợp i18n
