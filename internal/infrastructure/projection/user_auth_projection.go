package projection

import (
	"context"
	"fmt"
	"time"

	"whisko-petcare/internal/domain/aggregate"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserAuthReadModel represents the read model for user authentication
type UserAuthReadModel struct {
	UserID         string     `bson:"_id"`
	Email          string     `bson:"email"`
	HashedPassword string     `bson:"hashed_password"`
	Role           string     `bson:"role"`
	CreatedAt      time.Time  `bson:"created_at"`
	UpdatedAt      time.Time  `bson:"updated_at"`
	LastLoginAt    *time.Time `bson:"last_login_at,omitempty"`
	IsActive       bool       `bson:"is_active"`
}

// MongoUserAuthRepository implements UserAuthRepository using MongoDB
type MongoUserAuthRepository struct {
	collection *mongo.Collection
}

// NewMongoUserAuthRepository creates a new MongoDB user auth repository
func NewMongoUserAuthRepository(db *mongo.Database) *MongoUserAuthRepository {
	return &MongoUserAuthRepository{
		collection: db.Collection("users"), // Use same collection as user projection
	}
}

// Save saves or updates user authentication
func (r *MongoUserAuthRepository) Save(ctx context.Context, user *aggregate.User) error {
	filter := bson.M{"_id": user.ID()}

	update := bson.M{
		"$set": bson.M{
			"_id":             user.ID(),
			"email":           user.Email(),
			"hashed_password": user.HashedPassword(),
			"role":            string(user.Role()),
			"updated_at":      user.UpdatedAt(),
			"last_login_at":   user.LastLoginAt(),
			"is_active":       user.IsActive(),
		},
		"$setOnInsert": bson.M{
			"created_at": user.CreatedAt(),
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save user auth: %w", err)
	}

	return nil
}

// FindByEmail finds user authentication by email
func (r *MongoUserAuthRepository) FindByEmail(ctx context.Context, email string) (*aggregate.User, error) {
	var model UserAuthReadModel
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	// Reconstruct aggregate from read model
	// Note: Since the password is already hashed, we return the read model data
	// The aggregate is mainly used for business logic, not for read operations
	return nil, nil // For read operations, use GetByEmail instead
}

// FindByUserID finds user authentication by user ID
func (r *MongoUserAuthRepository) FindByUserID(ctx context.Context, userID string) (*aggregate.User, error) {
	var model UserAuthReadModel
	err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}

	return nil, nil // Implement proper reconstruction
}

// Delete deletes user authentication
func (r *MongoUserAuthRepository) Delete(ctx context.Context, userID string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return fmt.Errorf("failed to delete user auth: %w", err)
	}
	return nil
}

// ExistsByEmail checks if email already exists
func (r *MongoUserAuthRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return count > 0, nil
}

// GetByEmail returns the read model by email (helper method)
func (r *MongoUserAuthRepository) GetByEmail(ctx context.Context, email string) (*UserAuthReadModel, error) {
	var model UserAuthReadModel
	err := r.collection.FindOne(ctx, bson.M{"email": email, "is_active": true}).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	return &model, nil
}
