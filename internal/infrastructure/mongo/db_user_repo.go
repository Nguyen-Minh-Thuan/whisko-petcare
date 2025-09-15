package mongo

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"whisko-petcare/internal/infrastructure/projection"
)

// MongoProjectionRepository implements read model storage for MongoDB
type MongoProjectionRepository struct {
	// In production, use real MongoDB connection
	// collection *mongo.Collection

	// For now, in-memory storage
	users map[string]*projection.UserReadModel
	mutex sync.RWMutex
}

func NewMongoProjectionRepository() *MongoProjectionRepository {
	return &MongoProjectionRepository{
		users: make(map[string]*projection.UserReadModel),
	}
}

func (r *MongoProjectionRepository) Save(ctx context.Context, user *projection.UserReadModel) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.users[user.ID] = user
	return nil
}

func (r *MongoProjectionRepository) GetByID(ctx context.Context, id string) (*projection.UserReadModel, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	user, exists := r.users[id]
	if !exists || user.IsDeleted {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (r *MongoProjectionRepository) List(ctx context.Context, limit, offset int) ([]*projection.UserReadModel, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var users []*projection.UserReadModel
	count := 0

	for _, user := range r.users {
		if user.IsDeleted {
			continue
		}

		if count >= offset {
			users = append(users, user)
		}
		count++

		if len(users) >= limit {
			break
		}
	}

	return users, nil
}

func (r *MongoProjectionRepository) Search(ctx context.Context, name, email string) ([]*projection.UserReadModel, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var users []*projection.UserReadModel

	for _, user := range r.users {
		if user.IsDeleted {
			continue
		}

		matchesName := name == "" || strings.Contains(strings.ToLower(user.Name), strings.ToLower(name))
		matchesEmail := email == "" || strings.EqualFold(user.Email, email)

		if matchesName && matchesEmail {
			users = append(users, user)
		}
	}

	return users, nil
}
