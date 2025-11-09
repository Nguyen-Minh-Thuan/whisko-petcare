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

// VendorReadModel represents the read model for vendors
type VendorReadModel struct {
	ID        string    `bson:"_id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	Email     string    `bson:"email" json:"email"`
	Phone     string    `bson:"phone" json:"phone"`
	Address   string    `bson:"address" json:"address"`
	ImageUrl  string    `bson:"image_url" json:"image_url,omitempty"`
	IsActive  bool      `bson:"is_active" json:"is_active"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// MongoVendorProjection implements VendorProjection using MongoDB
type MongoVendorProjection struct {
	collection *mongo.Collection
}

// NewMongoVendorProjection creates a new MongoDB vendor projection
func NewMongoVendorProjection(db *mongo.Database) *MongoVendorProjection {
	collection := db.Collection("vendors")
	
	// Create indexes
	ctx := context.Background()
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "email", Value: 1},
			},
			Options: options.Index().SetUnique(false),
		},
		{
			Keys: bson.D{
				{Key: "is_active", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "name", Value: 1},
			},
		},
	}
	
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		fmt.Printf("Warning: failed to create vendor indexes: %v\n", err)
	}
	
	return &MongoVendorProjection{
		collection: collection,
	}
}

// GetByID retrieves a vendor by ID
func (p *MongoVendorProjection) GetByID(ctx context.Context, id string) (interface{}, error) {
	var vendor VendorReadModel
	err := p.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&vendor)
	if err != nil {
		return nil, err
	}
	return vendor, nil
}

// ListAll retrieves all vendors with pagination
func (p *MongoVendorProjection) ListAll(ctx context.Context, offset, limit int) ([]interface{}, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := p.collection.Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var vendors []interface{}
	for cursor.Next(ctx) {
		var vendor VendorReadModel
		if err := cursor.Decode(&vendor); err != nil {
			return nil, err
		}
		vendors = append(vendors, vendor)
	}
	
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	
	return vendors, nil
}

// HandleVendorCreated handles VendorCreated event
func (p *MongoVendorProjection) HandleVendorCreated(ctx context.Context, evt *event.VendorCreated) error {
	vendor := VendorReadModel{
		ID:        evt.VendorID,
		Name:      evt.Name,
		Email:     evt.Email,
		Phone:     evt.Phone,
		Address:   evt.Address,
		ImageUrl:  evt.ImageUrl,
		IsActive:  true,
		CreatedAt: evt.Timestamp,
		UpdatedAt: evt.Timestamp,
	}
	
	_, err := p.collection.InsertOne(ctx, vendor)
	if err != nil {
		return fmt.Errorf("failed to insert vendor: %w", err)
	}
	
	return nil
}

// HandleVendorUpdated handles VendorUpdated event
func (p *MongoVendorProjection) HandleVendorUpdated(ctx context.Context, evt *event.VendorUpdated) error {
	update := bson.M{
		"$set": bson.M{
			"name":       evt.Name,
			"email":      evt.Email,
			"phone":      evt.Phone,
			"address":    evt.Address,
			"updated_at": evt.Timestamp,
		},
	}
	
	_, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.VendorID}, update)
	if err != nil {
		return fmt.Errorf("failed to update vendor: %w", err)
	}
	
	return nil
}

// HandleVendorDeleted handles VendorDeleted event
func (p *MongoVendorProjection) HandleVendorDeleted(ctx context.Context, evt *event.VendorDeleted) error {
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": evt.Timestamp,
		},
	}
	
	_, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.VendorID}, update)
	if err != nil {
		return fmt.Errorf("failed to soft delete vendor: %w", err)
	}
	
	return nil
}

// HandleVendorImageUpdated handles VendorImageUpdated event
func (p *MongoVendorProjection) HandleVendorImageUpdated(ctx context.Context, evt *event.VendorImageUpdated) error {
	update := bson.M{
		"$set": bson.M{
			"image_url":  evt.ImageUrl,
			"updated_at": evt.Timestamp,
		},
	}
	
	_, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.VendorID}, update)
	if err != nil {
		return fmt.Errorf("failed to update vendor image: %w", err)
	}
	
	return nil
}
