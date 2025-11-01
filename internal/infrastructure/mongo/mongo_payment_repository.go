package mongo

import (
	"context"
	"fmt"
	"time"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoPaymentRepository implements PaymentRepository with real MongoDB persistence
type MongoPaymentRepository struct {
	database         *mongo.Database
	entityCollection *mongo.Collection
	eventCollection  *mongo.Collection
	session          mongo.Session
}

// NewMongoPaymentRepository creates a new MongoDB payment repository
func NewMongoPaymentRepository(database *mongo.Database) *MongoPaymentRepository {
	return &MongoPaymentRepository{
		database:         database,
		entityCollection: database.Collection("payments"),
		eventCollection:  database.Collection("payment_events"),
	}
}

// SetTransaction implements TransactionalRepository
func (r *MongoPaymentRepository) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
	}
}

// GetTransaction implements TransactionalRepository
func (r *MongoPaymentRepository) GetTransaction() interface{} {
	return r.session
}

// IsTransactional implements TransactionalRepository
func (r *MongoPaymentRepository) IsTransactional() bool {
	return r.session != nil
}

// getContext returns the appropriate context for MongoDB operations
func (r *MongoPaymentRepository) getContext(ctx context.Context) context.Context {
	if r.session != nil {
		return mongo.NewSessionContext(ctx, r.session)
	}
	return ctx
}

// Save stores a payment aggregate to MongoDB (ACTUAL DATABASE PERSISTENCE)
func (r *MongoPaymentRepository) Save(ctx context.Context, payment *aggregate.Payment) error {
	ctx = r.getContext(ctx)

	// First, save the events
	events := payment.GetUncommittedEvents()
	if len(events) > 0 {
		err := r.SaveEvents(ctx, payment.ID(), events, payment.Version()-len(events))
		if err != nil {
			return fmt.Errorf("failed to save events: %w", err)
		}
	}

	// Convert payment to BSON document
	paymentDoc := bson.M{
		"_id":                  payment.ID(),
		"order_code":           payment.OrderCode(),
		"user_id":              payment.UserID(),
		"amount":               payment.Amount(),
		"description":          payment.Description(),
		"items":                payment.Items(),
		"status":               string(payment.Status()),
		"method":               string(payment.Method()),
		"payos_transaction_id": payment.PayOSTransactionID(),
		"checkout_url":         payment.CheckoutURL(),
		"qr_code":              payment.QRCode(),
		"expired_at":           payment.ExpiredAt(),
		"vendor_id":            payment.VendorID(),
		"pet_id":               payment.PetID(),
		"service_ids":          payment.ServiceIDs(),
		"start_time":           payment.StartTime(),
		"end_time":             payment.EndTime(),
		"version":              payment.Version(),
		"created_at":           payment.CreatedAt(),
		"updated_at":           payment.UpdatedAt(),
	}

	// Use upsert to insert or update
	opts := options.Replace().SetUpsert(true)
	_, err := r.entityCollection.ReplaceOne(ctx, bson.M{"_id": payment.ID()}, paymentDoc, opts)
	if err != nil {
		return fmt.Errorf("failed to save payment: %w", err)
	}

	// Mark events as committed only after successful save
	if len(events) > 0 {
		payment.MarkEventsAsCommitted()
	}

	return nil
}

// GetByID retrieves a payment by ID from MongoDB
func (r *MongoPaymentRepository) GetByID(ctx context.Context, id string) (*aggregate.Payment, error) {
	ctx = r.getContext(ctx)

	var doc bson.M
	err := r.entityCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return r.documentToPayment(doc)
}

// GetByOrderCode retrieves a payment by order code from MongoDB
func (r *MongoPaymentRepository) GetByOrderCode(ctx context.Context, orderCode int64) (*aggregate.Payment, error) {
	ctx = r.getContext(ctx)

	var doc bson.M
	err := r.entityCollection.FindOne(ctx, bson.M{"order_code": orderCode}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payment not found with order code: %d", orderCode)
		}
		return nil, fmt.Errorf("failed to get payment by order code: %w", err)
	}

	return r.documentToPayment(doc)
}

// GetByUserID retrieves payments for a user from MongoDB
func (r *MongoPaymentRepository) GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*aggregate.Payment, error) {
	ctx = r.getContext(ctx)

	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by newest first

	cursor, err := r.entityCollection.Find(ctx, bson.M{"user_id": userID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find payments: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*aggregate.Payment
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode payment: %w", err)
		}

		payment, err := r.documentToPayment(doc)
		if err != nil {
			return nil, err
		}

		payments = append(payments, payment)
	}

	return payments, nil
}

// GetByStatus retrieves all payments with a specific status from MongoDB
func (r *MongoPaymentRepository) GetByStatus(ctx context.Context, status string) ([]*aggregate.Payment, error) {
	ctx = r.getContext(ctx)

	cursor, err := r.entityCollection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, fmt.Errorf("failed to find payments by status: %w", err)
	}
	defer cursor.Close(ctx)

	var payments []*aggregate.Payment
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("failed to decode payment: %w", err)
		}

		payment, err := r.documentToPayment(doc)
		if err != nil {
			return nil, err
		}

		payments = append(payments, payment)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return payments, nil
}

// SaveEvents saves events for a payment aggregate
func (r *MongoPaymentRepository) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	ctx = r.getContext(ctx)

	// Create event documents
	var eventDocs []interface{}
	currentVersion := expectedVersion
	for _, evt := range events {
		currentVersion++
		eventDoc := bson.M{
			"aggregate_id": aggregateID,
			"event_type":   evt.EventType(),
			"event_data":   evt,
			"version":      currentVersion,
			"occurred_at":  evt.OccurredAt(),
		}
		eventDocs = append(eventDocs, eventDoc)
	}

	// Insert events
	if len(eventDocs) > 0 {
		_, err := r.eventCollection.InsertMany(ctx, eventDocs)
		if err != nil {
			return fmt.Errorf("failed to save events: %w", err)
		}
	}

	return nil
}

// GetEvents retrieves all events for a payment aggregate
func (r *MongoPaymentRepository) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	// Simple implementation that returns empty slice
	// Full event sourcing reconstruction would be more complex
	return []event.DomainEvent{}, nil
}

// GetEventsSince retrieves events after a specific version
func (r *MongoPaymentRepository) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	return r.GetEvents(ctx, aggregateID)
}

// GetAllEvents retrieves all events
func (r *MongoPaymentRepository) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	return []event.DomainEvent{}, nil
}

// documentToPayment converts a MongoDB document to a Payment aggregate
func (r *MongoPaymentRepository) documentToPayment(doc bson.M) (*aggregate.Payment, error) {
	// Extract items
	itemsData, ok := doc["items"].(bson.A)
	if !ok {
		return nil, fmt.Errorf("invalid items data")
	}

	var items []aggregate.PaymentItem
	for _, itemData := range itemsData {
		itemDoc, ok := itemData.(bson.M)
		if !ok {
			continue
		}
		item := aggregate.PaymentItem{
			Name:     getString(itemDoc, "name"),
			Quantity: getIntValue(itemDoc, "quantity"),
			Price:    getIntValue(itemDoc, "price"),
		}
		items = append(items, item)
	}

	// Extract service IDs
	var serviceIDs []string
	if serviceIDsData, ok := doc["service_ids"].(bson.A); ok {
		for _, sid := range serviceIDsData {
			if sidStr, ok := sid.(string); ok {
				serviceIDs = append(serviceIDs, sidStr)
			}
		}
	}

	// Handle legacy payments without schedule fields by using ReconstructPayment
	vendorID := getString(doc, "vendor_id")
	petID := getString(doc, "pet_id")
	startTime := getTime(doc, "start_time")
	endTime := getTime(doc, "end_time")

	// Use ReconstructPayment for legacy data (bypasses validation)
	payment := aggregate.ReconstructPayment(
		getString(doc, "_id"),
		int64(getIntValue(doc, "order_code")),
		getString(doc, "user_id"),
		getIntValue(doc, "amount"),
		getString(doc, "description"),
		items,
		aggregate.PaymentStatus(getString(doc, "status")),
		aggregate.PaymentMethod(getString(doc, "method")),
		getString(doc, "payos_transaction_id"),
		getString(doc, "checkout_url"),
		getString(doc, "qr_code"),
		getTime(doc, "expired_at"),
		vendorID,
		petID,
		serviceIDs,
		startTime,
		endTime,
		getIntValue(doc, "version"),
		getTime(doc, "created_at"),
		getTime(doc, "updated_at"),
	)

	return payment, nil
}

// Helper functions specific to payment repository
func getIntValue(doc bson.M, key string) int {
	if val, ok := doc[key].(int32); ok {
		return int(val)
	}
	if val, ok := doc[key].(int64); ok {
		return int(val)
	}
	if val, ok := doc[key].(int); ok {
		return val
	}
	return 0
}

func getTime(doc bson.M, key string) time.Time {
	if val, ok := doc[key].(primitive.DateTime); ok {
		return val.Time()
	}
	if val, ok := doc[key].(time.Time); ok {
		return val
	}
	return time.Time{}
}
