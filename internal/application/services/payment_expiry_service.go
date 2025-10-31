package services

import (
	"context"
	"fmt"
	"time"

	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/internal/infrastructure/payos"
)

// PayOSService interface for cancelling payment links
type PayOSService interface {
	CancelPaymentLink(ctx context.Context, orderCode int64, cancelReason string) error
}

// PaymentExpiryService handles automatic expiration of pending payments
type PaymentExpiryService struct {
	uowFactory   repository.UnitOfWorkFactory
	eventBus     bus.EventBus
	payOSService PayOSService
	stopChan     chan struct{}
}

// NewPaymentExpiryService creates a new payment expiry service
func NewPaymentExpiryService(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus, payOSService *payos.Service) *PaymentExpiryService {
	return &PaymentExpiryService{
		uowFactory:   uowFactory,
		eventBus:     eventBus,
		payOSService: payOSService,
		stopChan:     make(chan struct{}),
	}
}

// Start begins the background job to check for expired payments
func (s *PaymentExpiryService) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	fmt.Println("✅ Payment expiry service started (checking every 1 minute)")

	for {
		select {
		case <-ticker.C:
			if err := s.expireOldPayments(ctx); err != nil {
				fmt.Printf("❌ Error expiring payments: %v\n", err)
			}
		case <-s.stopChan:
			fmt.Println("⏹️  Payment expiry service stopped")
			return
		case <-ctx.Done():
			fmt.Println("⏹️  Payment expiry service stopped (context done)")
			return
		}
	}
}

// Stop stops the background job
func (s *PaymentExpiryService) Stop() {
	close(s.stopChan)
}

// expireOldPayments finds and expires all pending payments that have passed their expiration time
func (s *PaymentExpiryService) expireOldPayments(ctx context.Context) error {
	uow := s.uowFactory.CreateUnitOfWork()
	defer uow.Close()

	if err := uow.Begin(ctx); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	paymentRepo := uow.PaymentRepository()
	
	// Get all pending payments
	payments, err := paymentRepo.GetByStatus(ctx, "PENDING")
	if err != nil {
		uow.Rollback(ctx)
		return fmt.Errorf("failed to get pending payments: %w", err)
	}

	expiredCount := 0
	for _, payment := range payments {
		if payment.IsExpired() {
			// First, cancel the payment link on PayOS platform
			cancelReason := "Payment expired automatically"
			if err := s.payOSService.CancelPaymentLink(ctx, payment.OrderCode(), cancelReason); err != nil {
				fmt.Printf("⚠️  Failed to cancel payment link on PayOS for payment %s (orderCode: %d): %v\n", payment.ID(), payment.OrderCode(), err)
				// Continue anyway to mark as expired locally even if PayOS cancellation fails
			} else {
				fmt.Printf("✅ Cancelled payment link on PayOS for payment %s (orderCode: %d)\n", payment.ID(), payment.OrderCode())
			}

			// Then mark as expired locally
			if err := payment.MarkAsExpired(); err != nil {
				fmt.Printf("⚠️  Failed to mark payment %s as expired: %v\n", payment.ID(), err)
				continue
			}

			if err := paymentRepo.Save(ctx, payment); err != nil {
				fmt.Printf("⚠️  Failed to save expired payment %s: %v\n", payment.ID(), err)
				continue
			}

			// Publish events
			events := payment.GetUncommittedEvents()
			if err := s.eventBus.PublishBatch(ctx, events); err != nil {
				fmt.Printf("⚠️  Failed to publish events for payment %s: %v\n", payment.ID(), err)
			}

			expiredCount++
		}
	}

	if err := uow.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if expiredCount > 0 {
		fmt.Printf("✅ Expired %d pending payment(s)\n", expiredCount)
	}

	return nil
}
