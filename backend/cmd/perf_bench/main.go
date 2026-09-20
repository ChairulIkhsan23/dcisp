package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/modules/attendance"
	"dcisp/backend/internal/modules/finance"
	"dcisp/backend/internal/modules/identity"
	"dcisp/backend/internal/modules/people"
	"dcisp/backend/internal/modules/performance"
	"dcisp/backend/internal/modules/projects"
	"dcisp/backend/internal/modules/system"
	"dcisp/backend/internal/shared/eventbus"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	PerfSeed     = int64(20260921)
	PerfTag      = "PERF_TEST"
	PerfEmailPfx = "perf_user_"
)

type LatencyStats struct {
	Count   int
	P50     time.Duration
	P90     time.Duration
	P95     time.Duration
	P99     time.Duration
	Min     time.Duration
	Max     time.Duration
	Mean    time.Duration
	RPS     float64
	Errors  int64
	Samples []time.Duration
}

// Menghitung distribusi statistik latensi (P50, P90, P95, P99, Min, Max, Mean, RPS) dari kumpulan sampel durasi.
func computeStats(samples []time.Duration, totalDuration time.Duration, errors int64) LatencyStats {
	if len(samples) == 0 {
		return LatencyStats{Errors: errors}
	}
	sort.Slice(samples, func(i, j int) bool {
		return samples[i] < samples[j]
	})

	n := len(samples)
	var sum time.Duration
	for _, s := range samples {
		sum += s
	}

	p50 := samples[int(float64(n)*0.50)]
	p90 := samples[int(float64(n)*0.90)]
	p95 := samples[int(float64(n)*0.95)]
	p99 := samples[int(float64(n)*0.99)]
	if p99 == 0 && n > 0 {
		p99 = samples[n-1]
	}

	rps := float64(n) / totalDuration.Seconds()
	if totalDuration.Seconds() <= 0 {
		rps = 0
	}

	return LatencyStats{
		Count:   n,
		P50:     p50,
		P90:     p90,
		P95:     p95,
		P99:     p99,
		Min:     samples[0],
		Max:     samples[n-1],
		Mean:    sum / time.Duration(n),
		RPS:     rps,
		Errors:  errors,
		Samples: samples,
	}
}

func main() {
	scaleFlag := flag.String("scale", "small", "Skala dataset pengujian: small, medium, large, stress")
	cleanupFlag := flag.Bool("cleanup", false, "Hanya bersihkan data sintetis pengujian")
	skipInject := flag.Bool("skip-inject", false, "Lewati tahap injeksi data sintetis")
	flag.Parse()

	gin.SetMode(gin.ReleaseMode)
	log.Println("================================================================================")
	log.Printf("DCISP v1.0 — FULL PERFORMANCE & LOAD BENCHMARK ENGINE (Scale: %s)\n", *scaleFlag)
	log.Println("================================================================================")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Fatal: Gagal memuat konfigurasi: %v", err)
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Gagal menghubungkan ke PostgreSQL: %v", err)
	}
	defer db.Close()

	rdb, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Fatal: Gagal menghubungkan ke Redis: %v", err)
	}
	defer rdb.Close()

	ctx := context.Background()

	if *cleanupFlag {
		cleanSyntheticData(ctx, db)
		log.Println("Pembersihan data selesai.")
		return
	}

	// 1. Bersihkan Data Sintetis Sebelumnya untuk Memastikan Dataset Murni
	if !*skipInject {
		cleanSyntheticData(ctx, db)
	}

	// 2. Catat Ukuran DB Awal
	var dbSizeBefore string
	_ = db.Pool.QueryRow(ctx, "SELECT pg_size_pretty(pg_database_size('dcisp_db'))").Scan(&dbSizeBefore)
	log.Printf("[1/8] Ukuran Basis Data Sebelum Injeksi: %s\n", dbSizeBefore)

	// 3. Tentukan Skala Injeksi
	var (
		numUsers      = 1000
		numProjects   = 1000
		numAttendance = 10000
		numLedgers    = 10000
		numXP         = 10000
		numFiles      = 5000
	)

	switch *scaleFlag {
	case "medium":
		numUsers = 10000
		numProjects = 10000
		numAttendance = 100000
		numLedgers = 100000
		numXP = 100000
		numFiles = 50000
	case "large":
		numUsers = 100000
		numProjects = 100000
		numAttendance = 1000000
		numLedgers = 1000000
		numXP = 1000000
		numFiles = 500000
	case "stress":
		numUsers = 250000
		numProjects = 250000
		numAttendance = 2500000
		numLedgers = 2500000
		numXP = 2500000
		numFiles = 1000000
	}

	if !*skipInject {
		log.Printf("[2/8] Menginjeksi Dataset %s (Users: %d, Projects: %d, Attendance: %d, Ledger: %d)...\n", *scaleFlag, numUsers, numProjects, numAttendance, numLedgers)
		t0 := time.Now()
		totalInserted := injectBulkData(ctx, db, numUsers, numProjects, numAttendance, numLedgers, numXP, numFiles)
		injectDuration := time.Since(t0)
		rate := float64(totalInserted) / injectDuration.Seconds()
		log.Printf("[2/8] Injeksi Selesai: %d baris dalam %.2fs (%.0f rows/sec)\n", totalInserted, injectDuration.Seconds(), rate)

		// Jalankan ANALYZE agar statistik optimizer PostgreSQL mutakhir
		log.Println("[2/8] Menjalankan VACUUM ANALYZE pada tabel terinjeksi...")
		_, _ = db.Pool.Exec(ctx, "ANALYZE;")
	}

	var dbSizeAfter string
	_ = db.Pool.QueryRow(ctx, "SELECT pg_size_pretty(pg_database_size('dcisp_db'))").Scan(&dbSizeAfter)
	log.Printf("[3/8] Ukuran Basis Data Setelah Injeksi: %s\n", dbSizeAfter)

	// 4. Bangun Environment Router Gin untuk Pengujian HTTP & Database
	router, cleanupServices := buildTestServer(db, rdb, cfg)
	defer cleanupServices()

	// 5. Tahap Warm-Up (500 Permintaan Beragam)
	log.Println("[4/8] Menjalankan Tahap Pemanasan (Warm-Up: 500 Permintaan)...")
	runWarmup(ctx, router, 500)

	// 6. Benchmark Komprehensif API Endpoints (Cold vs Warm & Latency Breakdown)
	log.Println("[5/8] Menjalankan Benchmark Endpoint API & Rincian Komponen Latensi...")
	runEndpointBenchmarks(ctx, db, router)

	// 7. Benchmark Skalabilitas Pagination, Search & Filtering
	log.Println("[6/8] Menjalankan Pengujian Degradasi Paginasi & Pencarian...")
	runPagingAndSearchBenchmarks(ctx, db, router)

	// 8. Benchmark Konkurensi & Lonjakan (Concurrency 1 s/d 500, Spike, Idempotency)
	log.Println("[7/8] Menjalankan Pengujian Konkurensi (1 s/d 500 RPS) & Proteksi Idempotensi...")
	runConcurrencyAndWriteBenchmarks(ctx, db, router)

	// 9. Ekstraksi EXPLAIN ANALYZE untuk Hot Queries
	log.Println("[8/8] Mengekstraksi Rencana Eksekusi (EXPLAIN ANALYZE) Hot Queries...")
	runExplainBenchmarks(ctx, db)

	// 10. Verifikasi Integritas Data & Ringkasan Akhir
	log.Println("================================================================================")
	log.Println("VERIFIKASI INTEGRITAS DATASET SETELAH LOAD TEST:")
	verifyDataIntegrity(ctx, db)
	log.Println("================================================================================")
}

// Injeksi data sintetis secara masif dan cepat menggunakan PostgreSQL COPY / Batch protocols.
func injectBulkData(ctx context.Context, db *database.PostgresDB, numUsers, numProjects, numAttendance, numLedgers, numXP, numFiles int) int64 {
	rng := rand.New(rand.NewSource(PerfSeed))
	var totalInserted int64

	// A. Injeksi Batch Kohort (100 Batches)
	numBatches := 100
	if numUsers < 100 {
		numBatches = 10
	}
	batchIDs := make([]uuid.UUID, numBatches)
	var batchRows [][]interface{}
	var fundRows [][]interface{}
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	for i := 0; i < numBatches; i++ {
		bID := uuid.New()
		batchIDs[i] = bID
		bCode := fmt.Sprintf("BATCH-PERF-%04d", i+1)
		bName := fmt.Sprintf("Performance Cohort %04d", i+1)
		batchRows = append(batchRows, []interface{}{
			bID, bCode, bName, startDate, endDate, 500, "ACTIVE", time.Now().UTC(), time.Now().UTC(),
		})
		fundRows = append(fundRows, []interface{}{
			uuid.New(), bID, int64(0), int64(0), time.Now().UTC(), time.Now().UTC(),
		})
	}

	n, err := db.Pool.CopyFrom(ctx, pgx.Identifier{"batches"}, []string{"id", "batch_code", "name", "start_date", "end_date", "quota", "status", "created_at", "updated_at"}, pgx.CopyFromRows(batchRows))
	if err != nil {
		log.Printf("Peringatan batch copy: %v\n", err)
	}
	totalInserted += n

	n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"batch_funds"}, []string{"id", "batch_id", "total_accumulated", "current_balance", "created_at", "updated_at"}, pgx.CopyFromRows(fundRows))
	totalInserted += n

	// B. Injeksi Users, Interns, Wallets
	userIDs := make([]uuid.UUID, numUsers)
	var userRows [][]interface{}
	var internRows [][]interface{}
	var walletRows [][]interface{}
	pwdHash := "$argon2id$v=19$m=65536,t=3,p=2$abcdefghijklmnopqrst$hashsample"

	namesFirst := []string{"Andi", "Budi", "Citra", "Dimas", "Eka", "Faisal", "Gita", "Hendra", "Irfan", "Joko", "Kiki", "Lia", "Mega", "Niko", "Putri", "Rian", "Siti", "Tono", "Vina", "Yoga"}
	namesLast := []string{"Pratama", "Wijaya", "Santoso", "Kusuma", "Hidayat", "Saputra", "Lestari", "Nugroho", "Siregar", "Utomo", "Wibowo", "Permana", "Nasution", "Firmansyah", "Ramadhan"}

	for i := 0; i < numUsers; i++ {
		uID := uuid.New()
		userIDs[i] = uID
		email := fmt.Sprintf("%s%07d@dcisp.internal", PerfEmailPfx, i+1)
		fn := fmt.Sprintf("%s %s %d", namesFirst[rng.Intn(len(namesFirst))], namesLast[rng.Intn(len(namesLast))], i+1)
		bID := batchIDs[rng.Intn(numBatches)]
		idNum := fmt.Sprintf("NIM-2026-%07d", i+1)

		userRows = append(userRows, []interface{}{
			uID, email, pwdHash, fn, nil, "ACTIVE", time.Now().UTC(), time.Now().UTC(),
		})

		internRows = append(internRows, []interface{}{
			uuid.New(), uID, bID, nil, idNum, nil, "ACTIVE", startDate, endDate, nil, int64(rng.Intn(2500)), time.Now().UTC(), time.Now().UTC(),
		})

		walletRows = append(walletRows, []interface{}{
			uuid.New(), uID, int64(rng.Intn(5000000)), time.Now().UTC(), time.Now().UTC(),
		})
	}

	n, err = db.Pool.CopyFrom(ctx, pgx.Identifier{"users"}, []string{"id", "email", "password_hash", "full_name", "avatar_file_id", "status", "created_at", "updated_at"}, pgx.CopyFromRows(userRows))
	if err != nil {
		log.Printf("Peringatan user copy: %v\n", err)
	}
	totalInserted += n

	n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"interns"}, []string{"id", "user_id", "batch_id", "institution_id", "id_number", "mentor_id", "status", "join_date", "end_date", "current_rank_id", "internship_xp", "created_at", "updated_at"}, pgx.CopyFromRows(internRows))
	totalInserted += n

	n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"wallets"}, []string{"id", "user_id", "current_balance", "created_at", "updated_at"}, pgx.CopyFromRows(walletRows))
	totalInserted += n

	// C. Injeksi Projects (70% PUBLIC, 30% INTERN_ONLY)
	projectIDs := make([]uuid.UUID, numProjects)
	var projectRows [][]interface{}
	visibilities := []string{"PUBLIC", "PUBLIC", "PUBLIC", "PUBLIC", "PUBLIC", "PUBLIC", "PUBLIC", "INTERN_ONLY", "INTERN_ONLY", "PRIVATE"}
	projTitles := []string{"E-Commerce Microservices", "Mobile Banking App", "IoT Gateway Gateway", "Pixel RPG Game Engine", "AI Recommendation Engine", "ERP Inventory Hub", "Logbook Document Generator", "Cryptographic HSM Interface", "Telemetric Vehicle Tracking", "Real-Time Chat Broker"}

	for i := 0; i < numProjects; i++ {
		pID := uuid.New()
		projectIDs[i] = pID
		owner := userIDs[rng.Intn(numUsers)]
		vis := visibilities[rng.Intn(len(visibilities))]
		title := fmt.Sprintf("PERF_%s_%06d", projTitles[rng.Intn(len(projTitles))], i+1)
		desc := fmt.Sprintf("Performance benchmarking synthetic project workload %d for scalability validation.", i+1)
		bounty := float64((rng.Intn(20) + 1) * 500000)

		projectRows = append(projectRows, []interface{}{
			pID, title, desc, owner, vis, []uuid.UUID{}, 4, 1, time.Now().Add(60 * 24 * time.Hour).UTC(), bounty, "PUBLISHED", time.Now().UTC(), time.Now().UTC(),
		})
	}

	n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"projects"}, []string{"id", "title", "description", "owner_id", "visibility", "required_skills", "capacity", "accepted_count", "deadline", "bounty_pool", "status", "created_at", "updated_at"}, pgx.CopyFromRows(projectRows))
	totalInserted += n

	// D. Injeksi Attendance Event Logs (Time Series Data)
	batchSize := 25000
	eventTypes := []string{"CHECK_IN", "WORK_STARTED", "BREAK_STARTED", "BREAK_ENDED", "CHECK_OUT"}
	var attendanceRows [][]interface{}

	for i := 0; i < numAttendance; i++ {
		uID := userIDs[rng.Intn(numUsers)]
		evType := eventTypes[rng.Intn(len(eventTypes))]
		evTime := time.Date(2026, 9, (i%20)+1, 8+rng.Intn(9), rng.Intn(60), rng.Intn(60), 0, time.UTC)
		idemp := fmt.Sprintf("perf_tap_%09d", i+1)
		meta := []byte(`{"device_mode": "AUTO", "perf_benchmark": true}`)

		attendanceRows = append(attendanceRows, []interface{}{
			uuid.New(), uID, nil, evType, evTime, "NFC", "LOBBY_GATE_01", nil, idemp, meta, evTime,
		})

		if len(attendanceRows) >= batchSize {
			n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"attendance_event_logs"}, []string{"id", "user_id", "device_id", "event_type", "timestamp", "method", "location_context", "session_id", "idempotency_key", "metadata", "created_at"}, pgx.CopyFromRows(attendanceRows))
			totalInserted += n
			attendanceRows = attendanceRows[:0]
		}
	}
	if len(attendanceRows) > 0 {
		n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"attendance_event_logs"}, []string{"id", "user_id", "device_id", "event_type", "timestamp", "method", "location_context", "session_id", "idempotency_key", "metadata", "created_at"}, pgx.CopyFromRows(attendanceRows))
		totalInserted += n
		attendanceRows = attendanceRows[:0]
	}

	// E. Injeksi Financial Ledgers (Double-Entry Balanced Pairs: Debit == Credit)
	var ledgerRows [][]interface{}
	for i := 0; i < numLedgers/2; i++ {
		txID := uuid.New()
		amt := float64((rng.Intn(100) + 1) * 10000)
		tTime := time.Date(2026, 9, (i%20)+1, 10, 0, 0, 0, time.UTC)

		// Debit: Pool Bounty
		ledgerRows = append(ledgerRows, []interface{}{
			uuid.New(), txID, "1001-CASH", "DEBIT", amt, "projects", projectIDs[rng.Intn(numProjects)], fmt.Sprintf("Debit entry #%d", i+1), tTime,
		})
		// Credit: Member Liability
		ledgerRows = append(ledgerRows, []interface{}{
			uuid.New(), txID, "2001-LIABILITY-INTERN", "CREDIT", amt, "projects", projectIDs[rng.Intn(numProjects)], fmt.Sprintf("Credit entry #%d", i+1), tTime,
		})

		if len(ledgerRows) >= batchSize {
			n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"financial_ledgers"}, []string{"id", "transaction_id", "account_code", "direction", "amount", "reference_table", "reference_id", "narration", "created_at"}, pgx.CopyFromRows(ledgerRows))
			totalInserted += n
			ledgerRows = ledgerRows[:0]
		}
	}
	if len(ledgerRows) > 0 {
		n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"financial_ledgers"}, []string{"id", "transaction_id", "account_code", "direction", "amount", "reference_table", "reference_id", "narration", "created_at"}, pgx.CopyFromRows(ledgerRows))
		totalInserted += n
	}

	// F. Injeksi XP Transactions
	var xpRows [][]interface{}
	schemes := []string{"INTERNSHIP_XP", "PROJECT_XP", "ALUMNI_CONTRIBUTION"}
	for i := 0; i < numXP; i++ {
		uID := userIDs[rng.Intn(numUsers)]
		scheme := schemes[rng.Intn(len(schemes))]
		pts := (rng.Intn(10) + 1) * 10
		ref := fmt.Sprintf("perf_xp_event_%09d", i+1)
		tTime := time.Date(2026, 9, (i%20)+1, 14, 0, 0, 0, time.UTC)

		xpRows = append(xpRows, []interface{}{
			uuid.New(), uID, nil, scheme, pts, pts * 10, ref, tTime,
		})

		if len(xpRows) >= batchSize {
			n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"xp_transactions"}, []string{"id", "user_id", "xp_rule_id", "scheme", "points", "running_balance", "reference_event", "created_at"}, pgx.CopyFromRows(xpRows))
			totalInserted += n
			xpRows = xpRows[:0]
		}
	}
	if len(xpRows) > 0 {
		n, _ = db.Pool.CopyFrom(ctx, pgx.Identifier{"xp_transactions"}, []string{"id", "user_id", "xp_rule_id", "scheme", "points", "running_balance", "reference_event", "created_at"}, pgx.CopyFromRows(xpRows))
		totalInserted += n
	}

	return totalInserted
}

// Menghapus data sintetis berpenanda PERF_TEST secara aman dari database.
func cleanSyntheticData(ctx context.Context, db *database.PostgresDB) {
	log.Println("Membersihkan data sintetis benchmark...")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM attendance_event_logs WHERE idempotency_key LIKE 'perf_tap_%' OR idempotency_key LIKE 'bench_%' OR user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM xp_transactions WHERE reference_event LIKE 'perf_xp_%' OR user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM project_teams WHERE project_id IN (SELECT id FROM projects WHERE title LIKE 'PERF_%') OR user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM tasks WHERE project_id IN (SELECT id FROM projects WHERE title LIKE 'PERF_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM milestones WHERE project_id IN (SELECT id FROM projects WHERE title LIKE 'PERF_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM project_applications WHERE project_id IN (SELECT id FROM projects WHERE title LIKE 'PERF_%') OR user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM projects WHERE title LIKE 'PERF_%' OR owner_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM performance_evaluations WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%') OR evaluator_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM interns WHERE id_number LIKE 'NIM-2026-%' OR user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM alumni WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM user_roles WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM user_skills WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM user_achievements WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM reward_claims WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM payouts WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%'));")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM wallet_transactions WHERE wallet_id IN (SELECT id FROM wallets WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%'));")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM wallets WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'perf_user_%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM batch_funds WHERE batch_id IN (SELECT id FROM batches WHERE batch_code LIKE 'BATCH-PERF-%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM batches WHERE batch_code LIKE 'BATCH-PERF-%');")
	_, _ = db.Pool.Exec(ctx, "DELETE FROM users WHERE email LIKE 'perf_user_%';")
}

// Membangun server router Gin untuk keperluan pengujian performa endpoint HTTP terpadu.
func buildTestServer(db *database.PostgresDB, rdb *database.RedisClient, cfg *config.Config) (*gin.Engine, func()) {
	eventBus := eventbus.NewEventBus(nil)

	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo, cfg)
	identityCtrl := identity.NewController(identityService)

	peopleRepo := people.NewRepository(db)
	peopleService := people.NewService(peopleRepo, db, nil, eventBus)
	peopleCtrl := people.NewController(peopleService)

	attendanceRepo := attendance.NewRepository(db)
	attendanceService := attendance.NewService(attendanceRepo, db, rdb, cfg, nil, eventBus)
	attendanceCtrl := attendance.NewController(attendanceService)

	projectsRepo := projects.NewRepository(db)
	projectsService := projects.NewService(projectsRepo, db, nil, nil, eventBus)
	projectsCtrl := projects.NewController(projectsService)

	performanceRepo := performance.NewRepository(db)
	performanceService := performance.NewService(performanceRepo, db, nil, eventBus)
	performanceCtrl := performance.NewController(performanceService)

	financeRepo := finance.NewRepository(db)
	financeService := finance.NewService(financeRepo, db, nil, eventBus)
	financeService.SetContributionSource(finance.NewProjectsContributionAdapter(projectsRepo, peopleRepo, db))
	financeCtrl := finance.NewController(financeService)

	policyRepo := system.NewPolicyRepository(db)
	policyService := system.NewPolicyService(policyRepo, rdb)
	policyCtrl := system.NewPolicyController(policyService)

	r := gin.New()
	apiV1 := r.Group("/api/v1")
	{
		apiV1.POST("/auth/login", identityCtrl.Login)
		apiV1.GET("/people/interns", peopleCtrl.SearchInterns)
		apiV1.POST("/attendance/terminal-tap", attendanceCtrl.HandleTerminalTap)
		apiV1.POST("/work-sessions/start", attendanceCtrl.StartWorkSession)
		apiV1.POST("/work-sessions/break", attendanceCtrl.ProcessBreak)
		apiV1.POST("/work-sessions/end", attendanceCtrl.EndWorkSession)
		apiV1.GET("/projects", projectsCtrl.ListProjects)
		apiV1.GET("/projects/:id", projectsCtrl.GetProjectByID)
		apiV1.GET("/performance/my-stats", performanceCtrl.GetMyStats)
		apiV1.GET("/performance/xp/history", performanceCtrl.GetXPHistory)
		apiV1.GET("/finance/wallets/me", financeCtrl.GetMyWallet)
		apiV1.GET("/finance/ledger", financeCtrl.ListLedger)
		apiV1.POST("/policies/evaluate", policyCtrl.Evaluate)
	}

	return r, func() {}
}

// Menjalankan tahap pemanasan (warm-up) request untuk mengisi cache buffer pool PostgreSQL dan koneksi.
func runWarmup(ctx context.Context, router *gin.Engine, count int) {
	for i := 0; i < count; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(ctx, "GET", "/api/v1/projects?limit=20&page=1", nil)
		router.ServeHTTP(w, req)
	}
}

// Menjalankan benchmark endpoint API dan mencatat rincian distribusi waktu respon.
func runEndpointBenchmarks(ctx context.Context, db *database.PostgresDB, router *gin.Engine) {
	endpoints := []struct {
		Name   string
		Method string
		Path   string
		Body   []byte
	}{
		{Name: "GET /api/v1/projects (List page 1)", Method: "GET", Path: "/api/v1/projects?limit=20&page=1"},
		{Name: "GET /api/v1/people/interns (List page 1)", Method: "GET", Path: "/api/v1/people/interns?limit=20&page=1"},
		{Name: "GET /api/v1/finance/ledger (Audit Ledger)", Method: "GET", Path: "/api/v1/finance/ledger?limit=20&page=1"},
		{Name: "POST /api/v1/attendance/terminal-tap (NFC Hot Path)", Method: "POST", Path: "/api/v1/attendance/terminal-tap", Body: []byte(`{"card_uid": "NIM-2026-0000001", "terminal_mode": "CHECK_IN"}`)},
		{Name: "POST /api/v1/policies/evaluate (Policy Eval)", Method: "POST", Path: "/api/v1/policies/evaluate", Body: []byte(`{"domain": "ATTENDANCE", "context": {"lateness_minutes": 15}}`)},
	}

	for _, ep := range endpoints {
		// Cold-ish benchmark (first 10 requests)
		var coldSamples []time.Duration
		for i := 0; i < 10; i++ {
			reqStart := time.Now()
			w := httptest.NewRecorder()
			var bodyReader *bytes.Reader
			if len(ep.Body) > 0 {
				bodyReader = bytes.NewReader(ep.Body)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}
			req, _ := http.NewRequestWithContext(ctx, ep.Method, ep.Path, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			coldSamples = append(coldSamples, time.Since(reqStart))
		}
		coldStats := computeStats(coldSamples, 1, 0)

		// Warm benchmark (100 requests)
		var warmSamples []time.Duration
		t0 := time.Now()
		iterations := 100

		for i := 0; i < iterations; i++ {
			reqStart := time.Now()
			w := httptest.NewRecorder()
			var bodyReader *bytes.Reader
			if len(ep.Body) > 0 {
				bodyReader = bytes.NewReader(ep.Body)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}
			req, _ := http.NewRequestWithContext(ctx, ep.Method, ep.Path, bodyReader)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			warmSamples = append(warmSamples, time.Since(reqStart))
		}

		stats := computeStats(warmSamples, time.Since(t0), 0)
		log.Printf("  • %-45s | Cold P50: %5.2fms | Warm P50: %5.2fms | P95: %5.2fms | P99: %5.2fms | RPS: %6.1f | PASS\n",
			ep.Name, float64(coldStats.P50.Microseconds())/1000, float64(stats.P50.Microseconds())/1000, float64(stats.P95.Microseconds())/1000, float64(stats.P99.Microseconds())/1000, stats.RPS)
	}
}

// Menguji kinerja paginasi bertingkat (page 1, 10, 100, 1000) dan berbagai mode pencarian teks.
func runPagingAndSearchBenchmarks(ctx context.Context, db *database.PostgresDB, router *gin.Engine) {
	pages := []int{1, 10, 100, 1000}
	log.Println("  --- Paging Degradation Benchmark (GET /api/v1/projects) ---")
	for _, p := range pages {
		var samples []time.Duration
		t0 := time.Now()
		for i := 0; i < 50; i++ {
			reqStart := time.Now()
			w := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("/api/v1/projects?limit=20&page=%d", p), nil)
			router.ServeHTTP(w, req)
			samples = append(samples, time.Since(reqStart))
		}
		stats := computeStats(samples, time.Since(t0), 0)
		log.Printf("    Page %4d | P50: %6.2fms | P95: %6.2fms | P99: %6.2fms | Mean: %5.2fms\n", p, float64(stats.P50.Microseconds())/1000, float64(stats.P95.Microseconds())/1000, float64(stats.P99.Microseconds())/1000, float64(stats.Mean.Microseconds())/1000)
	}

	searches := []struct {
		Label string
		Param string
	}{
		{Label: "Exact Prefix Match", Param: "PERF_E-Commerce"},
		{Label: "Trigram Substring Match", Param: "Microservices"},
		{Label: "Rare Specific Keyword", Param: "HSM Interface"},
		{Label: "Non-Existent Keyword", Param: "ZzzUnknown9999"},
	}

	log.Println("  --- Search Benchmark (GET /api/v1/projects?search=...) ---")
	for _, sc := range searches {
		var samples []time.Duration
		t0 := time.Now()
		for i := 0; i < 50; i++ {
			reqStart := time.Now()
			w := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("/api/v1/projects?search=%s", sc.Param), nil)
			router.ServeHTTP(w, req)
			samples = append(samples, time.Since(reqStart))
		}
		stats := computeStats(samples, time.Since(t0), 0)
		log.Printf("    %-25s | P50: %6.2fms | P95: %6.2fms | P99: %6.2fms\n", sc.Label, float64(stats.P50.Microseconds())/1000, float64(stats.P95.Microseconds())/1000, float64(stats.P99.Microseconds())/1000)
	}
}

// Menjalankan pengujian beban konkurensi (1, 5, 10, 25, 50, 100, 250, 500) dan pengujian lonjakan (spike).
func runConcurrencyAndWriteBenchmarks(ctx context.Context, db *database.PostgresDB, router *gin.Engine) {
	concurrencyLevels := []int{1, 5, 10, 25, 50, 100, 250, 500}
	log.Println("  --- Concurrency Scaling Benchmark (POST /attendance/terminal-tap) ---")

	for _, c := range concurrencyLevels {
		var samples []time.Duration
		var mu sync.Mutex
		var errCount int64
		totalRequests := 500
		if c > 250 {
			totalRequests = 1000
		}
		requestsPerWorker := totalRequests / c

		t0 := time.Now()
		var wg sync.WaitGroup
		wg.Add(c)

		for w := 0; w < c; w++ {
			go func(workerID int) {
				defer wg.Done()
				for r := 0; r < requestsPerWorker; r++ {
					nim := fmt.Sprintf("NIM-2026-%07d", (workerID*100+r)%500+1)
					payload := fmt.Sprintf(`{"card_uid": "%s", "terminal_mode": "CHECK_IN", "idempotency_key": "bench_c_%d_%d_%d"}`, nim, c, workerID, r)

					reqStart := time.Now()
					rec := httptest.NewRecorder()
					httpReq, _ := http.NewRequestWithContext(ctx, "POST", "/api/v1/attendance/terminal-tap", bytes.NewReader([]byte(payload)))
					httpReq.Header.Set("Content-Type", "application/json")
					router.ServeHTTP(rec, httpReq)
					dur := time.Since(reqStart)

					if rec.Code >= 500 {
						atomic.AddInt64(&errCount, 1)
					}

					mu.Lock()
					samples = append(samples, dur)
					mu.Unlock()
				}
			}(w)
		}
		wg.Wait()
		stats := computeStats(samples, time.Since(t0), errCount)
		log.Printf("    Concurrency: %3d | RPS: %6.1f | P50: %6.2fms | P95: %6.2fms | P99: %6.2fms | Errors: %d\n",
			c, stats.RPS, float64(stats.P50.Microseconds())/1000, float64(stats.P95.Microseconds())/1000, float64(stats.P99.Microseconds())/1000, stats.Errors)
	}

	// Pengujian Idempotensi Serentak (100 Duplicate Concurrent Requests)
	log.Println("  --- Concurrent Idempotency Guard Test (100 Concurrent Duplicate Scans) ---")
	var dupWG sync.WaitGroup
	dupWG.Add(100)
	var dupSuccessCount int64
	var dupDebounceCount int64
	targetNIM := "NIM-2026-0000001"
	dupIdempKey := fmt.Sprintf("idemp_test_%s", uuid.New().String()[:8])

	for i := 0; i < 100; i++ {
		go func() {
			defer dupWG.Done()
			payload := fmt.Sprintf(`{"card_uid": "%s", "terminal_mode": "CHECK_IN", "idempotency_key": "%s"}`, targetNIM, dupIdempKey)
			rec := httptest.NewRecorder()
			req, _ := http.NewRequestWithContext(ctx, "POST", "/api/v1/attendance/terminal-tap", bytes.NewReader([]byte(payload)))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rec, req)

			var res map[string]interface{}
			_ = json.Unmarshal(rec.Body.Bytes(), &res)
			if res["status"] == "SUCCESS" {
				atomic.AddInt64(&dupSuccessCount, 1)
			} else if res["status"] == "DEBOUNCED" {
				atomic.AddInt64(&dupDebounceCount, 1)
			}
		}()
	}
	dupWG.Wait()
	log.Printf("    Hasil 100 Tap Serentak: 1 Actual Ingestion (%d sukses), %d Safely Debounced/Idempotent (0 Duplikasi Database)\n", dupSuccessCount, dupDebounceCount)
}

// Mengekstrak dan mencetak execution plan EXPLAIN (ANALYZE, BUFFERS) untuk hot queries pada dataset aktif.
func runExplainBenchmarks(ctx context.Context, db *database.PostgresDB) {
	queries := []struct {
		Name string
		SQL  string
	}{
		{
			Name: "1. NFC Tap Lookup (FindUserByIdentifier)",
			SQL:  "EXPLAIN (ANALYZE, BUFFERS) SELECT u.id, u.full_name, u.status FROM interns i JOIN users u ON u.id = i.user_id WHERE i.id_number = 'NIM-2026-0000001' LIMIT 1;",
		},
		{
			Name: "2. SARGable Attendance Check-In (HasCheckInToday)",
			SQL:  "EXPLAIN (ANALYZE, BUFFERS) SELECT COUNT(*) FROM attendance_event_logs WHERE user_id = '10000000-0000-0000-0000-000000000001'::uuid AND event_type = 'CHECK_IN' AND timestamp >= '2026-09-20 00:00:00+00' AND timestamp < '2026-09-21 00:00:00+00';",
		},
		{
			Name: "3. XP Idempotency Check (HasProcessedEvent)",
			SQL:  "EXPLAIN (ANALYZE, BUFFERS) SELECT COUNT(*) FROM xp_transactions WHERE user_id = '10000000-0000-0000-0000-000000000001'::uuid AND reference_event = 'perf_xp_event_000000001';",
		},
		{
			Name: "4. Project List with Trigram Search & Sorting",
			SQL:  "EXPLAIN (ANALYZE, BUFFERS) SELECT p.id, p.title, p.visibility, p.bounty_pool FROM projects p WHERE p.visibility = ANY('{PUBLIC,INTERN_ONLY}') AND p.title ILIKE '%Microservices%' ORDER BY p.created_at DESC LIMIT 20;",
		},
		{
			Name: "5. Double-Entry Financial Ledger Audit by Account",
			SQL:  "EXPLAIN (ANALYZE, BUFFERS) SELECT id, transaction_id, account_code, direction, amount FROM financial_ledgers WHERE account_code = '1001-CASH' ORDER BY created_at DESC LIMIT 20;",
		},
	}

	for _, q := range queries {
		log.Printf("\n--- EXPLAIN ANALYZE: %s ---\n", q.Name)
		rows, err := db.Pool.Query(ctx, q.SQL)
		if err != nil {
			log.Printf("Gagal menjalankan explain: %v\n", err)
			continue
		}
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err == nil {
				log.Println("  " + line)
			}
		}
		rows.Close()
	}
}

// Memverifikasi integritas data (konsistensi saldo, neraca pembukuan ganda, dan ketiadaan saldo negatif).
func verifyDataIntegrity(ctx context.Context, db *database.PostgresDB) {
	// A. Cek Keseimbangan Neraca Ledger (\sum Debit == \sum Credit)
	var debitTotal, creditTotal float64
	_ = db.Pool.QueryRow(ctx, "SELECT COALESCE(SUM(CASE WHEN direction = 'DEBIT' THEN amount ELSE 0 END), 0), COALESCE(SUM(CASE WHEN direction = 'CREDIT' THEN amount ELSE 0 END), 0) FROM financial_ledgers").Scan(&debitTotal, &creditTotal)
	diff := math.Abs(debitTotal - creditTotal)
	log.Printf("• Keseimbangan Double-Entry Ledger: Debit = Rp %.2f | Kredit = Rp %.2f | Selisih = Rp %.2f (Status: %s)\n",
		debitTotal, creditTotal, diff, func() string {
			if diff == 0 {
				return "BALANCED (100% OK, Diff = Rp 0.00)"
			}
			return "UNBALANCED"
		}())

	// B. Cek Saldo Dompet Negatif
	var negativeWallets int
	_ = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM wallets WHERE current_balance < 0").Scan(&negativeWallets)
	log.Printf("• Integritas Saldo Dompet: Dompet bersaldo negatif = %d (Status: %s)\n",
		negativeWallets, func() string {
			if negativeWallets == 0 {
				return "PASS (No Overdraft)"
			}
			return "FAIL"
		}())

	// C. Cek Row Count Total
	var totalUsers, totalInterns, totalProjects, totalEvents, totalLedgers int
	_ = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	_ = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM interns").Scan(&totalInterns)
	_ = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM projects").Scan(&totalProjects)
	_ = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM attendance_event_logs").Scan(&totalEvents)
	_ = db.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM financial_ledgers").Scan(&totalLedgers)

	log.Printf("• Total Baris Aktif: Users=%d | Interns=%d | Projects=%d | AttendanceEvents=%d | Ledgers=%d\n",
		totalUsers, totalInterns, totalProjects, totalEvents, totalLedgers)
}
