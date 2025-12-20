package profiler

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

// LogEntry represents a single log entry with structured data
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	QueryID     string                 `json:"query_id,omitempty"`
	ExecutionMS float64                `json:"execution_ms,omitempty"`
	RowsAffected int64                 `json:"rows_affected,omitempty"`
	QueryHash   string                 `json:"query_hash,omitempty"`
	Query       string                 `json:"query,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Logger handles structured logging with file-based persistence
// Designed to be extensible for cloud logging services (CloudWatch, DataDog, etc.)
type Logger struct {
	mu              sync.Mutex
	logFile         *os.File
	logPath         string
	minLevel        LogLevel
	enableConsole   bool
	enableFile      bool
	maxFileSize     int64 // Max size in bytes before rotation
	rotationCount   int   // Number of rotated files to keep
	levelPriority   map[LogLevel]int
}

// NewLogger creates a new Logger instance with file-based logging
// logDir: directory to store log files (defaults to ./logs)
// enableConsole: whether to also print to console
func NewLogger(logDir string, enableConsole bool) (*Logger, error) {
	if logDir == "" {
		logDir = "./logs"
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "profiler.log")

	// Open log file in append mode
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := &Logger{
		logFile:       file,
		logPath:       logPath,
		minLevel:      INFO,
		enableConsole: enableConsole,
		enableFile:    true,
		maxFileSize:   100 * 1024 * 1024, // 100MB default
		rotationCount: 5,
		levelPriority: map[LogLevel]int{
			DEBUG: 0,
			INFO:  1,
			WARN:  2,
			ERROR: 3,
		},
	}

	return logger, nil
}

// SetMinLevel sets the minimum log level to be recorded
func (l *Logger) SetMinLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.minLevel = level
}

// shouldLog checks if a log level should be recorded based on minLevel
func (l *Logger) shouldLog(level LogLevel) bool {
	return l.levelPriority[level] >= l.levelPriority[l.minLevel]
}

// rotateFile performs log file rotation when size limit is exceeded
func (l *Logger) rotateFile() error {
	if err := l.logFile.Close(); err != nil {
		return err
	}

	// Rotate files: profiler.log.4 -> profiler.log.5, ... profiler.log.1 -> profiler.log.2
	for i := l.rotationCount - 1; i >= 1; i-- {
		oldName := fmt.Sprintf("%s.%d", l.logPath, i)
		newName := fmt.Sprintf("%s.%d", l.logPath, i+1)
		os.Rename(oldName, newName)
	}

	// Move current log to profiler.log.1
	if err := os.Rename(l.logPath, fmt.Sprintf("%s.1", l.logPath)); err != nil {
		return err
	}

	// Open new log file
	file, err := os.OpenFile(l.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.logFile = file
	return nil
}

// log writes a log entry to file and optionally to console
func (l *Logger) log(entry LogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.shouldLog(entry.Level) {
		return
	}

	// Marshal to JSON
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal log entry: %v\n", err)
		return
	}

	// Write to file
	if l.enableFile && l.logFile != nil {
		// Check file size and rotate if needed
		info, err := l.logFile.Stat()
		if err == nil && info.Size() > l.maxFileSize {
			if err := l.rotateFile(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to rotate log file: %v\n", err)
			}
		}

		if _, err := l.logFile.Write(append(data, '\n')); err != nil {
			fmt.Fprintf(os.Stderr, "failed to write log: %v\n", err)
		}
	}

	// Write to console
	if l.enableConsole {
		fmt.Printf("[%s] %s - %s\n", entry.Level, entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Message)
	}
}

// Debug logs a debug-level message
func (l *Logger) Debug(msg string, metadata map[string]interface{}) {
	l.log(LogEntry{
		Timestamp: time.Now(),
		Level:     DEBUG,
		Message:   msg,
		Metadata:  metadata,
	})
}

// Info logs an info-level message
func (l *Logger) Info(msg string, metadata map[string]interface{}) {
	l.log(LogEntry{
		Timestamp: time.Now(),
		Level:     INFO,
		Message:   msg,
		Metadata:  metadata,
	})
}

// Warn logs a warning-level message
func (l *Logger) Warn(msg string, metadata map[string]interface{}) {
	l.log(LogEntry{
		Timestamp: time.Now(),
		Level:     WARN,
		Message:   msg,
		Metadata:  metadata,
	})
}

// Error logs an error-level message
func (l *Logger) Error(msg string, err error, metadata map[string]interface{}) {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	l.log(LogEntry{
		Timestamp: time.Now(),
		Level:     ERROR,
		Message:   msg,
		Error:     errMsg,
		Metadata:  metadata,
	})
}

// LogQueryExecution logs a query execution with metrics
func (l *Logger) LogQueryExecution(queryID string, query string, queryHash string, executionMS float64, rowsAffected int64, err error) {
	entry := LogEntry{
		Timestamp:    time.Now(),
		Level:        INFO,
		Message:      "Query executed",
		QueryID:      queryID,
		ExecutionMS:  executionMS,
		RowsAffected: rowsAffected,
		QueryHash:    queryHash,
		Query:        query,
	}

	if err != nil {
		entry.Level = ERROR
		entry.Error = err.Error()
		entry.Message = "Query execution failed"
	}

	l.log(entry)
}

// LogSlowQuery logs queries that exceed the slow query threshold
func (l *Logger) LogSlowQuery(queryID string, query string, queryHash string, executionMS float64, threshold float64) {
	l.log(LogEntry{
		Timestamp:   time.Now(),
		Level:       WARN,
		Message:     fmt.Sprintf("Slow query detected (threshold: %.2fms)", threshold),
		QueryID:     queryID,
		ExecutionMS: executionMS,
		QueryHash:   queryHash,
		Query:       query,
		Metadata: map[string]interface{}{
			"slow_query_detected": true,
			"threshold_ms":        threshold,
			"exceeded_by_ms":      executionMS - threshold,
		},
	})
}

// Close closes the log file
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// GetLogPath returns the path to the current log file
func (l *Logger) GetLogPath() string {
	return l.logPath
}
