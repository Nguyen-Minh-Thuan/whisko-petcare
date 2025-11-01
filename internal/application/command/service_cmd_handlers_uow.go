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

// CreateServiceWithUoWHandler handles create service commands with Unit of Work
type CreateServiceWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewCreateServiceWithUoWHandler creates a new create service handler with UoW
func NewCreateServiceWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *CreateServiceWithUoWHandler {
	return &CreateServiceWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the create service command
func (h *CreateServiceWithUoWHandler) Handle(ctx context.Context, cmd *CreateService) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.VendorID == "" {
		return errors.NewValidationError("vendor_id is required")
	}
	if cmd.Name == "" {
		return errors.NewValidationError("name is required")
	}
	if cmd.Description == "" {
		return errors.NewValidationError("description is required")
	}
	if cmd.Price <= 0 {
		return errors.NewValidationError("price must be greater than 0")
	}
	if cmd.Duration <= 0 {
		return errors.NewValidationError("duration must be greater than 0")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Convert duration from minutes to time.Duration
	duration := time.Duration(cmd.Duration) * time.Minute

	// Create service aggregate (with optional imageUrl)
	service, err := aggregate.NewService(cmd.VendorID, cmd.Name, cmd.Description, cmd.Price, duration, cmd.Tags, cmd.ImageUrl)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to create service: %v", err))
	}

	// Save service using repository from unit of work
	serviceRepo := uow.ServiceRepository()
	if err := serviceRepo.Save(ctx, service); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save service: %v", err))
	}

	// Publish events asynchronously
	events := service.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish service events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}

// UpdateServiceWithUoWHandler handles update service commands with Unit of Work
type UpdateServiceWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewUpdateServiceWithUoWHandler creates a new update service handler with UoW
func NewUpdateServiceWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *UpdateServiceWithUoWHandler {
	return &UpdateServiceWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the update service command
func (h *UpdateServiceWithUoWHandler) Handle(ctx context.Context, cmd *UpdateService) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.ServiceID == "" {
		return errors.NewValidationError("service_id is required")
	}
	if cmd.Name == "" {
		return errors.NewValidationError("name is required")
	}
	if cmd.Description == "" {
		return errors.NewValidationError("description is required")
	}
	if cmd.Price <= 0 {
		return errors.NewValidationError("price must be greater than 0")
	}
	if cmd.Duration <= 0 {
		return errors.NewValidationError("duration must be greater than 0")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get service from repository
	serviceRepo := uow.ServiceRepository()
	serviceAggregate, err := serviceRepo.GetByID(ctx, cmd.ServiceID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("service")
	}

	// Convert duration from minutes to time.Duration
	duration := time.Duration(cmd.Duration) * time.Minute

	// Update service
	if err := serviceAggregate.UpdateService(cmd.Name, cmd.Description, cmd.Price, duration, cmd.Tags); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to update service: %v", err))
	}

	// Save updated service
	if err := serviceRepo.Save(ctx, serviceAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save service: %v", err))
	}

	// Publish events asynchronously
	events := serviceAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish service events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}

// DeleteServiceWithUoWHandler handles delete service commands with Unit of Work
type DeleteServiceWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewDeleteServiceWithUoWHandler creates a new delete service handler with UoW
func NewDeleteServiceWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *DeleteServiceWithUoWHandler {
	return &DeleteServiceWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the delete service command
func (h *DeleteServiceWithUoWHandler) Handle(ctx context.Context, cmd *DeleteService) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.ServiceID == "" {
		return errors.NewValidationError("service_id is required")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get service from repository
	serviceRepo := uow.ServiceRepository()
	serviceAggregate, err := serviceRepo.GetByID(ctx, cmd.ServiceID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("service")
	}

	// Delete service (soft delete)
	if err := serviceAggregate.Delete(); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to delete service: %v", err))
	}

	// Save updated service
	if err := serviceRepo.Save(ctx, serviceAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save service: %v", err))
	}

	// Publish events asynchronously
	events := serviceAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish service events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}
