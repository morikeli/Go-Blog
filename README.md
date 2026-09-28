# Go Blog API

## 1. Project Name

**Go Blog API** (`golang-rest-std-lib`)

## 2. Overview

Go Blog API is a REST API written in Go using the standard library HTTP server and `http.ServeMux`. It provides authentication and user-profile operations backed by PostgreSQL, with Redis used for caching and refresh-token revocation.

The development environment runs three Docker services:

- Go API with Air hot reloading
- PostgreSQL 16
- Redis 7

## 3. Developer Instructions

### Directory Structure

```text
.
├── main.go                 # Application entry point
├── internal/
│   ├── config/             # Environment configuration
│   ├── db/                 # Database and Redis connections, schema, migrations
│   ├── dtos/               # Request and response types
│   ├── handlers/           # HTTP handlers
│   ├── middlewares/        # Authentication middleware
│   ├── models/             # Domain models
│   ├── routes/             # Route registration
│   ├── store/              # sqlc-generated database access
│   └── utils/              # JWT, password, and validation helpers
├── scripts/                # Database initialization scripts
├── docker-compose.yaml     # API, PostgreSQL, and Redis services
├── Dockerfile              # Production image
├── Dockerfile.dev          # Development image with Air
├── Makefile                # Common development commands
└── sqlc.yaml               # sqlc configuration
```

### Prerequisites

- Docker Engine
- Docker Compose v2 (the `docker compose` command)
- Make

Go 1.27.1 is declared in `go.mod`. Go is not required when running the application through Docker, but it is required for local Go tooling.

### Environment Setup

Create the environment file before starting the application:

```sh
cp .env.example .env
```

Fill in the values in `.env`, especially `SERVER_PORT`, `DATABASE_URL`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `REDIS_ADDRESS`, `SECRET_KEY`, and `JWT_ISSUER`. When running through Docker Compose, use `db` as the database host and `db:5432` in the database URL. Use `redis:6379` for the Redis address.

### Run with Docker

Start the API and its dependencies:

```sh
make run
```

This runs `docker compose up --build`. The API is available at `http://localhost:${SERVER_PORT}`. The default Compose mappings expose PostgreSQL on host port `5433` and Redis on host port `6379`.

Useful commands:

```sh
docker compose up --build       # Start or rebuild all services
docker compose down             # Stop services
docker compose down -v          # Stop services and remove PostgreSQL data
make build                      # Build the Go binary locally
make deps                       # Tidy and download Go dependencies
make fmt                        # Format Go source files
make help                       # List Make targets
```

### Docker Installation

Install Docker Engine and the Compose plugin using the official instructions for your operating system:

<https://docs.docker.com/engine/install/>

Verify the installation:

```sh
docker --version
docker compose version
```

On Linux, ensure your user can run Docker without `sudo`, or run the Docker commands with the permissions required by your installation.

## 4. User Instructions

### Consuming the API

The API returns JSON. Successful responses use this shape:

```json
{
  "message": "...",
  "data": {}
}
```

Errors use:

```json
{
  "message": "..."
}
```

#### Health Check

```sh
curl http://localhost:8080/health
```

#### Authentication

Create an account:

```sh
curl -X POST http://localhost:8080/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","email":"alice@example.com","password":"password123"}'
```

Log in and save the refresh-token cookie:

```sh
curl -i -c cookies.txt -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"alice","password":"password123"}'
```

The login response contains a 15-minute `access_token`. Send it as a bearer token to protected routes:

```sh
curl http://localhost:8080/user/profile \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN'
```

Refresh the access token using the saved cookie:

```sh
curl -b cookies.txt -c cookies.txt -X POST http://localhost:8080/auth/refreshToken
```

Log out and revoke the refresh token:

```sh
curl -b cookies.txt -c cookies.txt -X POST http://localhost:8080/auth/logout
```

#### Endpoints

| Method | Path | Authentication | Description |
| --- | --- | --- | --- |
| `GET` | `/health` | No | Check API health |
| `POST` | `/auth/signup` | No | Create a user (`username`, `email`, `password`) |
| `POST` | `/auth/login` | No | Log in (`username`, `password`) |
| `POST` | `/auth/refreshToken` | Refresh cookie | Issue a new access token |
| `POST` | `/auth/logout` | Refresh cookie if available | Revoke the refresh token |
| `GET` | `/users?limit=10&offset=0` | Bearer token | List users with pagination |
| `GET` | `/user/profile` | Bearer token | Get the current user profile |
| `PUT` | `/user/profile` | Bearer token | Update `username` and/or upload `profile_picture` |

The profile update endpoint expects `multipart/form-data`, for example:

```sh
curl -X PUT http://localhost:8080/user/profile \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
  -F 'username=alice-updated' \
  -F 'profile_picture=@/path/to/photo.jpg'
```

The refresh cookie is marked `Secure` by the application. For browser clients, use HTTPS in environments where the cookie must be sent; API clients can inspect and manage the cookie returned by login.

## 5. Contribution

1. Create a focused branch for your change.
2. Update or add tests for changed behavior.
3. Run `make fmt`, `go test ./...`, and `go vet ./...` before opening a pull request.
4. Keep commits focused and explain any database or API contract changes.
5. Open a pull request with a summary, testing details, and any required environment or migration changes.

## Show some love ♥️
Don't forget to star the repo 🌟😉