package command

// ============================================
// Authentication Commands
// ============================================

// RegisterUserCommand represents a command to register a new user with credentials
type RegisterUserCommand struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone,omitempty"`
	Address  string `json:"address,omitempty"`
	Role     string `json:"role,omitempty"` // Optional, defaults to "User"
}

// RegisterUserResponse represents the response after user registration
type RegisterUserResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Token  string `json:"token"`
}

// LoginCommand represents a command to login a user
type LoginCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the response after successful login
type LoginResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	Role   string `json:"role"`
	Token  string `json:"token"`
}

// ChangePasswordCommand represents a command to change user password
type ChangePasswordCommand struct {
	UserID      string `json:"user_id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ResetPasswordCommand represents a command to reset password (with token)
type ResetPasswordCommand struct {
	Email       string `json:"email"`
	ResetToken  string `json:"reset_token"`
	NewPassword string `json:"new_password"`
}
