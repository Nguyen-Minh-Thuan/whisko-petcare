package mongo

import (
	"context"
	"fmt"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"
	"whisko-petcare/internal/domain/repository"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoPayoutRepository implements PayoutRepository with MongoDB persistence
type MongoPayoutRepository struct {
	database         *mongo.Database
	entityCollection *mongo.Collection
	eventCollection  *mongo.Collection
	session          mongo.Session
}

// NewMongoPayoutRepository creates a new MongoDB payout repository
func NewMongoPayoutRepository(database *mongo.Database) repository.PayoutRepository {
	return &MongoPayoutRepository{
		database:         database,
		entityCollection: database.Collection("payouts"),
		eventCollection:  database.Collection("payout_events"),
	}
}

// SetTransaction implements TransactionalRepository
func (r *MongoPayoutRepository) SetTransaction(tx interface{}) {
	if session, ok := tx.(mongo.Session); ok {
		r.session = session
	}
}

// GetTransaction implements TransactionalRepository
func (r *MongoPayoutRepository) GetTransaction() interface{} {
	return r.session
}

// IsTransactional implements TransactionalRepository
func (r *MongoPayoutRepository) IsTransactional() bool {
	return r.session != nil
}

// getContext returns the appropriate context for MongoDB operations
func (r *MongoPayoutRepository) getContext(ctx context.Context) context.Context {
	if r.session != nil {
		return mongo.NewSessionContext(ctx, r.session)
	}
	return ctx
}

// Save stores a payout aggregate to MongoDB
func (r *MongoPayoutRepository) Save(ctx context.Context, payout *aggregate.Payout) error {
	fmt.Printf("💾 MongoPayoutRepository.Save: Saving payout %s\n", payout.ID())
	ctx = r.getContext(ctx)

	// First, save the events
	events := payout.GetUncommittedEvents()
	if len(events) > 0 {
		fmt.Printf("💾 Saving %d events for payout %s\n", len(events), payout.ID())
		err := r.SaveEvents(ctx, payout.ID(), events, payout.Version()-len(events))
		if err != nil {
			fmt.Printf("❌ Failed to save payout events: %v\n", err)
			return fmt.Errorf("failed to save events: %w", err)
		}
	}

	// Convert payout to BSON document
	bankAccount := payout.BankAccount()
	payoutDoc := bson.M{
		"_id":         payout.ID(),
		"vendor_id":   payout.VendorID(),
		"payment_id":  payout.PaymentID(),
		"schedule_id": payout.ScheduleID(),
		"amount":      payout.Amount(),
		"bank_account": bson.M{
			"bank_name":      bankAccount.BankName,
			"account_number": bankAccount.AccountNumber,
			"account_name":   bankAccount.AccountName,
			"bank_branch":    bankAccount.BankBranch,
		},
		"status":         string(payout.Status()),
		"notes":          payout.Notes(),
		"failure_reason": payout.FailureReason(),
		"version":        payout.Version(),
		"created_at":     payout.CreatedAt(),
		"updated_at":     payout.UpdatedAt(),
	}

	// Use upsert to insert or update
	opts := options.Replace().SetUpsert(true)
	_, err := r.entityCollection.ReplaceOne(ctx, bson.M{"_id": payout.ID()}, payoutDoc, opts)
	if err != nil {
		fmt.Printf("❌ Failed to save payout to MongoDB: %v\n", err)
		return fmt.Errorf("failed to save payout: %w", err)
	}

	fmt.Printf("✅ MongoPayoutRepository.Save: Payout %s saved successfully\n", payout.ID())
	return nil
}

// GetByID retrieves a payout by ID
func (r *MongoPayoutRepository) GetByID(ctx context.Context, id string) (*aggregate.Payout, error) {
	fmt.Printf("🔍 MongoPayoutRepository.GetByID: Looking for payout %s\n", id)
	ctx = r.getContext(ctx)

	var result bson.M
	err := r.entityCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			fmt.Printf("❌ Payout not found: %s\n", id)
			return nil, fmt.Errorf("payout not found: %s", id)
		}
		fmt.Printf("❌ Failed to get payout: %v\n", err)
		return nil, fmt.Errorf("failed to get payout: %w", err)
	}

	// Extract bank account
	var bankAccount aggregate.BankAccount
	if bankAccountDoc, ok := result["bank_account"].(bson.M); ok {
		bankAccount = aggregate.BankAccount{
			BankName:      getString(bankAccountDoc, "bank_name"),
			AccountNumber: getString(bankAccountDoc, "account_number"),
			AccountName:   getString(bankAccountDoc, "account_name"),
			BankBranch:    getString(bankAccountDoc, "bank_branch"),
		}
	}

	// Reconstruct payout from database state
	payout := aggregate.ReconstructPayout(
		getString(result, "_id"),
		getString(result, "vendor_id"),
		getString(result, "payment_id"),
		getString(result, "schedule_id"),
		getInt(result, "amount"),
		bankAccount,
		getString(result, "status"),
		getString(result, "notes"),
		getString(result, "failure_reason"),
		getInt(result, "version"),
		getTime(result, "created_at"),
		getTime(result, "updated_at"),
	)

	fmt.Printf("✅ Payout found: %s (Status: %s)\n", id, payout.Status())
	return payout, nil
}

// GetByVendorID retrieves payouts by vendor ID with pagination
func (r *MongoPayoutRepository) GetByVendorID(ctx context.Context, vendorID string, offset, limit int) ([]*aggregate.Payout, error) {
	ctx = r.getContext(ctx)

	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.entityCollection.Find(ctx, bson.M{"vendor_id": vendorID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find payouts: %w", err)
	}
	defer cursor.Close(ctx)

	var payouts []*aggregate.Payout
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode payout: %w", err)
		}

		var bankAccount aggregate.BankAccount
		if bankAccountDoc, ok := result["bank_account"].(bson.M); ok {
			bankAccount = aggregate.BankAccount{
				BankName:      getString(bankAccountDoc, "bank_name"),
				AccountNumber: getString(bankAccountDoc, "account_number"),
				AccountName:   getString(bankAccountDoc, "account_name"),
				BankBranch:    getString(bankAccountDoc, "bank_branch"),
			}
		}

		payout := aggregate.ReconstructPayout(
			getString(result, "_id"),
			getString(result, "vendor_id"),
			getString(result, "payment_id"),
			getString(result, "schedule_id"),
			getInt(result, "amount"),
			bankAccount,
			getString(result, "status"),
			getString(result, "notes"),
			getString(result, "failure_reason"),
			getInt(result, "version"),
			getTime(result, "created_at"),
			getTime(result, "updated_at"),
		)
		payouts = append(payouts, payout)
	}

	return payouts, nil
}

// GetByPaymentID retrieves payout by payment ID
func (r *MongoPayoutRepository) GetByPaymentID(ctx context.Context, paymentID string) (*aggregate.Payout, error) {
	ctx = r.getContext(ctx)

	var result bson.M
	err := r.entityCollection.FindOne(ctx, bson.M{"payment_id": paymentID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payout not found for payment: %s", paymentID)
		}
		return nil, fmt.Errorf("failed to get payout: %w", err)
	}

	var bankAccount aggregate.BankAccount
	if bankAccountDoc, ok := result["bank_account"].(bson.M); ok {
		bankAccount = aggregate.BankAccount{
			BankName:      getString(bankAccountDoc, "bank_name"),
			AccountNumber: getString(bankAccountDoc, "account_number"),
			AccountName:   getString(bankAccountDoc, "account_name"),
			BankBranch:    getString(bankAccountDoc, "bank_branch"),
		}
	}

	payout := aggregate.ReconstructPayout(
		getString(result, "_id"),
		getString(result, "vendor_id"),
		getString(result, "payment_id"),
		getString(result, "schedule_id"),
		getInt(result, "amount"),
		bankAccount,
		getString(result, "status"),
		getString(result, "notes"),
		getString(result, "failure_reason"),
		getInt(result, "version"),
		getTime(result, "created_at"),
		getTime(result, "updated_at"),
	)

	return payout, nil
}

// GetByScheduleID retrieves payout by schedule ID
func (r *MongoPayoutRepository) GetByScheduleID(ctx context.Context, scheduleID string) (*aggregate.Payout, error) {
	ctx = r.getContext(ctx)

	var result bson.M
	err := r.entityCollection.FindOne(ctx, bson.M{"schedule_id": scheduleID}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("payout not found for schedule: %s", scheduleID)
		}
		return nil, fmt.Errorf("failed to get payout: %w", err)
	}

	var bankAccount aggregate.BankAccount
	if bankAccountDoc, ok := result["bank_account"].(bson.M); ok {
		bankAccount = aggregate.BankAccount{
			BankName:      getString(bankAccountDoc, "bank_name"),
			AccountNumber: getString(bankAccountDoc, "account_number"),
			AccountName:   getString(bankAccountDoc, "account_name"),
			BankBranch:    getString(bankAccountDoc, "bank_branch"),
		}
	}

	payout := aggregate.ReconstructPayout(
		getString(result, "_id"),
		getString(result, "vendor_id"),
		getString(result, "payment_id"),
		getString(result, "schedule_id"),
		getInt(result, "amount"),
		bankAccount,
		getString(result, "status"),
		getString(result, "notes"),
		getString(result, "failure_reason"),
		getInt(result, "version"),
		getTime(result, "created_at"),
		getTime(result, "updated_at"),
	)

	return payout, nil
}

// GetByStatus retrieves payouts by status with pagination
func (r *MongoPayoutRepository) GetByStatus(ctx context.Context, status aggregate.PayoutStatus, offset, limit int) ([]*aggregate.Payout, error) {
	ctx = r.getContext(ctx)

	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.entityCollection.Find(ctx, bson.M{"status": string(status)}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find payouts: %w", err)
	}
	defer cursor.Close(ctx)

	var payouts []*aggregate.Payout
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode payout: %w", err)
		}

		var bankAccount aggregate.BankAccount
		if bankAccountDoc, ok := result["bank_account"].(bson.M); ok {
			bankAccount = aggregate.BankAccount{
				BankName:      getString(bankAccountDoc, "bank_name"),
				AccountNumber: getString(bankAccountDoc, "account_number"),
				AccountName:   getString(bankAccountDoc, "account_name"),
				BankBranch:    getString(bankAccountDoc, "bank_branch"),
			}
		}

		payout := aggregate.ReconstructPayout(
			getString(result, "_id"),
			getString(result, "vendor_id"),
			getString(result, "payment_id"),
			getString(result, "schedule_id"),
			getInt(result, "amount"),
			bankAccount,
			getString(result, "status"),
			getString(result, "notes"),
			getString(result, "failure_reason"),
			getInt(result, "version"),
			getTime(result, "created_at"),
			getTime(result, "updated_at"),
		)
		payouts = append(payouts, payout)
	}

	return payouts, nil
}

// SaveEvents saves domain events for a payout
func (r *MongoPayoutRepository) SaveEvents(ctx context.Context, aggregateID string, events []event.DomainEvent, expectedVersion int) error {
	ctx = r.getContext(ctx)

	if len(events) == 0 {
		return nil
	}

	var eventDocs []interface{}
	for i, e := range events {
		eventDoc := bson.M{
			"aggregate_id":  aggregateID,
			"event_type":    e.EventType(),
			"event_version": expectedVersion + i + 1,
			"occurred_at":   e.OccurredAt(),
			"event_data":    e,
		}
		eventDocs = append(eventDocs, eventDoc)
	}

	_, err := r.eventCollection.InsertMany(ctx, eventDocs)
	if err != nil {
		return fmt.Errorf("failed to save payout events: %w", err)
	}

	return nil
}

// GetAllEvents retrieves all events for all payouts
func (r *MongoPayoutRepository) GetAllEvents(ctx context.Context) ([]event.DomainEvent, error) {
	ctx = r.getContext(ctx)

	cursor, err := r.eventCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []event.DomainEvent
	for cursor.Next(ctx) {
		var eventDoc bson.M
		if err := cursor.Decode(&eventDoc); err != nil {
			return nil, fmt.Errorf("failed to decode event: %w", err)
		}

		// Parse event_data back to domain event
		if eventData, ok := eventDoc["event_data"].(map[string]interface{}); ok {
			// This is simplified - in production you'd need proper event deserialization
			_ = eventData
			// For now, skip event reconstruction as it's optional for payout
		}
	}

	return events, nil
}

// GetEvents retrieves all events for a specific payout
func (r *MongoPayoutRepository) GetEvents(ctx context.Context, aggregateID string) ([]event.DomainEvent, error) {
	ctx = r.getContext(ctx)

	cursor, err := r.eventCollection.Find(ctx, bson.M{"aggregate_id": aggregateID})
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []event.DomainEvent
	for cursor.Next(ctx) {
		var eventDoc bson.M
		if err := cursor.Decode(&eventDoc); err != nil {
			return nil, fmt.Errorf("failed to decode event: %w", err)
		}

		// Parse event_data back to domain event
		if eventData, ok := eventDoc["event_data"].(map[string]interface{}); ok {
			// This is simplified - in production you'd need proper event deserialization
			_ = eventData
			// For now, skip event reconstruction as it's optional for payout
		}
	}

	return events, nil
}

// GetEventsSince retrieves events for a payout since a specific version
func (r *MongoPayoutRepository) GetEventsSince(ctx context.Context, aggregateID string, version int) ([]event.DomainEvent, error) {
	ctx = r.getContext(ctx)

	filter := bson.M{
		"aggregate_id":  aggregateID,
		"event_version": bson.M{"$gt": version},
	}

	cursor, err := r.eventCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find events: %w", err)
	}
	defer cursor.Close(ctx)

	var events []event.DomainEvent
	for cursor.Next(ctx) {
		var eventDoc bson.M
		if err := cursor.Decode(&eventDoc); err != nil {
			return nil, fmt.Errorf("failed to decode event: %w", err)
		}

		// Parse event_data back to domain event
		if eventData, ok := eventDoc["event_data"].(map[string]interface{}); ok {
			// This is simplified - in production you'd need proper event deserialization
			_ = eventData
			// For now, skip event reconstruction as it's optional for payout
		}
	}

	return events, nil
}

// GetPendingPayoutForVendor retrieves pending payout for a vendor
func (r *MongoPayoutRepository) GetPendingPayoutForVendor(ctx context.Context, vendorID string) (*aggregate.Payout, error) {
	ctx = r.getContext(ctx)

	var result bson.M
	err := r.entityCollection.FindOne(ctx, bson.M{
		"vendor_id": vendorID,
		"status":    string(aggregate.PayoutStatusPending),
	}).Decode(&result)
	
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // No pending payout found
		}
		return nil, fmt.Errorf("failed to get pending payout: %w", err)
	}

	var bankAccount aggregate.BankAccount
	if bankAccountDoc, ok := result["bank_account"].(bson.M); ok {
		bankAccount = aggregate.BankAccount{
			BankName:      getString(bankAccountDoc, "bank_name"),
			AccountNumber: getString(bankAccountDoc, "account_number"),
			AccountName:   getString(bankAccountDoc, "account_name"),
			BankBranch:    getString(bankAccountDoc, "bank_branch"),
		}
	}

	payout := aggregate.ReconstructPayout(
		getString(result, "_id"),
		getString(result, "vendor_id"),
		getString(result, "payment_id"),
		getString(result, "schedule_id"),
		getInt(result, "amount"),
		bankAccount,
		getString(result, "status"),
		getString(result, "notes"),
		getString(result, "failure_reason"),
		getInt(result, "version"),
		getTime(result, "created_at"),
		getTime(result, "updated_at"),
	)

	return payout, nil
}
