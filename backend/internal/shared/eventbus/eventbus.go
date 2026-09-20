package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"dcisp/backend/internal/worker"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const TaskTypeDomainEvent = "domain:event:dispatch"

type DomainEvent struct {
	ID          uuid.UUID              `json:"id"`
	Type        string                 `json:"type"`
	OccurredAt  time.Time              `json:"occurred_at"`
	AggregateID string                 `json:"aggregate_id"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	Payload     map[string]interface{} `json:"payload"`
}

type EventHandler func(ctx context.Context, event DomainEvent) error

type EventBus struct {
	workerPool *worker.WorkerPool
	mu         sync.RWMutex
	listeners  map[string][]EventHandler
}

// Menginisialisasi event bus terpusat untuk dispatch peristiwa domain antar modul.
func NewEventBus(workerPool *worker.WorkerPool) *EventBus {
	bus := &EventBus{
		workerPool: workerPool,
		listeners:  make(map[string][]EventHandler),
	}

	if workerPool != nil {
		workerPool.RegisterHandler(TaskTypeDomainEvent, func(ctx context.Context, t *asynq.Task) error {
			var event DomainEvent
			if err := json.Unmarshal(t.Payload(), &event); err != nil {
				return fmt.Errorf("gagal unmarshal payload event: %w", err)
			}
			return bus.dispatchLocal(ctx, event)
		})
	}

	return bus
}

// Mendaftarkan fungsi pendengar untuk jenis peristiwa domain tertentu.
func (b *EventBus) Subscribe(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.listeners[eventType] = append(b.listeners[eventType], handler)
}

// Mempublikasikan peristiwa domain baru secara asinkron ke antrean worker dan pendengar lokal.
func (b *EventBus) Publish(ctx context.Context, event DomainEvent) error {
	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}

	// 1. Dispatch ke listener lokal segera (untuk SSE stream atau notifikasi in-process)
	go func(ev DomainEvent) {
		_ = b.dispatchLocal(context.Background(), ev)
	}(event)

	// 2. Jika worker pool aktif, masukkan ke antrean persistent Asynq
	if b.workerPool != nil {
		_, err := b.workerPool.EnqueueTask(ctx, TaskTypeDomainEvent, event, worker.QueueDefault, 3, 0)
		if err != nil {
			return fmt.Errorf("gagal memasukkan event ke antrean worker: %w", err)
		}
	}

	return nil
}

// Mendistribusikan peristiwa ke seluruh fungsi penangan lokal yang terdaftar.
func (b *EventBus) dispatchLocal(ctx context.Context, event DomainEvent) error {
	b.mu.RLock()
	handlers := append([]EventHandler{}, b.listeners[event.Type]...)
	wildcardHandlers := append([]EventHandler{}, b.listeners["*"]...)
	b.mu.RUnlock()

	var errs []error
	for _, h := range append(handlers, wildcardHandlers...) {
		if err := h(ctx, event); err != nil {
			log.Printf("Kesalahan penangan event [%s]: %v", event.Type, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("terjadi %d kesalahan saat pemrosesan event", len(errs))
	}
	return nil
}
