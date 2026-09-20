package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji endpoint readiness probe untuk memastikan status layanan dan tidak adanya kebocoran detail error.
func TestReadinessProbeResponse(t *testing.T) {
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

	r := gin.New()
	r.GET("/health/readiness", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "UP"
		if err := db.Pool.Ping(ctx); err != nil {
			dbStatus = "DOWN"
		}

		redisStatus := "UP"
		if err := rdb.Client.Ping(ctx).Err(); err != nil {
			redisStatus = "DOWN"
		}

		status := http.StatusOK
		overallStatus := "UP"
		if dbStatus != "UP" || redisStatus != "UP" {
			status = http.StatusServiceUnavailable
			overallStatus = "DOWN"
		}

		c.JSON(status, gin.H{
			"status":    overallStatus,
			"database":  dbStatus,
			"redis":     redisStatus,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0",
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/health/readiness", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "UP", body["status"])
	assert.Equal(t, "UP", body["database"])
	assert.Equal(t, "UP", body["redis"])
}

// Menguji helper respon Conflict 409, TooManyRequests 429, dan pencegahan kebocoran error pada InternalError 500.
func TestResponseHelpers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/test-conflict", func(c *gin.Context) {
		response.Conflict(c, "Data sudah ada dalam sistem")
	})

	r.GET("/test-rate-limit", func(c *gin.Context) {
		response.TooManyRequests(c, "Terlalu banyak permintaan, coba lagi nanti")
	})

	r.GET("/test-internal-error", func(c *gin.Context) {
		response.InternalError(c, "Terjadi kesalahan internal pada server")
	})

	// 1. Conflict 409
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/test-conflict", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusConflict, w1.Code)

	// 2. TooManyRequests 429
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/test-rate-limit", nil)
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)

	// 3. InternalError 500
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/test-internal-error", nil)
	r.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusInternalServerError, w3.Code)

	var errBody response.APIResponse
	err := json.Unmarshal(w3.Body.Bytes(), &errBody)
	require.NoError(t, err)
	assert.False(t, errBody.Success)
	assert.Equal(t, "Terjadi kesalahan internal pada server", errBody.Message)
	assert.Nil(t, errBody.Errors) // Tidak membocorkan detail teknis internal
}

// Menguji pencatatan jejak audit interceptor yang menangkap status code respon HTTP secara asinkron.
func TestAuditInterceptor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	r := gin.New()
	r.Use(middleware.AuditInterceptor(db))

	testActorID := uuid.New()
	testEmail := fmt.Sprintf("auditor_%s@dcisp.internal", testActorID.String()[:8])
	_, err = db.Pool.Exec(context.Background(), `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Auditor Tester', 'ACTIVE')
	`, testActorID, testEmail)
	require.NoError(t, err)

	testPath := fmt.Sprintf("/api/v1/test-audit-%s", testActorID.String()[:8])

	r.POST(testPath, func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, testActorID)
		response.Success(c, http.StatusCreated, "Data uji berhasil dibuat", nil)
	})

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM audit_logs WHERE resource_id = $1", testPath)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testActorID)
	})

	body := bytes.NewBuffer([]byte(`{"test": true}`))
	req, _ := http.NewRequest(http.MethodPost, testPath, body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Beri jeda sejenak untuk eksekusi worker asinkron
	time.Sleep(200 * time.Millisecond)

	var count int
	var newStateBytes []byte
	err = db.Pool.QueryRow(context.Background(), `
		SELECT COUNT(*), COALESCE(MAX(new_state::text), '')
		FROM audit_logs
		WHERE resource_id = $1
	`, testPath).Scan(&count, &newStateBytes)
	require.NoError(t, err)
	assert.True(t, count >= 1)
	assert.Contains(t, string(newStateBytes), "201")
}
