package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/internal/infrastructure/projection"
)

// TestEventHandler demonstrates event handling
type TestEventHandler struct {
	name string
}

func (h *TestEventHandler) HandleEvent(ctx context.Context, evt event.DomainEvent) error {
	fmt.Printf("🎯 [%s] Handling event: %s for aggregate %s\n",
		h.name, evt.EventType(), evt.AggregateID())

	switch e := evt.(type) {
	case *event.UserCreated:
		fmt.Printf("   📝 User Created: %s (%s)\n", e.Name, e.Email)
	case *event.UserProfileUpdated:
		fmt.Printf("   ✏️  Profile Updated: %s (%s)\n", e.Name, e.Email)
	case *event.UserContactUpdated:
		fmt.Printf("   📱 Contact Updated: %s, %s\n", e.Phone, e.Address)
	case *event.UserDeleted:
		fmt.Printf("   🗑️  User Deleted: %s\n", e.UserID)
	}

	return nil
}

// TestUserRepositoryWithUoW demonstrates the Unit of Work and Event handling
func TestUserRepositoryWithUoW() {
	ctx := context.Background()

	fmt.Println("🚀 Starting Unit of Work and Event Handling Test...")

	// Create mock MongoDB setup (in-memory for testing)
	fmt.Println("📦 Setting up test infrastructure...")

	// Create event bus
	eventBus := bus.NewInMemoryEventBus()

	// Create test event handlers
	handler1 := &TestEventHandler{name: "Projection Handler"}
	handler2 := &TestEventHandler{name: "Audit Handler"}
	handler3 := &TestEventHandler{name: "Notification Handler"}

	// Subscribe handlers to events
	eventBus.Subscribe("UserCreated", bus.EventHandlerFunc(handler1.HandleEvent))
	eventBus.Subscribe("UserCreated", bus.EventHandlerFunc(handler2.HandleEvent))
	eventBus.Subscribe("UserProfileUpdated", bus.EventHandlerFunc(handler1.HandleEvent))
	eventBus.Subscribe("UserContactUpdated", bus.EventHandlerFunc(handler1.HandleEvent))
	eventBus.Subscribe("UserDeleted", bus.EventHandlerFunc(handler3.HandleEvent))

	// Start event bus
	if err := eventBus.Start(ctx); err != nil {
		log.Fatal("Failed to start event bus:", err)
	}
	defer eventBus.Stop()

	// Create projection for read models
	userProjection := projection.NewInMemoryUserProjection()

	// Subscribe projection to events
	eventBus.Subscribe("UserCreated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserCreated(ctx, e.(*event.UserCreated))
		}))

	eventBus.Subscribe("UserProfileUpdated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserProfileUpdated(ctx, e.(*event.UserProfileUpdated))
		}))

	eventBus.Subscribe("UserContactUpdated", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserContactUpdated(ctx, e.(*event.UserContactUpdated))
		}))

	eventBus.Subscribe("UserDeleted", bus.EventHandlerFunc(
		func(ctx context.Context, e event.DomainEvent) error {
			return userProjection.HandleUserDeleted(ctx, e.(*event.UserDeleted))
		}))

	// Test 1: Create users and see events
	fmt.Println("\n🧪 Test 1: Creating users and triggering events...")

	users := []struct {
		id, name, email, phone, address string
	}{
		{"user-1", "Zhu Yuan", "police@newereidu.com", "+1-555-0101", "New Eridu Police Station"},
		{"user-2", "Elysia", "herrscher@elysianrealm.com", "+1-555-0102", "Elysian Realm"},
		{"user-3", "Keanu Reeves", "theone@matrix.com", "+1-555-0103", "Neo's Apartment"},
	}

	var createdUsers []*aggregate.User

	for _, userData := range users {
		fmt.Printf("\n👤 Creating user: %s...\n", userData.name)

		user, err := aggregate.NewUser(userData.id, userData.name, userData.email)
		if err != nil {
			fmt.Printf("❌ Error creating user: %v\n", err)
			continue
		}

		// Update contact info
		if userData.phone != "" || userData.address != "" {
			user.UpdateContactInfo(userData.phone, userData.address)
		}

		// Publish events
		events := user.GetUncommittedEvents()
		for _, evt := range events {
			if err := eventBus.Publish(ctx, evt); err != nil {
				fmt.Printf("❌ Error publishing event: %v\n", err)
			}
		}

		// Mark events as committed
		user.MarkEventsAsCommitted()
		createdUsers = append(createdUsers, user)

		// Small delay to see the flow
		time.Sleep(100 * time.Millisecond)
	}

	// Test 2: Update user profiles
	fmt.Println("\n🧪 Test 2: Updating user profiles...")

	if len(createdUsers) > 0 {
		user := createdUsers[0]
		fmt.Printf("\n✏️ Updating profile for %s...\n", user.Name())

		user.UpdateProfile("Zhu Yuan (Chief)", "chief.police@newereidu.com")

		events := user.GetUncommittedEvents()
		for _, evt := range events {
			if err := eventBus.Publish(ctx, evt); err != nil {
				fmt.Printf("❌ Error publishing event: %v\n", err)
			}
		}
		user.MarkEventsAsCommitted()

		time.Sleep(100 * time.Millisecond)
	}

	// Test 3: Delete a user
	fmt.Println("\n🧪 Test 3: Deleting a user...")

	if len(createdUsers) > 2 {
		user := createdUsers[2]
		fmt.Printf("\n🗑️ Deleting user %s...\n", user.Name())

		user.Delete()

		events := user.GetUncommittedEvents()
		for _, evt := range events {
			if err := eventBus.Publish(ctx, evt); err != nil {
				fmt.Printf("❌ Error publishing event: %v\n", err)
			}
		}
		user.MarkEventsAsCommitted()

		time.Sleep(100 * time.Millisecond)
	}

	// Test 4: Show final projection state
	fmt.Println("\n🧪 Test 4: Checking projection state...")

	allUsers, err := userProjection.List(ctx, 10, 0)
	if err != nil {
		fmt.Printf("❌ Error getting users from projection: %v\n", err)
	} else {
		fmt.Printf("\n📊 Final projection state (%d users):\n", len(allUsers))
		for _, user := range allUsers {
			status := "✅ Active"
			if user.IsDeleted {
				status = "❌ Deleted"
			}
			fmt.Printf("   👤 %s (%s) - %s - Version: %d - %s\n",
				user.Name, user.Email, user.Phone, user.Version, status)
		}
	}

	fmt.Println("\n🎉 Test completed successfully!")
	fmt.Println("📋 Events were published and handled by multiple handlers:")
	fmt.Println("   - Projection Handler: Updates read models")
	fmt.Println("   - Audit Handler: Logs events for compliance")
	fmt.Println("   - Notification Handler: Handles user lifecycle events")
}

func main() {
	TestUserRepositoryWithUoW()
}
