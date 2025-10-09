# VPS Panel - Self-Hosted Deployment Platform

A Vercel-like deployment platform for self-hosting full-stack applications on your own VPS. Deploy SvelteKit, React, Vue, Angular apps with PocketBase, Supabase, or any other BaaS backend automatically.

## 🚀 Features

- **Multi-Framework Support**: SvelteKit, React, Vue, Angular, Next.js, Nuxt
- **BaaS Integration**: PocketBase, Supabase, Firebase, Appwrite
- **Auto-Deploy**: Git webhook integration (GitHub, GitLab, Bitbucket)
- **Caddy Integration**: Automatic reverse proxy configuration with SSL
- **Docker Orchestration**: Containerized deployments with Docker
- **Real-time Logs**: Live build and deployment logs
- **Environment Management**: Secure environment variable storage
- **Custom Domains**: Multi-domain support with automatic SSL
- **Modern UI**: SvelteKit + Tailwind CSS v4

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│  SvelteKit Frontend (Tailwind v4)       │
│  - Project management                   │
│  - Deployment dashboard                 │
│  - Real-time logs                       │
└─────────────────┬───────────────────────┘
                  │ REST API + WebSocket
┌─────────────────▼───────────────────────┐
│  Go Backend API (Fiber)                 │
│  - Authentication & Authorization       │
│  - Project CRUD operations              │
│  - Deployment orchestration             │
│  - Git webhook handling                 │
└─────────────────┬───────────────────────┘
                  │
        ┌─────────┴─────────┬─────────────┐
        ▼                   ▼             ▼
┌───────────────┐  ┌────────────┐  ┌─────────────┐
│ Docker Engine │  │   Caddy    │  │   SQLite/   │
│ (containers)  │  │  (proxy)   │  │ PostgreSQL  │
└───────────────┘  └────────────┘  └─────────────┘
```

## 📋 Prerequisites

- **Go 1.23+** (for backend)
- **Node.js 20+** (for frontend)
- **Docker** (for containerization)
- **Caddy** (for reverse proxy)
- **Git** (for repository cloning)

## 🔧 Installation

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/vps-panel.git
cd vps-panel
```

### 2. Install Dependencies

```bash
make install
```

Or manually:

```bash
# Frontend
cd frontend
npm install

# Backend
cd ../backend
go mod download
```

### 3. Configure Environment

```bash
cp .env .env
```

Edit `.env` and configure:
- `JWT_SECRET`: Generate a secure random string
- `WEBHOOK_SECRET`: For Git webhook verification
- `ADMIN_EMAIL` & `ADMIN_PASSWORD`: Initial admin credentials

### 4. Run Development Servers

```bash
make dev
```

This starts:
- Frontend at http://localhost:5173
- Backend API at http://localhost:8080

Or run separately:

```bash
# Terminal 1 - Backend
make dev-backend

# Terminal 2 - Frontend
make dev-frontend
```

## 🐳 Docker Deployment

### Production Deployment

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Services

- **Backend API**: http://localhost:8080
- **Frontend**: http://localhost:3000
- **Redis**: localhost:6379

## 📖 Usage

### 1. Create an Account

Navigate to http://localhost:3000 and register an account.

### 2. Create a Project

- Click "New Project"
- Enter project details:
  - Name
  - Git repository URL
  - Framework (SvelteKit, React, Vue, etc.)
  - BaaS type (PocketBase, Supabase, etc.)
  - Build configuration

### 3. Deploy

- Manual: Click "Deploy" button
- Auto: Configure Git webhooks for automatic deployments

### 4. Configure Domain

- Add custom domain(s)
- Caddy automatically handles SSL via Let's Encrypt

## 🔐 API Documentation

### Authentication

```bash
# Register
POST /api/v1/auth/register
{
  "email": "user@example.com",
  "password": "securepassword",
  "name": "John Doe"
}

# Login
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "securepassword"
}
```

### Projects

```bash
# List projects
GET /api/v1/projects
Authorization: Bearer <token>

# Create project
POST /api/v1/projects
Authorization: Bearer <token>
{
  "name": "My App",
  "git_url": "https://github.com/user/repo.git",
  "framework": "sveltekit",
  "baas_type": "pocketbase"
}

# Deploy project
POST /api/v1/projects/:id/deployments
Authorization: Bearer <token>
```

See [API.md](./docs/API.md) for complete API documentation.

## 🎯 Supported Frameworks

### Frontend Frameworks
- ✅ SvelteKit
- ✅ React (Vite/CRA)
- ✅ Vue 3 (Vite)
- ✅ Angular
- ✅ Next.js
- ✅ Nuxt

### BaaS Backends
- ✅ PocketBase
- ✅ Supabase
- ✅ Firebase
- ✅ Appwrite

## 🛠️ Development

```bash
# Run tests
make test-backend

# Format code
make format-backend
make format-frontend

# Lint
make lint-frontend

# Clean build artifacts
make clean
```

## 📁 Project Structure

```
vps-panel/
├── backend/               # Go backend
│   ├── cmd/server/        # Entry point
│   ├── internal/
│   │   ├── api/           # HTTP handlers & routes
│   │   ├── models/        # Database models
│   │   ├── services/      # Business logic
│   │   ├── config/        # Configuration
│   │   └── database/      # Database setup
│   └── pkg/               # Public packages
│
├── frontend/              # SvelteKit frontend
│   ├── src/
│   │   ├── routes/        # Pages & API routes
│   │   ├── lib/           # Components & utilities
│   │   └── app.css        # Tailwind v4 styles
│   └── static/
│
├── docker-compose.yml     # Docker orchestration
├── Makefile              # Development commands
└── README.md
```

## 🤝 Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md)

## 📄 License

MIT License - see [LICENSE](./LICENSE)

## 🙏 Acknowledgments

Built with:
- [SvelteKit](https://kit.svelte.dev)
- [Tailwind CSS v4](https://tailwindcss.com)
- [Go Fiber](https://gofiber.io)
- [Docker](https://docker.com)
- [Caddy](https://caddyserver.com)
- [GORM](https://gorm.io)

## 📞 Support

- Documentation: [docs/](./docs/)
- Issues: [GitHub Issues](https://github.com/yourusername/vps-panel/issues)

---

Built with ❤️ for self-hosting enthusiasts
