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

// MongoVendorRepository implements VendorRepository with real MongoDB persistence
type MongoVendorRepository struct {
	database         *mongo.Database
	entityCollection *mongo.Collection
	eventCollection  *mongo.Collection
	session          mongo.Session
}

// NewMongoVendorRepository creates a new MongoDB vendor repository
func NewMongoVendorRepository(database *mongo.Database) *MongoVendorRepository {
	return &MongoVendorRepository{
		database:         database,
		entityCollection: database.Collection("vendors"),
		eventCollection:  database.Collection("vendor_events"),
	}
}

// SetTransaction implements TransactionalRepository
func (r *MongoVendorRepository) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
	}
}

// GetTransaction implements TransactionalRepository
func (r *MongoVendorRepository) GetTransaction() interface{} {
	return r.session
}

// IsTransactional implements TransactionalRepository
func (r *MongoVendorRepository) IsTransactional() bool {
	return r.session != nil
}

// Save stores a vendor aggregate to MongoDB
func (r *MongoVendorRepository) Save(ctx context.Context, vendor *aggregate.Vendor) error {
	// Use session context if in transaction
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}

	// Prepare entity document for MongoDB
	entityDoc := bson.M{
		"_id":        vendor.GetID(),
		"version":    vendor.GetVersion(),
		"name":       vendor.Name(),
		"email":      vendor.Email(),
		"phone":      vendor.Phone(),
		"address":    vendor.Address(),
		"is_active":  vendor.IsActive(),
		"created_at": vendor.CreatedAt(),
		"updated_at": vendor.UpdatedAt(),
	}

	// Upsert entity document to MongoDB
	opts := options.Update().SetUpsert(true)
	_, err := r.entityCollection.UpdateOne(ctxToUse, bson.M{"_id": vendor.GetID()}, bson.M{"$set": entityDoc}, opts)
	if err != nil {
		return fmt.Errorf("failed to save vendor to MongoDB: %w", err)
	}

	// Save uncommitted events to MongoDB
	events := vendor.GetUncommittedEvents()
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
		vendor.MarkEventsAsCommitted()
	}

	return nil
}

// GetByID retrieves a vendor aggregate by ID from MongoDB
func (r *MongoVendorRepository) GetByID(ctx context.Context, id string) (*aggregate.Vendor, error) {
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
			return nil, fmt.Errorf("vendor not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get vendor from MongoDB: %w", err)
	}

	// Reconstruct vendor from document
	vendor, err := aggregate.NewVendor(
		getVendorString(result, "name"),
		getVendorString(result, "email"),
		getVendorString(result, "phone"),
		getVendorString(result, "address"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct vendor: %w", err)
	}

	// Set version from database
	vendor.SetVersion(getVendorInt(result, "version"))

	return vendor, nil
}

// getString safely extracts a string from a bson.M document
func getVendorString(doc bson.M, key string) string {
	if val, ok := doc[key].(string); ok {
		return val
	}
	return ""
}

// getVendorInt safely extracts an int from a bson.M document
func getVendorInt(doc bson.M, key string) int {
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

// SaveEvents saves events for a vendor aggregate
func (r *MongoVendorRepository) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	return nil // Stub implementation
}

// GetEvents retrieves all events for a vendor aggregate
func (r *MongoVendorRepository) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}

// GetEventsSince retrieves events since a specific version
func (r *MongoVendorRepository) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}

// GetAllEvents retrieves all vendor events
func (r *MongoVendorRepository) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}
