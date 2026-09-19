package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"dcisp/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	Client *redis.Client
}

// Menginisialisasi koneksi client ke Redis server.
func NewRedisClient(cfg *config.Config) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to ping redis at %s: %w", addr, err)
	}

	log.Printf("Connected to Redis successfully (%s)", addr)
	return &RedisClient{Client: rdb}, nil
}

// Menutup koneksi client Redis.
func (r *RedisClient) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}
