package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"runtime/debug"

	"whisko-petcare/pkg/errors"
)

// ErrorHandler middleware for handling panics and errors
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic: %v\n%s", err, debug.Stack())
				HandleError(w, errors.NewInternalError("Internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// HandleError writes an error response
func HandleError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	if appErr, ok := err.(*errors.ApplicationError); ok {
		w.WriteHeader(appErr.Status)
		json.NewEncoder(w).Encode(map[string]string{
			"error": appErr.Message,
			"code":  appErr.Code,
		})
	} else {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	}
}
