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

// ServiceReadModel represents the read model for services
type ServiceReadModel struct {
	ID          string    `bson:"_id" json:"id"`
	VendorID    string    `bson:"vendor_id" json:"vendor_id"`
	Name        string    `bson:"name" json:"name"`
	Description string    `bson:"description" json:"description"`
	Price       int       `bson:"price" json:"price"`             // Price in VND
	Duration    int       `bson:"duration" json:"duration"`       // Duration in minutes
	Tags        []string  `bson:"tags" json:"tags"`
	IsActive    bool      `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

// MongoServiceProjection implements ServiceProjection using MongoDB
type MongoServiceProjection struct {
	collection *mongo.Collection
}

// NewMongoServiceProjection creates a new MongoDB service projection
func NewMongoServiceProjection(db *mongo.Database) *MongoServiceProjection {
	collection := db.Collection("services")
	
	// Create indexes
	ctx := context.Background()
	indexes := []mongo.IndexModel{
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
				{Key: "name", Value: 1},
			},
		},
		{
			Keys: bson.D{
				{Key: "price", Value: 1},
			},
		},
	}
	
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		fmt.Printf("Warning: failed to create service indexes: %v\n", err)
	}
	
	return &MongoServiceProjection{
		collection: collection,
	}
}

// GetByID retrieves a service by ID
func (p *MongoServiceProjection) GetByID(ctx context.Context, id string) (interface{}, error) {
	var service ServiceReadModel
	err := p.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// GetByVendorID retrieves services by vendor ID with pagination
func (p *MongoServiceProjection) GetByVendorID(ctx context.Context, vendorID string, offset, limit int) ([]interface{}, error) {
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
	
	var services []interface{}
	for cursor.Next(ctx) {
		var service ServiceReadModel
		if err := cursor.Decode(&service); err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	
	return services, nil
}

// ListAll retrieves all services with pagination
func (p *MongoServiceProjection) ListAll(ctx context.Context, offset, limit int) ([]interface{}, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cursor, err := p.collection.Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	var services []interface{}
	for cursor.Next(ctx) {
		var service ServiceReadModel
		if err := cursor.Decode(&service); err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	
	return services, nil
}

// HandleServiceCreated handles ServiceCreated event
func (p *MongoServiceProjection) HandleServiceCreated(ctx context.Context, evt event.ServiceCreated) error {
	// Convert duration from time.Duration to minutes
	durationMinutes := int(evt.Duration.Minutes())
	
	service := ServiceReadModel{
		ID:          evt.ServiceID,
		VendorID:    evt.VendorID,
		Name:        evt.Name,
		Description: evt.Description,
		Price:       evt.Price,
		Duration:    durationMinutes,
		Tags:        evt.Tags,
		IsActive:    true,
		CreatedAt:   evt.Timestamp,
		UpdatedAt:   evt.Timestamp,
	}
	
	_, err := p.collection.InsertOne(ctx, service)
	if err != nil {
		return fmt.Errorf("failed to insert service: %w", err)
	}
	
	return nil
}

// HandleServiceUpdated handles ServiceUpdated event
func (p *MongoServiceProjection) HandleServiceUpdated(ctx context.Context, evt event.ServiceUpdated) error {
	// Convert duration from time.Duration to minutes
	durationMinutes := int(evt.Duration.Minutes())
	
	update := bson.M{
		"$set": bson.M{
			"name":        evt.Name,
			"description": evt.Description,
			"price":       evt.Price,
			"duration":    durationMinutes,
			"tags":        evt.Tags,
			"updated_at":  evt.Timestamp,
		},
	}
	
	_, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.ServiceID}, update)
	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}
	
	return nil
}

// HandleServiceDeleted handles ServiceDeleted event
func (p *MongoServiceProjection) HandleServiceDeleted(ctx context.Context, evt event.ServiceDeleted) error {
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": evt.Timestamp,
		},
	}
	
	_, err := p.collection.UpdateOne(ctx, bson.M{"_id": evt.ServiceID}, update)
	if err != nil {
		return fmt.Errorf("failed to soft delete service: %w", err)
	}
	
	return nil
}
