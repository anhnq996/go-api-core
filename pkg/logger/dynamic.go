package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

// DynamicLogger provides dynamic logger functionality
type DynamicLogger struct {
	defaultConfig Config
	defaultLogger zerolog.Logger
	loggers       map[string]zerolog.Logger
	mu            sync.RWMutex
}

// NewDynamicLogger creates a new dynamic logger
func NewDynamicLogger(defaultConfig Config, defaultLogger zerolog.Logger) *DynamicLogger {
	return &DynamicLogger{
		defaultConfig: defaultConfig,
		defaultLogger: defaultLogger,
		loggers:       make(map[string]zerolog.Logger),
	}
}

// GetLogger returns a logger for a specific job or file
func (dl *DynamicLogger) GetLogger(name string) zerolog.Logger {
	dl.mu.RLock()
	if logger, exists := dl.loggers[name]; exists {
		dl.mu.RUnlock()
		return logger
	}
	dl.mu.RUnlock()

	// Create new logger for this name
	dl.mu.Lock()
	defer dl.mu.Unlock()

	// Double-check after acquiring write lock
	if logger, exists := dl.loggers[name]; exists {
		return logger
	}

	// Create new logger with custom file path
	config := dl.defaultConfig
	config.FilePath = filepath.Join(filepath.Dir(config.FilePath), name+".log")

	logger := dl.createLogger(config)
	dl.loggers[name] = logger

	return logger
}

// SetJobLogger sets a logger for a specific job
func (dl *DynamicLogger) SetJobLogger(name string, config Config) zerolog.Logger {
	dl.mu.Lock()
	defer dl.mu.Unlock()

	logger := dl.createLogger(config)
	dl.loggers[name] = logger

	return logger
}

// createLogger creates a new logger with the given config
func (dl *DynamicLogger) createLogger(config Config) zerolog.Logger {
	// Parse log level
	level, err := zerolog.ParseLevel(config.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Setup output writers
	var writers []io.Writer
	outputs := strings.Split(strings.ToLower(config.Output), ",")

	for _, output := range outputs {
		output = strings.TrimSpace(output)
		switch output {
		case "console":
			writers = append(writers, getConsoleWriter(config.PrettyPrint))
		case "file":
			if config.FilePath != "" {
				var fileWriter io.Writer
				var err error

				if config.DailyRotation {
					fileWriter, err = getDailyFileWriter(config.FilePath)
				} else {
					fileWriter, err = getFileWriter(config.FilePath)
				}

				if err == nil {
					writers = append(writers, fileWriter)
				}
			}
		}
	}

	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	multi := zerolog.MultiLevelWriter(writers...)
	logger := zerolog.New(multi).With().Timestamp().Logger()

	if config.EnableCaller {
		logger = logger.With().Caller().Logger()
	}

	return logger.Level(level)
}

// Global dynamic logger instance
var Dynamic *DynamicLogger

// GetJobLogger returns a logger for a specific job
func GetJobLogger(name string) zerolog.Logger {
	if Dynamic == nil {
		return Logger
	}
	return Dynamic.GetLogger(name)
}

// SetJobLogger sets a logger for a specific job with custom config
func SetJobLogger(name string, config Config) zerolog.Logger {
	if Dynamic == nil {
		return Logger
	}
	return Dynamic.SetJobLogger(name, config)
}

// InitDynamic initializes the dynamic logger
func InitDynamic(config Config, defaultLogger zerolog.Logger) {
	Dynamic = NewDynamicLogger(config, defaultLogger)
}
