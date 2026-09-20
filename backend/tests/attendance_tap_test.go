package tests

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/modules/attendance"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menandatangani permintaan terminal uji sesuai skema kanonis device-auth
// (HMAC-SHA256 dengan kunci hash API terminal).
func signTestDeviceRequest(rawAPIKey, method, path, timestamp string, body []byte) (deviceToken, signature string) {
	hashSum := sha256.Sum256([]byte(rawAPIKey))
	deviceToken = hex.EncodeToString(hashSum[:])
	mac := hmac.New(sha256.New, []byte(deviceToken))
	mac.Write([]byte(method + path + timestamp + string(body)))
	signature = hex.EncodeToString(mac.Sum(nil))
	return deviceToken, signature
}

// Menyematkan header otentikasi perangkat yang sah pada permintaan HTTP pengujian.
func setTestDeviceHeaders(req *http.Request, terminalID, rawAPIKey, path string, body []byte) {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	_, signature := signTestDeviceRequest(rawAPIKey, http.MethodPost, path, timestamp, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Device-ID", terminalID)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-DCISP-Signature", signature)
}

// Menguji ingesti presensi terminal tap, debouncing 30 detik, klasifikasi keterlambatan 3-tier, dan sinkronisasi offline (T-033, T-034, T-035, T-042).
func TestAttendanceTerminalTapDebounceAndLateness(t *testing.T) {
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

	repo := attendance.NewRepository(db)
	service := attendance.NewService(repo, db, rdb, cfg, nil, nil)
	ctrl := attendance.NewController(service)

	r := gin.New()
	r.POST("/api/v1/attendance/terminal-tap", middleware.DeviceAuth(db, cfg), ctrl.HandleTerminalTap)
	r.POST("/api/v1/attendance/sync-offline", middleware.DeviceAuth(db, cfg), ctrl.SyncOffline)

	ctx := context.Background()

	// Daftarkan terminal uji dengan kunci API yang diketahui
	testTerminalID := fmt.Sprintf("TEST-GATE-%s", uuid.New().String()[:8])
	testDeviceKey := "kunci-rahasia-terminal-uji-2026-min16"
	_, err = service.RegisterDevice(ctx, &attendance.RegisterDeviceRequest{
		TerminalIdentifier: testTerminalID,
		DeviceType:         attendance.DeviceTypeESP32Terminal,
		LocationName:       "Gerbang Uji Audit",
		APIKey:             testDeviceKey,
		CurrentMode:        attendance.TerminalModeAuto,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM devices WHERE terminal_identifier = $1", testTerminalID)
	})

	// 1. Setup User akun dan Kartu Peserta
	testUserID := uuid.New()
	cardUID := "NFC-CARD-" + testUserID.String()[:8]
	testEmail := "tap_tester_" + testUserID.String()[:8] + "@dcisp.internal"

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Tap Tester Hero', 'ACTIVE')
	`, testUserID, testEmail)
	require.NoError(t, err)

	// Pastikan ada batch aktif
	batchID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO batches (id, batch_code, name, start_date, end_date, quota, status)
		VALUES ($1, $2, 'Batch Tap Test', '2026-01-01', '2026-12-31', 50, 'ACTIVE')
	`, batchID, "BATCH-TAP-"+batchID.String()[:6])
	require.NoError(t, err)

	// Petakan cardUID sebagai id_number di tabel interns
	internID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO interns (id, user_id, batch_id, id_number, status, join_date, end_date)
		VALUES ($1, $2, $3, $4, 'ACTIVE', '2026-01-01', '2026-12-31')
	`, internID, testUserID, batchID, cardUID)
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = rdb.Client.Del(context.Background(), fmt.Sprintf("attendance:debounce:%s", testUserID.String())).Err()
	})

	// 2. Uji Penolakan Kartu yang Tidak Terdaftar (404 Not Found & Audio card_unregistered)
	unregPayload := attendance.TerminalTapRequest{
		CardUID: "UNKNOWN-CARD-9999",
	}
	bodyUnreg, _ := json.Marshal(unregPayload)
	reqUnreg, _ := http.NewRequest(http.MethodPost, "/api/v1/attendance/terminal-tap", bytes.NewBuffer(bodyUnreg))
	setTestDeviceHeaders(reqUnreg, testTerminalID, testDeviceKey, "/api/v1/attendance/terminal-tap", bodyUnreg)
	wUnreg := httptest.NewRecorder()
	r.ServeHTTP(wUnreg, reqUnreg)

	assert.Equal(t, http.StatusNotFound, wUnreg.Code)
	assert.Contains(t, wUnreg.Body.String(), attendance.AudioCardUnregistered)

	// 3. Uji Presensi Masuk Tepat Waktu (ON_TIME, e.g. pukul 08:25 WIB, jadwal 08:30 + 10m grace)
	// Buat waktu jam 08:25:00 WIB (01:25:00 UTC)
	now := time.Now().In(attendance.WIB)
	onTime := time.Date(now.Year(), now.Month(), now.Day(), 8, 25, 0, 0, attendance.WIB).UTC()

	tapOnTime := attendance.TerminalTapRequest{
		CardUID:        cardUID,
		Timestamp:      onTime.Unix(),
		TerminalMode:   attendance.TerminalModeCheckIn,
		IdempotencyKey: fmt.Sprintf("on-time-%s-%d", testUserID.String()[:8], time.Now().UnixNano()),
	}
	bodyOnTime, _ := json.Marshal(tapOnTime)
	reqOnTime, _ := http.NewRequest(http.MethodPost, "/api/v1/attendance/terminal-tap", bytes.NewBuffer(bodyOnTime))
	setTestDeviceHeaders(reqOnTime, testTerminalID, testDeviceKey, "/api/v1/attendance/terminal-tap", bodyOnTime)
	wOnTime := httptest.NewRecorder()
	r.ServeHTTP(wOnTime, reqOnTime)

	assert.Equal(t, http.StatusOK, wOnTime.Code)
	var respOnTime attendance.TerminalTapResponse
	err = json.Unmarshal(wOnTime.Body.Bytes(), &respOnTime)
	require.NoError(t, err)
	assert.Equal(t, "SUCCESS", respOnTime.Status)
	assert.False(t, respOnTime.IsLate)
	assert.Equal(t, attendance.AudioCheckIn, respOnTime.AudioEvent)

	// 4. Uji Debouncing 30 Detik (BRULE-WF-013)
	// Tap ulang kartu yang sama sesaat setelahnya (< 30 detik)
	reqDebounce, _ := http.NewRequest(http.MethodPost, "/api/v1/attendance/terminal-tap", bytes.NewBuffer(bodyOnTime))
	setTestDeviceHeaders(reqDebounce, testTerminalID, testDeviceKey, "/api/v1/attendance/terminal-tap", bodyOnTime)
	wDebounce := httptest.NewRecorder()
	r.ServeHTTP(wDebounce, reqDebounce)

	assert.Equal(t, http.StatusOK, wDebounce.Code)
	var respDebounce attendance.TerminalTapResponse
	err = json.Unmarshal(wDebounce.Body.Bytes(), &respDebounce)
	require.NoError(t, err)
	assert.Equal(t, "DEBOUNCED", respDebounce.Status)
	assert.Equal(t, attendance.AudioTooFrequent, respDebounce.AudioEvent)

	// Hapus debounce di Redis untuk menguji kasus keterlambatan
	_ = rdb.Client.Del(ctx, fmt.Sprintf("attendance:debounce:%s", testUserID.String())).Err()

	// 5. Uji Klasifikasi Keterlambatan: Tier 2 (e.g. kedatangan pukul 08:48 WIB = terlambat 18 menit dari 08:30)
	lateTime := time.Date(now.Year(), now.Month(), now.Day(), 8, 48, 0, 0, attendance.WIB).UTC()
	tapLate := attendance.TerminalTapRequest{
		CardUID:        cardUID,
		Timestamp:      lateTime.Unix(),
		TerminalMode:   attendance.TerminalModeCheckIn,
		IdempotencyKey: fmt.Sprintf("late-%s-%d", testUserID.String()[:8], time.Now().UnixNano()),
	}
	bodyLate, _ := json.Marshal(tapLate)
	reqLate, _ := http.NewRequest(http.MethodPost, "/api/v1/attendance/terminal-tap", bytes.NewBuffer(bodyLate))
	setTestDeviceHeaders(reqLate, testTerminalID, testDeviceKey, "/api/v1/attendance/terminal-tap", bodyLate)
	wLate := httptest.NewRecorder()
	r.ServeHTTP(wLate, reqLate)

	assert.Equal(t, http.StatusOK, wLate.Code)
	var respLate attendance.TerminalTapResponse
	err = json.Unmarshal(wLate.Body.Bytes(), &respLate)
	require.NoError(t, err)
	assert.Equal(t, "SUCCESS", respLate.Status)
	assert.True(t, respLate.IsLate)
	assert.Equal(t, 2, respLate.LateTier) // Tier 2 (16-30 menit)
	assert.Equal(t, 18, respLate.LateMinutes)
	assert.Equal(t, attendance.AudioLate, respLate.AudioEvent)

	// 6. Uji Sinkronisasi Presensi Offline (T-042 - ON CONFLICT DO NOTHING)
	offlineKey := fmt.Sprintf("offline-%s-%d", testUserID.String()[:8], time.Now().UnixNano())
	offlinePayload := attendance.OfflineSyncRequest{
		Records: []attendance.OfflineSyncRecord{
			{
				TerminalIdentifier: "TERM-01",
				CardUID:            cardUID,
				Timestamp:          now.Unix() - 3600,
				EventType:          attendance.EventCheckIn,
				IdempotencyKey:     offlineKey,
			},
			{
				TerminalIdentifier: "TERM-01",
				CardUID:            cardUID,
				Timestamp:          now.Unix() - 3600,
				EventType:          attendance.EventCheckIn,
				IdempotencyKey:     offlineKey, // Duplikat key harus diskip tanpa error
			},
		},
	}
	bodyOffline, _ := json.Marshal(offlinePayload)
	reqOffline, _ := http.NewRequest(http.MethodPost, "/api/v1/attendance/sync-offline", bytes.NewBuffer(bodyOffline))
	setTestDeviceHeaders(reqOffline, testTerminalID, testDeviceKey, "/api/v1/attendance/sync-offline", bodyOffline)
	wOffline := httptest.NewRecorder()
	r.ServeHTTP(wOffline, reqOffline)

	assert.Equal(t, http.StatusOK, wOffline.Code)
	var respSync struct {
		Data attendance.OfflineSyncResponse `json:"data"`
	}
	err = json.Unmarshal(wOffline.Body.Bytes(), &respSync)
	require.NoError(t, err)
	assert.Equal(t, 1, respSync.Data.SyncedCount)
	assert.Equal(t, 1, respSync.Data.SkippedCount) // Idempotency check berhasil
}

// Menguji penegakan otentikasi perangkat pada endpoint terminal presensi (F-AUTH-01, SECURITY.md Section 1.3).
func TestDeviceAuthEnforcement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := attendance.NewRepository(db)
	service := attendance.NewService(repo, db, nil, cfg, nil, nil)
	ctrl := attendance.NewController(service)

	r := gin.New()
	r.POST("/api/v1/attendance/terminal-tap", middleware.DeviceAuth(db, cfg), ctrl.HandleTerminalTap)

	ctx := context.Background()

	testTerminalID := fmt.Sprintf("TEST-AUTH-%s", uuid.New().String()[:8])
	testDeviceKey := "kunci-uji-otentikasi-perangkat-2026"
	_, err = service.RegisterDevice(ctx, &attendance.RegisterDeviceRequest{
		TerminalIdentifier: testTerminalID,
		DeviceType:         attendance.DeviceTypeESP32Terminal,
		LocationName:       "Gerbang Uji Otentikasi",
		APIKey:             testDeviceKey,
		CurrentMode:        attendance.TerminalModeAuto,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM devices WHERE terminal_identifier = $1", testTerminalID)
	})

	buildTap := func() []byte {
		body, _ := json.Marshal(attendance.TerminalTapRequest{CardUID: "UNKNOWN-CARD-AUTH-PROBE"})
		return body
	}
	const tapPath = "/api/v1/attendance/terminal-tap"

	// 1. Tanpa header otentikasi sama sekali -> 401 dan tanpa mutasi
	body := buildTap()
	reqNoAuth, _ := http.NewRequest(http.MethodPost, tapPath, bytes.NewBuffer(body))
	reqNoAuth.Header.Set("Content-Type", "application/json")
	wNoAuth := httptest.NewRecorder()
	r.ServeHTTP(wNoAuth, reqNoAuth)
	assert.Equal(t, http.StatusUnauthorized, wNoAuth.Code)

	// 2. Tanda tangan salah -> 401
	reqBadSig, _ := http.NewRequest(http.MethodPost, tapPath, bytes.NewBuffer(body))
	setTestDeviceHeaders(reqBadSig, testTerminalID, testDeviceKey, tapPath, body)
	reqBadSig.Header.Set("X-DCISP-Signature", "deadbeef")
	wBadSig := httptest.NewRecorder()
	r.ServeHTTP(wBadSig, reqBadSig)
	assert.Equal(t, http.StatusUnauthorized, wBadSig.Code)

	// 3. Terminal tidak terdaftar -> 401
	reqUnknownDev, _ := http.NewRequest(http.MethodPost, tapPath, bytes.NewBuffer(body))
	setTestDeviceHeaders(reqUnknownDev, "TERMINAL-TIDAK-ADA", testDeviceKey, tapPath, body)
	wUnknownDev := httptest.NewRecorder()
	r.ServeHTTP(wUnknownDev, reqUnknownDev)
	assert.Equal(t, http.StatusUnauthorized, wUnknownDev.Code)

	// 4. Stempel waktu kedaluwarsa (>60 detik) -> 401 (anti-replay)
	expiredTS := strconv.FormatInt(time.Now().Unix()-300, 10)
	_, expiredSig := signTestDeviceRequest(testDeviceKey, http.MethodPost, tapPath, expiredTS, body)
	reqExpired, _ := http.NewRequest(http.MethodPost, tapPath, bytes.NewBuffer(body))
	reqExpired.Header.Set("Content-Type", "application/json")
	reqExpired.Header.Set("X-Device-ID", testTerminalID)
	reqExpired.Header.Set("X-Timestamp", expiredTS)
	reqExpired.Header.Set("X-DCISP-Signature", expiredSig)
	wExpired := httptest.NewRecorder()
	r.ServeHTTP(wExpired, reqExpired)
	assert.Equal(t, http.StatusUnauthorized, wExpired.Code)

	// 5. Tanda tangan valid -> lolos middleware (404 kartu tak terdaftar, bukan 401)
	reqValid, _ := http.NewRequest(http.MethodPost, tapPath, bytes.NewBuffer(body))
	setTestDeviceHeaders(reqValid, testTerminalID, testDeviceKey, tapPath, body)
	wValid := httptest.NewRecorder()
	r.ServeHTTP(wValid, reqValid)
	assert.Equal(t, http.StatusNotFound, wValid.Code)
	assert.Contains(t, wValid.Body.String(), attendance.AudioCardUnregistered)
}

// Menguji pembuatan dan validasi kode QR dinamis dengan masa berlaku ketat 30 detik (T-043, BRULE-WF-014).
func TestDynamicQRGenerationAndExpiration(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	rdb, err := database.NewRedisClient(cfg)
	require.NoError(t, err)
	defer rdb.Close()

	service := attendance.NewService(attendance.NewRepository(db), db, rdb, cfg, nil, nil)
	ctx := context.Background()

	testUserID := uuid.New()

	// 1. Buat token QR baru
	qrRes, err := service.GenerateDynamicQR(ctx, testUserID)
	require.NoError(t, err)
	assert.NotEmpty(t, qrRes.QRToken)
	assert.Equal(t, 30, qrRes.ExpiresInSeconds)

	// 2. Tap menggunakan token QR segar (harus valid secara verifikasi token)
	tapRes, err := service.ProcessTap(ctx, nil, &attendance.TerminalTapRequest{
		QRToken: qrRes.QRToken,
	})
	// Karena user belum ada di database, harus mengembalikan ErrCardUnregistered (artinya verifikasi QR lolos dan lanjut ke user lookup)
	assert.Equal(t, "ERROR", tapRes.Status)
	assert.Equal(t, attendance.AudioCardUnregistered, tapRes.AudioEvent)
}

// Menguji ketersediaan manifes katalog 15 audio master untuk terminal ESP32 (T-044).
func TestTerminalAudioCatalogManifest(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	service := attendance.NewService(nil, nil, nil, cfg, nil, nil)
	catalog, err := service.GetAudioCatalog(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "1.0.0", catalog.Version)
	assert.Len(t, catalog.Items, 15) // Wajib mencakup tepat 15 master event

	keys := make(map[string]bool)
	for _, item := range catalog.Items {
		keys[item.Key] = true
		assert.NotEmpty(t, item.Filename)
		assert.NotEmpty(t, item.SHA256Hash)
		assert.NotEmpty(t, item.URL)
	}

	assert.True(t, keys[attendance.AudioCheckIn])
	assert.True(t, keys[attendance.AudioLate])
	assert.True(t, keys[attendance.AudioTooFrequent])
	assert.True(t, keys[attendance.AudioCardUnregistered])
}
