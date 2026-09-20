package finance

import (
	"context"
	"fmt"
	"math"

	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/people"
	"dcisp/backend/internal/modules/projects"
	"github.com/google/uuid"
)

// ProjectsContributionAdapter menghubungkan modul finance ke data kontribusi final
// proyek tanpa membuat dependensi balik dari projects ke finance.
type ProjectsContributionAdapter struct {
	projectsRepo *projects.Repository
	peopleRepo   *people.Repository
}

// Menginisialisasi adapter sumber kontribusi proyek untuk distribusi bounty.
func NewProjectsContributionAdapter(projectsRepo *projects.Repository, peopleRepo *people.Repository, db *database.PostgresDB) *ProjectsContributionAdapter {
	_ = db
	return &ProjectsContributionAdapter{
		projectsRepo: projectsRepo,
		peopleRepo:   peopleRepo,
	}
}

// Mengambil anggota tim proyek beserta persentase kontribusi final yang telah dikunci supervisor.
func (a *ProjectsContributionAdapter) GetDistributionMembers(ctx context.Context, projectID uuid.UUID) (*DistributionInput, error) {
	project, err := a.projectsRepo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, fmt.Errorf("proyek tidak ditemukan")
	}

	members, err := a.projectsRepo.ListTeamMembers(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// Konversi bounty pool ke Rupiah penuh dengan pembulatan half-up yang deterministik
	pool := int64(math.Round(project.BountyPool))
	if pool < 0 {
		return nil, fmt.Errorf("bounty pool proyek tidak valid")
	}

	input := &DistributionInput{
		ProjectID: project.ID,
		Title:     project.Title,
		Status:    project.Status,
		Pool:      pool,
	}

	for _, m := range members {
		if m.FinalContributionPct == nil {
			return nil, fmt.Errorf("kontribusi final anggota %s belum ditetapkan supervisor (finalisasi kontribusi wajib mencakup seluruh tim)", m.MemberFullName)
		}
		dm := DistributionMember{
			UserID:   m.UserID,
			FullName: m.MemberFullName,
			FinalPct: *m.FinalContributionPct,
			IsLocked: m.IsLocked,
		}
		// Petakan batch anggota untuk akumulasi kas bersama (BRULE-FIN-002)
		if intern, err := a.peopleRepo.GetInternByUserID(ctx, m.UserID); err == nil && intern != nil {
			dm.BatchID = intern.BatchID
			dm.HasBatch = true
		}
		input.Members = append(input.Members, dm)
	}

	return input, nil
}
