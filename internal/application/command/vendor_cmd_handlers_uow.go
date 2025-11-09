package command

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/pkg/errors"
)

// CreateVendorWithUoWHandler handles create vendor commands with Unit of Work
type CreateVendorWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewCreateVendorWithUoWHandler creates a new create vendor handler with UoW
func NewCreateVendorWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *CreateVendorWithUoWHandler {
	return &CreateVendorWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the create vendor command
func (h *CreateVendorWithUoWHandler) Handle(ctx context.Context, cmd *CreateVendor) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.Name == "" {
		return errors.NewValidationError("name is required")
	}
	if cmd.Email == "" {
		return errors.NewValidationError("email is required")
	}
	if cmd.Phone == "" {
		return errors.NewValidationError("phone is required")
	}
	if cmd.Address == "" {
		return errors.NewValidationError("address is required")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Create vendor aggregate (with optional imageUrl)
	vendor, err := aggregate.NewVendor(cmd.VendorID, cmd.Name, cmd.Email, cmd.Phone, cmd.Address, cmd.ImageUrl)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to create vendor: %v", err))
	}

	// Get events BEFORE saving (Save() will clear them)
	events := vendor.GetUncommittedEvents()
	fmt.Printf("🔍 CreateVendor: Got %d uncommitted events before save\n", len(events))
	for i, evt := range events {
		fmt.Printf("  Event %d: Type=%s\n", i+1, evt.EventType())
	}

	// Save vendor using repository from unit of work
	vendorRepo := uow.VendorRepository()
	if err := vendorRepo.Save(ctx, vendor); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save vendor: %v", err))
	}

	// Commit transaction FIRST
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	fmt.Printf("📤 CreateVendor: Publishing %d events...\n", len(events))
	// Publish events AFTER successful commit (eventual consistency)
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish vendor events: %v\n", err)
	}

	return nil
}

// UpdateVendorWithUoWHandler handles update vendor commands with Unit of Work
type UpdateVendorWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewUpdateVendorWithUoWHandler creates a new update vendor handler with UoW
func NewUpdateVendorWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *UpdateVendorWithUoWHandler {
	return &UpdateVendorWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the update vendor command
func (h *UpdateVendorWithUoWHandler) Handle(ctx context.Context, cmd *UpdateVendor) error {
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
	if cmd.Email == "" {
		return errors.NewValidationError("email is required")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get vendor from repository
	vendorRepo := uow.VendorRepository()
	vendorAggregate, err := vendorRepo.GetByID(ctx, cmd.VendorID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("vendor")
	}

	// Update vendor
	if err := vendorAggregate.UpdateProfile(cmd.Name, cmd.Email, cmd.Phone, cmd.Address); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to update vendor: %v", err))
	}

	// Save updated vendor
	if err := vendorRepo.Save(ctx, vendorAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save vendor: %v", err))
	}

	// Publish events asynchronously
	events := vendorAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish vendor events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}

// DeleteVendorWithUoWHandler handles delete vendor commands with Unit of Work
type DeleteVendorWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewDeleteVendorWithUoWHandler creates a new delete vendor handler with UoW
func NewDeleteVendorWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *DeleteVendorWithUoWHandler {
	return &DeleteVendorWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the delete vendor command
func (h *DeleteVendorWithUoWHandler) Handle(ctx context.Context, cmd *DeleteVendor) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.VendorID == "" {
		return errors.NewValidationError("vendor_id is required")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get vendor from repository
	vendorRepo := uow.VendorRepository()
	vendorAggregate, err := vendorRepo.GetByID(ctx, cmd.VendorID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("vendor")
	}

	// Delete vendor (soft delete)
	if err := vendorAggregate.Delete(); err != nil{
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to delete vendor: %v", err))
	}

	// Save updated vendor
	if err := vendorRepo.Save(ctx, vendorAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save vendor: %v", err))
	}

	// Publish events asynchronously
	events := vendorAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish vendor events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}
