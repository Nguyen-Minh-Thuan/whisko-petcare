package repository

import (
	"context"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"
)

// PaymentRepository defines operations for event-sourced payment aggregate
type PaymentRepository interface {
	// Event store operations
	SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error
	GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error)

	// Aggregate operations (built from events)
	Save(ctx context.Context, payment *aggregate.Payment) error
	GetByID(ctx context.Context, id string) (*aggregate.Payment, error)
	GetByOrderCode(ctx context.Context, orderCode int64) (*aggregate.Payment, error)
	GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*aggregate.Payment, error)
	GetByStatus(ctx context.Context, status string) ([]*aggregate.Payment, error)

	// Event stream operations
	GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error)
	GetAllEvents(ctx context.Context) ([]event.DomainEvent, error)
}
