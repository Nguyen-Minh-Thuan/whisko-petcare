package bus

import (
	"context"
	"fmt"
	"sync"

	"whisko-petcare/internal/domain/event"
)

// EventBus defines the contract for event publishing/subscribing
type EventBus interface {
	Publish(ctx context.Context, event event.DomainEvent) error
	PublishSync(ctx context.Context, event event.DomainEvent) error
	PublishBatch(ctx context.Context, events []event.DomainEvent) error
	Subscribe(eventType string, handler EventHandler) error
	Start(ctx context.Context) error
	Stop() error
	Wait() // Wait for async operations to complete
}

// EventHandler handles domain events
type EventHandler interface {
	Handle(ctx context.Context, event event.DomainEvent) error
}

// EventHandlerFunc allows functions to implement EventHandler
type EventHandlerFunc func(ctx context.Context, event event.DomainEvent) error

func (f EventHandlerFunc) Handle(ctx context.Context, event event.DomainEvent) error {
	return f(ctx, event)
}

// InMemoryEventBus implements EventBus in memory
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
	running  bool
}

func NewInMemoryEventBus() EventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (b *InMemoryEventBus) Publish(ctx context.Context, event event.DomainEvent) error {
	b.mutex.RLock()
	handlers := b.handlers[event.EventType()]
	b.mutex.RUnlock()

	fmt.Printf("📢 EventBus: Publishing event '%s' to %d handlers\n", event.EventType(), len(handlers))

	if len(handlers) == 0 {
		fmt.Printf("⚠️  WARNING: No handlers registered for event type '%s'\n", event.EventType())
		return nil
	}

	var errs []error

	for i, handler := range handlers {
		fmt.Printf("  → Handler %d/%d processing event '%s'\n", i+1, len(handlers), event.EventType())
		if err := handler.Handle(ctx, event); err != nil {
			fmt.Printf("  ❌ Handler %d failed: %v\n", i+1, err)
			errs = append(errs, fmt.Errorf("handler error for %s: %w", event.EventType(), err))
		} else {
			fmt.Printf("  ✅ Handler %d completed successfully\n", i+1)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("event handling errors: %v", errs)
	}

	return nil
}

func (b *InMemoryEventBus) PublishSync(ctx context.Context, event event.DomainEvent) error {
	return b.Publish(ctx, event)
}

func (b *InMemoryEventBus) PublishBatch(ctx context.Context, events []event.DomainEvent) error {
	fmt.Printf("📦 EventBus: Publishing batch of %d events\n", len(events))
	for i, evt := range events {
		fmt.Printf("  Event %d/%d: %s\n", i+1, len(events), evt.EventType())
		if err := b.Publish(ctx, evt); err != nil {
			fmt.Printf("  ❌ Batch publishing failed at event %d: %v\n", i+1, err)
			return err
		}
	}
	fmt.Printf("✅ Batch published successfully (%d events)\n", len(events))
	return nil
}

func (b *InMemoryEventBus) Wait() {
	// No-op for synchronous implementation
}

func (b *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	return nil
}

func (b *InMemoryEventBus) Start(ctx context.Context) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.running = true
	return nil
}

func (b *InMemoryEventBus) Stop() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.running = false
	return nil
}
