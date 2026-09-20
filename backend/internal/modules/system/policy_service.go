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

type PolicyService struct {
	repo *PolicyRepository
	rdb  *database.RedisClient
}

// Menginisialisasi instance baru service policy engine dengan dukungan cache Redis.
func NewPolicyService(repo *PolicyRepository, rdb *database.RedisClient) *PolicyService {
	return &PolicyService{repo: repo, rdb: rdb}
}

// Menghapus cache kebijakan pada Redis berdasarkan domain kebijakan.
func (s *PolicyService) invalidateDomainCache(ctx context.Context, domain string) {
	if s.rdb != nil && s.rdb.Client != nil {
		cacheKey := fmt.Sprintf("policy:domain:%s", strings.ToUpper(domain))
		if err := s.rdb.Client.Del(ctx, cacheKey).Err(); err != nil {
			log.Printf("Peringatan: Gagal menghapus cache policy domain %s: %v", domain, err)
		}
	}
}

// Membuat kebijakan baru setelah memvalidasi format JSONB dan aturan tanggal efektif.
func (s *PolicyService) CreatePolicy(ctx context.Context, req *CreatePolicyRequest) (*Policy, error) {
	var condMap map[string]interface{}
	if err := json.Unmarshal(req.ConditionRules, &condMap); err != nil {
		return nil, errors.New("format condition_rules harus berupa objek JSON yang valid")
	}
	if len(req.ActionDefinitions) > 0 {
		var actMap map[string]interface{}
		if err := json.Unmarshal(req.ActionDefinitions, &actMap); err != nil {
			return nil, errors.New("format action_definitions harus berupa objek JSON yang valid")
		}
	}

	effectiveDate, err := time.Parse("2006-01-02", req.EffectiveDate)
	if err != nil {
		return nil, fmt.Errorf("format tanggal efektif tidak valid (harus YYYY-MM-DD): %w", err)
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	actions := req.ActionDefinitions
	if len(actions) == 0 {
		actions = json.RawMessage("{}")
	}

	p := &Policy{
		ID:                uuid.New(),
		PolicyDomain:      strings.ToUpper(req.PolicyDomain),
		Name:              req.Name,
		EffectiveDate:     effectiveDate,
		ConditionRules:    req.ConditionRules,
		ActionDefinitions: actions,
		PriorityOrder:     req.PriorityOrder,
		IsActive:          isActive,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		return nil, err
	}

	s.invalidateDomainCache(ctx, p.PolicyDomain)
	return p, nil
}

// Mengambil data kebijakan berdasarkan identitas unik UUID.
func (s *PolicyService) GetPolicyByID(ctx context.Context, id uuid.UUID) (*Policy, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("kebijakan tidak ditemukan")
	}
	return p, nil
}

// Mengambil daftar kebijakan aktif untuk domain tertentu dengan dukungan cache Redis.
func (s *PolicyService) GetActivePolicies(ctx context.Context, domain string) ([]Policy, error) {
	upperDomain := strings.ToUpper(domain)
	cacheKey := fmt.Sprintf("policy:domain:%s", upperDomain)

	if s.rdb != nil && s.rdb.Client != nil {
		val, err := s.rdb.Client.Get(ctx, cacheKey).Result()
		if err == nil {
			var cachedPolicies []Policy
			if err := json.Unmarshal([]byte(val), &cachedPolicies); err == nil {
				return cachedPolicies, nil
			}
		} else if err != redis.Nil {
			log.Printf("Peringatan: Gagal membaca cache policy dari Redis: %v", err)
		}
	}

	policies, err := s.repo.GetActiveByDomain(ctx, upperDomain)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil && s.rdb.Client != nil && len(policies) > 0 {
		if data, err := json.Marshal(policies); err == nil {
			_ = s.rdb.Client.Set(ctx, cacheKey, data, 10*time.Minute).Err()
		}
	}

	return policies, nil
}

// Mengambil daftar kebijakan dengan paginasi dan filter opsional.
func (s *PolicyService) ListPolicies(ctx context.Context, domain string, isActive *bool, page, limit int) ([]Policy, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	return s.repo.List(ctx, strings.ToUpper(domain), isActive, limit, offset)
}

// Memperbarui konfigurasi kebijakan dan mengosongkan cache Redis terkait.
func (s *PolicyService) UpdatePolicy(ctx context.Context, id uuid.UUID, req *UpdatePolicyRequest) (*Policy, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("kebijakan tidak ditemukan")
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.EffectiveDate != "" {
		parsedDate, err := time.Parse("2006-01-02", req.EffectiveDate)
		if err != nil {
			return nil, fmt.Errorf("format tanggal efektif tidak valid: %w", err)
		}
		existing.EffectiveDate = parsedDate
	}
	if len(req.ConditionRules) > 0 {
		if !json.Valid(req.ConditionRules) {
			return nil, errors.New("format condition_rules tidak valid")
		}
		existing.ConditionRules = req.ConditionRules
	}
	if len(req.ActionDefinitions) > 0 {
		if !json.Valid(req.ActionDefinitions) {
			return nil, errors.New("format action_definitions tidak valid")
		}
		existing.ActionDefinitions = req.ActionDefinitions
	}
	if req.PriorityOrder != nil {
		existing.PriorityOrder = *req.PriorityOrder
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, err
	}

	s.invalidateDomainCache(ctx, existing.PolicyDomain)
	return existing, nil
}

// Menghapus data kebijakan dari sistem dan memvalidasi penghapusan cache.
func (s *PolicyService) DeletePolicy(ctx context.Context, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("kebijakan tidak ditemukan")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.invalidateDomainCache(ctx, existing.PolicyDomain)
	return nil
}

// Mengevaluasi aturan kebijakan aktif terhadap data konteks masukan dan mengembalikan tindakan yang berlaku.
func (s *PolicyService) EvaluatePolicy(ctx context.Context, domain string, contextData map[string]interface{}) (*PolicyEvaluationResult, error) {
	policies, err := s.GetActivePolicies(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil kebijakan aktif untuk evaluasi: %w", err)
	}

	if len(policies) == 0 {
		return &PolicyEvaluationResult{
			Allowed:    true,
			Domain:     strings.ToUpper(domain),
			Actions:    map[string]interface{}{},
			Conditions: map[string]interface{}{},
			Reason:     "Tidak ada kebijakan aktif untuk domain ini, menggunakan perilaku default",
		}, nil
	}

	// Evaluasi kebijakan berdasarkan prioritas urutan (priority_order terendah diproses duluan)
	for _, p := range policies {
		var conditions map[string]interface{}
		if err := json.Unmarshal(p.ConditionRules, &conditions); err != nil {
			continue
		}

		var actions map[string]interface{}
		if len(p.ActionDefinitions) > 0 {
			_ = json.Unmarshal(p.ActionDefinitions, &actions)
		}

		// Evaluasi kesesuaian: periksa apakah aturan melarang atau mengizinkan
		// Jika condition_rules memuat aturan "block": true, maka allowed = false
		isBlocked := false
		if b, ok := conditions["blocked"].(bool); ok && b {
			isBlocked = true
		}

		return &PolicyEvaluationResult{
			Allowed:    !isBlocked,
			PolicyID:   p.ID,
			Domain:     p.PolicyDomain,
			Actions:    actions,
			Conditions: conditions,
			Reason:     fmt.Sprintf("Dievaluasi berdasarkan kebijakan '%s'", p.Name),
		}, nil
	}

	return &PolicyEvaluationResult{
		Allowed: true,
		Domain:  strings.ToUpper(domain),
		Actions: map[string]interface{}{},
	}, nil
}
