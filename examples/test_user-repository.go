package examples

import (
	"context"
	"fmt"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/infrastructure/mongo"
	"whisko-petcare/internal/infrastructure/projection"
)

// aggregateToReadModel converts an aggregate.User to a projection.UserReadModel
func aggregateToReadModel(u *aggregate.User) *projection.UserReadModel {
	return &projection.UserReadModel{
		ID:        u.ID(),
		Name:      u.Name(),
		Email:     u.Email(),
		Phone:     u.Phone(),
		Address:   u.Address(),
		Version:   u.Version(),
		CreatedAt: u.CreatedAt(),
		UpdatedAt: u.UpdatedAt(),
		IsDeleted: false, // or add a method to aggregate.User if you track deletion
	}
}

// TestUserRepository demonstrates the Generic Repository functionality
func TestUserRepository() {
	ctx := context.Background()

	// Create a user repository
	userRepo := mongo.NewMongoProjectionRepository()

	// Create some users
	user1, _ := aggregate.NewUser("1", "Zhu Yuan", "police@example.com")
	user2, _ := aggregate.NewUser("2", "Elysia", "yurijesus@example.com")
	user3, _ := aggregate.NewUser("3", "Keanu Reeves", "thegoat@example.com")

	// Convert to read models and save
	userRepo.Save(ctx, aggregateToReadModel(user1))
	userRepo.Save(ctx, aggregateToReadModel(user2))
	userRepo.Save(ctx, aggregateToReadModel(user3))

	fmt.Printf("Saved 3 users\n")
	fmt.Printf("Repository test completed successfully!\n")
}

func main() {
	TestUserRepository()
}
