package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"whisko-petcare/internal/application/command"
	"whisko-petcare/internal/domain/aggregate"
	"whisko-petcare/internal/domain/event"
	"whisko-petcare/internal/domain/repository"
	"whisko-petcare/internal/infrastructure/bus"
	"whisko-petcare/internal/infrastructure/projection"
)

// ExampleAsyncEventHandling demonstrates async event handling patterns
func ExampleAsyncEventHandling() {
	ctx := context.Background()

	// 1. Initialize async event bus
	eventBus := bus.NewAsyncEventBus()
	defer eventBus.Stop()

	// Start error monitoring
	eventBus.Start(ctx)

	// 2. Setup projections
	// userProjection := projection.NewMongoUserProjection(db)
	// paymentProjection := projection.NewMongoPaymentProjection(db)

	// 3. Subscribe event handlers to event bus
	setupEventHandlers(eventBus /* , userProjection, paymentProjection */)

	// 4. Create command handlers with UoW
	// uowFactory := mongo.NewMongoUnitOfWorkFactory(client, db)
	// createUserHandler := command.NewCreateUserWithUoWHandler(uowFactory, eventBus)

	// 5. Execute command - events are published asynchronously
	// cmd := command.CreateUser{
	// 	UserID: "user_123",
	// 	Name:   "John Doe",
	// 	Email:  "john@example.com",
	// }
	// err := createUserHandler.Handle(ctx, cmd)
	// if err != nil {
	// 	log.Printf("Command failed: %v", err)
	// 	return
	// }

	// 6. Command returns immediately, events process in background
	log.Println("Command completed, events are being processed asynchronously")

	// 7. Wait for all async event handlers to complete (if needed)
	eventBus.Wait()

	log.Println("All events processed")
}

// setupEventHandlers subscribes all event handlers to the event bus
func setupEventHandlers(eventBus bus.EventBus /* projections... */) {
	// User events
	eventBus.Subscribe("UserCreated", bus.EventHandlerFunc(handleUserCreated))
	eventBus.Subscribe("UserProfileUpdated", bus.EventHandlerFunc(handleUserProfileUpdated))
	eventBus.Subscribe("UserContactUpdated", bus.EventHandlerFunc(handleUserContactUpdated))
	eventBus.Subscribe("UserDeleted", bus.EventHandlerFunc(handleUserDeleted))

	// Payment events
	eventBus.Subscribe("PaymentCreated", bus.EventHandlerFunc(handlePaymentCreated))
	eventBus.Subscribe("PaymentStatusChanged", bus.EventHandlerFunc(handlePaymentStatusChanged))

	// Pet events
	eventBus.Subscribe("PetCreated", bus.EventHandlerFunc(handlePetCreated))
	eventBus.Subscribe("PetUpdated", bus.EventHandlerFunc(handlePetUpdated))

	// Service events
	eventBus.Subscribe("ServiceCreated", bus.EventHandlerFunc(handleServiceCreated))
	eventBus.Subscribe("ServiceUpdated", bus.EventHandlerFunc(handleServiceUpdated))

	// Add more subscriptions as needed...
}

// Example event handlers
func handleUserCreated(ctx context.Context, evt event.DomainEvent) error {
	userCreated, ok := evt.(*event.UserCreated)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling UserCreated event: UserID=%s, Name=%s", userCreated.UserID, userCreated.Name)

	// Update projection (in real implementation)
	// projection.HandleUserCreated(ctx, userCreated)

	// Send welcome email (example side effect)
	// emailService.SendWelcomeEmail(userCreated.Email)

	return nil
}

func handleUserProfileUpdated(ctx context.Context, evt event.DomainEvent) error {
	updated, ok := evt.(*event.UserProfileUpdated)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling UserProfileUpdated event: UserID=%s", updated.AggregateID())

	// Update projection
	// projection.HandleUserProfileUpdated(ctx, updated)

	return nil
}

func handleUserContactUpdated(ctx context.Context, evt event.DomainEvent) error {
	updated, ok := evt.(*event.UserContactUpdated)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling UserContactUpdated event: UserID=%s", updated.AggregateID())
	return nil
}

func handleUserDeleted(ctx context.Context, evt event.DomainEvent) error {
	deleted, ok := evt.(*event.UserDeleted)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling UserDeleted event: UserID=%s", deleted.AggregateID())
	return nil
}

func handlePaymentCreated(ctx context.Context, evt event.DomainEvent) error {
	created, ok := evt.(*event.PaymentCreated)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling PaymentCreated event: PaymentID=%s, Amount=%d", created.PaymentID, created.Amount)
	return nil
}

func handlePaymentStatusChanged(ctx context.Context, evt event.DomainEvent) error {
	changed, ok := evt.(*event.PaymentStatusChanged)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling PaymentStatusChanged event: PaymentID=%s, Status=%s", changed.PaymentID, changed.NewStatus)
	return nil
}

func handlePetCreated(ctx context.Context, evt event.DomainEvent) error {
	created, ok := evt.(*event.PetCreated)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling PetCreated event: PetID=%s", created.AggregateID())
	return nil
}

func handlePetUpdated(ctx context.Context, evt event.DomainEvent) error {
	updated, ok := evt.(*event.PetUpdated)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling PetUpdated event: PetID=%s", updated.AggregateID())
	return nil
}

func handleServiceCreated(ctx context.Context, evt event.DomainEvent) error {
	created, ok := evt.(*event.ServiceCreated)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling ServiceCreated event: ServiceID=%s", created.AggregateID())
	return nil
}

func handleServiceUpdated(ctx context.Context, evt event.DomainEvent) error {
	updated, ok := evt.(*event.ServiceUpdated)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	log.Printf("Handling ServiceUpdated event: ServiceID=%s", updated.AggregateID())
	return nil
}

// ExampleCommandWithAsyncEvents demonstrates a complete command flow with async events
func ExampleCommandWithAsyncEvents(
	uowFactory repository.UnitOfWorkFactory,
	eventBus bus.EventBus,
) {
	ctx := context.Background()

	// Create command handler with UoW
	handler := command.NewCreateUserWithUoWHandler(uowFactory, eventBus)

	// Execute command
	cmd := &command.CreateUser{
		UserID: "user_456",
		Name:   "Jane Smith",
		Email:  "jane@example.com",
		Phone:  "+1234567890",
	}

	log.Println("Executing CreateUser command...")
	err := handler.Handle(ctx, cmd)
	if err != nil {
		log.Printf("Command failed: %v", err)
		return
	}

	log.Println("Command completed successfully!")
	log.Println("Events are being processed asynchronously in the background")

	// At this point:
	// 1. User aggregate is saved to event store ✓
	// 2. Events are being published asynchronously
	// 3. Projections are being updated in background
	// 4. Side effects (emails, etc.) are being triggered
	// 5. Command handler has returned to the HTTP controller
}

// ExampleWithErrorHandling demonstrates error handling in async scenarios
func ExampleWithErrorHandling() {
	ctx := context.Background()

	// Initialize async event bus
	eventBus := bus.NewAsyncEventBus()
	defer func() {
		eventBus.Wait()
		eventBus.Stop()
	}()

	eventBus.Start(ctx)

	// Subscribe a failing handler
	eventBus.Subscribe("TestEvent", bus.EventHandlerFunc(func(ctx context.Context, evt event.DomainEvent) error {
		// Simulate error
		return fmt.Errorf("handler failed: database connection lost")
	}))

	// Publish event
	testEvent := &event.UserCreated{
		UserID: "test_user",
		Name:   "Test User",
		Email:  "test@example.com",
	}

	err := eventBus.Publish(ctx, testEvent)
	if err != nil {
		log.Printf("Publish error: %v", err)
	}

	// Wait for async processing
	eventBus.Wait()

	// Check for errors
	select {
	case err := <-eventBus.GetErrors():
		log.Printf("Event handler error detected: %v", err)
		// In production: Send to error tracking, retry queue, etc.
	default:
		log.Println("No errors detected")
	}
}

// ExampleBatchPublishing demonstrates publishing multiple events at once
func ExampleBatchPublishing(uowFactory repository.UnitOfWorkFactory, eventBus bus.EventBus) {
	ctx := context.Background()

	// Create unit of work
	uow := uowFactory.CreateUnitOfWork()
	defer uow.Close()

	uow.Begin(ctx)

	// Create user aggregate
	user, _ := aggregate.NewUser("user_789", "Bob Johnson", "bob@example.com")
	user.UpdateContactInfo("+9876543210", "123 Main St")

	// Save to event store using UoW
	userRepo := uow.UserRepository()
	_ = userRepo.Save(ctx, user)

	// Commit transaction
	uow.Commit(ctx)

	// Publish all events in batch (async)
	events := user.GetUncommittedEvents()
	log.Printf("Publishing %d events in batch...", len(events))

	err := eventBus.PublishBatch(ctx, events)
	if err != nil {
		log.Printf("Warning: failed to publish events: %v", err)
	}

	log.Println("Batch publish completed (events processing asynchronously)")
}

// ExampleProjectionEventHandler shows how projections handle events
type UserProjectionEventHandler struct {
	projection projection.UserProjection
}

func (h *UserProjectionEventHandler) Handle(ctx context.Context, evt event.DomainEvent) error {
	switch e := evt.(type) {
	case *event.UserCreated:
		log.Printf("Updating projection for UserCreated: %s", e.UserID)
		return h.projection.HandleUserCreated(ctx, e)

	case *event.UserProfileUpdated:
		log.Printf("Updating projection for UserProfileUpdated: %s", e.AggregateID())
		return h.projection.HandleUserProfileUpdated(ctx, e)

	case *event.UserContactUpdated:
		log.Printf("Updating projection for UserContactUpdated: %s", e.AggregateID())
		return h.projection.HandleUserContactUpdated(ctx, e)

	case *event.UserDeleted:
		log.Printf("Updating projection for UserDeleted: %s", e.AggregateID())
		return h.projection.HandleUserDeleted(ctx, e)

	default:
		return fmt.Errorf("unknown event type: %T", evt)
	}
}

// ExampleGracefulShutdown demonstrates proper shutdown sequence
func ExampleGracefulShutdown() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize event bus
	eventBus := bus.NewAsyncEventBus()
	eventBus.Start(ctx)

	// ... application runs ...

	// Shutdown signal received
	log.Println("Shutdown signal received, starting graceful shutdown...")

	// Cancel context (stops error monitor)
	cancel()

	// Wait for all in-flight events to complete
	log.Println("Waiting for in-flight events to complete...")
	eventBus.Wait()

	// Stop event bus
	log.Println("Stopping event bus...")
	eventBus.Stop()

	log.Println("Graceful shutdown completed")
}

// ExampleWithTimeout demonstrates context timeout handling
func ExampleWithTimeout(eventBus bus.EventBus) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publish event with timeout
	evt := &event.UserCreated{
		UserID: "timeout_test",
		Name:   "Timeout Test",
		Email:  "timeout@example.com",
	}

	err := eventBus.Publish(ctx, evt)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Event publishing timed out")
		} else {
			log.Printf("Event publishing failed: %v", err)
		}
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		eventBus.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All events processed within timeout")
	case <-ctx.Done():
		log.Println("Timeout waiting for events to complete")
	}
}
