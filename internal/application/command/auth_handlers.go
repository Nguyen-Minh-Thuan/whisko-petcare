package command

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/projection"
	"whisko-petcare/pkg/errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// RegisterUserHandler handles user registration
type RegisterUserHandler struct {
	userAuthRepo      *projection.MongoUserAuthRepository
	createUserHandler *CreateUserWithUoWHandler
}

// NewRegisterUserHandler creates a new register user handler
func NewRegisterUserHandler(
	userAuthRepo *projection.MongoUserAuthRepository,
	createUserHandler *CreateUserWithUoWHandler,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		userAuthRepo:      userAuthRepo,
		createUserHandler: createUserHandler,
	}
}

// Handle executes the register user command
func (h *RegisterUserHandler) Handle(ctx context.Context, cmd *RegisterUserCommand) (*RegisterUserResponse, error) {
	// Validate input
	if cmd.Email == "" || cmd.Password == "" || cmd.Name == "" {
		return nil, errors.NewValidationError("email, password, and name are required")
	}

	// Check if email already exists
	exists, err := h.userAuthRepo.ExistsByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to check email: %v", err))
	}
	if exists {
		return nil, errors.NewValidationError("email already registered")
	}

	// Create user ID
	userID := uuid.New().String()

	// Create user aggregate with password
	user, err := aggregate.NewUserWithPassword(userID, cmd.Name, cmd.Email, cmd.Password)
	if err != nil {
		return nil, errors.NewValidationError(fmt.Sprintf("failed to create user: %v", err))
	}

	// Update contact info if provided
	if cmd.Phone != "" || cmd.Address != "" {
		if err := user.UpdateContactInfo(cmd.Phone, cmd.Address); err != nil {
			return nil, err
		}
	}

	// Save user auth
	err = h.userAuthRepo.Save(ctx, user)
	if err != nil {
		return nil, errors.NewInternalError(fmt.Sprintf("failed to save user auth: %v", err))
	}

	// Create user profile using command handler
	createUserCmd := &CreateUser{
		UserID:  userID,
		Name:    cmd.Name,
		Email:   cmd.Email,
		Phone:   cmd.Phone,
		Address: cmd.Address,
	}
	err = h.createUserHandler.Handle(ctx, createUserCmd)
	if err != nil {
		// Rollback user auth if user creation fails
		_ = h.userAuthRepo.Delete(ctx, userID)
		return nil, errors.NewInternalError(fmt.Sprintf("failed to create user: %v", err))
	}

	// Generate token
	token := generateToken()

	return &RegisterUserResponse{
		UserID: userID,
		Email:  cmd.Email,
		Name:   cmd.Name,
		Token:  token,
	}, nil
}

// LoginHandler handles user login
type LoginHandler struct {
	userAuthRepo *projection.MongoUserAuthRepository
	userRepo     repository.UserRepository
}

// NewLoginHandler creates a new login handler
func NewLoginHandler(
	userAuthRepo *projection.MongoUserAuthRepository,
	userRepo repository.UserRepository,
) *LoginHandler {
	return &LoginHandler{
		userAuthRepo: userAuthRepo,
		userRepo:     userRepo,
	}
}

// Handle executes the login command
func (h *LoginHandler) Handle(ctx context.Context, cmd *LoginCommand) (*LoginResponse, error) {
	// Validate input
	if cmd.Email == "" || cmd.Password == "" {
		return nil, errors.NewValidationError("email and password are required")
	}

	// Find user by email
	userAuthModel, err := h.userAuthRepo.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, errors.NewValidationError("invalid email or password")
	}

	// Verify password directly using bcrypt
	err = verifyPassword(userAuthModel.HashedPassword, cmd.Password)
	if err != nil {
		return nil, errors.NewValidationError("invalid email or password")
	}

	// Get user details to update last login
	user, err := h.userRepo.GetByID(ctx, userAuthModel.UserID)
	if err != nil {
		return nil, errors.NewInternalError("failed to get user details")
	}

	// Update last login
	user.UpdateLastLogin()
	_ = h.userAuthRepo.Save(ctx, user)

	// Generate token
	token := generateToken()

	return &LoginResponse{
		UserID: user.ID(),
		Email:  userAuthModel.Email,
		Name:   user.Name(),
		Token:  token,
	}, nil
}

// Helper functions

func generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func verifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
