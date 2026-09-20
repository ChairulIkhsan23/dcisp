package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/shared/eventbus"
	"dcisp/backend/internal/worker"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji siklus hidup antrean worker Asynq termasuk enqueue task dan eksekusi handler.
func TestAsynqWorkerEnqueueAndExecution(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	w := worker.NewWorkerPool(cfg)
	defer w.Shutdown()

	taskProcessed := make(chan bool, 1)
	testTaskType := "test:task:sample"

	w.RegisterHandler(testTaskType, func(ctx context.Context, task *asynq.Task) error {
		taskProcessed <- true
		return nil
	})

	err = w.Start()
	require.NoError(t, err)

	ctx := context.Background()
	payload := map[string]string{"message": "Halo Worker"}
	info, err := w.EnqueueTask(ctx, testTaskType, payload, worker.QueueDefault, 2, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, info.ID)

	select {
	case <-taskProcessed:
		// Sukses diproses worker
	case <-time.After(3 * time.Second):
		t.Fatal("Batas waktu habis: Task Asynq tidak diproses dalam 3 detik")
	}
}

// Menguji publikasi dan penerimaan peristiwa domain secara asinkron melalui event bus terpusat.
func TestEventBusPublishAndSubscribe(t *testing.T) {
	bus := eventbus.NewEventBus(nil)

	var wg sync.WaitGroup
	wg.Add(2)

	receivedSpecific := false
	receivedWildcard := false

	bus.Subscribe("attendance.scanned", func(ctx context.Context, event eventbus.DomainEvent) error {
		receivedSpecific = true
		assert.Equal(t, "12345", event.AggregateID)
		wg.Done()
		return nil
	})

	bus.Subscribe("*", func(ctx context.Context, event eventbus.DomainEvent) error {
		receivedWildcard = true
		wg.Done()
		return nil
	})

	event := eventbus.DomainEvent{
		ID:          uuid.New(),
		Type:        "attendance.scanned",
		AggregateID: "12345",
		Payload: map[string]interface{}{
			"user_id": uuid.New().String(),
			"status":  "CHECK_IN",
		},
	}

	err := bus.Publish(context.Background(), event)
	require.NoError(t, err)

	// Tunggu dispatch lokal selesai
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		assert.True(t, receivedSpecific)
		assert.True(t, receivedWildcard)
	case <-time.After(2 * time.Second):
		t.Fatal("Batas waktu habis: Event tidak diterima oleh subscriber dalam 2 detik")
	}
}
