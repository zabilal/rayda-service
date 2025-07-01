# Rayda Service

A multi-tenant SaaS platform with external integrations built with Go, PostgreSQL, Redis, and Kafka.

## Features

- Multi-tenant architecture with data isolation
- JWT-based authentication with role management
- RESTful API for user and organization management
- Webhook processing for external service integration
- Event-driven architecture with Kafka
- Rate limiting and request validation
- Health monitoring and metrics
- **Caching Layer** with Redis and in-memory backends
  - Configurable TTL for cached items
  - Automatic cache invalidation
  - No-op implementation for development and testing

## Prerequisites

- Go 1.21 or later
- PostgreSQL 13+
- Redis 6+
- Kafka 3.0+
- Docker (optional, for development environment)

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/rayda-service.git
cd rayda-service
```

### 2. Set up environment variables

Copy the example environment file and update the values as needed:

```bash
cp .env.example .env
```

### 3. Install dependencies

```bash
go mod download

# Install additional tools for development
# go install github.com/swaggo/swag/cmd/swag@latest  # For API documentation
go install github.com/golang/mock/mockgen@latest    # For generating mocks
```

### 4. Configuration

The application can be configured using environment variables. Copy the example environment file and update the values as needed:

```bash
cp .env.example .env
```

Key configuration options:

- **Server**: `SERVER_PORT`, `ENVIRONMENT`, `SHUTDOWN_TIMEOUT`
- **Database**: `DB_*` variables for PostgreSQL connection
- **Redis**: `REDIS_*` variables for Redis connection
- **JWT**: `JWT_*` variables for authentication
- **Cache**: `CACHE_*` variables to control caching behavior
  - `CACHE_ENABLED`: Set to `false` to disable caching
  - `CACHE_TTL`: Time-to-live for cache entries in seconds (default: 300)
  - `CACHE_KEY_PREFIX`: Prefix for all cache keys (default: "rayda")

### 5. Run the application

```bash
# Development mode with in-memory cache
go run cmd/api/main.go

# Production mode with Redis cache
ENVIRONMENT=production go run cmd/api/main.go
```

The API server will start on `http://localhost:8080` by default.

## API Documentation

Once the server is running, you can access the API documentation at:
- Swagger UI: `http://localhost:8080/swagger/index.html`
- OpenAPI spec: `http://localhost:8080/swagger/doc.json`

## Caching

The application includes a flexible caching layer with the following features:

- **Multiple Backends**:
  - **Redis**: Used in production for distributed caching
  - **In-Memory**: Used in development for simplicity
  - **No-Op**: Used when caching is disabled

- **Automatic Caching**:
  - Audit log queries are automatically cached
  - Cache keys are namespaced by tenant and query parameters
  - Automatic cache invalidation on data modification

- **Configuration**:
  - Enable/disable caching via `CACHE_ENABLED`
  - Set TTL with `CACHE_TTL` in seconds
  - Configure key prefix with `CACHE_KEY_PREFIX`

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# Run integration tests (requires database and Redis)
ENVIRONMENT=test go test -tags=integration ./...
```

### Code generation

To generate mock implementations for testing:

```bash
go install github.com/golang/mock/mockgen@latest
go generate ./...
```

### Linting and formatting

```bash
go fmt ./...
staticcheck ./...
```

## Deployment

### Building the application

```bash
go build -o bin/rayda-service cmd/api/main.go
```

### Using Docker

```bash
docker build -t rayda-service .
docker run -p 8080:8080 --env-file .env rayda-service
```

## Project Structure

```
.
├── cmd/                  # Application entry points
│   └── api/              # Main API server
├── internal/             # Private application code
│   ├── api/              # HTTP handlers and routes
│   ├── config/           # Configuration management
│   ├── model/            # Domain models
│   ├── repository/       # Data access layer
│   ├── service/          # Business logic
│   └── pkg/              # Reusable packages
│       ├── auth/         # Authentication and authorization
│       ├── logger/       # Logging utilities
│       ├── validator/    # Request validation
│       ├── webhook/      # Webhook processing
│       └── event/        # Event publishing/subscribing
├── migrations/           # Database migrations
├── scripts/              # Utility scripts
└── test/                 # Test files
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
