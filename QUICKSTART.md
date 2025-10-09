# Quick Start Guide

Get VPS Panel up and running in 5 minutes!

## Prerequisites Check

```bash
# Check Go installation
go version  # Should be 1.23+

# Check Node.js installation
node --version  # Should be 20+

# Check Docker installation
docker --version

# Check if Caddy is installed (for production)
caddy version
```

## Step 1: Clone & Install

```bash
# Clone repository
git clone https://github.com/T-SOFT-TECH/vps-panel.git
cd vps-panel

# Install all dependencies (frontend + backend)
make install
```

## Step 2: Configure Environment

```bash
# Copy environment template
cp .env .env

# Edit with your favorite editor
nano .env  # or vim, code, etc.
```

**Minimum required configuration:**

```env
JWT_SECRET=generate-a-random-secure-string-here
WEBHOOK_SECRET=another-random-string-for-webhooks
```

**Optional but recommended:**

```env
ADMIN_EMAIL=your-email@example.com
ADMIN_PASSWORD=your-secure-password
```

## Step 3: Run Development Servers

```bash
# Start both frontend and backend
make dev
```

This will start:
- ✅ Frontend at http://localhost:5173
- ✅ Backend API at http://localhost:8080

Or run them separately in different terminals:

```bash
# Terminal 1 - Backend
make dev-backend

# Terminal 2 - Frontend
make dev-frontend
```

## Step 4: Access the Application

1. Open your browser to **http://localhost:5173**
2. Create an account
3. Create your first project!

## Creating Your First Project

### 1. Click "New Project"

### 2. Fill in Project Details

```
Name: My Awesome App
Git URL: https://github.com/yourusername/your-app.git
Branch: main
Framework: SvelteKit (or your framework)
BaaS Type: PocketBase (if applicable)
```

### 3. Configure Build Settings

```
Install Command: npm install
Build Command: npm run build
Output Directory: build (or dist, .next, etc.)
Node Version: 20
Frontend Port: 3000
Backend Port: 8090 (if using BaaS)
```

### 4. Deploy!

Click the "Deploy" button and watch your app build in real-time!

## Production Deployment (Docker)

### On Your VPS

```bash
# 1. Clone repository
git clone https://github.com/T-SOFT-TECH/vps-panel.git
cd vps-panel

# 2. Configure production environment
cp .env .env
nano .env  # Set production values

# 3. Build and start services
docker-compose up -d

# 4. View logs
docker-compose logs -f
```

### Configure Caddy (Reverse Proxy)

Create `/etc/caddy/Caddyfile`:

```caddy
# Import all site configs
import /etc/caddy/sites/*

# VPS Panel UI
panel.yourdomain.com {
    reverse_proxy localhost:3000
}

# VPS Panel API
api.yourdomain.com {
    reverse_proxy localhost:8080
}
```

Reload Caddy:

```bash
sudo systemctl reload caddy
```

## Setting Up Git Webhooks

### GitHub

1. Go to your repository → Settings → Webhooks
2. Click "Add webhook"
3. **Payload URL**: `https://api.yourdomain.com/api/v1/webhooks/github`
4. **Content type**: `application/json`
5. **Secret**: Your `WEBHOOK_SECRET` from `.env`
6. **Events**: Just the push event
7. Click "Add webhook"

### GitLab

1. Go to your project → Settings → Webhooks
2. **URL**: `https://api.yourdomain.com/api/v1/webhooks/gitlab`
3. **Secret Token**: Your `WEBHOOK_SECRET`
4. **Trigger**: Push events
5. Click "Add webhook"

## Troubleshooting

### Port Already in Use

```bash
# Find what's using the port
lsof -i :8080  # or :5173

# Kill the process
kill -9 <PID>
```

### Database Issues

```bash
# Reset database (WARNING: Deletes all data!)
rm backend/data/vps-panel.db
make dev-backend  # Will recreate database
```

### Docker Issues

```bash
# Check Docker is running
docker ps

# Restart Docker service
sudo systemctl restart docker

# Check VPS Panel containers
docker-compose ps
```

### Build Failures

```bash
# Clear node_modules and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install

# For backend
cd ../backend
rm go.sum
go mod download
go mod tidy
```

## Next Steps

- 📖 Read the [Architecture Documentation](./ARCHITECTURE.md)
- 🔐 Set up [SSL certificates](./docs/SSL.md)
- 🚀 Configure [auto-deployment](./docs/AUTO_DEPLOY.md)
- 📊 Enable [monitoring](./docs/MONITORING.md)

## Common Commands

```bash
# Development
make dev              # Start dev servers
make install          # Install dependencies
make clean            # Clean build artifacts

# Production
make build            # Build for production
make docker-build     # Build Docker images
make docker-up        # Start Docker containers
make docker-down      # Stop Docker containers

# Testing
make test-backend     # Run backend tests

# Code Quality
make format-backend   # Format Go code
make format-frontend  # Format Svelte code
make lint-frontend    # Lint frontend code
```

## Getting Help

- 📚 [Full Documentation](./README.md)
- 🐛 [Report Issues](https://github.com/T-SOFT-TECH/vps-panel/issues)
- 💬 [Discussions](https://github.com/T-SOFT-TECH/vps-panel/discussions)

---

Happy deploying! 🚀
