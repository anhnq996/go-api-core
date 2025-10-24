package config

import (
	"os"
	"strings"
)

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// GetDefaultCORSConfig returns default CORS configuration
func GetDefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false, // Set to false when using wildcard origins
		MaxAge:           300,
	}
}

// LoadCORSConfig loads CORS configuration from environment variables
func LoadCORSConfig() *CORSConfig {
	config := GetDefaultCORSConfig()

	// Load allowed origins from environment
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		config.AllowedOrigins = strings.Split(origins, ",")
	}

	// Load allowed methods from environment
	if methods := os.Getenv("CORS_ALLOWED_METHODS"); methods != "" {
		config.AllowedMethods = strings.Split(methods, ",")
	}

	// Load allowed headers from environment
	if headers := os.Getenv("CORS_ALLOWED_HEADERS"); headers != "" {
		config.AllowedHeaders = strings.Split(headers, ",")
	}

	// Load exposed headers from environment
	if exposedHeaders := os.Getenv("CORS_EXPOSED_HEADERS"); exposedHeaders != "" {
		config.ExposedHeaders = strings.Split(exposedHeaders, ",")
	}

	// Load allow credentials from environment
	if allowCredentials := os.Getenv("CORS_ALLOW_CREDENTIALS"); allowCredentials != "" {
		config.AllowCredentials = allowCredentials == "true"
	}

	// Load max age from environment
	if maxAge := os.Getenv("CORS_MAX_AGE"); maxAge != "" {
		config.MaxAge = parseInt(maxAge, 300)
	}

	return config
}

// parseInt parses string to int with default value
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}

	// Simple parseInt implementation
	result := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return defaultValue
		}
	}
	return result
}
