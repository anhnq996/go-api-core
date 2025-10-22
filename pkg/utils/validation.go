package utils

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

// IsEmail kiểm tra email hợp lệ
func IsEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// IsPhone kiểm tra số điện thoại VN hợp lệ
func IsPhone(phone string) bool {
	// Remove whitespace
	phone = RemoveWhitespace(phone)

	// Check format: 10 digits starting with 0
	matched, _ := regexp.MatchString(`^0\d{9}$`, phone)
	return matched
}

// IsURL kiểm tra URL hợp lệ
func IsURL(url string) bool {
	matched, _ := regexp.MatchString(`^https?://`, url)
	return matched
}

// IsAlphanumeric kiểm tra chỉ chứa chữ và số
func IsAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

// IsNumeric kiểm tra chỉ chứa số
func IsNumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

// IsAlpha kiểm tra chỉ chứa chữ
func IsAlpha(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// MinLength kiểm tra độ dài tối thiểu
func MinLength(s string, min int) bool {
	return len([]rune(s)) >= min
}

// MaxLength kiểm tra độ dài tối đa
func MaxLength(s string, max int) bool {
	return len([]rune(s)) <= max
}

// LengthBetween kiểm tra độ dài trong khoảng
func LengthBetween(s string, min, max int) bool {
	length := len([]rune(s))
	return length >= min && length <= max
}

// IsStrongPassword kiểm tra mật khẩu mạnh
// Ít nhất 8 ký tự, có chữ hoa, chữ thường, số và ký tự đặc biệt
func IsStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsNumber(r):
			hasNumber = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasNumber && hasSpecial
}

// IsUsername kiểm tra username hợp lệ (chữ, số, underscore, dash, 3-20 ký tự)
func IsUsername(username string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]{3,20}$`, username)
	return matched
}

// IsSlug kiểm tra slug hợp lệ (chữ thường, số, dash)
func IsSlug(slug string) bool {
	matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
	return matched
}

// IsCreditCard kiểm tra credit card hợp lệ (Luhn algorithm)
func IsCreditCard(number string) bool {
	// Remove spaces
	number = strings.ReplaceAll(number, " ", "")

	// Check only digits
	if !IsNumeric(number) {
		return false
	}

	// Check length
	if len(number) < 13 || len(number) > 19 {
		return false
	}

	// Luhn algorithm
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// IsIPv4 kiểm tra IPv4 hợp lệ
func IsIPv4(ip string) bool {
	matched, _ := regexp.MatchString(`^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`, ip)
	return matched
}

// IsHexColor kiểm tra hex color hợp lệ (#FFFFFF hoặc #FFF)
func IsHexColor(color string) bool {
	matched, _ := regexp.MatchString(`^#([A-Fa-f0-9]{6}|[A-Fa-f0-9]{3})$`, color)
	return matched
}

// IsBase64 kiểm tra base64 string hợp lệ
func IsBase64(s string) bool {
	matched, _ := regexp.MatchString(`^[A-Za-z0-9+/]*={0,2}$`, s)
	return matched && len(s)%4 == 0
}
