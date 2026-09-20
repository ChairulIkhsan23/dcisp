package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/integrations"
	"dcisp/backend/internal/integrations/apiindonesia"
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/modules/attendance"
	"dcisp/backend/internal/modules/documents"
	"dcisp/backend/internal/modules/identity"
	"dcisp/backend/internal/modules/people"
	"dcisp/backend/internal/modules/performance"
	"dcisp/backend/internal/modules/projects"
	"dcisp/backend/internal/modules/system"
	"dcisp/backend/internal/shared/eventbus"
	"dcisp/backend/internal/shared/response"
	"dcisp/backend/internal/worker"
	"github.com/gin-gonic/gin"
)

// Menginisialisasi konfigurasi, koneksi database, routing middleware, dan menjalankan server HTTP Go backend.
func main() {
	log.Println("==========================================================")
	log.Println("Starting DCISP Platform v1.0 — Golang Backend Server")
	log.Println("==========================================================")

	// 1. Load Configurations (Fail-hard if essential configs missing)
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Fatal: Kesalahan konfigurasi aplikasi: %v", err)
	}
	gin.SetMode(cfg.GinMode)

	// 2. Initialize PostgreSQL Pool
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: Koneksi PostgreSQL gagal: %v", err)
	}
	defer db.Close()

	// 3. Initialize Redis Client
	rdb, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Fatal: Koneksi Redis gagal: %v", err)
	}
	defer rdb.Close()

	// 4. Initialize Background Worker & Event Bus
	workerPool := worker.NewWorkerPool(cfg)
	if err := workerPool.Start(); err != nil {
		log.Printf("Peringatan: Gagal menjalankan Asynq worker server: %v", err)
	}
	eventBus := eventbus.NewEventBus(workerPool)

	// 5. Initialize Modular Services & Controllers
	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo, cfg)
	identityCtrl := identity.NewController(identityService)

	storageService := documents.NewStorageService(db, cfg)
	storageCtrl := documents.NewStorageController(storageService)

	auditService := system.NewAuditService(db)
	auditCtrl := system.NewAuditController(auditService)

	// Registry API eksternal: daftarkan setiap provider di sini agar modul domain
	// cukup mengonsumsi instance terdaftar (contoh: api-indonesia, provider lain menyusul).
	externalRegistry := integrations.NewRegistry()
	apiIndonesiaClient := apiindonesia.NewClient(cfg.APIIndonesiaBaseURL, cfg.APIIndonesiaKey)
	if err := externalRegistry.Register(apiIndonesiaClient); err != nil {
		log.Fatalf("Fatal: Registrasi provider API eksternal gagal: %v", err)
	}

	peopleRepo := people.NewRepository(db)
	peopleService := people.NewService(peopleRepo, db, auditService, eventBus)
	peopleService.SetExternalDependencies(apiIndonesiaClient, rdb)
	peopleCtrl := people.NewController(peopleService)

	attendanceRepo := attendance.NewRepository(db)
	attendanceService := attendance.NewService(attendanceRepo, db, rdb, cfg, auditService, eventBus)
	attendanceCtrl := attendance.NewController(attendanceService)

	projectsRepo := projects.NewRepository(db)
	projectsService := projects.NewService(projectsRepo, db, storageService, auditService, eventBus)
	projectsCtrl := projects.NewController(projectsService)

	performanceRepo := performance.NewRepository(db)
	performanceService := performance.NewService(performanceRepo, db, auditService, eventBus)
	performanceCtrl := performance.NewController(performanceService)

	policyRepo := system.NewPolicyRepository(db)
	policyService := system.NewPolicyService(policyRepo, rdb)
	policyCtrl := system.NewPolicyController(policyService)

	settingsRepo := system.NewSettingsRepository(db)
	settingsService := system.NewSettingsService(settingsRepo, rdb)
	settingsCtrl := system.NewSettingsController(settingsService)

	sseCtrl := system.NewSSEController(cfg, eventBus)

	// 6. Setup Gin Router
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.AuditInterceptor(db))

	// CORS Middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Device-ID, X-DCISP-Signature")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// 7. System Health Probes
	r.GET("/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "UP",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"service":   "dcisp-backend-api",
		})
	})

	r.GET("/health/readiness", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		dbStatus := "UP"
		if err := db.Pool.Ping(ctx); err != nil {
			log.Printf("Readiness probe gagal memeriksa koneksi database: %v", err)
			dbStatus = "DOWN"
		}

		redisStatus := "UP"
		if err := rdb.Client.Ping(ctx).Err(); err != nil {
			log.Printf("Readiness probe gagal memeriksa koneksi redis: %v", err)
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

	// 8. API v1 Routing
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.GeneralRateLimiter(rdb))
	{
		apiV1.GET("/ping", func(c *gin.Context) {
			response.Success(c, http.StatusOK, "DCISP API Gateway v1.0 Siap", gin.H{
				"realm": "The 8-Bit Professional Realm",
			})
		})

		// Real-time Server-Sent Events (SSE) Stream
		apiV1.GET("/events/stream", sseCtrl.StreamEvents)

		// Identity & Authentication Routes
		authRoutes := apiV1.Group("/auth")
		authRoutes.Use(middleware.AuthRateLimiter(rdb))
		{
			authRoutes.POST("/login", identityCtrl.Login)
			authRoutes.POST("/refresh", identityCtrl.RefreshToken)

			protectedAuth := authRoutes.Group("")
			protectedAuth.Use(middleware.AuthJWT(cfg))
			{
				protectedAuth.GET("/me", identityCtrl.GetMe)
				protectedAuth.GET("/roles", middleware.RequirePermission(db, rdb, "system.roles", "view", "SYSTEM"), identityCtrl.GetRoles)
			}
		}

		// Cloudflare R2 Storage Routes (FR-044)
		storageRoutes := apiV1.Group("/storage")
		storageRoutes.Use(middleware.AuthJWT(cfg))
		{
			storageRoutes.POST("/presigned-upload", storageCtrl.PresignedUpload)
			storageRoutes.GET("/files/:id", storageCtrl.GetMetadata)
			storageRoutes.GET("/files/:id/download-url", storageCtrl.GetDownloadURL)
		}

		// Unified Policy Engine Routes (FR-045)
		policyRoutes := apiV1.Group("/policies")
		policyRoutes.Use(middleware.AuthJWT(cfg))
		{
			policyRoutes.POST("", middleware.RequirePermission(db, rdb, "system.policies", "create", "SYSTEM"), policyCtrl.Create)
			policyRoutes.GET("", middleware.RequirePermission(db, rdb, "system.policies", "view", "SYSTEM"), policyCtrl.List)
			policyRoutes.GET("/:id", middleware.RequirePermission(db, rdb, "system.policies", "view", "SYSTEM"), policyCtrl.GetByID)
			policyRoutes.PUT("/:id", middleware.RequirePermission(db, rdb, "system.policies", "update", "SYSTEM"), policyCtrl.Update)
			policyRoutes.DELETE("/:id", middleware.RequirePermission(db, rdb, "system.policies", "delete", "SYSTEM"), policyCtrl.Delete)
			policyRoutes.POST("/evaluate", middleware.RequirePermission(db, rdb, "system.policies", "evaluate", "SYSTEM"), policyCtrl.Evaluate)
		}

		// Centralized System Settings & Audit Logs (FR-047, FR-048)
		systemRoutes := apiV1.Group("/system")
		systemRoutes.Use(middleware.AuthJWT(cfg))
		{
			systemRoutes.GET("/audit-logs", middleware.RequirePermission(db, rdb, "system.audit_logs", "view", "SYSTEM"), auditCtrl.List)
			systemRoutes.GET("/audit-logs/:id", middleware.RequirePermission(db, rdb, "system.audit_logs", "view", "SYSTEM"), auditCtrl.GetByID)

			systemRoutes.POST("/settings", middleware.RequirePermission(db, rdb, "system.settings", "create", "SYSTEM"), settingsCtrl.SetSetting)
			systemRoutes.GET("/settings", middleware.RequirePermission(db, rdb, "system.settings", "view", "SYSTEM"), settingsCtrl.ListSettings)
			systemRoutes.GET("/settings/:key", middleware.RequirePermission(db, rdb, "system.settings", "view", "SYSTEM"), settingsCtrl.GetSetting)
			systemRoutes.DELETE("/settings/:key", middleware.RequirePermission(db, rdb, "system.settings", "delete", "SYSTEM"), settingsCtrl.DeleteSetting)
		}

		// People & Lifecycle Management (Domain 2: FR-002 s/d FR-006)
		peopleRoutes := apiV1.Group("/people")
		peopleRoutes.Use(middleware.AuthJWT(cfg))
		{
			// Institutions (FR-005, T-030) — katalog eksternal API Indonesia + database internal
			peopleRoutes.POST("/institutions", middleware.RequirePermission(db, rdb, "people.institutions", "create", "WORKFORCE_AND_PEOPLE"), peopleCtrl.CreateInstitution)
			peopleRoutes.GET("/institutions", middleware.RequirePermission(db, rdb, "people.institutions", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.ListInstitutions)
			peopleRoutes.GET("/institutions/search-external", middleware.RequirePermission(db, rdb, "people.institutions", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.SearchExternalInstitutions)
			peopleRoutes.GET("/institutions/external/:source/:external_id", middleware.RequirePermission(db, rdb, "people.institutions", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.GetExternalInstitutionDetail)
			peopleRoutes.POST("/institutions/import-external", middleware.RequirePermission(db, rdb, "people.institutions", "create", "WORKFORCE_AND_PEOPLE"), peopleCtrl.ImportExternalInstitution)
			peopleRoutes.GET("/institutions/:id", middleware.RequirePermission(db, rdb, "people.institutions", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.GetInstitutionByID)
			peopleRoutes.PUT("/institutions/:id", middleware.RequirePermission(db, rdb, "people.institutions", "update", "WORKFORCE_AND_PEOPLE"), peopleCtrl.UpdateInstitution)
			peopleRoutes.DELETE("/institutions/:id", middleware.RequirePermission(db, rdb, "people.institutions", "delete", "WORKFORCE_AND_PEOPLE"), peopleCtrl.DeleteInstitution)

			// Batches & Cohorts (FR-004, T-028)
			peopleRoutes.POST("/batches", middleware.RequirePermission(db, rdb, "people.batches", "create", "WORKFORCE_AND_PEOPLE"), peopleCtrl.CreateBatch)
			peopleRoutes.GET("/batches", middleware.RequirePermission(db, rdb, "people.batches", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.ListBatches)
			peopleRoutes.GET("/batches/:id", middleware.RequirePermission(db, rdb, "people.batches", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.GetBatchByID)
			peopleRoutes.PUT("/batches/:id", middleware.RequirePermission(db, rdb, "people.batches", "update", "WORKFORCE_AND_PEOPLE"), peopleCtrl.UpdateBatch)

			// Interns & Lifecycle State Machine (FR-002, T-027, T-032)
			peopleRoutes.POST("/interns", middleware.RequirePermission(db, rdb, "people.interns", "create", "WORKFORCE_AND_PEOPLE"), peopleCtrl.RegisterIntern)
			peopleRoutes.GET("/interns", middleware.RequirePermission(db, rdb, "people.interns", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.SearchInterns)
			peopleRoutes.GET("/interns/:id", middleware.RequirePermission(db, rdb, "people.interns", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.GetInternByID)
			peopleRoutes.GET("/interns/by-user/:user_id", middleware.RequirePermission(db, rdb, "people.interns", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.GetInternByUserID)
			peopleRoutes.PATCH("/interns/:id/status", middleware.RequirePermission(db, rdb, "people.interns", "update", "WORKFORCE_AND_PEOPLE"), peopleCtrl.ChangeInternStatus)
			peopleRoutes.PATCH("/interns/:id/mentor", middleware.RequirePermission(db, rdb, "people.interns", "update", "WORKFORCE_AND_PEOPLE"), peopleCtrl.AssignMentor)
			peopleRoutes.PATCH("/interns/:id/batch", middleware.RequirePermission(db, rdb, "people.interns", "update", "WORKFORCE_AND_PEOPLE"), peopleCtrl.PlotBatch)

			// Alumni (FR-003, BR-003, T-029)
			peopleRoutes.GET("/alumni", middleware.RequirePermission(db, rdb, "people.alumni", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.ListAlumni)
			peopleRoutes.GET("/alumni/by-user/:user_id", middleware.RequirePermission(db, rdb, "people.alumni", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.GetAlumniByUserID)

			// Skills & Skill Matrix (FR-006, T-031)
			peopleRoutes.POST("/skills", middleware.RequirePermission(db, rdb, "people.skills", "create", "WORKFORCE_AND_PEOPLE"), peopleCtrl.CreateSkill)
			peopleRoutes.GET("/skills", middleware.RequirePermission(db, rdb, "people.skills", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.ListSkills)
			peopleRoutes.POST("/skills/user-skills/:user_id", middleware.RequirePermission(db, rdb, "people.skills", "update", "WORKFORCE_AND_PEOPLE"), peopleCtrl.AssignUserSkill)
			peopleRoutes.GET("/skills/user-skills/:user_id", middleware.RequirePermission(db, rdb, "people.skills", "view", "WORKFORCE_AND_PEOPLE"), peopleCtrl.GetUserSkills)
			peopleRoutes.DELETE("/skills/user-skills/:user_id/:skill_id", middleware.RequirePermission(db, rdb, "people.skills", "delete", "WORKFORCE_AND_PEOPLE"), peopleCtrl.DeleteUserSkill)
		}

		// Attendance & Presensi (Domain 3: FR-007 s/d FR-015)
		attendanceRoutes := apiV1.Group("/attendance")
		{
			// Terminal Tap & Offline Sync (FR-008, T-034, T-042)
			attendanceRoutes.POST("/terminal-tap", attendanceCtrl.HandleTerminalTap)
			attendanceRoutes.POST("/sync-offline", attendanceCtrl.SyncOffline)
			attendanceRoutes.GET("/audio-catalog", attendanceCtrl.GetAudioCatalog)

			// Protected Attendance endpoints
			protectedAttendance := attendanceRoutes.Group("")
			protectedAttendance.Use(middleware.AuthJWT(cfg))
			{
				// Dynamic QR Code (T-043)
				protectedAttendance.POST("/qr/generate", attendanceCtrl.GenerateQR)

				// Overtime (FR-012, T-038)
				protectedAttendance.POST("/overtime/request", attendanceCtrl.RequestOvertime)
				protectedAttendance.PATCH("/overtime/:id/review", middleware.RequirePermission(db, rdb, "attendance.overtime", "review", "ASSIGNED_TEAM"), attendanceCtrl.ReviewOvertime)

				// Leave (FR-014, T-039)
				protectedAttendance.POST("/leave/request", attendanceCtrl.RequestLeave)
				protectedAttendance.PATCH("/leave/:id/review", middleware.RequirePermission(db, rdb, "attendance.leave", "review", "WORKFORCE_AND_PEOPLE"), attendanceCtrl.ReviewLeave)

				// Corrections (FR-015, T-040)
				protectedAttendance.POST("/corrections", attendanceCtrl.RequestCorrection)
				protectedAttendance.PATCH("/corrections/:id/review", middleware.RequirePermission(db, rdb, "attendance.corrections", "review", "ASSIGNED_TEAM"), attendanceCtrl.ReviewCorrection)

				// Device Registry (FR-013, T-041)
				protectedAttendance.POST("/devices", middleware.RequirePermission(db, rdb, "attendance.devices", "create", "SYSTEM"), attendanceCtrl.RegisterDevice)
				protectedAttendance.GET("/devices", middleware.RequirePermission(db, rdb, "attendance.devices", "view", "SYSTEM"), attendanceCtrl.ListDevices)
				protectedAttendance.POST("/devices/:id/heartbeat", attendanceCtrl.DeviceHeartbeat)
			}
		}

		// Work Sessions & Breaks (FR-009, FR-010, T-036, T-037)
		workSessionRoutes := apiV1.Group("/work-sessions")
		workSessionRoutes.Use(middleware.AuthJWT(cfg))
		{
			workSessionRoutes.POST("/start", attendanceCtrl.StartWorkSession)
			workSessionRoutes.POST("/break", attendanceCtrl.ProcessBreak)
			workSessionRoutes.POST("/end", attendanceCtrl.EndWorkSession)
		}

		// Projects & Marketplace (Domain 4: FR-016 s/d FR-024)
		projectsRoutes := apiV1.Group("/projects")
		projectsRoutes.Use(middleware.AuthJWT(cfg))
		{
			// Marketplace & Projects (FR-016, T-045)
			projectsRoutes.POST("", middleware.RequirePermission(db, rdb, "projects.marketplace", "create", "ASSIGNED_PROJECTS"), projectsCtrl.CreateProject)
			projectsRoutes.GET("", projectsCtrl.ListProjects)
			projectsRoutes.GET("/:id", projectsCtrl.GetProjectByID)
			projectsRoutes.PUT("/:id", middleware.RequirePermission(db, rdb, "projects.marketplace", "update", "ASSIGNED_PROJECTS"), projectsCtrl.UpdateProject)

			// Applications & Quota (FR-017, T-046)
			projectsRoutes.POST("/:id/apply", projectsCtrl.ApplyProject)
			projectsRoutes.GET("/:id/applications", middleware.RequirePermission(db, rdb, "projects.applications", "view", "ASSIGNED_PROJECTS"), projectsCtrl.ListApplications)
			projectsRoutes.PATCH("/applications/:id/review", middleware.RequirePermission(db, rdb, "projects.applications", "review", "ASSIGNED_PROJECTS"), projectsCtrl.ReviewApplication)

			// Teams & Planned Contribution (FR-019, T-047)
			projectsRoutes.GET("/:id/team", projectsCtrl.ListTeamMembers)
			projectsRoutes.POST("/:id/planned-contributions", middleware.RequirePermission(db, rdb, "projects.team", "update", "ASSIGNED_PROJECTS"), projectsCtrl.FinalizePlannedContribution)

			// Milestones (FR-020, T-048)
			projectsRoutes.POST("/:id/milestones", middleware.RequirePermission(db, rdb, "projects.milestones", "create", "ASSIGNED_PROJECTS"), projectsCtrl.CreateMilestone)
			projectsRoutes.GET("/:id/milestones", projectsCtrl.ListMilestones)

			// Tasks Kanban (FR-021, T-049)
			projectsRoutes.POST("/:id/tasks", middleware.RequirePermission(db, rdb, "projects.tasks", "create", "ASSIGNED_PROJECTS"), projectsCtrl.CreateTask)
			projectsRoutes.GET("/:id/tasks", projectsCtrl.ListTasks)

			// Three-Layer Contribution (FR-024, T-052)
			projectsRoutes.GET("/:id/actual-contributions", projectsCtrl.CalculateActualContribution)
			projectsRoutes.POST("/:id/final-contributions", middleware.RequirePermission(db, rdb, "projects.contribution", "finalize", "ASSIGNED_PROJECTS"), projectsCtrl.FinalizeContribution)
		}

		// Tasks & Submissions (FR-021, FR-022, FR-023, T-049, T-050, T-051)
		tasksRoutes := apiV1.Group("/tasks")
		tasksRoutes.Use(middleware.AuthJWT(cfg))
		{
			tasksRoutes.GET("/:id", projectsCtrl.GetTaskByID)
			tasksRoutes.PATCH("/:id/status", projectsCtrl.ChangeTaskStatus)
			tasksRoutes.POST("/:id/submissions", projectsCtrl.SubmitWorkReport)
			tasksRoutes.PATCH("/submissions/:id/review", middleware.RequirePermission(db, rdb, "tasks.submissions", "review", "ASSIGNED_TASKS"), projectsCtrl.ReviewWorkReport)
		}

		// Performance & Gamification Routes (Domain 5: FR-025 s/d FR-030)
		performanceRoutes := apiV1.Group("/performance")
		performanceRoutes.Use(middleware.AuthJWT(cfg))
		{
			// XP & Ranks (FR-025, FR-026, T-054, T-056)
			performanceRoutes.GET("/my-stats", performanceCtrl.GetMyStats)
			performanceRoutes.GET("/users/:user_id/stats", performanceCtrl.GetUserStats)
			performanceRoutes.POST("/xp/mutate", middleware.RequirePermission(db, rdb, "performance.xp", "mutate", "SYSTEM"), performanceCtrl.MutateXP)
			performanceRoutes.GET("/xp/history", performanceCtrl.GetXPHistory)
			performanceRoutes.GET("/ranks", performanceCtrl.ListRanks)

			// Evaluations (FR-027, T-058)
			performanceRoutes.POST("/evaluations", middleware.RequirePermission(db, rdb, "performance.evaluations", "create", "ASSIGNED_TEAM"), performanceCtrl.CreateEvaluation)
			performanceRoutes.GET("/evaluations/:id", performanceCtrl.GetEvaluationByID)

			// Top Performer (FR-028, T-059)
			performanceRoutes.POST("/batches/:batch_id/top-performer", middleware.RequirePermission(db, rdb, "performance.top_performer", "create", "WORKFORCE_AND_PEOPLE"), performanceCtrl.DetermineTopPerformer)

			// Achievements (FR-029, T-060)
			performanceRoutes.GET("/achievements", performanceCtrl.ListAchievements)
			performanceRoutes.POST("/achievements", middleware.RequirePermission(db, rdb, "performance.achievements", "create", "SYSTEM"), performanceCtrl.CreateAchievement)
			performanceRoutes.GET("/my-achievements", performanceCtrl.GetMyAchievements)
			performanceRoutes.POST("/achievements/:code/unlock", performanceCtrl.UnlockAchievement)

			// Skill Growth Matrix (FR-030, T-061)
			performanceRoutes.GET("/my-skill-growth", performanceCtrl.GetMySkillGrowth)
		}
	}

	// 9. Graceful Server Start
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on http://localhost:%s\n", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listener failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	// Shutdown Asynq worker pool and audit logger
	workerPool.Shutdown()
	middleware.CloseAuditWorker()

	log.Println("Server exiting successfully")
}
