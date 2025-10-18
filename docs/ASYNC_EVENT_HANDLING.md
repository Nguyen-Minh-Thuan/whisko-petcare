# Async Event Handling Implementation

## Overview

This project implements asynchronous event publishing using Go's goroutines and channels to improve system performance and resilience. Events are published in a non-blocking manner, allowing command handlers to return quickly while events are processed in the background.

## Architecture

### Event Bus Implementations

We provide two EventBus implementations:

1. **InMemoryEventBus** (Synchronous)
   - Simple, blocking event publishing
   - Handlers execute sequentially
   - Errors are propagated immediately
   - Good for: Testing, simple scenarios

2. **AsyncEventBus** (Asynchronous)
   - Non-blocking event publishing
   - Handlers execute concurrently in goroutines
   - Errors are collected in a buffered channel
   - Good for: Production, high-throughput scenarios

### EventBus Interface

```go
type EventBus interface {
    Publish(ctx context.Context, event event.DomainEvent) error
    PublishSync(ctx context.Context, event event.DomainEvent) error
    PublishBatch(ctx context.Context, events []event.DomainEvent) error
    Subscribe(eventType string, handler EventHandler) error
    Start(ctx context.Context) error
    Stop() error
    Wait() // Wait for async operations to complete
}
```

## Usage Patterns

### 1. Publishing Events in Command Handlers

**Before (Synchronous):**
```go
// Publish events synchronously - blocks until all handlers complete
for _, event := range user.GetUncommittedEvents() {
    if err := h.eventBus.Publish(ctx, event); err != nil {
        // Error stops execution
        return err
    }
}
```

**After (Asynchronous):**
```go
// Publish events asynchronously - returns immediately
events := user.GetUncommittedEvents()
if err := h.eventBus.PublishBatch(ctx, events); err != nil {
    // Log warning but don't fail the command
    fmt.Printf("Warning: failed to publish events: %v\n", err)
}
// Command returns while events are being processed
```

### 2. Event Handler Implementation

```go
type UserProjectionHandler struct {
    projection projection.UserProjection
}

func (h *UserProjectionHandler) Handle(ctx context.Context, evt event.DomainEvent) error {
    switch e := evt.(type) {
    case *event.UserCreated:
        return h.projection.HandleUserCreated(ctx, e)
    case *event.UserProfileUpdated:
        return h.projection.HandleUserProfileUpdated(ctx, e)
    default:
        return nil
    }
}
```

### 3. Subscribing to Events

```go
// Initialize event bus
eventBus := bus.NewAsyncEventBus()

// Subscribe handlers
userProjection := projection.NewMongoUserProjection(db)
userHandler := &UserProjectionHandler{projection: userProjection}

eventBus.Subscribe("UserCreated", userHandler)
eventBus.Subscribe("UserProfileUpdated", userHandler)
eventBus.Subscribe("UserContactUpdated", userHandler)
eventBus.Subscribe("UserDeleted", userHandler)

// Start event bus with error monitoring
ctx := context.Background()
eventBus.Start(ctx)
```

### 4. Error Monitoring

```go
// Start background error monitor
eventBus.StartErrorMonitor(ctx)

// Or manually check errors
go func() {
    for err := range eventBus.GetErrors() {
        log.Printf("Event handler error: %v", err)
        // Send to error tracking service (Sentry, etc.)
    }
}()
```

### 5. Graceful Shutdown

```go
// Wait for all in-flight events to complete
eventBus.Wait()

// Stop the event bus
eventBus.Stop()
```

## Benefits

### 1. Performance
- **Non-blocking**: Command handlers return immediately after saving events
- **Concurrent**: Multiple event handlers execute in parallel
- **Scalable**: Can handle high event throughput

### 2. Resilience
- **Error Isolation**: Event handler failures don't fail commands
- **Eventual Consistency**: Projections update asynchronously
- **Retry Logic**: Can implement retry in event handlers

### 3. Decoupling
- **Loose Coupling**: Commands don't wait for projections
- **Independent Scaling**: Event handlers can scale independently
- **Flexible**: Easy to add new event handlers

## Best Practices

### 1. Event Handler Idempotency

Always make event handlers idempotent (safe to execute multiple times):

```go
func (p *MongoUserProjection) HandleUserCreated(ctx context.Context, evt *event.UserCreated) error {
    // Use upsert to make it idempotent
    filter := bson.M{"_id": evt.AggregateID()}
    update := bson.M{
        "$setOnInsert": UserReadModel{
            ID:        evt.UserID,
            Name:      evt.Name,
            Email:     evt.Email,
            CreatedAt: evt.OccurredAt(),
        },
    }
    opts := options.Update().SetUpsert(true)
    _, err := p.collection.UpdateOne(ctx, filter, update, opts)
    return err
}
```

### 2. Error Handling Strategy

```go
// In command handlers - log but don't fail
if err := h.eventBus.PublishBatch(ctx, events); err != nil {
    log.Printf("Warning: failed to publish events: %v", err)
    // Consider: Send to dead letter queue
    // Consider: Implement outbox pattern for guaranteed delivery
}

// In event handlers - return errors for monitoring
func (h *Handler) Handle(ctx context.Context, evt event.DomainEvent) error {
    if err := h.process(evt); err != nil {
        // Error will be sent to error channel
        return fmt.Errorf("failed to process %s: %w", evt.EventType(), err)
    }
    return nil
}
```

### 3. Context Handling

```go
// Use context for cancellation and timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := eventBus.Publish(ctx, event); err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        // Handle timeout
    }
}
```

### 4. Monitoring and Observability

```go
// Add metrics
func (h *Handler) Handle(ctx context.Context, evt event.DomainEvent) error {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        metrics.RecordEventHandling(evt.EventType(), duration)
    }()
    
    return h.process(evt)
}
```

## Advanced Patterns

### 1. Outbox Pattern (Guaranteed Delivery)

For critical events that must be processed:

```go
// Save events to outbox table atomically with aggregate
func (r *Repository) Save(ctx context.Context, agg aggregate.Aggregate) error {
    session, err := r.client.StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(ctx)
    
    _, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
        // Save events to event store
        if err := r.saveEvents(sc, agg); err != nil {
            return nil, err
        }
        
        // Save events to outbox
        if err := r.saveToOutbox(sc, agg.GetUncommittedEvents()); err != nil {
            return nil, err
        }
        
        return nil, nil
    })
    
    return err
}

// Background worker processes outbox
func (w *OutboxWorker) Run(ctx context.Context) {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            w.processOutbox(ctx)
        case <-ctx.Done():
            return
        }
    }
}
```

### 2. Event Versioning

Handle multiple event versions:

```go
func (h *Handler) Handle(ctx context.Context, evt event.DomainEvent) error {
    switch e := evt.(type) {
    case *event.UserCreatedV1:
        return h.handleUserCreatedV1(ctx, e)
    case *event.UserCreatedV2:
        return h.handleUserCreatedV2(ctx, e)
    default:
        return fmt.Errorf("unknown event version: %T", evt)
    }
}
```

### 3. Saga Pattern

For distributed transactions:

```go
type PaymentSaga struct {
    eventBus bus.EventBus
}

func (s *PaymentSaga) Handle(ctx context.Context, evt event.DomainEvent) error {
    switch e := evt.(type) {
    case *event.PaymentCreated:
        // Trigger inventory reservation
        return s.reserveInventory(ctx, e)
    case *event.InventoryReserved:
        // Trigger shipping
        return s.createShipment(ctx, e)
    case *event.InventoryReservationFailed:
        // Compensate: cancel payment
        return s.cancelPayment(ctx, e)
    }
    return nil
}
```

## Testing

### Testing Async Event Handlers

```go
func TestAsyncEventPublishing(t *testing.T) {
    // Use sync bus for deterministic tests
    eventBus := bus.NewInMemoryEventBus()
    
    // Or use async bus with Wait()
    asyncBus := bus.NewAsyncEventBus()
    defer asyncBus.Stop()
    
    // Publish event
    event := &event.UserCreated{...}
    err := asyncBus.Publish(ctx, event)
    assert.NoError(t, err)
    
    // Wait for async processing
    asyncBus.Wait()
    
    // Verify results
    user, err := projection.GetByID(ctx, event.UserID)
    assert.NoError(t, err)
    assert.NotNil(t, user)
}
```

## Migration Guide

### Migrating from Sync to Async

1. **Update Command Handlers**
   ```go
   // Replace loop with batch publish
   - for _, event := range agg.GetUncommittedEvents() {
   -     h.eventBus.Publish(ctx, event)
   - }
   + events := agg.GetUncommittedEvents()
   + h.eventBus.PublishBatch(ctx, events)
   ```

2. **Update Event Bus Initialization**
   ```go
   - eventBus := bus.NewInMemoryEventBus()
   + eventBus := bus.NewAsyncEventBus()
   + eventBus.Start(ctx)
   + defer eventBus.Stop()
   ```

3. **Add Error Monitoring**
   ```go
   + eventBus.StartErrorMonitor(ctx)
   ```

4. **Update Shutdown Logic**
   ```go
   + eventBus.Wait() // Wait for in-flight events
   + eventBus.Stop() // Stop gracefully
   ```

## Performance Considerations

- **Buffer Size**: Error channel has 100-item buffer (configurable)
- **Goroutine Pool**: Consider limiting concurrent handlers for resource control
- **Context Timeouts**: Add timeouts to prevent hanging handlers
- **Monitoring**: Track event processing latency and error rates

## Conclusion

Asynchronous event handling improves system performance and resilience while maintaining the benefits of event-driven architecture. Use `AsyncEventBus` for production workloads and `InMemoryEventBus` for testing or simple scenarios.
