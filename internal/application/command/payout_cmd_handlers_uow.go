package command

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/pkg/errors"

	"github.com/google/uuid"
)

// ============================================
// Request Payout Handler (UoW)
// ============================================

// RequestPayoutWithUoWHandler handles payout request commands with Unit of Work
type RequestPayoutWithUoWHandler struct {
	uowFactory repository.UnitOfWorkFactory
	eventBus   bus.EventBus
}

// NewRequestPayoutWithUoWHandler creates a new request payout handler with UoW
func NewRequestPayoutWithUoWHandler(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) *RequestPayoutWithUoWHandler {
	return &RequestPayoutWithUoWHandler{
		uowFactory: uowFactory,
		eventBus:   eventBus,
	}
}

// Handle processes the request payout command
func (h *RequestPayoutWithUoWHandler) Handle(ctx context.Context, cmd *RequestPayout) error {
	fmt.Printf("\n💰 === RequestPayoutHandler DEBUG START ===\n")
	fmt.Printf("📋 Creating payout for vendor: %s\n", cmd.VendorID)
	fmt.Printf("   Payment ID: %s\n", cmd.PaymentID)
	fmt.Printf("   Schedule ID: %s\n", cmd.ScheduleID)
	fmt.Printf("   Amount: %d VND\n", cmd.Amount)
	
	if cmd == nil {
		fmt.Printf("❌ Command is nil\n")
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewValidationError("command cannot be nil")
	}

	// Validate command
	if cmd.VendorID == "" {
		fmt.Printf("❌ Vendor ID is required\n")
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewValidationError("vendor_id is required")
	}
	if cmd.PaymentID == "" {
		fmt.Printf("❌ Payment ID is required\n")
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewValidationError("payment_id is required")
	}
	if cmd.ScheduleID == "" {
		fmt.Printf("❌ Schedule ID is required\n")
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewValidationError("schedule_id is required")
	}
	if cmd.Amount <= 0 {
		fmt.Printf("❌ Amount must be positive: %d\n", cmd.Amount)
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewValidationError("amount must be positive")
	}

	// Create unit of work
	fmt.Printf("🔧 Creating Unit of Work...\n")
	uow := h.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	// Begin transaction
	fmt.Printf("🔧 Beginning transaction...\n")
	if err := uow.Begin(ctx); err != nil {
		fmt.Printf("❌ Failed to begin transaction: %v\n", err)
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewInternalError(fmt.Sprintf("failed to begin transaction: %v", err))
	}

	// Get vendor and verify bank account
	fmt.Printf("🏢 Fetching vendor: %s\n", cmd.VendorID)
	vendorRepo := uow.VendorRepository()
	vendor, err := vendorRepo.GetByID(ctx, cmd.VendorID)
	if err != nil {
		fmt.Printf("❌ Vendor not found: %v\n", err)
		uow.Rollback(ctx)
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewValidationError(fmt.Sprintf("vendor not found: %v", err))
	}

	// Check if vendor has bank account
	if !vendor.HasBankAccount() {
		fmt.Printf("❌ Vendor %s has no bank account configured\n", cmd.VendorID)
		uow.Rollback(ctx)
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewValidationError("vendor must have a bank account configured for payouts")
	}

	bankAccount := vendor.GetBankAccount()
	fmt.Printf("🏦 Vendor bank account found:\n")
	fmt.Printf("   Bank: %s\n", bankAccount.BankName)
	fmt.Printf("   Account: %s\n", bankAccount.AccountNumber)
	fmt.Printf("   Name: %s\n", bankAccount.AccountName)

	// Create aggregate bank account for payout
	payoutBankAccount := aggregate.BankAccount{
		BankName:      bankAccount.BankName,
		AccountNumber: bankAccount.AccountNumber,
		AccountName:   bankAccount.AccountName,
		BankBranch:    bankAccount.BankBranch,
	}

	// Create payout aggregate
	fmt.Printf("💸 Creating payout aggregate...\n")
	payoutID := uuid.New().String()
	payout, err := aggregate.NewPayout(
		payoutID,
		cmd.VendorID,
		cmd.PaymentID,
		cmd.ScheduleID,
		cmd.Amount,
		payoutBankAccount,
		cmd.Notes,
	)
	if err != nil {
		fmt.Printf("❌ Failed to create payout: %v\n", err)
		uow.Rollback(ctx)
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewValidationError(fmt.Sprintf("failed to create payout: %v", err))
	}

	fmt.Printf("✅ Payout aggregate created: %s\n", payoutID)
	fmt.Printf("   Status: %s\n", payout.Status())
	fmt.Printf("   Amount: %d VND\n", payout.Amount())

	// Save payout
	fmt.Printf("💾 Saving payout to database...\n")
	payoutRepo := uow.PayoutRepository()
	if err := payoutRepo.Save(ctx, payout); err != nil {
		fmt.Printf("❌ Failed to save payout: %v\n", err)
		uow.Rollback(ctx)
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewInternalError(fmt.Sprintf("failed to save payout: %v", err))
	}

	fmt.Printf("✅ Payout saved successfully\n")

	// Get events before commit
	events := payout.GetUncommittedEvents()
	fmt.Printf("📦 Captured %d events\n", len(events))

	// Commit transaction
	fmt.Printf("💾 Committing transaction...\n")
	if err := uow.Commit(ctx); err != nil {
		fmt.Printf("❌ Failed to commit transaction: %v\n", err)
		fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")
		return errors.NewInternalError(fmt.Sprintf("failed to commit transaction: %v", err))
	}

	fmt.Printf("✅ Transaction committed successfully\n")

	// Publish events asynchronously
	fmt.Printf("📢 Publishing payout events...\n")
	if err := h.eventBus.PublishBatch(ctx, events); err != nil {
		fmt.Printf("⚠️  Warning: failed to publish payout events: %v\n", err)
	} else {
		fmt.Printf("✅ Events published successfully\n")
	}

	fmt.Printf("✅ Payout request completed successfully!\n")
	fmt.Printf("   Payout ID: %s\n", payoutID)
	fmt.Printf("   Vendor: %s\n", cmd.VendorID)
	fmt.Printf("   Amount: %d VND\n", cmd.Amount)
	fmt.Printf("   Status: PENDING (awaiting automatic processing)\n")
	fmt.Printf("💰 === RequestPayoutHandler DEBUG END ===\n\n")

	return nil
}
