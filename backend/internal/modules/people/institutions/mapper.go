package institutions

import (
	"strings"

	"dcisp/backend/internal/integrations/apiindonesia"
	"dcisp/backend/internal/modules/people"
)

// Memetakan respons kampus eksternal menjadi draf institusi internal tanpa menyimpan kredensial eksternal.
func MapKampusToInstitutionDraft(ext *apiindonesia.Kampus) *people.Institution {
	inst := &people.Institution{
		Name:   ext.Name,
		Source: strPtr(people.InstitutionSourceAPIKampus),
	}
	externalID := "kampus:" + ext.ID
	inst.ExternalID = &externalID
	parts := []string{}
	if ext.Address != nil && strings.TrimSpace(*ext.Address) != "" {
		parts = append(parts, strings.TrimSpace(*ext.Address))
	}
	if ext.RegencyName != nil && strings.TrimSpace(*ext.RegencyName) != "" {
		parts = append(parts, strings.TrimSpace(*ext.RegencyName))
	}
	if ext.ProvinceName != nil && strings.TrimSpace(*ext.ProvinceName) != "" {
		parts = append(parts, strings.TrimSpace(*ext.ProvinceName))
	}
	if len(parts) > 0 {
		joined := strings.Join(parts, ", ")
		inst.Address = &joined
	}
	if ext.Phone != nil && strings.TrimSpace(*ext.Phone) != "" {
		inst.Phone = ext.Phone
	}
	if ext.Email != nil && strings.TrimSpace(*ext.Email) != "" {
		inst.Email = ext.Email
	}
	return inst
}

// Memetakan respons sekolah eksternal menjadi draf institusi internal.
func MapSekolahToInstitutionDraft(ext *apiindonesia.Sekolah) *people.Institution {
	inst := &people.Institution{
		Name:   ext.Name,
		Source: strPtr(people.InstitutionSourceAPISekolah),
	}
	externalID := "sekolah:" + ext.NPSN
	inst.ExternalID = &externalID
	if ext.Address != nil {
		inst.Address = ext.Address
	}
	if ext.Phone != nil {
		inst.Phone = ext.Phone
	}
	if ext.Email != nil {
		inst.Email = ext.Email
	}
	return inst
}

// Mengembalikan pointer string untuk nilai opsional.
func strPtr(s string) *string {
	copied := s
	return &copied
}
