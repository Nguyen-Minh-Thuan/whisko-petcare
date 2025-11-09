package mongo

import (
	"context"
	"fmt"
	"time"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoServiceRepository implements ServiceRepository with real MongoDB persistence
type MongoServiceRepository struct {
	database         *mongo.Database
	entityCollection *mongo.Collection
	eventCollection  *mongo.Collection
	session          mongo.Session
}

// NewMongoServiceRepository creates a new MongoDB service repository
func NewMongoServiceRepository(database *mongo.Database) *MongoServiceRepository {
	return &MongoServiceRepository{
		database:         database,
		entityCollection: database.Collection("services"),
		eventCollection:  database.Collection("service_events"),
	}
}

// SetTransaction implements TransactionalRepository
func (r *MongoServiceRepository) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
	}
}

// GetTransaction implements TransactionalRepository
func (r *MongoServiceRepository) GetTransaction() interface{} {
	return r.session
}

// IsTransactional implements TransactionalRepository
func (r *MongoServiceRepository) IsTransactional() bool {
	return r.session != nil
}

// Save stores a service aggregate to MongoDB
func (r *MongoServiceRepository) Save(ctx context.Context, service *aggregate.Service) error {
	// Use session context if in transaction
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}

	// Prepare entity document for MongoDB
	entityDoc := bson.M{
		"_id":         service.GetID(),
		"version":     service.GetVersion(),
		"vendor_id":   service.VendorID(),
		"name":        service.Name(),
		"description": service.Description(),
		"price":       service.Price(),
		"duration":    service.Duration().Minutes(), // Store as minutes
		"tags":        service.Tags(),
		"is_active":   service.IsActive(),
		"created_at":  service.CreatedAt(),
		"updated_at":  service.UpdatedAt(),
	}

	// Upsert entity document to MongoDB
	opts := options.Update().SetUpsert(true)
	_, err := r.entityCollection.UpdateOne(ctxToUse, bson.M{"_id": service.GetID()}, bson.M{"$set": entityDoc}, opts)
	if err != nil {
		return fmt.Errorf("failed to save service to MongoDB: %w", err)
	}

	// Save uncommitted events to MongoDB
	events := service.GetUncommittedEvents()
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
		service.MarkEventsAsCommitted()
	}

	return nil
}

// GetByID retrieves a service aggregate by ID from MongoDB
func (r *MongoServiceRepository) GetByID(ctx context.Context, id string) (*aggregate.Service, error) {
	// Use session context if in transaction
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}

	// Find entity document
	var result bson.M
	err := r.entityCollection.FindOne(ctxToUse, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("service not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get service from MongoDB: %w", err)
	}

	// Extract duration
	durationMinutes := getServiceFloat64(result, "duration")
	duration := time.Duration(durationMinutes) * time.Minute

	// Extract tags
	tags := getServiceTags(result, "tags")

	// Reconstruct service from database state WITHOUT raising events
	service := aggregate.ReconstructService(
		getServiceString(result, "_id"),
		getServiceString(result, "vendor_id"),
		getServiceString(result, "name"),
		getServiceString(result, "description"),
		getServiceString(result, "image_url"),
		getServiceInt(result, "price"),
		duration,
		tags,
		getServiceInt(result, "version"),
		getServiceTime(result, "created_at"),
		getServiceTime(result, "updated_at"),
		getServiceBool(result, "is_active"),
	)

	return service, nil
}

// getServiceString safely extracts a string from a bson.M document
func getServiceString(doc bson.M, key string) string {
	if val, ok := doc[key].(string); ok {
		return val
	}
	return ""
}

// getServiceInt safely extracts an int from a bson.M document
func getServiceInt(doc bson.M, key string) int {
	if val, ok := doc[key].(int); ok {
		return val
	}
	if val, ok := doc[key].(int32); ok {
		return int(val)
	}
	if val, ok := doc[key].(int64); ok {
		return int(val)
	}
	return 0
}

func getServiceBool(doc bson.M, key string) bool {
	if val, ok := doc[key].(bool); ok {
		return val
	}
	return false
}

func getServiceTime(doc bson.M, key string) time.Time {
	if val, ok := doc[key].(time.Time); ok {
		return val
	}
	return time.Time{}
}

// getServiceFloat64 safely extracts a float64 from a bson.M document
func getServiceFloat64(doc bson.M, key string) float64 {
	if val, ok := doc[key].(float64); ok {
		return val
	}
	if val, ok := doc[key].(float32); ok {
		return float64(val)
	}
	if val, ok := doc[key].(int); ok {
		return float64(val)
	}
	if val, ok := doc[key].(int32); ok {
		return float64(val)
	}
	if val, ok := doc[key].(int64); ok {
		return float64(val)
	}
	return 0
}

// getServiceTags safely extracts a string slice from a bson.M document
func getServiceTags(doc bson.M, key string) []string {
	if val, ok := doc[key].([]interface{}); ok {
		tags := make([]string, 0, len(val))
		for _, v := range val {
			if str, ok := v.(string); ok {
				tags = append(tags, str)
			}
		}
		return tags
	}
	if val, ok := doc[key].([]string); ok {
		return val
	}
	return []string{}
}

// SaveEvents saves events for a service aggregate
func (r *MongoServiceRepository) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	return nil // Stub implementation
}

// GetEvents retrieves all events for a service aggregate
func (r *MongoServiceRepository) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}

// GetEventsSince retrieves events since a specific version
func (r *MongoServiceRepository) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}

// GetAllEvents retrieves all service events
func (r *MongoServiceRepository) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}
