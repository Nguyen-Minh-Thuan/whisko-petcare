package command

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/pkg/errors"
)

// CreateUserWithUoWHandler handles user creation with Unit of Work
type CreateUserWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewCreateUserWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *CreateUserWithUoWHandler {
	return &CreateUserWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *CreateUserWithUoWHandler) Handle(ctx context.Context, cmd *CreateUser) error {
	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create user aggregate (with optional imageUrl)
	user, err := aggregate.NewUser(cmd.UserID, cmd.Name, cmd.Email, cmd.ImageUrl)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("invalid user data: %v", err))
	}

	// Update contact info if provided
	if cmd.Phone != "" || cmd.Address != "" {
		if err := user.UpdateContactInfo(cmd.Phone, cmd.Address); err != nil {
			uow.Rollback(ctx)
			return err
		}
	}

	// Save user using repository from unit of work
	userRepo := uow.UserRepository()
	if err := userRepo.Save(ctx, user); err != nil {
		uow.Rollback(ctx)
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Publish events
	events := user.GetUncommittedEvents()
	for _, event := range events {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			uow.Rollback(ctx)
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateUserProfileWithUoWHandler handles user profile updates with Unit of Work
type UpdateUserProfileWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewUpdateUserProfileWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *UpdateUserProfileWithUoWHandler {
	return &UpdateUserProfileWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *UpdateUserProfileWithUoWHandler) Handle(ctx context.Context, cmd *UpdateUserProfile) error {
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	if err := uow.Begin(ctx); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Get user from repository
	userRepo := uow.UserRepository()
	user, err := userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		uow.Rollback(ctx)
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update profile
	if err := user.UpdateProfile(cmd.Name, cmd.Email); err != nil {
		uow.Rollback(ctx)
		return err
	}

	// Save changes
	if err := userRepo.Save(ctx, user); err != nil {
		uow.Rollback(ctx)
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Publish events
	events := user.GetUncommittedEvents()
	for _, event := range events {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			uow.Rollback(ctx)
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}

	if err := uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// UpdateUserContactWithUoWHandler handles user contact updates with Unit of Work
type UpdateUserContactWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewUpdateUserContactWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *UpdateUserContactWithUoWHandler {
	return &UpdateUserContactWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *UpdateUserContactWithUoWHandler) Handle(ctx context.Context, cmd *UpdateUserContact) error {
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	if err := uow.Begin(ctx); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	userRepo := uow.UserRepository()
	user, err := userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		uow.Rollback(ctx)
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := user.UpdateContactInfo(cmd.Phone, cmd.Address); err != nil {
		uow.Rollback(ctx)
		return err
	}

	if err := userRepo.Save(ctx, user); err != nil {
		uow.Rollback(ctx)
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Publish events
	events := user.GetUncommittedEvents()
	for _, event := range events {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			uow.Rollback(ctx)
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}

	if err := uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteUserWithUoWHandler handles user deletion with Unit of Work
type DeleteUserWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewDeleteUserWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *DeleteUserWithUoWHandler {
	return &DeleteUserWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *DeleteUserWithUoWHandler) Handle(ctx context.Context, cmd *DeleteUser) error {
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	if err := uow.Begin(ctx); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	userRepo := uow.UserRepository()
	user, err := userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		uow.Rollback(ctx)
		return fmt.Errorf("failed to get user: %w", err)
	}

	if err := user.Delete(); err != nil {
		uow.Rollback(ctx)
		return err
	}

	if err := userRepo.Save(ctx, user); err != nil {
		uow.Rollback(ctx)
		return fmt.Errorf("failed to save user: %w", err)
	}

	// Publish events
	events := user.GetUncommittedEvents()
	for _, event := range events {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			uow.Rollback(ctx)
			return fmt.Errorf("failed to publish event: %w", err)
		}
	}

	if err := uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
