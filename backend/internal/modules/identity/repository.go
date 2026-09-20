package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// RefreshSession adalah proyeksi sesi refresh token tersimpan (hash, bukan token mentah).
type RefreshSession struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

// Menyimpan sesi refresh token baru (hanya hash SHA-256 token, SECURITY.md Section 1.2).
func (r *Repository) CreateRefreshSession(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*RefreshSession, error) {
	var s RefreshSession
	err := r.db.Pool.QueryRow(ctx, `
		INSERT INTO refresh_sessions (id, user_id, token_hash, expires_at, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, CURRENT_TIMESTAMP)
		RETURNING id, user_id, token_hash, expires_at, revoked_at
	`, userID, tokenHash, expiresAt).Scan(&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt, &s.RevokedAt)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan sesi refresh: %w", err)
	}
	return &s, nil
}

// Mengambil sesi refresh aktif berdasarkan hash tokennya.
func (r *Repository) FindActiveRefreshSession(ctx context.Context, tokenHash string) (*RefreshSession, error) {
	var s RefreshSession
	err := r.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, revoked_at
		FROM refresh_sessions
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP
	`, tokenHash).Scan(&s.ID, &s.UserID, &s.TokenHash, &s.ExpiresAt, &s.RevokedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari sesi refresh: %w", err)
	}
	return &s, nil
}

// Mencabut satu sesi refresh token berdasarkan hash tokennya (logout per perangkat).
func (r *Repository) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	_, err := r.db.Pool.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE token_hash = $1 AND revoked_at IS NULL
	`, tokenHash)
	if err != nil {
		return fmt.Errorf("gagal mencabut sesi refresh: %w", err)
	}
	return nil
}

// Mencabut seluruh sesi refresh aktif milik pengguna (dipakai saat penonaktifan akun paksa).
func (r *Repository) RevokeAllUserRefreshSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	cmdTag, err := r.db.Pool.Exec(ctx, `
		UPDATE refresh_sessions
		SET revoked_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	if err != nil {
		return 0, fmt.Errorf("gagal mencabut seluruh sesi pengguna: %w", err)
	}
	return cmdTag.RowsAffected(), nil
}

// Mengambil data pengguna dari database berdasarkan alamat email terdaftar.
func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, full_name, avatar_file_id, status, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	row := r.db.Pool.QueryRow(ctx, query, email)

	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.AvatarFileID, &u.Status, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("gagal mencari pengguna berdasarkan email: %w", err)
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
		return nil, fmt.Errorf("gagal mencari pengguna berdasarkan id: %w", err)
	}

	return &u, nil
}

// Mengambil seluruh daftar peran dan cakupan scope yang dimiliki oleh pengguna.
func (r *Repository) GetUserRolesAndScopes(ctx context.Context, userID uuid.UUID) ([]string, []string, error) {
	query := `
		SELECT DISTINCT r.name, COALESCE(s.scope_type, 'OWN_DATA')
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		LEFT JOIN scopes s ON s.id = ur.scope_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mengueri peran dan scope pengguna: %w", err)
	}
	defer rows.Close()

	roleMap := make(map[string]bool)
	scopeMap := make(map[string]bool)
	var roles []string
	var scopes []string

	for rows.Next() {
		var roleName, scopeType string
		if err := rows.Scan(&roleName, &scopeType); err != nil {
			return nil, nil, fmt.Errorf("gagal memindai data peran dan scope: %w", err)
		}
		if !roleMap[roleName] {
			roleMap[roleName] = true
			roles = append(roles, roleName)
		}
		if !scopeMap[scopeType] {
			scopeMap[scopeType] = true
			scopes = append(scopes, scopeType)
		}
	}

	if len(roles) == 0 {
		return []string{"GUEST"}, []string{"OWN_DATA"}, nil
	}

	return roles, scopes, nil
}

// Mengambil nama role utama dan cakupan seluruh scope yang dimiliki oleh pengguna untuk kompatibilitas.
func (r *Repository) GetUserRoleAndScopes(ctx context.Context, userID uuid.UUID) (string, []string, error) {
	roles, scopes, err := r.GetUserRolesAndScopes(ctx, userID)
	if err != nil {
		return "", nil, err
	}
	primaryRole := "GUEST"
	if len(roles) > 0 {
		primaryRole = roles[0]
	}
	return primaryRole, scopes, nil
}

// Mengambil daftar izin akses spesifik yang dimiliki oleh pengguna dari seluruh perannya.
func (r *Repository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT CONCAT(p.resource, ':', p.action)
		FROM permissions p
		JOIN user_roles ur ON ur.role_id = p.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengueri izin pengguna: %w", err)
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
		return nil, fmt.Errorf("gagal mengueri daftar peran: %w", err)
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
