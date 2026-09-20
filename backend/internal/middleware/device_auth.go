package middleware

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/shared/response"
	"dcisp/backend/internal/shared/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	// ContextDeviceIDKey menyimpan UUID terminal yang terautentikasi pada request context.
	ContextDeviceIDKey = "device_id"
	// DeviceSignatureToleranceSeconds adalah jendela waktu anti-replay (SECURITY.md Section 1.3).
	DeviceSignatureToleranceSeconds = 60
)

// Mengautentikasi terminal presensi IoT melalui tanda tangan HMAC-SHA256 atau token operator yang sah.
// Skema kanonis: Signature = HMAC-SHA256(device_token, HTTP_Method + Path + Timestamp + Payload),
// dengan device_token adalah hash API key terminal yang tersimpan (kolom devices.api_key_hash).
// Alternatif yang diterima: Bearer access token operator (API.md: Device Auth OR Operator Token).
func DeviceAuth(db *database.PostgresDB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Jalur 1: Token operator (Bearer access token)
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				if claims, err := utils.ValidateToken(cfg.JWTSecret, parts[1], utils.TokenTypeAccess); err == nil {
					c.Set(ContextUserIDKey, claims.UserID)
					c.Set(ContextEmailKey, claims.Email)
					c.Set(ContextRoleKey, claims.Role)
					c.Set(ContextRolesKey, claims.Roles)
					c.Set(ContextScopesKey, claims.Scopes)
					c.Next()
					return
				}
			}
		}

		// Jalur 2: Tanda tangan perangkat HMAC-SHA256
		deviceIdentifier := strings.TrimSpace(c.GetHeader("X-Device-ID"))
		timestampStr := strings.TrimSpace(c.GetHeader("X-Timestamp"))
		signature := strings.TrimSpace(c.GetHeader("X-DCISP-Signature"))

		if deviceIdentifier == "" || timestampStr == "" || signature == "" {
			response.Unauthorized(c, "Otentikasi terminal tidak lengkap (wajib: X-Device-ID, X-Timestamp, X-DCISP-Signature)")
			c.Abort()
			return
		}

		device, err := findDeviceByIdentifier(c.Request.Context(), db, deviceIdentifier)
		if err != nil || device == nil || !device.IsActive {
			response.Unauthorized(c, "Terminal tidak terdaftar atau tidak aktif")
			c.Abort()
			return
		}

		ts, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			response.Unauthorized(c, "Format X-Timestamp tidak valid")
			c.Abort()
			return
		}

		if math.Abs(float64(deviceNowUnix()-ts)) > DeviceSignatureToleranceSeconds {
			response.Unauthorized(c, "Stempel waktu kedaluwarsa, sinkronisasi ulang jam terminal")
			c.Abort()
			return
		}

		bodyBytes, err := readAndRestoreBody(c)
		if err != nil {
			response.Unauthorized(c, "Payload permintaan tidak dapat dibaca")
			c.Abort()
			return
		}

		if !verifyDeviceHMAC(device.APIKeyHash, c.Request.Method, c.Request.URL.Path, timestampStr, bodyBytes, signature) {
			response.Unauthorized(c, "Tanda tangan terminal tidak valid")
			c.Abort()
			return
		}

		c.Set(ContextDeviceIDKey, device.ID)
		c.Next()
	}
}

// deviceRecord adalah proyeksi minimal tabel devices untuk otentikasi terminal.
type deviceRecord struct {
	ID         uuid.UUID
	APIKeyHash string
	IsActive   bool
}

// Mencari terminal terdaftar berdasarkan pengenal terminalnya (devices.terminal_identifier).
func findDeviceByIdentifier(ctx context.Context, db *database.PostgresDB, identifier string) (*deviceRecord, error) {
	var dev deviceRecord
	err := db.Pool.QueryRow(ctx, `
		SELECT id, api_key_hash, is_active
		FROM devices
		WHERE terminal_identifier = $1
	`, identifier).Scan(&dev.ID, &dev.APIKeyHash, &dev.IsActive)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &dev, nil
}

// Mengembalikan waktu epoch server saat ini dalam detik.
func deviceNowUnix() int64 {
	return time.Now().Unix()
}

// Membaca body permintaan mentah dan mengembalikannya agar dapat di-bind ulang oleh handler.
func readAndRestoreBody(c *gin.Context) ([]byte, error) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return bodyBytes, nil
}

// Memverifikasi HMAC-SHA256 tanda tangan perangkat dengan perbandingan waktu-konstan.
func verifyDeviceHMAC(deviceTokenHex, httpMethod, path, timestamp string, body []byte, signatureHex string) bool {
	canonical := strings.ToUpper(httpMethod) + path + timestamp + string(body)
	mac := hmac.New(sha256.New, []byte(deviceTokenHex))
	mac.Write([]byte(canonical))
	expected := hex.EncodeToString(mac.Sum(nil))

	sigBytes, err := hex.DecodeString(strings.TrimSpace(signatureHex))
	if err != nil {
		return false
	}
	expBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}
	return hmac.Equal(sigBytes, expBytes)
}
