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
	"whisko-petcare/internal/infrastructure/cloudinary"
	"whisko-petcare/pkg/errors"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"
)

// HTTPUserController implements UserController for HTTP transport (Event Sourcing)
type HTTPUserController struct {
	userService *services.UserService
	cloudinary  *cloudinary.Service
}

// NewHTTPUserController creates a new HTTP user controller
func NewHTTPUserController(userService *services.UserService, cloud *cloudinary.Service) *HTTPUserController {
	return &HTTPUserController{
		userService: userService,
		cloudinary:  cloud,
	}
}

// CreateUser handles POST /users - supports both JSON and multipart/form-data with image
func (c *HTTPUserController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var name, email, phone, address, imageUrl string
	userID := generateUserID()

	// Check if multipart form (with image file)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Failed to parse form"))
			return
		}

		// Get form fields
		name = r.FormValue("name")
		email = r.FormValue("email")
		phone = r.FormValue("phone")
		address = r.FormValue("address")

		// Check if image file is provided
		file, fileHeader, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			
			// Upload to Cloudinary
			if c.cloudinary == nil {
				middleware.HandleError(w, r, errors.NewInternalError("Cloudinary not configured"))
				return
			}
			
			uploadRes, err := c.cloudinary.UploadUserAvatar(r.Context(), file, fileHeader.Filename, userID)
			if err != nil {
				middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("Failed to upload image: %v", err)))
				return
			}
			imageUrl = uploadRes.SecureURL
		}
	} else {
		// JSON body
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Phone    string `json:"phone,omitempty"`
			Address  string `json:"address,omitempty"`
			ImageUrl string `json:"image_url,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
			return
		}

		name = req.Name
		email = req.Email
		phone = req.Phone
		address = req.Address
		imageUrl = req.ImageUrl
	}

	// Create user command
	cmd := command.CreateUser{
		UserID:   userID,
		Name:     name,
		Email:    email,
		Phone:    phone,
		Address:  address,
		ImageUrl: imageUrl,
	}

	if err := c.userService.CreateUser(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	// Response
	responseData := map[string]interface{}{
		"id":      cmd.UserID,
		"message": "User created successfully",
	}
	if imageUrl != "" {
		responseData["image_url"] = imageUrl
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

// UpdateUserImage handles PUT /users/{id}/image - supports multipart/form-data with image file
func (c *HTTPUserController) UpdateUserImage(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	parts := strings.Split(path, "/")
	
	if len(parts) < 1 || parts[0] == "" {
		middleware.HandleError(w, r, errors.NewValidationError("user ID is required"))
		return
	}
	
	id := parts[0]

	var imageUrl string

	// Check if multipart form (with image file)
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Failed to parse form"))
			return
		}

		// Get image file
		file, fileHeader, err := r.FormFile("image")
		if err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Image file is required"))
			return
		}
		defer file.Close()
		
		// Upload to Cloudinary
		if c.cloudinary == nil {
			middleware.HandleError(w, r, errors.NewInternalError("Cloudinary not configured"))
			return
		}
		
		uploadRes, err := c.cloudinary.UploadUserImage(r.Context(), file, fileHeader.Filename, id)
		if err != nil {
			middleware.HandleError(w, r, errors.NewInternalError(fmt.Sprintf("Failed to upload image: %v", err)))
			return
		}
		imageUrl = uploadRes.SecureURL
	} else {
		// JSON body (backward compatibility)
		var req struct {
			ImageUrl string `json:"image_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.HandleError(w, r, errors.NewValidationError("Invalid JSON format"))
			return
		}

		if req.ImageUrl == "" {
			middleware.HandleError(w, r, errors.NewValidationError("image_url is required"))
			return
		}
		imageUrl = req.ImageUrl
	}

	cmd := command.UpdateUserImage{
		UserID:   id,
		ImageUrl: imageUrl,
	}

	if err := c.userService.UpdateUserImage(r.Context(), cmd); err != nil {
		middleware.HandleError(w, r, err)
		return
	}

	response.SendSuccess(w, r, map[string]interface{}{
		"message":   "User image updated successfully",
		"image_url": imageUrl,
	})
}
