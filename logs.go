package accesslog

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Logger struct holds the logging configuration
type Logger struct {
	logFormat string
	logFile   *os.File
	logger    *log.Logger
}

// NewLogger creates and returns a new Logger instance
func NewLogger(logFilePath string, logFormat string) (*Logger, error) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("could not open log file: %v", err)
	}

	logger := log.New(logFile, "", log.LstdFlags)

	return &Logger{
		logFormat: logFormat,
		logFile:   logFile,
		logger:    logger,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	return l.logFile.Close()
}

// LogRequest is the middleware function that logs the details of an HTTP request
func (l *Logger) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture the status code
		rec := &statusCodeResponseWriter{ResponseWriter: w}

		// Log request details before calling the next handler
		l.logRequestDetails(r, rec.statusCode, time.Since(start), "START")

		// Call the next handler (actual request processing)
		next.ServeHTTP(rec, r)

		// After the handler finishes, log the response status code and duration
		duration := time.Since(start)
		l.logRequestDetails(r, rec.statusCode, duration, "END")
	})
}

// logRequestDetails logs the request's details in a formatted string
func (l *Logger) logRequestDetails(r *http.Request, statusCode int, duration time.Duration, stage string) {
	timestamp := time.Now().Format(time.RFC1123)

	// Default log format: "TIMESTAMP STAGE IP METHOD PATH STATUS_CODE DURATION"
	var logMessage string
	if l.logFormat == "json" {
		logMessage = fmt.Sprintf(`{"timestamp": "%s", "stage": "%s", "client_ip": "%s", "method": "%s", "path": "%s", "status_code": %d, "duration": "%s"}`,
			timestamp, stage, r.RemoteAddr, r.Method, r.URL.Path, statusCode, duration)
	} else {
		logMessage = fmt.Sprintf("%s [%s] %s %s %s %d %s", timestamp, stage, r.RemoteAddr, r.Method, r.URL.Path, statusCode, duration)
	}

	l.logger.Println(logMessage)
}

// statusCodeResponseWriter captures the status code of the HTTP response
type statusCodeResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the HTTP status code
func (w *statusCodeResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

