package integration_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	authApi "github.com/KiranRajeev-KV/nyx-backend/api/auth"
	claimsApi "github.com/KiranRajeev-KV/nyx-backend/api/claims"
	hubsApi "github.com/KiranRajeev-KV/nyx-backend/api/hubs"
	itemsApi "github.com/KiranRajeev-KV/nyx-backend/api/items"
	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	"github.com/KiranRajeev-KV/nyx-backend/internal/email"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	testRouter *gin.Engine
	testDBPool *pgxpool.Pool
)

// ensureRootChanges changes dir to project root for resolving env files relative paths.
func ensureRootChanges() {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		log.Fatalf("failed to change directory: %v", err)
	}
}

// TestMain acts as the entry point for all tests in this package.
func TestMain(m *testing.M) {
	// Only run if specific env var is set to prevent these heavy tests from running during standard `go test ./...`
	// OR if we explicitly ask for them. Here we'll just run them if they check out.
	integrationEnabled := os.Getenv("RUN_INTEGRATION_TESTS")
	if integrationEnabled != "true" {
		log.Println("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run.")
		os.Exit(0)
	}

	ensureRootChanges()

	// 1. Initialize DB Pool for Tests (using standard init from cmd)
	// Make sure .env is loaded (if running locally) or standard ENVs are present
	cmd.Env = &cmd.EnvConfig{
		DatabaseURL: os.Getenv("TEST_DATABASE_URL"),
		Environment: "TEST",
	}

	// Fallback to local default if TEST_DATABASE_URL is not provided
	if cmd.Env.DatabaseURL == "" {
		cmd.Env.DatabaseURL = "postgres://postgres:1234@localhost:5432/postgres" // Standard local pg container setup
	}

	err := cmd.InitDBPool()
	if err != nil {
		log.Fatalf("Could not connect to test database: %v", err)
	}
	testDBPool = cmd.DBPool

	// 2. Init Crypto
	err = pkg.InitPaseto()
	if err != nil {
		log.Fatalf("failed to initialize PASETO keys: %v", err)
	}

	// Try initializing RSA if available; tests usually only mandate PASETO for tokens
	// No explicit InitRSAPair() required if test routes only handle PASETO auth

	// Init Logger for intercepting validations
	logger.Log, err = logger.InitLogger("DEV")
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	// 3. Setup global Gin Router
	gin.SetMode(gin.TestMode)
	testRouter = gin.Default()

	// Mock email service
	mockEmailService := email.NewMockEmailService(true)
	authApi.InitAuthRoutes(mockEmailService)

	// API Groups
	apiGroup := testRouter.Group("/api/v1")

	authApi.AuthRoutes(apiGroup)
	hubsApi.HubRoutes(apiGroup)
	itemsApi.ItemRoutes(apiGroup)
	claimsApi.ClaimRoutes(apiGroup)

	// Run Tests
	code := m.Run()

	// Clean up
	if testDBPool != nil {
		testDBPool.Close()
	}

	os.Exit(code)
}

// Helper to clean specific tables before tests
func cleanDB(t *testing.T) {
	if testDBPool == nil {
		t.Fatal("Database pool not initialized")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Soft delete everything or hard delete for clean slate
	tables := []string{"items", "claims", "hubs", "users", "user_onboarding"}

	for _, table := range tables {
		query := fmt.Sprintf("TRUNCATE TABLE %s CASCADE;", table)
		_, err := testDBPool.Exec(ctx, query)
		if err != nil {
			t.Fatalf("Failed to truncate %s: %v", table, err)
		}
	}
}
