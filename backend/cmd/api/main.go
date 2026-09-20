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
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/modules/documents"
	"dcisp/backend/internal/modules/identity"
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

	policyRepo := system.NewPolicyRepository(db)
	policyService := system.NewPolicyService(policyRepo, rdb)
	policyCtrl := system.NewPolicyController(policyService)

	auditService := system.NewAuditService(db)
	auditCtrl := system.NewAuditController(auditService)

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
