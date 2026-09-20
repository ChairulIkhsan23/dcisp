package identity

import (
	"context"
	"errors"
	"fmt"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/shared/utils"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("email atau kata sandi tidak valid")
	ErrUserNotFound       = errors.New("pengguna tidak ditemukan")
	ErrAccountInactive    = errors.New("akun tidak aktif")
)

type Service struct {
	repo *Repository
	cfg  *config.Config
}

// Menginisialisasi instance baru identity service.
func NewService(repo *Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

// Memverifikasi kredensial login pengguna dan menghasilkan token akses JWT serta token refresh.
func (s *Service) Login(ctx context.Context, email, password string) (*utils.TokenPair, *UserProfileResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, ErrInvalidCredentials
	}

	valid, err := utils.VerifyPassword(password, user.PasswordHash)
	if err != nil || !valid {
		return nil, nil, ErrInvalidCredentials
	}

	role, scopes, err := s.repo.GetUserRoleAndScopes(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mengambil peran pengguna: %w", err)
	}

	permissions, err := s.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		permissions = []string{}
	}

	tokens, err := utils.GenerateTokenPair(
		s.cfg.JWTSecret,
		s.cfg.JWTAccessDurationMinutes,
		s.cfg.JWTRefreshDurationDays,
		user.ID,
		user.Email,
		role,
		scopes,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal membuat pasangan token: %w", err)
	}

	profile := &UserProfileResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		Status:      user.Status,
		Role:        role,
		Scopes:      scopes,
		Permissions: permissions,
	}

	return tokens, profile, nil
}

// Mengambil informasi profil lengkap, role, scope, dan permission pengguna.
func (s *Service) GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfileResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	role, scopes, err := s.repo.GetUserRoleAndScopes(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	permissions, err := s.repo.GetUserPermissions(ctx, user.ID)
	if err != nil {
		permissions = []string{}
	}

	return &UserProfileResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		Status:      user.Status,
		Role:        role,
		Scopes:      scopes,
		Permissions: permissions,
	}, nil
}

// Menghasilkan pasangan token baru menggunakan token refresh yang valid.
func (s *Service) RefreshToken(ctx context.Context, refreshTokenString string) (*utils.TokenPair, error) {
	claims, err := utils.ValidateToken(s.cfg.JWTSecret, refreshTokenString)
	if err != nil {
		return nil, errors.New("refresh token tidak valid atau telah kedaluwarsa")
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil || user == nil {
		return nil, errors.New("pengguna tidak ditemukan atau tidak aktif")
	}

	role, scopes, err := s.repo.GetUserRoleAndScopes(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return utils.GenerateTokenPair(
		s.cfg.JWTSecret,
		s.cfg.JWTAccessDurationMinutes,
		s.cfg.JWTRefreshDurationDays,
		user.ID,
		user.Email,
		role,
		scopes,
	)
}

// Mengambil seluruh daftar role master dari repository.
func (s *Service) GetRoles(ctx context.Context) ([]Role, error) {
	return s.repo.GetAllRoles(ctx)
}
