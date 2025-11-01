package command

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/internal/infrastructure/projection"
	"whisko-petcare/pkg/errors"

	"github.com/google/uuid"
)

// CreateVendorStaffWithUoWHandler handles create vendor staff commands with Unit of Work
type CreateVendorStaffWithUoWHandler struct {
	uowFactory     repository.UnitOfWorkFactory
	eventBus       bus.EventBus
	userProjection projection.UserProjection
}

// NewCreateVendorStaffWithUoWHandler creates a new create vendor staff handler with UoW
func NewCreateVendorStaffWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
	userProjection projection.UserProjection,
) *CreateVendorStaffWithUoWHandler {
	return &CreateVendorStaffWithUoWHandler{
		uowFactory:     uowFactory,
		eventBus:       eventBus,
		userProjection: userProjection,
	}
}

// Handle processes the create vendor staff command
// Workflow: Find user by email -> Create vendor with default values -> Create vendor staff relationship
func (h *CreateVendorStaffWithUoWHandler) Handle(ctx context.Context, cmd *CreateVendorStaff) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.Email == "" {
		return errors.NewValidationError("email is required")
	}

	// Find user by email
	user, err := h.userProjection.GetByEmail(ctx, cmd.Email)
	if err != nil {
		return errors.NewNotFoundError(fmt.Sprintf("user with email %s not found", cmd.Email))
	}

	if user == nil {
		return errors.NewNotFoundError(fmt.Sprintf("user with email %s not found", cmd.Email))
	}

	// Prepare vendor details with defaults from user
	vendorID := uuid.New().String()
	vendorName := cmd.VendorName
	if vendorName == "" {
		vendorName = user.Name + "'s Vendor"
	}

	vendorEmail := cmd.VendorEmail
	if vendorEmail == "" {
		vendorEmail = user.Email
	}

	vendorPhone := cmd.VendorPhone
	if vendorPhone == "" {
		vendorPhone = user.Phone
	}

	vendorAddress := cmd.VendorAddress
	if vendorAddress == "" {
		vendorAddress = user.Address
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	err = uow.Begin(ctx)
	if err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Create vendor aggregate with default or provided values
	vendor, err := aggregate.NewVendor(vendorID, vendorName, vendorEmail, vendorPhone, vendorAddress)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to create vendor: %v", err))
	}

	// Save vendor using repository from unit of work
	vendorRepo := uow.VendorRepository()
	err = vendorRepo.Save(ctx, vendor)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save vendor: %v", err))
	}

	// Create vendor staff aggregate linking user and vendor as OWNER (first creator)
	vendorStaff, err := aggregate.NewVendorStaff(user.ID, vendorID, aggregate.VendorStaffRoleOwner)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to create vendor staff: %v", err))
	}

	// Save vendor staff using repository from unit of work
	vendorStaffRepo := uow.VendorStaffRepository()
	err = vendorStaffRepo.Save(ctx, vendorStaff)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save vendor staff: %v", err))
	}

	// Update user role to "vendor" only if current role is "customer"
	// Get user aggregate to update role
	userRepo := uow.UserRepository()
	userAggregate, err := userRepo.GetByID(ctx, user.ID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to load user aggregate: %v", err))
	}

	if userAggregate != nil && userAggregate.Role() == "customer" {
		// Update user role to vendor
		if err := userAggregate.UpdateRole("vendor"); err != nil {
			uow.Rollback(ctx)
			return errors.NewInternalError(fmt.Sprintf("failed to update user role: %v", err))
		}

		// Save updated user aggregate
		if err := userRepo.Save(ctx, userAggregate); err != nil {
			uow.Rollback(ctx)
			return errors.NewInternalError(fmt.Sprintf("failed to save user with updated role: %v", err))
		}

		// Publish user role updated events
		userEvents := userAggregate.GetUncommittedEvents()
		if publishErr := h.eventBus.PublishBatch(ctx, userEvents); publishErr != nil {
			fmt.Printf("Warning: failed to publish user role updated events: %v\n", publishErr)
		}
	}

	// Publish vendor events asynchronously
	vendorEvents := vendor.GetUncommittedEvents()
	if publishErr := h.eventBus.PublishBatch(ctx, vendorEvents); publishErr != nil {
		fmt.Printf("Warning: failed to publish vendor events: %v\n", publishErr)
	}

	// Publish vendor staff events asynchronously
	vendorStaffEvents := vendorStaff.GetUncommittedEvents()
	if publishErr := h.eventBus.PublishBatch(ctx, vendorStaffEvents); publishErr != nil {
		fmt.Printf("Warning: failed to publish vendor staff events: %v\n", publishErr)
	}

	// Commit transaction
	err = uow.Commit(ctx)
	if err != nil {
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
