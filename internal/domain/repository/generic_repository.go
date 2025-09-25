package repository

import (
	"context"
	"whisko-petcare/internal/domain/event"
)

// Entity represents any domain entity that can be persisted
type Entity interface {
	GetID() string
	SetID(id string)
	GetVersion() int
	SetVersion(version int)
}

// AggregateRoot represents an aggregate root in DDD/Event Sourcing
type AggregateRoot interface {
	Entity
	GetUncommittedEvents() []event.DomainEvent
	MarkEventsAsCommitted()
	LoadFromHistory(events []event.DomainEvent) error
}

// GenericRepository provides CRUD operations for any entity type
type GenericRepository[T Entity] interface {
	// Basic CRUD operations
	Save(ctx context.Context, entity T) error
	GetByID(ctx context.Context, id string) (T, error)
	Update(ctx context.Context, entity T) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]T, error)
	Count(ctx context.Context) (int64, error)

	// Query operations
	FindBy(ctx context.Context, filter map[string]interface{}) ([]T, error)
	FindOneBy(ctx context.Context, filter map[string]interface{}) (T, error)
	Exists(ctx context.Context, id string) (bool, error)
}

// EventSourcedRepository provides event sourcing operations for aggregate roots
type EventSourcedRepository[T AggregateRoot] interface {
	GenericRepository[T]

	// Event sourcing operations
	SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error
	GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error)
	GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error)
	GetAllEvents(ctx context.Context) ([]event.DomainEvent, error)

	// Aggregate operations (built from events)
	LoadAggregate(ctx context.Context, aggregateID string) (T, error)
	SaveAggregate(ctx context.Context, aggregate T) error
}

// ReadOnlyRepository provides read-only operations
type ReadOnlyRepository[T Entity] interface {
	GetByID(ctx context.Context, id string) (T, error)
	List(ctx context.Context, offset, limit int) ([]T, error)
	Count(ctx context.Context) (int64, error)
	FindBy(ctx context.Context, filter map[string]interface{}) ([]T, error)
	FindOneBy(ctx context.Context, filter map[string]interface{}) (T, error)
	Exists(ctx context.Context, id string) (bool, error)
}
