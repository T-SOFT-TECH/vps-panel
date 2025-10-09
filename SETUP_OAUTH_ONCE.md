# One-Time OAuth Setup (For App Owner Only)

## Important: You Only Do This ONCE

Just like Vercel, Railway, or Netlify - **YOU** (the app owner) set this up once, and then **ALL YOUR USERS** can just click "Connect GitHub" and it works seamlessly.

Your users will NEVER see any of this setup. They just click "Connect GitHub" → Authorize → Done.

---

## Step 1: Install OAuth Package (One Command)

```bash
cd backend
go get golang.org/x/oauth2
go get golang.org/x/oauth2/github
```

**That's it for packages.**

---

## Step 2: Create GitHub OAuth App (5 Minutes)

### 2.1 Go to GitHub
- Visit: https://github.com/settings/developers
- Click: **"New OAuth App"**

### 2.2 Fill in the form:

**Application name:**
```
VPS Panel
```
(or whatever you want to call your app)

**Homepage URL:**
```
http://localhost:5173
```
(In production: `https://yourdomain.com`)

**Application description:** (optional)
```
Self-hosted deployment platform
```

**Authorization callback URL:** ⚠️ **IMPORTANT**
```
http://localhost:8080/api/v1/auth/oauth/callback/github
```
(In production: `https://api.yourdomain.com/api/v1/auth/oauth/callback/github`)

### 2.3 Click "Register application"

### 2.4 Copy your credentials:
- **Client ID**: Copy this (looks like: `Iv1.a1b2c3d4e5f6g7h8`)
- Click **"Generate a new client secret"**
- **Client Secret**: Copy this (looks like: `1a2b3c4d5e6f...`)

**⚠️ Save these somewhere safe!** You'll need them in the next step.

---

## Step 3: Configure Backend (30 Seconds)

Create `.env` file in `backend/` directory:

```bash
# Copy your GitHub OAuth credentials here
GITHUB_CLIENT_ID=Iv1.a1b2c3d4e5f6g7h8
GITHUB_CLIENT_SECRET=1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p7q8r9s0

# These should already be set
OAUTH_CALLBACK_URL=http://localhost:8080/api/v1/auth/oauth/callback
JWT_SECRET=your-secret-key-here
PORT=8080
CORS_ORIGINS=http://localhost:5173,http://localhost:4173
```

**Just replace:**
- `GITHUB_CLIENT_ID` with your Client ID
- `GITHUB_CLIENT_SECRET` with your Client Secret

---

## Step 4: Build & Run (One Command)

```bash
cd backend
go build -o vps-panel.exe ./cmd/server
./vps-panel.exe
```

---

## ✅ DONE! That's It!

### What Your Users See:

1. **Login** → Create account (normal flow)
2. **Go to Settings** → See "Connected Accounts"
3. **Click "Connect GitHub"** → Redirected to GitHub
4. **Click "Authorize"** on GitHub → Redirected back
5. **✅ Connected!** → Can now see all their repos

### What They DON'T See:
- ❌ No OAuth app setup
- ❌ No client ID/secret
- ❌ No configuration
- ❌ No tokens to copy
- ❌ No manual credentials

**It just works.** Like Vercel. Like Netlify. Like Railway.

---

## The Complete User Flow

```
User → New Project Page
         ↓
     Shows ALL their GitHub repos
         ↓
     Click a repo
         ↓
     Everything auto-filled
         ↓
     Click "Create Project"
         ↓
     ✅ Deployed!
```

**No setup. No configuration. Just works.**

---

## For Production Deployment

When you deploy to production:

### 1. Update GitHub OAuth App:
- Go back to https://github.com/settings/developers
- Edit your OAuth App
- Change URLs to production:
  - Homepage URL: `https://yourdomain.com`
  - Callback URL: `https://api.yourdomain.com/api/v1/auth/oauth/callback/github`

### 2. Update `.env` in production:
```bash
GITHUB_CLIENT_ID=same_as_before
GITHUB_CLIENT_SECRET=same_as_before
OAUTH_CALLBACK_URL=https://api.yourdomain.com/api/v1/auth/oauth/callback
CORS_ORIGINS=https://yourdomain.com
```

### 3. Restart backend
```bash
./vps-panel
```

**Done!** All your users can now connect GitHub seamlessly.

---

## Troubleshooting

### "Failed to connect GitHub"
- Check that `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` are correct in `.env`
- Check that backend is running on port 8080
- Check that callback URL in GitHub matches: `http://localhost:8080/api/v1/auth/oauth/callback/github`

### "Invalid state parameter"
- This happens if cookies are blocked. Make sure your browser allows cookies.

### "GitHub not connected" when clicking repos
- Make sure user went to Settings → Connect GitHub first
- Check that backend saved the connection (check database)

---

## Security Notes

✅ **Tokens are stored securely:**
- OAuth tokens are stored in database with `json:"-"` tag
- Never sent to frontend
- Only used server-side for Git operations

✅ **State parameter for CSRF protection:**
- Random state generated for each OAuth flow
- Validated on callback to prevent attacks

✅ **Secure cookies:**
- HTTPOnly cookies in production
- SameSite=Lax for CSRF protection

---

## Summary

**You (App Owner):**
- Set up GitHub OAuth app: **5 minutes** ✅
- Add credentials to `.env`: **30 seconds** ✅
- Deploy backend: **Done** ✅

**Your Users:**
- Click "Connect GitHub": **3 seconds** ✅
- Everything else is automatic ✅

**No more:**
- ❌ Manual Git tokens
- ❌ Copying repository URLs
- ❌ Entering credentials
- ❌ Configuration headaches

**Just seamless deployment. Like the big boys do it.** 🚀
