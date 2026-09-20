package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/modules/identity"
	"dcisp/backend/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menginisialisasi router pengujian Gin beserta koneksi database PostgreSQL dan Redis.
func setupTestRouter(t *testing.T) (*gin.Engine, *database.PostgresDB, *database.RedisClient, *config.Config) {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Overload("../../.env")
	_ = godotenv.Overload("../.env")
	_ = godotenv.Overload(".env")

	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)

	rdb, err := database.NewRedisClient(cfg)
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
			protected.GET("/roles", middleware.RequirePermission(db, rdb, "system.roles", "view", "SYSTEM"), ctrl.GetRoles)
		}
	}

	return r, db, rdb, cfg
}

// Menguji fungsi hashing kata sandi dan verifikasi menggunakan algoritma Argon2id.
func TestArgon2idPasswordHashing(t *testing.T) {
	password := "SecretP@ssword2026!"

	hash, err := utils.HashPassword(password)
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.Contains(t, hash, "$argon2id$")

	valid, err := utils.VerifyPassword(password, hash)
	require.NoError(t, err)
	assert.True(t, valid)

	invalid, err := utils.VerifyPassword("WrongPassword123", hash)
	require.NoError(t, err)
	assert.False(t, invalid)
}

// Menguji pembuatan pasangan token JWT dan validasi klaim identitas di dalamnya.
func TestJWTTokenGenerationAndValidation(t *testing.T) {
	secret := "test_secret_key_1234567890123456"
	userID := uuid.New()
	email := "intern_hero@dcisp.internal"
	role := "INTERN"
	scopes := []string{"OWN_DATA"}

	tokens, err := utils.GenerateTokenPair(secret, 15, 7, userID, email, role, scopes)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.AccessToken)
	assert.NotEmpty(t, tokens.RefreshToken)
	assert.Equal(t, int64(15*60), tokens.ExpiresIn)

	// Validasi Access Token
	claims, err := utils.ValidateToken(secret, tokens.AccessToken, utils.TokenTypeAccess)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, role, claims.Role)
	assert.Equal(t, scopes, claims.Scopes)
	assert.Equal(t, utils.TokenTypeAccess, claims.TokenType)

	// Validasi Refresh Token
	refreshClaims, err := utils.ValidateToken(secret, tokens.RefreshToken, utils.TokenTypeRefresh)
	require.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, utils.TokenTypeRefresh, refreshClaims.TokenType)
}

// Menguji pemisahan ketat antara token akses dan token refresh untuk mencegah pertukaran tipe token.
func TestJWTTokenTypeValidation(t *testing.T) {
	secret := "test_secret_key_1234567890123456"
	userID := uuid.New()
	email := "intern_hero@dcisp.internal"
	role := "INTERN"
	scopes := []string{"OWN_DATA"}

	tokens, err := utils.GenerateTokenPair(secret, 15, 7, userID, email, role, scopes)
	require.NoError(t, err)

	// 1. Refresh token digunakan sebagai access token -> harus ditolak
	_, err = utils.ValidateToken(secret, tokens.RefreshToken, utils.TokenTypeAccess)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tipe token tidak valid")

	// 2. Access token digunakan sebagai refresh token -> harus ditolak
	_, err = utils.ValidateToken(secret, tokens.AccessToken, utils.TokenTypeRefresh)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tipe token tidak valid")

	// 3. Token tanpa klaim tipe -> harus ditolak
	legacyClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		Subject:   userID.String(),
		Issuer:    "dcisp-auth-service",
	}
	legacyToken := jwt.NewWithClaims(jwt.SigningMethodHS256, legacyClaims)
	legacySigned, err := legacyToken.SignedString([]byte(secret))
	require.NoError(t, err)

	_, err = utils.ValidateToken(secret, legacySigned, utils.TokenTypeAccess)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tipe token tidak ditemukan")
}

// Menguji kegagalan inisialisasi konfigurasi jika kunci rahasia JWT tidak ditentukan.
func TestJWTSecretFailHard(t *testing.T) {
	origSecret, hadOrig := os.LookupEnv("JWT_SECRET")
	defer func() {
		if hadOrig && origSecret != "" {
			_ = os.Setenv("JWT_SECRET", origSecret)
		} else {
			_ = godotenv.Overload("../../.env")
		}
	}()

	_ = os.Setenv("JWT_SECRET", "")
	_, err := config.LoadConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

// Menguji penolakan permintaan login ketika kombinasi email atau kata sandi salah.
func TestLoginInvalidCredentials(t *testing.T) {
	r, db, rdb, _ := setupTestRouter(t)
	defer db.Close()
	defer rdb.Close()

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

// Menguji penolakan proses login bagi pengguna yang berstatus tidak aktif.
func TestLoginInactiveUserRejected(t *testing.T) {
	r, db, rdb, _ := setupTestRouter(t)
	defer db.Close()
	defer rdb.Close()

	ctx := context.Background()
	inactiveUserID := uuid.New()
	inactiveEmail := "inactive_" + inactiveUserID.String()[:8] + "@dcisp.internal"
	pass := "InactivePass2026!"
	hash, err := utils.HashPassword(pass)
	require.NoError(t, err)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, $3, 'Inactive User Test', 'INACTIVE')
	`, inactiveUserID, inactiveEmail, hash)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", inactiveUserID)
	})

	loginPayload := map[string]string{
		"email":    inactiveEmail,
		"password": pass,
	}
	body, _ := json.Marshal(loginPayload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Menguji penolakan pembaruan token bagi pengguna yang berstatus tidak aktif.
func TestRefreshTokenInactiveUserRejected(t *testing.T) {
	r, db, rdb, cfg := setupTestRouter(t)
	defer db.Close()
	defer rdb.Close()

	ctx := context.Background()
	suspendedUserID := uuid.New()
	suspendedEmail := "suspended_" + suspendedUserID.String()[:8] + "@dcisp.internal"
	pass := "SuspendedPass2026!"
	hash, err := utils.HashPassword(pass)
	require.NoError(t, err)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, $3, 'Suspended User Test', 'SUSPENDED')
	`, suspendedUserID, suspendedEmail, hash)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", suspendedUserID)
	})

	tokens, err := utils.GenerateTokenPair(cfg.JWTSecret, 15, 7, suspendedUserID, suspendedEmail, "INTERN", []string{"OWN_DATA"})
	require.NoError(t, err)

	refreshPayload := map[string]string{
		"refresh_token": tokens.RefreshToken,
	}
	refreshBody, _ := json.Marshal(refreshPayload)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewBuffer(refreshBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// Menguji alur lengkap autentikasi Super Admin, pengambilan profil, dan pembaruan token dengan pembersihan data otomatis.
func TestSuperAdminLoginAndProfile(t *testing.T) {
	r, db, rdb, _ := setupTestRouter(t)
	defer db.Close()
	defer rdb.Close()

	ctx := context.Background()

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

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_roles WHERE user_id = $1", testAdminID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testAdminID)
	})

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

// Menguji keamanan type assertion UUID pada controller profil pengguna.
func TestSafeUUIDTypeAssertion(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := identity.NewController(nil)

	r := gin.New()
	r.GET("/test-me", func(c *gin.Context) {
		// Set invalid type non-UUID into context
		c.Set(middleware.ContextUserIDKey, "not-a-uuid-string")
		c.Next()
	}, ctrl.GetMe)

	req, _ := http.NewRequest(http.MethodGet, "/test-me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Harusnya tidak panic dan mengembalikan 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
