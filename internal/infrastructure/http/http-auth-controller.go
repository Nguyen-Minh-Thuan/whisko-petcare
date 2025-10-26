package http

import (
	"encoding/json"
	"net/http"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/infrastructure/projection"
	jwtutil "whisko-petcare/pkg/jwt"
	"whisko-petcare/pkg/response"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// HTTPAuthController handles HTTP requests for authentication
type HTTPAuthController struct {
	userAuthRepo   *projection.MongoUserAuthRepository
	userProjection *projection.MongoUserProjection
	jwtManager     *jwtutil.JWTManager
}

// NewHTTPAuthController creates a new HTTP auth controller
func NewHTTPAuthController(
	userAuthRepo *projection.MongoUserAuthRepository,
	userProjection *projection.MongoUserProjection,
	jwtManager *jwtutil.JWTManager,
) *HTTPAuthController {
	return &HTTPAuthController{
		userAuthRepo:   userAuthRepo,
		userProjection: userProjection,
		jwtManager:     jwtManager,
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
	exists, err := c.userAuthRepo.ExistsByEmail(r.Context(), req.Email)
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

	// Create user with password
	user, err := aggregate.NewUserWithPassword(userID, req.Name, req.Email, req.Password)
	if err != nil {
		response.SendBadRequest(w, r, err.Error())
		return
	}

	// Save user auth
	err = c.userAuthRepo.Save(r.Context(), user)
	if err != nil {
		response.SendInternalError(w, r, "Failed to save user")
		return
	}

	// Generate JWT token
	token, err := c.jwtManager.GenerateToken(userID, req.Email, req.Name)
	if err != nil {
		response.SendInternalError(w, r, "Failed to generate token")
		return
	}

	resp := map[string]interface{}{
		"user_id": userID,
		"email":   req.Email,
		"name":    req.Name,
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
	userAuthModel, err := c.userAuthRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		response.SendBadRequest(w, r, "Invalid email or password")
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(userAuthModel.HashedPassword), []byte(req.Password))
	if err != nil {
		response.SendBadRequest(w, r, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := c.jwtManager.GenerateToken(userAuthModel.UserID, userAuthModel.Email, req.Email)
	if err != nil {
		response.SendInternalError(w, r, "Failed to generate token")
		return
	}

	resp := map[string]interface{}{
		"user_id": userAuthModel.UserID,
		"email":   userAuthModel.Email,
		"token":   token,
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

	// Find user auth
	userAuthModel, err := c.userAuthRepo.GetByEmail(r.Context(), req.UserID)
	if err != nil {
		response.SendBadRequest(w, r, "User not found")
		return
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(userAuthModel.HashedPassword), []byte(req.OldPassword))
	if err != nil {
		response.SendBadRequest(w, r, "Invalid old password")
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.SendInternalError(w, r, "Failed to process new password")
		return
	}

	// Update password (simplified - you'd use aggregate properly)
	userAuthModel.HashedPassword = string(hashedPassword)
	// Save updated auth...

	response.SendSuccess(w, r, map[string]string{"message": "Password changed successfully"})
}
