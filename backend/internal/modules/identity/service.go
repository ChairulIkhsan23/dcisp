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

// Memverifikasi kredensial login pengguna aktif dan menghasilkan pasangan token akses serta token refresh.
func (s *Service) Login(ctx context.Context, email, password string) (*utils.TokenPair, *UserProfileResponse, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mencari data pengguna: %w", err)
	}
	if user == nil {
		return nil, nil, ErrInvalidCredentials
	}

	valid, err := utils.VerifyPassword(password, user.PasswordHash)
	if err != nil || !valid {
		return nil, nil, ErrInvalidCredentials
	}

	if user.Status != "ACTIVE" {
		return nil, nil, ErrAccountInactive
	}

	roles, scopes, err := s.repo.GetUserRolesAndScopes(ctx, user.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("gagal mengambil peran pengguna: %w", err)
	}
	primaryRole := "GUEST"
	if len(roles) > 0 {
		primaryRole = roles[0]
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
		primaryRole,
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
		Role:        primaryRole,
		Roles:       roles,
		Scopes:      scopes,
		Permissions: permissions,
	}

	return tokens, profile, nil
}

// Mengambil informasi profil lengkap, seluruh peran, cakupan scope, dan izin aktif pengguna.
func (s *Service) GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfileResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("gagal mencari data pengguna berdasarkan id: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	if user.Status != "ACTIVE" {
		return nil, ErrAccountInactive
	}

	roles, scopes, err := s.repo.GetUserRolesAndScopes(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil peran pengguna: %w", err)
	}
	primaryRole := "GUEST"
	if len(roles) > 0 {
		primaryRole = roles[0]
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
		Role:        primaryRole,
		Roles:       roles,
		Scopes:      scopes,
		Permissions: permissions,
	}, nil
}

// Menghasilkan pasangan token baru menggunakan token refresh yang valid dari pengguna berstatus aktif.
func (s *Service) RefreshToken(ctx context.Context, refreshTokenString string) (*utils.TokenPair, error) {
	claims, err := utils.ValidateToken(s.cfg.JWTSecret, refreshTokenString, utils.TokenTypeRefresh)
	if err != nil {
		return nil, fmt.Errorf("refresh token tidak valid atau telah kedaluwarsa: %w", err)
	}

	user, err := s.repo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("gagal mencari data pengguna saat pembaruan token: %w", err)
	}
	if user == nil || user.Status != "ACTIVE" {
		return nil, errors.New("pengguna tidak ditemukan atau tidak aktif")
	}

	roles, scopes, err := s.repo.GetUserRolesAndScopes(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil peran pengguna saat pembaruan token: %w", err)
	}
	primaryRole := "GUEST"
	if len(roles) > 0 {
		primaryRole = roles[0]
	}

	return utils.GenerateTokenPair(
		s.cfg.JWTSecret,
		s.cfg.JWTAccessDurationMinutes,
		s.cfg.JWTRefreshDurationDays,
		user.ID,
		user.Email,
		primaryRole,
		scopes,
	)
}

// Mengambil seluruh daftar master role sistem dari repository.
func (s *Service) GetRoles(ctx context.Context) ([]Role, error) {
	roles, err := s.repo.GetAllRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar peran: %w", err)
	}
	return roles, nil
}
