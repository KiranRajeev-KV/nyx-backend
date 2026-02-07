package tests

import (
	"io"
	"os"

	"github.com/KiranRajeev-KV/nyx-backend/internal/logger"
	"github.com/rs/zerolog"
)

// InitTestLogger initializes a test logger that logs to io.Discard
// This should be called in test files that need the logger
func InitTestLogger() {
	testLogger := zerolog.New(io.Discard).With().Logger()
	logger.Log = &logger.LoggerService{
		Logger: testLogger,
		Env:    "test",
	}
}

// InitTestLoggerWithOutput initializes a test logger that logs to the provided output
// This is useful for debugging tests
func InitTestLoggerWithOutput(output *os.File) {
	testLogger := zerolog.New(output).With().Logger()
	logger.Log = &logger.LoggerService{
		Logger: testLogger,
		Env:    "test",
	}
}