package mongo

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrNotFound = errors.New("entity not found")

// Entity interface that all entities must implement
type Entity interface {
	GetID() string
	SetID(id string)
	GetVersion() int
	SetVersion(version int)
}

// GenericRepository provides CRUD operations for any entity type
type GenericRepository[T Entity] struct {
	data  map[string]T
	mutex sync.RWMutex
}

// NewGenericRepository creates a new generic repository
func NewGenericRepository[T Entity]() *GenericRepository[T] {
	return &GenericRepository[T]{
		data: make(map[string]T),
	}
}

// Save stores an entity
func (r *GenericRepository[T]) Save(ctx context.Context, entity T) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if entity.GetID() == "" {
		// Generate ID if not set
		entity.SetID(generateID())
	}

	r.data[entity.GetID()] = entity
	return nil
}

// GetByID retrieves an entity by ID
func (r *GenericRepository[T]) GetByID(ctx context.Context, id string) (T, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var zero T
	entity, exists := r.data[id]
	if !exists {
		return zero, ErrNotFound
	}

	return entity, nil
}

// GetAll retrieves all entities
func (r *GenericRepository[T]) GetAll(ctx context.Context) ([]T, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	entities := make([]T, 0, len(r.data))
	for _, entity := range r.data {
		entities = append(entities, entity)
	}

	return entities, nil
}

// Delete removes an entity by ID
func (r *GenericRepository[T]) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.data[id]; !exists {
		return ErrNotFound
	}

	delete(r.data, id)
	return nil
}

// FindBy finds entities matching a predicate
func (r *GenericRepository[T]) FindBy(ctx context.Context, predicate func(T) bool) ([]T, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	var result []T
	for _, entity := range r.data {
		if predicate(entity) {
			result = append(result, entity)
		}
	}

	return result, nil
}

// Count returns the number of entities
func (r *GenericRepository[T]) Count(ctx context.Context) int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.data)
}

// Exists checks if an entity exists
func (r *GenericRepository[T]) Exists(ctx context.Context, id string) bool {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	_, exists := r.data[id]
	return exists
}

// Clear removes all entities (for testing)
func (r *GenericRepository[T]) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.data = make(map[string]T)
}

func generateID() string {
	return "id_" + time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}
