# GitHub OAuth - Vercel-Style Integration

## ✅ What's Been Built

Your VPS Panel now has **seamless GitHub integration** just like Vercel, Netlify, and Railway!

### For Your Users:
1. Click "Connect GitHub" in Settings
2. Authorize on GitHub
3. See ALL their repositories on New Project page
4. Click any repo → Auto-fills everything
5. Deploy with one click

**No setup. No tokens. No configuration. It just works.**

---

## 🚀 Quick Start (For You - The App Owner)

### Step 1: Install Package
```bash
cd backend
go get golang.org/x/oauth2
go get golang.org/x/oauth2/github
```

### Step 2: Create GitHub OAuth App
1. Go to: https://github.com/settings/developers
2. Click "New OAuth App"
3. Fill in:
   - **Name**: `VPS Panel`
   - **Homepage**: `http://localhost:5173`
   - **Callback**: `http://localhost:8080/api/v1/auth/oauth/callback/github`
4. Save **Client ID** and **Client Secret**

### Step 3: Add to `.env`
```bash
GITHUB_CLIENT_ID=your_client_id_here
GITHUB_CLIENT_SECRET=your_client_secret_here
OAUTH_CALLBACK_URL=http://localhost:8080/api/v1/auth/oauth/callback
```

### Step 4: Build & Run
```bash
go build -o vps-panel.exe ./cmd/server
./vps-panel.exe
```

**Done! Your users can now connect GitHub seamlessly.**

---

## 📱 What Your Users See

### Settings Page
```
Connected Accounts
┌─────────────────────────────────────────┐
│ 🔷 GitHub                               │
│ Not connected                           │
│                      [Connect GitHub]   │
└─────────────────────────────────────────┘
```

After clicking "Connect GitHub":
- Redirects to GitHub
- User clicks "Authorize"
- Redirects back
- ✅ Connected!

```
Connected Accounts
┌─────────────────────────────────────────┐
│ 🔷 GitHub                               │
│ ✅ Connected as @username               │
│                      [Disconnect]       │
└─────────────────────────────────────────┘
```

### New Project Page
```
Import Git Repository
┌─────────────────────────────────────────┐
│ Search repositories...                  │
├─────────────────────────────────────────┤
│ 🔷 my-sveltekit-app    🔒 Private    →  │
│ username/my-sveltekit-app               │
├─────────────────────────────────────────┤
│ 🔷 react-dashboard                   →  │
│ username/react-dashboard                │
├─────────────────────────────────────────┤
│ 🔷 nextjs-blog         🔒 Private    →  │
│ username/nextjs-blog                    │
└─────────────────────────────────────────┘
```

Click any repo → Form auto-fills → Click "Create Project" → Deployed!

---

## 🎯 Features

✅ **Auto-load repositories** - Shows ALL user repos (public + private)
✅ **Search & filter** - Find repos instantly
✅ **One-click selection** - Click to auto-fill form
✅ **Framework detection** - Automatically detects SvelteKit, React, Next.js, etc.
✅ **Private repo support** - Works seamlessly with OAuth token
✅ **Secure tokens** - Stored server-side, never exposed to frontend
✅ **No manual setup** - Users never see OAuth configuration

---

## 🔒 Security

- **OAuth tokens** stored in database with `json:"-"` tag (never sent to frontend)
- **State parameter** for CSRF protection (includes user ID + random token)
- **HTTPOnly cookies** in production
- **SameSite=Lax** for additional CSRF protection
- **Tokens only used server-side** for Git operations

---

## 🛠️ How It Works

### 1. User Clicks "Connect GitHub"
```
Frontend → GET /api/v1/auth/oauth/github/init
           ↓
Backend → Generates state (userID:randomToken)
        → Stores in cookie
        → Returns GitHub OAuth URL
           ↓
Frontend → Redirects user to GitHub
```

### 2. User Authorizes on GitHub
```
GitHub → User clicks "Authorize"
      → Redirects to callback URL with code
```

### 3. Backend Processes Callback
```
GET /api/v1/auth/oauth/callback/github?code=xxx&state=yyy
    ↓
Backend → Verifies state matches cookie
        → Extracts user ID from state
        → Exchanges code for access token
        → Fetches GitHub username
        → Saves to database
        → Redirects to /settings?github=connected
```

### 4. User Goes to New Project
```
Frontend → GET /api/v1/auth/oauth/github/repositories
           ↓
Backend → Fetches repos using stored token
        → Returns list of repositories
           ↓
Frontend → Displays all repos
         → User clicks one
         → Form auto-fills
```

---

## 📂 Files Changed/Created

### Backend
- ✅ `internal/config/config.go` - Added OAuth config
- ✅ `internal/models/user.go` - Added GitHub fields
- ✅ `internal/services/oauth/github.go` - GitHub OAuth service
- ✅ `internal/api/handlers/auth.go` - OAuth handlers
- ✅ `internal/api/routes/routes.go` - OAuth routes
- ✅ `migrations/004_add_oauth_fields.sql` - Database migration

### Frontend
- ✅ `routes/(app)/settings/+page.svelte` - Settings page
- ✅ `routes/(app)/projects/new/+page.svelte` - Repo selector
- ✅ `lib/api/oauth.ts` - OAuth API client
- ✅ `lib/types.ts` - OAuth types
- ✅ `lib/components/Navbar.svelte` - Added Settings link

---

## 🧪 Testing

1. **Start backend**: `./vps-panel.exe`
2. **Start frontend**: `npm run dev`
3. **Register/Login** at http://localhost:5173
4. **Go to Settings** → Click "Connect GitHub"
5. **Authorize** on GitHub
6. **Go to New Project** → See all your repos!
7. **Click a repo** → Everything auto-filled
8. **Create Project** → Deployed!

---

## 🚀 Production Deployment

### Update GitHub OAuth App
1. Go to https://github.com/settings/developers
2. Edit your OAuth App
3. Update URLs:
   - Homepage: `https://yourdomain.com`
   - Callback: `https://api.yourdomain.com/api/v1/auth/oauth/callback/github`

### Update `.env`
```bash
GITHUB_CLIENT_ID=same_as_before
GITHUB_CLIENT_SECRET=same_as_before
OAUTH_CALLBACK_URL=https://api.yourdomain.com/api/v1/auth/oauth/callback
CORS_ORIGINS=https://yourdomain.com
```

### Restart
```bash
./vps-panel
```

---

## 💡 Key Differences from Manual Approach

| Feature | OAuth (Vercel-style) | Manual Tokens |
|---------|---------------------|---------------|
| User setup | ✅ One click | ❌ Copy/paste tokens |
| List repos | ✅ Auto-loaded | ❌ Manual search |
| Private repos | ✅ Works seamlessly | ⚠️ Need credentials |
| Security | ✅ Server-side only | ⚠️ User handles tokens |
| UX | ✅ Like Vercel | ❌ Manual workflow |

---

## 📝 Summary

**You (App Owner):**
- Create GitHub OAuth app once: 5 minutes ✅
- Add credentials to `.env`: 30 seconds ✅
- Build & deploy: Done ✅

**Your Users:**
- Click "Connect GitHub": 3 seconds ✅
- Select repo: 2 seconds ✅
- Deploy: 1 click ✅

**Total user effort: 6 seconds + 1 click. Just like Vercel.** 🎉

---

For detailed setup instructions, see: `SETUP_OAUTH_ONCE.md`
