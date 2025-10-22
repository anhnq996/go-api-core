package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hash password với bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword kiểm tra password với hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MD5Hash tạo MD5 hash
func MD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// SHA256Hash tạo SHA256 hash
func SHA256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// GenerateToken tạo random token
func GenerateToken(length int) string {
	return RandomString(length)
}

// GenerateAPIKey tạo API key
func GenerateAPIKey() string {
	return "ak_" + RandomString(32)
}

// GenerateSecretKey tạo secret key
func GenerateSecretKey() string {
	return "sk_" + RandomString(48)
}
