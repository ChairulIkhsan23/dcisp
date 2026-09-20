package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/shared/utils"
	"github.com/google/uuid"
)

// Menghitung hash SHA-256 dari token refresh mentah untuk penyimpanan sesi yang aman.
func hashRefreshToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

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

	// Mendaftarkan sesi refresh terotasi (hanya hash yang disimpan, SECURITY.md Section 1.2)
	if _, err := s.repo.CreateRefreshSession(ctx, user.ID, hashRefreshToken(tokens.RefreshToken), time.Now().Add(time.Duration(s.cfg.JWTRefreshDurationDays)*24*time.Hour)); err != nil {
		return nil, nil, fmt.Errorf("gagal mendaftarkan sesi refresh: %w", err)
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

// Menghasilkan pasangan token baru dengan rotasi: token refresh lama dicabut seketika
// dan digantikan pasangan baru (SECURITY.md Section 1.2). Token yang sudah dicabut,
// kedaluwarsa, atau milik pengguna nonaktif ditolak.
func (s *Service) RefreshToken(ctx context.Context, refreshTokenString string) (*utils.TokenPair, error) {
	claims, err := utils.ValidateRefreshToken(s.cfg.JWTSecret, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("refresh token tidak valid atau telah kedaluwarsa: %w", err)
	}

	// Sesi refresh wajib terdaftar dan belum dicabut (revocation seketika)
	session, err := s.repo.FindActiveRefreshSession(ctx, hashRefreshToken(refreshTokenString))
	if err != nil {
		return nil, fmt.Errorf("gagal memeriksa sesi refresh: %w", err)
	}
	if session == nil {
		return nil, errors.New("sesi refresh tidak ditemukan atau telah dicabut, silakan login ulang")
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

	newTokens, err := utils.GenerateTokenPair(
		s.cfg.JWTSecret,
		s.cfg.JWTAccessDurationMinutes,
		s.cfg.JWTRefreshDurationDays,
		user.ID,
		user.Email,
		primaryRole,
		scopes,
	)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat pasangan token: %w", err)
	}

	// Rotasi: cabut token lama, daftarkan token baru
	if err := s.repo.RevokeRefreshSession(ctx, hashRefreshToken(refreshTokenString)); err != nil {
		return nil, fmt.Errorf("gagal mencabut sesi refresh lama: %w", err)
	}
	if _, err := s.repo.CreateRefreshSession(ctx, user.ID, hashRefreshToken(newTokens.RefreshToken), time.Now().Add(time.Duration(s.cfg.JWTRefreshDurationDays)*24*time.Hour)); err != nil {
		return nil, fmt.Errorf("gagal mendaftarkan sesi refresh baru: %w", err)
	}

	return newTokens, nil
}

// Mencabut sesi refresh token milik pengguna yang sedang login (logout per perangkat).
func (s *Service) Logout(ctx context.Context, userID uuid.UUID, refreshTokenString string) error {
	if err := s.repo.RevokeRefreshSession(ctx, hashRefreshToken(refreshTokenString)); err != nil {
		return fmt.Errorf("gagal mencabut sesi refresh: %w", err)
	}
	return nil
}

// Mencabut seluruh sesi refresh aktif milik pengguna (dipakai saat penonaktifan akun paksa).
func (s *Service) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.RevokeAllUserRefreshSessions(ctx, userID)
}

// Mengambil seluruh daftar master role sistem dari repository.
func (s *Service) GetRoles(ctx context.Context) ([]Role, error) {
	roles, err := s.repo.GetAllRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil daftar peran: %w", err)
	}
	return roles, nil
}
