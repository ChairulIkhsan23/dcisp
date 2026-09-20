package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/shared/response"
	"dcisp/backend/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji isolasi hak akses peran Scanner Operator hanya pada cakupan scanner_only dan penolakan pada modul finansial.
func TestScannerOperatorIsolation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	rdb, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer rdb.Close()

	ctx := context.Background()

	// Seed user dengan peran SCANNER_OPERATOR
	testScannerUserID := uuid.New()
	testEmail := fmt.Sprintf("gatekeeper_%s@dcisp.internal", testScannerUserID.String()[:8])
	scannerRoleID := uuid.MustParse("10000000-0000-0000-0000-000000000008")
	scannerScopeID := uuid.MustParse("20000000-0000-0000-0000-000000000006")

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Gatekeeper Tester', 'ACTIVE')
	`, testScannerUserID, testEmail)
	require.NoError(t, err)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, scope_id)
		VALUES ($1, $2, $3)
	`, testScannerUserID, scannerRoleID, scannerScopeID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_roles WHERE user_id = $1", testScannerUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testScannerUserID)
		_ = rdb.Client.Del(context.Background(), fmt.Sprintf("rbac:permissions:%s", testScannerUserID.String())).Err()
	})

	r := gin.New()
	r.Use(middleware.AuthJWT(cfg))

	// Endpoint terproteksi yang membutuhkan akses finansial
	r.GET("/api/v1/finance/ledger", middleware.RequirePermission(db, rdb, "finance.ledger", "view", "FINANCIAL_DATA"), func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Buku besar berhasil diakses", nil)
	})

	// Endpoint terproteksi yang diizinkan untuk scanner operator
	r.GET("/api/v1/attendance/scanner/view", middleware.RequirePermission(db, rdb, "attendance.scanner", "view", "SCANNER_ONLY"), func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Tampilan pemindai berhasil diakses", nil)
	})

	tokens, err := utils.GenerateTokenPair(cfg.JWTSecret, 15, 7, testScannerUserID, testEmail, "SCANNER_OPERATOR", []string{"SCANNER_ONLY"})
	require.NoError(t, err)

	// 1. Akses ke Finance Ledger -> Harus 403 Forbidden (BR-002)
	req1, _ := http.NewRequest(http.MethodGet, "/api/v1/finance/ledger", nil)
	req1.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusForbidden, w1.Code)

	// 2. Akses ke Scanner View -> Harus 200 OK
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/attendance/scanner/view", nil)
	req2.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
}

// Menguji mekanisme caching izin RBAC pada Redis dengan verifikasi cache miss, cache hit, dan TTL.
func TestDynamicRBACRedisCache(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	rdb, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer rdb.Close()

	ctx := context.Background()
	testUserID := uuid.New()
	testEmail := fmt.Sprintf("cache_tester_%s@dcisp.internal", testUserID.String()[:8])
	scannerRoleID := uuid.MustParse("10000000-0000-0000-0000-000000000008")
	scannerScopeID := uuid.MustParse("20000000-0000-0000-0000-000000000006")

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Cache Tester', 'ACTIVE')
	`, testUserID, testEmail)
	require.NoError(t, err)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, scope_id)
		VALUES ($1, $2, $3)
	`, testUserID, scannerRoleID, scannerScopeID)
	require.NoError(t, err)

	cacheKey := fmt.Sprintf("rbac:permissions:%s", testUserID.String())

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_roles WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
		_ = rdb.Client.Del(context.Background(), cacheKey).Err()
	})

	r := gin.New()
	r.Use(middleware.AuthJWT(cfg))
	r.GET("/api/v1/attendance/scanner/view", middleware.RequirePermission(db, rdb, "attendance.scanner", "view", "SCANNER_ONLY"), func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Tampilan pemindai berhasil diakses", nil)
	})

	tokens, err := utils.GenerateTokenPair(cfg.JWTSecret, 15, 7, testUserID, testEmail, "SCANNER_OPERATOR", []string{"SCANNER_ONLY"})
	require.NoError(t, err)

	// Pastikan cache awal kosong
	_ = rdb.Client.Del(ctx, cacheKey).Err()

	// 1. Request pertama (Cache Miss -> DB Query -> Redis Cache Set)
	req1, _ := http.NewRequest(http.MethodGet, "/api/v1/attendance/scanner/view", nil)
	req1.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	// Verifikasi bahwa data tersimpan di Redis dan memiliki TTL
	cachedVal, err := rdb.Client.Get(ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.NotEmpty(t, cachedVal)

	ttl, err := rdb.Client.TTL(ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.True(t, ttl > 0 && ttl <= 5*time.Minute)

	// 2. Request kedua (Cache Hit -> membaca langsung dari Redis)
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/attendance/scanner/view", nil)
	req2.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
}

// Menguji evaluasi izin bagi pengguna yang memiliki lebih dari satu peran (multiple roles).
func TestMultiRolePermissionEvaluation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	rdb, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer rdb.Close()

	ctx := context.Background()
	testUserID := uuid.New()
	testEmail := fmt.Sprintf("multirole_%s@dcisp.internal", testUserID.String()[:8])

	// Role 1: SCANNER_OPERATOR (attendance.scanner:view,scan)
	scannerRoleID := uuid.MustParse("10000000-0000-0000-0000-000000000008")
	scannerScopeID := uuid.MustParse("20000000-0000-0000-0000-000000000006")

	// Role 2: SUPER_ADMIN (*.* on SYSTEM)
	superAdminRoleID := uuid.MustParse("10000000-0000-0000-0000-000000000001")
	systemScopeID := uuid.MustParse("20000000-0000-0000-0000-000000000001")

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Multi Role Tester', 'ACTIVE')
	`, testUserID, testEmail)
	require.NoError(t, err)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO user_roles (user_id, role_id, scope_id)
		VALUES ($1, $2, $3), ($1, $4, $5)
	`, testUserID, scannerRoleID, scannerScopeID, superAdminRoleID, systemScopeID)
	require.NoError(t, err)

	cacheKey := fmt.Sprintf("rbac:permissions:%s", testUserID.String())

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM user_roles WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
		_ = rdb.Client.Del(context.Background(), cacheKey).Err()
	})

	r := gin.New()
	r.Use(middleware.AuthJWT(cfg))
	r.GET("/api/v1/finance/ledger", middleware.RequirePermission(db, rdb, "finance.ledger", "view", "FINANCIAL_DATA"), func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Buku besar berhasil diakses", nil)
	})
	r.GET("/api/v1/attendance/scanner/view", middleware.RequirePermission(db, rdb, "attendance.scanner", "view", "SCANNER_ONLY"), func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Tampilan pemindai berhasil diakses", nil)
	})

	tokens, err := utils.GenerateTokenPair(cfg.JWTSecret, 15, 7, testUserID, testEmail, "SCANNER_OPERATOR", []string{"SCANNER_ONLY", "SYSTEM"})
	require.NoError(t, err)

	// Karena user juga memiliki peran SUPER_ADMIN (*.*), akses ke finance ledger HARUS diizinkan (200 OK)
	req1, _ := http.NewRequest(http.MethodGet, "/api/v1/finance/ledger", nil)
	req1.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	// Dan akses ke scanner view juga diizinkan (200 OK)
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/attendance/scanner/view", nil)
	req2.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
}
