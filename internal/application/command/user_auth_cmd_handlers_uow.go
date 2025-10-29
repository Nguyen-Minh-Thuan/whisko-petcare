package command

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/pkg/errors"
)

// RegisterUserWithUoWHandler handles user registration with Unit of Work
type RegisterUserWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewRegisterUserWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *RegisterUserWithUoWHandler {
	return &RegisterUserWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *RegisterUserWithUoWHandler) Handle(ctx context.Context, cmd *RegisterUser) error {
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	if err := uow.Begin(ctx); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create user aggregate with password and role
	user, err := aggregate.NewUserWithPasswordAndRole(cmd.UserID, cmd.Name, cmd.Email, cmd.Password, aggregate.UserRole(cmd.Role))
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("invalid user data: %v", err))
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

// ChangeUserPasswordWithUoWHandler handles password changes with Unit of Work
type ChangeUserPasswordWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewChangeUserPasswordWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *ChangeUserPasswordWithUoWHandler {
	return &ChangeUserPasswordWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *ChangeUserPasswordWithUoWHandler) Handle(ctx context.Context, cmd *ChangeUserPassword) error {
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

	// Change password on aggregate
	if err := user.ChangePassword(cmd.OldPassword, cmd.NewPassword); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to change password: %v", err))
	}

	// Save user
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

// RecordUserLoginWithUoWHandler handles login recording with Unit of Work
type RecordUserLoginWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewRecordUserLoginWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *RecordUserLoginWithUoWHandler {
	return &RecordUserLoginWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *RecordUserLoginWithUoWHandler) Handle(ctx context.Context, cmd *RecordUserLogin) error {
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

	// Update last login
	user.UpdateLastLogin()

	// Save user
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
