package bus

import (
	"context"
	"log"
	"sync"

	"whisko-petcare/internal/domain/event"
)

// AsyncEventBus implements EventBus with asynchronous publishing
type AsyncEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
	wg       sync.WaitGroup
	errorCh  chan error
}

// NewAsyncEventBus creates a new async event bus
func NewAsyncEventBus() *AsyncEventBus {
	return &AsyncEventBus{
		handlers: make(map[string][]EventHandler),
		errorCh:  make(chan error, 100), // Buffered channel for errors
	}
}

// Subscribe registers a handler for a specific event type
func (b *AsyncEventBus) Subscribe(eventType string, handler EventHandler) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.handlers[eventType] = append(b.handlers[eventType], handler)
	return nil
}

// Start initializes the event bus
func (b *AsyncEventBus) Start(ctx context.Context) error {
	b.StartErrorMonitor(ctx)
	return nil
}

// Stop gracefully shuts down the event bus
func (b *AsyncEventBus) Stop() error {
	b.Wait()
	b.Close()
	return nil
}

// Publish publishes an event asynchronously to all subscribed handlers
func (b *AsyncEventBus) Publish(ctx context.Context, evt event.DomainEvent) error {
	b.mu.RLock()
	handlers, exists := b.handlers[evt.EventType()]
	b.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		// No handlers registered, silently continue
		return nil
	}

	// Increment wait group for async operations
	b.wg.Add(len(handlers))

	// Publish to each handler asynchronously
	for _, handler := range handlers {
		go b.publishToHandler(ctx, handler, evt)
	}

	return nil
}

// PublishSync publishes an event synchronously (waits for all handlers to complete)
func (b *AsyncEventBus) PublishSync(ctx context.Context, evt event.DomainEvent) error {
	b.mu.RLock()
	handlers, exists := b.handlers[evt.EventType()]
	b.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		return nil
	}

	// Process synchronously
	for _, handler := range handlers {
		if err := handler.Handle(ctx, evt); err != nil {
			log.Printf("Error handling event %s: %v", evt.EventType(), err)
			return err
		}
	}

	return nil
}

// PublishBatch publishes multiple events asynchronously
func (b *AsyncEventBus) PublishBatch(ctx context.Context, events []event.DomainEvent) error {
	for _, evt := range events {
		if err := b.Publish(ctx, evt); err != nil {
			return err
		}
	}
	return nil
}

// Wait waits for all async event handlers to complete
func (b *AsyncEventBus) Wait() {
	b.wg.Wait()
}

// GetErrors returns a channel to receive errors from async handlers
func (b *AsyncEventBus) GetErrors() <-chan error {
	return b.errorCh
}

// Close closes the error channel
func (b *AsyncEventBus) Close() {
	close(b.errorCh)
}

// publishToHandler handles the async publishing to a single handler
func (b *AsyncEventBus) publishToHandler(ctx context.Context, handler EventHandler, evt event.DomainEvent) {
	defer b.wg.Done()

	if err := handler.Handle(ctx, evt); err != nil {
		log.Printf("Error handling event %s: %v", evt.EventType(), err)
		
		// Send error to channel (non-blocking)
		select {
		case b.errorCh <- err:
		default:
			log.Printf("Error channel full, dropping error: %v", err)
		}
	}
}

// StartErrorMonitor starts a goroutine to monitor and log errors
func (b *AsyncEventBus) StartErrorMonitor(ctx context.Context) {
	go func() {
		for {
			select {
			case err := <-b.errorCh:
				log.Printf("Async event handler error: %v", err)
			case <-ctx.Done():
				return
			}
		}
	}()
}
