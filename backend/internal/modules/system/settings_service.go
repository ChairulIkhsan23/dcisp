package system

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"dcisp/backend/internal/database"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SettingsService struct {
	repo *SettingsRepository
	rdb  *database.RedisClient
}

// Menginisialisasi instance baru service pengaturan sistem dengan dukungan cache Redis.
func NewSettingsService(repo *SettingsRepository, rdb *database.RedisClient) *SettingsService {
	return &SettingsService{repo: repo, rdb: rdb}
}

// Menghapus cache pengaturan sistem pada Redis berdasarkan kunci pengaturan.
func (s *SettingsService) invalidateCache(ctx context.Context, key string) {
	if s.rdb != nil && s.rdb.Client != nil {
		cacheKey := fmt.Sprintf("settings:%s", strings.ToLower(key))
		if err := s.rdb.Client.Del(ctx, cacheKey).Err(); err != nil {
			log.Printf("Peringatan: Gagal menghapus cache setting %s: %v", key, err)
		}
	}
}

// Mengambil nilai pengaturan sistem dengan pengecekan cache Redis terlebih dahulu.
func (s *SettingsService) GetSetting(ctx context.Context, key string) (*SystemSetting, error) {
	lowerKey := strings.ToLower(key)
	cacheKey := fmt.Sprintf("settings:%s", lowerKey)

	if s.rdb != nil && s.rdb.Client != nil {
		val, err := s.rdb.Client.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedSetting SystemSetting
			if err := json.Unmarshal([]byte(val), &cachedSetting); err == nil {
				return &cachedSetting, nil
			}
		} else if err != redis.Nil {
			log.Printf("Peringatan: Gagal membaca cache setting dari Redis: %v", err)
		}
	}

	setting, err := s.repo.GetByKey(ctx, lowerKey)
	if err != nil {
		return nil, err
	}
	if setting == nil {
		return nil, errors.New("pengaturan sistem tidak ditemukan")
	}

	if s.rdb != nil && s.rdb.Client != nil {
		if data, err := json.Marshal(setting); err == nil {
			_ = s.rdb.Client.Set(ctx, cacheKey, data, 30*time.Minute).Err()
		}
	}

	return setting, nil
}

// Menyimpan atau memperbarui nilai pengaturan sistem terpusat.
func (s *SettingsService) SetSetting(ctx context.Context, req *SetSettingRequest) (*SystemSetting, error) {
	lowerKey := strings.ToLower(req.SettingKey)
	existing, err := s.repo.GetByKey(ctx, lowerKey)
	if err != nil {
		return nil, err
	}

	if existing == nil {
		newSetting := &SystemSetting{
			ID:           uuid.New(),
			SettingKey:   lowerKey,
			SettingValue: req.SettingValue,
			IsEncrypted:  req.IsEncrypted,
			Description:  req.Description,
		}
		if err := s.repo.Create(ctx, newSetting); err != nil {
			return nil, err
		}
		s.invalidateCache(ctx, lowerKey)
		return newSetting, nil
	}

	existing.SettingValue = req.SettingValue
	existing.IsEncrypted = req.IsEncrypted
	if req.Description != "" {
		existing.Description = req.Description
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.invalidateCache(ctx, lowerKey)
	return existing, nil
}

// Mengambil seluruh daftar pengaturan sistem dengan penyamaran nilai untuk data terenkripsi.
func (s *SettingsService) ListSettings(ctx context.Context) ([]SystemSetting, error) {
	settings, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Samarkan nilai pengaturan terenkripsi untuk respon publik
	sanitized := make([]SystemSetting, len(settings))
	for i, st := range settings {
		sanitized[i] = st
		if st.IsEncrypted {
			sanitized[i].SettingValue = "********"
		}
	}

	return sanitized, nil
}

// Menghapus data pengaturan sistem berdasarkan kunci unik dan mengosongkan cachenya.
func (s *SettingsService) DeleteSetting(ctx context.Context, key string) error {
	lowerKey := strings.ToLower(key)
	existing, err := s.repo.GetByKey(ctx, lowerKey)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("pengaturan sistem tidak ditemukan")
	}

	if err := s.repo.Delete(ctx, lowerKey); err != nil {
		return err
	}

	s.invalidateCache(ctx, lowerKey)
	return nil
}
