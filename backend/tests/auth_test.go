package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/modules/identity"
	"dcisp/backend/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *database.PostgresDB, *config.Config) {
	gin.SetMode(gin.TestMode)
	cfg := config.LoadConfig()

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)

	repo := identity.NewRepository(db)
	service := identity.NewService(repo, cfg)
	ctrl := identity.NewController(service)

	r := gin.New()
	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", ctrl.Login)
		authGroup.POST("/refresh", ctrl.RefreshToken)

		protected := authGroup.Group("")
		protected.Use(middleware.AuthJWT(cfg))
		{
			protected.GET("/me", ctrl.GetMe)
			protected.GET("/roles", middleware.RequirePermission(db, "system.roles", "view", "SYSTEM"), ctrl.GetRoles)
		}
	}

	return r, db, cfg
}

func TestArgon2idPasswordHashing(t *testing.T) {
	password := "SecretP@ssword2026!"

	// 1. Hash password
	hash, err := utils.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")

	// 2. Verify with correct password
	valid, err := utils.VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)

	// 3. Verify with wrong password
	invalid, err := utils.VerifyPassword("WrongPassword123", hash)
	require.NoError(t, err)
	assert.False(t, invalid)
}

func TestJWTTokenGenerationAndValidation(t *testing.T) {
	secret := "test_secret_key_1234567890123456"
	userID := uuid.New()
	email := "intern_hero@dcisp.internal"
	role := "INTERN"
	scopes := []string{"OWN_DATA"}

	// 1. Generate Token Pair
	tokens, err := utils.GenerateTokenPair(secret, 15, 7, userID, email, role, scopes)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, int64(15*60), tokens.ExpiresIn)

	// 2. Validate Access Token
	claims, err := utils.ValidateToken(secret, tokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, scopes, claims.Scopes)
}

func TestLoginInvalidCredentials(t *testing.T) {
	r, db, _ := setupTestRouter(t)
	defer db.Close()

	payload := map[string]string{
		"email":    "nonexistent_user@dcisp.internal",
		"password": "wrongpassword123",
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestSuperAdminLoginAndProfile(t *testing.T) {
	r, db, _ := setupTestRouter(t)
	defer db.Close()

	ctx := context.Background()

	// Seed unique test superadmin with known password
	testAdminID := uuid.New()
	testAdminEmail := "creator_" + testAdminID.String()[:8] + "@dcisp.internal"
	testAdminPass := "CreatorSecretPass2026!"
	hash, err := utils.HashPassword(testAdminPass)
	require.NoError(t, err)

	superAdminRoleID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	systemScopeID := uuid.MustParse("20000000-0000-0000-0000-000000000001")

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, $3, 'The Creator Test', 'ACTIVE')
	`, testAdminID, testAdminEmail, hash)
	require.NoError(t, err)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, scope_id)
		VALUES ($1, $2, $3)
	`, testAdminID, superAdminRoleID, systemScopeID)
	require.NoError(t, err)

	// 1. Test Login Endpoint
	loginPayload := map[string]string{
		"email":    testAdminEmail,
		"password": testAdminPass,
	}
	body, _ := json.Marshal(loginPayload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Tokens struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
			} `json:"tokens"`
			Profile struct {
				Role string `json:"role"`
			} `json:"profile"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotEmpty(t, resp.Data.Tokens.AccessToken)
	assert.NotEmpty(t, resp.Data.Tokens.RefreshToken)
	assert.Equal(t, "SUPER_ADMIN", resp.Data.Profile.Role)

	// 2. Test Protected /me Endpoint
	reqMe, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+resp.Data.Tokens.AccessToken)
	wMe := httptest.NewRecorder()
	r.ServeHTTP(wMe, reqMe)

	assert.Equal(t, http.StatusOK, wMe.Code)

	// 3. Test Refresh Token Endpoint
	refreshPayload := map[string]string{
		"refresh_token": resp.Data.Tokens.RefreshToken,
	}
	refreshBody, _ := json.Marshal(refreshPayload)

	reqRefresh, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
	reqRefresh.Header.Set("Content-Type", "application/json")
	wRefresh := httptest.NewRecorder()
	r.ServeHTTP(wRefresh, reqRefresh)

	assert.Equal(t, http.StatusOK, wRefresh.Code)
}
