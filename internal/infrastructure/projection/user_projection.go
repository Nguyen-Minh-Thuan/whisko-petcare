package projection

import (
	"context"
	"fmt"
	"strings"
	"time"

	"whisko-petcare/internal/domain/event"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserReadModel represents the read model for users
type UserReadModel struct {
	ID             string     `json:"id" bson:"_id"`
	Name           string     `json:"name" bson:"name"`
	Email          string     `json:"email" bson:"email"`
	Phone          string     `json:"phone" bson:"phone"`
	Address        string     `json:"address" bson:"address"`
	HashedPassword string     `json:"-" bson:"hashed_password,omitempty"` // Hidden from JSON
	Role           string     `json:"role,omitempty" bson:"role,omitempty"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty" bson:"last_login_at,omitempty"`
	IsActive       bool       `json:"is_active,omitempty" bson:"is_active,omitempty"`
	Version        int        `json:"version" bson:"version"`
	CreatedAt      time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" bson:"updated_at"`
	IsDeleted      bool       `json:"is_deleted" bson:"is_deleted"`
}

// UserProjection defines operations for user read model
type UserProjection interface {
	GetByID(ctx context.Context, id string) (*UserReadModel, error)
	List(ctx context.Context, limit, offset int) ([]*UserReadModel, error)
	Search(ctx context.Context, name, email string) ([]*UserReadModel, error)

	// Event handlers
	HandleUserCreated(ctx context.Context, event *event.UserCreated) error
	HandleUserProfileUpdated(ctx context.Context, event *event.UserProfileUpdated) error
	HandleUserContactUpdated(ctx context.Context, event *event.UserContactUpdated) error
	HandleUserDeleted(ctx context.Context, event *event.UserDeleted) error
}

// MongoUserProjection implements UserProjection using MongoDB
type MongoUserProjection struct {
	collection *mongo.Collection
}

func NewMongoUserProjection(database *mongo.Database) UserProjection {
	return &MongoUserProjection{
		collection: database.Collection("users"), // Read from the same collection where data is written
	}
}

func (p *MongoUserProjection) GetByID(ctx context.Context, id string) (*UserReadModel, error) {
	var result bson.M
	err := p.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result) // Removed is_deleted filter
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	user := &UserReadModel{
		ID:        result["_id"].(string),
		Name:      result["name"].(string),
		Email:     result["email"].(string),
		Phone:     getStringFromResult(result, "phone"),
		Address:   getStringFromResult(result, "address"),
		Version:   getIntFromResult(result, "version"),
		CreatedAt: getTimeFromResult(result, "created_at"),
		UpdatedAt: getTimeFromResult(result, "updated_at"),
		IsDeleted: false, // Default to false since users collection doesn't track this
	}

	return user, nil
}

func (p *MongoUserProjection) List(ctx context.Context, limit, offset int) ([]*UserReadModel, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // Sort by newest first

	cursor, err := p.collection.Find(ctx, bson.M{}, opts) // Removed is_deleted filter
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*UserReadModel
	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			return nil, fmt.Errorf("failed to decode user: %w", err)
		}

		user := &UserReadModel{
			ID:        result["_id"].(string),
			Name:      result["name"].(string),
			Email:     result["email"].(string),
			Phone:     getStringFromResult(result, "phone"),
			Address:   getStringFromResult(result, "address"),
			Version:   getIntFromResult(result, "version"),
			CreatedAt: getTimeFromResult(result, "created_at"),
			UpdatedAt: getTimeFromResult(result, "updated_at"),
			IsDeleted: false,
		}
		users = append(users, user)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return users, nil
}

func (p *MongoUserProjection) Search(ctx context.Context, name, email string) ([]*UserReadModel, error) {
	filter := bson.M{"is_deleted": false}

	if name != "" {
		filter["name"] = bson.M{"$regex": name, "$options": "i"} // Case insensitive
	}
	if email != "" {
		filter["email"] = bson.M{"$regex": "^" + strings.ToLower(email) + "$", "$options": "i"}
	}

	cursor, err := p.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	defer cursor.Close(ctx)

	var users []*UserReadModel
	for cursor.Next(ctx) {
		var user UserReadModel
		if err := cursor.Decode(&user); err != nil {
			return nil, fmt.Errorf("failed to decode user: %w", err)
		}
		// Set ID from _id field
		var result bson.M
		cursor.Decode(&result)
		user.ID = result["_id"].(string)
		users = append(users, &user)
	}

	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("cursor error: %w", err)
	}

	return users, nil
}

// Event handlers
func (p *MongoUserProjection) HandleUserCreated(ctx context.Context, event *event.UserCreated) error {
	userReadModel := bson.M{
		"_id":        event.UserID,
		"name":       event.Name,
		"email":      event.Email,
		"phone":      event.Phone,
		"address":    event.Address,
		"version":    1,
		"created_at": event.Timestamp,
		"updated_at": event.Timestamp,
		"is_deleted": false,
	}

	_, err := p.collection.InsertOne(ctx, userReadModel)
	if err != nil {
		return fmt.Errorf("failed to create user projection: %w", err)
	}

	return nil
}

func (p *MongoUserProjection) HandleUserProfileUpdated(ctx context.Context, event *event.UserProfileUpdated) error {
	filter := bson.M{"_id": event.UserID}
	update := bson.M{
		"$set": bson.M{
			"name":       event.Name,
			"email":      event.Email,
			"version":    event.EventVersion,
			"updated_at": event.Timestamp,
		},
	}

	result, err := p.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user profile projection: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found for profile update")
	}

	return nil
}

func (p *MongoUserProjection) HandleUserContactUpdated(ctx context.Context, event *event.UserContactUpdated) error {
	filter := bson.M{"_id": event.UserID}
	update := bson.M{
		"$set": bson.M{
			"phone":      event.Phone,
			"address":    event.Address,
			"version":    event.EventVersion,
			"updated_at": event.Timestamp,
		},
	}

	result, err := p.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update user contact projection: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found for contact update")
	}

	return nil
}

func (p *MongoUserProjection) HandleUserDeleted(ctx context.Context, event *event.UserDeleted) error {
	filter := bson.M{"_id": event.UserID}
	update := bson.M{
		"$set": bson.M{
			"is_deleted": true,
			"updated_at": event.Timestamp,
		},
	}

	result, err := p.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to mark user as deleted in projection: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found for deletion")
	}

	return nil
}

// Helper functions for safe type conversion
func getStringFromResult(m bson.M, key string) string {
	if val, ok := m[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getIntFromResult(m bson.M, key string) int {
	if val, ok := m[key]; ok && val != nil {
		if i, ok := val.(int32); ok {
			return int(i)
		}
		if i, ok := val.(int64); ok {
			return int(i)
		}
		if i, ok := val.(int); ok {
			return i
		}
	}
	return 0
}

func getTimeFromResult(m bson.M, key string) time.Time {
	if val, ok := m[key]; ok && val != nil {
		if t, ok := val.(time.Time); ok {
			return t
		}
		// Handle Unix timestamp (int64)
		if ts, ok := val.(int64); ok {
			return time.Unix(ts/1000, (ts%1000)*1000000) // Convert milliseconds to nanoseconds
		}
	}
	return time.Time{}
}
