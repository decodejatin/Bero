# Bero Backend

A high-performance Go backend for the Bero home services marketplace.

## Tech Stack

- **Go 1.23+** - Language
- **Echo v4** - Web framework
- **GORM** - ORM
- **PostgreSQL** - Database
- **JWT** - Authentication

## Quick Start

### Prerequisites

- Go 1.23+
- Docker & Docker Compose
- PostgreSQL (or use Docker)

### Setup

1. **Start PostgreSQL with Docker:**
   ```bash
   docker-compose up -d postgres
   ```

2. **Install dependencies:**
   ```bash
   go mod tidy
   ```

3. **Run the server:**
   ```bash
   go run cmd/server/main.go
   ```

4. **Access the API:**
   - Health check: http://localhost:8080/health
   - API base: http://localhost:8080/api/v1

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/send-otp` | Send OTP to phone |
| POST | `/api/v1/auth/verify-otp` | Verify OTP and get tokens |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Logout (protected) |
| GET | `/api/v1/auth/me` | Get current user (protected) |

### Jobs
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/jobs` | Create a job |
| GET | `/api/v1/jobs` | List available jobs |
| GET | `/api/v1/jobs/my` | Get my jobs |
| GET | `/api/v1/jobs/:id` | Get job details |
| POST | `/api/v1/jobs/:id/accept` | Accept a job (worker) |
| POST | `/api/v1/jobs/:id/start` | Start a job |
| POST | `/api/v1/jobs/:id/complete` | Complete a job |
| POST | `/api/v1/jobs/:id/cancel` | Cancel a job |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | 8080 | Server port |
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | password | Database password |
| `DB_NAME` | bero | Database name |
| `JWT_SECRET` | (random) | JWT signing secret |

## Project Structure

```
backend/
├── cmd/server/          # Application entry point
├── config/              # Configuration loading
├── internal/
│   ├── api/             # HTTP handlers & routing
│   ├── domain/          # Domain models
│   ├── repository/      # Database operations
│   └── service/         # Business logic
├── pkg/database/        # Database connection
├── migrations/          # SQL migrations
└── docker-compose.yml   # Docker setup
```

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o bero-server cmd/server/main.go
```

## License

Proprietary - Bero
