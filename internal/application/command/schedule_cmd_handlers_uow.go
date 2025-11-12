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

// CreateScheduleWithUoWHandler handles create schedule commands with Unit of Work
type CreateScheduleWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewCreateScheduleWithUoWHandler creates a new create schedule handler with UoW
func NewCreateScheduleWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *CreateScheduleWithUoWHandler {
	return &CreateScheduleWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the create schedule command
func (h *CreateScheduleWithUoWHandler) Handle(ctx context.Context, cmd *CreateSchedule) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.UserID == "" {
		return errors.NewValidationError("user_id is required")
	}
	if cmd.VendorID == "" {
		return errors.NewValidationError("shop_id is required")
	}
	if cmd.PetID == "" {
		return errors.NewValidationError("pet_id is required")
	}
	if cmd.StartTime == "" {
		return errors.NewValidationError("start_time is required")
	}
	if cmd.EndTime == "" {
		return errors.NewValidationError("end_time is required")
	}

	// Parse time strings
	startTime, err := time.Parse(time.RFC3339, cmd.StartTime)
	if err != nil {
		return errors.NewValidationError(fmt.Sprintf("invalid start_time format: %v", err))
	}
	endTime, err := time.Parse(time.RFC3339, cmd.EndTime)
	if err != nil {
		return errors.NewValidationError(fmt.Sprintf("invalid end_time format: %v", err))
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Validate that User exists
	userRepo := uow.UserRepository()
	user, err := userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("user not found: %v", err))
	}
	// Create booking user with real data from User aggregate
	bookingUser := aggregate.BookingUser{
		UserID:  cmd.UserID,
		Name:    user.Name(),
		Email:   user.Email(),
		Phone:   user.Phone(),
		Address: user.Address(),
	}

	// Validate that Vendor/Shop exists
	vendorRepo := uow.VendorRepository()
	vendor, err := vendorRepo.GetByID(ctx, cmd.VendorID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("vendor/shop not found: %v", err))
	}
	// Validate that Pet exists and belongs to the user
	petRepo := uow.PetRepository()
	pet, err := petRepo.GetByID(ctx, cmd.PetID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("pet not found: %v", err))
	}
	if pet.UserID() != cmd.UserID {
		uow.Rollback(ctx)
		return errors.NewValidationError("pet does not belong to this user")
	}
	// Create assigned pet with real data
	assignedPet := aggregate.PetAssigned{
		PetID:   cmd.PetID,
		Name:    pet.Name(),
		Species: pet.Species(),
		Breed:   pet.Breed(),
		Age:     pet.Age(),
		Weight:  pet.Weight(),
	}

	// Validate that all Services exist and belong to the vendor
	serviceRepo := uow.ServiceRepository()
	var bookedServices []aggregate.BookedServices
	for _, serviceID := range cmd.ServiceIDs {
		service, err := serviceRepo.GetByID(ctx, serviceID)
		if err != nil {
			uow.Rollback(ctx)
			return errors.NewValidationError(fmt.Sprintf("service %s not found: %v", serviceID, err))
		}
		if service.VendorID() != cmd.VendorID {
			uow.Rollback(ctx)
			return errors.NewValidationError(fmt.Sprintf("service %s does not belong to vendor %s", serviceID, cmd.VendorID))
		}
		// Add service with real data
		bookedServices = append(bookedServices, aggregate.BookedServices{
			ServiceID: serviceID,
			Name:      service.Name(),
		})
	}

	// Create booked shop with real data
	bookedVendor := aggregate.BookedVendor{
		ShopID:         cmd.VendorID,
		Name:           vendor.Name(),
		Location:       vendor.Address(),
		Phone:          vendor.Phone(),
		BookedServices: bookedServices,
	}

	// Create schedule aggregate with validated data
	schedule, err := aggregate.NewSchedule(bookingUser, bookedVendor, assignedPet, startTime, endTime)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to create schedule: %v", err))
	}

	// Save schedule using repository from unit of work
	scheduleRepo := uow.ScheduleRepository()
	if err := scheduleRepo.Save(ctx, schedule); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save schedule: %v", err))
	}

	// Publish events asynchronously
	events := schedule.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish schedule events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	fmt.Printf("✅ Schedule created successfully: %s\n", schedule.ID())

	// Auto-create payout if vendor has bank account configured
	fmt.Printf("🔍 Checking if vendor %s has bank account for payout...\n", cmd.VendorID)
	if vendor.HasBankAccount() {
		fmt.Printf("✅ Vendor has bank account - Creating payout for schedule %s\n", schedule.ID())
		
		// Get payment ID from the schedule
		paymentID := cmd.PaymentID
		if paymentID == "" {
			fmt.Printf("⚠️ No payment ID in schedule command - payout creation skipped\n")
			return nil
		}

		// Calculate payout amount (for now, use total amount - in real system, deduct platform fee)
		payoutAmount := cmd.TotalPrice
		
		// Create payout using vendor's bank account
		vendorBankAccount := vendor.GetBankAccount()
		if vendorBankAccount == nil {
			fmt.Printf("⚠️ Vendor bank account is nil - payout creation skipped\n")
			return nil
		}
		
		// Convert VendorBankAccount to BankAccount
		bankAccount := aggregate.BankAccount{
			BankName:      vendorBankAccount.BankName,
			AccountNumber: vendorBankAccount.AccountNumber,
			AccountName:   vendorBankAccount.AccountName,
			BankBranch:    vendorBankAccount.BankBranch,
		}
		
		payoutID := fmt.Sprintf("PAYOUT-%d", time.Now().UnixNano())
		
		fmt.Printf("💰 Creating payout: ID=%s, Amount=%d, VendorID=%s, PaymentID=%s, ScheduleID=%s\n",
			payoutID, payoutAmount, cmd.VendorID, paymentID, schedule.ID())
		
		payout, err := aggregate.NewPayout(
			payoutID,
			cmd.VendorID,
			paymentID,
			schedule.ID(),
			payoutAmount,
			bankAccount,
			fmt.Sprintf("Payout for schedule %s", schedule.ID()),
		)
		if err != nil {
			fmt.Printf("⚠️ Failed to create payout: %v\n", err)
			return nil // Don't fail the whole operation if payout creation fails
		}

		// Save payout in a new transaction
		fmt.Printf("💾 Saving payout to database...\n")
		uow2 := h.uowFactory.CreateUnitOfWork()
		defer uow2.Close()

		if err := uow2.Begin(ctx); err != nil {
			fmt.Printf("⚠️ Failed to begin payout transaction: %v\n", err)
			return nil // Don't fail the whole operation
		}

		payoutRepo := uow2.PayoutRepository()
		if err := payoutRepo.Save(ctx, payout); err != nil {
			_ = uow2.Rollback(ctx)
			fmt.Printf("⚠️ Failed to save payout: %v\n", err)
			return nil // Don't fail the whole operation
		}

		// Publish payout events
		payoutEvents := payout.GetUncommittedEvents()
		if err := h.eventBus.PublishBatch(ctx, payoutEvents); err != nil {
			fmt.Printf("⚠️ Failed to publish payout events: %v\n", err)
		}

		if err := uow2.Commit(ctx); err != nil {
			fmt.Printf("⚠️ Failed to commit payout transaction: %v\n", err)
			return nil // Don't fail the whole operation
		}

		fmt.Printf("✅ Payout created successfully: %s (Status: %s)\n", payout.ID(), payout.Status())
	} else {
		fmt.Printf("⚠️ Vendor does not have bank account configured - payout creation skipped\n")
	}

	return nil
}

// ChangeScheduleStatusWithUoWHandler handles change schedule status commands with Unit of Work
type ChangeScheduleStatusWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewChangeScheduleStatusWithUoWHandler creates a new change schedule status handler with UoW
func NewChangeScheduleStatusWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *ChangeScheduleStatusWithUoWHandler {
	return &ChangeScheduleStatusWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the change schedule status command
func (h *ChangeScheduleStatusWithUoWHandler) Handle(ctx context.Context, cmd *ChangeScheduleStatus) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.ScheduleID == "" {
		return errors.NewValidationError("schedule_id is required")
	}
	if cmd.Status == "" {
		return errors.NewValidationError("status is required")
	}

	// Validate status
	status := aggregate.ScheduleStatus(cmd.Status)
	if status != aggregate.ScheduleStatusPending &&
		status != aggregate.ScheduleStatusConfirmed &&
		status != aggregate.ScheduleStatusCompleted &&
		status != aggregate.ScheduleStatusCancelled {
		return errors.NewValidationError("invalid status value")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get schedule from repository
	scheduleRepo := uow.ScheduleRepository()
	scheduleAggregate, err := scheduleRepo.GetByID(ctx, cmd.ScheduleID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("schedule")
	}

	// Change status
	if err := scheduleAggregate.ChangeStatus(status); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to change status: %v", err))
	}

	// Save updated schedule
	if err := scheduleRepo.Save(ctx, scheduleAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save schedule: %v", err))
	}

	// Publish events asynchronously
	events := scheduleAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish schedule events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}

// CompleteScheduleWithUoWHandler handles complete schedule commands with Unit of Work
type CompleteScheduleWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewCompleteScheduleWithUoWHandler creates a new complete schedule handler with UoW
func NewCompleteScheduleWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *CompleteScheduleWithUoWHandler {
	return &CompleteScheduleWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the complete schedule command
func (h *CompleteScheduleWithUoWHandler) Handle(ctx context.Context, cmd *CompleteSchedule) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.ScheduleID == "" {
		return errors.NewValidationError("schedule_id is required")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get schedule from repository
	scheduleRepo := uow.ScheduleRepository()
	scheduleAggregate, err := scheduleRepo.GetByID(ctx, cmd.ScheduleID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("schedule")
	}

	// Complete schedule
	if err := scheduleAggregate.Complete(); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to complete schedule: %v", err))
	}

	// Save updated schedule
	if err := scheduleRepo.Save(ctx, scheduleAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save schedule: %v", err))
	}

	// Publish events asynchronously
	events := scheduleAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish schedule events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}

// CancelScheduleWithUoWHandler handles cancel schedule commands with Unit of Work
type CancelScheduleWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewCancelScheduleWithUoWHandler creates a new cancel schedule handler with UoW
func NewCancelScheduleWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *CancelScheduleWithUoWHandler {
	return &CancelScheduleWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the cancel schedule command
func (h *CancelScheduleWithUoWHandler) Handle(ctx context.Context, cmd *CancelSchedule) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.ScheduleID == "" {
		return errors.NewValidationError("schedule_id is required")
	}
	if cmd.Reason == "" {
		return errors.NewValidationError("reason is required")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get schedule from repository
	scheduleRepo := uow.ScheduleRepository()
	scheduleAggregate, err := scheduleRepo.GetByID(ctx, cmd.ScheduleID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("schedule")
	}

	// Cancel schedule
	if err := scheduleAggregate.Cancel(cmd.Reason); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to cancel schedule: %v", err))
	}

	// Save updated schedule
	if err := scheduleRepo.Save(ctx, scheduleAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save schedule: %v", err))
	}

	// Publish events asynchronously
	events := scheduleAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish schedule events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}
