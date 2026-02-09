# Nyx Backend

## Description
Nyx is a backend service designed to manage items, claims, and user interactions in a streamlined manner. It utilizes PostgreSQL for data storage and provides a RESTful API for client interactions.

## Tech Stack
- **Programming Language:** Go
- **Framework:** Gin
- **Database:** PostgreSQL with pgvector
- **ORM/Query Builder:** SQLC (type-safe SQL)
- **Authentication:** PASETO tokens
- **Containerization:** Docker
- **Task Management:** Taskfile
- **Testing:** Go testing + Bruno API collections
- **Logging:** Zerolog

## Setup
1. **Clone the Repository**
   ```sh
   git clone https://github.com/KiranRajeev-KV/nyx-backend.git
   cd nyx-backend
   ```

2. **Setup the project**
   Ensure you have Go, Docker, and [Taskfile](https://taskfile.dev/docs/installation) installed on your machine.
   ```sh
   task setup
   ```
   This will install air (for hot reload), goose (for migrations), lefthook (for git hooks), and sqlc (for code generation).

3. **Run Migrations**
   Apply the database migrations:
   ```sh
   task up
   ```

4. **Environment Configuration**
   Copy and configure your environment variables:
   ```sh
   cp .env.sample .env
   ```
   Edit `.env` with your database credentials and other settings:
   ```bash
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=1234
   DB_NAME=postgres
   
   # Server
   SERVER_PORT=8080
   
   # JWT/PASETO
   JWT_SECRET=your-secret-key
   
   # Other settings as needed
   ```

## Development
To start the development server with hot reload, run:
```sh
task dev
```
This will start the server and allow you to make changes without restarting it manually.

## API Documentation

### Authentication Endpoints
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/auth/register` | User registration | No |
| POST | `/auth/login` | User login | No |
| POST | `/auth/verify-otp` | Verify OTP | No |
| POST | `/auth/resend-otp` | Resend OTP | No |
| POST | `/auth/logout` | User logout | Yes |
| POST | `/auth/session` | Check session | Yes |
| POST | `/auth/refresh` | Refresh tokens | Yes |
| POST | `/auth/forgot-password` | Forgot password | No |
| POST | `/auth/reset-password` | Reset password | No |

### Items Endpoints
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/items` | Get all items | No |
| GET | `/items/:id` | Get item by ID | No |
| GET | `/items/user/:userId` | Get items by user | Yes |
| POST | `/items` | Create new item | Yes |
| PUT | `/items/:id` | Update item | Yes |
| DELETE | `/items/:id` | Delete item | Yes |
| PUT | `/items/:id/status` | Update item status | Yes |

### Claims Endpoints
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/claims` | Get all claims by user | Yes |
| GET | `/claims/item/:itemId` | Get claims for item | No |
| POST | `/claims` | Create new claim | Yes |

### Hubs Endpoints
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/hubs` | Get all hubs | No |
| GET | `/hubs/:id` | Get hub by ID | No |
| POST | `/hubs` | Create new hub | Yes |
| PUT | `/hubs/:id` | Update hub | Yes |
| DELETE | `/hubs/:id` | Delete hub | Yes |

### Admin Endpoints
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/admin/claims` | Get all claims | Admin |
| POST | `/admin/claims/:id/process` | Process claim | Admin |

**Testing APIs**: Use Bruno collections in `/bruno/` directory for API testing.

## Task Commands

### Setup & Dependencies
| Command | Description |
|---------|-------------|
| `task setup` | Complete project setup (deps + env + docker) |
| `task deps` | Install Go dependencies and tools |
| `task env` | Create .env file from .env.sample |

### Development
| Command | Description |
|---------|-------------|
| `task dev` | Start dev server with hot reload (Air) |
| `task build` | Build application to ./bin/nyx-backend |
| `task run` | Build and run the application |

### Database Management
| Command | Description |
|---------|-------------|
| `task docker:up` | Start PostgreSQL and Drizzle Gateway |
| `task docker:down` | Stop and remove containers |
| `task start` | Start existing containers |
| `task stop` | Stop containers |
| `task docker:logs` | View database logs |
| `task docker:reset` | Reset database (⚠️ deletes data) |

### Database Migrations
| Command | Description |
|---------|-------------|
| `task up` | Run all pending migrations |
| `task down` | Rollback last migration |
| `task status` | Show migration status |

### Database Seeding
| Command | Description |
|---------|-------------|
| `task db:seed` | Seed database with dummy data |
| `task db:truncate` | Truncate all tables |
| `task db:rebuild` | Truncate and seed database |

### Code Generation
| Command | Description |
|---------|-------------|
| `task gen` | Generate SQLC code |
| `task gen:watch` | Watch changes and regenerate |

### Testing
| Command | Description |
|---------|-------------|
| `task test` | Run all tests |
| `task test:coverage` | Run tests with coverage report |
| `task test:unit` | Run unit tests only |
| `task test:integration` | Run integration tests only |

### Code Quality
| Command | Description |
|---------|-------------|
| `task fmt` | Format Go code |
| `task vet` | Run static analysis |
| `task lint` | Run golangci-lint |

## IMPROVEMENTS TO BE MADE
- [ ] Update the logic of `/verify-otp` endpoint
- [ ] Add `/resend-otp` endpoint
- [ ] Add `/reset-password` endpoint
- [ ] Add `/refresh` endpoint for token refresh
- [ ] Setup Mailer service for sending OTPs
- [ ] Setup CI pipeline
- [x] Add seeding scripts for initial data population
- [ ] Update logic of POST `/items` to include image upload
- [ ] Add POST `/items/:id/image` endpoint to update item images
- [ ] Generate redacted image URLs after storing original images


## Authors
- Kiran Rajeev K V - [GitHub](www.github.com/KiranRajeev-KV)



