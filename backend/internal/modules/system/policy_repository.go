package system

import (
	"context"
	"errors"
	"fmt"

	"dcisp/backend/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type PolicyRepository struct {
	db *database.PostgresDB
}

// Menginisialisasi instance baru repository policy engine.
func NewPolicyRepository(db *database.PostgresDB) *PolicyRepository {
	return &PolicyRepository{db: db}
}

// Menyimpan entitas kebijakan baru ke dalam tabel policies database.
func (r *PolicyRepository) Create(ctx context.Context, p *Policy) error {
	query := `
		INSERT INTO policies (id, policy_domain, name, effective_date, condition_rules, action_definitions, priority_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err := r.db.Pool.Exec(ctx, query, p.ID, p.PolicyDomain, p.Name, p.EffectiveDate, p.ConditionRules, p.ActionDefinitions, p.PriorityOrder, p.IsActive)
	if err != nil {
		return fmt.Errorf("gagal menyimpan kebijakan baru: %w", err)
	}
	return nil
}

// Mengambil satu data kebijakan dari database berdasarkan identitas unik UUID.
func (r *PolicyRepository) GetByID(ctx context.Context, id uuid.UUID) (*Policy, error) {
	query := `
		SELECT id, policy_domain, name, effective_date, condition_rules, action_definitions, priority_order, is_active, created_at, updated_at
		FROM policies
		WHERE id = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var p Policy
	err := row.Scan(&p.ID, &p.PolicyDomain, &p.Name, &p.EffectiveDate, &p.ConditionRules, &p.ActionDefinitions, &p.PriorityOrder, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari kebijakan berdasarkan id: %w", err)
	}
	return &p, nil
}

// Mengambil seluruh kebijakan aktif untuk domain tertentu diurutkan berdasarkan prioritas.
func (r *PolicyRepository) GetActiveByDomain(ctx context.Context, domain string) ([]Policy, error) {
	query := `
		SELECT id, policy_domain, name, effective_date, condition_rules, action_definitions, priority_order, is_active, created_at, updated_at
		FROM policies
		WHERE policy_domain = $1 AND is_active = TRUE
		ORDER BY priority_order ASC, created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, domain)
	if err != nil {
		return nil, fmt.Errorf("gagal mengueri kebijakan aktif berdasarkan domain: %w", err)
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.PolicyDomain, &p.Name, &p.EffectiveDate, &p.ConditionRules, &p.ActionDefinitions, &p.PriorityOrder, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err == nil {
			policies = append(policies, p)
		}
	}
	return policies, nil
}

// Mengambil daftar kebijakan dengan filter domain dan status serta paginasi.
func (r *PolicyRepository) List(ctx context.Context, domain string, isActive *bool, limit, offset int) ([]Policy, int, error) {
	baseQuery := `FROM policies WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if domain != "" {
		baseQuery += fmt.Sprintf(" AND policy_domain = $%d", argIdx)
		args = append(args, domain)
		argIdx++
	}

	if isActive != nil {
		baseQuery += fmt.Sprintf(" AND is_active = $%d", argIdx)
		args = append(args, *isActive)
		argIdx++
	}

	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int
	err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal menghitung total kebijakan: %w", err)
	}

	selectQuery := fmt.Sprintf(`
		SELECT id, policy_domain, name, effective_date, condition_rules, action_definitions, priority_order, is_active, created_at, updated_at
		%s
		ORDER BY priority_order ASC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal mengambil daftar kebijakan: %w", err)
	}
	defer rows.Close()

	var policies []Policy
	for rows.Next() {
		var p Policy
		if err := rows.Scan(&p.ID, &p.PolicyDomain, &p.Name, &p.EffectiveDate, &p.ConditionRules, &p.ActionDefinitions, &p.PriorityOrder, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err == nil {
			policies = append(policies, p)
		}
	}

	return policies, total, nil
}

// Memperbarui data kebijakan yang sudah ada pada database.
func (r *PolicyRepository) Update(ctx context.Context, p *Policy) error {
	query := `
		UPDATE policies
		SET name = $2, effective_date = $3, condition_rules = $4, action_definitions = $5, priority_order = $6, is_active = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	cmdTag, err := r.db.Pool.Exec(ctx, query, p.ID, p.Name, p.EffectiveDate, p.ConditionRules, p.ActionDefinitions, p.PriorityOrder, p.IsActive)
	if err != nil {
		return fmt.Errorf("gagal memperbarui data kebijakan: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("kebijakan tidak ditemukan")
	}
	return nil
}

// Menghapus data kebijakan dari database berdasarkan identitas unik.
func (r *PolicyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM policies WHERE id = $1`
	cmdTag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("gagal menghapus kebijakan: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return errors.New("kebijakan tidak ditemukan")
	}
	return nil
}
