package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// fileLogger manages daily log file rotation
type fileLogger struct {
	mu         sync.Mutex
	logDir     string
	currentDay string
	file       *os.File
}

var dailyLogger *fileLogger

func init() {
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "./logs"
	}
	os.MkdirAll(logDir, 0755)
	dailyLogger = &fileLogger{logDir: logDir}
}

// getLogFile returns the current day's log file, creating a new one if the day has changed
func (fl *fileLogger) getLogFile() (*os.File, error) {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	today := time.Now().Format("2006-01-02")

	if fl.currentDay == today && fl.file != nil {
		return fl.file, nil
	}

	// Close previous day's file
	if fl.file != nil {
		fl.file.Close()
	}

	fileName := fmt.Sprintf("crm-%s.txt", today)
	filePath := filepath.Join(fl.logDir, fileName)

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	fl.currentDay = today
	fl.file = file
	return file, nil
}

// writeLog writes a log line to the daily log file
func (fl *fileLogger) writeLog(line string) {
	file, err := fl.getLogFile()
	if err != nil {
		log.Printf("Failed to write to log file: %v", err)
		return
	}

	fl.mu.Lock()
	defer fl.mu.Unlock()
	file.WriteString(line + "\n")
}

// Logger logs each incoming HTTP request with method, path, status, and duration.
// It writes to both stdout (console) AND a daily rotating log file.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)
		duration := time.Since(start)

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logLine := fmt.Sprintf("[%s] %s %s %s %d %v",
			timestamp,
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.status,
			duration,
		)

		// Write to stdout (console)
		log.Printf("[%s] %s %s %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			wrapped.status,
			duration,
		)

		// Write to daily log file
		dailyLogger.writeLog(logLine)
	})
}
