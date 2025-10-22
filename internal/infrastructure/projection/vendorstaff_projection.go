package projection

import (
	"context"
	"fmt"
	"time"

	"whisko-petcare/internal/domain/event"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// VendorStaffReadModel represents the read model for vendor staff
type VendorStaffReadModel struct {
	ID        string    `bson:"_id" json:"id"`
	UserID    string    `bson:"user_id" json:"user_id"`
	VendorID  string    `bson:"vendor_id" json:"vendor_id"`
	IsActive  bool      `bson:"is_active" json:"is_active"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// MongoVendorStaffProjection implements VendorStaffProjection using MongoDB
type MongoVendorStaffProjection struct {
	collection *mongo.Collection
}

// NewMongoVendorStaffProjection creates a new MongoDB vendor staff projection
func NewMongoVendorStaffProjection(db *mongo.Database) *MongoVendorStaffProjection {
	collection := db.Collection("vendor_staffs")
	
	// Create indexes
	ctx := context.Background()
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "vendor_id", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "is_active", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "vendor_id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	}
	
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		fmt.Printf("Warning: failed to create vendor staff indexes: %v\n", err)
	}
	
	return &MongoVendorStaffProjection{
		collection: collection,
	}
}

// GetByID retrieves a vendor staff by composite ID (userID-vendorID)
func (p *MongoVendorStaffProjection) GetByID(ctx context.Context, userID, vendorID string) (interface{}, error) {
	compositeID := userID + "-" + vendorID
	var vendorStaff VendorStaffReadModel
	err := p.collection.FindOne(ctx, bson.M{"_id": compositeID}).Decode(&vendorStaff)
	if err != nil {
		return nil, err
	}
	return vendorStaff, nil
}

// GetByVendorID retrieves vendor staffs by vendor ID with pagination
func (p *MongoVendorStaffProjection) GetByVendorID(ctx context.Context, vendorID string, offset, limit int) ([]interface{}, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := p.collection.Find(ctx, bson.M{
		"vendor_id": vendorID,
		"is_active": true,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var vendorStaffs []interface{}
	for cursor.Next(ctx) {
		var vendorStaff VendorStaffReadModel
		if err := cursor.Decode(&vendorStaff); err != nil {
			return nil, err
		}
		vendorStaffs = append(vendorStaffs, vendorStaff)
	}
	
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	
	return vendorStaffs, nil
}

// GetByUserID retrieves vendor staffs by user ID with pagination
func (p *MongoVendorStaffProjection) GetByUserID(ctx context.Context, userID string, offset, limit int) ([]interface{}, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := p.collection.Find(ctx, bson.M{
		"user_id":   userID,
		"is_active": true,
	}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var vendorStaffs []interface{}
	for cursor.Next(ctx) {
		var vendorStaff VendorStaffReadModel
		if err := cursor.Decode(&vendorStaff); err != nil {
			return nil, err
		}
		vendorStaffs = append(vendorStaffs, vendorStaff)
	}
	
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	
	return vendorStaffs, nil
}

// ListAll retrieves all vendor staffs with pagination
func (p *MongoVendorStaffProjection) ListAll(ctx context.Context, offset, limit int) ([]interface{}, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := p.collection.Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var vendorStaffs []interface{}
	for cursor.Next(ctx) {
		var vendorStaff VendorStaffReadModel
		if err := cursor.Decode(&vendorStaff); err != nil {
			return nil, err
		}
		vendorStaffs = append(vendorStaffs, vendorStaff)
	}
	
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	
	return vendorStaffs, nil
}

// HandleVendorStaffCreated handles VendorStaffCreated event
func (p *MongoVendorStaffProjection) HandleVendorStaffCreated(ctx context.Context, evt event.VendorStaffCreated) error {
	compositeID := evt.UserID + "-" + evt.VendorID
	
	vendorStaff := VendorStaffReadModel{
		ID:        compositeID,
		UserID:    evt.UserID,
		VendorID:  evt.VendorID,
		IsActive:  true,
		CreatedAt: evt.Timestamp,
		UpdatedAt: evt.Timestamp,
	}
	
	_, err := p.collection.InsertOne(ctx, vendorStaff)
	if err != nil {
		return fmt.Errorf("failed to insert vendor staff: %w", err)
	}
	
	return nil
}

// HandleVendorStaffDeleted handles VendorStaffDeleted event
func (p *MongoVendorStaffProjection) HandleVendorStaffDeleted(ctx context.Context, evt event.VendorStaffDeleted) error {
	compositeID := evt.UserID + "-" + evt.VendorID
	
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": evt.Timestamp,
		},
	}
	
	_, err := p.collection.UpdateOne(ctx, bson.M{"_id": compositeID}, update)
	if err != nil {
		return fmt.Errorf("failed to soft delete vendor staff: %w", err)
	}
	
	return nil
}
