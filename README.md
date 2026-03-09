# Nyx Backend

## Description
Nyx is a lost-and-found platform backend that manages items, claims, hubs, and user interactions. It provides a RESTful API with PASETO-based authentication, S3-compatible image uploads via pre-signed URLs, and email-based OTP verification.

## Tech Stack
- **Language:** Go
- **Framework:** Gin
- **Database:** PostgreSQL with pgvector
- **Query Builder:** SQLC (type-safe SQL)
- **Authentication:** PASETO v4 tokens
- **Object Storage:** S3-compatible (Garage for local dev, Cloudflare R2 / Supabase for production)
- **Containerization:** Docker Compose
- **Task Runner:** [Taskfile](https://taskfile.dev)
- **Testing:** Go testing + Bruno API collections
- **Logging:** Zerolog
- **CI:** GitHub Actions

## Setup

### 1. Clone the Repository
```sh
git clone https://github.com/KiranRajeev-KV/nyx-backend.git
cd nyx-backend
```

### 2. Setup the project
Ensure you have Go, Docker, and [Taskfile](https://taskfile.dev/docs/installation) installed.
```sh
task setup
```
This installs Air (hot reload), Goose (migrations), Lefthook (git hooks), and sqlc (code generation).

### 3. Start Services
```sh
task docker:up      # Start PostgreSQL, Drizzle Gateway, and Garage
task garage:setup   # Provision Garage S3 (layout, key, bucket) — run once after first docker:up
task up             # Run database migrations
```

### 4. Environment Configuration
```sh
cp .env.sample .env
```
Edit `.env` with your settings. Key variables:
```bash
# Server
PORT=8080
ENVIRONMENT=development
COOKIE_DOMAIN=localhost
COOKIE_SECURE=false

# Database
GOOSE_DBSTRING="postgres://postgres:1234@localhost:5432/postgres"

# Email (optional - disable for local dev without email)
EMAIL_ENABLE=true
EMAIL_SMTP_HOST=smtp.gmail.com
EMAIL_SMTP_PORT=587
EMAIL_FROM_EMAIL=your-email@gmail.com
EMAIL_FROM_PASSWORD=your-app-password
EMAIL_FROM_NAME="Nyx System"

# S3 Storage (Garage for local dev)
S3_ENDPOINT="http://localhost:3900"
S3_REGION="us-east-1"
S3_BUCKET_NAME="nyx-items"
S3_ACCESS_KEY_ID="<from task garage:setup output>"
S3_SECRET_ACCESS_KEY="<from task garage:setup output>"
```

### 5. Run
```sh
task dev    # Hot reload with Air
task run    # Build and run
```

## API Documentation

### Authentication
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/auth/register` | User registration (sends OTP) | No |
| POST | `/auth/verify-otp` | Verify OTP to complete registration | Temp |
| POST | `/auth/resend-otp` | Resend OTP | Temp |
| POST | `/auth/login` | User login | No |
| POST | `/auth/refresh` | Refresh access token | Cookie |
| GET | `/auth/session` | Check current session | Yes |
| GET | `/auth/logout` | Logout (revoke tokens) | Yes |
| POST | `/auth/forgot-password` | Request password reset OTP | No |
| POST | `/auth/reset-password` | Reset password with OTP | Temp |

### Items
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/items/` | Get all items (with optional `?type=LOST\|FOUND`) | Yes |
| GET | `/items/:id` | Get item by ID | Yes |
| GET | `/items/me` | Get current user's items | User |
| POST | `/items/` | Create new item | User |
| POST | `/items/:id/image` | Get pre-signed URL for image upload | User (Owner) |
| PATCH | `/items/:id` | Update item | User (Owner) |
| PATCH | `/items/:id/status` | Update item status | User (Owner) |
| DELETE | `/items/:id` | Soft delete item | User (Owner) |

### Claims
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/claims/` | Create a claim on an item | User |
| GET | `/claims/me` | Get current user's claims | User |
| GET | `/claims/item/:id` | Get claims for a specific item | Yes |
| GET | `/claims/admin` | Get all claims | Admin |
| PATCH | `/claims/:id` | Process (approve/reject) a claim | Admin |

### Hubs
| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/hubs/` | Get all hubs | No |
| GET | `/hubs/:id` | Get hub by ID | No |
| POST | `/hubs/` | Create a hub | Admin |
| PATCH | `/hubs/:id` | Update a hub | Admin |
| DELETE | `/hubs/:id` | Delete a hub | Admin |

**Testing APIs**: Use Bruno collections in the `/bruno/` directory.

## Task Commands

### Setup & Development
| Command | Description |
|---------|-------------|
| `task setup` | Complete project setup (deps + env + docker) |
| `task deps` | Install Go dependencies and tools |
| `task env` | Create `.env` from `.env.sample` |
| `task dev` | Start dev server with hot reload (Air) |
| `task build` | Build to `./bin/nyx-backend` |
| `task run` | Build and run |

### Docker & Database
| Command | Description |
|---------|-------------|
| `task docker:up` | Start all containers (Postgres, Drizzle Gateway, Garage) |
| `task docker:down` | Stop and remove containers |
| `task garage:setup` | Provision Garage S3 (layout, key, bucket) |
| `task start` / `task stop` | Start/stop existing containers |
| `task docker:logs` | View database logs |
| `task docker:reset` | Reset database (⚠️ deletes data) |

### Migrations & Seeding
| Command | Description |
|---------|-------------|
| `task up` | Run all pending migrations |
| `task down` | Rollback last migration |
| `task status` | Show migration status |
| `task db:seed` | Seed database with dummy data |
| `task db:truncate` | Truncate all tables |
| `task db:rebuild` | Truncate and seed |

### Code Generation & Quality
| Command | Description |
|---------|-------------|
| `task gen` | Generate SQLC code |
| `task test` | Run all tests |
| `task test:coverage` | Run tests with coverage report |
| `task fmt` | Format Go code |
| `task lint` | Run golangci-lint |
| `task vet` | Run static analysis |

## Image Upload Flow

Items support image uploads via **pre-signed URLs**:

1. Client calls `POST /items/:id/image` with `{ "content_type": "image/png" }`
2. Backend returns a pre-signed PUT URL (valid for 15 minutes)
3. Client uploads the image directly to S3 using the URL
4. The item's `image_url_original` is updated in the database

For production, swap the S3 env vars to point at Cloudflare R2 or Supabase Storage — no code changes needed.

## Authors
- Kiran Rajeev K V — [GitHub](https://github.com/KiranRajeev-KV)
