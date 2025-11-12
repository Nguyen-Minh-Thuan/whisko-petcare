package repository

import (
	"context"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"
)

// PayoutRepository defines operations for event-sourced payout aggregate
type PayoutRepository interface {
	// Event store operations
	SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error
	GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error)

	// Aggregate operations (built from events)
	Save(ctx context.Context, payout *aggregate.Payout) error
	GetByID(ctx context.Context, id string) (*aggregate.Payout, error)
	GetByVendorID(ctx context.Context, vendorID string, offset, limit int) ([]*aggregate.Payout, error)
	GetByStatus(ctx context.Context, status aggregate.PayoutStatus, offset, limit int) ([]*aggregate.Payout, error)
	GetPendingPayoutForVendor(ctx context.Context, vendorID string) (*aggregate.Payout, error) // Check if vendor has pending payout
	
	// Event stream operations
	GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error)
	GetAllEvents(ctx context.Context) ([]event.DomainEvent, error)
}
