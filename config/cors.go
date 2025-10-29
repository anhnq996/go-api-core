package config

import (
	"api-core/pkg/utils"
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

// LoadCORSConfig loads CORS configuration from environment variables
func LoadCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:   utils.GetEnvStringSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		AllowedMethods:   utils.GetEnvStringSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}),
		AllowedHeaders:   utils.GetEnvStringSlice("CORS_ALLOWED_HEADERS", []string{"*"}),
		ExposedHeaders:   utils.GetEnvStringSlice("CORS_EXPOSED_HEADERS", []string{"Link"}),
		AllowCredentials: utils.GetEnvBool("CORS_ALLOW_CREDENTIALS", false),
		MaxAge:           utils.GetEnvInt("CORS_MAX_AGE", 300),
	}
}
