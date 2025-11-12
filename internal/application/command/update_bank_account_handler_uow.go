package command

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/pkg/errors"
)

// ============================================
// Update Vendor Bank Account Handler (UoW)
// ============================================

type UpdateVendorBankAccountWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

func NewUpdateVendorBankAccountWithUoWHandler(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) *UpdateVendorBankAccountWithUoWHandler {
	return &UpdateVendorBankAccountWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

func (h *UpdateVendorBankAccountWithUoWHandler) Handle(ctx context.Context, cmd *UpdateVendorBankAccount) error {
	fmt.Printf("\n🔧 === UpdateVendorBankAccountHandler DEBUG START ===\n")
	fmt.Printf("📥 Command received:\n")
	fmt.Printf("   VendorID: '%s'\n", cmd.VendorID)
	fmt.Printf("   BankName: '%s'\n", cmd.BankName)
	fmt.Printf("   AccountNumber: '%s'\n", cmd.AccountNumber)
	fmt.Printf("   AccountName: '%s'\n", cmd.AccountName)
	fmt.Printf("   BankBranch: '%s'\n", cmd.BankBranch)
	
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	fmt.Printf("🔄 Beginning transaction...\n")
	if err := uow.Begin(ctx); err != nil {
		fmt.Printf("❌ Failed to begin transaction: %v\n", err)
		fmt.Printf("🔧 === UpdateVendorBankAccountHandler DEBUG END ===\n\n")
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	fmt.Printf("🔍 Fetching vendor from repository...\n")
	vendorRepo := uow.VendorRepository()
	vendor, err := vendorRepo.GetByID(ctx, cmd.VendorID)
	if err != nil {
		fmt.Printf("❌ Vendor not found: %v\n", err)
		uow.Rollback(ctx)
		fmt.Printf("🔧 === UpdateVendorBankAccountHandler DEBUG END ===\n\n")
		return errors.NewNotFoundError(fmt.Sprintf("vendor not found: %v", err))
	}

	fmt.Printf("✅ Vendor found: ID='%s', Name='%s'\n", vendor.ID(), vendor.Name())
	
	fmt.Printf("🏦 Updating vendor bank account...\n")
	fmt.Printf("   Bank: %s, Account: %s, Name: %s, Branch: %s\n", cmd.BankName, cmd.AccountNumber, cmd.AccountName, cmd.BankBranch)
	if err := vendor.UpdateBankAccount(cmd.BankName, cmd.AccountNumber, cmd.AccountName, cmd.BankBranch); err != nil {
		fmt.Printf("❌ Failed to update bank account: %v\n", err)
		uow.Rollback(ctx)
		fmt.Printf("🔧 === UpdateVendorBankAccountHandler DEBUG END ===\n\n")
		return errors.NewValidationError(fmt.Sprintf("failed to update bank account: %v", err))
	}

	fmt.Printf("✅ Bank account updated in aggregate\n")
	
	// Get events BEFORE saving (Save() will clear them)
	events := vendor.GetUncommittedEvents()
	fmt.Printf("📢 Uncommitted events count: %d\n", len(events))
	for i, evt := range events {
		fmt.Printf("   Event %d: %s\n", i+1, evt.EventType())
	}

	fmt.Printf("💾 Saving vendor to repository...\n")
	if err := vendorRepo.Save(ctx, vendor); err != nil {
		fmt.Printf("❌ Failed to save vendor: %v\n", err)
		uow.Rollback(ctx)
		fmt.Printf("🔧 === UpdateVendorBankAccountHandler DEBUG END ===\n\n")
		return errors.NewInternalError(fmt.Sprintf("failed to save vendor: %v", err))
	}

	fmt.Printf("✅ Vendor saved successfully\n")
	
	// Commit transaction FIRST
	fmt.Printf("✔️  Committing transaction...\n")
	if err := uow.Commit(ctx); err != nil {
		fmt.Printf("❌ Failed to commit transaction: %v\n", err)
		fmt.Printf("🔧 === UpdateVendorBankAccountHandler DEBUG END ===\n\n")
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	fmt.Printf("✅ Transaction committed successfully\n")
	
	// Publish events AFTER successful commit
	fmt.Printf("📡 Publishing %d events...\n", len(events))
	for i, event := range events {
		fmt.Printf("   Publishing event %d/%d: %s\n", i+1, len(events), event.EventType())
		if err := h.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("⚠️  Warning: failed to publish event: %v\n", err)
		} else {
			fmt.Printf("   ✅ Event published\n")
		}
	}

	fmt.Printf("🎉 Bank account update completed successfully!\n")
	fmt.Printf("🔧 === UpdateVendorBankAccountHandler DEBUG END ===\n\n")
	return nil
}
