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
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji pembatasan laju permintaan menggunakan algoritma sliding window Redis dan header Retry-After saat terlampaui.
func TestRedisSlidingWindowRateLimiter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	rdb, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer rdb.Close()

	testUser := uuid.New()
	scope := fmt.Sprintf("test-limit-%s", testUser.String()[:8])
	limit := 3
	window := 2 * time.Second

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, testUser)
		c.Next()
	})
	r.Use(middleware.RateLimiter(rdb, limit, window, scope))
	r.GET("/limited-resource", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "Akses berhasil", nil)
	})

	cacheKey := fmt.Sprintf("ratelimit:%s:%s", scope, testUser.String())
	t.Cleanup(func() {
		_ = rdb.Client.Del(context.Background(), cacheKey).Err()
	})

	// 1. Permintaan ke-1 hingga ke-3 harus berhasil (200 OK)
	for i := 1; i <= limit; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/limited-resource", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code, "Permintaan ke-%d harusnya sukses", i)
	}

	// 2. Permintaan ke-4 harus ditolak dengan status 429 Too Many Requests
	reqBlocked, _ := http.NewRequest(http.MethodGet, "/limited-resource", nil)
	wBlocked := httptest.NewRecorder()
	r.ServeHTTP(wBlocked, reqBlocked)

	assert.Equal(t, http.StatusTooManyRequests, wBlocked.Code)
	assert.NotEmpty(t, wBlocked.Header().Get("Retry-After"))

	// 3. Tunggu hingga window sliding berakhir, permintaan berikutnya harus kembali diizinkan (200 OK)
	time.Sleep(window + 100*time.Millisecond)

	reqAllowed, _ := http.NewRequest(http.MethodGet, "/limited-resource", nil)
	wAllowed := httptest.NewRecorder()
	r.ServeHTTP(wAllowed, reqAllowed)
	assert.Equal(t, http.StatusOK, wAllowed.Code)
}
