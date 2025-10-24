package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"anhnq/api-core/pkg/logger"
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
	config := &LoggerConfig{
		// Default values
		Level:         "debug",
		Output:        "console,file",
		LogPath:       "storages/logs",
		LokiURL:       "http://localhost:3100",
		EnableCaller:  false,
		PrettyPrint:   true,
		DailyRotation: true,
	}

	// Load from environment variables
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		config.Level = level
	}

	if output := os.Getenv("LOG_OUTPUT"); output != "" {
		config.Output = output
	}

	if logPath := os.Getenv("LOG_PATH"); logPath != "" {
		config.LogPath = logPath
	}

	if lokiURL := os.Getenv("LOG_LOKI_URL"); lokiURL != "" {
		config.LokiURL = lokiURL
	}

	if enableCaller := os.Getenv("LOG_ENABLE_CALLER"); enableCaller != "" {
		if parsed, err := strconv.ParseBool(enableCaller); err == nil {
			config.EnableCaller = parsed
		}
	}

	if prettyPrint := os.Getenv("LOG_PRETTY_PRINT"); prettyPrint != "" {
		if parsed, err := strconv.ParseBool(prettyPrint); err == nil {
			config.PrettyPrint = parsed
		}
	}

	if dailyRotation := os.Getenv("LOG_DAILY_ROTATION"); dailyRotation != "" {
		if parsed, err := strconv.ParseBool(dailyRotation); err == nil {
			config.DailyRotation = parsed
		}
	}

	return config
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
