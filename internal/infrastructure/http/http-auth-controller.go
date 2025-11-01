package http

import (
	"encoding/json"
	"net/http"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/infrastructure/projection"
	jwtutil "whisko-petcare/pkg/jwt"
	"whisko-petcare/pkg/middleware"
	"whisko-petcare/pkg/response"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HTTPAuthController handles HTTP requests for authentication
type HTTPAuthController struct {
	registerHandler       *command.RegisterUserWithUoWHandler
	changePasswordHandler *command.ChangeUserPasswordWithUoWHandler
	recordLoginHandler    *command.RecordUserLoginWithUoWHandler
	userProjection        *projection.MongoUserProjection
	jwtManager            *jwtutil.JWTManager
}

// NewHTTPAuthController creates a new HTTP auth controller
func NewHTTPAuthController(
	registerHandler *command.RegisterUserWithUoWHandler,
	changePasswordHandler *command.ChangeUserPasswordWithUoWHandler,
	recordLoginHandler *command.RecordUserLoginWithUoWHandler,
	userProjection *projection.MongoUserProjection,
	jwtManager *jwtutil.JWTManager,
) *HTTPAuthController {
	return &HTTPAuthController{
		registerHandler:       registerHandler,
		changePasswordHandler: changePasswordHandler,
		recordLoginHandler:    recordLoginHandler,
		userProjection:        userProjection,
		jwtManager:            jwtManager,
	}
}

// Register handles POST /auth/register
func (c *HTTPAuthController) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Phone    string `json:"phone"`
		Address  string `json:"address"`
		Role     string `json:"role"` // Optional, defaults to "User"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate
	if req.Email == "" || req.Password == "" || req.Name == "" {
		response.SendBadRequest(w, r, "Email, password, and name are required")
		return
	}

	if len(req.Password) < 6 {
		response.SendBadRequest(w, r, "Password must be at least 6 characters")
		return
	}

	// Check if email exists
	exists, err := c.userProjection.ExistsByEmail(r.Context(), req.Email)
	if err != nil {
		response.SendInternalError(w, r, "Failed to check email")
		return
	}
	if exists {
		response.SendBadRequest(w, r, "Email already registered")
		return
	}

	// Create user ID
	userID := uuid.New().String()

	// Determine role (default to User if not specified or invalid)
	role := aggregate.RoleUser
	if req.Role != "" {
		requestedRole := aggregate.UserRole(req.Role)
		if requestedRole.IsValid() {
			role = requestedRole
		}
	}

	// Register user using command handler
	registerCmd := &command.RegisterUser{
		UserID:   userID,
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Role:     string(role),
	}

	err = c.registerHandler.Handle(r.Context(), registerCmd)
	if err != nil {
		response.SendInternalError(w, r, "Failed to register user")
		return
	}

	// Generate JWT token
	token, err := c.jwtManager.GenerateToken(userID, req.Email, req.Name, string(role))
	if err != nil {
		response.SendInternalError(w, r, "Failed to generate token")
		return
	}

	resp := map[string]interface{}{
		"user_id": userID,
		"email":   req.Email,
		"name":    req.Name,
		"role":    string(role),
		"token":   token,
	}

	response.SendCreated(w, r, resp)
}

// Login handles POST /auth/login
func (c *HTTPAuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate
	if req.Email == "" || req.Password == "" {
		response.SendBadRequest(w, r, "Email and password are required")
		return
	}

	// Find user by email
	userModel, err := c.userProjection.GetByEmail(r.Context(), req.Email)
	if err != nil {
		response.SendBadRequest(w, r, "Invalid email or password")
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(userModel.HashedPassword), []byte(req.Password))
	if err != nil {
		response.SendBadRequest(w, r, "Invalid email or password")
		return
	}

	// Record login through event sourcing
	loginCmd := &command.RecordUserLogin{
		UserID: userModel.ID,
	}
	if loginErr := c.recordLoginHandler.Handle(r.Context(), loginCmd); loginErr != nil {
		// Log error but don't fail login
		// In production, use proper logger
	}

	// Generate JWT token
	token, err := c.jwtManager.GenerateToken(userModel.ID, userModel.Email, userModel.Name, userModel.Role)
	if err != nil {
		response.SendInternalError(w, r, "Failed to generate token")
		return
	}

	resp := map[string]interface{}{
		"user_id": userModel.ID,
		"email":   userModel.Email,
		"name":    userModel.Name,
		"role":    userModel.Role,
		"token":   token,
	}

	response.SendSuccess(w, r, resp)
}

// GetCurrentUser handles GET /auth/me - returns current user data from JWT token
func (c *HTTPAuthController) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by JWT middleware)
	// Use the middleware's GetUserIDFromContext helper to properly extract the typed context key
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		response.SendUnauthorized(w, r, "User not authenticated")
		return
	}

	// Get user from projection
	userModel, err := c.userProjection.GetByID(r.Context(), userID)
	if err != nil {
		response.SendNotFound(w, r, "User not found")
		return
	}

	// Return user data (exclude hashed password)
	resp := map[string]interface{}{
		"id":         userModel.ID,
		"name":       userModel.Name,
		"email":      userModel.Email,
		"phone":      userModel.Phone,
		"address":    userModel.Address,
		"role":       userModel.Role,
		"image_url":  userModel.ImageUrl,
		"is_active":  userModel.IsActive,
		"created_at": userModel.CreatedAt,
		"updated_at": userModel.UpdatedAt,
	}

	response.SendSuccess(w, r, resp)
}

// ChangePassword handles POST /auth/change-password
func (c *HTTPAuthController) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID      string `json:"user_id"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendBadRequest(w, r, "Invalid request body")
		return
	}

	// Validate
	if req.UserID == "" || req.OldPassword == "" || req.NewPassword == "" {
		response.SendBadRequest(w, r, "All fields are required")
		return
	}

	if len(req.NewPassword) < 6 {
		response.SendBadRequest(w, r, "New password must be at least 6 characters")
		return
	}

	// Change password through command handler (it will verify old password internally)
	changeCmd := &command.ChangeUserPassword{
		UserID:      req.UserID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	err := c.changePasswordHandler.Handle(r.Context(), changeCmd)
	if err != nil {
		response.SendBadRequest(w, r, err.Error())
		return
	}

	response.SendSuccess(w, r, map[string]string{"message": "Password changed successfully"})
}
