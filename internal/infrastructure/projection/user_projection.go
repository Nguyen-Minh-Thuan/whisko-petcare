package projection

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"whisko-petcare/internal/domain/event"
)

// UserReadModel represents the read model for users
type UserReadModel struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Address   string    `json:"address"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsDeleted bool      `json:"is_deleted"`
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

// InMemoryUserProjection implements UserProjection in memory
type InMemoryUserProjection struct {
	users map[string]*UserReadModel
	mutex sync.RWMutex
}

func NewInMemoryUserProjection() UserProjection {
	return &InMemoryUserProjection{
		users: make(map[string]*UserReadModel),
	}
}

func (p *InMemoryUserProjection) GetByID(ctx context.Context, id string) (*UserReadModel, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	user, exists := p.users[id]
	if !exists || user.IsDeleted {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (p *InMemoryUserProjection) List(ctx context.Context, limit, offset int) ([]*UserReadModel, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var users []*UserReadModel
	count := 0

	for _, user := range p.users {
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

func (p *InMemoryUserProjection) Search(ctx context.Context, name, email string) ([]*UserReadModel, error) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var users []*UserReadModel

	for _, user := range p.users {
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

// Event handlers
func (p *InMemoryUserProjection) HandleUserCreated(ctx context.Context, event *event.UserCreated) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.users[event.UserID] = &UserReadModel{
		ID:        event.UserID,
		Name:      event.Name,
		Email:     event.Email,
		Phone:     event.Phone,
		Address:   event.Address,
		Version:   1,
		CreatedAt: event.Timestamp,
		UpdatedAt: event.Timestamp,
		IsDeleted: false,
	}

	return nil
}

func (p *InMemoryUserProjection) HandleUserProfileUpdated(ctx context.Context, event *event.UserProfileUpdated) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	user, exists := p.users[event.UserID]
	if !exists {
		return fmt.Errorf("user not found for update")
	}

	user.Name = event.Name
	user.Email = event.Email
	user.Version = event.EventVersion
	user.UpdatedAt = event.Timestamp

	return nil
}

func (p *InMemoryUserProjection) HandleUserContactUpdated(ctx context.Context, event *event.UserContactUpdated) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	user, exists := p.users[event.UserID]
	if !exists {
		return fmt.Errorf("user not found for update")
	}

	user.Phone = event.Phone
	user.Address = event.Address
	user.Version = event.EventVersion
	user.UpdatedAt = event.Timestamp

	return nil
}

func (p *InMemoryUserProjection) HandleUserDeleted(ctx context.Context, event *event.UserDeleted) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	user, exists := p.users[event.UserID]
	if !exists {
		return fmt.Errorf("user not found for deletion")
	}

	user.IsDeleted = true
	user.UpdatedAt = event.Timestamp

	return nil
}
