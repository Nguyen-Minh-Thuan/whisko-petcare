package mongo

import (
	"context"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoUserRepository implements UserRepository with MongoDB
type MongoUserRepository struct {
	*MongoEventSourcedRepository[*aggregate.User]
	session mongo.Session
}

// NewMongoUserRepository creates a new MongoDB user repository
func NewMongoUserRepository(database *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{
		MongoEventSourcedRepository: NewMongoEventSourcedRepository[*aggregate.User](
			database,
			"users",
			"user_events",
		),
	}
}

// SetTransaction implements TransactionalRepository
func (r *MongoUserRepository) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
		r.MongoEventSourcedRepository.SetTransaction(tx)
	}
}

// GetTransaction implements TransactionalRepository
func (r *MongoUserRepository) GetTransaction() interface{} {
	return r.session
}

// IsTransactional implements TransactionalRepository
func (r *MongoUserRepository) IsTransactional() bool {
	return r.session != nil
}

// SaveEvents saves events for a user aggregate
func (r *MongoUserRepository) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	return r.MongoEventSourcedRepository.SaveEvents(ctx, aggregateID, events, expectedVersion)
}

// GetEvents gets all events for a user aggregate
func (r *MongoUserRepository) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	return r.MongoEventSourcedRepository.GetEvents(ctx, aggregateID)
}

// Save saves a user aggregate by persisting its events
func (r *MongoUserRepository) Save(ctx context.Context, user *aggregate.User) error {
	return r.MongoEventSourcedRepository.SaveAggregate(ctx, user)
}

// GetByID gets a user by ID, reconstructed from events
func (r *MongoUserRepository) GetByID(ctx context.Context, id string) (*aggregate.User, error) {
	return r.MongoEventSourcedRepository.LoadAggregate(ctx, id)
}

// GetEventsSince gets events since a specific version
func (r *MongoUserRepository) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	return r.MongoEventSourcedRepository.GetEventsSince(ctx, aggregateID, version)
}

// GetAllEvents gets all events from the event store
func (r *MongoUserRepository) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	return r.MongoEventSourcedRepository.GetAllEvents(ctx)
}
