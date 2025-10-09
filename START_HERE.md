# 🚀 START HERE - VPS Panel

Welcome to VPS Panel! This guide will get you up and running in **5 minutes**.

## 📋 What You've Built

A **self-hosted deployment platform** (like Vercel) that can:
- Deploy SvelteKit, React, Vue, Angular apps
- Integrate with PocketBase, Supabase, or any BaaS
- Auto-deploy on Git push via webhooks
- Manage multiple projects
- View real-time deployment logs
- Handle custom domains with auto-SSL

## 🎯 Quick Start (Choose Your Path)

### Path A: Just Want to Test? (5 minutes)

```bash
# 1. Set up environment
cp .env .env
cp frontend/.env frontend/.env

# 2. Install dependencies
make install
# Or manually: cd backend && go mod download && cd ../frontend && npm install

# 3. Start both servers
make dev
# Or see manual steps below
```

**Then:** Open http://localhost:5173 and follow **[test-manual.md](./test-manual.md)**

---

### Path B: Want to Understand the Code? (15 minutes)

Read in this order:
1. [README.md](./README.md) - Overview
2. [ARCHITECTURE.md](./ARCHITECTURE.md) - System design
3. [backend/README.md](./backend/README.md) - Backend API
4. [frontend/README.md](./frontend/README.md) - Frontend UI

---

### Path C: Ready to Deploy to Production? (30 minutes)

1. Read [QUICKSTART.md](./QUICKSTART.md)
2. Set up your VPS
3. Configure Caddy
4. Deploy with Docker Compose

---

## 🔧 Manual Setup Steps

### Step 1: Prerequisites

**Required:**
- Go 1.23+ → `go version`
- Node.js 20+ → `node --version`

**Optional (for deployment features):**
- Docker → `docker --version`
- Caddy → `caddy version` (for production)

### Step 2: Clone & Install

```bash
# You're already here if you're reading this!

# Install backend dependencies
cd backend
go mod download
cd ..

# Install frontend dependencies
cd frontend
npm install
cd ..
```

### Step 3: Configure Environment

```bash
# Copy environment files
cp .env .env
cp frontend/.env frontend/.env
```

**Edit `.env` and set:**
```bash
# Generate with: openssl rand -base64 32
JWT_SECRET=your-random-secret-here
WEBHOOK_SECRET=your-webhook-secret-here
```

### Step 4: Start Servers

**Terminal 1 - Backend:**
```bash
cd backend
go run cmd/server/main.go
```

Wait for: `🚀 Server starting on 0.0.0.0:8080`

**Terminal 2 - Frontend:**
```bash
cd frontend
npm run dev
```

Wait for: `Local: http://localhost:5173/`

### Step 5: Access Application

Open **http://localhost:5173** in your browser!

---

## ✅ Verify Everything Works

### Quick Health Check

**Option 1: Use the script**
```bash
# Linux/Mac
./scripts/check-health.sh

# Windows PowerShell
.\scripts\check-health.ps1
```

**Option 2: Manual check**
```bash
# Check backend
curl http://localhost:8080/health
# Should return: {"status":"ok","service":"vps-panel-api"}

# Check frontend
curl http://localhost:5173
# Should return HTML
```

---

## 🧪 Testing

### Option 1: Automated API Tests

**Linux/Mac:**
```bash
chmod +x test-api.sh
./test-api.sh
```

**Windows PowerShell:**
```powershell
.\test-api.ps1
```

This will test:
- ✅ Backend health
- ✅ User registration
- ✅ Login/logout
- ✅ Project creation
- ✅ Project listing
- ✅ Authorization
- ✅ Security

### Option 2: Manual UI Testing

Follow the detailed checklist in **[test-manual.md](./test-manual.md)**

Tests 23 different scenarios including:
- Authentication flows
- Project management
- Deployments
- UI responsiveness
- Security
- Data persistence

---

## 📚 Documentation Map

| Document | Purpose | Time |
|----------|---------|------|
| **START_HERE.md** ← You are here | Quick start guide | 5 min |
| [test-manual.md](./test-manual.md) | Step-by-step testing | 20 min |
| [TESTING.md](./TESTING.md) | Complete testing guide | - |
| [README.md](./README.md) | Project overview | 10 min |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | System design | 15 min |
| [QUICKSTART.md](./QUICKSTART.md) | Production deployment | 30 min |
| [GETTING_STARTED.md](./GETTING_STARTED.md) | Detailed setup guide | - |
| [backend/README.md](./backend/README.md) | Backend API docs | - |
| [frontend/README.md](./frontend/README.md) | Frontend docs | - |

---

## 🎯 Your First Test

### 1. Register an Account

1. Open http://localhost:5173
2. Click "Register here"
3. Fill in your details
4. Click "Create account"

### 2. Create a Project

1. Click "New Project"
2. Fill in:
   - Name: `My Test App`
   - Git URL: `https://github.com/sveltejs/kit`
   - Framework: `SvelteKit`
3. Click "Create Project"

### 3. View Your Project

1. Click on the project card
2. See project details
3. Try clicking "Deploy" (will need Docker to complete)

**🎉 Congratulations! You've successfully tested VPS Panel!**

---

## 🐛 Troubleshooting

### "Backend won't start"

```bash
# Check if port 8080 is in use
netstat -ano | findstr :8080  # Windows
lsof -i :8080                 # Mac/Linux

# Change port in .env if needed
PORT=8081
```

### "Frontend shows 'Failed to fetch'"

Check that:
1. Backend is running: `curl http://localhost:8080/health`
2. `frontend/.env` has: `VITE_API_URL=http://localhost:8080/api/v1`
3. No CORS errors in browser console (F12)

### "Database errors"

```bash
# Reset database
rm backend/data/vps-panel.db

# Restart backend
cd backend
go run cmd/server/main.go
```

### "Login doesn't work"

1. Open browser DevTools (F12)
2. Check Console for errors
3. Check Network tab for failed requests
4. Verify JWT_SECRET is set in `.env`

---

## 🚀 Next Steps After Testing

### 1. Customize the UI

Edit files in `frontend/src/`:
- `app.css` - Tailwind CSS v4 styles
- `lib/components/` - Reusable components
- `routes/` - Pages

### 2. Add Features

Ideas:
- Email notifications
- Slack/Discord webhooks
- Build caching
- Custom build scripts
- Multi-user teams
- Resource monitoring

### 3. Deploy to Production

Follow [QUICKSTART.md](./QUICKSTART.md) to deploy to your VPS with:
- Docker Compose
- Caddy reverse proxy
- Automatic SSL
- Git webhooks

### 4. Set Up Git Webhooks

**GitHub:**
1. Repo → Settings → Webhooks
2. URL: `https://your-domain.com/api/v1/webhooks/github`
3. Secret: Your `WEBHOOK_SECRET` from `.env`
4. Events: Push events

**GitLab:**
1. Project → Settings → Webhooks
2. URL: `https://your-domain.com/api/v1/webhooks/gitlab`
3. Secret Token: Your `WEBHOOK_SECRET`

---

## 📞 Getting Help

### Check These First:
1. [TESTING.md](./TESTING.md) - Troubleshooting section
2. Browser console (F12) for errors
3. Backend terminal for error messages
4. [GETTING_STARTED.md](./GETTING_STARTED.md) - Detailed guide

### Common Issues:

**"Cannot find module"** → Run `npm install` in frontend
**"Database locked"** → Stop backend, delete `.db` file, restart
**"Port in use"** → Change PORT in `.env`
**"Unauthorized"** → Check JWT_SECRET is set

---

## 🎯 Success Checklist

After following this guide, you should be able to:

- [x] Start backend server
- [x] Start frontend server
- [x] Register a new account
- [x] Login successfully
- [x] Create a project
- [x] View projects list
- [x] See deployment interface
- [x] Navigate the UI smoothly

**If you can do all of these → You're ready to use VPS Panel! 🎉**

---

## 🎨 Project Structure Overview

```
vps-panel/
├── backend/              # Go API server
│   ├── cmd/server/       # Entry point
│   ├── internal/         # Business logic
│   └── Dockerfile
│
├── frontend/             # SvelteKit UI
│   ├── src/
│   │   ├── routes/       # Pages
│   │   └── lib/          # Components & API
│   └── Dockerfile
│
├── scripts/              # Helper scripts
├── test-api.sh           # API tests (Bash)
├── test-api.ps1          # API tests (PowerShell)
├── test-manual.md        # Manual test guide
├── docker-compose.yml    # Full stack
└── Makefile              # Quick commands
```

---

## 💡 Helpful Commands

```bash
# Start development
make dev

# Install dependencies
make install

# Build for production
make build

# Run with Docker
make docker-up

# Check health
./scripts/check-health.sh    # Linux/Mac
.\scripts\check-health.ps1   # Windows

# Run API tests
./test-api.sh                # Linux/Mac
.\test-api.ps1               # Windows

# View all commands
make help
```

---

## 🎓 Learning Resources

**Backend (Go):**
- [Fiber Documentation](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [Go by Example](https://gobyexample.com/)

**Frontend (SvelteKit):**
- [SvelteKit Docs](https://kit.svelte.dev/docs)
- [Svelte 5 Tutorial](https://svelte.dev/tutorial)
- [Tailwind CSS v4](https://tailwindcss.com/docs)

---

## ✨ What Makes This Special?

✅ **Modern Stack** - Svelte 5 + Tailwind v4 + Go
✅ **Type-Safe** - TypeScript + Go types
✅ **Production-Ready** - Docker, Caddy, SSL
✅ **Developer-Friendly** - Hot reload, clear errors
✅ **Well-Documented** - Extensive guides
✅ **Fully Functional** - Complete CRUD operations
✅ **Secure** - JWT auth, CORS, XSS protection
✅ **Responsive** - Mobile, tablet, desktop

---

**Ready to build something amazing? Let's go! 🚀**

---

*Built with ❤️ using SvelteKit, Tailwind CSS v4, and Go*
