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
	Subscribe(eventType string, handler EventHandler) error
	Start(ctx context.Context) error
	Stop() error
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

	var errs []error

	for _, handler := range handlers {
		if err := handler.Handle(ctx, event); err != nil {
			errs = append(errs, fmt.Errorf("handler error for %s: %w", event.EventType(), err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("event handling errors: %v", errs)
	}

	return nil
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
