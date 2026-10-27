package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

// LogBroadcaster manages real-time broadcast of log events to SSE subscribers
type LogBroadcaster struct {
	mu      sync.RWMutex
	clients map[chan string]bool
}

var GlobalLogBroadcaster = &LogBroadcaster{
	clients: make(map[chan string]bool),
}

func (b *LogBroadcaster) Subscribe() chan string {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan string, 100)
	b.clients[ch] = true
	return ch
}

func (b *LogBroadcaster) Unsubscribe(ch chan string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
}

func (b *LogBroadcaster) Broadcast(msg string) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default:
			// Channel full, drop to prevent blocking
		}
	}
}

// LiveLogsStreamHandler streams new log entries as Server-Sent Events (SSE)
func LiveLogsStreamHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	logChan := GlobalLogBroadcaster.Subscribe()
	defer GlobalLogBroadcaster.Unsubscribe(logChan)

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"connected\",\"service\":\"core-service-aws\"}\n\n")
	flusher.Flush()

	notify := r.Context().Done()
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-notify:
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case logJSON, ok := <-logChan:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", logJSON)
			flusher.Flush()
		}
	}
}

// Logger logs each incoming HTTP request with method, path, status, and duration.
// It writes to stdout, a daily rotating log file, and broadcasts to live SSE subscribers.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip logging SSE streaming endpoints themselves to avoid recursive loops
		if strings.Contains(r.URL.Path, "/live") {
			next.ServeHTTP(w, r)
			return
		}

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

		// Broadcast structured log entry to real-time subscribers
		payload := fmt.Sprintf(`{"raw":"%s","timestamp":"%s","method":"%s","path":"%s","remote_addr":"%s","status":"%d","duration":"%v"}`,
			escapeJSON(logLine), timestamp, r.Method, r.RequestURI, r.RemoteAddr, wrapped.status, duration)
		GlobalLogBroadcaster.Broadcast(payload)
	})
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

