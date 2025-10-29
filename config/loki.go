package config

import (
	"api-core/pkg/loki"
	"api-core/pkg/utils"
)

// LokiConfig cấu hình cho Loki events
type LokiConfig struct {
	URL         string `json:"url"`
	Job         string `json:"job"`
	Environment string `json:"environment"`
	Enabled     bool   `json:"enabled"`
}

// LoadLokiConfig load Loki config từ environment variables
func LoadLokiConfig() *LokiConfig {
	return &LokiConfig{
		URL:         utils.GetEnv("LOKI_URL", "http://localhost:3100"),
		Job:         utils.GetEnv("LOKI_JOB", "action_events"),
		Environment: utils.GetEnv("LOKI_ENVIRONMENT", "development"),
		Enabled:     utils.GetEnvBool("LOKI_ENABLED", true),
	}
}

// ToLokiConfig convert sang loki.Config
func (c *LokiConfig) ToLokiConfig() loki.Config {
	return loki.Config{
		URL:         c.URL,
		Job:         c.Job,
		Environment: c.Environment,
		Labels:      make(map[string]string),
	}
}
