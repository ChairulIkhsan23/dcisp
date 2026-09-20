package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/modules/identity"
	"dcisp/backend/internal/modules/performance"
	"dcisp/backend/internal/modules/projects"
	"dcisp/backend/internal/shared/eventbus"
	"dcisp/backend/internal/shared/response"
	"dcisp/backend/internal/shared/utils"
	"dcisp/backend/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Menguji rotasi refresh token, penolakan token lama yang sudah diputasi, dan pencabutan sesi saat logout (F-AUTH-02, SECURITY.md Section 1.2).
func TestRefreshRotationAndRevocation(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	repo := identity.NewRepository(db)
	service := identity.NewService(repo, cfg)
	ctx := context.Background()

	testUserID := uuid.New()
	testEmail := fmt.Sprintf("rotation_%s@dcisp.internal", testUserID.String()[:8])
	testPass := "RotationPass2026!"
	hash, err := utils.HashPassword(testPass)
	require.NoError(t, err)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, $3, 'Rotation Tester', 'ACTIVE')
	`, testUserID, testEmail, hash)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM refresh_sessions WHERE user_id = $1", testUserID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
	})

	// 1. Login menerbitkan pasangan token dan mendaftarkan sesi refresh
	tokens, _, err := service.Login(ctx, testEmail, testPass)
	require.NoError(t, err)

	// 2. Refresh valid memutar token: pasangan baru terbit
	rotated, err := service.RefreshToken(ctx, tokens.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, rotated.AccessToken)
	assert.NotEqual(t, tokens.RefreshToken, rotated.RefreshToken)

	// 3. Token lama yang sudah diputasi wajib ditolak (anti-reuse)
	_, err = service.RefreshToken(ctx, tokens.RefreshToken)
	require.Error(t, err)

	// 4. Logout mencabut sesi baru; refresh berikutnya wajib ditolak
	require.NoError(t, service.Logout(ctx, testUserID, rotated.RefreshToken))
	_, err = service.RefreshToken(ctx, rotated.RefreshToken)
	require.Error(t, err)

	// 5. Revokasi massal mencabut seluruh sesi pengguna
	tokens2, _, err := service.Login(ctx, testEmail, testPass)
	require.NoError(t, err)
	revokedCount, err := service.RevokeAllUserSessions(ctx, testUserID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, revokedCount, int64(1))
	_, err = service.RefreshToken(ctx, tokens2.RefreshToken)
	require.Error(t, err)
}

// Menguji kebijakan kompleksitas kata sandi sesuai SECURITY.md Section 1.1.
func TestPasswordComplexityPolicy(t *testing.T) {
	weakPasswords := []string{
		"pendek1!",       // kurang dari 8 karakter
		"alllowercase1!", // tanpa huruf besar
		"ALLUPPERCASE1!", // tanpa huruf kecil
		"TanpaAngka!!",   // tanpa angka
		"TanpaSpesial12", // tanpa karakter khusus
	}
	for _, weak := range weakPasswords {
		assert.Error(t, utils.ValidatePasswordComplexity(weak), "kata sandi lemah harus ditolak: %s", weak)
		_, err := utils.HashPassword(weak)
		assert.Error(t, err)
	}

	require.NoError(t, utils.ValidatePasswordComplexity("KuatPass2026!"))
	hash, err := utils.HashPassword("KuatPass2026!")
	require.NoError(t, err)
	assert.NotEmpty(t, hash)
}

// Menguji kebijakan CORS berbasis allowlist: origin tepercaya lolos, tidak dikenal ditolak (F-SEC-01).
func TestCORSAllowlistEnforcement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{AllowedOrigins: []string{"http://localhost:3000"}}

	newRouter := func() *gin.Engine {
		r := gin.New()
		r.Use(middleware.CORS(cfg))
		r.GET("/api/v1/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
		return r
	}

	// 1. Origin tepercaya menerima ACAO eksak + kredensial
	r := newRouter()
	reqAllowed, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	reqAllowed.Header.Set("Origin", "http://localhost:3000")
	wAllowed := httptest.NewRecorder()
	r.ServeHTTP(wAllowed, reqAllowed)
	assert.Equal(t, http.StatusOK, wAllowed.Code)
	assert.Equal(t, "http://localhost:3000", wAllowed.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", wAllowed.Header().Get("Access-Control-Allow-Credentials"))

	// 2. Origin tidak dikenal tidak menerima ACAO (deny) namun request non-preflight tetap diproses
	reqUnknown, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	reqUnknown.Header.Set("Origin", "http://evil.example")
	wUnknown := httptest.NewRecorder()
	r.ServeHTTP(wUnknown, reqUnknown)
	assert.Equal(t, http.StatusOK, wUnknown.Code)
	assert.Empty(t, wUnknown.Header().Get("Access-Control-Allow-Origin"))

	// 3. Preflight dari origin tidak dikenal ditolak 403
	reqPreflight, _ := http.NewRequest(http.MethodOptions, "/api/v1/ping", nil)
	reqPreflight.Header.Set("Origin", "http://evil.example")
	wPreflight := httptest.NewRecorder()
	r.ServeHTTP(wPreflight, reqPreflight)
	assert.Equal(t, http.StatusForbidden, wPreflight.Code)
}

// Menguji penyematan security headers HTTP termasuk HSTS khusus mode release (F-SEC-02).
func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	newRouter := func(mode string) *gin.Engine {
		r := gin.New()
		r.Use(middleware.SecurityHeaders(&config.Config{GinMode: mode}))
		r.GET("/api/v1/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
		return r
	}

	// Mode debug: header dasar ada, HSTS tidak ada (tidak mengganggu pengembangan lokal)
	rDebug := newRouter("debug")
	wDebug := httptest.NewRecorder()
	reqDebug, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rDebug.ServeHTTP(wDebug, reqDebug)
	assert.Equal(t, "DENY", wDebug.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", wDebug.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", wDebug.Header().Get("Referrer-Policy"))
	assert.Contains(t, wDebug.Header().Get("Content-Security-Policy"), "default-src 'self'")
	assert.Empty(t, wDebug.Header().Get("Strict-Transport-Security"))

	// Mode release: HSTS aktif
	rRelease := newRouter("release")
	wRelease := httptest.NewRecorder()
	reqRelease, _ := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	rRelease.ServeHTTP(wRelease, reqRelease)
	assert.Contains(t, wRelease.Header().Get("Strict-Transport-Security"), "max-age=31536000")
}

// Menguji jalur dispatch tunggal: satu event hanya menghasilkan tepat satu efek samping meskipun worker aktif (F-EVT-01).
func TestSingleDispatchNoDoubleProcessing(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	pool := worker.NewWorkerPool(cfg)
	require.NoError(t, pool.Start())
	defer pool.Shutdown()

	bus := eventbus.NewEventBus(pool)

	var counter int64
	bus.Subscribe("audit.probe_single", func(ctx context.Context, event eventbus.DomainEvent) error {
		atomic.AddInt64(&counter, 1)
		return nil
	})

	require.NoError(t, bus.Publish(context.Background(), eventbus.DomainEvent{
		Type:        "audit.probe_single",
		AggregateID: uuid.New().String(),
		Payload:     map[string]interface{}{"probe": true},
	}))

	// Beri jeda agar worker Asynq sempat memproses bila terjadi dispatch ganda
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
	}

	assert.Equal(t, int64(1), atomic.LoadInt64(&counter), "satu event wajib menghasilkan tepat satu efek samping")
}

// Menguji jalur produksi nyata: aksi review laporan APPROVE menerbitkan event yang
// diproses subscriber XP menjadi mutasi Project XP (F-EVT-02, STEP 24).
func TestRealEventPathReportReviewedToXP(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	bus := eventbus.NewEventBus(nil)
	perfRepo := performance.NewRepository(db)
	perfService := performance.NewService(perfRepo, db, nil, bus)
	_ = perfService

	projRepo := projects.NewRepository(db)
	projService := projects.NewService(projRepo, db, nil, nil, bus)
	ctx := context.Background()

	ownerID := uuid.New()
	memberID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'PM Event Path', 'ACTIVE'),
		       ($3, $4, 'hash_pass', 'Member Event Path', 'ACTIVE')
	`, ownerID, fmt.Sprintf("pm_evt_%s@dcisp.internal", ownerID.String()[:8]),
		memberID, fmt.Sprintf("mb_evt_%s@dcisp.internal", memberID.String()[:8]))
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM evidence WHERE submission_id IN (SELECT id FROM work_reports WHERE user_id = $1)", memberID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM work_reports WHERE user_id = $1", memberID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM tasks WHERE project_id IN (SELECT id FROM projects WHERE owner_id = $1)", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM project_teams WHERE project_id IN (SELECT id FROM projects WHERE owner_id = $1)", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM projects WHERE owner_id = $1", ownerID)
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id IN ($1, $2)", ownerID, memberID)
	})

	project, err := projService.CreateProject(ctx, ownerID, &projects.CreateProjectRequest{
		Title: "Proyek Jalur Event XP", Description: "Uji produksi event report reviewed",
		Visibility: projects.VisibilityPublic, Capacity: 2, Deadline: "2026-12-31",
	})
	require.NoError(t, err)

	_, err = db.Pool.Exec(ctx, `
		INSERT INTO project_teams (id, project_id, user_id, project_role, planned_contribution_pct, is_locked)
		VALUES (gen_random_uuid(), $1, $2, 'MEMBER', 100.00, FALSE)
	`, project.ID, memberID)
	require.NoError(t, err)

	task, err := projService.CreateTask(ctx, project.ID, &projects.CreateTaskRequest{
		AssigneeID: &memberID, Title: "Tugas Event XP", EstimatedHours: 4.0,
		DifficultyWeight: 4, Priority: projects.TaskPriorityHigh, Deadline: "2026-10-15",
	})
	require.NoError(t, err)

	_, err = projService.ChangeTaskStatus(ctx, task.ID, projects.TaskStatusInProgress)
	require.NoError(t, err)

	report, err := projService.SubmitWorkReport(ctx, task.ID, memberID, &projects.SubmitWorkReportRequest{
		ProgressPercentage: 100, WhatIDid: "Menyelesaikan tugas jalur event secara penuh",
		EvidenceType: projects.EvidenceTypeURL, EvidenceURLOrKey: "https://example.com/bukti-event-xp",
	})
	require.NoError(t, err)

	_, err = projService.ReviewWorkReport(ctx, report.ID, "APPROVE")
	require.NoError(t, err)

	// Tunggu pemrosesan subscriber asinkron, lalu verifikasi mutasi XP benar-benar tercatat
	expectedRef := fmt.Sprintf("task:completion:%s", task.ID.String())
	deadline := time.Now().Add(5 * time.Second)
	for {
		var count int
		_ = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM xp_transactions WHERE reference_event = $1`, expectedRef).Scan(&count)
		if count >= 1 || time.Now().After(deadline) {
			assert.GreaterOrEqual(t, count, 1, "event task.report_reviewed wajib menghasilkan tepat satu mutasi XP")
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	balance, err := perfRepo.GetRunningBalance(ctx, memberID, performance.XPSchemeProject)
	require.NoError(t, err)
	assert.Equal(t, 50, balance)
}

// Menguji penerbitan event rank.promoted saat mutasi XP melewati ambang rank (F-EVT-02).
func TestRankPromotedEventFiresOnThresholdCross(t *testing.T) {
	_ = godotenv.Overload("../../.env")
	cfg, err := config.LoadConfig()
	require.NoError(t, err)

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err)
	defer db.Close()

	bus := eventbus.NewEventBus(nil)
	perfRepo := performance.NewRepository(db)
	perfService := performance.NewService(perfRepo, db, nil, bus)
	ctx := context.Background()

	testUserID := uuid.New()
	_, err = db.Pool.Exec(ctx, `
		INSERT INTO users (id, email, password_hash, full_name, status)
		VALUES ($1, $2, 'hash_pass', 'Rank Event Tester', 'ACTIVE')
	`, testUserID, fmt.Sprintf("rankevt_%s@dcisp.internal", testUserID.String()[:8]))
	require.NoError(t, err)

	received := make(chan eventbus.DomainEvent, 4)
	bus.Subscribe("rank.promoted", func(ctx context.Context, event eventbus.DomainEvent) error {
		received <- event
		return nil
	})

	t.Cleanup(func() {
		_, _ = db.Pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", testUserID)
	})

	bigPoints := 300
	_, err = perfService.RecordXPMutation(ctx, &performance.RecordXPMutationRequest{
		UserID:         testUserID,
		EventTrigger:   "CHECK_IN_ON_TIME",
		PointsOverride: &bigPoints,
		ReferenceEvent: fmt.Sprintf("rank-event-probe-%s", testUserID.String()[:8]),
	})
	require.NoError(t, err)

	select {
	case evt := <-received:
		assert.Equal(t, "rank.promoted", evt.Type)
		assert.Equal(t, testUserID.String(), evt.Payload["user_id"])
		assert.NotEmpty(t, evt.Payload["rank_id"])
	case <-time.After(5 * time.Second):
		t.Fatal("event rank.promoted tidak pernah diterbitkan saat ambang rank terlampaui")
	}
}

// Menguji sanitasi error: detail driver database disamarkan, pesan validasi aman diteruskan (F-ERR-01).
func TestErrorSanitizationBoundary(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 1. Error beraroma driver database wajib disamarkan
	r := gin.New()
	r.GET("/leak", func(c *gin.Context) {
		response.SafeBadRequest(c, "Permintaan tidak dapat diproses", fmt.Errorf("gagal menyimpan: ERROR: duplicate key value violates unique constraint \"uq_x\" (SQLSTATE 23505)"))
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/leak", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.NotContains(t, w.Body.String(), "SQLSTATE")
	assert.NotContains(t, w.Body.String(), "duplicate key")
	assert.NotContains(t, w.Body.String(), "uq_x")
	assert.Contains(t, w.Body.String(), "Permintaan tidak dapat diproses")

	// 2. Pesan validasi statis yang aman tetap diteruskan apa adanya
	r2 := gin.New()
	r2.GET("/valid", func(c *gin.Context) {
		response.SafeBadRequest(c, "Fallback", fmt.Errorf("kuota peserta proyek telah penuh"))
	})
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/valid", nil)
	r2.ServeHTTP(w2, req2)
	assert.Contains(t, w2.Body.String(), "kuota peserta proyek telah penuh")
}
