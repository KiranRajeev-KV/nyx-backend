package tests

import (
	"context"
	"fmt"
	"time"

	"github.com/KiranRajeev-KV/nyx-backend/cmd"
	"github.com/jackc/pgx/v5/pgxpool"
)

var TestDBPool *pgxpool.Pool

func SetupTestDB() error {
	if cmd.Env == nil {
		return fmt.Errorf("config not loaded, call Setup() first")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(cmd.Env.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to parse test database config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to create test database pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return fmt.Errorf("failed to connect to test database: %w", err)
	}

	TestDBPool = pool
	return nil
}

func ClearTestDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// List of tables to truncate (order matters due to foreign keys)
	tables := []string{
		"password_resets",
		"user_onboardings",
		"claims",
		"items",
		"hubs",
		"users",
	}

	// Disable foreign key constraints temporarily
	_, err := TestDBPool.Exec(ctx, "SET session_replication_role = replica;")
	if err != nil {
		return fmt.Errorf("failed to disable foreign key constraints: %w", err)
	}

	// Truncate each table
	for _, table := range tables {
		_, err := TestDBPool.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE;", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	// Re-enable foreign key constraints
	_, err = TestDBPool.Exec(ctx, "SET session_replication_role = DEFAULT;")
	if err != nil {
		return fmt.Errorf("failed to re-enable foreign key constraints: %w", err)
	}

	return nil
}

// TeardownTestDB closes the test database pool
func TeardownTestDB() error {
	if TestDBPool != nil {
		TestDBPool.Close()

		// Wait for all connections to close
		for TestDBPool.Stat().TotalConns() > 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return nil
}
