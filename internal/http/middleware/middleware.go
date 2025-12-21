package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware logs all incoming requests and responses
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log incoming request
		log.Printf(
			"[%s] %s %s - %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			r.UserAgent(),
		)

		// Wrap response writer to capture status code and size
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call the next handler
		next.ServeHTTP(wrapped, r)

		// Log response
		duration := time.Since(start)
		log.Printf(
			"[%s] %s - Status: %d - Duration: %v",
			r.Method,
			r.RequestURI,
			wrapped.statusCode,
			duration,
		)
	})
}

// TimingMiddleware measures and logs request execution time
func TimingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		w.Header().Set("X-Response-Time", duration.String())
		log.Printf("Request %s completed in %v", r.RequestURI, duration)
	})
}

// RecoveryMiddleware recovers from panics and logs them
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":"Internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// ValidateJSONMiddleware validates that request content-type is JSON
func ValidateJSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			contentType := r.Header.Get("Content-Type")
			if contentType != "" && contentType != "application/json" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":"Content-Type must be application/json"}`))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// QueryExecutionDecorator measures query execution performance
type QueryExecutionDecorator struct {
	executionTime time.Duration
	rowsReturned  int
	querySize     int
}

// NewQueryExecutionDecorator creates a new query execution decorator
func NewQueryExecutionDecorator() *QueryExecutionDecorator {
	return &QueryExecutionDecorator{}
}

// RecordExecution records performance metrics for a query
func (qed *QueryExecutionDecorator) RecordExecution(duration time.Duration, rowCount int, queryLength int) {
	qed.executionTime = duration
	qed.rowsReturned = rowCount
	qed.querySize = queryLength
}

// GetMetrics returns the recorded metrics
func (qed *QueryExecutionDecorator) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"execution_time_ms": qed.executionTime.Milliseconds(),
		"rows_returned":     qed.rowsReturned,
		"query_size_bytes":  qed.querySize,
	}
}

// SlowQueryThreshold logs queries that exceed a threshold
func SlowQueryThreshold(threshold time.Duration) func(duration time.Duration, query string) {
	return func(duration time.Duration, query string) {
		if duration > threshold {
			log.Printf("SLOW QUERY: Duration: %v\nQuery: %s", duration, query)
		}
	}
}
