# VPS Panel - Backend API

Go-based backend API for VPS Panel deployment platform.

## 🏗️ Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/            # HTTP request handlers
│   │   │   ├── auth.go          # Authentication endpoints
│   │   │   ├── project.go       # Project management
│   │   │   ├── deployment.go    # Deployment operations
│   │   │   └── webhook.go       # Git webhook handling
│   │   ├── middleware/          # HTTP middleware
│   │   │   └── auth.go          # JWT authentication
│   │   └── routes/              # Route definitions
│   │       └── routes.go
│   ├── config/                  # Configuration management
│   │   └── config.go
│   ├── database/                # Database setup
│   │   └── database.go
│   ├── models/                  # Database models (GORM)
│   │   ├── user.go
│   │   ├── project.go
│   │   ├── deployment.go
│   │   └── environment.go
│   └── services/                # Business logic
│       ├── deployment/          # Deployment orchestration
│       │   └── deployment.go
│       ├── docker/              # Docker operations
│       │   └── docker.go
│       ├── git/                 # Git operations
│       │   └── git.go
│       └── caddy/               # Caddy config management
│           └── caddy.go
├── pkg/                         # Public packages
│   └── utils/
├── migrations/                  # Database migrations
├── .env.example                 # Environment variables template
├── Dockerfile                   # Docker build instructions
├── go.mod                       # Go module definition
└── README.md
```

## 🚀 Getting Started

### Prerequisites

- Go 1.23 or higher
- Docker (for deployment features)
- Redis (for background jobs)

### Installation

1. **Install dependencies:**

```bash
go mod download
```

2. **Configure environment:**

```bash
cp .env .env
# Edit .env with your settings
```

3. **Run the server:**

```bash
go run cmd/server/main.go
```

The API will be available at `http://localhost:8080`

### Development

**With hot reload (using air):**

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

## 📡 API Endpoints

### Authentication

```
POST   /api/v1/auth/register      # Register new user
POST   /api/v1/auth/login         # Login
POST   /api/v1/auth/refresh       # Refresh token
```

### Users

```
GET    /api/v1/users/me           # Get current user
PUT    /api/v1/users/me           # Update profile
```

### Projects

```
GET    /api/v1/projects           # List projects
POST   /api/v1/projects           # Create project
GET    /api/v1/projects/:id       # Get project
PUT    /api/v1/projects/:id       # Update project
DELETE /api/v1/projects/:id       # Delete project
```

### Deployments

```
GET    /api/v1/projects/:id/deployments              # List deployments
POST   /api/v1/projects/:id/deployments              # Create deployment
GET    /api/v1/projects/:id/deployments/:deployId    # Get deployment
POST   /api/v1/projects/:id/deployments/:deployId/cancel  # Cancel
GET    /api/v1/projects/:id/deployments/:deployId/logs    # Get logs
```

### Environment Variables

```
GET    /api/v1/projects/:id/environments        # List env vars
POST   /api/v1/projects/:id/environments        # Add env var
PUT    /api/v1/projects/:id/environments/:envId # Update env var
DELETE /api/v1/projects/:id/environments/:envId # Delete env var
```

### Domains

```
GET    /api/v1/projects/:id/domains           # List domains
POST   /api/v1/projects/:id/domains           # Add domain
DELETE /api/v1/projects/:id/domains/:domainId # Delete domain
```

### Webhooks

```
POST   /api/v1/webhooks/github      # GitHub webhook
POST   /api/v1/webhooks/gitlab      # GitLab webhook
POST   /api/v1/webhooks/bitbucket   # Bitbucket webhook
```

## 🔧 Configuration

Configuration is done via environment variables. See `.env.example` for all options.

### Key Configuration Options

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DB_DRIVER` | Database driver (sqlite/postgres) | `sqlite` |
| `DB_PATH` | SQLite database path | `./data/vps-panel.db` |
| `DOCKER_HOST` | Docker socket path | `unix:///var/run/docker.sock` |
| `REDIS_ADDR` | Redis address | `localhost:6379` |
| `CADDY_CONFIG_PATH` | Caddy config directory | `/etc/caddy/sites` |
| `JWT_SECRET` | JWT signing secret | *required* |
| `WEBHOOK_SECRET` | Git webhook secret | *required* |

## 🗄️ Database

### SQLite (Development)

Default database for development. No additional setup required.

```bash
DB_DRIVER=sqlite
DB_PATH=./data/vps-panel.db
```

### PostgreSQL (Production)

For production deployments:

```bash
DB_DRIVER=postgres
DB_HOST=localhost
DB_PORT=5432
DB_NAME=vps_panel
DB_USER=postgres
DB_PASSWORD=your-password
```

### Migrations

GORM AutoMigrate runs automatically on startup. Models are defined in `internal/models/`.

## 🐳 Docker

### Build Image

```bash
docker build -t vps-panel-backend .
```

### Run Container

```bash
docker run -p 8080:8080 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v $(pwd)/data:/data \
  -e JWT_SECRET=your-secret \
  vps-panel-backend
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/services/git
```

## 📝 Adding New Features

### Adding a New API Endpoint

1. **Define model** in `internal/models/`
2. **Add database migration** (AutoMigrate in `database.go`)
3. **Create handler** in `internal/api/handlers/`
4. **Register route** in `internal/api/routes/routes.go`

Example:

```go
// 1. Model (internal/models/feature.go)
type Feature struct {
    ID        uint      `gorm:"primarykey"`
    Name      string    `gorm:"not null"`
    CreatedAt time.Time
}

// 2. Handler (internal/api/handlers/feature.go)
func (h *FeatureHandler) Create(c *fiber.Ctx) error {
    // Implementation
}

// 3. Route (internal/api/routes/routes.go)
features := protected.Group("/features")
features.Post("/", featureHandler.Create)
```

## 🛠️ Development Tools

### Recommended Tools

- **Air**: Hot reload for Go
- **golangci-lint**: Linting
- **go-swagger**: API documentation
- **Docker Desktop**: Container management

### Code Style

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run
```

## 🔒 Security

- JWT tokens for authentication
- Bcrypt for password hashing
- HMAC SHA-256 for webhook verification
- Role-based access control
- Input validation

## 📊 Performance

- Fiber framework (one of the fastest Go frameworks)
- Connection pooling for database
- Redis for caching and queuing
- Efficient Docker SDK usage

## 🚧 Roadmap

- [ ] WebSocket support for real-time logs
- [ ] Background job processing with Asynq
- [ ] Metrics and monitoring
- [ ] Multi-VPS support
- [ ] Custom build scripts
- [ ] Rollback functionality

## 📞 Support

For issues and questions, please open an issue on GitHub.
