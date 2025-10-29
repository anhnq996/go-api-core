package config

import (
	"api-core/pkg/utils"
)

// ActionEventConfig cấu hình cho action events
type ActionEventConfig struct {
	LokiURL     string `json:"loki_url"`
	Environment string `json:"environment"`
	Enabled     bool   `json:"enabled"`
	DefaultJob  string `json:"default_job"`
}

// LoadActionEventConfig load action event config từ environment variables
func LoadActionEventConfig() *ActionEventConfig {
	return &ActionEventConfig{
		LokiURL:     utils.GetEnv("ACTION_EVENT_LOKI_URL", "http://localhost:3100"),
		Environment: utils.GetEnv("ACTION_EVENT_ENVIRONMENT", "development"),
		Enabled:     utils.GetEnvBool("ACTION_EVENT_ENABLED", true),
		DefaultJob:  utils.GetEnv("ACTION_EVENT_DEFAULT_JOB", "action_events"),
	}
}
