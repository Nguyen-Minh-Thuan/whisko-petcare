package repository

import (
	"context"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"
)

// ServiceRepository defines operations for event-sourced service aggregate
type ServiceRepository interface {
	// Event store operations
	SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error
	GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error)

	// Aggregate operations (built from events)
	Save(ctx context.Context, service *aggregate.Service) error
	GetByID(ctx context.Context, id string) (*aggregate.Service, error)

	// Event stream operations
	GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error)
	GetAllEvents(ctx context.Context) ([]event.DomainEvent, error)
}