package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"whisko-petcare/pkg/errors"
)

// Context key for request ID
type contextKey string

const requestIDKey contextKey = "requestID"

// ErrorHandler middleware for handling panics and errors
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())
				log.Printf("[%s] Panic: %v\n%s", requestID, err, debug.Stack())
				HandleError(w, r, errors.NewInternalError("Internal server error"))
			}
		}()

		// Check for request timeout/cancellation
		select {
		case <-r.Context().Done():
			if r.Context().Err() == context.DeadlineExceeded {
				HandleError(w, r, errors.NewRequestTimeoutError("Request timeout"))
			} else {
				HandleError(w, r, errors.NewInternalError("Request cancelled"))
			}
			return
		default:
		}

		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)

		// Add request ID to response header
		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimitMiddleware provides basic rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		now := time.Now()

		// Clean old requests
		if requests, exists := rl.requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < rl.window {
					validRequests = append(validRequests, reqTime)
				}
			}
			rl.requests[clientIP] = validRequests
		}

		// Check rate limit
		if len(rl.requests[clientIP]) >= rl.limit {
			HandleError(w, r, errors.NewTooManyRequestsError("Rate limit exceeded"))
			return
		}

		// Add current request
		rl.requests[clientIP] = append(rl.requests[clientIP], now)

		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware adds request timeout
func TimeoutMiddleware(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoggingMiddleware logs HTTP requests with request ID
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := GetRequestID(r.Context())

		log.Printf("[%s] Started %s %s from %s", requestID, r.Method, r.URL.Path, r.RemoteAddr)

		next.ServeHTTP(w, r)

		duration := time.Since(start)
		log.Printf("[%s] Completed %s %s in %v", requestID, r.Method, r.URL.Path, duration)
	})
}

// HandleError writes an error response with enhanced logging using ApiResponse format
func HandleError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := GetRequestID(r.Context())

	if appErr, ok := err.(*errors.ApplicationError); ok {
		log.Printf("[%s] Error %d: %s (Code: %s) for %s %s",
			requestID, appErr.Status, appErr.Message, appErr.Code, r.Method, r.URL.Path)

		sendApiErrorResponse(w, requestID, appErr.Status, appErr.Code, appErr.Message)
	} else {
		log.Printf("[%s] Unexpected error 500: %s for %s %s",
			requestID, err.Error(), r.Method, r.URL.Path)

		sendApiErrorResponse(w, requestID, 500, "INTERNAL_ERROR", "Internal server error")
	}
}

// sendApiErrorResponse sends a standardized API error response
func sendApiErrorResponse(w http.ResponseWriter, requestID string, statusCode int, code, message string) {
	response := map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
		"request_id": requestID,
		"timestamp":  time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return "unknown"
}

// generateRequestID creates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xForwardedFor := r.Header.Get("X-Forwarded-For"); xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		parts := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	if xRealIP := r.Header.Get("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		return ip[:idx]
	}
	return ip
}

// RecoveryMiddleware provides enhanced panic recovery
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := GetRequestID(r.Context())

				// Log the panic with full stack trace
				log.Printf("[%s] PANIC: %v\nStack Trace:\n%s", requestID, err, debug.Stack())

				// Check if response has already been written
				if w.Header().Get("Content-Type") == "" {
					HandleError(w, r, errors.NewInternalError("Internal server error"))
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// DatabaseErrorHandler converts database-specific errors to application errors
func DatabaseErrorHandler(err error) error {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "connection"):
		return errors.NewServiceUnavailableError("Database connection error")
	case strings.Contains(errStr, "timeout"):
		return errors.NewRequestTimeoutError("Database operation timeout")
	case strings.Contains(errStr, "duplicate") || strings.Contains(errStr, "unique"):
		return errors.NewConflictError("Resource already exists")
	case strings.Contains(errStr, "not found"):
		return errors.NewNotFoundError("resource")
	default:
		return errors.NewInternalError("Database operation failed")
	}
}
