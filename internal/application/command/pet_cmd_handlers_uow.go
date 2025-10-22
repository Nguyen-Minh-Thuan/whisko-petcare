package command

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/pkg/errors"
)

// CreatePetWithUoWHandler handles create pet commands with Unit of Work
type CreatePetWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewCreatePetWithUoWHandler creates a new create pet handler with UoW
func NewCreatePetWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *CreatePetWithUoWHandler {
	return &CreatePetWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the create pet command
func (h *CreatePetWithUoWHandler) Handle(ctx context.Context, cmd *CreatePet) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.UserID == "" {
		return errors.NewValidationError("user_id is required")
	}
	if cmd.Name == "" {
		return errors.NewValidationError("name is required")
	}
	if cmd.Species == "" {
		return errors.NewValidationError("species is required")
	}
	if cmd.Age < 0 {
		return errors.NewValidationError("age cannot be negative")
	}
	if cmd.Weight < 0 {
		return errors.NewValidationError("weight cannot be negative")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Create pet aggregate
	pet, err := aggregate.NewPet(cmd.UserID, cmd.Name, cmd.Species, cmd.Breed, cmd.Age, cmd.Weight)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to create pet: %v", err))
	}

	// Save pet using repository from unit of work
	petRepo := uow.PetRepository()
	if err := petRepo.Save(ctx, pet); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save pet: %v", err))
	}

	// Publish events asynchronously
	events := pet.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		// Log warning but don't fail the command (eventual consistency)
		fmt.Printf("Warning: failed to publish pet events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}

// UpdatePetWithUoWHandler handles update pet commands with Unit of Work
type UpdatePetWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewUpdatePetWithUoWHandler creates a new update pet handler with UoW
func NewUpdatePetWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *UpdatePetWithUoWHandler {
	return &UpdatePetWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the update pet command
func (h *UpdatePetWithUoWHandler) Handle(ctx context.Context, cmd *UpdatePet) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.PetID == "" {
		return errors.NewValidationError("pet_id is required")
	}
	if cmd.Name == "" {
		return errors.NewValidationError("name is required")
	}
	if cmd.Age < 0 {
		return errors.NewValidationError("age cannot be negative")
	}
	if cmd.Weight < 0 {
		return errors.NewValidationError("weight cannot be negative")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get pet from repository
	petRepo := uow.PetRepository()
	petAggregate, err := petRepo.GetByID(ctx, cmd.PetID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("pet")
	}

	// Update pet
	if err := petAggregate.UpdateProfile(cmd.Name, cmd.Species, cmd.Breed, cmd.Age, cmd.Weight); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to update pet: %v", err))
	}

	// Save updated pet
	if err := petRepo.Save(ctx, petAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save pet: %v", err))
	}

	// Publish events asynchronously
	events := petAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish pet events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}

// DeletePetWithUoWHandler handles delete pet commands with Unit of Work
type DeletePetWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewDeletePetWithUoWHandler creates a new delete pet handler with UoW
func NewDeletePetWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *DeletePetWithUoWHandler {
	return &DeletePetWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the delete pet command
func (h *DeletePetWithUoWHandler) Handle(ctx context.Context, cmd *DeletePet) error {
	if cmd == nil {
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.PetID == "" {
		return errors.NewValidationError("pet_id is required")
	}

	// Create unit of work
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get pet from repository
	petRepo := uow.PetRepository()
	petAggregate, err := petRepo.GetByID(ctx, cmd.PetID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError("pet")
	}

	// Delete pet (soft delete)
	if err := petAggregate.Delete(); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to delete pet: %v", err))
	}

	// Save updated pet
	if err := petRepo.Save(ctx, petAggregate); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save pet: %v", err))
	}

	// Publish events asynchronously
	events := petAggregate.GetUncommittedEvents()
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("Warning: failed to publish pet events: %v\n", err)
	}

	// Commit transaction
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	return nil
}
