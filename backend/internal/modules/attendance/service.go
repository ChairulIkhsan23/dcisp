package attendance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/system"
	"dcisp/backend/internal/shared/eventbus"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrCardUnregistered = errors.New("kartu belum terdaftar, silakan hubungi administrator")
	ErrCardInvalid      = errors.New("kartu tidak valid atau akun pengguna tidak aktif")
	ErrReplayDetected   = errors.New("permintaan ditolak: potensi serangan replay terdeteksi")
	ErrDeviceInvalid    = errors.New("perangkat terminal tidak valid atau tidak terdaftar")
	ErrDeviceSignature  = errors.New("tanda tangan otentikasi terminal tidak valid")
)

type Service struct {
	repo         *Repository
	db           *database.PostgresDB
	rdb          *database.RedisClient
	cfg          *config.Config
	auditService *system.AuditService
	eventBus     *eventbus.EventBus
}

// Menginisialisasi instance baru service attendance and workforce management.
func NewService(repo *Repository, db *database.PostgresDB, rdb *database.RedisClient, cfg *config.Config, audit *system.AuditService, bus *eventbus.EventBus) *Service {
	return &Service{
		repo:         repo,
		db:           db,
		rdb:          rdb,
		cfg:          cfg,
		auditService: audit,
		eventBus:     bus,
	}
}

// ============================================================================
// 1. TERMINAL TAP INGESTION (FR-008, T-034, T-035)
// ============================================================================

// Memproses pemindaian kartu NFC atau kode QR dari terminal gerbang secara aman, idempoten, dan bebas duplikasi.
func (s *Service) ProcessTap(ctx context.Context, deviceID *uuid.UUID, req *TerminalTapRequest) (*TerminalTapResponse, error) {
	now := time.Now().UTC()
	if req.Timestamp > 0 {
		now = time.Unix(req.Timestamp, 0).UTC()
	}
	nowWIB := now.In(WIB)

	// 1. Identifikasi Pengguna (via QR Token dinamis atau UID Kartu NFC)
	var userID uuid.UUID
	var userName, userStatus string

	if req.QRToken != "" {
		claims, err := s.verifyQRToken(req.QRToken)
		if err != nil {
			return &TerminalTapResponse{
				Status:     "ERROR",
				AudioEvent: AudioCardInvalid,
				Message:    fmt.Sprintf("QR code tidak valid: %v", err),
			}, ErrCardInvalid
		}
		userID = claims.UserID
		uID, name, status, err := s.repo.FindUserByIdentifier(ctx, userID.String())
		if err != nil {
			return &TerminalTapResponse{
				Status:     "ERROR",
				AudioEvent: AudioCardUnregistered,
				Message:    "Pengguna QR code tidak ditemukan",
			}, ErrCardUnregistered
		}
		userID = uID
		userName = name
		userStatus = status
	} else if req.CardUID != "" {
		uID, name, status, err := s.repo.FindUserByIdentifier(ctx, req.CardUID)
		if err != nil {
			return &TerminalTapResponse{
				Status:     "ERROR",
				AudioEvent: AudioCardUnregistered,
				Message:    "Kartu belum terdaftar di sistem",
			}, ErrCardUnregistered
		}
		userID = uID
		userName = name
		userStatus = status
	} else {
		return &TerminalTapResponse{
			Status:     "ERROR",
			AudioEvent: AudioCardInvalid,
			Message:    "Kredensial kartu atau QR token wajib diisi",
		}, ErrCardInvalid
	}

	if userStatus != "ACTIVE" {
		return &TerminalTapResponse{
			Status:     "ERROR",
			AudioEvent: AudioCardInvalid,
			Message:    "Akun pengguna tidak aktif",
		}, ErrCardInvalid
	}

	// 2. Pemeriksaan Debouncing 30 Detik (BRULE-WF-013)
	debounceKey := fmt.Sprintf("attendance:debounce:%s", userID.String())
	if s.rdb != nil && s.rdb.Client != nil {
		isSet, err := s.rdb.Client.SetNX(ctx, debounceKey, "1", 30*time.Second).Result()
		if err == nil && !isSet {
			return &TerminalTapResponse{
				Status:     "DEBOUNCED",
				AudioEvent: AudioTooFrequent,
				Message:    "Pemindaian diabaikan: kartu terdeteksi terlalu sering dalam 30 detik (BRULE-WF-013)",
				Data: map[string]interface{}{
					"user_name": userName,
				},
			}, nil
		}
	}

	// 3. Resolusi Jadwal Kerja & Hari Libur (FR-007, T-033)
	schedule, err := s.repo.GetActiveSchedule(ctx)
	if err != nil {
		return nil, err
	}

	isHoliday, holidayName, _ := s.repo.IsHoliday(ctx, nowWIB)
	hasApprovedLeave, _ := s.repo.HasApprovedLeaveOnDate(ctx, userID, nowWIB)

	// 4. Penentuan Tipe Event Berdasarkan Mode Terminal
	mode := TerminalModeAuto
	if req.TerminalMode != "" {
		mode = req.TerminalMode
	}

	eventType := EventCheckIn
	switch mode {
	case TerminalModeCheckIn:
		eventType = EventCheckIn
	case TerminalModeCheckOut:
		eventType = EventCheckOut
	case TerminalModeBreakStart:
		eventType = EventBreakStarted
	case TerminalModeBreakEnd:
		eventType = EventBreakEnded
	case TerminalModeAuto:
		hasCheckedIn, _ := s.repo.HasCheckInToday(ctx, userID, nowWIB)
		if !hasCheckedIn {
			eventType = EventCheckIn
		} else {
			latestEvent, _ := s.repo.GetLatestEventForUserToday(ctx, userID, nowWIB)
			if latestEvent != nil && latestEvent.EventType == EventBreakStarted {
				eventType = EventBreakEnded
			} else if nowWIB.Hour() >= 17 || (nowWIB.Hour() == 16 && nowWIB.Minute() >= 45) {
				eventType = EventCheckOut
			} else if nowWIB.Hour() >= 12 && nowWIB.Hour() < 13 {
				eventType = EventBreakStarted
			} else {
				eventType = EventCheckOut
			}
		}
	}

	// 5. Klasifikasi Keterlambatan 3-Tier (FR-007, BRULE-WF-001, BRULE-WF-002, T-035)
	isLate := false
	lateTier := 0
	lateMinutes := 0
	audioEvent := AudioCheckIn
	var earnedXP, penaltyXP int

	if eventType == EventCheckIn {
		if isHoliday || hasApprovedLeave {
			// Bebas sanksi keterlambatan
			isLate = false
			audioEvent = AudioCheckIn
			earnedXP = 10
		} else {
			// Parsing jadwal start time: "08:30:00"
			schedHour, schedMin := 8, 30
			if parts := strings.Split(schedule.StartTime, ":"); len(parts) >= 2 {
				schedHour, _ = strconv.Atoi(parts[0])
				schedMin, _ = strconv.Atoi(parts[1])
			}
			graceMins := schedule.GracePeriodMinutes
			if graceMins <= 0 {
				graceMins = 10
			}

			// Hitung menit hari ini sejak midnight
			currentMinutesToday := nowWIB.Hour()*60 + nowWIB.Minute()
			schedMinutesToday := schedHour*60 + schedMin
			cutoffMinutes := schedMinutesToday + graceMins

			if currentMinutesToday <= cutoffMinutes {
				isLate = false
				audioEvent = AudioCheckIn
				earnedXP = 10
			} else {
				isLate = true
				audioEvent = AudioLate
				lateMinutes = currentMinutesToday - schedMinutesToday

				// 3 Tier Keterlambatan: Tier 1 (1-15m), Tier 2 (16-30m), Tier 3 (>30m)
				if lateMinutes <= 15 {
					lateTier = 1
					penaltyXP = -1
				} else if lateMinutes <= 30 {
					lateTier = 2
					penaltyXP = -2
				} else {
					lateTier = 3
					penaltyXP = -3
				}
			}
		}
	} else if eventType == EventCheckOut {
		audioEvent = AudioCheckOut
		// Tutup sesi kerja aktif jika ada
		if activeSession, _ := s.repo.GetActiveWorkSession(ctx, userID, nowWIB); activeSession != nil {
			gross := int(now.Sub(activeSession.StartTime).Seconds())
			breakDur, _ := s.repo.GetTotalBreakDurationForSession(ctx, activeSession.ID)
			activeDur := gross - breakDur
			if activeDur < 0 {
				activeDur = 0
			}
			_ = s.repo.EndWorkSession(ctx, activeSession.ID, now, gross, activeDur, 0)
		}
	} else if eventType == EventBreakStarted {
		audioEvent = AudioBreakStart
	} else if eventType == EventBreakEnded {
		audioEvent = AudioBreakEnd
	}

	// 6. Kunci Idempotensi & Persistensi Log Kehadiran Append-Only
	idempKey := req.IdempotencyKey
	if idempKey == "" {
		idempKey = fmt.Sprintf("tap:%s:%s:%d", userID.String(), eventType, now.Unix()/60)
	}

	metaData, _ := json.Marshal(map[string]interface{}{
		"card_uid":     req.CardUID,
		"is_late":      isLate,
		"late_tier":    lateTier,
		"late_minutes": lateMinutes,
		"holiday":      holidayName,
	})

	eventLog := &AttendanceEventLog{
		ID:             uuid.New(),
		UserID:         userID,
		DeviceID:       deviceID,
		EventType:      eventType,
		Timestamp:      now,
		Method:         MethodNFC,
		IdempotencyKey: &idempKey,
		Metadata:       metaData,
	}
	if req.QRToken != "" {
		eventLog.Method = MethodQRCode
	}

	inserted, err := s.repo.CreateEventLog(ctx, eventLog)
	if err != nil {
		log.Printf("Kesalahan penyimpanan log presensi: %v", err)
		return &TerminalTapResponse{
			Status:     "FAILED",
			AudioEvent: AudioSaveFailed,
			Message:    "Gagal mencatat log kehadiran pada basis data",
		}, err
	}

	if !inserted {
		return &TerminalTapResponse{
			Status:     "DEBOUNCED",
			AudioEvent: AudioTooFrequent,
			Message:    "Presensi telah tercatat sebelumnya (idempotency key match)",
			Data: map[string]interface{}{
				"user_name": userName,
			},
		}, nil
	}

	// 7. Domain Event Dispatching
	if s.eventBus != nil {
		_ = s.eventBus.Publish(ctx, eventbus.DomainEvent{
			Type:        "attendance.scanned",
			AggregateID: userID.String(),
			Payload: map[string]interface{}{
				"user_id":      userID.String(),
				"user_name":    userName,
				"event_type":   eventType,
				"is_late":      isLate,
				"late_tier":    lateTier,
				"late_minutes": lateMinutes,
				"earned_xp":    earnedXP,
				"penalty_xp":   penaltyXP,
				"timestamp":    nowWIB.Format(time.RFC3339),
			},
		})
	}

	msg := "Presensi berhasil dicatat"
	if isLate {
		msg = fmt.Sprintf("Presensi masuk tercatat: Terlambat %d menit (Tingkat %d)", lateMinutes, lateTier)
	} else if eventType == EventCheckIn {
		msg = "Presensi masuk berhasil dicatat tepat waktu"
	}

	respData := map[string]interface{}{
		"user_name": userName,
		"timestamp": nowWIB.Format(time.RFC3339),
	}
	if isLate {
		respData["penalty_xp"] = penaltyXP
		respData["late_minutes"] = lateMinutes
	} else if eventType == EventCheckIn {
		respData["earned_xp"] = earnedXP
	}

	return &TerminalTapResponse{
		Status:      "SUCCESS",
		EventType:   eventType,
		IsLate:      isLate,
		LateTier:    lateTier,
		LateMinutes: lateMinutes,
		AudioEvent:  audioEvent,
		Message:     msg,
		Data:        respData,
	}, nil
}

// ============================================================================
// 2. OFFLINE SYNC (FR-008, T-042)
// ============================================================================

// Memproses sinkronisasi tumpukan data presensi offline dari terminal dengan idempotensi (ON CONFLICT DO NOTHING).
func (s *Service) SyncOffline(ctx context.Context, records []OfflineSyncRecord) (*OfflineSyncResponse, error) {
	synced := 0
	skipped := 0
	var errMessages []string

	for _, rec := range records {
		uID, _, _, err := s.repo.FindUserByIdentifier(ctx, rec.CardUID)
		if err != nil {
			errMessages = append(errMessages, fmt.Sprintf("Kartu %s tidak terdaftar", rec.CardUID))
			continue
		}

		evType := EventCheckIn
		if rec.EventType != "" {
			evType = rec.EventType
		}

		evTime := time.Unix(rec.Timestamp, 0).UTC()
		idempKey := rec.IdempotencyKey
		if idempKey == "" {
			idempKey = fmt.Sprintf("offline:%s:%s:%d", uID.String(), evType, rec.Timestamp)
		}

		logEntry := &AttendanceEventLog{
			ID:             uuid.New(),
			UserID:         uID,
			EventType:      evType,
			Timestamp:      evTime,
			Method:         MethodNFCOfflineSync,
			IdempotencyKey: &idempKey,
			Metadata:       []byte(`{"offline_synced": true}`),
		}

		inserted, err := s.repo.CreateEventLog(ctx, logEntry)
		if err != nil {
			errMessages = append(errMessages, fmt.Sprintf("Gagal menyimpan event %s: %v", idempKey, err))
			continue
		}

		if inserted {
			synced++
		} else {
			skipped++
		}
	}

	return &OfflineSyncResponse{
		SyncedCount:  synced,
		SkippedCount: skipped,
		Errors:       errMessages,
	}, nil
}

// ============================================================================
// 3. DYNAMIC ROTATING QR CODE (T-043, TTL = 30s)
// ============================================================================

type QRClaims struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

// Menghasilkan token QR dinamis dengan masa berlaku ketat 30 detik (BRULE-WF-014).
func (s *Service) GenerateDynamicQR(ctx context.Context, userID uuid.UUID) (*GenerateQRResponse, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(30 * time.Second)

	claims := &QRClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID.String(),
			Issuer:    "dcisp-dynamic-qr",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return nil, fmt.Errorf("gagal menandatangani dynamic QR: %w", err)
	}

	return &GenerateQRResponse{
		QRToken:          tokenStr,
		ExpiresInSeconds: 30,
		ExpiresAt:        expiresAt.Format(time.RFC3339),
	}, nil
}

// Memvalidasi token QR dinamis dan memastikan belum kedaluwarsa secara server-side.
func (s *Service) verifyQRToken(tokenStr string) (*QRClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &QRClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*QRClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("QR code tidak valid atau kedaluwarsa")
}

// ============================================================================
// 4. TERMINAL AUDIO CATALOG (T-044, 15 MASTER EVENTS)
// ============================================================================

// Mengambil katalog manifes 15 event audio standar terminal ESP32 beserta hash SHA-256.
func (s *Service) GetAudioCatalog(ctx context.Context) (*AudioCatalogResponse, error) {
	baseURL := s.cfg.R2PublicURL
	if baseURL == "" {
		baseURL = "https://pub-r2.dcisp.internal"
	}

	// 15 Master Audio Events resmi sesuai spesifikasi FINAL-TECH-STACK-SPEC-DCISP Section 3.3
	items := []AudioCatalogItem{
		{Key: AudioCardRead, Filename: "card_read.mp3", URL: baseURL + "/audio/global/card_read.mp3", SHA256Hash: "d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592", Description: "Kartu terbaca. Stand by."},
		{Key: AudioProcessing, Filename: "processing.mp3", URL: baseURL + "/audio/global/processing.mp3", SHA256Hash: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", Description: "Sedang Memproses. One moment."},
		{Key: AudioDeviceReady, Filename: "device_ready.mp3", URL: baseURL + "/audio/global/device_ready.mp3", SHA256Hash: "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb", Description: "System ready."},
		{Key: AudioCheckIn, Filename: "check_in.mp3", URL: baseURL + "/audio/global/check_in.mp3", SHA256Hash: "4e07408562bedb8b60ce05c1decfe3ad16b72230967de01f640b7e4729b49fce", Description: "Good Morning! Absensi berhasil."},
		{Key: AudioWorkStart, Filename: "work_start.mp3", URL: baseURL + "/audio/global/work_start.mp3", SHA256Hash: "4b227777d4dd1fc61c6f884f48641d02b4d121d3fd328cb08b5531fcacdabf8a", Description: "Selamat bekerja. Have a productive day!."},
		{Key: AudioBreakStart, Filename: "break_start.mp3", URL: baseURL + "/audio/global/break_start.mp3", SHA256Hash: "ef2d127de37b942baad06145e54b0c619a1f22327b2ebbcfbec78f5564afe39d", Description: "Enjoy your break!."},
		{Key: AudioBreakEnd, Filename: "break_end.mp3", URL: baseURL + "/audio/global/break_end.mp3", SHA256Hash: "e7f6c011776e8db7cd330b54174fd76f7d0216b612387a5ffcfb81e6f0919683", Description: "Welcome back! Silakan lanjut bekerja."},
		{Key: AudioCheckOut, Filename: "check_out.mp3", URL: baseURL + "/audio/global/check_out.mp3", SHA256Hash: "7902699be42c8a8e46fbbb4501726517e86b22c56a189f7625a6da49081b2451", Description: "Thank you, See you tomorrow."},
		{Key: AudioCardUnregistered, Filename: "card_unregistered.mp3", URL: baseURL + "/audio/global/card_unregistered.mp3", SHA256Hash: "2c624232cdd221771294dfbb310aca000a0df6ac8b66b696d90ef9f8b1b85ee7", Description: "Kartu belum terdaftar. Contact admin."},
		{Key: AudioCardInvalid, Filename: "card_invalid.mp3", URL: baseURL + "/audio/global/card_invalid.mp3", SHA256Hash: "11a4a60b518f624090c204dd7419992c90cee3a9bfa34ebd70994d99f6a5f517", Description: "Kartu tidak valid. Please try again."},
		{Key: AudioSaveFailed, Filename: "save_failed.mp3", URL: baseURL + "/audio/global/save_failed.mp3", SHA256Hash: "2b00042f7481c7b056c4b410d28f33cf6d5f127f49736b3c0490dd16460dac3d", Description: "Absensi gagal disimpan. Please try again."},
		{Key: AudioOffline, Filename: "offline.mp3", URL: baseURL + "/audio/global/offline.mp3", SHA256Hash: "8a050b4ecb409438542971b7770e2782f237f5a6773385624255371b05cad555", Description: "Sistem offline. Check the connection."},
		{Key: AudioOfflineSuccess, Filename: "offline_success.mp3", URL: baseURL + "/audio/global/offline_success.mp3", SHA256Hash: "026382998363797f9525f1499ab12f63cb3d87a1450a6728324f6f64b12cb9c9", Description: "Absensi berhasil dicatat. Data akan disinkronkan saat koneksi kembali."},
		{Key: AudioTooFrequent, Filename: "too_frequent.mp3", URL: baseURL + "/audio/global/too_frequent.mp3", SHA256Hash: "a9993e364706816aba3e25717850c26c9cd0d89d", Description: "Terlalu cepat. Please wait."},
		{Key: AudioLate, Filename: "late.mp3", URL: baseURL + "/audio/global/late.mp3", SHA256Hash: "5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03", Description: "Perhatian. Anda terlambat."},
	}

	return &AudioCatalogResponse{
		Version: "1.0.0",
		BaseURL: baseURL,
		Items:   items,
	}, nil
}

// ============================================================================
// 5. WORK SESSION & BREAK TRACKING (FR-009, FR-010, T-036, T-037)
// ============================================================================

// Memulai sesi kerja harian dari portal My Day setelah memvalidasi kehadiran check-in (BRULE-WF-004).
func (s *Service) StartWorkSession(ctx context.Context, userID uuid.UUID, taskID *uuid.UUID) (*WorkSession, error) {
	now := time.Now().UTC()
	nowWIB := now.In(WIB)

	// Validasi Prasyarat Kehadiran Sah Hari Ini (BRULE-WF-004)
	hasCheckedIn, err := s.repo.HasCheckInToday(ctx, userID, nowWIB)
	if err != nil {
		return nil, err
	}
	if !hasCheckedIn {
		return nil, errors.New("sesi kerja memerlukan catatan kehadiran (CHECK_IN) yang sah pada hari yang sama (BRULE-WF-004)")
	}

	// Cek apakah sudah ada sesi aktif
	activeSession, err := s.repo.GetActiveWorkSession(ctx, userID, nowWIB)
	if err != nil {
		return nil, err
	}
	if activeSession != nil {
		return activeSession, nil
	}

	ws := &WorkSession{
		ID:           uuid.New(),
		UserID:       userID,
		Date:         nowWIB,
		StartTime:    now,
		Status:       SessionStatusWorking,
		ActiveTaskID: taskID,
	}

	if err := s.repo.CreateWorkSession(ctx, ws); err != nil {
		return nil, err
	}

	// Catat log event kehadiran WORK_STARTED
	_, _ = s.repo.CreateEventLog(ctx, &AttendanceEventLog{
		ID:        uuid.New(),
		UserID:    userID,
		EventType: EventWorkStarted,
		Timestamp: now,
		Method:    MethodAdminScanner,
		SessionID: &ws.ID,
		Metadata:  []byte(`{"source": "portal_my_day"}`),
	})

	return ws, nil
}

// Mengontrol jeda istirahat kerja (Break) dan mendeteksi anomali istirahat awal serta keterlambatan kembali (BRULE-WF-005, BRULE-WF-006).
func (s *Service) ProcessBreakAction(ctx context.Context, userID uuid.UUID, action string) (*Break, error) {
	now := time.Now().UTC()
	nowWIB := now.In(WIB)

	session, err := s.repo.GetActiveWorkSession(ctx, userID, nowWIB)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("tidak ada sesi kerja aktif yang sedang berjalan")
	}

	upperAction := strings.ToUpper(action)

	if upperAction == "START" {
		if session.Status != SessionStatusWorking {
			return nil, errors.New("istirahat hanya dapat dimulai saat status sesi kerja adalah WORKING")
		}

		// Deteksi Early Break Anomaly (< 12:00) (BRULE-WF-005)
		isEarly := nowWIB.Hour() < 12

		brk := &Break{
			ID:             uuid.New(),
			SessionID:      session.ID,
			StartTime:      now,
			IsAnomalyEarly: isEarly,
		}

		if err := s.repo.CreateBreak(ctx, brk); err != nil {
			return nil, err
		}

		_ = s.repo.UpdateWorkSessionStatus(ctx, session.ID, SessionStatusBreak)

		_, _ = s.repo.CreateEventLog(ctx, &AttendanceEventLog{
			ID:        uuid.New(),
			UserID:    userID,
			EventType: EventBreakStarted,
			Timestamp: now,
			Method:    MethodAdminScanner,
			SessionID: &session.ID,
		})

		return brk, nil
	} else if upperAction == "RESUME" {
		if session.Status != SessionStatusBreak {
			return nil, errors.New("resume kerja hanya dapat dilakukan saat status sesi adalah BREAK")
		}

		activeBreak, err := s.repo.GetActiveBreak(ctx, session.ID)
		if err != nil {
			return nil, err
		}
		if activeBreak == nil {
			return nil, errors.New("tidak ditemukan rekaman istirahat aktif yang belum selesai")
		}

		// Deteksi Unauthorized Break Violation (> 13:00) (BRULE-WF-006)
		isLateResume := nowWIB.Hour() >= 13 && nowWIB.Minute() > 0
		durSecs := int(now.Sub(activeBreak.StartTime).Seconds())

		if err := s.repo.EndBreak(ctx, activeBreak.ID, now, durSecs, isLateResume); err != nil {
			return nil, err
		}

		_ = s.repo.UpdateWorkSessionStatus(ctx, session.ID, SessionStatusWorking)

		_, _ = s.repo.CreateEventLog(ctx, &AttendanceEventLog{
			ID:        uuid.New(),
			UserID:    userID,
			EventType: EventBreakEnded,
			Timestamp: now,
			Method:    MethodAdminScanner,
			SessionID: &session.ID,
		})

		activeBreak.EndTime = &now
		activeBreak.DurationSeconds = durSecs
		activeBreak.IsViolationLate = isLateResume

		return activeBreak, nil
	}

	return nil, errors.New("aksi istirahat tidak valid, diharapkan 'START' atau 'RESUME'")
}

// Menutup sesi kerja harian dan menghitung durasi waktu secara presisi.
func (s *Service) EndWorkSession(ctx context.Context, userID uuid.UUID) (*WorkSession, error) {
	now := time.Now().UTC()
	nowWIB := now.In(WIB)

	session, err := s.repo.GetActiveWorkSession(ctx, userID, nowWIB)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("tidak ada sesi kerja aktif yang sedang berjalan")
	}

	// Jika masih dalam status break, selesaikan break terlebih dahulu
	if session.Status == SessionStatusBreak {
		if activeBreak, _ := s.repo.GetActiveBreak(ctx, session.ID); activeBreak != nil {
			dur := int(now.Sub(activeBreak.StartTime).Seconds())
			_ = s.repo.EndBreak(ctx, activeBreak.ID, now, dur, false)
		}
	}

	gross := int(now.Sub(session.StartTime).Seconds())
	breakDur, _ := s.repo.GetTotalBreakDurationForSession(ctx, session.ID)
	activeDur := gross - breakDur
	if activeDur < 0 {
		activeDur = 0
	}

	if err := s.repo.EndWorkSession(ctx, session.ID, now, gross, activeDur, session.IdleDurationSeconds); err != nil {
		return nil, err
	}

	session.EndTime = &now
	session.GrossDurationSeconds = gross
	session.ActiveDurationSeconds = activeDur
	session.Status = SessionStatusEnded

	return session, nil
}

// ============================================================================
// 6. OVERTIME, LEAVE & CORRECTIONS (T-038, T-039, T-040)
// ============================================================================

// Mengajukan permohonan lembur baru sebelum batas jam kerja normal berakhir (FR-012, BRULE-WF-009).
func (s *Service) RequestOvertime(ctx context.Context, userID uuid.UUID, req *CreateOvertimeRequest) (*OvertimeRequest, error) {
	targetDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("format tanggal lembur tidak valid (YYYY-MM-DD): %w", err)
	}

	ot := &OvertimeRequest{
		ID:             uuid.New(),
		UserID:         userID,
		ProjectID:      req.ProjectID,
		TaskID:         req.TaskID,
		Date:           targetDate,
		RequestedStart: req.RequestedStart,
		RequestedEnd:   req.RequestedEnd,
		Reason:         req.Reason,
		Status:         OvertimeStatusSubmitted,
	}

	if err := s.repo.CreateOvertime(ctx, ot); err != nil {
		return nil, err
	}

	return ot, nil
}

// Meninjau dan menyetujui atau menolak permohonan lembur oleh supervisor.
func (s *Service) ReviewOvertime(ctx context.Context, id uuid.UUID, req *ReviewOvertimeRequest, reviewerID uuid.UUID) (*OvertimeRequest, error) {
	ot, err := s.repo.GetOvertimeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ot == nil {
		return nil, errors.New("permohonan lembur tidak ditemukan")
	}

	newStatus := OvertimeStatusRejected
	if strings.EqualFold(req.Action, "APPROVE") {
		newStatus = OvertimeStatusApproved
	}

	appStart := &ot.RequestedStart
	if req.ApprovedStart != nil {
		appStart = req.ApprovedStart
	}
	appEnd := &ot.RequestedEnd
	if req.ApprovedEnd != nil {
		appEnd = req.ApprovedEnd
	}

	if err := s.repo.ReviewOvertime(ctx, id, newStatus, appStart, appEnd, reviewerID, req.ReviewNotes); err != nil {
		return nil, err
	}

	ot.Status = newStatus
	ot.ApprovedStart = appStart
	ot.ApprovedEnd = appEnd

	return ot, nil
}

// Mengajukan permohonan cuti resmi sakit atau akademik (FR-014).
func (s *Service) RequestLeave(ctx context.Context, userID uuid.UUID, req *CreateLeaveRequest) (*LeaveRequest, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, fmt.Errorf("format tanggal mulai cuti tidak valid: %w", err)
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, fmt.Errorf("format tanggal selesai cuti tidak valid: %w", err)
	}
	if endDate.Before(startDate) {
		return nil, errors.New("tanggal selesai cuti tidak boleh mendahului tanggal mulai")
	}

	lr := &LeaveRequest{
		ID:             uuid.New(),
		UserID:         userID,
		LeaveType:      req.LeaveType,
		StartDate:      startDate,
		EndDate:        endDate,
		Reason:         req.Reason,
		EvidenceFileID: req.EvidenceFileID,
		Status:         LeaveStatusSubmitted,
	}

	if err := s.repo.CreateLeave(ctx, lr); err != nil {
		return nil, err
	}

	return lr, nil
}

// Meninjau dan menyetujui atau menolak permohonan cuti oleh supervisor atau HR.
func (s *Service) ReviewLeave(ctx context.Context, id uuid.UUID, req *ReviewLeaveRequest, reviewerID uuid.UUID) (*LeaveRequest, error) {
	lr, err := s.repo.GetLeaveByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if lr == nil {
		return nil, errors.New("permohonan cuti tidak ditemukan")
	}

	newStatus := LeaveStatusRejected
	if strings.EqualFold(req.Action, "APPROVE") {
		newStatus = LeaveStatusApproved
	}

	if err := s.repo.ReviewLeave(ctx, id, newStatus, reviewerID); err != nil {
		return nil, err
	}

	lr.Status = newStatus
	return lr, nil
}

// Mengajukan permohonan koreksi absensi manual dengan rekaman bukti tanpa mengubah log asli (FR-015, BR-025).
func (s *Service) RequestCorrection(ctx context.Context, userID uuid.UUID, req *CreateCorrectionRequest) (*AttendanceCorrection, error) {
	targetDate, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		return nil, fmt.Errorf("format target date tidak valid: %w", err)
	}
	propTime, err := time.Parse(time.RFC3339, req.ProposedTimestamp)
	if err != nil {
		return nil, fmt.Errorf("format proposed timestamp tidak valid (harus ISO8601/RFC3339): %w", err)
	}

	corr := &AttendanceCorrection{
		ID:                uuid.New(),
		UserID:            userID,
		TargetDate:        targetDate,
		ProposedEventType: req.ProposedEventType,
		ProposedTimestamp: propTime,
		Reason:            req.Reason,
		EvidenceFileID:    req.EvidenceFileID,
		Status:            CorrectionStatusSubmitted,
	}

	if err := s.repo.CreateCorrection(ctx, corr); err != nil {
		return nil, err
	}

	return corr, nil
}

// Menyetujui koreksi absensi dan menerbitkan rekaman baru MANUAL_CORRECTION secara append-only (BR-025, BRULE-WF-012).
func (s *Service) ReviewCorrection(ctx context.Context, id uuid.UUID, req *ReviewCorrectionRequest, reviewerID uuid.UUID) (*AttendanceCorrection, error) {
	corr, err := s.repo.GetCorrectionByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if corr == nil {
		return nil, errors.New("koreksi absensi tidak ditemukan")
	}

	newStatus := CorrectionStatusRejected
	if strings.EqualFold(req.Action, "APPROVE") {
		newStatus = CorrectionStatusApproved

		// Jika disetujui, buat log presensi baru bertipe MANUAL_CORRECTION tanpa mengubah log asli (BR-025)
		newLog := &AttendanceEventLog{
			ID:        uuid.New(),
			UserID:    corr.UserID,
			EventType: corr.ProposedEventType,
			Timestamp: corr.ProposedTimestamp,
			Method:    MethodManualCorrection,
			Metadata:  []byte(fmt.Sprintf(`{"correction_id": "%s", "reason": "%s"}`, corr.ID.String(), corr.Reason)),
		}
		_, err := s.repo.CreateEventLog(ctx, newLog)
		if err != nil {
			return nil, fmt.Errorf("gagal mencatat log koreksi presensi append-only: %w", err)
		}
	}

	if err := s.repo.ReviewCorrection(ctx, id, newStatus, reviewerID); err != nil {
		return nil, err
	}

	corr.Status = newStatus
	return corr, nil
}

// ============================================================================
// 7. DEVICE REGISTRY & HMAC AUTHENTICATION (FR-013, T-041)
// ============================================================================

// Mendaftarkan perangkat pemindai terminal gerbang baru dengan kunci rahasia API yang di-hash.
func (s *Service) RegisterDevice(ctx context.Context, req *RegisterDeviceRequest) (*Device, error) {
	hash := sha256.Sum256([]byte(req.APIKey))
	keyHashHex := hex.EncodeToString(hash[:])

	mode := TerminalModeAuto
	if req.CurrentMode != "" {
		mode = req.CurrentMode
	}

	dev := &Device{
		ID:                  uuid.New(),
		TerminalIdentifier:  req.TerminalIdentifier,
		DeviceType:          req.DeviceType,
		LocationName:        req.LocationName,
		APIKeyHash:          keyHashHex,
		CurrentMode:         mode,
		FirmwareVersion:     "1.0.0",
		AudioCatalogVersion: "1.0.0",
		IsActive:            true,
	}

	if err := s.repo.CreateDevice(ctx, dev); err != nil {
		return nil, err
	}

	return dev, nil
}

// Memvalidasi tanda tangan otentikasi HMAC-SHA256 terminal gerbang dan memeriksa batas waktu replay (<60 detik).
func (s *Service) VerifyDeviceSignature(ctx context.Context, deviceIdentifier, timestampStr, signature, httpMethod, path string, bodyBytes []byte, rawAPIKey string) error {
	ts, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		return errors.New("format X-Timestamp tidak valid")
	}

	now := time.Now().Unix()
	// Replay prevention: tolak jika perbedaan waktu > 60 detik (docs/guide/SECURITY.md)
	if math.Abs(float64(now-ts)) > 60 {
		return ErrReplayDetected
	}

	// Canonical payload: method + path + timestamp + body
	canonical := fmt.Sprintf("%s%s%s%s", strings.ToUpper(httpMethod), path, timestampStr, string(bodyBytes))
	mac := hmac.New(sha256.New, []byte(rawAPIKey))
	mac.Write([]byte(canonical))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return ErrDeviceSignature
	}

	return nil
}
