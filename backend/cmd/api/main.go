package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dcisp/backend/internal/config"
	"dcisp/backend/internal/database"
	"dcisp/backend/internal/middleware"
	"dcisp/backend/internal/modules/identity"
	"dcisp/backend/internal/shared/response"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("==========================================================")
	log.Println("🚀 Starting DCISP Platform v1.0 — Golang Backend Server")
	log.Println("==========================================================")

	// 1. Load Configurations
	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	// 2. Initialize PostgreSQL Pool
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Fatal: PostgreSQL connection failed: %v", err)
	}
	defer db.Close()

	// 3. Initialize Redis Client
	rdb, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Fatal: Redis connection failed: %v", err)
	}
	defer rdb.Close()

	// 4. Initialize Modular Services & Controllers
	identityRepo := identity.NewRepository(db)
	identityService := identity.NewService(identityRepo, cfg)
	identityCtrl := identity.NewController(identityService)

	// 5. Setup Gin Router
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

	// 6. System Health Probes
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
			dbStatus = fmt.Sprintf("DOWN: %v", err)
		}

		redisStatus := "UP"
		if err := rdb.Client.Ping(ctx).Err(); err != nil {
			redisStatus = fmt.Sprintf("DOWN: %v", err)
		}

		status := http.StatusOK
		if dbStatus != "UP" || redisStatus != "UP" {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status":     "UP",
			"database":   dbStatus,
			"redis":      redisStatus,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"version":    "1.0.0",
		})
	})

	// 7. API v1 Routing
	apiV1 := r.Group("/api/v1")
	{
		apiV1.GET("/ping", func(c *gin.Context) {
			response.Success(c, http.StatusOK, "DCISP API Gateway v1.0 Ready", gin.H{
				"realm": "The 8-Bit Professional Realm",
			})
		})

		// Identity & Authentication Routes
		authRoutes := apiV1.Group("/auth")
		{
			authRoutes.POST("/login", identityCtrl.Login)

			// Protected routes
			protected := authRoutes.Group("")
			protected.Use(middleware.AuthJWT(cfg))
			{
				protected.GET("/me", identityCtrl.GetMe)
				protected.GET("/roles", middleware.RequirePermission(db, "system.roles", "view", "SYSTEM"), identityCtrl.GetRoles)
			}
		}
	}

	// 8. Graceful Server Start
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

	log.Println("Server exiting successfully")
}
