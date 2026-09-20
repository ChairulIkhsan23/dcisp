package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"dcisp/backend/internal/config"
	"github.com/hibiken/asynq"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)

type WorkerPool struct {
	client *asynq.Client
	server *asynq.Server
	mux    *asynq.ServeMux
}

// Menginisialisasi infrastruktur antrean latar belakang Asynq yang terhubung ke Redis.
func NewWorkerPool(cfg *config.Config) *WorkerPool {
	redisOpt := asynq.RedisClientOpt{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	}

	client := asynq.NewClient(redisOpt)
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Concurrency: 10,
			Queues: map[string]int{
				QueueCritical: 6,
				QueueDefault:  3,
				QueueLow:      1,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				log.Printf("Kesalahan pemrosesan task [%s]: %v", task.Type(), err)
			}),
		},
	)

	return &WorkerPool{
		client: client,
		server: server,
		mux:    asynq.NewServeMux(),
	}
}

// Mendaftarkan fungsi penangan untuk tipe task tertentu pada antrean Asynq.
func (w *WorkerPool) RegisterHandler(taskType string, handler func(context.Context, *asynq.Task) error) {
	w.mux.HandleFunc(taskType, handler)
}

// Menjalankan pemrosesan antrean latar belakang secara asinkron dalam goroutine.
func (w *WorkerPool) Start() error {
	return w.server.Start(w.mux)
}

// Menutup server worker secara aman dan menunggu penyelesaian task aktif.
func (w *WorkerPool) Shutdown() {
	w.server.Shutdown()
	_ = w.client.Close()
}

// Memasukkan task baru ke dalam antrean latar belakang dengan opsi retry dan penundaan.
func (w *WorkerPool) EnqueueTask(ctx context.Context, taskType string, payload interface{}, queue string, maxRetry int, delay time.Duration) (*asynq.TaskInfo, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("gagal mem-marshal payload task: %w", err)
	}

	task := asynq.NewTask(taskType, data)
	opts := []asynq.Option{
		asynq.Queue(queue),
		asynq.MaxRetry(maxRetry),
	}
	if delay > 0 {
		opts = append(opts, asynq.ProcessIn(delay))
	}

	info, err := w.client.EnqueueContext(ctx, task, opts...)
	if err != nil {
		return nil, fmt.Errorf("gagal memasukkan task ke antrean: %w", err)
	}
	return info, nil
}
