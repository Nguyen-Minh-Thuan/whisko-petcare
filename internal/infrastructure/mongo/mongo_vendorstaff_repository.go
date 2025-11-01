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

// MongoVendorStaffRepository implements VendorStaffRepository with real MongoDB persistence
type MongoVendorStaffRepository struct {
	database         *mongo.Database
	entityCollection *mongo.Collection
	eventCollection  *mongo.Collection
	session          mongo.Session
}

// NewMongoVendorStaffRepository creates a new MongoDB vendor staff repository
func NewMongoVendorStaffRepository(database *mongo.Database) *MongoVendorStaffRepository {
	return &MongoVendorStaffRepository{
		database:         database,
		entityCollection: database.Collection("vendor_staffs"),
		eventCollection:  database.Collection("vendor_staff_events"),
	}
}

// SetTransaction implements TransactionalRepository
func (r *MongoVendorStaffRepository) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
	}
}

// GetTransaction implements TransactionalRepository
func (r *MongoVendorStaffRepository) GetTransaction() interface{} {
	return r.session
}

// IsTransactional implements TransactionalRepository
func (r *MongoVendorStaffRepository) IsTransactional() bool {
	return r.session != nil
}

// Save stores a vendor staff aggregate to MongoDB
func (r *MongoVendorStaffRepository) Save(ctx context.Context, vendorStaff *aggregate.VendorStaff) error {
	// Use session context if in transaction
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}

	// Prepare entity document for MongoDB
	entityDoc := bson.M{
		"_id":        vendorStaff.GetID(),
		"version":    vendorStaff.GetVersion(),
		"user_id":    vendorStaff.UserID(),
		"vendor_id":  vendorStaff.VendorID(),
		"role":       string(vendorStaff.Role()),
		"is_active":  vendorStaff.IsActive(),
		"created_at": vendorStaff.CreatedAt(),
		"updated_at": vendorStaff.UpdatedAt(),
	}

	// Upsert entity document to MongoDB
	opts := options.Update().SetUpsert(true)
	_, err := r.entityCollection.UpdateOne(ctxToUse, bson.M{"_id": vendorStaff.GetID()}, bson.M{"$set": entityDoc}, opts)
	if err != nil {
		return fmt.Errorf("failed to save vendor staff to MongoDB: %w", err)
	}

	// Save uncommitted events to MongoDB
	events := vendorStaff.GetUncommittedEvents()
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
		vendorStaff.MarkEventsAsCommitted()
	}

	return nil
}

// GetByID retrieves a vendor staff aggregate by ID from MongoDB
func (r *MongoVendorStaffRepository) GetByID(ctx context.Context, id string) (*aggregate.VendorStaff, error) {
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
			return nil, fmt.Errorf("vendor staff not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get vendor staff from MongoDB: %w", err)
	}

	// Reconstruct vendor staff from document
	role := getVendorStaffString(result, "role")
	if role == "" {
		role = string(aggregate.VendorStaffRoleStaff) // Default if missing
	}
	vendorStaff, err := aggregate.NewVendorStaff(
		getVendorStaffString(result, "user_id"),
		getVendorStaffString(result, "vendor_id"),
		aggregate.VendorStaffRole(role),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct vendor staff: %w", err)
	}

	// Set version from database
	vendorStaff.SetVersion(getVendorStaffInt(result, "version"))

	return vendorStaff, nil
}

// getVendorStaffString safely extracts a string from a bson.M document
func getVendorStaffString(doc bson.M, key string) string {
	if val, ok := doc[key].(string); ok {
		return val
	}
	return ""
}

// getVendorStaffInt safely extracts an int from a bson.M document
func getVendorStaffInt(doc bson.M, key string) int {
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

// SaveEvents saves events for a vendor staff aggregate
func (r *MongoVendorStaffRepository) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	return nil // Stub implementation
}

// GetEvents retrieves all events for a vendor staff aggregate
func (r *MongoVendorStaffRepository) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}

// GetEventsSince retrieves events since a specific version
func (r *MongoVendorStaffRepository) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}

// GetAllEvents retrieves all vendor staff events
func (r *MongoVendorStaffRepository) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}
