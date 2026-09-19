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
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrAccountInactive    = errors.New("account is not active")
)

type Service struct {
	repo *Repository
	cfg  *config.Config
}

func NewService(repo *Repository, cfg *config.Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

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
		return nil, nil, fmt.Errorf("failed to retrieve user role: %w", err)
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
		return nil, nil, fmt.Errorf("failed to generate token pair: %w", err)
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

func (s *Service) GetRoles(ctx context.Context) ([]Role, error) {
	return s.repo.GetAllRoles(ctx)
}
