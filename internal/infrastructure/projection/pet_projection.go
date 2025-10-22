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

// PetReadModel represents the read model for pet queries
type PetReadModel struct {
	ID          string    `bson:"_id" json:"id"`
	UserID      string    `bson:"user_id" json:"user_id"`
	Name        string    `bson:"name" json:"name"`
	Species     string    `bson:"species" json:"species"`
	Breed       string    `bson:"breed" json:"breed"`
	Age         int       `bson:"age" json:"age"`
	Weight      float64   `bson:"weight" json:"weight"`
	IsActive    bool      `bson:"is_active" json:"is_active"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at" json:"updated_at"`
}

// PetProjection handles pet read model operations
type PetProjection interface {
	GetByID(ctx context.Context, id string) (*PetReadModel, error)
	GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*PetReadModel, error)
	ListAll(ctx context.Context, offset, limit int) ([]*PetReadModel, error)
	HandlePetCreated(ctx context.Context, event *event.PetCreated) error
	HandlePetUpdated(ctx context.Context, event *event.PetUpdated) error
	HandlePetDeleted(ctx context.Context, event *event.PetDeleted) error
}

// MongoPetProjection implements PetProjection using MongoDB
type MongoPetProjection struct {
	collection *mongo.Collection
}

// NewMongoPetProjection creates a new MongoDB pet projection
func NewMongoPetProjection(db *mongo.Database) *MongoPetProjection {
	collection := db.Collection("pets")
	
	// Create indexes
	ctx := context.Background()
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "species", Value: 1}},
		},
	}
	
	_, err := collection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		fmt.Printf("Warning: failed to create pet indexes: %v\n", err)
	}
	
	return &MongoPetProjection{
		collection: collection,
	}
}

// GetByID retrieves a pet by ID
func (p *MongoPetProjection) GetByID(ctx context.Context, id string) (*PetReadModel, error) {
	var pet PetReadModel
	err := p.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&pet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("pet not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get pet: %w", err)
	}
	return &pet, nil
}

// GetByUserID retrieves all pets for a user with pagination
func (p *MongoPetProjection) GetByUserID(ctx context.Context, userID string, offset, limit int) ([]*PetReadModel, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	filter := bson.M{
		"user_id":   userID,
		"is_active": true,
	}
	
	cursor, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find pets: %w", err)
	}
	defer cursor.Close(ctx)
	
	var pets []*PetReadModel
	if err := cursor.All(ctx, &pets); err != nil {
		return nil, fmt.Errorf("failed to decode pets: %w", err)
	}
	
	return pets, nil
}

// ListAll retrieves all pets with pagination
func (p *MongoPetProjection) ListAll(ctx context.Context, offset, limit int) ([]*PetReadModel, error) {
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	filter := bson.M{"is_active": true}
	
	cursor, err := p.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find pets: %w", err)
	}
	defer cursor.Close(ctx)
	
	var pets []*PetReadModel
	if err := cursor.All(ctx, &pets); err != nil {
		return nil, fmt.Errorf("failed to decode pets: %w", err)
	}
	
	return pets, nil
}

// HandlePetCreated handles the PetCreated event
func (p *MongoPetProjection) HandlePetCreated(ctx context.Context, event *event.PetCreated) error {
	pet := &PetReadModel{
		ID:        event.PetID,
		UserID:    event.UserID,
		Name:      event.Name,
		Species:   event.Species,
		Breed:     event.Breed,
		Age:       event.Age,
		Weight:    event.Weight,
		IsActive:  true,
		CreatedAt: event.Timestamp,
		UpdatedAt: event.Timestamp,
	}
	
	_, err := p.collection.InsertOne(ctx, pet)
	if err != nil {
		return fmt.Errorf("failed to insert pet: %w", err)
	}
	
	return nil
}

// HandlePetUpdated handles the PetUpdated event
func (p *MongoPetProjection) HandlePetUpdated(ctx context.Context, event *event.PetUpdated) error {
	filter := bson.M{"_id": event.PetID}
	update := bson.M{
		"$set": bson.M{
			"name":       event.Name,
			"species":    event.Species,
			"breed":      event.Breed,
			"age":        event.Age,
			"weight":     event.Weight,
			"updated_at": event.Timestamp,
		},
	}
	
	result, err := p.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update pet: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return fmt.Errorf("pet not found: %s", event.PetID)
	}
	
	return nil
}

// HandlePetDeleted handles the PetDeleted event
func (p *MongoPetProjection) HandlePetDeleted(ctx context.Context, event *event.PetDeleted) error {
	filter := bson.M{"_id": event.PetID}
	update := bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": event.Timestamp,
		},
	}
	
	result, err := p.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to delete pet: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return fmt.Errorf("pet not found: %s", event.PetID)
	}
	
	return nil
}
