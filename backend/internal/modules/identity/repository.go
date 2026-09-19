package identity

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

// Menginisialisasi instance baru identity repository.
func NewRepository(db *database.PostgresDB) *Repository {
	return &Repository{db: db}
}

// Mengambil data pengguna aktif dari database berdasarkan alamat email.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, avatar_file_id, status, created_at, updated_at
		FROM users
		WHERE email = $1 AND status = 'ACTIVE'
	`
	row := r.db.Pool.QueryRow(ctx, query, email)

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.AvatarFileID, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding user by email: %w", err)
	}

	return &u, nil
}

// Mengambil data pengguna dari database berdasarkan ID uniknya.
func (r *Repository) FindByID(ctx context.Context, id uuid.UUID) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, avatar_file_id, status, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, id)

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.AvatarFileID, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding user by id: %w", err)
	}

	return &u, nil
}

// Mengambil nama role aktif dan cakupan scope yang dimiliki oleh pengguna.
func (r *Repository) GetUserRoleAndScopes(ctx context.Context, userID uuid.UUID) (string, []string, error) {
	query := `
		SELECT r.name, COALESCE(s.scope_type, 'OWN_DATA')
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		LEFT JOIN scopes s ON s.id = ur.scope_id
		WHERE ur.user_id = $1
		LIMIT 1
	`
	var roleName, scopeType string
	err := r.db.Pool.QueryRow(ctx, query, userID).Scan(&roleName, &scopeType)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "GUEST", []string{"OWN_DATA"}, nil
		}
		return "", nil, fmt.Errorf("error querying user role: %w", err)
	}

	return roleName, []string{scopeType}, nil
}

// Mengambil daftar izin akses spesifik yang dimiliki oleh pengguna.
func (r *Repository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT CONCAT(p.resource, ':', p.action)
		FROM permissions p
		JOIN user_roles ur ON ur.role_id = p.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error querying user permissions: %w", err)
	}
	defer rows.Close()

	var perms []string
	for rows.Next() {
		var perm string
		if err := rows.Scan(&perm); err == nil {
			perms = append(perms, perm)
		}
	}

	return perms, nil
}

// Mengambil seluruh daftar master role sistem dari database.
func (r *Repository) GetAllRoles(ctx context.Context) ([]Role, error) {
	query := `SELECT id, name, description, is_system FROM roles ORDER BY name ASC`
	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying roles: %w", err)
	}
	defer rows.Close()

	var rolesList []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.IsSystem); err == nil {
			rolesList = append(rolesList, role)
		}
	}

	return rolesList, nil
}
