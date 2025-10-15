# VPS Panel - Self-Hosted Deployment Platform

A modern, self-hosted deployment platform for managing and deploying full-stack applications on your own VPS. Deploy SvelteKit, React, Vue, Angular, and Next.js apps with integrated BaaS backends like PocketBase.

![TSOFT Technologies](https://img.shields.io/badge/TSOFT-Technologies-0A6522?style=flat-square)
![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go)
![Node](https://img.shields.io/badge/Node-20+-339933?style=flat-square&logo=node.js)
![Docker](https://img.shields.io/badge/Docker-Required-2496ED?style=flat-square&logo=docker)

## ✨ Features

### 🚀 Deployment & Management
- **Multi-Framework Support**: SvelteKit, React, Vue, Angular, Next.js, Nuxt
- **Auto Framework Detection**: Automatically detects project framework and configuration
- **Monorepo Support**: Deploy specific directories from monorepo projects
- **Real-time Build Logs**: Live streaming of build and deployment progress
- **Git Integration**: OAuth integration with GitHub, GitLab, and Gitea

### 🔐 Authentication & Security
- **OAuth Git Providers**: Connect GitHub, GitLab, or Gitea accounts
- **Single User Lock**: First registered user becomes admin, registration auto-disables
- **JWT Authentication**: Secure token-based authentication
- **Private Repository Support**: Deploy from private Git repositories

### 🌐 Domain & SSL
- **Custom Domains**: Add multiple custom domains per project
- **Automatic SSL**: Let's Encrypt SSL certificates via Caddy
- **Multi-Domain Support**: Host multiple domains for a single project
- **Domain Management**: Edit, activate/deactivate domains on the fly

### 🛠️ Backend as a Service
- **PocketBase Integration**: Full PocketBase deployment with admin dashboard
- **Supabase Support**: Connect to Supabase projects
- **Firebase Support**: Firebase backend integration
- **Appwrite Support**: Appwrite backend integration

### 🎨 Modern UI/UX
- **Dark/Light Mode**: System-aware theme switching
- **Responsive Design**: Mobile-friendly interface
- **Real-time Updates**: WebSocket-based live updates
- **TSOFT Brand Colors**: Custom-branded interface

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│  SvelteKit Frontend (Port 3000)             │
│  - Project management dashboard             │
│  - OAuth Git provider integration           │
│  - Domain management                        │
│  - Real-time deployment logs                │
│  - Environment variable management          │
└─────────────────┬───────────────────────────┘
                  │ REST API (http://localhost:3456)
┌─────────────────▼───────────────────────────┐
│  Go Backend API (Fiber) - Port 3456         │
│  - JWT authentication                       │
│  - OAuth callbacks (GitHub/GitLab/Gitea)    │
│  - Project & deployment orchestration       │
│  - Git repository management                │
│  - Docker container lifecycle               │
│  - Caddy configuration management           │
└─────────────────┬───────────────────────────┘
                  │
        ┌─────────┴─────────┬──────────────────┐
        ▼                   ▼                  ▼
┌───────────────┐  ┌────────────────┐  ┌──────────────┐
│ Docker Engine │  │ Caddy Proxy    │  │   SQLite     │
│ (containers)  │  │ (SSL/domains)  │  │  (database)  │
└───────────────┘  └────────────────┘  └──────────────┘
```

## 📋 Prerequisites

- **Ubuntu 20.04+** or **Debian 11+** (for production)
- **Go 1.23+** (for development)
- **Node.js 20+** (for development)
- **Docker & Docker Compose** (required for deployments)
- **Caddy** (installed automatically by install script)
- **Git** (for repository cloning)
- **Root/sudo access** (for production installation)

## 🚀 Quick Start

### Production Installation (Ubuntu/Debian VPS)

The automated installation script will:
- Install all required dependencies (Docker, Caddy, Go, Node.js)
- Set up system users and permissions
- Configure systemd services
- Generate SSL certificates
- Start the VPS Panel

```bash
# Clone the repository
git clone [https://github.com/yourusername/vps-panel](https://github.com/T-SOFT-TECH/vps-panel).git
cd vps-panel

sudo chmod +x install.sh

# Run the installation script
sudo ./install.sh
```

During installation, you'll be prompted for:
- **Domain name**: Your VPS Panel domain (e.g., panel.example.com)
- **Email**: For Let's Encrypt SSL certificates

After installation:
1. Navigate to `https://your-domain.com`
2. Register the first admin account (registration auto-locks after first user)
3. Start deploying projects!

### Development Setup

```bash
# Clone the repository
git clone https://github.com/yourusername/vps-panel.git
cd vps-panel

# Backend setup
cd backend
go mod download
go run cmd/server/main.go

# Frontend setup (new terminal)
cd frontend
npm install
npm run dev
```

Development servers:
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:3456

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the backend directory:

```env
# Server
PORT=3456
GIN_MODE=release

# Database
DATABASE_PATH=/var/lib/vps-panel/database/vps-panel.db

# Authentication
JWT_SECRET=your-secure-random-string-here

# Projects
PROJECTS_DIR=/var/lib/vps-panel/projects

# OAuth
OAUTH_CALLBACK_URL=https://panel.example.com/api/v1/auth/oauth/callback

# Caddy
CADDY_CONFIG_PATH=/var/lib/vps-panel/caddy
PANEL_DOMAIN=panel.example.com
```

### OAuth Setup

To enable Git provider integration:

#### GitHub OAuth
1. Go to GitHub Settings → Developer settings → OAuth Apps
2. Create a new OAuth App:
   - **Application name**: VPS Panel
   - **Homepage URL**: `https://panel.example.com`
   - **Authorization callback URL**: `https://panel.example.com/api/v1/auth/oauth/callback`
3. Copy Client ID and Client Secret
4. Add provider in VPS Panel Settings → Git Providers

#### GitLab OAuth
1. Go to GitLab Settings → Applications
2. Create a new application:
   - **Name**: VPS Panel
   - **Redirect URI**: `https://panel.example.com/api/v1/auth/oauth/callback`
   - **Scopes**: `api`, `read_repository`
3. Copy Application ID and Secret
4. Add provider in VPS Panel Settings → Git Providers

#### Gitea OAuth
1. Go to Gitea Settings → Applications → OAuth2 Applications
2. Create a new OAuth2 Application:
   - **Application Name**: VPS Panel
   - **Redirect URI**: `https://panel.example.com/api/v1/auth/oauth/callback`
3. Copy Client ID and Client Secret
4. Add provider in VPS Panel Settings → Git Providers

## 📖 Usage Guide

### 1. Connect Git Provider

1. Navigate to **Settings** → **Git Providers**
2. Click **Add Git Provider**
3. Select provider type (GitHub, GitLab, or Gitea)
4. Enter OAuth credentials
5. Click **Connect**

### 2. Create a Project

1. Click **New Project**
2. Select your Git provider
3. Choose a repository
4. Select branch (auto-detects available branches)
5. Framework is auto-detected (or manually select)
6. Configure build settings:
   - Root directory (for monorepos)
   - Build command
   - Output directory
   - Environment variables
7. Click **Create Project**

### 3. Deploy

Projects automatically deploy on creation. For subsequent deployments:
- Click **Deploy** button in project dashboard
- View real-time build logs
- Access your deployed app via the generated domain

### 4. Manage Domains

1. Open project details
2. Scroll to **Domains** section
3. Click **Add Domain**
4. Enter your custom domain
5. Enable SSL/HTTPS (recommended)
6. Update your DNS:
   - Add an `A` record pointing to your VPS IP
7. Domain is automatically configured with SSL

### 5. Environment Variables

1. Open project details
2. Navigate to **Environment Variables** section
3. Add key-value pairs
4. Mark sensitive values as secrets
5. Redeploy project for changes to take effect

## 🎯 Supported Technologies

### Frontend Frameworks
- ✅ **SvelteKit** - Full SSR and static support
- ✅ **React** - Vite or Create React App
- ✅ **Vue 3** - Vite-based projects
- ✅ **Angular** - Angular CLI projects
- ✅ **Next.js** - Pages and App router
- ✅ **Nuxt** - Nuxt 3 projects

### Backend as a Service
- ✅ **PocketBase** - Auto-deployed with admin dashboard
- ✅ **Supabase** - External Supabase projects
- ✅ **Firebase** - Firebase SDK integration
- ✅ **Appwrite** - Self-hosted Appwrite

### Git Providers
- ✅ **GitHub** - OAuth integration
- ✅ **GitLab** - Self-hosted and GitLab.com
- ✅ **Gitea** - Self-hosted Gitea instances

## 📁 Project Structure

```
vps-panel/
├── backend/                    # Go backend
│   ├── cmd/
│   │   └── server/            # Application entry point
│   ├── internal/
│   │   ├── api/
│   │   │   ├── handlers/      # HTTP request handlers
│   │   │   ├── middleware/    # Auth, CORS, logging
│   │   │   └── routes/        # Route definitions
│   │   ├── config/            # Configuration management
│   │   ├── database/          # Database setup
│   │   ├── models/            # Database models (GORM)
│   │   └── services/
│   │       ├── caddy/         # Caddy management
│   │       ├── deployment/    # Deployment orchestration
│   │       ├── detector/      # Framework detection
│   │       ├── docker/        # Docker container management
│   │       ├── git/           # Git operations
│   │       └── oauth/         # OAuth implementations
│   └── go.mod
│
├── frontend/                   # SvelteKit frontend
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api/           # API client functions
│   │   │   ├── components/    # Reusable components
│   │   │   └── stores/        # Svelte stores (auth, theme)
│   │   ├── routes/
│   │   │   ├── (app)/         # Protected routes
│   │   │   ├── login/         # Authentication pages
│   │   │   └── register/
│   │   └── app.css            # Global styles
│   ├── static/
│   │   └── img/               # WebP optimized images
│   └── package.json
│
├── scripts/                    # Utility scripts
├── .gitignore
├── ARCHITECTURE.md            # Detailed architecture docs
├── deploy.sh                  # Quick deployment script
├── docker-compose.yml         # Docker services
├── install.sh                 # Production installation
├── Makefile                   # Development commands
└── README.md                  # This file
```

## 🛠️ Development Commands

```bash
# Backend
make dev-backend          # Run backend in development mode
make build-backend        # Build backend binary
make test-backend         # Run backend tests

# Frontend
make dev-frontend         # Run frontend dev server
make build-frontend       # Build frontend for production

# Combined
make dev                  # Run both backend and frontend
make build                # Build both for production
make clean                # Clean build artifacts
```

## 🐳 Docker Deployment

For containerized deployment:

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f vps-panel-backend
docker-compose logs -f vps-panel-frontend

# Stop services
docker-compose down

# Rebuild after changes
docker-compose up -d --build
```

## 📊 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user (first user only)
- `POST /api/v1/auth/login` - Login
- `GET /api/v1/auth/me` - Get current user
- `GET /api/v1/auth/registration-status` - Check if registration is enabled

### OAuth
- `GET /api/v1/auth/oauth/:provider/:provider_id` - Initiate OAuth flow
- `GET /api/v1/auth/oauth/callback` - OAuth callback handler

### Git Providers
- `GET /api/v1/git-providers` - List providers
- `POST /api/v1/git-providers` - Add provider
- `GET /api/v1/git-providers/:id` - Get provider
- `PUT /api/v1/git-providers/:id` - Update provider
- `DELETE /api/v1/git-providers/:id` - Delete provider
- `POST /api/v1/git-providers/:id/disconnect` - Disconnect OAuth
- `GET /api/v1/git-providers/:id/repositories` - List repositories

### Projects
- `GET /api/v1/projects` - List projects
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects/:id` - Get project details
- `PUT /api/v1/projects/:id` - Update project
- `DELETE /api/v1/projects/:id` - Delete project
- `POST /api/v1/projects/detect-framework` - Auto-detect framework
- `POST /api/v1/projects/list-branches` - List Git branches
- `POST /api/v1/projects/list-directories` - List monorepo directories

### Deployments
- `POST /api/v1/projects/:id/deployments` - Create deployment
- `GET /api/v1/projects/:id/deployments` - List deployments
- `GET /api/v1/projects/:id/deployments/:deploymentId` - Get deployment
- `GET /api/v1/projects/:id/deployments/:deploymentId/logs` - Get build logs

### Domains
- `GET /api/v1/projects/:id/domains` - List domains
- `POST /api/v1/projects/:id/domains` - Add domain
- `PUT /api/v1/projects/:id/domains/:domainId` - Update domain
- `DELETE /api/v1/projects/:id/domains/:domainId` - Delete domain

### Environment Variables
- `GET /api/v1/projects/:id/environments` - List env vars
- `POST /api/v1/projects/:id/environments` - Add env var
- `PUT /api/v1/projects/:id/environments/:envId` - Update env var
- `DELETE /api/v1/projects/:id/environments/:envId` - Delete env var

## 🤝 Contributing

We welcome contributions! Please see [ARCHITECTURE.md](./ARCHITECTURE.md) for detailed technical documentation.

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## 🙏 Built With

- [**SvelteKit**](https://kit.svelte.dev) - Frontend framework
- [**Svelte 5**](https://svelte.dev) - Reactive UI framework
- [**Tailwind CSS**](https://tailwindcss.com) - Utility-first CSS
- [**Go**](https://golang.org) - Backend language
- [**Fiber**](https://gofiber.io) - Go web framework
- [**GORM**](https://gorm.io) - Go ORM
- [**SQLite**](https://sqlite.org) - Database
- [**Docker**](https://docker.com) - Containerization
- [**Caddy**](https://caddyserver.com) - Reverse proxy & SSL
- [**PocketBase**](https://pocketbase.io) - BaaS integration

## 📞 Support & Resources

- **Documentation**: See [ARCHITECTURE.md](./ARCHITECTURE.md) for technical details
- **Issues**: Report bugs via [GitHub Issues](https://github.com/T-SOFT-TECH/vps-panel/issues)
- **Installation Help**: Check `install.sh` comments for troubleshooting

## 🔒 Security

- First registered user becomes admin
- Registration automatically locks after first user
- JWT-based authentication
- OAuth tokens securely stored
- Environment variables encrypted in database
- Let's Encrypt SSL/TLS certificates
- Isolated Docker containers per project

## 🎨 Brand

VPS Panel by **TSOFT Technologies**
- Primary: `#0A6522` (Forest Green)
- Secondary: `#083D16` (Dark Green)

---

**Built with ❤️ by TSOFT Technologies for the self-hosting community**

*Deploy anywhere. Host everywhere.*
