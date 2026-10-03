# Go Blog API

## Overview

Go Blog API is a REST API written in Go using the standard library HTTP server and `http.ServeMux`. It provides user authentication and profile-management operations backed by PostgreSQL, with Redis used for caching and refresh-token revocation.

The development environment runs three Docker services:

- Go API with Air hot reloading
- PostgreSQL 16
- Redis 7

The project follows a layered structure that separates HTTP handlers, business logic, repositories, database access, models, DTOs, middleware, and utility functions.

### Key features

- 🔐 **JWT Authentication** – Short-lived access tokens and refresh tokens with token-type validation.
- 🍪 **Secure Refresh Tokens** – Refresh tokens are stored in `HttpOnly`, `Secure` cookies and rotated when refreshed.
- 🚫 **Refresh-token Revocation** – Redis is used to blacklist revoked refresh tokens.
- 👤 **User Management** – Create accounts, authenticate users, list users, and manage the current user's profile.
- 📄 **Profile Picture Uploads** – Profile images are validated and uploaded to Cloudinary.
- ⚡ **Redis Caching** – Authenticated user profiles are cached for faster repeated access.
- 📑 **Pagination** – User listing supports `limit`, `offset`, and pagination metadata.
- 🛡️ **Request Validation** – Request payloads and uploaded profile images are validated before processing.
- 🐘 **PostgreSQL Persistence** – Relational data is stored in PostgreSQL using `pgx/v5`.
- 🧩 **SQLC-generated Data Access** – SQL queries are compiled into type-safe Go database code.
- 🐳 **Dockerized Development** – API, PostgreSQL, and Redis can be started together with Docker Compose.
- 🔥 **Hot Reloading** – Air is used during Docker-based development.

## Tech Stack

- 🐹 **Language:** Go 1.27.1
- 🌐 **HTTP:** Go standard library `net/http`
- 🗄️ **Database:** PostgreSQL 16
- ⚡ **Cache / Token Revocation:** Redis 7
- 🔑 **Authentication:** JWT using `golang-jwt/jwt`
- 🖼️ **Image Storage:** Cloudinary
- 🧬 **Database Driver:** pgx/v5
- 🧾 **Database Code Generation:** sqlc
- 🔥 **Development:** Air
- 🐳 **Containers:** Docker + Docker Compose
- 🛠️ **Build / Commands:** Make

### Go packages

- `github.com/golang-jwt/jwt/v5` – JWT creation and verification
- `github.com/jackc/pgx/v5` – PostgreSQL connectivity
- `github.com/redis/go-redis/v9` – Redis client
- `github.com/cloudinary/cloudinary-go/v2` – Profile image uploads
- `github.com/go-playground/validator/v10` – Request validation
- `golang.org/x/crypto` – Password hashing utilities
- `golang.org/x/image` – Image processing/validation support
- `github.com/joho/godotenv` – Environment configuration

## Architecture

The application uses a layered architecture:

```text
HTTP Request
    │
    ▼
Routes
    │
    ▼
Middleware
    │
    ▼
Handlers
    │
    ▼
Services
    │
    ▼
Repositories
    │
    ▼
SQLC-generated Store
    │
    ▼
PostgreSQL

Handlers / Services
    │
    ├──► Redis
    │
    └──► Cloudinary
```

### Directory structure

```text
.
├── cmd/
│   └── main.go                  # Application entry point
├── internal/
│   ├── config/                  # Environment configuration
│   ├── db/                      # Database, Redis, schema and queries
│   │   ├── migrations/
│   │   ├── queries/
│   │   └── schema/
│   ├── dtos/                    # Request and response DTOs
│   │   ├── requests/
│   │   └── responses/
│   ├── handlers/                # HTTP request handlers
│   ├── middlewares/             # Authentication middleware
│   ├── models/                  # Domain models
│   ├── repositories/            # Data-access abstraction
│   ├── routes/                  # HTTP route registration
│   ├── services/                # Business logic
│   ├── store/                   # SQLC-generated database access
│   └── utils/                   # JWT, password, validation and request helpers
├── scripts/
│   └── db-init.sh               # Database initialization
├── Dockerfile                   # Production image
├── Dockerfile.dev               # Development image with Air
├── docker-compose.yaml          # API, PostgreSQL and Redis services
├── Makefile                     # Common development commands
└── sqlc.yaml                    # SQLC configuration
```

## Product thinking

The project is built around a simple backend-first goal: provide a clean foundation for a blog application while focusing on secure authentication, user management, persistence, caching, and external media storage.

The current implementation focuses on the account and user-management foundation:

- 🔐 Authentication is separated from user-management logic.
- 🧱 Business logic lives in services rather than HTTP handlers.
- 🗃️ Database access is isolated behind repositories and SQLC-generated code.
- ⚡ Redis reduces repeated database reads for authenticated user profiles.
- 🖼️ Cloudinary handles profile-picture storage rather than storing image files directly in PostgreSQL.
- 🐳 Docker Compose provides a repeatable development environment.

The database schema also includes a `blogs` table related to users, providing a foundation for blog functionality as the API evolves.

## Developer instructions

> [!WARNING]
> The recommended way to run this project is with Docker Compose. When using the Docker-based setup, you do not need to install Go locally to run the API.
>
> You will need Docker Engine and Docker Compose v2. Make is recommended for the provided development commands.

### Prerequisites

- Docker Engine
- Docker Compose v2 (`docker compose`)
- Make
- Go 1.27.1 if you want to run Go tooling or the application outside Docker

Install Docker using the official documentation:

<https://docs.docker.com/engine/install/>

Verify your installation:

```bash
docker --version
docker compose version
make --version
```

## Installation guide

### 1. Clone the repository

```bash
git clone https://github.com/morikeli/golangrestapi.git
cd golangrestapi
```

### 2. Create the environment file

Copy the example environment file:

```bash
cp .env.example .env
```

Update `.env` with the values required by your environment.

Example:

```env
SERVER_PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug

SECRET_KEY=your-secret-key
JWT_ISSUER=go-blog-api

CLOUDINARY_URL=cloudinary://...

# Redis
REDIS_ADDRESS=redis:6379
REDIS_PASSWORD=

# Database
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=go_blog
DB_PORT=5432
DB_HOST=db
DATABASE_URL=postgres://postgres:your-password@db:5432/go_blog
```

> [!NOTE]
> When running through Docker Compose, use `db` as the PostgreSQL host and `redis:6379` as the Redis address. For a locally running Go application, use `localhost` with the appropriate host ports.

### 3. Run with Docker

The recommended development command is:

```bash
make run
```

This starts the application with:

```bash
docker compose up --build
```

The Docker Compose stack starts:

- Go API → `http://localhost:${SERVER_PORT}`
- PostgreSQL → `localhost:5433`
- Redis → `localhost:6379`

Useful Docker commands:

```bash
docker compose up --build       # Start or rebuild all services
docker compose up -d            # Start services in detached mode
docker compose logs -f          # Follow service logs
docker compose down             # Stop services
docker compose down -v          # Stop services and remove PostgreSQL data
```

## Make commands

The project includes a `Makefile` for common development tasks:

```bash
make build          # Build the Go binary
make clean          # Remove build artifacts
make deps           # Tidy and download Go dependencies
make fmt            # Format Go source files
make help           # Show available Make targets
make migrate-up     # Run database migrations with dbmate
make migrate-down   # Roll back database migrations
make run            # Start the Docker Compose development environment
```

## API usage

The API returns JSON responses.

### Successful response

```json
{
  "message": "Operation completed successfully!",
  "data": {}
}
```

### Error response

```json
{
  "message": "Something went wrong!"
}
```

## Authentication flow

The API uses two JWT token types:

- **Access token** – valid for 15 minutes.
- **Refresh token** – valid for 7 days.

After login, the access token is returned in the JSON response while the refresh token is stored in an `HttpOnly` and `Secure` cookie.

Refresh tokens are single-use. When `/auth/token/refresh` is called:

1. The refresh-token cookie is read.
2. The JWT signature, issuer, expiry, and token type are verified.
3. Redis is checked to ensure the refresh token has not been revoked.
4. The current refresh token is revoked.
5. A new access token is generated.
6. A new refresh token is generated and placed in the cookie.

This provides refresh-token rotation and allows revoked refresh tokens to be rejected.

## API endpoints

| Method | Path | Authentication | Description |
| --- | --- | --- | --- |
| `GET` | `/health` | No | Check API health |
| `POST` | `/auth/signup` | No | Create a new user |
| `POST` | `/auth/login` | No | Authenticate a user |
| `POST` | `/auth/token/refresh` | Refresh cookie | Rotate the refresh token and issue a new access token |
| `POST` | `/auth/logout` | Refresh cookie if available | Revoke the refresh token and clear the cookie |
| `GET` | `/users` | Bearer token | Retrieve paginated users |
| `GET` | `/user/me` | Bearer token | Retrieve the authenticated user's profile |
| `PATCH` | `/user/me` | Bearer token | Update username and/or profile picture |

### Health check

```bash
curl http://localhost:8080/health
```

### Create an account

```bash
curl -X POST http://localhost:8080/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "password123"
  }'
```

### Login

```bash
curl -i -c cookies.txt -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "alice",
    "password": "password123"
  }'
```

The login response contains the access token. The refresh token is returned as an `HttpOnly` cookie.

### Get current user

```bash
curl http://localhost:8080/user/me \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN'
```

### Refresh access token

Use the refresh-token cookie saved during login:

```bash
curl -b cookies.txt -c cookies.txt \
  -X POST http://localhost:8080/auth/token/refresh
```

### Logout

```bash
curl -b cookies.txt -c cookies.txt \
  -X POST http://localhost:8080/auth/logout
```

### List users

The endpoint supports pagination using `limit` and `offset`.

```bash
curl 'http://localhost:8080/users?limit=10&offset=0' \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN'
```

The default limit is `10`, and the maximum limit is `100`.

### Update user profile

The profile endpoint accepts `multipart/form-data`.

Update a username:

```bash
curl -X PATCH http://localhost:8080/user/me \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
  -F 'username=alice-updated'
```

Update a profile picture:

```bash
curl -X PATCH http://localhost:8080/user/me \
  -H 'Authorization: Bearer YOUR_ACCESS_TOKEN' \
  -F 'profile_picture=@/path/to/photo.jpg'
```

The supported profile image formats are JPEG, PNG, and WebP, with a maximum profile-image size of 5 MB.

## Database

PostgreSQL is used as the primary data store.

The database schema currently includes:

- `users` – user account and profile information
- `blogs` – blog records associated with users

The `blogs.user_id` column references `users.id` with cascading deletion.

SQLC generates type-safe Go database access code from the SQL files in:

```text
internal/db/schema/
internal/db/queries/
```

The generated code is stored in:

```text
internal/store/
```

### PostgreSQL connection

When the Docker stack is running, PostgreSQL is exposed on:

```text
localhost:5433
```

The PostgreSQL container uses:

```text
db:5432
```

for internal Docker networking.

You can connect to the running database with:

```bash
docker exec -it goapp_postgres_db psql -U postgres -d go_blog
```

Use the actual `DB_USER` and `DB_NAME` values configured in `.env` if they differ from the example above.

## Redis

Redis is used for two purposes:

1. **User profile caching**
   - Authenticated user profiles are cached for 15 minutes.
   - The cache is invalidated after a profile update.

2. **Refresh-token revocation**
   - Revoked refresh tokens are stored using their JWT ID.
   - The Redis entry expires when the corresponding refresh token expires.

Redis is exposed on:

```text
localhost:6379
```

and is available to the Go API inside Docker as:

```text
redis:6379
```

## Security considerations

The project includes several security-focused measures:

- Passwords are hashed before being stored.
- Access and refresh tokens have different token types.
- JWT issuer validation is enabled.
- JWT signing is restricted to HS256.
- Refresh tokens are stored in `HttpOnly` and `Secure` cookies.
- Refresh tokens are rotated after successful refresh.
- Revoked refresh tokens are blacklisted in Redis.
- JSON request bodies are size-limited.
- Profile image uploads are size-limited and type-validated.
- Protected user routes require authentication middleware.
- Profile responses exclude the stored password.

> [!NOTE]
> The current configuration is intended for development and learning workflows. Production deployments should review HTTPS, secrets management, cookie configuration, logging, infrastructure security, database permissions, and other operational controls before deployment.

## Contributor expectations

If you want to contribute to the project:

1. Create a focused branch from the main branch.
2. Make the relevant changes.
3. Update or add tests for changed behavior.
4. Run formatting and validation checks locally.
5. Keep database and API contract changes documented.
6. Open a pull request with a clear summary of the changes and testing performed.

Recommended checks:

```bash
make fmt
go test ./...
go vet ./...
```

Keep environment configuration in `.env` and never commit production secrets.

## 🐞 Known issues / limitations

1. **Blog endpoints are not currently exposed** – The database schema contains a `blogs` table, but the current HTTP routes focus on authentication and user-management operations.
2. **Development-oriented configuration** – The Docker Compose setup is intended primarily for local development.
3. **Secure cookies require HTTPS in production** – Refresh cookies are configured with the `Secure` attribute.
4. **Environment configuration is required** – The application depends on values in `.env`, including database, Redis, JWT, and Cloudinary configuration.

## 🙏 Show some love

Don't forget to star the repo 🌟😉
