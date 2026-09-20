package people

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

// Menginisialisasi instance baru repository people and lifecycle management.
func NewRepository(db *database.PostgresDB) *Repository {
	return &Repository{db: db}
}

// ============================================================================
// 1. INSTITUTIONS
// ============================================================================

// Menyimpan entitas institusi pendidikan mitra baru ke dalam database.
func (r *Repository) CreateInstitution(ctx context.Context, inst *Institution) error {
	query := `
		INSERT INTO institutions (id, name, address, contact_person, email, phone, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if inst.ID == uuid.Nil {
		inst.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, inst.ID, inst.Name, inst.Address, inst.ContactPerson, inst.Email, inst.Phone)
	if err != nil {
		return fmt.Errorf("gagal menyimpan data institusi: %w", err)
	}
	return nil
}

// Mengambil data institusi pendidikan berdasarkan identitas unik UUID.
func (r *Repository) GetInstitutionByID(ctx context.Context, id uuid.UUID) (*Institution, error) {
	query := `
		SELECT id, name, address, contact_person, email, phone, created_at, updated_at
		FROM institutions
		WHERE id = $1
	`
	var inst Institution
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&inst.ID, &inst.Name, &inst.Address, &inst.ContactPerson, &inst.Email, &inst.Phone, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari institusi berdasarkan id: %w", err)
	}
	return &inst, nil
}

// Mengambil seluruh daftar institusi pendidikan mitra dari database.
func (r *Repository) ListInstitutions(ctx context.Context) ([]Institution, error) {
	query := `
		SELECT id, name, address, contact_person, email, phone, created_at, updated_at
		FROM institutions
		ORDER BY name ASC
	`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar institusi: %w", err)
	}
	defer rows.Close()

	var list []Institution
	for rows.Next() {
		var inst Institution
		if err := rows.Scan(&inst.ID, &inst.Name, &inst.Address, &inst.ContactPerson, &inst.Email, &inst.Phone, &inst.CreatedAt, &inst.UpdatedAt); err == nil {
			list = append(list, inst)
		}
	}
	return list, nil
}

// Memperbarui informasi institusi pendidikan mitra yang sudah ada.
func (r *Repository) UpdateInstitution(ctx context.Context, inst *Institution) error {
	query := `
		UPDATE institutions
		SET name = $2, address = $3, contact_person = $4, email = $5, phone = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, inst.ID, inst.Name, inst.Address, inst.ContactPerson, inst.Email, inst.Phone)
	if err != nil {
		return fmt.Errorf("gagal memperbarui institusi: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("institusi tidak ditemukan")
	}
	return nil
}

// Menghitung jumlah peserta magang aktif maupun riwayat yang terhubung ke institusi tertentu.
func (r *Repository) CountInternsByInstitution(ctx context.Context, instID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM interns WHERE institution_id = $1`
	err := r.db.Pool.QueryRow(ctx, query, instID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("gagal menghitung peserta pada institusi: %w", err)
	}
	return count, nil
}

// Menghapus data institusi dari database berdasarkan identitas unik.
func (r *Repository) DeleteInstitution(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM institutions WHERE id = $1`
	cmdTag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus institusi: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("institusi tidak ditemukan")
	}
	return nil
}

// ============================================================================
// 2. BATCHES & BATCH FUNDS
// ============================================================================

// Membuat batch baru dan menginisialisasi Batch Fund secara atomik dalam satu transaksi.
func (r *Repository) CreateBatchWithFundTx(ctx context.Context, tx pgx.Tx, b *Batch) (*BatchFund, error) {
	queryBatch := `
		INSERT INTO batches (id, batch_code, name, start_date, end_date, quota, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, queryBatch, b.ID, b.BatchCode, b.Name, b.StartDate, b.EndDate, b.Quota, b.Status)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat batch baru dalam transaksi: %w", err)
	}

	fundID := uuid.New()
	queryFund := `
		INSERT INTO batch_funds (id, batch_id, total_accumulated, current_balance, created_at, updated_at)
		VALUES ($1, $2, 0.00, 0.00, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err = tx.Exec(ctx, queryFund, fundID, b.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal menginisialisasi batch fund: %w", err)
	}

	return &BatchFund{
		ID:               fundID,
		BatchID:          b.ID,
		TotalAccumulated: 0.00,
		CurrentBalance:   0.00,
	}, nil
}

// Mengambil data kohort batch dari database berdasarkan ID unik.
func (r *Repository) GetBatchByID(ctx context.Context, id uuid.UUID) (*Batch, error) {
	query := `
		SELECT id, batch_code, name, start_date, end_date, quota, status, created_at, updated_at
		FROM batches
		WHERE id = $1
	`
	var b Batch
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(&b.ID, &b.BatchCode, &b.Name, &b.StartDate, &b.EndDate, &b.Quota, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari batch berdasarkan id: %w", err)
	}
	return &b, nil
}

// Mengambil data kohort batch dari database dengan penguncian baris eksklusif untuk konkurensi alokasi kuota.
func (r *Repository) LockBatchForUpdateTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*Batch, error) {
	query := `
		SELECT id, batch_code, name, start_date, end_date, quota, status, created_at, updated_at
		FROM batches
		WHERE id = $1
		FOR UPDATE
	`
	var b Batch
	err := tx.QueryRow(ctx, query, id).Scan(&b.ID, &b.BatchCode, &b.Name, &b.StartDate, &b.EndDate, &b.Quota, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mengunci data batch untuk pembaruan: %w", err)
	}
	return &b, nil
}

// Mengambil data kohort batch berdasarkan kode unik batch.
func (r *Repository) GetBatchByCode(ctx context.Context, code string) (*Batch, error) {
	query := `
		SELECT id, batch_code, name, start_date, end_date, quota, status, created_at, updated_at
		FROM batches
		WHERE batch_code = $1
	`
	var b Batch
	err := r.db.Pool.QueryRow(ctx, query, code).Scan(&b.ID, &b.BatchCode, &b.Name, &b.StartDate, &b.EndDate, &b.Quota, &b.Status, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari batch berdasarkan kode: %w", err)
	}
	return &b, nil
}

// Mengambil daftar seluruh batch dengan opsi filter status.
func (r *Repository) ListBatches(ctx context.Context, status string) ([]Batch, error) {
	query := `
		SELECT id, batch_code, name, start_date, end_date, quota, status, created_at, updated_at
		FROM batches
	`
	var args []interface{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY start_date DESC"

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar batch: %w", err)
	}
	defer rows.Close()

	var list []Batch
	for rows.Next() {
		var b Batch
		if err := rows.Scan(&b.ID, &b.BatchCode, &b.Name, &b.StartDate, &b.EndDate, &b.Quota, &b.Status, &b.CreatedAt, &b.UpdatedAt); err == nil {
			list = append(list, b)
		}
	}
	return list, nil
}

// Memperbarui data informasi kohort batch pada database.
func (r *Repository) UpdateBatch(ctx context.Context, b *Batch) error {
	query := `
		UPDATE batches
		SET name = $2, start_date = $3, end_date = $4, quota = $5, status = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, b.ID, b.Name, b.StartDate, b.EndDate, b.Quota, b.Status)
	if err != nil {
		return fmt.Errorf("gagal memperbarui batch: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("batch tidak ditemukan")
	}
	return nil
}

// Menghitung jumlah peserta magang terdaftar yang telah diterima (memakai kuota) pada batch tertentu.
func (r *Repository) CountInternsInBatch(ctx context.Context, batchID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM interns WHERE batch_id = $1 AND status IN ('ONBOARDING', 'ACTIVE', 'ON_LEAVE', 'SUSPENDED', 'GRADUATED')`
	var count int
	err := r.db.Pool.QueryRow(ctx, query, batchID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("gagal menghitung peserta pada batch: %w", err)
	}
	return count, nil
}

// Mengambil data kas bersama (Batch Fund) berdasarkan identitas unik batch.
func (r *Repository) GetBatchFundByBatchID(ctx context.Context, batchID uuid.UUID) (*BatchFund, error) {
	query := `
		SELECT id, batch_id, total_accumulated, current_balance, created_at, updated_at
		FROM batch_funds
		WHERE batch_id = $1
	`
	var f BatchFund
	err := r.db.Pool.QueryRow(ctx, query, batchID).Scan(&f.ID, &f.BatchID, &f.TotalAccumulated, &f.CurrentBalance, &f.CreatedAt, &f.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari kas batch: %w", err)
	}
	return &f, nil
}

// ============================================================================
// 3. INTERNS & STATE MACHINE
// ============================================================================

// Menyimpan entitas peserta magang baru ke dalam tabel interns.
func (r *Repository) CreateIntern(ctx context.Context, intern *Intern) error {
	query := `
		INSERT INTO interns (id, user_id, batch_id, institution_id, id_number, mentor_id, status, join_date, end_date, current_rank_id, internship_xp, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if intern.ID == uuid.Nil {
		intern.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, intern.ID, intern.UserID, intern.BatchID, intern.InstitutionID, intern.IDNumber, intern.MentorID, intern.Status, intern.JoinDate, intern.EndDate, intern.CurrentRankID, intern.InternshipXP)
	if err != nil {
		return fmt.Errorf("gagal mendaftarkan peserta magang baru: %w", err)
	}
	return nil
}

// Mengambil data lengkap peserta magang berdasarkan ID profil magang.
func (r *Repository) GetInternByID(ctx context.Context, id uuid.UUID) (*Intern, error) {
	query := `
		SELECT i.id, i.user_id, i.batch_id, i.institution_id, i.id_number, i.mentor_id, i.status, i.join_date, i.end_date, i.current_rank_id, i.internship_xp, i.created_at, i.updated_at,
		       u.full_name, u.email, b.name, inst.name, m.full_name, rk.name
		FROM interns i
		JOIN users u ON u.id = i.user_id
		JOIN batches b ON b.id = i.batch_id
		LEFT JOIN institutions inst ON inst.id = i.institution_id
		LEFT JOIN users m ON m.id = i.mentor_id
		LEFT JOIN ranks rk ON rk.id = i.current_rank_id
		WHERE i.id = $1
	`
	var intern Intern
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&intern.ID, &intern.UserID, &intern.BatchID, &intern.InstitutionID, &intern.IDNumber, &intern.MentorID, &intern.Status, &intern.JoinDate, &intern.EndDate, &intern.CurrentRankID, &intern.InternshipXP, &intern.CreatedAt, &intern.UpdatedAt,
		&intern.UserFullName, &intern.UserEmail, &intern.BatchName, &intern.InstitutionName, &intern.MentorFullName, &intern.RankName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari data peserta magang berdasarkan id: %w", err)
	}
	return &intern, nil
}

// Mengambil data lengkap peserta magang berdasarkan user ID akun.
func (r *Repository) GetInternByUserID(ctx context.Context, userID uuid.UUID) (*Intern, error) {
	query := `
		SELECT i.id, i.user_id, i.batch_id, i.institution_id, i.id_number, i.mentor_id, i.status, i.join_date, i.end_date, i.current_rank_id, i.internship_xp, i.created_at, i.updated_at,
		       u.full_name, u.email, b.name, inst.name, m.full_name, rk.name
		FROM interns i
		JOIN users u ON u.id = i.user_id
		JOIN batches b ON b.id = i.batch_id
		LEFT JOIN institutions inst ON inst.id = i.institution_id
		LEFT JOIN users m ON m.id = i.mentor_id
		LEFT JOIN ranks rk ON rk.id = i.current_rank_id
		WHERE i.user_id = $1
	`
	var intern Intern
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&intern.ID, &intern.UserID, &intern.BatchID, &intern.InstitutionID, &intern.IDNumber, &intern.MentorID, &intern.Status, &intern.JoinDate, &intern.EndDate, &intern.CurrentRankID, &intern.InternshipXP, &intern.CreatedAt, &intern.UpdatedAt,
		&intern.UserFullName, &intern.UserEmail, &intern.BatchName, &intern.InstitutionName, &intern.MentorFullName, &intern.RankName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari peserta magang berdasarkan user_id: %w", err)
	}
	return &intern, nil
}

// Memperbarui status siklus hidup peserta magang dalam sebuah transaksi database.
func (r *Repository) UpdateInternStatusTx(ctx context.Context, tx pgx.Tx, internID uuid.UUID, newStatus string) error {
	query := `
		UPDATE interns
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := tx.Exec(ctx, query, internID, newStatus)
	if err != nil {
		return fmt.Errorf("gagal memperbarui status peserta magang: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("peserta magang tidak ditemukan")
	}
	return nil
}

// Memperbarui penugasan pembimbing (mentor) bagi peserta magang.
func (r *Repository) UpdateInternMentor(ctx context.Context, internID uuid.UUID, mentorID *uuid.UUID) error {
	query := `
		UPDATE interns
		SET mentor_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, internID, mentorID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui pembimbing peserta: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("peserta magang tidak ditemukan")
	}
	return nil
}

// Memperbarui penugasan kohort batch bagi peserta magang dalam transaksi.
func (r *Repository) UpdateInternBatchTx(ctx context.Context, tx pgx.Tx, internID, newBatchID uuid.UUID) error {
	query := `
		UPDATE interns
		SET batch_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := tx.Exec(ctx, query, internID, newBatchID)
	if err != nil {
		return fmt.Errorf("gagal memperbarui batch peserta magang: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("peserta magang tidak ditemukan")
	}
	return nil
}

// Mengambil daftar peserta magang berdasarkan kriteria pencarian dan paginasi.
func (r *Repository) ListInterns(ctx context.Context, search string, batchID, institutionID *uuid.UUID, status string, limit, offset int) ([]Intern, int, error) {
	baseQuery := `
		FROM interns i
		JOIN users u ON u.id = i.user_id
		JOIN batches b ON b.id = i.batch_id
		LEFT JOIN institutions inst ON inst.id = i.institution_id
		LEFT JOIN users m ON m.id = i.mentor_id
		LEFT JOIN ranks rk ON rk.id = i.current_rank_id
		WHERE 1=1
	`
	var args []interface{}
	argIdx := 1

	if search != "" {
		baseQuery += fmt.Sprintf(" AND (u.full_name ILIKE $%d OR u.email ILIKE $%d OR i.id_number ILIKE $%d)", argIdx, argIdx, argIdx)
		args = append(args, "%"+search+"%")
		argIdx++
	}

	if batchID != nil {
		baseQuery += fmt.Sprintf(" AND i.batch_id = $%d", argIdx)
		args = append(args, *batchID)
		argIdx++
	}

	if institutionID != nil {
		baseQuery += fmt.Sprintf(" AND i.institution_id = $%d", argIdx)
		args = append(args, *institutionID)
		argIdx++
	}

	if status != "" {
		baseQuery += fmt.Sprintf(" AND i.status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) " + baseQuery
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total peserta magang: %w", err)
	}

	selectQuery := fmt.Sprintf(`
		SELECT i.id, i.user_id, i.batch_id, i.institution_id, i.id_number, i.mentor_id, i.status, i.join_date, i.end_date, i.current_rank_id, i.internship_xp, i.created_at, i.updated_at,
		       u.full_name, u.email, b.name, inst.name, m.full_name, rk.name
		%s
		ORDER BY i.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar peserta magang: %w", err)
	}
	defer rows.Close()

	var list []Intern
	for rows.Next() {
		var intern Intern
		if err := rows.Scan(
			&intern.ID, &intern.UserID, &intern.BatchID, &intern.InstitutionID, &intern.IDNumber, &intern.MentorID, &intern.Status, &intern.JoinDate, &intern.EndDate, &intern.CurrentRankID, &intern.InternshipXP, &intern.CreatedAt, &intern.UpdatedAt,
			&intern.UserFullName, &intern.UserEmail, &intern.BatchName, &intern.InstitutionName, &intern.MentorFullName, &intern.RankName,
		); err == nil {
			list = append(list, intern)
		}
	}

	return list, total, nil
}

// ============================================================================
// 4. ALUMNI
// ============================================================================

// Menyimpan rekod profil alumni baru dalam transaksi saat kelulusan peserta disahkan.
func (r *Repository) CreateAlumniTx(ctx context.Context, tx pgx.Tx, a *Alumni) error {
	query := `
		INSERT INTO alumni (id, user_id, batch_id, graduation_date, alumni_xp, certificate_id, is_public_profile, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, query, a.ID, a.UserID, a.BatchID, a.GraduationDate, a.AlumniXP, a.CertificateID, a.IsPublicProfile)
	if err != nil {
		return fmt.Errorf("gagal membuat rekod alumni dalam transaksi: %w", err)
	}
	return nil
}

// Mengambil data alumni berdasarkan user ID akun.
func (r *Repository) GetAlumniByUserID(ctx context.Context, userID uuid.UUID) (*Alumni, error) {
	query := `
		SELECT a.id, a.user_id, a.batch_id, a.graduation_date, a.alumni_xp, a.certificate_id, a.is_public_profile, a.created_at, a.updated_at,
		       u.full_name, u.email, b.name
		FROM alumni a
		JOIN users u ON u.id = a.user_id
		JOIN batches b ON b.id = a.batch_id
		WHERE a.user_id = $1
	`
	var a Alumni
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(
		&a.ID, &a.UserID, &a.BatchID, &a.GraduationDate, &a.AlumniXP, &a.CertificateID, &a.IsPublicProfile, &a.CreatedAt, &a.UpdatedAt,
		&a.UserFullName, &a.UserEmail, &a.BatchName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari data alumni berdasarkan user_id: %w", err)
	}
	return &a, nil
}

// Mengambil daftar alumni terdaftar dengan paginasi.
func (r *Repository) ListAlumni(ctx context.Context, limit, offset int) ([]Alumni, int, error) {
	var total int
	err := r.db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM alumni").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total alumni: %w", err)
	}

	query := `
		SELECT a.id, a.user_id, a.batch_id, a.graduation_date, a.alumni_xp, a.certificate_id, a.is_public_profile, a.created_at, a.updated_at,
		       u.full_name, u.email, b.name
		FROM alumni a
		JOIN users u ON u.id = a.user_id
		JOIN batches b ON b.id = a.batch_id
		ORDER BY a.graduation_date DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar alumni: %w", err)
	}
	defer rows.Close()

	var list []Alumni
	for rows.Next() {
		var a Alumni
		if err := rows.Scan(
			&a.ID, &a.UserID, &a.BatchID, &a.GraduationDate, &a.AlumniXP, &a.CertificateID, &a.IsPublicProfile, &a.CreatedAt, &a.UpdatedAt,
			&a.UserFullName, &a.UserEmail, &a.BatchName,
		); err == nil {
			list = append(list, a)
		}
	}
	return list, total, nil
}

// Mengubah peran pengguna dari INTERN menjadi ALUMNI pada tabel user_roles dalam transaksi.
func (r *Repository) SwitchUserRoleToAlumniTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID) error {
	// Ambil role ID untuk ALUMNI
	var alumniRoleID uuid.UUID
	err := tx.QueryRow(ctx, "SELECT id FROM roles WHERE name = 'ALUMNI'").Scan(&alumniRoleID)
	if err != nil {
		return fmt.Errorf("gagal mencari role ALUMNI: %w", err)
	}

	// Hapus peran INTERN
	_, _ = tx.Exec(ctx, "DELETE FROM user_roles WHERE user_id = $1 AND role_id IN (SELECT id FROM roles WHERE name = 'INTERN')", userID)

	// Tambahkan peran ALUMNI
	query := `
		INSERT INTO user_roles (id, user_id, role_id, scope_id)
		VALUES (gen_random_uuid(), $1, $2, (SELECT id FROM scopes WHERE scope_type = 'PUBLIC_DATA' LIMIT 1))
		ON CONFLICT (user_id, role_id, scope_id) DO NOTHING
	`
	_, err = tx.Exec(ctx, query, userID, alumniRoleID)
	if err != nil {
		return fmt.Errorf("gagal menambahkan peran ALUMNI pada pengguna: %w", err)
	}
	return nil
}

// ============================================================================
// 5. SKILLS & USER SKILLS
// ============================================================================

// Membuat master keahlian teknis baru ke dalam katalog skills.
func (r *Repository) CreateSkill(ctx context.Context, skill *Skill) error {
	query := `
		INSERT INTO skills (id, name, category, description, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
	`
	if skill.ID == uuid.Nil {
		skill.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, skill.ID, skill.Name, skill.Category, skill.Description)
	if err != nil {
		return fmt.Errorf("gagal membuat master skill baru: %w", err)
	}
	return nil
}

// Mengambil seluruh master katalog skill dengan filter kategori opsional.
func (r *Repository) ListSkills(ctx context.Context, category string) ([]Skill, error) {
	query := `SELECT id, name, category, description, created_at FROM skills`
	var args []interface{}
	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}
	query += " ORDER BY name ASC"

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar skill: %w", err)
	}
	defer rows.Close()

	var list []Skill
	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.ID, &s.Name, &s.Category, &s.Description, &s.CreatedAt); err == nil {
			list = append(list, s)
		}
	}
	return list, nil
}

// Menetapkan atau memperbarui tingkat kemahiran keahlian peserta pada tabel user_skills.
func (r *Repository) UpsertUserSkill(ctx context.Context, us *UserSkill) error {
	query := `
		INSERT INTO user_skills (id, user_id, skill_id, proficiency_level, created_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id, skill_id) DO UPDATE
		SET proficiency_level = EXCLUDED.proficiency_level
	`
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	_, err := r.db.Pool.Exec(ctx, query, us.ID, us.UserID, us.SkillID, us.ProficiencyLevel)
	if err != nil {
		return fmt.Errorf("gagal menetapkan skill pada pengguna: %w", err)
	}
	return nil
}

// Mengambil daftar keahlian dan tingkat kemahiran yang dimiliki oleh pengguna tertentu.
func (r *Repository) GetUserSkills(ctx context.Context, userID uuid.UUID) ([]UserSkill, error) {
	query := `
		SELECT us.id, us.user_id, us.skill_id, us.proficiency_level, us.created_at,
		       s.name, s.category, s.description
		FROM user_skills us
		JOIN skills s ON s.id = us.skill_id
		WHERE us.user_id = $1
		ORDER BY us.proficiency_level DESC, s.name ASC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar skill pengguna: %w", err)
	}
	defer rows.Close()

	var list []UserSkill
	for rows.Next() {
		var us UserSkill
		if err := rows.Scan(&us.ID, &us.UserID, &us.SkillID, &us.ProficiencyLevel, &us.CreatedAt, &us.SkillName, &us.SkillCategory, &us.Description); err == nil {
			list = append(list, us)
		}
	}
	return list, nil
}

// Menghapus rekaman keahlian yang terhubung dengan pengguna.
func (r *Repository) DeleteUserSkill(ctx context.Context, userID, skillID uuid.UUID) error {
	query := `DELETE FROM user_skills WHERE user_id = $1 AND skill_id = $2`
	_, err := r.db.Pool.Exec(ctx, query, userID, skillID)
	if err != nil {
		return fmt.Errorf("gagal menghapus skill pengguna: %w", err)
	}
	return nil
}
