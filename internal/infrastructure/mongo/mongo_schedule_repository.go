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

// MongoScheduleRepository implements ScheduleRepository with real MongoDB persistence
type MongoScheduleRepository struct {
	database         *mongo.Database
	entityCollection *mongo.Collection
	eventCollection  *mongo.Collection
	session          mongo.Session
}

// NewMongoScheduleRepository creates a new MongoDB schedule repository
func NewMongoScheduleRepository(database *mongo.Database) *MongoScheduleRepository {
	return &MongoScheduleRepository{
		database:         database,
		entityCollection: database.Collection("schedules"),
		eventCollection:  database.Collection("schedule_events"),
	}
}

// SetTransaction implements TransactionalRepository
func (r *MongoScheduleRepository) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
	}
}

// GetTransaction implements TransactionalRepository
func (r *MongoScheduleRepository) GetTransaction() interface{} {
	return r.session
}

// IsTransactional implements TransactionalRepository
func (r *MongoScheduleRepository) IsTransactional() bool {
	return r.session != nil
}

// Save stores a schedule aggregate to MongoDB
func (r *MongoScheduleRepository) Save(ctx context.Context, schedule *aggregate.Schedule) error {
	// Use session context if in transaction
	var ctxToUse context.Context = ctx
	if r.session != nil {
		ctxToUse = mongo.NewSessionContext(ctx, r.session)
	}

	// Prepare entity document for MongoDB
	entityDoc := bson.M{
		"_id":          schedule.GetID(),
		"version":      schedule.GetVersion(),
		"booking_user": schedule.BookingUser(),
		"booked_shop":  schedule.BookedShop(),
		"assigned_pet": schedule.AssignedPet(),
		"start_time":   schedule.StartTime(),
		"end_time":     schedule.EndTime(),
		"status":       schedule.Status(),
		"is_active":    schedule.IsActive(),
		"created_at":   schedule.CreatedAt(),
		"updated_at":   schedule.UpdatedAt(),
	}

	// Upsert entity document to MongoDB
	opts := options.Update().SetUpsert(true)
	_, err := r.entityCollection.UpdateOne(ctxToUse, bson.M{"_id": schedule.GetID()}, bson.M{"$set": entityDoc}, opts)
	if err != nil {
		return fmt.Errorf("failed to save schedule to MongoDB: %w", err)
	}

	// Save uncommitted events to MongoDB
	events := schedule.GetUncommittedEvents()
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
		schedule.MarkEventsAsCommitted()
	}

	return nil
}

// GetByID retrieves a schedule aggregate by ID from MongoDB
func (r *MongoScheduleRepository) GetByID(ctx context.Context, id string) (*aggregate.Schedule, error) {
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
			return nil, fmt.Errorf("schedule not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get schedule from MongoDB: %w", err)
	}

	// Extract nested objects
	var bookingUser aggregate.BookingUser
	if bu, ok := result["booking_user"].(bson.M); ok {
		bookingUser = aggregate.BookingUser{
			UserID:  getScheduleString(bu, "user_id"),
			Name:    getScheduleString(bu, "name"),
			Email:   getScheduleString(bu, "email"),
			Phone:   getScheduleString(bu, "phone"),
			Address: getScheduleString(bu, "address"),
		}
	}

	var bookedShop aggregate.BookedShop
	if bs, ok := result["booked_shop"].(bson.M); ok {
		bookedShop = aggregate.BookedShop{
			ShopID:   getScheduleString(bs, "shop_id"),
			Name:     getScheduleString(bs, "name"),
			Location: getScheduleString(bs, "location"),
			Phone:    getScheduleString(bs, "phone"),
		}
		// Extract services
		if services, ok := bs["booked_services"].(bson.A); ok {
			for _, svc := range services {
				if svcDoc, ok := svc.(bson.M); ok {
					bookedShop.BookedServices = append(bookedShop.BookedServices, aggregate.BookedServices{
						ServiceID: getScheduleString(svcDoc, "service_id"),
						Name:      getScheduleString(svcDoc, "name"),
					})
				}
			}
		}
	}

	var assignedPet aggregate.PetAssigned
	if ap, ok := result["assigned_pet"].(bson.M); ok {
		assignedPet = aggregate.PetAssigned{
			PetID:   getScheduleString(ap, "pet_id"),
			Name:    getScheduleString(ap, "name"),
			Species: getScheduleString(ap, "species"),
			Breed:   getScheduleString(ap, "breed"),
			Age:     getScheduleInt(ap, "age"),
			Weight:  getScheduleFloat64(ap, "weight"),
		}
	}

	// Extract times
	startTime, _ := result["start_time"].(time.Time)
	endTime, _ := result["end_time"].(time.Time)

	// Reconstruct schedule from document
	schedule, err := aggregate.NewSchedule(bookingUser, bookedShop, assignedPet, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to reconstruct schedule: %w", err)
	}

	// Set version from database
	schedule.SetVersion(getScheduleInt(result, "version"))

	return schedule, nil
}

// getScheduleString safely extracts a string from a bson.M document
func getScheduleString(doc bson.M, key string) string {
	if val, ok := doc[key].(string); ok {
		return val
	}
	return ""
}

// getScheduleInt safely extracts an int from a bson.M document
func getScheduleInt(doc bson.M, key string) int {
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

// getScheduleFloat64 safely extracts a float64 from a bson.M document
func getScheduleFloat64(doc bson.M, key string) float64 {
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

// SaveEvents saves events for a schedule aggregate
func (r *MongoScheduleRepository) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	return nil // Stub implementation
}

// GetEvents retrieves all events for a schedule aggregate
func (r *MongoScheduleRepository) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}

// GetEventsSince retrieves events since a specific version
func (r *MongoScheduleRepository) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}

// GetAllEvents retrieves all schedule events
func (r *MongoScheduleRepository) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil // Stub implementation
}
