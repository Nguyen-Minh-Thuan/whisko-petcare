package mongo

import (
	"context"
	"fmt"
	"sync"

	"whisko-petcare/internal/domain/repository"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoUnitOfWork implements the Unit of Work pattern for MongoDB
type MongoUnitOfWork struct {
	client        *mongo.Client
	database      *mongo.Database
	session       mongo.Session
	repositories  map[string]interface{}
	mutex         sync.RWMutex
	inTransaction bool

	// Repository instances
	userRepo         repository.UserRepository
	paymentRepo      repository.PaymentRepository
	petRepo          repository.PetRepository
	vendorRepo       repository.VendorRepository
	serviceRepo      repository.ServiceRepository
	scheduleRepo     repository.ScheduleRepository
	vendorStaffRepo  repository.VendorStaffRepository
	// payoutRepo       repository.PayoutRepository  // TODO: Implement MongoDB payout repository
}

// NewMongoUnitOfWork creates a new MongoDB unit of work
func NewMongoUnitOfWork(client *mongo.Client, database *mongo.Database) *MongoUnitOfWork {
	return &MongoUnitOfWork{
		client:       client,
		database:     database,
		repositories: make(map[string]interface{}),
	}
}

// Begin starts a new transaction
func (uow *MongoUnitOfWork) Begin(ctx context.Context) error {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if uow.inTransaction {
		return fmt.Errorf("unit of work is already in transaction")
	}

	session, err := uow.client.StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}

	err = session.StartTransaction()
	if err != nil {
		session.EndSession(ctx)
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	uow.session = session
	uow.inTransaction = true

	// Set transaction context for all repositories
	uow.setTransactionForRepositories()

	return nil
}

// Commit commits the current transaction
func (uow *MongoUnitOfWork) Commit(ctx context.Context) error {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if !uow.inTransaction {
		return fmt.Errorf("no active transaction to commit")
	}

	err := uow.session.CommitTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	uow.endTransaction(ctx)
	return nil
}

// Rollback rolls back the current transaction
func (uow *MongoUnitOfWork) Rollback(ctx context.Context) error {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if !uow.inTransaction {
		return fmt.Errorf("no active transaction to rollback")
	}

	err := uow.session.AbortTransaction(ctx)
	if err != nil {
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	uow.endTransaction(ctx)
	return nil
}

// UserRepository returns the user repository
func (uow *MongoUnitOfWork) UserRepository() repository.UserRepository {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if uow.userRepo == nil {
		uow.userRepo = NewMongoUserRepository(uow.database)
		if uow.inTransaction {
			if transactionalRepo, ok := uow.userRepo.(repository.TransactionalRepository); ok {
				transactionalRepo.SetTransaction(uow.session)
			}
		}
	}

	return uow.userRepo
}

// PaymentRepository returns the payment repository
func (uow *MongoUnitOfWork) PaymentRepository() repository.PaymentRepository {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if uow.paymentRepo == nil {
		uow.paymentRepo = NewMongoPaymentRepository(uow.database)
		if uow.inTransaction {
			if transactionalRepo, ok := uow.paymentRepo.(repository.TransactionalRepository); ok {
				transactionalRepo.SetTransaction(uow.session)
			}
		}
	}

	return uow.paymentRepo
}

// PetRepository returns the pet repository
func (uow *MongoUnitOfWork) PetRepository() repository.PetRepository {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if uow.petRepo == nil {
		uow.petRepo = NewMongoPetRepository(uow.database)
		if uow.inTransaction {
			if transactionalRepo, ok := uow.petRepo.(repository.TransactionalRepository); ok {
				transactionalRepo.SetTransaction(uow.session)
			}
		}
	}

	return uow.petRepo
}

// VendorRepository returns the vendor repository
func (uow *MongoUnitOfWork) VendorRepository() repository.VendorRepository {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if uow.vendorRepo == nil {
		uow.vendorRepo = NewMongoVendorRepository(uow.database)
		if uow.inTransaction {
			if transactionalRepo, ok := uow.vendorRepo.(repository.TransactionalRepository); ok {
				transactionalRepo.SetTransaction(uow.session)
			}
		}
	}

	return uow.vendorRepo
}

// ServiceRepository returns the service repository
func (uow *MongoUnitOfWork) ServiceRepository() repository.ServiceRepository {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if uow.serviceRepo == nil {
		uow.serviceRepo = NewMongoServiceRepository(uow.database)
		if uow.inTransaction {
			if transactionalRepo, ok := uow.serviceRepo.(repository.TransactionalRepository); ok {
				transactionalRepo.SetTransaction(uow.session)
			}
		}
	}

	return uow.serviceRepo
}

// ScheduleRepository returns the schedule repository
func (uow *MongoUnitOfWork) ScheduleRepository() repository.ScheduleRepository {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if uow.scheduleRepo == nil {
		uow.scheduleRepo = NewMongoScheduleRepository(uow.database)
		if uow.inTransaction {
			if transactionalRepo, ok := uow.scheduleRepo.(repository.TransactionalRepository); ok {
				transactionalRepo.SetTransaction(uow.session)
			}
		}
	}

	return uow.scheduleRepo
}

// VendorStaffRepository returns the vendor staff repository
func (uow *MongoUnitOfWork) VendorStaffRepository() repository.VendorStaffRepository {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if uow.vendorStaffRepo == nil {
		uow.vendorStaffRepo = NewMongoVendorStaffRepository(uow.database)
		if uow.inTransaction {
			if transactionalRepo, ok := uow.vendorStaffRepo.(repository.TransactionalRepository); ok {
				transactionalRepo.SetTransaction(uow.session)
			}
		}
	}

	return uow.vendorStaffRepo
}

// PayoutRepository returns the payout repository
// TODO: Implement MongoDB payout repository
func (uow *MongoUnitOfWork) PayoutRepository() repository.PayoutRepository {
	panic("PayoutRepository not yet implemented - MongoDB payout repository needs to be created")
}

// Repository returns a generic repository for the specified entity type
func (uow *MongoUnitOfWork) Repository(entityType string) interface{} {
	uow.mutex.RLock()
	defer uow.mutex.RUnlock()

	if repo, exists := uow.repositories[entityType]; exists {
		return repo
	}

	// Create repository based on entity type
	// This can be extended to support more entity types
	switch entityType {
	case "user":
		return uow.UserRepository()
	case "payment":
		return uow.PaymentRepository()
	default:
		return nil
	}
}

// SaveChanges persists all changes in the current unit of work
func (uow *MongoUnitOfWork) SaveChanges(ctx context.Context) error {
	// In MongoDB with transactions, changes are automatically persisted on commit
	// This method can be used for additional validation or business logic
	return nil
}

// Close closes the unit of work and cleans up resources
func (uow *MongoUnitOfWork) Close() error {
	uow.mutex.Lock()
	defer uow.mutex.Unlock()

	if uow.inTransaction && uow.session != nil {
		ctx := context.Background()
		uow.session.AbortTransaction(ctx)
		uow.endTransaction(ctx)
	}

	return nil
}

// IsInTransaction returns whether the unit of work is in a transaction
func (uow *MongoUnitOfWork) IsInTransaction() bool {
	uow.mutex.RLock()
	defer uow.mutex.RUnlock()
	return uow.inTransaction
}

// endTransaction cleans up transaction resources
func (uow *MongoUnitOfWork) endTransaction(ctx context.Context) {
	if uow.session != nil {
		uow.session.EndSession(ctx)
		uow.session = nil
	}
	uow.inTransaction = false

	// Clear transaction context from repositories
	uow.clearTransactionFromRepositories()
}

// setTransactionForRepositories sets transaction context for all repositories
func (uow *MongoUnitOfWork) setTransactionForRepositories() {
	if uow.userRepo != nil {
		if transactionalRepo, ok := uow.userRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(uow.session)
		}
	}

	if uow.paymentRepo != nil {
		if transactionalRepo, ok := uow.paymentRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(uow.session)
		}
	}

	if uow.petRepo != nil {
		if transactionalRepo, ok := uow.petRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(uow.session)
		}
	}

	if uow.vendorRepo != nil {
		if transactionalRepo, ok := uow.vendorRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(uow.session)
		}
	}

	if uow.serviceRepo != nil {
		if transactionalRepo, ok := uow.serviceRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(uow.session)
		}
	}

	if uow.scheduleRepo != nil {
		if transactionalRepo, ok := uow.scheduleRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(uow.session)
		}
	}

	if uow.vendorStaffRepo != nil {
		if transactionalRepo, ok := uow.vendorStaffRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(uow.session)
		}
	}

	// Set transaction for other repositories in the map
	for _, repo := range uow.repositories {
		if transactionalRepo, ok := repo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(uow.session)
		}
	}
}

// clearTransactionFromRepositories clears transaction context from all repositories
func (uow *MongoUnitOfWork) clearTransactionFromRepositories() {
	if uow.userRepo != nil {
		if transactionalRepo, ok := uow.userRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(nil)
		}
	}

	if uow.paymentRepo != nil {
		if transactionalRepo, ok := uow.paymentRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(nil)
		}
	}

	if uow.petRepo != nil {
		if transactionalRepo, ok := uow.petRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(nil)
		}
	}

	if uow.vendorRepo != nil {
		if transactionalRepo, ok := uow.vendorRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(nil)
		}
	}

	if uow.serviceRepo != nil {
		if transactionalRepo, ok := uow.serviceRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(nil)
		}
	}

	if uow.scheduleRepo != nil {
		if transactionalRepo, ok := uow.scheduleRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(nil)
		}
	}

	if uow.vendorStaffRepo != nil {
		if transactionalRepo, ok := uow.vendorStaffRepo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(nil)
		}
	}

	// Clear transaction for other repositories in the map
	for _, repo := range uow.repositories {
		if transactionalRepo, ok := repo.(repository.TransactionalRepository); ok {
			transactionalRepo.SetTransaction(nil)
		}
	}
}

// MongoUnitOfWorkFactory creates MongoDB unit of work instances
type MongoUnitOfWorkFactory struct {
	client   *mongo.Client
	database *mongo.Database
}

// NewMongoUnitOfWorkFactory creates a new MongoDB unit of work factory
func NewMongoUnitOfWorkFactory(client *mongo.Client, database *mongo.Database) *MongoUnitOfWorkFactory {
	return &MongoUnitOfWorkFactory{
		client:   client,
		database: database,
	}
}

// CreateUnitOfWork creates a new unit of work instance
func (f *MongoUnitOfWorkFactory) CreateUnitOfWork() repository.UnitOfWork {
	return NewMongoUnitOfWork(f.client, f.database)
}
