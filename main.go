package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	"github.com/KiranRajeev-KV/nyx-backend/internal/email"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	mw "github.com/KiranRajeev-KV/nyx-backend/internal/middleware"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/KiranRajeev-KV/nyx-backend/pkg/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	apiAdmin "github.com/KiranRajeev-KV/nyx-backend/api/admin"
	apiAudit "github.com/KiranRajeev-KV/nyx-backend/api/audit"
	apiAuth "github.com/KiranRajeev-KV/nyx-backend/api/auth"
	apiClaims "github.com/KiranRajeev-KV/nyx-backend/api/claims"
	apiHubs "github.com/KiranRajeev-KV/nyx-backend/api/hubs"
	apiItems "github.com/KiranRajeev-KV/nyx-backend/api/items"
)

func StartServer() {
	gin.SetMode(gin.ReleaseMode)

	// === Init ===

	cfg, err := cmd.LoadConfig()
	if err != nil {
		fmt.Println("[FATAL] Could not load EnvConfig: ", err)
		panic(err)

	}
	fmt.Println("[OK]: EnvConfig loaded successfully")

	cmd.Env = cfg

	// Initialize Logger
	log, err := logger.InitLogger(cfg.Environment)
	if err != nil {
		fmt.Println("[FATAL]: Could not initialize Logger: ", err)
		panic(err)
	}
	logger.Log = log

	// Initialize DB Pool
	if err := cmd.InitDBPool(); err != nil {
		logger.Log.Error("[FATAL]: Could not initialize DB Pool: ", err)
		return
	}
	logger.Log.Info("[OK]: DB Pool initialized successfully")

	// Initialize RSA
	err = cmd.CheckRSAKeyPairExists()
	if err != nil {
		err = cmd.GenerateRSAKeyPair()
		if err != nil {
			logger.Log.Fatal("[CRASH]: Failed to initialize rsa", err)
		}
		logger.Log.Info("[OK]: RSA keypair generated and saved successfully.")
	} else {
		logger.Log.Info("[OK]: Using existing RSA keypair.")
	}

	// Initialize Paseto
	if err := pkg.InitPaseto(); err != nil {
		logger.Log.Error("[FATAL]: Could not initialize Paseto: ", err)
		return
	}
	logger.Log.Info("[OK]: Paseto initialized successfully")

	// Initialize Email Service
	emailService := email.NewEmailService(cmd.Env, logger.Log)
	if emailService.IsEnabled() {
		logger.Log.Info("[OK]: Email service initialized successfully")
	} else {
		logger.Log.Info("[INFO]: Email service is disabled")
	}

	// Initialize S3 Storage
	if err := storage.InitS3(cmd.Env.S3Endpoint, cmd.Env.S3Region, cmd.Env.S3BucketName, cmd.Env.S3AccessKeyID, cmd.Env.S3SecretAccessKey); err != nil {
		logger.Log.Error("[FATAL]: Could not initialize S3 storage: ", err)
		return
	}
	logger.Log.Info("[OK]: S3 storage initialized successfully")

	// === Router Setup ===

	config := cors.Config{
		AllowOrigins:              []string{cmd.Env.ClientDomain},
		AllowWildcard:             true,
		AllowMethods:              []string{"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS"},
		AllowHeaders:              []string{"X-Csrf-Token", "Origin", "Content-Type"},
		AllowCredentials:          true,
		OptionsResponseStatusCode: 204,
		MaxAge:                    12 * time.Hour,
	}

	router := gin.New()

	// middlewares
	router.Use(cors.New(config))
	router.Use(pkg.TagRequestWithId)
	router.Use(mw.LogMiddleware(logger.Log))

	// router.RedirectTrailingSlash = false
	// router.RedirectFixedPath = false

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Server is running!",
		})
		logger.Log.SuccessCtx(c)
	})

	v2 := router.Group("/api/v2")

	// Initialize auth routes with email service
	apiAuth.InitAuthRoutes(emailService)
	apiAuth.AuthRoutes(v2)
	apiAudit.AuditRoutes(v2)
	apiItems.ItemRoutes(v2)
	apiHubs.HubRoutes(v2)
	apiClaims.ClaimRoutes(v2)
	apiAdmin.AdminRoutes(v2)

	// === Server Setup ===

	server := &http.Server{
		Addr:    ":" + strconv.Itoa(cmd.Env.Port),
		Handler: router,
	}

	go func() {
		logger.Log.Info("[OK]: Starting server on port " + strconv.Itoa(cmd.Env.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Error("[FATAL]: Could not start server: ", err)
		} // Blocking in nature
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("[OK]: Shutting down server...")

	// 10 seconds timeout for the server to shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error("[FATAL]: Server forced to shutdown: ", err)
	}
}

func main() {
	StartServer()
}
