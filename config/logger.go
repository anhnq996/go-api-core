package config

import (
	"fmt"
	"strings"

	"anhnq/api-core/pkg/logger"
	"anhnq/api-core/pkg/utils"
)

// LoggerConfig cấu hình cho logger
type LoggerConfig struct {
	Level         string // debug, info, warn, error
	Output        string // console, file, loki (có thể kết hợp)
	LogPath       string // đường dẫn thư mục chứa logs
	LokiURL       string // Loki server URL
	EnableCaller  bool   // hiển thị file:line
	PrettyPrint   bool   // format đẹp cho console
	DailyRotation bool   // bật daily rotation
}

// LoadLoggerConfig load logger config từ environment variables
func LoadLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:         utils.GetEnv("LOG_LEVEL", "debug"),
		Output:        utils.GetEnv("LOG_OUTPUT", "console,file"),
		LogPath:       utils.GetEnv("LOG_PATH", "storages/logs"),
		LokiURL:       utils.GetEnv("LOG_LOKI_URL", "http://localhost:3100"),
		EnableCaller:  utils.GetEnvBool("LOG_ENABLE_CALLER", false),
		PrettyPrint:   utils.GetEnvBool("LOG_PRETTY_PRINT", true),
		DailyRotation: utils.GetEnvBool("LOG_DAILY_ROTATION", true),
	}
}

// ValidateLoggerConfig kiểm tra config có hợp lệ không
func (c *LoggerConfig) Validate() error {
	// Validate level
	validLevels := []string{"debug", "info", "warn", "error"}
	if !contains(validLevels, c.Level) {
		return fmt.Errorf("invalid log level: %s, must be one of %v", c.Level, validLevels)
	}

	// Validate output
	validOutputs := []string{"console", "file", "loki"}
	outputs := strings.Split(strings.ToLower(c.Output), ",")
	for _, output := range outputs {
		output = strings.TrimSpace(output)
		if output != "" && !contains(validOutputs, output) {
			return fmt.Errorf("invalid log output: %s, must be one of %v", output, validOutputs)
		}
	}

	// Validate Loki URL if loki output is enabled
	if strings.Contains(strings.ToLower(c.Output), "loki") {
		if c.LokiURL == "" {
			return fmt.Errorf("loki URL is required when loki output is enabled")
		}
	}

	return nil
}

// ToLoggerConfig convert sang logger.Config
func (c *LoggerConfig) ToLoggerConfig() logger.Config {
	return logger.Config{
		Level:          c.Level,
		Output:         c.Output,
		FilePath:       c.LogPath + "/app.log",
		RequestLogPath: c.LogPath + "/request.log",
		LokiURL:        c.LokiURL,
		EnableCaller:   c.EnableCaller,
		PrettyPrint:    c.PrettyPrint,
		DailyRotation:  c.DailyRotation,
	}
}

// contains kiểm tra slice có chứa string không
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
