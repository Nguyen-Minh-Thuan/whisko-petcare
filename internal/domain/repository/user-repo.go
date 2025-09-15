package repository

import (
	"context"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"
)

// UserRepository defines operations for event-sourced user aggregate
type UserRepository interface {
	// Event store operations
	SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error
	GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error)

	// Aggregate operations (built from events)
	Save(ctx context.Context, user *aggregate.User) error
	GetByID(ctx context.Context, id string) (*aggregate.User, error)

	// Event stream operations
	GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error)
	GetAllEvents(ctx context.Context) ([]event.DomainEvent, error)
}
