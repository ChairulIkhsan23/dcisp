package people

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"dcisp/backend/internal/modules/system"
	"github.com/google/uuid"
)

// Membangun entri audit untuk impor institusi eksternal tanpa membocorkan kredensial API.
func importAuditEntry(draft *Institution) system.MutationAuditEntry {
	externalID := ""
	if draft.ExternalID != nil {
		externalID = *draft.ExternalID
	}
	source := ""
	if draft.Source != nil {
		source = *draft.Source
	}
	return system.MutationAuditEntry{
		EventName:    "IMPORT_EXTERNAL_INSTITUTION",
		ResourceType: "INSTITUTION",
		ResourceID:   draft.ID.String(),
		NewState: map[string]interface{}{
			"name":        draft.Name,
			"external_id": externalID,
			"source":      source,
		},
	}
}

// Hasil pencarian gabungan institusi eksternal (kampus + sekolah).
type ExternalInstitutionSearchResult struct {
	Kampus  *SearchKampusResult  `json:"kampus,omitempty"`
	Sekolah *SearchSekolahResult `json:"sekolah,omitempty"`
}

// Permintaan impor institusi eksternal ke database internal.
type ImportExternalInstitutionRequest struct {
	Source     string `json:"source" binding:"required"` // API_KAMPUS atau API_SEKOLAH
	ExternalID string `json:"external_id" binding:"required"`
}

// Mencari institusi pendidikan pada Public API API Indonesia dengan cache Redis 30 menit.
func (s *Service) SearchExternalInstitutions(ctx context.Context, query, source, province, regency string, page, perPage int) (*ExternalInstitutionSearchResult, error) {
	if s.externalClient == nil {
		return nil, fmt.Errorf("integrasi API Indonesia belum dikonfigurasi pada server")
	}
	if len(strings.TrimSpace(query)) < 2 && strings.TrimSpace(province) == "" && strings.TrimSpace(regency) == "" {
		return nil, fmt.Errorf("kata kunci pencarian minimal 2 karakter atau filter wilayah wajib diisi")
	}
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 200 {
		perPage = 20
	}

	cacheKey := fmt.Sprintf("apiindonesia:institutions:%s:%s:%s:%s:%d:%d", strings.ToLower(source), strings.ToLower(query), province, regency, page, perPage)
	if s.redisClient != nil && s.redisClient.Client != nil {
		if cached, err := s.redisClient.Client.Get(ctx, cacheKey).Result(); err == nil && cached != "" {
			var cachedResult ExternalInstitutionSearchResult
			if err := json.Unmarshal([]byte(cached), &cachedResult); err == nil {
				return &cachedResult, nil
			}
		}
	}

	result := &ExternalInstitutionSearchResult{}
	normalizedSource := strings.ToUpper(strings.TrimSpace(source))

	if normalizedSource == "" || normalizedSource == "KAMPUS" || normalizedSource == "API_KAMPUS" || normalizedSource == "ALL" {
		kampusRes, err := s.externalClient.SearchKampus(ctx, SearchKampusParams{
			Query:    query,
			Province: province,
			Regency:  regency,
			Page:     page,
			PerPage:  perPage,
		})
		if err != nil {
			return nil, err
		}
		result.Kampus = kampusRes
	}

	if normalizedSource == "" || normalizedSource == "SEKOLAH" || normalizedSource == "API_SEKOLAH" || normalizedSource == "ALL" {
		sekolahRes, err := s.externalClient.SearchSekolah(ctx, SearchSekolahParams{
			Query:    query,
			Province: province,
			Regency:  regency,
			Page:     page,
			PerPage:  perPage,
		})
		if err != nil {
			// Jika kampus sudah berhasil, kembalikan kampus saja agar parsial tetap berguna
			if result.Kampus != nil {
				return result, nil
			}
			return nil, err
		}
		result.Sekolah = sekolahRes
	}

	if s.redisClient != nil && s.redisClient.Client != nil {
		if payload, err := json.Marshal(result); err == nil {
			_ = s.redisClient.Client.Set(ctx, cacheKey, payload, 30*time.Minute).Err()
		}
	}

	return result, nil
}

// Mengambil detail institusi eksternal berdasarkan sumber dan ID eksternal.
func (s *Service) GetExternalInstitutionDetail(ctx context.Context, source, externalID string) (interface{}, error) {
	if s.externalClient == nil {
		return nil, fmt.Errorf("integrasi API Indonesia belum dikonfigurasi pada server")
	}
	switch strings.ToUpper(strings.TrimSpace(source)) {
	case "API_KAMPUS", "KAMPUS":
		cleanID := strings.TrimPrefix(externalID, "kampus:")
		return s.externalClient.GetKampusDetail(ctx, cleanID)
	case "API_SEKOLAH", "SEKOLAH":
		cleanID := strings.TrimPrefix(externalID, "sekolah:")
		return s.externalClient.GetSekolahDetail(ctx, cleanID)
	default:
		return nil, fmt.Errorf("sumber institusi eksternal tidak valid (gunakan API_KAMPUS atau API_SEKOLAH)")
	}
}

// Mengimpor satu institusi eksternal dari API Indonesia ke database internal secara idempoten.
func (s *Service) ImportExternalInstitution(ctx context.Context, req *ImportExternalInstitutionRequest) (*Institution, bool, error) {
	if s.externalClient == nil {
		return nil, false, fmt.Errorf("integrasi API Indonesia belum dikonfigurasi pada server")
	}
	source := strings.ToUpper(strings.TrimSpace(req.Source))
	externalID := strings.TrimSpace(req.ExternalID)
	if externalID == "" {
		return nil, false, fmt.Errorf("external_id wajib diisi")
	}

	var normalizedExternalID string
	var draft *Institution

	switch source {
	case "API_KAMPUS", "KAMPUS":
		cleanID := strings.TrimPrefix(externalID, "kampus:")
		detail, err := s.externalClient.GetKampusDetail(ctx, cleanID)
		if err != nil {
			return nil, false, err
		}
		draft = MapKampusToInstitutionDraft(detail)
		normalizedExternalID = "kampus:" + detail.ID
	case "API_SEKOLAH", "SEKOLAH":
		cleanID := strings.TrimPrefix(externalID, "sekolah:")
		detail, err := s.externalClient.GetSekolahDetail(ctx, cleanID)
		if err != nil {
			return nil, false, err
		}
		draft = MapSekolahToInstitutionDraft(detail)
		normalizedExternalID = "sekolah:" + detail.NPSN
	default:
		return nil, false, fmt.Errorf("sumber institusi eksternal tidak valid (gunakan API_KAMPUS atau API_SEKOLAH)")
	}

	// Idempotensi: kembalikan record existing apabila external_id sudah pernah diimpor
	existing, err := s.repo.GetInstitutionByExternalID(ctx, normalizedExternalID)
	if err != nil {
		return nil, false, err
	}
	if existing != nil {
		return existing, false, nil
	}

	now := time.Now().UTC()
	draft.ID = uuid.New()
	draft.ExternalID = &normalizedExternalID
	draft.SyncedAt = &now

	if err := s.repo.CreateInstitution(ctx, draft); err != nil {
		// Tangani race condition duplikasi: baca ulang apabila constraint unique terlanggar
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") || strings.Contains(strings.ToLower(err.Error()), "unique") {
			if retryExisting, retryErr := s.repo.GetInstitutionByExternalID(ctx, normalizedExternalID); retryErr == nil && retryExisting != nil {
				return retryExisting, false, nil
			}
		}
		return nil, false, err
	}

	if s.auditService != nil {
		_ = s.auditService.RecordMutation(ctx, importAuditEntry(draft))
	}

	return draft, true, nil
}
