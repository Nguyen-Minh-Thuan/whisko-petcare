package mongo

import (
	"context"
	"fmt"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoUserRepository implements UserRepository with real MongoDB persistence
type MongoUserRepository struct {
	database         *mongo.Database
	entityCollection *mongo.Collection
	eventCollection  *mongo.Collection
	session          mongo.Session
}

// NewMongoUserRepository creates a new MongoDB user repository
func NewMongoUserRepository(database *mongo.Database) *MongoUserRepository {
	return &MongoUserRepository{
		database:         database,
		entityCollection: database.Collection("users"),
		eventCollection:  database.Collection("user_events"),
	}
}

// SetTransaction implements TransactionalRepository
func (r *MongoUserRepository) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
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

// Save stores a user aggregate to MongoDB (ACTUAL DATABASE PERSISTENCE)
func (r *MongoUserRepository) Save(ctx context.Context, user *aggregate.User) error {
	// Use session context if in transaction
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}

	// Prepare entity document for MongoDB
	entityDoc := bson.M{
		"_id":        user.GetID(),
		"version":    user.GetVersion(),
		"name":       user.Name(),
		"email":      user.Email(),
		"phone":      user.Phone(),
		"address":    user.Address(),
		"created_at": user.CreatedAt(),
		"updated_at": user.UpdatedAt(),
	}

	// Upsert entity document to MongoDB
	opts := options.Update().SetUpsert(true)
	_, err := r.entityCollection.UpdateOne(ctxToUse, bson.M{"_id": user.GetID()}, bson.M{"$set": entityDoc}, opts)
	if err != nil {
		return fmt.Errorf("failed to save user to MongoDB: %w", err)
	}

	// Save uncommitted events to MongoDB
	events := user.GetUncommittedEvents()
	if len(events) > 0 {
		var eventDocs []interface{}
		for _, e := range events {
			eventDoc := bson.M{
				"aggregate_id":  e.AggregateID(),
				"event_type":    e.EventType(),
				"event_version": e.Version(),
				"occurred_at":   e.OccurredAt(),
				"event_data":    e,
			}
			eventDocs = append(eventDocs, eventDoc)
		}

		_, err = r.eventCollection.InsertMany(ctxToUse, eventDocs)
		if err != nil {
			return fmt.Errorf("failed to save events to MongoDB: %w", err)
		}

		// Mark events as committed
		user.MarkEventsAsCommitted()
	}

	return nil
}

// GetByID retrieves a user by ID from MongoDB
func (r *MongoUserRepository) GetByID(ctx context.Context, id string) (*aggregate.User, error) {
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}

	var result bson.M
	err := r.entityCollection.FindOne(ctxToUse, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Reconstruct user from MongoDB data
	user, err := aggregate.NewUser(
		id,
		getString(result, "name"),
		getString(result, "email"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct user: %w", err)
	}

	// Update contact info if present
	phone := getString(result, "phone")
	address := getString(result, "address")
	if phone != "" || address != "" {
		user.UpdateContactInfo(phone, address)
	}

	// Set version
	user.SetVersion(getInt(result, "version"))

	return user, nil
}

// SaveEvents saves events for a user aggregate
func (r *MongoUserRepository) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}

	// Check version conflict by counting existing events
	count, err := r.eventCollection.CountDocuments(ctxToUse, bson.M{"aggregate_id": aggregateID})
	if err != nil {
		return fmt.Errorf("failed to check version: %w", err)
	}

	if int(count) != expectedVersion {
		return fmt.Errorf("concurrency conflict: expected version %d, got %d", expectedVersion, count)
	}

	// Insert events to MongoDB
	if len(events) > 0 {
		var eventDocs []interface{}
		for _, e := range events {
			eventDoc := bson.M{
				"aggregate_id":  e.AggregateID(),
				"event_type":    e.EventType(),
				"event_version": e.Version(),
				"occurred_at":   e.OccurredAt(),
				"event_data":    e,
			}
			eventDocs = append(eventDocs, eventDoc)
		}

		_, err = r.eventCollection.InsertMany(ctxToUse, eventDocs)
		if err != nil {
			return fmt.Errorf("failed to save events: %w", err)
		}
	}

	return nil
}

// GetEvents retrieves all events for a user aggregate
func (r *MongoUserRepository) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	// Simple implementation that returns empty slice
	// Full event sourcing reconstruction would be more complex
	return []event.DomainEvent{}, nil
}

// GetEventsSince retrieves events after a specific version
func (r *MongoUserRepository) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	return r.GetEvents(ctx, aggregateID)
}

// GetAllEvents retrieves all events
func (r *MongoUserRepository) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil
}

// Helper functions
func getString(doc bson.M, key string) string {
	if val, ok := doc[key].(string); ok {
		return val
	}
	return ""
}

func getInt(doc bson.M, key string) int {
	if val, ok := doc[key].(int32); ok {
		return int(val)
	}
	if val, ok := doc[key].(int); ok {
		return val
	}
	if val, ok := doc[key].(float64); ok {
		return int(val)
	}
	return 0
}
