package command

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/pkg/errors"
)

// ============================================
// Update User Image Handler
// ============================================

type UpdateUserImageWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewUpdateUserImageWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *UpdateUserImageWithUoWHandler {
	return &UpdateUserImageWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *UpdateUserImageWithUoWHandler) Handle(ctx context.Context, cmd *UpdateUserImage) error {
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	userRepo := uow.UserRepository()
	user, err := userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError(fmt.Sprintf("user not found: %v", err))
	}

	if err := user.UpdateImageUrl(cmd.ImageUrl); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to update image URL: %v", err))
	}

	// Get events BEFORE saving (Save() will clear them)
	events := user.GetUncommittedEvents()
	fmt.Printf("🔍 UpdateUserImage: Got %d uncommitted events before save\n", len(events))
	for i, evt := range events {
		fmt.Printf("  Event %d: Type=%s\n", i+1, evt.EventType())
	}

	if err := userRepo.Save(ctx, user); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save user: %v", err))
	}

	// Commit transaction FIRST
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	fmt.Printf("📤 UpdateUserImage: Publishing %d events...\n", len(events))
	// Publish events AFTER successful commit
	for _, event := range events {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Warning: failed to publish user image event: %v\n", err)
		}
	}

	return nil
}

// ============================================
// Update Pet Image Handler
// ============================================

type UpdatePetImageWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewUpdatePetImageWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *UpdatePetImageWithUoWHandler {
	return &UpdatePetImageWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *UpdatePetImageWithUoWHandler) Handle(ctx context.Context, cmd *UpdatePetImage) error {
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	petRepo := uow.PetRepository()
	pet, err := petRepo.GetByID(ctx, cmd.PetID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError(fmt.Sprintf("pet not found: %v", err))
	}

	if err := pet.UpdateImageUrl(cmd.ImageUrl); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to update image URL: %v", err))
	}

	// Get events BEFORE saving (Save() will clear them)
	events := pet.GetUncommittedEvents()

	if err := petRepo.Save(ctx, pet); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save pet: %v", err))
	}

	// Commit transaction FIRST
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	// Publish events AFTER successful commit
	for _, event := range events {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Warning: failed to publish pet image event: %v\n", err)
		}
	}

	return nil
}

// ============================================
// Update Vendor Image Handler
// ============================================

type UpdateVendorImageWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewUpdateVendorImageWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *UpdateVendorImageWithUoWHandler {
	return &UpdateVendorImageWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *UpdateVendorImageWithUoWHandler) Handle(ctx context.Context, cmd *UpdateVendorImage) error {
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	vendorRepo := uow.VendorRepository()
	vendor, err := vendorRepo.GetByID(ctx, cmd.VendorID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError(fmt.Sprintf("vendor not found: %v", err))
	}

	if err := vendor.UpdateImageUrl(cmd.ImageUrl); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to update image URL: %v", err))
	}

	// Get events BEFORE saving (Save() will clear them)
	events := vendor.GetUncommittedEvents()

	if err := vendorRepo.Save(ctx, vendor); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save vendor: %v", err))
	}

	// Commit transaction FIRST
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	// Publish events AFTER successful commit
	for _, event := range events {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Warning: failed to publish vendor image event: %v\n", err)
		}
	}

	return nil
}

// ============================================
// Update Service Image Handler
// ============================================

type UpdateServiceImageWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewUpdateServiceImageWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *UpdateServiceImageWithUoWHandler {
	return &UpdateServiceImageWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *UpdateServiceImageWithUoWHandler) Handle(ctx context.Context, cmd *UpdateServiceImage) error {
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	if err := uow.Begin(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	serviceRepo := uow.ServiceRepository()
	service, err := serviceRepo.GetByID(ctx, cmd.ServiceID)
	if err != nil {
		uow.Rollback(ctx)
		return errors.NewNotFoundError(fmt.Sprintf("service not found: %v", err))
	}

	if err := service.UpdateImageUrl(cmd.ImageUrl); err != nil {
		uow.Rollback(ctx)
		return errors.NewValidationError(fmt.Sprintf("failed to update image URL: %v", err))
	}

	// Get events BEFORE saving (Save() will clear them)
	events := service.GetUncommittedEvents()

	if err := serviceRepo.Save(ctx, service); err != nil {
		uow.Rollback(ctx)
		return errors.NewInternalError(fmt.Sprintf("failed to save service: %v", err))
	}

	// Commit transaction FIRST
	if err := uow.Commit(ctx); err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	// Publish events AFTER successful commit
	for _, event := range events {
		if err := h.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Warning: failed to publish service image event: %v\n", err)
		}
	}

	return nil
}
