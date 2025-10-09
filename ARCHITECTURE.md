# VPS Panel - Architecture Documentation

## Overview

VPS Panel is a self-hosted deployment platform built with a modern tech stack focusing on performance, scalability, and ease of use.

## Tech Stack

### Frontend
- **Framework**: SvelteKit (latest with Svelte 5 runes)
- **Styling**: Tailwind CSS v4.1
- **Build Tool**: Vite
- **Language**: TypeScript

### Backend
- **Language**: Go 1.23
- **Web Framework**: Fiber v2
- **ORM**: GORM
- **Database**: SQLite (dev) / PostgreSQL (production)
- **Queue**: Asynq (Redis-based)
- **Container**: Docker SDK
- **Git**: go-git

### Infrastructure
- **Reverse Proxy**: Caddy (automatic HTTPS)
- **Containerization**: Docker
- **Queue/Cache**: Redis

## System Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                         User Browser                          │
└────────────────┬─────────────────────────────────────────────┘
                 │
                 ▼
┌──────────────────────────────────────────────────────────────┐
│                      Caddy (Reverse Proxy)                    │
│  - Automatic HTTPS (Let's Encrypt)                           │
│  - Route to projects based on domain                         │
│  - Load balancing                                            │
└────────┬────────────────────────────────────────────┬────────┘
         │                                             │
         │ Panel UI                                    │ User Projects
         ▼                                             ▼
┌─────────────────┐                          ┌──────────────────┐
│ Frontend        │                          │ Deployed Apps    │
│ (SvelteKit)     │                          │ (Docker)         │
│ Port: 3000      │                          │ Port: 3001-9999  │
└────────┬────────┘                          └──────────────────┘
         │
         │ API Calls
         ▼
┌─────────────────────────────────────────────────────────────┐
│                    Backend API (Go Fiber)                    │
│  Port: 8080                                                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ API Layer                                              │ │
│  │  - Authentication (JWT)                                │ │
│  │  - RESTful endpoints                                   │ │
│  │  - Webhook handlers                                    │ │
│  │  - WebSocket (for logs)                                │ │
│  └──────────────────┬─────────────────────────────────────┘ │
│                     ▼                                        │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Service Layer                                          │ │
│  │  - Deployment Service (orchestration)                  │ │
│  │  - Docker Service (container management)               │ │
│  │  - Git Service (repo operations)                       │ │
│  │  - Caddy Service (config generation)                   │ │
│  └──────────────────┬─────────────────────────────────────┘ │
│                     ▼                                        │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Data Layer                                             │ │
│  │  - GORM (ORM)                                          │ │
│  │  - Models (User, Project, Deployment, etc.)           │ │
│  └────────────────────────────────────────────────────────┘ │
└──────┬────────────────────────────────────────────┬─────────┘
       │                                             │
       ▼                                             ▼
┌─────────────┐                              ┌─────────────────┐
│ SQLite/     │                              │ Redis           │
│ PostgreSQL  │                              │ - Queue (Asynq) │
│ - Projects  │                              │ - Cache         │
│ - Users     │                              │                 │
│ - Deploys   │                              │                 │
└─────────────┘                              └─────────────────┘
       │
       │ Volumes
       ▼
┌─────────────┐     ┌────────────┐     ┌────────────────┐
│ Docker      │     │ Git Repos  │     │ Caddy Configs  │
│ Containers  │     │ /projects  │     │ /etc/caddy/    │
└─────────────┘     └────────────┘     └────────────────┘
```

## Component Details

### 1. Frontend (SvelteKit)

**Responsibilities:**
- User interface for project management
- Deployment triggering and monitoring
- Real-time log viewing
- Domain and environment configuration

**Key Features:**
- Server-side rendering (SSR)
- Progressive enhancement
- Real-time updates via WebSocket
- Responsive design with Tailwind v4

**Routes:**
```
/                           - Dashboard
/login                      - Authentication
/projects                   - Project list
/projects/new               - Create project
/projects/:id               - Project details
/projects/:id/deployments   - Deployment history
/projects/:id/settings      - Project settings
```

### 2. Backend API (Go)

**Package Structure:**

```
backend/
├── cmd/server/             # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/       # HTTP request handlers
│   │   │   ├── auth.go
│   │   │   ├── project.go
│   │   │   ├── deployment.go
│   │   │   └── webhook.go
│   │   ├── middleware/     # HTTP middleware
│   │   │   └── auth.go
│   │   └── routes/         # Route definitions
│   │       └── routes.go
│   ├── models/             # Database models
│   │   ├── user.go
│   │   ├── project.go
│   │   ├── deployment.go
│   │   └── environment.go
│   ├── services/           # Business logic
│   │   ├── deployment/     # Deployment orchestration
│   │   ├── docker/         # Docker operations
│   │   ├── git/            # Git operations
│   │   └── caddy/          # Caddy configuration
│   ├── config/             # Configuration management
│   └── database/           # Database initialization
└── pkg/                    # Public packages
```

### 3. Deployment Flow

```
┌─────────────────────────────────────────────────────────────┐
│ 1. Trigger                                                   │
│    - Manual (UI button)                                     │
│    - Webhook (Git push)                                     │
│    - API call                                               │
└────────────────────┬────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. Queue Job (Asynq)                                        │
│    - Create deployment record (status: pending)             │
│    - Add to Redis queue                                     │
└────────────────────┬────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. Clone Repository (Git Service)                           │
│    - Clone from Git URL                                     │
│    - Checkout specified branch                              │
│    - Get commit info                                        │
└────────────────────┬────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. Detect & Configure                                       │
│    - Auto-detect framework from package.json                │
│    - Generate Dockerfile if not exists                      │
│    - Load environment variables                             │
└────────────────────┬────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Build Application                                        │
│    - npm install                                            │
│    - npm run build                                          │
│    - Build Docker image                                     │
└────────────────────┬────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Deploy Container (Docker Service)                        │
│    - Stop old container (if exists)                         │
│    - Create new container                                   │
│    - Configure port mappings                                │
│    - Start container                                        │
└────────────────────┬────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. Update Proxy (Caddy Service)                            │
│    - Generate Caddy config                                  │
│    - Write to /etc/caddy/sites/                            │
│    - Reload Caddy                                           │
│    - SSL auto-provisioned                                   │
└────────────────────┬────────────────────────────────────────┘
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ 8. Finalize                                                 │
│    - Update deployment status (success/failed)              │
│    - Store build logs                                       │
│    - Send notifications (future)                            │
└─────────────────────────────────────────────────────────────┘
```

### 4. Database Schema

#### Users
```sql
id          INTEGER PRIMARY KEY
email       TEXT UNIQUE NOT NULL
password    TEXT NOT NULL
name        TEXT
role        TEXT (admin/user)
created_at  TIMESTAMP
updated_at  TIMESTAMP
```

#### Projects
```sql
id              INTEGER PRIMARY KEY
user_id         INTEGER REFERENCES users(id)
name            TEXT NOT NULL
git_url         TEXT NOT NULL
git_branch      TEXT DEFAULT 'main'
framework       TEXT (sveltekit/react/vue/etc)
baas_type       TEXT (pocketbase/supabase/etc)
frontend_port   INTEGER
backend_port    INTEGER
auto_deploy     BOOLEAN DEFAULT true
status          TEXT (pending/active/failed)
created_at      TIMESTAMP
updated_at      TIMESTAMP
```

#### Deployments
```sql
id              INTEGER PRIMARY KEY
project_id      INTEGER REFERENCES projects(id)
commit_hash     TEXT
commit_message  TEXT
branch          TEXT
status          TEXT (pending/building/success/failed)
triggered_by    TEXT (manual/webhook/api)
started_at      TIMESTAMP
completed_at    TIMESTAMP
duration        INTEGER (seconds)
error_message   TEXT
created_at      TIMESTAMP
```

#### Environments
```sql
id          INTEGER PRIMARY KEY
project_id  INTEGER REFERENCES projects(id)
key         TEXT NOT NULL
value       TEXT NOT NULL
is_secret   BOOLEAN DEFAULT false
created_at  TIMESTAMP
```

### 5. Security

**Authentication:**
- JWT tokens (24-hour expiry)
- Refresh tokens (7-day expiry)
- Bcrypt password hashing

**Authorization:**
- Role-based access (admin/user)
- Project ownership verification
- API route protection

**Secrets Management:**
- Environment variables encrypted at rest
- Webhook signature verification
- Docker socket access control

### 6. Scalability Considerations

**Current (Single VPS):**
- SQLite for simplicity
- Local file storage
- Single Redis instance

**Future (Multi-VPS):**
- PostgreSQL with replication
- S3-compatible object storage
- Redis cluster
- Load balancer (multiple API instances)

## Development Workflow

1. **Local Development**
   - Frontend: Vite dev server (HMR)
   - Backend: Hot reload with `air`
   - SQLite database

2. **Testing**
   - Backend: Go tests
   - Frontend: Vitest/Playwright

3. **Production Build**
   - Multi-stage Docker builds
   - Optimized images
   - Docker Compose orchestration

## Monitoring & Logging

- **Application Logs**: Docker logs
- **Build Logs**: Stored in database + real-time streaming
- **Caddy Logs**: JSON format in `/var/log/caddy/`
- **Metrics** (Future): Prometheus + Grafana

## Deployment Strategies

### Rolling Updates
- New container starts before old stops
- Zero-downtime deployments
- Automatic rollback on failure

### Blue-Green (Future)
- Two identical environments
- Instant switch via Caddy
- Quick rollback capability

---

This architecture provides a solid foundation for a self-hosted deployment platform with room for growth and optimization.
