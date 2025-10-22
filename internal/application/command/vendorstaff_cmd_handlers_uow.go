package command

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/pkg/errors"
)

// CreateVendorStaffWithUoWHandler handles create vendor staff commands with Unit of Work
type CreateVendorStaffWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewCreateVendorStaffWithUoWHandler creates a new create vendor staff handler with UoW
func NewCreateVendorStaffWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *CreateVendorStaffWithUoWHandler {
	return &CreateVendorStaffWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the create vendor staff command
func (h *CreateVendorStaffWithUoWHandler) Handle(ctx context.Context, cmd *CreateVendorStaff) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.UserID == "" {
		return errors.NewValidationError("user_id is required")
	}
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

	// Create vendor staff aggregate
	vendorStaff, err := aggregate.NewVendorStaff(cmd.UserID, cmd.VendorID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to create vendor staff: %v", err))
	}

	// Save vendor staff using repository from unit of work
	vendorStaffRepo := uow.VendorStaffRepository()
	if err := vendorStaffRepo.Save(ctx, vendorStaff); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save vendor staff: %v", err))
	}

	// Publish events asynchronously
	events := vendorStaff.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish vendor staff events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}

// DeleteVendorStaffWithUoWHandler handles delete vendor staff commands with Unit of Work
type DeleteVendorStaffWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewDeleteVendorStaffWithUoWHandler creates a new delete vendor staff handler with UoW
func NewDeleteVendorStaffWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *DeleteVendorStaffWithUoWHandler {
	return &DeleteVendorStaffWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the delete vendor staff command
func (h *DeleteVendorStaffWithUoWHandler) Handle(ctx context.Context, cmd *DeleteVendorStaff) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.UserID == "" {
		return errors.NewValidationError("user_id is required")
	}
	if cmd.VendorID == "" {
		return errors.NewValidationError("vendor_id is required")
	}

	// Composite ID
	compositeID := cmd.UserID + "-" + cmd.VendorID

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get vendor staff from repository
	vendorStaffRepo := uow.VendorStaffRepository()
	vendorStaffAggregate, err := vendorStaffRepo.GetByID(ctx, compositeID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("vendor staff")
	}

	// Delete vendor staff
	if err := vendorStaffAggregate.Delete(); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to delete vendor staff: %v", err))
	}

	// Save updated vendor staff
	if err := vendorStaffRepo.Save(ctx, vendorStaffAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save vendor staff: %v", err))
	}

	// Publish events asynchronously
	events := vendorStaffAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish vendor staff events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}
