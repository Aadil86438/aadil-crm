package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// LogViewerHandler handles log file browsing, viewing, and downloading
type LogViewerHandler struct {
	logDir string
}

// NewLogViewerHandler creates a new log viewer handler
func NewLogViewerHandler() *LogViewerHandler {
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "./logs"
	}
	// Ensure the logs directory exists
	os.MkdirAll(logDir, 0755)
	return &LogViewerHandler{logDir: logDir}
}

// LogFileInfo represents metadata about a single log file
type LogFileInfo struct {
	FileName     string `json:"file_name"`
	Date         string `json:"date"`
	SizeBytes    int64  `json:"size_bytes"`
	SizeDisplay  string `json:"size_display"`
	LineCount    int    `json:"line_count"`
	LastModified string `json:"last_modified"`
}

// LogStats represents overall logging statistics
type LogStats struct {
	TotalFiles     int    `json:"total_files"`
	TotalSizeBytes int64  `json:"total_size_bytes"`
	TotalSizeDisp  string `json:"total_size_display"`
	OldestLog      string `json:"oldest_log"`
	NewestLog      string `json:"newest_log"`
	TodayRequests  int    `json:"today_requests"`
	TodayErrors    int    `json:"today_errors"`
}

// ListLogs returns a list of all available log files with metadata
func (h *LogViewerHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	files, err := ioutil.ReadDir(h.logDir)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"files": []LogFileInfo{},
				"stats": LogStats{},
			},
		})
		return
	}

	var logFiles []LogFileInfo
	var totalSize int64
	var oldestDate, newestDate string
	todayRequests := 0
	todayErrors := 0
	todayStr := time.Now().Format("2006-01-02")

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".txt") {
			continue
		}

		filePath := filepath.Join(h.logDir, f.Name())
		content, _ := ioutil.ReadFile(filePath)
		lineCount := 0
		if len(content) > 0 {
			lineCount = strings.Count(string(content), "\n")
			if !strings.HasSuffix(string(content), "\n") {
				lineCount++
			}
		}

		// Extract date from filename (format: crm-YYYY-MM-DD.txt)
		date := strings.TrimPrefix(f.Name(), "crm-")
		date = strings.TrimSuffix(date, ".txt")

		sizeDisplay := formatSize(f.Size())

		logFile := LogFileInfo{
			FileName:     f.Name(),
			Date:         date,
			SizeBytes:    f.Size(),
			SizeDisplay:  sizeDisplay,
			LineCount:    lineCount,
			LastModified: f.ModTime().Format("2006-01-02 15:04:05"),
		}
		logFiles = append(logFiles, logFile)
		totalSize += f.Size()

		if oldestDate == "" || date < oldestDate {
			oldestDate = date
		}
		if newestDate == "" || date > newestDate {
			newestDate = date
		}

		// Count today's requests and errors
		if date == todayStr {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					todayRequests++
					if strings.Contains(line, " 4") || strings.Contains(line, " 5") {
						// Check for 4xx or 5xx status codes
						parts := strings.Fields(line)
						for _, p := range parts {
							if len(p) == 3 && (strings.HasPrefix(p, "4") || strings.HasPrefix(p, "5")) {
								todayErrors++
								break
							}
						}
					}
				}
			}
		}
	}

	// Sort by date descending (newest first)
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].Date > logFiles[j].Date
	})

	stats := LogStats{
		TotalFiles:     len(logFiles),
		TotalSizeBytes: totalSize,
		TotalSizeDisp:  formatSize(totalSize),
		OldestLog:      oldestDate,
		NewestLog:      newestDate,
		TodayRequests:  todayRequests,
		TodayErrors:    todayErrors,
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"files": logFiles,
			"stats": stats,
		},
	})
}

// ViewLog returns the content of a specific log file
func (h *LogViewerHandler) ViewLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "Missing 'date' query parameter (format: YYYY-MM-DD)",
		})
		return
	}

	fileName := fmt.Sprintf("crm-%s.txt", date)
	filePath := filepath.Join(h.logDir, fileName)

	// Security: prevent directory traversal
	absLogDir, _ := filepath.Abs(h.logDir)
	absFile, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFile, absLogDir) {
		writeJSON(w, http.StatusForbidden, map[string]interface{}{
			"success": false,
			"message": "Access denied",
		})
		return
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("No log file found for date: %s", date),
		})
		return
	}

	// Parse log lines into structured data
	lines := strings.Split(string(content), "\n")
	var logEntries []map[string]interface{}
	apiCounts := make(map[string]int)
	methodCounts := make(map[string]int)
	statusCounts := make(map[string]int)
	totalRequests := 0
	errorCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		totalRequests++

		entry := map[string]interface{}{
			"raw": trimmed,
		}

		// Parse structured log: [TIMESTAMP] METHOD PATH FROM STATUS DURATION
		// Example: [2026-09-07 14:22:01] GET /api/leads 192.168.1.1 200 12ms
		if strings.HasPrefix(trimmed, "[") {
			closeBracket := strings.Index(trimmed, "]")
			if closeBracket > 0 {
				entry["timestamp"] = strings.Trim(trimmed[1:closeBracket], " ")
				rest := strings.TrimSpace(trimmed[closeBracket+1:])
				parts := strings.Fields(rest)
				if len(parts) >= 4 {
					entry["method"] = parts[0]
					entry["path"] = parts[1]
					entry["remote_addr"] = parts[2]
					entry["status"] = parts[3]
					methodCounts[parts[0]]++
					apiCounts[parts[1]]++

					// Group status codes
					if len(parts[3]) == 3 {
						prefix := string(parts[3][0])
						switch prefix {
						case "2":
							statusCounts["2xx"]++
						case "3":
							statusCounts["3xx"]++
						case "4":
							statusCounts["4xx"]++
							errorCount++
						case "5":
							statusCounts["5xx"]++
							errorCount++
						}
					}

					if len(parts) >= 5 {
						entry["duration"] = parts[4]
					}
				}
			}
		}

		logEntries = append(logEntries, entry)
	}

	// Get top 10 most called APIs
	type apiHit struct {
		Path  string `json:"path"`
		Count int    `json:"count"`
	}
	var topAPIs []apiHit
	for path, count := range apiCounts {
		topAPIs = append(topAPIs, apiHit{Path: path, Count: count})
	}
	sort.Slice(topAPIs, func(i, j int) bool {
		return topAPIs[i].Count > topAPIs[j].Count
	})
	if len(topAPIs) > 10 {
		topAPIs = topAPIs[:10]
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"date":           date,
			"file_name":      fileName,
			"total_requests": totalRequests,
			"error_count":    errorCount,
			"entries":        logEntries,
			"top_apis":       topAPIs,
			"method_counts":  methodCounts,
			"status_counts":  statusCounts,
		},
	})
}

// DownloadLog serves a log file as a downloadable attachment
func (h *LogViewerHandler) DownloadLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "Missing date parameter", http.StatusBadRequest)
		return
	}

	fileName := fmt.Sprintf("crm-%s.txt", date)
	filePath := filepath.Join(h.logDir, fileName)

	// Security: prevent directory traversal
	absLogDir, _ := filepath.Abs(h.logDir)
	absFile, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFile, absLogDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		http.Error(w, "Log file not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
	w.Write(content)
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	} else {
		return fmt.Sprintf("%.2f MB", float64(bytes)/(1024*1024))
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
