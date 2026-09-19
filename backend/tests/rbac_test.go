package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/shared/response"
	"dcisp/backend/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScannerOperatorIsolation(t *testing.T) {
	// Rule BR-002: Scanner Operator is strictly isolated to scanner_only scope
	gin.SetMode(gin.TestMode)
	cfg := config.LoadConfig()

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Seed unique test scanner operator user and map to role SCANNER_OPERATOR
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

	r := gin.New()
	r.Use(middleware.AuthJWT(cfg))

	// Protected endpoint that requires finance/project access
	r.GET("/api/v1/finance/ledger", middleware.RequirePermission(db, "finance.ledger", "view", "FINANCIAL_DATA"), func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Ledger accessed", nil)
	})

	// Protected endpoint that scanner operator IS allowed to access
	r.GET("/api/v1/attendance/scanner/view", middleware.RequirePermission(db, "attendance.scanner", "view", "SCANNER_ONLY"), func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Scanner view accessed", nil)
	})

	tokens, err := utils.GenerateTokenPair(cfg.JWTSecret, 15, 7, testScannerUserID, testEmail, "SCANNER_OPERATOR", []string{"SCANNER_ONLY"})
	require.NoError(t, err)

	// 1. Scanner Operator attempts to access Finance Ledger -> MUST RETURN 403 FORBIDDEN (BR-002)
	req1, _ := http.NewRequest(http.MethodGet, "/api/v1/finance/ledger", nil)
	req1.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusForbidden, w1.Code)

	// 2. Scanner Operator accesses Scanner View -> MUST BE ALLOWED (200 OK)
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/attendance/scanner/view", nil)
	req2.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
}
