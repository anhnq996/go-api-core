package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var Logger zerolog.Logger
var RequestLogger zerolog.Logger // Logger riêng cho requests

// LoggerManager manages dynamic loggers
type LoggerManager struct {
	defaultConfig Config
	defaultLogger zerolog.Logger
	loggers       map[string]zerolog.Logger
	mu            sync.RWMutex
}

var Manager *LoggerManager

// Config cấu hình cho logger
type Config struct {
	Level          string // debug, info, warn, error
	Output         string // console, file, loki (có thể kết hợp: "console,file,loki")
	FilePath       string // đường dẫn file log
	RequestLogPath string // đường dẫn file log cho requests (mặc định: request.log)
	LokiURL        string // Loki server URL (ví dụ: http://localhost:3100)
	EnableCaller   bool   // hiển thị file:line
	PrettyPrint    bool   // format đẹp cho console
	DailyRotation  bool   // bật daily rotation cho file logs
}

// Init khởi tạo logger với config
func Init(cfg Config) error {
	// Set error stack marshaler
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Setup output writers - parse comma-separated outputs
	var writers []io.Writer
	outputs := strings.Split(strings.ToLower(cfg.Output), ",")

	for _, output := range outputs {
		output = strings.TrimSpace(output)

		switch output {
		case "console":
			writers = append(writers, getConsoleWriter(cfg.PrettyPrint))
		case "file":
			var fileWriter io.Writer
			var err error

			if cfg.DailyRotation {
				fileWriter, err = getDailyFileWriter(cfg.FilePath)
			} else {
				fileWriter, err = getFileWriter(cfg.FilePath)
			}

			if err != nil {
				return fmt.Errorf("failed to create file writer: %w", err)
			}
			writers = append(writers, fileWriter)
		case "loki":
			if cfg.LokiURL == "" {
				return fmt.Errorf("loki URL is required when output contains loki")
			}
			lokiWriter, err := getLokiWriter(cfg.LokiURL)
			if err != nil {
				return fmt.Errorf("failed to create loki writer: %w", err)
			}
			writers = append(writers, lokiWriter)
		}
	}

	// Default to console if no valid output
	if len(writers) == 0 {
		writers = append(writers, getConsoleWriter(cfg.PrettyPrint))
	}

	multi := zerolog.MultiLevelWriter(writers...)

	// Create logger
	Logger = zerolog.New(multi).With().Timestamp().Logger()

	// Enable caller if needed
	if cfg.EnableCaller {
		Logger = Logger.With().Caller().Logger()
	}

	// Create request logger with separate Loki writer (job="request")
	var requestWriters []io.Writer
	for _, output := range outputs {
		output = strings.TrimSpace(output)

		switch output {
		case "console":
			requestWriters = append(requestWriters, getConsoleWriter(cfg.PrettyPrint))
		case "file":
			// Sử dụng RequestLogPath nếu có, nếu không thì dùng FilePath
			requestLogPath := cfg.RequestLogPath
			if requestLogPath == "" {
				// Tạo tên file request từ FilePath
				dir := filepath.Dir(cfg.FilePath)
				requestLogPath = filepath.Join(dir, "request.log")
			}

			var fileWriter io.Writer
			var err error

			fmt.Printf("🔍 Creating RequestLogger file writer: path=%s, dailyRotation=%v\n", requestLogPath, cfg.DailyRotation)

			if cfg.DailyRotation {
				fileWriter, err = getDailyFileWriter(requestLogPath)
				fmt.Printf("✅ Using DailyWriter for request logs\n")
			} else {
				fileWriter, err = getFileWriter(requestLogPath)
				fmt.Printf("⚠️ Using static file writer for request logs\n")
			}

			if err != nil {
				return fmt.Errorf("failed to create request file writer: %w", err)
			}
			requestWriters = append(requestWriters, fileWriter)
		case "loki":
			if cfg.LokiURL == "" {
				return fmt.Errorf("loki URL is required when output contains loki")
			}
			// Loki writer riêng với job="request"
			lokiWriter, err := getLokiWriterWithJob(cfg.LokiURL, "request")
			if err != nil {
				return fmt.Errorf("failed to create request loki writer: %w", err)
			}
			requestWriters = append(requestWriters, lokiWriter)
		}
	}

	if len(requestWriters) == 0 {
		requestWriters = append(requestWriters, getConsoleWriter(cfg.PrettyPrint))
	}

	multiRequest := zerolog.MultiLevelWriter(requestWriters...)
	RequestLogger = zerolog.New(multiRequest).With().Timestamp().Logger()

	if cfg.EnableCaller {
		RequestLogger = RequestLogger.With().Caller().Logger()
	}

	// Log initialization success
	fmt.Printf("✅ RequestLogger initialized with %d writers (DailyRotation: %v)\n", len(requestWriters), cfg.DailyRotation)
	if cfg.DailyRotation {
		fmt.Printf("✅ Request logs will be saved to: %s\n", cfg.RequestLogPath)
		fmt.Printf("✅ Daily rotation enabled - files will be: %s-YYYY-MM-DD.log\n", cfg.RequestLogPath)
	} else {
		fmt.Printf("⚠️ Daily rotation DISABLED - using static file: %s\n", cfg.RequestLogPath)
	}

	// Initialize dynamic logger
	InitDynamic(cfg, Logger)

	return nil
}

// getConsoleWriter tạo console writer với màu sắc
func getConsoleWriter(prettyPrint bool) io.Writer {
	if prettyPrint {
		return zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
			NoColor:    false,
		}
	}
	return os.Stdout
}

// getFileWriter tạo file writer
func getFileWriter(filePath string) (io.Writer, error) {
	if filePath == "" {
		filePath = "storages/logs/app.log"
	}

	// Tạo directory nếu chưa tồn tại
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Mở file log
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// getLokiWriter tạo Loki writer với job="apicore"
func getLokiWriter(lokiURL string) (io.Writer, error) {
	return getLokiWriterWithJob(lokiURL, "apicore")
}

// getLokiWriterWithJob tạo Loki writer với custom job label
func getLokiWriterWithJob(lokiURL, job string) (io.Writer, error) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	return &lokiWriter{
		lokiURL: lokiURL,
		labels: map[string]string{
			"job":         job,
			"environment": "development",
			"host":        hostname,
		},
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

// lokiWriter implements io.Writer interface for Loki
type lokiWriter struct {
	lokiURL    string
	labels     map[string]string
	httpClient *http.Client
}

// LokiPushRequest represents the Loki push API request
type LokiPushRequest struct {
	Streams []LokiStream `json:"streams"`
}

// LokiStream represents a log stream
type LokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

func (w *lokiWriter) Write(p []byte) (n int, err error) {
	// Prepare log entry
	timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
	logLine := string(p)

	// Prepare Loki push request
	pushReq := LokiPushRequest{
		Streams: []LokiStream{
			{
				Stream: w.labels,
				Values: [][]string{
					{timestamp, logLine},
				},
			},
		},
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(pushReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loki: failed to marshal: %v\n", err)
		return len(p), nil // Don't fail the write
	}

	// Send to Loki
	url := w.lokiURL + "/loki/api/v1/push"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loki: failed to create request: %v\n", err)
		return len(p), nil
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Loki: failed to send: %v\n", err)
		return len(p), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintf(os.Stderr, "Loki: bad status %d: %s\n", resp.StatusCode, string(body))
	}

	return len(p), nil
}

// Helper functions để log dễ dàng hơn

// Debug log debug message
func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

// Debugf log debug message with format
func Debugf(format string, v ...interface{}) {
	Logger.Debug().Msgf(format, v...)
}

// Info log info message
func Info(msg string) {
	Logger.Info().Msg(msg)
}

// Infof log info message with format
func Infof(format string, v ...interface{}) {
	Logger.Info().Msgf(format, v...)
}

// Warn log warning message
func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

// Warnf log warning message with format
func Warnf(format string, v ...interface{}) {
	Logger.Warn().Msgf(format, v...)
}

// Error log error message
func Error(msg string) {
	Logger.Error().Msg(msg)
}

// Errorf log error message with format
func Errorf(format string, v ...interface{}) {
	Logger.Error().Msgf(format, v...)
}

// ErrorWithErr log error with error object
func ErrorWithErr(err error, msg string) {
	Logger.Error().Err(err).Msg(msg)
}

// Fatal log fatal message and exit
func Fatal(msg string) {
	Logger.Fatal().Msg(msg)
}

// Fatalf log fatal message with format and exit
func Fatalf(format string, v ...interface{}) {
	Logger.Fatal().Msgf(format, v...)
}

// WithFields tạo logger với fields
func WithFields(fields map[string]interface{}) *zerolog.Logger {
	ctx := Logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	log := ctx.Logger()
	return &log
}

// WithField tạo logger với 1 field
func WithField(key string, value interface{}) *zerolog.Logger {
	log := Logger.With().Interface(key, value).Logger()
	return &log
}
