package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Slug tạo slug từ string (ví dụ: "Hello World" -> "hello-world")
func Slug(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Remove accents
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ = transform.String(t, s)

	// Replace non-alphanumeric with dash
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")

	// Remove leading/trailing dashes
	s = strings.Trim(s, "-")

	return s
}

// RandomString tạo random string với độ dài n
func RandomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:n]
}

// RandomNumericString tạo random numeric string
func RandomNumericString(n int) string {
	const digits = "0123456789"
	b := make([]byte, n)
	rand.Read(b)

	for i := range b {
		b[i] = digits[b[i]%byte(len(digits))]
	}

	return string(b)
}

// Truncate cắt string về độ dài tối đa, thêm suffix nếu cần
func Truncate(s string, maxLen int, suffix string) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-len(suffix)] + suffix
}

// CamelToSnake chuyển CamelCase sang snake_case
func CamelToSnake(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// SnakeToCamel chuyển snake_case sang CamelCase
func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}
	return strings.Join(parts, "")
}

// Contains kiểm tra string có trong slice không
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase kiểm tra string có trong slice không (ignore case)
func ContainsIgnoreCase(slice []string, item string) bool {
	item = strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == item {
			return true
		}
	}
	return false
}

// IsEmpty kiểm tra string có empty không
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotEmpty kiểm tra string có empty không
func IsNotEmpty(s string) bool {
	return !IsEmpty(s)
}

// Mask che giấu một phần string (ví dụ: email, phone)
func Mask(s string, visible int, mask rune) string {
	if len(s) <= visible {
		return s
	}

	runes := []rune(s)
	for i := visible; i < len(runes); i++ {
		runes[i] = mask
	}

	return string(runes)
}

// MaskEmail che giấu email (example@gmail.com -> ex*****@gmail.com)
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	localPart := parts[0]
	if len(localPart) > 2 {
		localPart = localPart[:2] + strings.Repeat("*", len(localPart)-2)
	}

	return localPart + "@" + parts[1]
}

// MaskPhone che giấu số điện thoại (0123456789 -> 012****789)
func MaskPhone(phone string) string {
	if len(phone) <= 6 {
		return phone
	}

	return phone[:3] + strings.Repeat("*", len(phone)-6) + phone[len(phone)-3:]
}

// FirstN lấy n ký tự đầu
func FirstN(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}

// LastN lấy n ký tự cuối
func LastN(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[len(runes)-n:])
}

// ReverseString đảo ngược string
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// PadLeft thêm padding bên trái
func PadLeft(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(pad, length-len(s)) + s
}

// PadRight thêm padding bên phải
func PadRight(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(pad, length-len(s))
}

// RemoveWhitespace xóa tất cả whitespace
func RemoveWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

// FormatPhoneVN format số điện thoại VN (0123456789 -> 0123 456 789)
func FormatPhoneVN(phone string) string {
	phone = RemoveWhitespace(phone)
	if len(phone) != 10 {
		return phone
	}
	return fmt.Sprintf("%s %s %s", phone[:4], phone[4:7], phone[7:])
}
