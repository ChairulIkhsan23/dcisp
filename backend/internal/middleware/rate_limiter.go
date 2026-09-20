package middleware

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"dcisp/backend/internal/database"
	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Membatasi laju permintaan menggunakan algoritma sliding window Redis berdasarkan identitas pengguna atau IP.
func RateLimiter(rdb *database.RedisClient, limit int, window time.Duration, scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil || rdb.Client == nil {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		now := time.Now()
		nowMillis := now.UnixNano() / int64(time.Millisecond)
		windowMillis := window.Milliseconds()
		clearBefore := nowMillis - windowMillis

		// Tentukan identitas klien: Utamakan ID pengguna terautentikasi, fallback ke Client IP
		identifier := c.ClientIP()
		if val, exists := c.Get(ContextUserIDKey); exists {
			if uid, ok := val.(uuid.UUID); ok {
				identifier = uid.String()
			}
		}

		key := fmt.Sprintf("ratelimit:%s:%s", scope, identifier)

		// Eksekusi transaksi pipeline atomic pada Redis
		pipe := rdb.Client.Pipeline()
		pipe.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(clearBefore, 10))
		cardCmd := pipe.ZCard(ctx, key)
		memberID := fmt.Sprintf("%d-%s", nowMillis, uuid.New().String()[:8])
		pipe.ZAdd(ctx, key, redis.Z{Score: float64(nowMillis), Member: memberID})
		pipe.Expire(ctx, key, window)

		_, err := pipe.Exec(ctx)
		if err != nil {
			log.Printf("Peringatan: Gagal memproses rate limiter pada Redis: %v", err)
			// Untuk scope auth sensitif: fail closed demi mencegah serangan brute force
			if scope == "auth" {
				response.TooManyRequests(c, "Layanan pembatas laju sementara tidak tersedia, silakan coba lagi")
				c.Abort()
				return
			}
			c.Next()
			return
		}

		currentCount := cardCmd.Val()
		if currentCount >= int64(limit) {
			retryAfterSec := int(window.Seconds())
			c.Header("Retry-After", strconv.Itoa(retryAfterSec))
			response.TooManyRequests(c, fmt.Sprintf("Batas permintaan terlampaui (%d permintaan per menit). Silakan tunggu %d detik.", limit, retryAfterSec))
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.FormatInt(int64(limit)-currentCount-1, 10))

		c.Next()
	}
}

// Membatasi laju permintaan untuk endpoint autentikasi maksimal 10 permintaan per menit.
func AuthRateLimiter(rdb *database.RedisClient) gin.HandlerFunc {
	return RateLimiter(rdb, 10, 1*time.Minute, "auth")
}

// Membatasi laju permintaan umum API maksimal 120 permintaan per menit.
func GeneralRateLimiter(rdb *database.RedisClient) gin.HandlerFunc {
	return RateLimiter(rdb, 120, 1*time.Minute, "general")
}
