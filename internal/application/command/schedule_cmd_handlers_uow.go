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

	// Create booking user
	bookingUser := aggregate.BookingUser{
		UserID:  cmd.UserID,
		Name:    cmd.BookingUser.Name,
		Email:   cmd.BookingUser.Email,
		Phone:   cmd.BookingUser.Phone,
		Address: cmd.BookingUser.Address,
	}

	// Create booked services
	var bookedServices []aggregate.BookedServices
	for _, s := range cmd.BookedVendor.Services {
		bookedServices = append(bookedServices, aggregate.BookedServices{
			ServiceID: s.ServiceID,
			Name:      s.Name,
		})
	}

	// Create booked shop
	bookedVendor := aggregate.BookedVendor{
		ShopID:         cmd.VendorID,
		Name:           cmd.BookedVendor.Name,
		Location:       cmd.BookedVendor.Location,
		Phone:          cmd.BookedVendor.Phone,
		BookedServices: bookedServices,
	}

	// Create assigned pet
	assignedPet := aggregate.PetAssigned{
		PetID:   cmd.PetID,
		Name:    cmd.AssignedPet.Name,
		Species: cmd.AssignedPet.Species,
		Breed:   cmd.AssignedPet.Breed,
		Age:     cmd.AssignedPet.Age,
		Weight:  cmd.AssignedPet.Weight,
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Create schedule aggregate
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
