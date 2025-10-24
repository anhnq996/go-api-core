package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DailyWriter writes logs to daily rotated files
type DailyWriter struct {
	basePath string
	current  *os.File
	date     string
	mu       sync.Mutex
}

// NewDailyWriter creates a new daily writer
func NewDailyWriter(basePath string) (*DailyWriter, error) {
	// Tạo directory nếu chưa tồn tại
	dir := filepath.Dir(basePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	dw := &DailyWriter{
		basePath: basePath,
	}

	// Khởi tạo file đầu tiên
	if err := dw.rotate(); err != nil {
		return nil, err
	}

	return dw, nil
}

// Write implements io.Writer interface
func (dw *DailyWriter) Write(p []byte) (n int, err error) {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	// Kiểm tra xem có cần rotate không
	today := time.Now().Format("2006-01-02")
	if dw.date != today {
		if err := dw.rotate(); err != nil {
			return 0, err
		}
	}

	return dw.current.Write(p)
}

// Close closes the current file
func (dw *DailyWriter) Close() error {
	dw.mu.Lock()
	defer dw.mu.Unlock()

	if dw.current != nil {
		return dw.current.Close()
	}
	return nil
}

// rotate creates a new file for the current date
func (dw *DailyWriter) rotate() error {
	// Đóng file cũ nếu có
	if dw.current != nil {
		dw.current.Close()
	}

	// Tạo tên file mới với ngày
	today := time.Now().Format("2006-01-02")
	dw.date = today

	// Tạo tên file với ngày
	ext := filepath.Ext(dw.basePath)
	baseName := strings.TrimSuffix(dw.basePath, ext)
	fileName := fmt.Sprintf("%s-%s%s", baseName, today, ext)

	// Mở file mới
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}

	dw.current = file
	return nil
}

// getDailyFileWriter tạo daily file writer
func getDailyFileWriter(filePath string) (io.Writer, error) {
	if filePath == "" {
		filePath = "storages/logs/app.log"
	}

	return NewDailyWriter(filePath)
}
