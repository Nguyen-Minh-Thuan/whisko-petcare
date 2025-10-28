package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/application/query"
	"whisko-petcare/internal/application/services"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"
)

// HTTPUserController implements UserController for HTTP transport (Event Sourcing)
type HTTPUserController struct {
	userService *services.UserService
}

// NewHTTPUserController creates a new HTTP user controller
func NewHTTPUserController(userService *services.UserService) *HTTPUserController {
	return &HTTPUserController{
		userService: userService,
	}
}

// CreateUser handles POST /users
func (c *HTTPUserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone,omitempty"`
		Address string `json:"address,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	cmd := command.CreateUser{
		UserID:  generateUserID(),
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
	}

	if err := c.userService.CreateUser(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	// Use ApiResponse for success
	responseData := map[string]interface{}{
		"id":      cmd.UserID,
		"message": "User created successfully",
	}
	response.SendCreated(w, r, responseData)
}

// GetUser handles GET /users/{id}
func (c *HTTPUserController) GetUser(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		middleware.HandleError(w, r, errors.NewValidationError("user ID is required"))
		return
	}

	userReadModel, err := c.userService.GetUser(r.Context(), query.GetUser{UserID: id})
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	// Use ApiResponse for success
	userData := map[string]interface{}{
		"id":         userReadModel.ID,
		"name":       userReadModel.Name,
		"email":      userReadModel.Email,
		"phone":      userReadModel.Phone,
		"address":    userReadModel.Address,
		"role":       userReadModel.Role,
		"is_active":  userReadModel.IsActive,
		"version":    userReadModel.Version,
		"created_at": userReadModel.CreatedAt,
		"updated_at": userReadModel.UpdatedAt,
	}
	if userReadModel.LastLoginAt != nil {
		userData["last_login_at"] = userReadModel.LastLoginAt
	}
	response.SendSuccess(w, r, userData)
}

// ListUsers handles GET /users
func (c *HTTPUserController) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.userService.ListUsers(r.Context(), query.ListUsers{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	var userList []map[string]interface{}
	for _, user := range users {
		userData := map[string]interface{}{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"phone":      user.Phone,
			"address":    user.Address,
			"role":       user.Role,
			"is_active":  user.IsActive,
			"version":    user.Version,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		}
		if user.LastLoginAt != nil {
			userData["last_login_at"] = user.LastLoginAt
		}
		userList = append(userList, userData)
	}

	// Use ApiResponse with metadata for pagination
	meta := &response.Meta{
		Page:  1,
		Limit: 10,
		Total: len(userList),
	}
	response.SendSuccessWithMeta(w, r, userList, meta)
}

// UpdateUser handles PUT /users/{id}
func (c *HTTPUserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		middleware.HandleError(w, r, errors.NewValidationError("user ID is required"))
		return
	}

	var req struct {
		Name    string `json:"name,omitempty"`
		Email   string `json:"email,omitempty"`
		Phone   string `json:"phone,omitempty"`
		Address string `json:"address,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
		return
	}

	if req.Name != "" || req.Email != "" {
		cmd := command.UpdateUserProfile{
			UserID: id,
			Name:   req.Name,
			Email:  req.Email,
		}
		if err := c.userService.UpdateUserProfile(r.Context(), cmd); err != nil {
			middleware.HandleError(w, r, err)
			return
		}
	}

	if req.Phone != "" || req.Address != "" {
		cmd := command.UpdateUserContact{
			UserID:  id,
			Phone:   req.Phone,
			Address: req.Address,
		}
		if err := c.userService.UpdateUserContact(r.Context(), cmd); err != nil {
			middleware.HandleError(w, r, err)
			return
		}
	}

	// Use ApiResponse for success
	responseData := map[string]interface{}{
		"message": "User updated successfully",
		"user_id": id,
	}
	response.SendSuccess(w, r, responseData)
}

// DeleteUser handles DELETE /users/{id}
func (c *HTTPUserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := extractIDFromPath(r.URL.Path)
	if id == "" {
		middleware.HandleError(w, r, errors.NewValidationError("user ID is required"))
		return
	}

	cmd := command.DeleteUser{UserID: id}
	if err := c.userService.DeleteUser(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	// Use ApiResponse for success
	responseData := map[string]interface{}{
		"message": "User deleted successfully",
		"user_id": id,
	}
	response.SendSuccess(w, r, responseData)
}

func extractIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) >= 2 && parts[len(parts)-1] != "" {
		return parts[len(parts)-1]
	}
	return ""
}

func generateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}
