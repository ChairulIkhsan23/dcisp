package finance

import (
	"context"

	"dcisp/backend/internal/modules/performance"
	"github.com/google/uuid"
)

// PerformanceXPAdapter menghubungkan modul finance ke mesin XP untuk komponen
// reward BONUS_XP tanpa membuat dependensi balik dari performance ke finance.
// Bonus dicatat pada skema INTERNSHIP_XP melalui trigger CHECK_IN_ON_TIME yang
// selalu tersedia di seed, dengan nominal override dan referensi idempoten unik.
type PerformanceXPAdapter struct {
	performanceSvc *performance.Service
}

// Menginisialisasi adapter mutasi XP bonus untuk reward rank.
func NewPerformanceXPAdapter(svc *performance.Service) *PerformanceXPAdapter {
	return &PerformanceXPAdapter{performanceSvc: svc}
}

// Mencatat mutasi XP bonus dari komponen reward secara idempoten berdasarkan referensi event.
func (a *PerformanceXPAdapter) RecordBonusXP(ctx context.Context, userID uuid.UUID, points int, referenceEvent string) error {
	if a.performanceSvc == nil {
		return nil
	}
	_, err := a.performanceSvc.RecordXPMutation(ctx, &performance.RecordXPMutationRequest{
		UserID:         userID,
		EventTrigger:   "CHECK_IN_ON_TIME",
		PointsOverride: &points,
		ReferenceEvent: referenceEvent,
	})
	return err
}
