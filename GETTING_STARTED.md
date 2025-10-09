# Getting Started with VPS Panel

Complete guide to run your VPS Panel deployment platform locally.

## 📋 Prerequisites

Ensure you have these installed:
- **Go 1.23+** - `go version`
- **Node.js 20+** - `node --version`
- **Docker** - `docker --version` (optional for deployment features)
- **Git** - `git --version`

## 🚀 Quick Start (5 minutes)

### 1. Install Dependencies

```bash
# Install both frontend and backend dependencies
make install
```

Or manually:

```bash
# Backend
cd backend
go mod download

# Frontend
cd ../frontend
npm install
```

### 2. Configure Environment

```bash
# Copy environment files
cp .env .env
cp frontend/.env frontend/.env
```

**Important:** Edit `.env` and set:
```env
JWT_SECRET=your-random-secret-here
WEBHOOK_SECRET=your-webhook-secret-here
```

Generate secure secrets:
```bash
# On Linux/Mac
openssl rand -base64 32

# On Windows (PowerShell)
[Convert]::ToBase64String((1..32 | ForEach-Object { Get-Random -Maximum 256 }))
```

### 3. Start Development Servers

```bash
# Start both frontend and backend
make dev
```

This will start:
- ✅ **Backend API** → http://localhost:8080
- ✅ **Frontend UI** → http://localhost:5173

Or run separately:

```bash
# Terminal 1 - Backend
cd backend
go run cmd/server/main.go

# Terminal 2 - Frontend
cd frontend
npm run dev
```

### 4. Access the Application

1. Open http://localhost:5173 in your browser
2. Click "Register here" to create an account
3. Fill in your details and create your first project!

## 📱 First Login

### Create an Account

1. Navigate to http://localhost:5173
2. Click **"Register here"**
3. Fill in:
   - **Name**: Your name
   - **Email**: your@email.com
   - **Password**: At least 8 characters
4. Click **"Create account"**

You'll be automatically logged in and redirected to the dashboard!

### Explore the Dashboard

The dashboard shows:
- **Total Projects** - Number of projects you've created
- **Active Projects** - Projects currently deployed
- **Total Deployments** - All deployment attempts
- **Success Rate** - Deployment success percentage

## 🎯 Create Your First Project

### Step 1: Click "New Project"

### Step 2: Fill in Project Details

**Basic Information:**
```
Name: My Awesome App
Description: A cool SvelteKit application
```

**Git Repository:**
```
Repository URL: https://github.com/yourusername/your-repo.git
Branch: main
```

**Framework & Backend:**
```
Framework: SvelteKit
Backend/BaaS: PocketBase (or None)
```

**Build Configuration:**
```
Install Command: npm install
Build Command: npm run build
Output Directory: build
Node Version: Node.js 20
```

**Port Configuration:**
```
Frontend Port: 3000
Backend Port: 8090 (if using BaaS)
```

**Deployment Settings:**
- ✅ Auto Deploy (automatically deploy on git push)

### Step 3: Click "Create Project"

### Step 4: Deploy Your Project

1. Click the **"Deploy"** button
2. Watch the deployment logs in real-time
3. Your app will be live!

## 🎨 Frontend Features

### Pages

| Route | Description |
|-------|-------------|
| `/` | Home (redirects to dashboard or login) |
| `/login` | Sign in to your account |
| `/register` | Create a new account |
| `/projects` | View all your projects (dashboard) |
| `/projects/new` | Create a new project |
| `/projects/:id` | Project details & deployments |
| `/projects/:id/deployments/:deploymentId` | Deployment logs |

### UI Components

All built with **Svelte 5 runes** and **Tailwind CSS v4**:
- Button (with loading states)
- Card
- Input (with validation)
- Select dropdown
- Badge (status indicators)
- Modal dialogs
- Alert messages
- Navbar (responsive)

## 🔧 Backend API Endpoints

### Authentication

```bash
# Register
POST http://localhost:8080/api/v1/auth/register
{
  "email": "user@example.com",
  "password": "password123",
  "name": "John Doe"
}

# Login
POST http://localhost:8080/api/v1/auth/login
{
  "email": "user@example.com",
  "password": "password123"
}
```

### Projects

```bash
# Get all projects
GET http://localhost:8080/api/v1/projects
Authorization: Bearer <token>

# Create project
POST http://localhost:8080/api/v1/projects
Authorization: Bearer <token>
{
  "name": "My App",
  "git_url": "https://github.com/user/repo.git",
  "framework": "sveltekit"
}

# Deploy project
POST http://localhost:8080/api/v1/projects/1/deployments
Authorization: Bearer <token>
```

## 🐛 Troubleshooting

### Backend won't start

**Error: Port 8080 already in use**
```bash
# Find process using port 8080
# Linux/Mac
lsof -i :8080

# Windows
netstat -ano | findstr :8080

# Kill the process or change the port in .env
PORT=8081
```

**Error: Database locked**
```bash
# Remove database and restart
rm backend/data/vps-panel.db
go run backend/cmd/server/main.go
```

### Frontend won't start

**Error: Port 5173 already in use**
```bash
# Kill process or change port
# Edit vite.config.ts and add:
server: {
  port: 5174
}
```

**Error: Module not found**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

### Can't login

**Check backend is running:**
```bash
curl http://localhost:8080/health
# Should return: {"status":"ok","service":"vps-panel-api"}
```

**Check credentials:**
- Email must be valid format
- Password must be at least 8 characters

**Clear browser data:**
- Open DevTools (F12)
- Application → Local Storage → Clear All
- Refresh page

### TypeScript errors in frontend

```bash
cd frontend
npm run check
```

Fix any type errors shown.

## 📚 Next Steps

### Set Up Git Webhooks

Configure your Git provider to automatically deploy on push:

**GitHub:**
1. Go to repo → Settings → Webhooks
2. Add webhook: `http://yourserver.com/api/v1/webhooks/github`
3. Content type: `application/json`
4. Secret: Your `WEBHOOK_SECRET` from `.env`
5. Events: Just the push event

**GitLab:**
1. Go to project → Settings → Webhooks
2. URL: `http://yourserver.com/api/v1/webhooks/gitlab`
3. Secret Token: Your `WEBHOOK_SECRET`
4. Trigger: Push events

### Deploy to Production

See [QUICKSTART.md](./QUICKSTART.md) for production deployment with Docker.

### Configure Caddy

For production deployments with custom domains, configure Caddy:

```bash
# Install Caddy
# Linux
sudo apt install caddy

# Create sites directory
sudo mkdir -p /etc/caddy/sites

# Edit Caddyfile
sudo nano /etc/caddy/Caddyfile
```

Add:
```caddy
import /etc/caddy/sites/*
```

VPS Panel will automatically generate Caddy configs for each project!

## 🎓 Learn More

- [Architecture Documentation](./ARCHITECTURE.md) - System design
- [Backend Documentation](./backend/README.md) - API details
- [Frontend Documentation](./frontend/README.md) - UI components
- [Quick Start Guide](./QUICKSTART.md) - Production deployment

## ❓ Getting Help

### Common Issues

1. **"Failed to connect to API"**
   - Check backend is running on port 8080
   - Check VITE_API_URL in frontend/.env

2. **"Token expired"**
   - Tokens expire after 24 hours
   - Just log in again

3. **"Project not found"**
   - Make sure project belongs to your user
   - Check URL parameters

### Debug Mode

**Backend:**
```bash
# Set log level
ENV=development go run cmd/server/main.go
```

**Frontend:**
```bash
# Open browser DevTools (F12)
# Check Console and Network tabs
```

## 🎉 You're Ready!

Start building and deploying your applications with VPS Panel!

**Quick Commands:**
```bash
make dev          # Start dev servers
make build        # Build for production
make docker-up    # Start with Docker
make help         # See all commands
```

Happy deploying! 🚀
