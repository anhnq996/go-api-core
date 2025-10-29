# Utils Package

Package chứa các utility functions thường dùng cho backend REST API.

## Modules

### 1. String Utils (`string.go`)

Xử lý string operations.

```go
import "api-core/pkg/utils"

// Slug generation
slug := utils.Slug("Hello World Tiếng Việt")
// Output: "hello-world-tieng-viet"

// Random string
token := utils.RandomString(32)

// Random numeric string
otp := utils.RandomNumericString(6) // "123456"

// Truncate with suffix
short := utils.Truncate("Long text here", 10, "...")
// Output: "Long te..."

// Case conversion
snake := utils.CamelToSnake("HelloWorld")     // "hello_world"
camel := utils.SnakeToCamel("hello_world")    // "HelloWorld"

// Mask sensitive data
email := utils.MaskEmail("user@example.com")  // "us*****@example.com"
phone := utils.MaskPhone("0123456789")        // "012****789"

// Format phone VN
formatted := utils.FormatPhoneVN("0123456789") // "0123 456 789"

// Check contains
has := utils.Contains([]string{"a", "b"}, "a") // true
```

### 2. Hash Utils (`hash.go`)

Mã hóa và hash functions.

```go
// Hash password với bcrypt
hashed, err := utils.HashPassword("mypassword")

// Verify password
valid := utils.CheckPassword("mypassword", hashed) // true

// MD5 hash
md5 := utils.MD5Hash("hello") // "5d41402abc4b2a76b9719d911017c592"

// SHA256 hash
sha := utils.SHA256Hash("hello")

// Generate tokens
token := utils.GenerateToken(32)
apiKey := utils.GenerateAPIKey()    // "ak_xxxxx"
secret := utils.GenerateSecretKey()  // "sk_xxxxx"
```

### 3. Validation Utils (`validation.go`)

Kiểm tra validation.

```go
// Email validation
valid := utils.IsEmail("user@example.com") // true

// Phone VN validation
valid := utils.IsPhone("0123456789") // true

// URL validation
valid := utils.IsURL("https://example.com") // true

// Strong password check
strong := utils.IsStrongPassword("MyPass123!") // true

// Username validation
valid := utils.IsUsername("john_doe") // true

// Credit card validation (Luhn algorithm)
valid := utils.IsCreditCard("4532015112830366") // true

// Length checks
valid := utils.MinLength("hello", 3)              // true
valid := utils.MaxLength("hello", 10)             // true
valid := utils.LengthBetween("hello", 3, 10)      // true

// Character checks
valid := utils.IsAlpha("abc")         // true
valid := utils.IsNumeric("123")       // true
valid := utils.IsAlphanumeric("abc123") // true
```

### 4. Time Utils (`time.go`)

Xử lý thời gian.

```go
// Current time helpers
now := utils.Now()
today := utils.Today()       // 00:00:00 hôm nay
tomorrow := utils.Tomorrow()
yesterday := utils.Yesterday()

// Month/Year helpers
start := utils.StartOfMonth(time.Now())
end := utils.EndOfMonth(time.Now())
startYear := utils.StartOfYear(time.Now())
endYear := utils.EndOfYear(time.Now())

// Format
datetime := utils.FormatDateTime(time.Now()) // "2024-01-15 10:30:00"
date := utils.FormatDate(time.Now())         // "2024-01-15"
timeStr := utils.FormatTime(time.Now())      // "10:30:00"

// Parse
t, err := utils.ParseDateTime("2024-01-15 10:30:00")
t, err := utils.ParseDate("2024-01-15")

// Calculations
days := utils.DiffDays(time.Now(), future)
hours := utils.DiffHours(time.Now(), future)

age := utils.Age(birthDate)
future := utils.AddDays(time.Now(), 7)
future := utils.AddMonths(time.Now(), 3)

// Checks
isToday := utils.IsToday(time.Now())      // true
isPast := utils.IsPast(yesterday)          // true
isFuture := utils.IsFuture(tomorrow)       // true
isWeekend := utils.IsWeekend(time.Now())  // depends

// Time ago format
ago := utils.TimeAgo(pastTime) // "2 hours ago", "3 days ago"
```

### 5. Number Utils (`number.go`)

Xử lý số.

```go
// Type conversion
num := utils.ToInt("123")         // 123
num64 := utils.ToInt64("123")     // 123 (int64)
f := utils.ToFloat64("123.45")    // 123.45
str := utils.ToString(123)        // "123"

// Rounding
rounded := utils.Round(123.456, 2)      // 123.46
up := utils.RoundUp(123.456, 2)         // 123.46
down := utils.RoundDown(123.456, 2)     // 123.45

// Money formatting
money := utils.FormatMoney(1000000)        // "1,000,000"
vnd := utils.FormatMoneyVND(1000000)       // "1.000.000đ"

// Calculations
percent := utils.Percentage(25, 100)       // 25.0
change := utils.PercentageChange(100, 150) // 50.0

// Math helpers
min := utils.Min(1.5, 2.0, 3.0)           // 1.5
max := utils.Max(1.5, 2.0, 3.0)           // 3.0
sum := utils.Sum(1.0, 2.0, 3.0)           // 6.0
avg := utils.Average(1.0, 2.0, 3.0)       // 2.0

// Range checks
inRange := utils.InRange(5, 1, 10)        // true
clamped := utils.Clamp(15, 1, 10)         // 10
```

### 6. Array Utils (`array.go`)

Xử lý arrays/slices.

```go
// Remove duplicates
unique := utils.UniqueStrings([]string{"a", "b", "a"})
// ["a", "b"]

uniqueInts := utils.UniqueInts([]int{1, 2, 1})
// [1, 2]

// Filter
filtered := utils.FilterStrings([]string{"a", "ab", "abc"}, func(s string) bool {
    return len(s) > 1
})
// ["ab", "abc"]

// Map/Transform
upper := utils.MapStrings([]string{"a", "b"}, strings.ToUpper)
// ["A", "B"]

// Chunk
chunks := utils.ChunkStrings([]string{"a", "b", "c", "d"}, 2)
// [["a", "b"], ["c", "d"]]

// Reverse
reversed := utils.ReverseStrings([]string{"a", "b", "c"})
// ["c", "b", "a"]

// Contains
has := utils.ContainsInt([]int{1, 2, 3}, 2) // true

// Set operations
diff := utils.DifferenceStrings([]string{"a", "b"}, []string{"b", "c"})
// ["a"]

intersection := utils.IntersectionStrings([]string{"a", "b"}, []string{"b", "c"})
// ["b"]

union := utils.UnionStrings([]string{"a", "b"}, []string{"b", "c"})
// ["a", "b", "c"]
```

### 7. JSON Utils (`json.go`)

Xử lý JSON.

```go
// Marshal
jsonStr, err := utils.ToJSON(object)
pretty, err := utils.ToJSONPretty(object)

// Unmarshal
var result MyStruct
err := utils.FromJSON(jsonStr, &result)

// Check if valid JSON
valid := utils.IsJSON(`{"key":"value"}`) // true

// Merge JSON objects
merged, err := utils.JSONMerge(json1, json2)

// Extract field
value, err := utils.JSONExtract(jsonStr, "fieldName")

// Copy struct via JSON
err := utils.CopyStruct(source, &destination)
```

### 8. HTTP Utils (`http.go`)

Xử lý HTTP requests.

```go
// Get client IP
ip := utils.GetClientIP(r)

// Get user agent
ua := utils.GetUserAgent(r)

// Check if AJAX
isAjax := utils.IsAjax(r)

// Get query parameters
lang := utils.GetQueryParam(r, "lang", "en")
page := utils.GetQueryParamInt(r, "page", 1)

// Cookie operations
utils.SetCookie(w, "session", "value", 3600)
value := utils.GetCookie(r, "session")
utils.DeleteCookie(w, "session")

// Get Bearer token
token := utils.GetBearerToken(r)

// Set headers
utils.SetJSONContentType(w)
utils.SetNoCacheHeaders(w)
```

### 9. Pagination Utils (`pagination.go`)

Xử lý phân trang.

```go
// Create pagination
pagination := utils.NewPagination(1, 10, 100)
// page=1, perPage=10, total=100, totalPages=10

// From HTTP request
pagination := utils.PaginationFromRequest(r, total)

// Properties
offset := pagination.Offset     // For SQL OFFSET
limit := pagination.Limit        // For SQL LIMIT
hasNext := pagination.HasNextPage()
hasPrev := pagination.HasPrevPage()

// Use in SQL query
query := "SELECT * FROM users LIMIT ? OFFSET ?"
db.Raw(query, pagination.Limit, pagination.Offset).Find(&users)
```

## Examples

### Example 1: User Registration

```go
func RegisterUser(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())

    // Parse input
    var input struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        Phone    string `json:"phone"`
    }
    json.NewDecoder(r.Body).Decode(&input)

    // Validate
    errors := []response.ErrorDetail{}

    if !utils.IsEmail(input.Email) {
        errors = append(errors, response.ErrorDetail{
            Field: "email", Message: "Invalid email format",
        })
    }

    if !utils.IsStrongPassword(input.Password) {
        errors = append(errors, response.ErrorDetail{
            Field: "password", Message: "Password too weak",
        })
    }

    if !utils.IsPhone(input.Phone) {
        errors = append(errors, response.ErrorDetail{
            Field: "phone", Message: "Invalid phone number",
        })
    }

    if len(errors) > 0 {
        response.ValidationError(w, lang, response.CodeValidationFailed, errors)
        return
    }

    // Hash password
    hashed, err := utils.HashPassword(input.Password)
    if err != nil {
        response.InternalServerError(w, lang, response.CodeInternalServerError)
        return
    }

    // Create user
    user := User{
        Email:    input.Email,
        Password: hashed,
        Phone:    input.Phone,
    }

    // Save...
    response.Created(w, lang, response.CodeCreated, user)
}
```

### Example 2: Search với Pagination

```go
func SearchProducts(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())

    // Get query params
    keyword := utils.GetQueryParam(r, "q", "")

    // Get total count
    var total int64
    db.Model(&Product{}).Where("name LIKE ?", "%"+keyword+"%").Count(&total)

    // Create pagination
    pagination := utils.PaginationFromRequest(r, total)

    // Get paginated results
    var products []Product
    db.Where("name LIKE ?", "%"+keyword+"%").
        Limit(pagination.Limit).
        Offset(pagination.Offset).
        Find(&products)

    // Return with pagination meta
    response.SuccessWithMeta(w, lang, response.CodeSuccess, products, &response.Meta{
        Page:       pagination.Page,
        PerPage:    pagination.PerPage,
        Total:      pagination.Total,
        TotalPages: pagination.TotalPages,
    })
}
```

### Example 3: Generate & Send OTP

```go
func SendOTP(phone string) error {
    // Generate 6-digit OTP
    otp := utils.RandomNumericString(6)

    // Store in cache (5 minutes expiry)
    cacheKey := "otp:" + phone
    cache.Set(cacheKey, otp, 5*time.Minute)

    // Send SMS
    message := fmt.Sprintf("Your OTP is: %s", otp)
    return smsService.Send(phone, message)
}

func VerifyOTP(phone, otp string) bool {
    cacheKey := "otp:" + phone
    cached, err := cache.Get(cacheKey)

    if err != nil || cached != otp {
        return false
    }

    // Delete after verification
    cache.Delete(cacheKey)
    return true
}
```

### Example 4: File Upload with Validation

```go
func UploadFile(w http.ResponseWriter, r *http.Request) {
    lang := i18n.GetLanguageFromContext(r.Context())

    // Parse multipart form
    r.ParseMultipartForm(10 << 20) // 10 MB max

    file, handler, err := r.FormFile("file")
    if err != nil {
        response.BadRequest(w, lang, response.CodeInvalidInput, nil)
        return
    }
    defer file.Close()

    // Generate unique filename
    ext := filepath.Ext(handler.Filename)
    newName := utils.RandomString(16) + ext

    // Save file...

    response.Success(w, lang, response.CodeSuccess, map[string]string{
        "filename": newName,
        "url":      "/uploads/" + newName,
    })
}
```

## Best Practices

### 1. Always Validate User Input

```go
// ✅ Good
if !utils.IsEmail(email) {
    return errors.New("invalid email")
}

// ❌ Bad - no validation
user.Email = email
```

### 2. Use Appropriate Hash Functions

```go
// ✅ Good - bcrypt for passwords
hashed, _ := utils.HashPassword(password)

// ❌ Bad - MD5 not secure for passwords
hashed := utils.MD5Hash(password)
```

### 3. Handle Pagination Properly

```go
// ✅ Good
pagination := utils.PaginationFromRequest(r, total)
db.Limit(pagination.Limit).Offset(pagination.Offset).Find(&results)

// ❌ Bad - no pagination, can cause performance issues
db.Find(&results)
```

### 4. Mask Sensitive Data in Logs

```go
// ✅ Good
logger.Info("User login: " + utils.MaskEmail(email))

// ❌ Bad - exposing sensitive data
logger.Info("User login: " + email)
```

## Thread Safety

Tất cả functions trong utils package đều **thread-safe** và có thể sử dụng an toàn trong concurrent environments.

## Performance Tips

- `RandomString` sử dụng `crypto/rand` - secure nhưng chậm hơn `math/rand`
- `Slug` function xử lý Unicode - có thể cache nếu call nhiều lần
- Array operations tạo slice mới - consider memory usage với large datasets

## Testing

```go
func TestSlug(t *testing.T) {
    result := utils.Slug("Hello World")
    if result != "hello-world" {
        t.Errorf("Expected 'hello-world', got '%s'", result)
    }
}
```

## Contributing

Khi thêm utility functions mới:

1. Đặt vào file phù hợp theo category
2. Viết doc comments rõ ràng
3. Thêm examples vào README
4. Viết unit tests
