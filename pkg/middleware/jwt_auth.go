package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	jwtutil "whisko-petcare/pkg/jwt"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// EmailKey is the context key for email
	EmailKey ContextKey = "email"
	// NameKey is the context key for name
	NameKey ContextKey = "name"
)

// JWTAuthMiddleware creates a middleware for JWT authentication
func JWTAuthMiddleware(jwtManager *jwtutil.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				sendUnauthorized(w, "Missing authorization header")
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				sendUnauthorized(w, "Invalid authorization header format")
				return
			}

			tokenString := parts[1]

			// Validate token
			claims, err := jwtManager.ValidateToken(tokenString)
			if err != nil {
				sendUnauthorized(w, "Invalid or expired token")
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, NameKey, claims.Name)

			// Call next handler with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalJWTAuthMiddleware creates a middleware that doesn't require authentication
// but extracts user info if token is present
func OptionalJWTAuthMiddleware(jwtManager *jwtutil.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					claims, err := jwtManager.ValidateToken(parts[1])
					if err == nil {
						ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
						ctx = context.WithValue(ctx, EmailKey, claims.Email)
						ctx = context.WithValue(ctx, NameKey, claims.Name)
						r = r.WithContext(ctx)
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetEmailFromContext extracts email from context
func GetEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(EmailKey).(string)
	return email, ok
}

// GetNameFromContext extracts name from context
func GetNameFromContext(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(NameKey).(string)
	return name, ok
}

// sendUnauthorized sends an unauthorized response
func sendUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]string{
			"code":    "UNAUTHORIZED",
			"message": message,
		},
	})
}
