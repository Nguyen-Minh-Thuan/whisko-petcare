package http

import (
	"encoding/json"
	"net/http"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"

	"github.com/go-chi/chi/v5"
)

// AdminController handles admin-only endpoints
type AdminController struct{}

// NewAdminController creates a new admin controller
func NewAdminController() *AdminController {
	return &AdminController{}
}

// GetSystemStats returns system statistics (Admin only)
func (c *AdminController) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	// Get user info from context
	userID := middleware.GetUserID(r.Context())
	role, _ := middleware.GetUserRole(r.Context())

	stats := map[string]interface{}{
		"admin_user_id":    userID,
		"admin_role":       string(role),
		"total_users":      150,
		"total_vendors":    25,
		"total_services":   200,
		"total_bookings":   500,
		"total_payments":   450,
		"active_users":     120,
		"pending_bookings": 15,
		"system_status":    "healthy",
		"uptime":           "5 days",
	}

	response.SendSuccess(w, r, stats)
}

// ListAllUsers returns all users in the system (Admin only)
func (c *AdminController) ListAllUsers(w http.ResponseWriter, r *http.Request) {
	// Mock data - replace with actual database query
	users := []map[string]interface{}{
		{
			"user_id": "user-001",
			"name":    "John Doe",
			"email":   "john@example.com",
			"role":    "User",
			"active":  true,
		},
		{
			"user_id": "user-002",
			"name":    "Jane Smith",
			"email":   "jane@example.com",
			"role":    "Vendor",
			"active":  true,
		},
		{
			"user_id": "user-003",
			"name":    "Admin User",
			"email":   "admin@whisko.com",
			"role":    "Admin",
			"active":  true,
		},
	}

	response.SendSuccess(w, r, map[string]interface{}{
		"users": users,
		"total": len(users),
	})
}

// UpdateUserRole updates a user's role (Admin only)
func (c *AdminController) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var req struct {
		Role string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate role
	role := aggregate.UserRole(req.Role)
	if !role.IsValid() {
		response.SendBadRequest(w, r, "Invalid role. Must be: Admin, Vendor, or User")
		return
	}

	// Mock update - replace with actual database update
	result := map[string]interface{}{
		"user_id":  userID,
		"new_role": req.Role,
		"updated":  true,
		"message":  "User role updated successfully",
	}

	response.SendSuccess(w, r, result)
}

// DeleteUser deletes a user from the system (Admin only)
func (c *AdminController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Prevent self-deletion
	currentUserID := middleware.GetUserID(r.Context())
	if userID == currentUserID {
		response.SendBadRequest(w, r, "Cannot delete your own account")
		return
	}

	// Mock deletion - replace with actual database deletion
	result := map[string]interface{}{
		"user_id": userID,
		"deleted": true,
		"message": "User deleted successfully",
	}

	response.SendSuccess(w, r, result)
}

// GetUserDetails returns detailed information about a user (Admin only)
func (c *AdminController) GetUserDetails(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	// Mock data - replace with actual database query
	user := map[string]interface{}{
		"user_id":       userID,
		"name":          "John Doe",
		"email":         "john@example.com",
		"phone":         "+84123456789",
		"address":       "123 Street, City",
		"role":          "User",
		"active":        true,
		"created_at":    "2025-01-15T10:30:00Z",
		"updated_at":    "2025-10-27T14:20:00Z",
		"last_login_at": "2025-10-27T08:15:00Z",
		"total_pets":    3,
		"total_bookings": 10,
		"total_payments": 5,
	}

	response.SendSuccess(w, r, user)
}

// ApproveVendor approves a vendor registration (Admin only)
func (c *AdminController) ApproveVendor(w http.ResponseWriter, r *http.Request) {
	vendorID := chi.URLParam(r, "id")

	var req struct {
		Approved bool   `json:"approved"`
		Notes    string `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	// Mock approval - replace with actual business logic
	result := map[string]interface{}{
		"vendor_id": vendorID,
		"approved":  req.Approved,
		"notes":     req.Notes,
		"status":    "approved",
		"message":   "Vendor approved successfully",
	}

	response.SendSuccess(w, r, result)
}
