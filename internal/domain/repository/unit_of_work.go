package repository

import (
	"context"
)

// UnitOfWork represents a unit of work pattern that manages repositories and transactions
type UnitOfWork interface {
	// Transaction management
	Begin(ctx context.Context) error
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error

	// Repository factory methods
	UserRepository() UserRepository
	PaymentRepository() PaymentRepository
	PetRepository() PetRepository
	VendorRepository() VendorRepository
	ServiceRepository() ServiceRepository
	ScheduleRepository() ScheduleRepository
	VendorStaffRepository() VendorStaffRepository
	PayoutRepository() PayoutRepository

	// Generic repository factory
	Repository(entityType string) interface{}

	// Bulk operations
	SaveChanges(ctx context.Context) error

	// Resource management
	Close() error

	// Transaction state
	IsInTransaction() bool
}

// UnitOfWorkFactory creates new unit of work instances
type UnitOfWorkFactory interface {
	CreateUnitOfWork() UnitOfWork
}

// TransactionalRepository extends repository with transaction support
type TransactionalRepository interface {
	// Set transaction context for the repository
	SetTransaction(tx interface{})

	// Get current transaction context
	GetTransaction() interface{}

	// Check if repository is in transaction
	IsTransactional() bool
}
