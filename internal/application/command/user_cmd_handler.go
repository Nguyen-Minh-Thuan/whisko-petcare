package command

import (
	"context"
	"fmt"
	"time"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/pkg/errors"
)

// Commands
type CreateUser struct {
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone,omitempty"`
	Address string `json:"address,omitempty"`
}

type UpdateUserProfile struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
}

type UpdateUserContact struct {
	UserID  string `json:"user_id"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type DeleteUser struct {
	UserID string `json:"user_id"`
}

// CreateUserHandler handles user creation
type CreateUserHandler struct {
	userRepo repository.UserRepository
	eventBus bus.EventBus
}

func NewCreateUserHandler(userRepo repository.UserRepository, eventBus bus.EventBus) *CreateUserHandler {
	return &CreateUserHandler{
		userRepo: userRepo,
		eventBus: eventBus,
	}
}

func (h *CreateUserHandler) Handle(ctx context.Context, cmd CreateUser) error {
	// Generate ID if not provided
	if cmd.UserID == "" {
		cmd.UserID = generateUserID()
	}

	// Business rule: Check if user exists
	existingUser, err := h.userRepo.GetByID(ctx, cmd.UserID)
	if err == nil && existingUser != nil {
		return errors.NewConflictError("user already exists")
	}

	// Create aggregate
	user, err := aggregate.NewUser(cmd.UserID, cmd.Name, cmd.Email)
	if err != nil {
		return errors.NewValidationError(err.Error())
	}

	// Update contact info if provided
	if cmd.Phone != "" || cmd.Address != "" {
		if err := user.UpdateContactInfo(cmd.Phone, cmd.Address); err != nil {
			return errors.NewValidationError(err.Error())
		}
	}

	// Save to event store
	if err := h.userRepo.Save(ctx, user); err != nil {
		return errors.NewInternalError("failed to save user")
	}

	// Publish events to event bus
	for _, event := range user.GetUncommittedEvents() {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			// Log error but don't fail the command
			// In production, use outbox pattern or saga
		}
	}

	return nil
}

// UpdateUserProfileHandler handles profile updates
type UpdateUserProfileHandler struct {
	userRepo repository.UserRepository
	eventBus bus.EventBus
}

func NewUpdateUserProfileHandler(userRepo repository.UserRepository, eventBus bus.EventBus) *UpdateUserProfileHandler {
	return &UpdateUserProfileHandler{
		userRepo: userRepo,
		eventBus: eventBus,
	}
}

func (h *UpdateUserProfileHandler) Handle(ctx context.Context, cmd UpdateUserProfile) error {
	// Load aggregate from event store
	user, err := h.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// Execute business logic
	if err := user.UpdateProfile(cmd.Name, cmd.Email); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// Save events
	if err := h.userRepo.Save(ctx, user); err != nil {
		return errors.NewInternalError("failed to save user")
	}

	// Publish events
	for _, event := range user.GetUncommittedEvents() {
		h.eventBus.Publish(ctx, event)
	}

	return nil
}

// UpdateUserContactHandler handles contact updates
type UpdateUserContactHandler struct {
	userRepo repository.UserRepository
	eventBus bus.EventBus
}

func NewUpdateUserContactHandler(userRepo repository.UserRepository, eventBus bus.EventBus) *UpdateUserContactHandler {
	return &UpdateUserContactHandler{
		userRepo: userRepo,
		eventBus: eventBus,
	}
}

func (h *UpdateUserContactHandler) Handle(ctx context.Context, cmd UpdateUserContact) error {
	// Load aggregate from event store
	user, err := h.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// Execute business logic
	if err := user.UpdateContactInfo(cmd.Phone, cmd.Address); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// Save events
	if err := h.userRepo.Save(ctx, user); err != nil {
		return errors.NewInternalError("failed to save user")
	}

	// Publish events
	for _, event := range user.GetUncommittedEvents() {
		h.eventBus.Publish(ctx, event)
	}

	return nil
}

// DeleteUserHandler handles user deletion
type DeleteUserHandler struct {
	userRepo repository.UserRepository
	eventBus bus.EventBus
}

func NewDeleteUserHandler(userRepo repository.UserRepository, eventBus bus.EventBus) *DeleteUserHandler {
	return &DeleteUserHandler{
		userRepo: userRepo,
		eventBus: eventBus,
	}
}

func (h *DeleteUserHandler) Handle(ctx context.Context, cmd DeleteUser) error {
	// Load aggregate from event store
	user, err := h.userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return errors.NewNotFoundError("user")
	}

	// Execute business logic
	if err := user.Delete(); err != nil {
		return errors.NewValidationError(err.Error())
	}

	// Save events
	if err := h.userRepo.Save(ctx, user); err != nil {
		return errors.NewInternalError("failed to save user")
	}

	// Publish events
	for _, event := range user.GetUncommittedEvents() {
		h.eventBus.Publish(ctx, event)
	}

	return nil
}

func generateUserID() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}
