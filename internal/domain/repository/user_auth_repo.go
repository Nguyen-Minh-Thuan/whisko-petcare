package repository

import (
	"context"
	"whisko-petcare/internal/domain/aggregate"
)

// UserAuthRepository defines operations for user authentication storage
type UserAuthRepository interface {
	// Save saves or updates user authentication
	Save(ctx context.Context, user *aggregate.User) error

	// FindByEmail finds user authentication by email
	FindByEmail(ctx context.Context, email string) (*aggregate.User, error)

	// FindByUserID finds user authentication by user ID
	FindByUserID(ctx context.Context, userID string) (*aggregate.User, error)

	// Delete deletes user authentication
	Delete(ctx context.Context, userID string) error

	// ExistsByEmail checks if email already exists
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}
