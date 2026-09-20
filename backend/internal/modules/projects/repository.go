package projects

import (
	"context"
	"errors"
	"fmt"

	"dcisp/backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *database.PostgresDB
}

// Menginisialisasi instance baru repository projects and tasks management.
func NewRepository(db *database.PostgresDB) *Repository {
	return &Repository{db: db}
}

// ProjectPermission adalah proyeksi izin marketplace proyek milik pengguna dari tabel permissions.
type ProjectPermission struct {
	Action    string
	ScopeType string
}

// Mengambil seluruh izin marketplace proyek milik pengguna untuk evaluasi visibilitas dinamis (BR-001).
func (r *Repository) GetUserProjectPermissions(ctx context.Context, userID uuid.UUID) ([]ProjectPermission, error) {
	rows, err := r.db.Pool.Query(ctx, `
		SELECT p.action, p.scope_type
		FROM permissions p
		JOIN user_roles ur ON ur.role_id = p.role_id
		WHERE ur.user_id = $1 AND p.resource IN ('projects.marketplace', '*')
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengueri izin marketplace proyek: %w", err)
	}
	defer rows.Close()

	var perms []ProjectPermission
	for rows.Next() {
		var perm ProjectPermission
		if err := rows.Scan(&perm.Action, &perm.ScopeType); err == nil {
			perms = append(perms, perm)
		}
	}
	return perms, nil
}

// ============================================================================
// 1. PROJECTS (FR-016)
// ============================================================================

// Menyimpan entitas master proyek baru ke dalam database.
func (r *Repository) CreateProject(ctx context.Context, p *Project) error {
	query := `
		INSERT INTO projects (id, title, description, owner_id, visibility, required_skills, capacity, accepted_count, deadline, bounty_pool, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, p.ID, p.Title, p.Description, p.OwnerID, p.Visibility, p.RequiredSkills, p.Capacity, p.AcceptedCount, p.Deadline, p.BountyPool, p.Status)
	if err != nil {
		return fmt.Errorf("gagal menyimpan data proyek baru: %w", err)
	}
	return nil
}

// Mengambil data proyek berdasarkan ID unik beserta nama pemiliknya.
func (r *Repository) GetProjectByID(ctx context.Context, id uuid.UUID) (*Project, error) {
	query := `
		SELECT p.id, p.title, p.description, p.owner_id, p.visibility, p.required_skills, p.capacity, p.accepted_count, p.deadline, p.bounty_pool, p.status, p.created_at, p.updated_at,
		       u.full_name
		FROM projects p
		JOIN users u ON u.id = p.owner_id
		WHERE p.id = $1
	`
	var p Project
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Title, &p.Description, &p.OwnerID, &p.Visibility, &p.RequiredSkills, &p.Capacity, &p.AcceptedCount, &p.Deadline, &p.BountyPool, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		&p.OwnerFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari proyek berdasarkan id: %w", err)
	}
	return &p, nil
}

// Mengambil proyek dan mengunci baris dalam transaksi SERIALIZABLE untuk alokasi kuota yang aman secara konkurensi (FR-017).
func (r *Repository) LockProjectForUpdateTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*Project, error) {
	query := `
		SELECT id, title, description, owner_id, visibility, required_skills, capacity, accepted_count, deadline, bounty_pool, status, created_at, updated_at
		FROM projects
		WHERE id = $1
		FOR UPDATE
	`
	var p Project
	err := tx.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Title, &p.Description, &p.OwnerID, &p.Visibility, &p.RequiredSkills, &p.Capacity, &p.AcceptedCount, &p.Deadline, &p.BountyPool, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mengunci data proyek dalam transaksi: %w", err)
	}
	return &p, nil
}

// Mengambil daftar proyek bursa dengan filter visibilitas, status, pencarian judul, dan paginasi (FR-016).
func (r *Repository) ListProjects(ctx context.Context, allowedVisibilities []string, status, search string, limit, offset int) ([]Project, int, error) {
	baseQuery := `FROM projects p JOIN users u ON u.id = p.owner_id WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if len(allowedVisibilities) > 0 {
		baseQuery += fmt.Sprintf(" AND p.visibility = ANY($%d)", argIdx)
		args = append(args, allowedVisibilities)
		argIdx++
	}

	if status != "" {
		baseQuery += fmt.Sprintf(" AND p.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	if search != "" {
		baseQuery += fmt.Sprintf(" AND p.title ILIKE $%d", argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total proyek: %w", err)
	}

	selectQuery := fmt.Sprintf(`
		SELECT p.id, p.title, p.description, p.owner_id, p.visibility, p.required_skills, p.capacity, p.accepted_count, p.deadline, p.bounty_pool, p.status, p.created_at, p.updated_at,
		       u.full_name
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar proyek: %w", err)
	}
	defer rows.Close()

	var list []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(
			&p.ID, &p.Title, &p.Description, &p.OwnerID, &p.Visibility, &p.RequiredSkills, &p.Capacity, &p.AcceptedCount, &p.Deadline, &p.BountyPool, &p.Status, &p.CreatedAt, &p.UpdatedAt,
			&p.OwnerFullName,
		); err == nil {
			list = append(list, p)
		}
	}

	return list, total, nil
}

// Memperbarui data informasi master proyek.
func (r *Repository) UpdateProject(ctx context.Context, p *Project) error {
	query := `
		UPDATE projects
		SET title = $2, description = $3, visibility = $4, required_skills = $5, capacity = $6, deadline = $7, bounty_pool = $8, status = $9, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, p.ID, p.Title, p.Description, p.Visibility, p.RequiredSkills, p.Capacity, p.Deadline, p.BountyPool, p.Status)
	if err != nil {
		return fmt.Errorf("gagal memperbarui proyek: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("proyek tidak ditemukan")
	}
	return nil
}

// Menambahkan jumlah peserta yang diterima dalam transaksi atomik saat pendaftaran disetujui.
func (r *Repository) IncrementAcceptedCountTx(ctx context.Context, tx pgx.Tx, projectID uuid.UUID) error {
	query := `
		UPDATE projects
		SET accepted_count = accepted_count + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND accepted_count < capacity
	`
	cmdTag, err := tx.Exec(ctx, query, projectID)
	if err != nil {
		return fmt.Errorf("gagal menambah jumlah peserta diterima: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("kuota proyek telah penuh atau proyek tidak ditemukan")
	}
	return nil
}

// ============================================================================
// 2. PROJECT APPLICATIONS & REGISTRATION (FR-017)
// ============================================================================

// Menyimpan berkas lamaran proyek baru ke dalam database.
func (r *Repository) CreateApplication(ctx context.Context, app *ProjectApplication) error {
	query := `
		INSERT INTO project_applications (id, project_id, user_id, cover_letter, skill_match_percentage, status, applied_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if app.ID == uuid.Nil {
		app.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, app.ID, app.ProjectID, app.UserID, app.CoverLetter, app.SkillMatchPercentage, app.Status)
	if err != nil {
		return fmt.Errorf("gagal mengajukan lamaran proyek: %w", err)
	}
	return nil
}

// Mengambil data lamaran proyek berdasarkan project ID dan user ID pemohon.
func (r *Repository) GetApplication(ctx context.Context, projectID, userID uuid.UUID) (*ProjectApplication, error) {
	query := `
		SELECT id, project_id, user_id, cover_letter, skill_match_percentage, status, applied_at, updated_at
		FROM project_applications
		WHERE project_id = $1 AND user_id = $2
	`
	var app ProjectApplication
	err := r.db.Pool.QueryRow(ctx, query, projectID, userID).Scan(
		&app.ID, &app.ProjectID, &app.UserID, &app.CoverLetter, &app.SkillMatchPercentage, &app.Status, &app.AppliedAt, &app.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari lamaran proyek: %w", err)
	}
	return &app, nil
}

// Mengambil berkas lamaran proyek berdasarkan ID unik lamaran.
func (r *Repository) GetApplicationByID(ctx context.Context, id uuid.UUID) (*ProjectApplication, error) {
	query := `
		SELECT a.id, a.project_id, a.user_id, a.cover_letter, a.skill_match_percentage, a.status, a.applied_at, a.updated_at,
		       u.full_name, u.email
		FROM project_applications a
		JOIN users u ON u.id = a.user_id
		WHERE a.id = $1
	`
	var app ProjectApplication
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&app.ID, &app.ProjectID, &app.UserID, &app.CoverLetter, &app.SkillMatchPercentage, &app.Status, &app.AppliedAt, &app.UpdatedAt,
		&app.ApplicantFullName, &app.ApplicantEmail,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari lamaran berdasarkan id: %w", err)
	}
	return &app, nil
}

// Mengambil seluruh daftar pelamar pada proyek tertentu.
func (r *Repository) ListApplicationsByProject(ctx context.Context, projectID uuid.UUID) ([]ProjectApplication, error) {
	query := `
		SELECT a.id, a.project_id, a.user_id, a.cover_letter, a.skill_match_percentage, a.status, a.applied_at, a.updated_at,
		       u.full_name, u.email
		FROM project_applications a
		JOIN users u ON u.id = a.user_id
		WHERE a.project_id = $1
		ORDER BY a.applied_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar pelamar proyek: %w", err)
	}
	defer rows.Close()

	var list []ProjectApplication
	for rows.Next() {
		var a ProjectApplication
		if err := rows.Scan(
			&a.ID, &a.ProjectID, &a.UserID, &a.CoverLetter, &a.SkillMatchPercentage, &a.Status, &a.AppliedAt, &a.UpdatedAt,
			&a.ApplicantFullName, &a.ApplicantEmail,
		); err == nil {
			list = append(list, a)
		}
	}
	return list, nil
}

// Memperbarui status lamaran proyek dalam transaksi.
func (r *Repository) UpdateApplicationStatusTx(ctx context.Context, tx pgx.Tx, appID uuid.UUID, status string) error {
	query := `
		UPDATE project_applications
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := tx.Exec(ctx, query, appID, status)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status lamaran proyek: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("lamaran tidak ditemukan")
	}
	return nil
}

// ============================================================================
// 3. PROJECT TEAMS & CONTRIBUTION (FR-019, FR-024)
// ============================================================================

// Menambahkan anggota baru ke dalam tim proyek dalam transaksi database.
func (r *Repository) AddTeamMemberTx(ctx context.Context, tx pgx.Tx, m *ProjectTeamMember) error {
	query := `
		INSERT INTO project_teams (id, project_id, user_id, project_role, responsibility, planned_contribution_pct, is_locked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (project_id, user_id) DO UPDATE
		SET project_role = EXCLUDED.project_role, responsibility = EXCLUDED.responsibility
	`
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, query, m.ID, m.ProjectID, m.UserID, m.ProjectRole, m.Responsibility, m.PlannedContributionPct, m.IsLocked)
	if err != nil {
		return fmt.Errorf("gagal menambahkan anggota tim proyek: %w", err)
	}
	return nil
}

// Menambahkan anggota baru ke dalam tim proyek secara langsung.
func (r *Repository) AddTeamMember(ctx context.Context, m *ProjectTeamMember) error {
	query := `
		INSERT INTO project_teams (id, project_id, user_id, project_role, responsibility, planned_contribution_pct, is_locked, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (project_id, user_id) DO UPDATE
		SET project_role = EXCLUDED.project_role, responsibility = EXCLUDED.responsibility
	`
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, m.ID, m.ProjectID, m.UserID, m.ProjectRole, m.Responsibility, m.PlannedContributionPct, m.IsLocked)
	if err != nil {
		return fmt.Errorf("gagal menambahkan anggota tim proyek: %w", err)
	}
	return nil
}

// Mengambil rekaman data anggota tim proyek berdasarkan project ID dan user ID.
func (r *Repository) GetTeamMember(ctx context.Context, projectID, userID uuid.UUID) (*ProjectTeamMember, error) {
	query := `
		SELECT id, project_id, user_id, project_role, responsibility, planned_contribution_pct, actual_contribution_pct, final_contribution_pct, is_locked, created_at, updated_at
		FROM project_teams
		WHERE project_id = $1 AND user_id = $2
	`
	var m ProjectTeamMember
	err := r.db.Pool.QueryRow(ctx, query, projectID, userID).Scan(
		&m.ID, &m.ProjectID, &m.UserID, &m.ProjectRole, &m.Responsibility, &m.PlannedContributionPct, &m.ActualContributionPct, &m.FinalContributionPct, &m.IsLocked, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari anggota tim: %w", err)
	}
	return &m, nil
}

// Mengambil seluruh daftar anggota tim pada proyek tertentu.
func (r *Repository) ListTeamMembers(ctx context.Context, projectID uuid.UUID) ([]ProjectTeamMember, error) {
	query := `
		SELECT pt.id, pt.project_id, pt.user_id, pt.project_role, pt.responsibility, pt.planned_contribution_pct, pt.actual_contribution_pct, pt.final_contribution_pct, pt.is_locked, pt.created_at, pt.updated_at,
		       u.full_name, u.email, i.batch_id
		FROM project_teams pt
		JOIN users u ON u.id = pt.user_id
		LEFT JOIN interns i ON i.user_id = pt.user_id
		WHERE pt.project_id = $1
		ORDER BY pt.created_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar anggota tim: %w", err)
	}
	defer rows.Close()

	var list []ProjectTeamMember
	for rows.Next() {
		var m ProjectTeamMember
		if err := rows.Scan(
			&m.ID, &m.ProjectID, &m.UserID, &m.ProjectRole, &m.Responsibility, &m.PlannedContributionPct, &m.ActualContributionPct, &m.FinalContributionPct, &m.IsLocked, &m.CreatedAt, &m.UpdatedAt,
			&m.MemberFullName, &m.MemberEmail, &m.BatchID,
		); err == nil {
			list = append(list, m)
		}
	}
	return list, nil
}

// Memperbarui nilai Planned Contribution persentase pada anggota tim dalam transaksi.
func (r *Repository) UpdatePlannedContributionTx(ctx context.Context, tx pgx.Tx, projectID, userID uuid.UUID, plannedPct float64) error {
	query := `
		UPDATE project_teams
		SET planned_contribution_pct = $3, updated_at = CURRENT_TIMESTAMP
		WHERE project_id = $1 AND user_id = $2 AND is_locked = FALSE
	`
	cmdTag, err := tx.Exec(ctx, query, projectID, userID, plannedPct)
	if err != nil {
		return fmt.Errorf("gagal memperbarui planned contribution: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("anggota tim tidak ditemukan atau kontribusi telah terkunci")
	}
	return nil
}

// Memperbarui nilai Actual Contribution persentase hasil komputasi sistem pada anggota tim.
func (r *Repository) UpdateActualContribution(ctx context.Context, projectID, userID uuid.UUID, actualPct float64) error {
	query := `
		UPDATE project_teams
		SET actual_contribution_pct = $3, updated_at = CURRENT_TIMESTAMP
		WHERE project_id = $1 AND user_id = $2
	`
	_, err := r.db.Pool.Exec(ctx, query, projectID, userID, actualPct)
	if err != nil {
		return fmt.Errorf("gagal memperbarui actual contribution: %w", err)
	}
	return nil
}

// Mengesahkan dan mengunci Final Contribution persentase anggota tim dalam transaksi (BR-016).
func (r *Repository) LockFinalContributionTx(ctx context.Context, tx pgx.Tx, projectID, userID uuid.UUID, finalPct float64) error {
	query := `
		UPDATE project_teams
		SET final_contribution_pct = $3, is_locked = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE project_id = $1 AND user_id = $2
	`
	cmdTag, err := tx.Exec(ctx, query, projectID, userID, finalPct)
	if err != nil {
		return fmt.Errorf("gagal mengunci final contribution: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("anggota tim proyek tidak ditemukan")
	}
	return nil
}

// ============================================================================
// 4. MILESTONES (FR-020)
// ============================================================================

// Membuat milestone target baru pada proyek.
func (r *Repository) CreateMilestone(ctx context.Context, m *Milestone) error {
	query := `
		INSERT INTO milestones (id, project_id, title, description, deadline, weight_pct, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, m.ID, m.ProjectID, m.Title, m.Description, m.Deadline, m.WeightPct, m.Status)
	if err != nil {
		return fmt.Errorf("gagal membuat milestone baru: %w", err)
	}
	return nil
}

// Mengambil milestone proyek berdasarkan ID unik.
func (r *Repository) GetMilestoneByID(ctx context.Context, id uuid.UUID) (*Milestone, error) {
	query := `
		SELECT id, project_id, title, description, deadline, weight_pct, status, created_at, updated_at
		FROM milestones
		WHERE id = $1
	`
	var m Milestone
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&m.ID, &m.ProjectID, &m.Title, &m.Description, &m.Deadline, &m.WeightPct, &m.Status, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari milestone berdasarkan id: %w", err)
	}
	return &m, nil
}

// Mengambil daftar seluruh milestone pada proyek tertentu diurutkan berdasarkan tenggat waktu.
func (r *Repository) ListMilestones(ctx context.Context, projectID uuid.UUID) ([]Milestone, error) {
	query := `
		SELECT id, project_id, title, description, deadline, weight_pct, status, created_at, updated_at
		FROM milestones
		WHERE project_id = $1
		ORDER BY deadline ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar milestone: %w", err)
	}
	defer rows.Close()

	var list []Milestone
	for rows.Next() {
		var m Milestone
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.Title, &m.Description, &m.Deadline, &m.WeightPct, &m.Status, &m.CreatedAt, &m.UpdatedAt); err == nil {
			list = append(list, m)
		}
	}
	return list, nil
}

// Menghitung total bobot persentase seluruh milestone pada suatu proyek.
func (r *Repository) GetTotalMilestoneWeight(ctx context.Context, projectID uuid.UUID) (float64, error) {
	query := `SELECT COALESCE(SUM(weight_pct), 0.00) FROM milestones WHERE project_id = $1`
	var total float64
	err := r.db.Pool.QueryRow(ctx, query, projectID).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("gagal menghitung total bobot milestone: %w", err)
	}
	return total, nil
}

// Memperbarui data informasi milestone proyek.
func (r *Repository) UpdateMilestone(ctx context.Context, m *Milestone) error {
	query := `
		UPDATE milestones
		SET title = $2, description = $3, deadline = $4, weight_pct = $5, status = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, m.ID, m.Title, m.Description, m.Deadline, m.WeightPct, m.Status)
	if err != nil {
		return fmt.Errorf("gagal memperbarui milestone: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("milestone tidak ditemukan")
	}
	return nil
}

// ============================================================================
// 5. TASKS KANBAN (FR-021)
// ============================================================================

// Menyimpan tugas baru ke dalam papan kanban proyek.
func (r *Repository) CreateTask(ctx context.Context, t *Task) error {
	query := `
		INSERT INTO tasks (id, project_id, milestone_id, assignee_id, title, description, estimated_hours, difficulty_weight, priority, deadline, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, t.ID, t.ProjectID, t.MilestoneID, t.AssigneeID, t.Title, t.Description, t.EstimatedHours, t.DifficultyWeight, t.Priority, t.Deadline, t.Status)
	if err != nil {
		return fmt.Errorf("gagal membuat tugas baru: %w", err)
	}
	return nil
}

// Mengambil data tugas kanban berdasarkan ID unik.
func (r *Repository) GetTaskByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	query := `
		SELECT t.id, t.project_id, t.milestone_id, t.assignee_id, t.title, t.description, t.estimated_hours, t.difficulty_weight, t.priority, t.deadline, t.status, t.created_at, t.updated_at,
		       u.full_name, m.title
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		LEFT JOIN milestones m ON m.id = t.milestone_id
		WHERE t.id = $1
	`
	var t Task
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.ProjectID, &t.MilestoneID, &t.AssigneeID, &t.Title, &t.Description, &t.EstimatedHours, &t.DifficultyWeight, &t.Priority, &t.Deadline, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		&t.AssigneeFullName, &t.MilestoneTitle,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari tugas berdasarkan id: %w", err)
	}
	return &t, nil
}

// Mengambil daftar tugas kanban dengan filter milestone, penugasan, status, dan paginasi.
func (r *Repository) ListTasks(ctx context.Context, projectID uuid.UUID, milestoneID, assigneeID *uuid.UUID, status string, limit, offset int) ([]Task, int, error) {
	baseQuery := `
		FROM tasks t
		LEFT JOIN users u ON u.id = t.assignee_id
		LEFT JOIN milestones m ON m.id = t.milestone_id
		WHERE t.project_id = $1
	`
	args := []interface{}{projectID}
	argIdx := 2

	if milestoneID != nil {
		baseQuery += fmt.Sprintf(" AND t.milestone_id = $%d", argIdx)
		args = append(args, *milestoneID)
		argIdx++
	}

	if assigneeID != nil {
		baseQuery += fmt.Sprintf(" AND t.assignee_id = $%d", argIdx)
		args = append(args, *assigneeID)
		argIdx++
	}

	if status != "" {
		baseQuery += fmt.Sprintf(" AND t.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total tugas: %w", err)
	}

	selectQuery := fmt.Sprintf(`
		SELECT t.id, t.project_id, t.milestone_id, t.assignee_id, t.title, t.description, t.estimated_hours, t.difficulty_weight, t.priority, t.deadline, t.status, t.created_at, t.updated_at,
		       u.full_name, m.title
		%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar tugas: %w", err)
	}
	defer rows.Close()

	var list []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.MilestoneID, &t.AssigneeID, &t.Title, &t.Description, &t.EstimatedHours, &t.DifficultyWeight, &t.Priority, &t.Deadline, &t.Status, &t.CreatedAt, &t.UpdatedAt,
			&t.AssigneeFullName, &t.MilestoneTitle,
		); err == nil {
			list = append(list, t)
		}
	}

	return list, total, nil
}

// Memperbarui status siklus tugas pada papan kanban dalam transaksi database.
func (r *Repository) UpdateTaskStatusTx(ctx context.Context, tx pgx.Tx, taskID uuid.UUID, status string) error {
	query := `
		UPDATE tasks
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := tx.Exec(ctx, query, taskID, status)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status tugas: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("tugas tidak ditemukan")
	}
	return nil
}

// Mengambil seluruh tugas yang telah berstatus COMPLETED pada suatu proyek untuk komputasi Actual Contribution.
func (r *Repository) GetCompletedTasksByProject(ctx context.Context, projectID uuid.UUID) ([]Task, error) {
	query := `
		SELECT id, project_id, milestone_id, assignee_id, title, description, estimated_hours, difficulty_weight, priority, deadline, status, created_at, updated_at
		FROM tasks
		WHERE project_id = $1 AND status = 'COMPLETED' AND assignee_id IS NOT NULL
	`
	rows, err := r.db.Pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil tugas selesai proyek: %w", err)
	}
	defer rows.Close()

	var list []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.MilestoneID, &t.AssigneeID, &t.Title, &t.Description, &t.EstimatedHours, &t.DifficultyWeight, &t.Priority, &t.Deadline, &t.Status, &t.CreatedAt, &t.UpdatedAt); err == nil {
			list = append(list, t)
		}
	}
	return list, nil
}

// ============================================================================
// 6. WORK REPORTS & EVIDENCE (FR-022, FR-023)
// ============================================================================

// Menyimpan laporan kemajuan pekerjaan tugas baru dalam transaksi.
func (r *Repository) CreateWorkReportTx(ctx context.Context, tx pgx.Tx, rep *WorkReport) error {
	query := `
		INSERT INTO work_reports (id, task_id, user_id, date, progress_percentage, what_i_did, evidence_type, evidence_url_or_key, problems, next_actions, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if rep.ID == uuid.Nil {
		rep.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, query, rep.ID, rep.TaskID, rep.UserID, rep.Date, rep.ProgressPercentage, rep.WhatIDid, rep.EvidenceType, rep.EvidenceURLOrKey, rep.Problems, rep.NextActions, rep.Status)
	if err != nil {
		return fmt.Errorf("gagal membuat laporan pekerjaan: %w", err)
	}
	return nil
}

// Mengambil laporan pekerjaan berdasarkan ID unik.
func (r *Repository) GetWorkReportByID(ctx context.Context, id uuid.UUID) (*WorkReport, error) {
	query := `
		SELECT r.id, r.task_id, r.user_id, r.date, r.progress_percentage, r.what_i_did, r.evidence_type, r.evidence_url_or_key, r.problems, r.next_actions, r.status, r.created_at, r.updated_at,
		       t.title, u.full_name
		FROM work_reports r
		JOIN tasks t ON t.id = r.task_id
		JOIN users u ON u.id = r.user_id
		WHERE r.id = $1
	`
	var rep WorkReport
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&rep.ID, &rep.TaskID, &rep.UserID, &rep.Date, &rep.ProgressPercentage, &rep.WhatIDid, &rep.EvidenceType, &rep.EvidenceURLOrKey, &rep.Problems, &rep.NextActions, &rep.Status, &rep.CreatedAt, &rep.UpdatedAt,
		&rep.TaskTitle, &rep.AuthorFullName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari laporan pekerjaan berdasarkan id: %w", err)
	}
	return &rep, nil
}

// Memperbarui status laporan pekerjaan hasil telaah reviewer dalam transaksi.
func (r *Repository) UpdateWorkReportStatusTx(ctx context.Context, tx pgx.Tx, reportID uuid.UUID, status string) error {
	query := `
		UPDATE work_reports
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := tx.Exec(ctx, query, reportID, status)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status laporan pekerjaan: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("laporan pekerjaan tidak ditemukan")
	}
	return nil
}

// Menyimpan rekaman artefak bukti deliverable baru ke dalam tabel evidence dalam transaksi.
func (r *Repository) CreateEvidenceTx(ctx context.Context, tx pgx.Tx, ev *Evidence) error {
	query := `
		INSERT INTO evidence (id, submission_id, file_metadata_id, evidence_type, external_url, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, query, ev.ID, ev.SubmissionID, ev.FileMetadataID, ev.EvidenceType, ev.ExternalURL, ev.Description)
	if err != nil {
		return fmt.Errorf("gagal menyimpan artefak bukti deliverable: %w", err)
	}
	return nil
}

// Menyimpan rekaman artefak bukti deliverable baru ke dalam tabel evidence.
func (r *Repository) CreateEvidence(ctx context.Context, ev *Evidence) error {
	query := `
		INSERT INTO evidence (id, submission_id, file_metadata_id, evidence_type, external_url, description, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
	`
	if ev.ID == uuid.Nil {
		ev.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, ev.ID, ev.SubmissionID, ev.FileMetadataID, ev.EvidenceType, ev.ExternalURL, ev.Description)
	if err != nil {
		return fmt.Errorf("gagal menyimpan artefak bukti deliverable: %w", err)
	}
	return nil
}

// Mengambil daftar seluruh artefak bukti deliverable untuk laporan tugas tertentu.
func (r *Repository) GetEvidenceBySubmission(ctx context.Context, submissionID uuid.UUID) ([]Evidence, error) {
	query := `
		SELECT id, submission_id, file_metadata_id, evidence_type, external_url, description, created_at
		FROM evidence
		WHERE submission_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, submissionID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil bukti deliverable: %w", err)
	}
	defer rows.Close()

	var list []Evidence
	for rows.Next() {
		var ev Evidence
		if err := rows.Scan(&ev.ID, &ev.SubmissionID, &ev.FileMetadataID, &ev.EvidenceType, &ev.ExternalURL, &ev.Description, &ev.CreatedAt); err == nil {
			list = append(list, ev)
		}
	}
	return list, nil
}
