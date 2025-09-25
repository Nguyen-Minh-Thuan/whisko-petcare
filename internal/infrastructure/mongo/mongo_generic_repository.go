package mongo

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"whisko-petcare/internal/domain/event"
	"whisko-petcare/internal/domain/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrEntityNotFound = errors.New("entity not found")
var ErrOptimisticLock = errors.New("optimistic locking failed - entity was modified by another operation")

// MongoGenericRepository implements GenericRepository for MongoDB
type MongoGenericRepository[T repository.Entity] struct {
	database       *mongo.Database
	collection     *mongo.Collection
	collectionName string
	session        mongo.Session
}

// NewMongoGenericRepository creates a new MongoDB generic repository
func NewMongoGenericRepository[T repository.Entity](database *mongo.Database, collectionName string) *MongoGenericRepository[T] {
	return &MongoGenericRepository[T]{
		database:       database,
		collection:     database.Collection(collectionName),
		collectionName: collectionName,
	}
}

// SetTransaction implements TransactionalRepository
func (r *MongoGenericRepository[T]) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
	}
}

// GetTransaction implements TransactionalRepository
func (r *MongoGenericRepository[T]) GetTransaction() interface{} {
	return r.session
}

// IsTransactional implements TransactionalRepository
func (r *MongoGenericRepository[T]) IsTransactional() bool {
	return r.session != nil
}

// getContext returns the appropriate context for operations
func (r *MongoGenericRepository[T]) getContext(ctx context.Context) context.Context {
	if r.session != nil {
		return mongo.NewSessionContext(ctx, r.session)
	}
	return ctx
}

// Save stores an entity
func (r *MongoGenericRepository[T]) Save(ctx context.Context, entity T) error {
	ctx = r.getContext(ctx)

	if entity.GetID() == "" {
		entity.SetID(generateID())
		entity.SetVersion(1)

		_, err := r.collection.InsertOne(ctx, entity)
		if err != nil {
			return fmt.Errorf("failed to insert entity: %w", err)
		}
		return nil
	}

	return r.Update(ctx, entity)
}

// GetByID retrieves an entity by ID
func (r *MongoGenericRepository[T]) GetByID(ctx context.Context, id string) (T, error) {
	ctx = r.getContext(ctx)
	var zero T

	filter := bson.M{"_id": id}
	var result T

	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return zero, ErrEntityNotFound
		}
		return zero, fmt.Errorf("failed to get entity by ID: %w", err)
	}

	return result, nil
}

// Update updates an existing entity with optimistic locking
func (r *MongoGenericRepository[T]) Update(ctx context.Context, entity T) error {
	ctx = r.getContext(ctx)

	currentVersion := entity.GetVersion()
	entity.SetVersion(currentVersion + 1)

	filter := bson.M{
		"_id":     entity.GetID(),
		"version": currentVersion,
	}

	result, err := r.collection.ReplaceOne(ctx, filter, entity)
	if err != nil {
		return fmt.Errorf("failed to update entity: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrOptimisticLock
	}

	return nil
}

// Delete removes an entity by ID
func (r *MongoGenericRepository[T]) Delete(ctx context.Context, id string) error {
	ctx = r.getContext(ctx)

	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrEntityNotFound
	}

	return nil
}

// List retrieves entities with pagination
func (r *MongoGenericRepository[T]) List(ctx context.Context, offset, limit int) ([]T, error) {
	ctx = r.getContext(ctx)

	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit))
	cursor, err := r.collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}
	defer cursor.Close(ctx)

	var results []T
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, fmt.Errorf("failed to decode entity: %w", err)
		}
		results = append(results, entity)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return results, nil
}

// Count returns the total number of entities
func (r *MongoGenericRepository[T]) Count(ctx context.Context) (int64, error) {
	ctx = r.getContext(ctx)

	count, err := r.collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, fmt.Errorf("failed to count entities: %w", err)
	}

	return count, nil
}

// FindBy finds entities matching the given filter
func (r *MongoGenericRepository[T]) FindBy(ctx context.Context, filter map[string]interface{}) ([]T, error) {
	ctx = r.getContext(ctx)

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find entities: %w", err)
	}
	defer cursor.Close(ctx)

	var results []T
	for cursor.Next(ctx) {
		var entity T
		if err := cursor.Decode(&entity); err != nil {
			return nil, fmt.Errorf("failed to decode entity: %w", err)
		}
		results = append(results, entity)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return results, nil
}

// FindOneBy finds a single entity matching the given filter
func (r *MongoGenericRepository[T]) FindOneBy(ctx context.Context, filter map[string]interface{}) (T, error) {
	ctx = r.getContext(ctx)
	var zero T

	var result T
	err := r.collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return zero, ErrEntityNotFound
		}
		return zero, fmt.Errorf("failed to find entity: %w", err)
	}

	return result, nil
}

// Exists checks if an entity exists by ID
func (r *MongoGenericRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	ctx = r.getContext(ctx)

	filter := bson.M{"_id": id}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("failed to check entity existence: %w", err)
	}

	return count > 0, nil
}

// MongoEventSourcedRepository implements EventSourcedRepository for MongoDB
type MongoEventSourcedRepository[T repository.AggregateRoot] struct {
	*MongoGenericRepository[T]
	eventCollection *mongo.Collection
}

// NewMongoEventSourcedRepository creates a new MongoDB event-sourced repository
func NewMongoEventSourcedRepository[T repository.AggregateRoot](database *mongo.Database, collectionName, eventCollectionName string) *MongoEventSourcedRepository[T] {
	return &MongoEventSourcedRepository[T]{
		MongoGenericRepository: NewMongoGenericRepository[T](database, collectionName),
		eventCollection:        database.Collection(eventCollectionName),
	}
}

// SaveEvents saves events to the event store
func (r *MongoEventSourcedRepository[T]) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	ctx = r.getContext(ctx)

	if len(events) == 0 {
		return nil
	}

	// Create event documents
	var eventDocs []interface{}
	for i, evt := range events {
		eventDoc := bson.M{
			"aggregateId":   aggregateID,
			"aggregateType": reflect.TypeOf(evt).Name(),
			"eventType":     evt.EventType(),
			"eventData":     evt,
			"version":       expectedVersion + i + 1,
			"timestamp":     evt.OccurredAt(),
		}
		eventDocs = append(eventDocs, eventDoc)
	}

	// Insert events atomically
	_, err := r.eventCollection.InsertMany(ctx, eventDocs)
	if err != nil {
		return fmt.Errorf("failed to save events: %w", err)
	}

	return nil
}

// GetEvents retrieves all events for an aggregate
func (r *MongoEventSourcedRepository[T]) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	return r.GetEventsSince(ctx, aggregateID, 0)
}

// GetEventsSince retrieves events for an aggregate since a specific version
func (r *MongoEventSourcedRepository[T]) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	ctx = r.getContext(ctx)

	filter := bson.M{
		"aggregateId": aggregateID,
		"version":     bson.M{"$gt": version},
	}

	opts := options.Find().SetSort(bson.D{{Key: "version", Value: 1}})
	cursor, err := r.eventCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []event.DomainEvent
	for cursor.Next(ctx) {
		var eventDoc bson.M
		if err := cursor.Decode(&eventDoc); err != nil {
			return nil, fmt.Errorf("failed to decode event: %w", err)
		}

		// Convert to domain event (this would need proper event deserialization)
		// For now, we'll return a placeholder - you'll need to implement proper event deserialization
		// based on your specific event types
	}

	return events, nil
}

// GetAllEvents retrieves all events from the event store
func (r *MongoEventSourcedRepository[T]) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	ctx = r.getContext(ctx)

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})
	cursor, err := r.eventCollection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get all events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []event.DomainEvent
	// Implementation would be similar to GetEventsSince
	// You'll need to implement proper event deserialization

	return events, nil
}

// LoadAggregate loads an aggregate from its event history
func (r *MongoEventSourcedRepository[T]) LoadAggregate(ctx context.Context, aggregateID string) (T, error) {
	var zero T

	events, err := r.GetEvents(ctx, aggregateID)
	if err != nil {
		return zero, fmt.Errorf("failed to load events for aggregate: %w", err)
	}

	if len(events) == 0 {
		return zero, ErrEntityNotFound
	}

	// Create new aggregate instance
	aggregate := zero
	aggregate.SetID(aggregateID)

	// Load from history
	if err := aggregate.LoadFromHistory(events); err != nil {
		return zero, fmt.Errorf("failed to load aggregate from history: %w", err)
	}

	return aggregate, nil
}

// SaveAggregate saves an aggregate by persisting its uncommitted events
func (r *MongoEventSourcedRepository[T]) SaveAggregate(ctx context.Context, aggregate T) error {
	events := aggregate.GetUncommittedEvents()
	if len(events) == 0 {
		return nil // No changes to save
	}

	expectedVersion := aggregate.GetVersion() - len(events)
	err := r.SaveEvents(ctx, aggregate.GetID(), events, expectedVersion)
	if err != nil {
		return err
	}

	// Mark events as committed
	aggregate.MarkEventsAsCommitted()
	return nil
}
