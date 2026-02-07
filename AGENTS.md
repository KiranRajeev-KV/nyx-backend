# AGENTS.md

This file contains guidelines and commands for agentic coding agents working on the Nyx backend project.

## Project Overview

Nyx is a Go-based backend service for managing lost-and-found items with user authentication, item tracking, and claim processing. It uses PostgreSQL with pgvector, Gin web framework, and follows Clean Architecture principles. The project includes comprehensive testing infrastructure, API testing with Bruno, and database seeding capabilities.

## Build & Development Commands

### Setup & Dependencies
```bash
task setup              # Install dependencies, create .env, start Docker
task deps               # Download and tidy Go dependencies
task env                # Create .env from .env.sample
```

### Development Server
```bash
task dev                # Start development server with hot reload (Air)
task build              # Build the application to ./bin/nyx-backend
task run                # Build and run the application
```

### Database Management
```bash
task docker:up          # Start PostgreSQL and Drizzle Gateway containers
task docker:down        # Stop and remove containers
task start              # Start existing containers
task stop               # Stop containers
task docker:logs        # View database logs
task docker:reset       # Reset database (WARNING: deletes data)
```

### Database Migrations
```bash
task up                 # Run all pending migrations
task down               # Rollback last migration
task status             # Show migration status
```

### Database Seeding
```bash
task db:seed            # Seed database with dummy data (DEV only)
task db:truncate        # Truncate all tables (DEV only)
task db:rebuild         # Truncate and seed database (DEV only)
```

### Code Generation
```bash
task gen                # Generate SQLC code from SQL queries
task gen:watch          # Watch for changes and regenerate SQLC code
```

### Code Quality
```bash
task fmt                # Format Go code with go fmt
task vet                # Run static analysis with go vet
task lint               # Run golangci-lint (requires installation)
```

### Testing
```bash
task test              # Run all tests
task test:coverage     # Run tests with coverage report
task test:unit         # Run unit tests only (pkg/...)
task test:integration  # Run integration tests only (tests/integration/...)
```

## Code Style Guidelines

### Import Organization
- Group imports in three blocks: standard library, third-party, internal
- Use blank lines between groups
- Prefer aliasing for clarity when needed
```go
import (
    "context"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"

    db "github.com/KiranRajeev-KV/nyx-backend/internal/db/gen"
    "github.com/KiranRajeev-KV/nyx-backend/internal/logger"
)
```

### Naming Conventions
- **Packages**: lowercase, short, descriptive (e.g., `models`, `middleware`, `logger`)
- **Functions**: PascalCase for exported, camelCase for private
- **Variables**: camelCase, descriptive names
- **Constants**: PascalCase for exported, camelCase for private
- **Database**: Use SQLC-generated types and naming

### Error Handling Patterns
- Always handle errors explicitly
- Use structured logging with context
- Follow the project's error handling utilities:
```go
// Database transaction errors
if pkg.HandleDbTxnErr(c, err, "OPERATION") {
    return
}
defer pkg.RollbackTx(c, tx, ctx, "OPERATION")

// Database connection errors
if pkg.HandleDbAcquireErr(c, err, "OPERATION") {
    return
}
defer conn.Release()

// Request validation
req, ok := pkg.ValidateRequest[models.RequestType](c)
if !ok {
    return
}
```

### Request/Response Models
- Define request models in `internal/models/`
- Implement `Validate()` method using ozzo-validation
- Return error message and validation error:
```go
func (r RequestType) Validate() (errorMsg string, err error) {
    err = v.ValidateStruct(&r,
        v.Field(&r.Field, v.Required, v.Length(3, 100)),
        v.Field(&r.Email, v.Required, is.Email),
    )
    return "Invalid request format for operation", err
}
```

### Database Operations
- Use SQLC for type-safe SQL operations
- Prefer transactions for multi-step operations
- Use context with timeout for database operations
- Follow the pattern: acquire, use, release
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

tx, err := cmd.DBPool.Begin(ctx)
if pkg.HandleDbTxnErr(c, err, "OPERATION") {
    return
}
defer pkg.RollbackTx(c, tx, ctx, "OPERATION")

q := db.New()
// ... operations ...

err = tx.Commit(ctx)
if pkg.HandleDbTxnCommitErr(c, err, "OPERATION") {
    return
}
```

### Logging Patterns
- Use structured logging with context
- Include operation context in log messages
- Follow the log level conventions:
```go
logger.Log.ErrorCtx(c, "[OPERATION-ERROR]: Description", err)
logger.Log.InfoCtx(c, "[OPERATION-INFO]: Description")
logger.Log.SuccessCtx(c)  // For successful operations
logger.Log.WarnCtx(c, "[OPERATION-WARN]: Description")
```

### Authentication & Authorization
- Use PASETO tokens for authentication
- Implement role-based access control
- Use middleware for authentication checks
- Follow the token pattern: access + refresh tokens
```go
// Set auth cookies
pkg.SetAuthCookie(c, accessToken)
pkg.SetRefreshCookie(c, refreshToken)

// Get user context
email, ok := pkg.GetEmail(c, "OPERATION")
if !ok {
    return
}
```

### API Response Patterns
- Use consistent JSON response format
- Include descriptive error messages
- Use appropriate HTTP status codes
- Log all operations with context
```go
// Success response
c.JSON(http.StatusOK, gin.H{
    "message": "Operation completed successfully",
})
logger.Log.SuccessCtx(c)

// Error response
c.JSON(http.StatusInternalServerError, gin.H{
    "message": "Oops! Something happened. Please try again later.",
})
logger.Log.ErrorCtx(c, "[OPERATION-ERROR]: Description", err)
```

### File Structure Conventions
- **API handlers**: `api/{domain}/controllers.go`
- **Routes**: `api/{domain}/routes.go`
- **Models**: `internal/models/{domain}.models.go`
- **Middleware**: `internal/middleware/`
- **Database**: `internal/db/` (migrations, queries, generated code)
- **Utilities**: `pkg/`
- **Configuration**: `cmd/`

### Environment Configuration
- Use `.env` file for local development
- Reference `.env.sample` for required variables
- Configuration management using Koanf
- Support for loading from different environment file paths
- Test-specific configuration in `tests/.env.test`

### Code Generation
- SQLC generates type-safe database code
- Run `task gen` after modifying SQL queries
- Generated code is in `internal/db/gen/`
- Don't edit generated files directly

### Security Best Practices
- Hash passwords before storage
- Use parameterized queries (SQLC handles this)
- Validate all input data
- Use HTTPS in production
- Set appropriate cookie flags
- Implement rate limiting for sensitive operations

### Testing Guidelines
- Write unit tests for business logic
- Write integration tests for API endpoints
- Use table-driven tests for multiple scenarios
- Mock external dependencies using `go.uber.org/mock`
- Test error handling paths
- Use test-specific database setup for integration tests
- Generate mocks using `go generate` command
- Test both happy paths and error scenarios

### Performance Considerations
- Use database connection pooling
- Implement proper indexing in database
- Use context timeouts for external calls
- Consider caching for frequently accessed data
- Monitor and log slow operations

## Bruno API Testing

The project includes Bruno collections for API testing located in `/bruno/`:

```bash
# API Collections Structure
/bruno/
├── admin/              # Admin-specific endpoints
├── auth/               # Authentication endpoints
├── claims/             # Claims management
├── hubs/               # Hub/location management
├── items/              # Item management
└── environments/       # Test environment configurations
```

- Use Bruno desktop app or CLI to run API tests
- Environment variables are configured in `bruno/environments/`
- Collections are organized by domain for comprehensive endpoint coverage

## Testing Infrastructure

The project includes a comprehensive testing setup:

### Test Database
- Separate test database container (`postgres-test`) on port 5433
- Test-specific configuration in `tests/.env.test`
- Automatic test database setup and teardown

### Test Files Structure
```
/tests/
├── .env.test          # Test environment variables
├── app.test.rsa       # Test private key for PASETO
├── app.test.pub.rsa   # Test public key for PASETO
├── setup.go           # Test initialization
└── test_db.go         # Test database utilities
```

### Mock Testing
- Uses `go.uber.org/mock` for generating mocks
- Mock interfaces for external dependencies
- Table-driven test patterns for multiple scenarios

### Test Patterns
```go
// Test setup example
func TestExampleFunction(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Test implementation
}
```

## Configuration Management

The project uses Koanf for configuration management:

```go
// Configuration loading
config := koanf.New(".")
config.LoadFile(envFile, toml.Parser())

// Access configuration values
dbHost := config.String("database.host")
dbPort := config.Int("database.port")
```

### Environment Files
- `.env` - Local development
- `.env.sample` - Template with required variables
- `tests/.env.test` - Test environment
- Support for custom environment file paths

### CORS Configuration
- CORS middleware is configured using `gin-contrib/cors`
- Configured for development and production environments
- Proper headers and origins handling

## File Structure

```
/                           # Project root
├── api/                    # API handlers and routes
│   ├── auth/              # Authentication endpoints
│   ├── claims/            # Claims management
│   ├── hubs/              # Location management
│   └── items/             # Item management
├── bruno/                  # Bruno API testing collections
│   ├── admin/             # Admin endpoints
│   ├── auth/              # Auth endpoints
│   ├── claims/            # Claims tests
│   ├── hubs/              # Hub tests
│   ├── items/             # Item tests
│   └── environments/      # Test environments
├── cmd/                    # Application entry point
├── internal/               # Internal application code
│   ├── db/                # Database layer
│   │   ├── gen/           # SQLC generated code
│   │   ├── migrations/     # Database migrations
│   │   └── queries/        # SQL queries
│   ├── logger/            # Logging utilities
│   ├── middleware/        # HTTP middleware
│   ├── models/            # Data models
│   └── seed/              # Database seeding
├── pkg/                    # Public utilities
├── tests/                  # Test infrastructure
│   ├── integration/       # Integration tests
│   └── setup files        # Test configuration
├── .vscode/               # VS Code configuration
└── tmp/                   # Temporary build directory
```

## Development Workflow

1. **Setup**: Run `task setup` to initialize the project
2. **Database**: Start services with `task docker:up`
3. **Migrations**: Apply with `task up`
4. **Seeding**: (Optional) Seed with `task db:seed` for development
5. **Development**: Use `task dev` for hot reload
6. **Code Generation**: Run `task gen` after SQL changes
7. **Quality**: Use `task fmt` and `task vet` before commits
8. **Testing**: Run `task test` for all tests, `task test:unit` or `task test:integration` for specific test types
9. **API Testing**: Use Bruno collections in `/bruno/` for manual API testing

## Key Dependencies

- **Gin**: Web framework
- **SQLC**: Type-safe SQL code generation
- **PGX**: PostgreSQL driver
- **PASETO**: Token-based authentication
- **Ozzo-validation**: Request validation
- **Zerolog**: Structured logging
- **Goose**: Database migrations
- **gin-contrib/cors**: CORS middleware
- **google/uuid**: UUID generation
- **pgvector**: Vector operations for PostgreSQL
- **koanf**: Configuration management
- **go-paseto**: PASETO token implementation
- **go.uber.org/mock**: Mocking framework for tests

## Database Schema

- Uses PostgreSQL with pgvector extension
- UUIDv7 for primary keys
- Comprehensive audit logging
- JSONB for flexible metadata
- Migration files in `internal/db/migrations/`
- Separate test database (`postgres-test`) on port 5433 for testing
- Drizzle Gateway container for database management

## Common Patterns

### Transaction Pattern
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

tx, err := cmd.DBPool.Begin(ctx)
if pkg.HandleDbTxnErr(c, err, "OPERATION") {
    return
}
defer pkg.RollbackTx(c, tx, ctx, "OPERATION")

// ... operations ...

err = tx.Commit(ctx)
if pkg.HandleDbTxnCommitErr(c, err, "OPERATION") {
    return
}
```

### Validation Pattern
```go
req, ok := pkg.ValidateRequest[models.RequestType](c)
if !ok {
    return
}
```

### Response Pattern
```go
c.JSON(http.StatusOK, gin.H{
    "message": "Success message",
})
logger.Log.SuccessCtx(c)
```

Remember to follow these guidelines and maintain consistency with the existing codebase. Always run formatting and static analysis before committing changes.