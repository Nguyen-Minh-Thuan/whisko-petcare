package middleware

import (
	"context"
	"net/http"
	"whisko-petcare/internal/domain/aggregate"
)

// RoleAuthMiddleware checks if the user has one of the required roles
func RoleAuthMiddleware(allowedRoles ...aggregate.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user role from context (set by JWT middleware)
			userRole, ok := r.Context().Value("user_role").(string)
			if !ok || userRole == "" {
				sendUnauthorized(w, "User role not found")
				return
			}

			// Check if user has one of the allowed roles
			role := aggregate.UserRole(userRole)
			hasPermission := false
			for _, allowedRole := range allowedRoles {
				if role == allowedRole {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				sendForbidden(w, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin middleware that requires Admin role
func RequireAdmin(next http.Handler) http.Handler {
	return RoleAuthMiddleware(aggregate.RoleAdmin)(next)
}

// RequireVendor middleware that requires Vendor or Admin role
func RequireVendor(next http.Handler) http.Handler {
	return RoleAuthMiddleware(aggregate.RoleVendor, aggregate.RoleAdmin)(next)
}

// RequireUser middleware that allows any authenticated user
func RequireUser(next http.Handler) http.Handler {
	return RoleAuthMiddleware(aggregate.RoleUser, aggregate.RoleVendor, aggregate.RoleAdmin)(next)
}

// Helper function to send forbidden response
func sendForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error": "` + message + `"}`))
}

// Context helper functions for roles
func GetUserRole(ctx context.Context) (aggregate.UserRole, bool) {
	role, ok := ctx.Value("user_role").(string)
	if !ok {
		return "", false
	}
	return aggregate.UserRole(role), true
}

func WithUserRole(ctx context.Context, role aggregate.UserRole) context.Context {
	return context.WithValue(ctx, "user_role", string(role))
}
