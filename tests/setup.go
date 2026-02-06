package tests

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"aidanwoods.dev/go-paseto"
	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/KiranRajeev-KV/nyx-backend/pkg"
	"github.com/gin-gonic/gin"
)

var (
	setupOnce sync.Once
	setupErr  error
)

const (
	testPrivateKeyPath = "tests/app.test.rsa"
	testPublicKeyPath  = "tests/app.test.pub.rsa"
	testEnvPath        = "tests/.env.test"
)

// Setup initializes all dependencies for testing (mirrors main.go initialization).
// Safe to call multiple times - only runs once.
func Setup() error {
	setupOnce.Do(func() {
		setupErr = doSetup()
	})
	return setupErr
}

// getProjectRoot returns the absolute path to the project root directory
func getProjectRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	// This file is at tests/setup.go, so go up one level
	return filepath.Dir(filepath.Dir(filename))
}

// initTestPaseto generates or loads test-specific PASETO keys in the tests directory
func initTestPaseto(projectRoot string) error {
	privateKeyPath := filepath.Join(projectRoot, testPrivateKeyPath)
	publicKeyPath := filepath.Join(projectRoot, testPublicKeyPath)

	// Check if test keys exist, generate if not
	if _, err := os.Stat(privateKeyPath); os.IsNotExist(err) {
		privateKey := paseto.NewV4AsymmetricSecretKey()
		publicKey := privateKey.Public()

		if err := os.WriteFile(privateKeyPath, privateKey.ExportBytes(), 0644); err != nil {
			return fmt.Errorf("failed to write test private key: %w", err)
		}
		if err := os.WriteFile(publicKeyPath, publicKey.ExportBytes(), 0644); err != nil {
			return fmt.Errorf("failed to write test public key: %w", err)
		}
	}

	// Load the test keys
	privateKeyBinary, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read test private key: %w", err)
	}
	privateKeyHex := hex.EncodeToString(privateKeyBinary)

	publicKeyBinary, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read test public key: %w", err)
	}
	publicKeyHex := hex.EncodeToString(publicKeyBinary)

	// Set the package-level PASETO keys directly
	pkg.VerifyKey, err = paseto.NewV4AsymmetricPublicKeyFromHex(publicKeyHex)
	if err != nil {
		return fmt.Errorf("failed to parse test public key: %w", err)
	}
	pkg.SignKey, err = paseto.NewV4AsymmetricSecretKeyFromHex(privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to parse test private key: %w", err)
	}

	return nil
}

func doSetup() error {
	projectRoot := getProjectRoot()

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Load config from tests/.env.test
	envPath := filepath.Join(projectRoot, testEnvPath)
	cfg, err := cmd.LoadConfigFrom(envPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cmd.Env = cfg

	// Initialize Logger (silent for tests)
	log, err := logger.InitLogger("TEST")
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	logger.Log = log

	// Initialize test-specific PASETO keys (stored in tests/ directory)
	if err := initTestPaseto(projectRoot); err != nil {
		return fmt.Errorf("failed to initialize test Paseto: %w", err)
	}

	// Initialize test database pool
	if err := SetupTestDB(); err != nil {
		return fmt.Errorf("failed to initialize test database: %w", err)
	}

	return nil
}

// MustSetup calls Setup and panics if it fails.
// Useful for TestMain functions.
func MustSetup() {
	if err := Setup(); err != nil {
		panic("test setup failed: " + err.Error())
	}
}
