# Testing Guide - VPS Panel

Complete guide to test your VPS Panel deployment platform.

## 🚀 Quick Test (2 minutes)

### 1. Start the Application

```bash
# Terminal 1 - Backend
cd backend
go run cmd/server/main.go

# Terminal 2 - Frontend
cd frontend
npm run dev
```

Wait for both servers to start:
- ✅ Backend: `🚀 Server starting on 0.0.0.0:8080`
- ✅ Frontend: `Local: http://localhost:5173/`

### 2. Test Backend Health

```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "ok",
  "service": "vps-panel-api"
}
```

✅ If you see this, backend is working!

### 3. Test Frontend

Open http://localhost:5173 in your browser. You should see:
- A loading spinner briefly
- Then redirect to login page

✅ If you see the login page, frontend is working!

## 📋 Complete Testing Workflow

### Phase 1: Authentication Testing

#### Test 1: Register a New User

1. **Open Frontend**: http://localhost:5173
2. **Click**: "Register here"
3. **Fill in**:
   ```
   Name: Test User
   Email: test@example.com
   Password: password123
   Confirm Password: password123
   ```
4. **Click**: "Create account"

**Expected Result:**
- ✅ Redirected to dashboard
- ✅ See "Test User" in navbar
- ✅ See empty dashboard with stats showing 0s

#### Test 2: Logout and Login

1. **Click**: "Sign out" (top right)
2. **Verify**: Redirected to login page
3. **Login**:
   ```
   Email: test@example.com
   Password: password123
   ```
4. **Click**: "Sign in"

**Expected Result:**
- ✅ Redirected to dashboard
- ✅ Logged in successfully

#### Test 3: Wrong Password

1. **Go to**: Login page
2. **Try wrong password**:
   ```
   Email: test@example.com
   Password: wrongpassword
   ```

**Expected Result:**
- ❌ Error message: "Invalid credentials"
- ✅ Stays on login page

### Phase 2: Project Management Testing

#### Test 4: Create a New Project

1. **Click**: "New Project" button
2. **Fill in the form**:

   **Basic Information:**
   ```
   Name: Test SvelteKit App
   Description: My first test project
   ```

   **Git Repository:**
   ```
   Repository URL: https://github.com/sveltejs/kit
   Branch: main
   ```

   **Framework & Backend:**
   ```
   Framework: SvelteKit
   Backend/BaaS: None
   ```

   **Leave other fields as defaults**

3. **Click**: "Create Project"

**Expected Result:**
- ✅ Success message appears
- ✅ Redirected to project detail page
- ✅ Project status shows "pending"

#### Test 5: View Projects List

1. **Click**: "Projects" in navbar
2. **Verify**:
   - ✅ See your "Test SvelteKit App" project
   - ✅ Shows framework badge
   - ✅ Shows branch name

#### Test 6: Search Projects

1. **On projects page**, type in search: "sveltekit"
2. **Verify**: Project appears

3. **Type**: "react"
4. **Verify**: "No projects found" message

#### Test 7: View Project Details

1. **Click** on your project card
2. **Verify**:
   - ✅ Project name and description shown
   - ✅ "Deploy" button visible
   - ✅ Project info sidebar shows framework, branch, etc.
   - ✅ "No deployments yet" message

### Phase 3: Deployment Testing (Simplified)

**Note:** Full deployment requires Docker and Git access. Here's what to expect:

#### Test 8: Trigger Deployment

1. **On project detail page**, click "Deploy"
2. **Verify**:
   - ✅ Button shows "Deploying..." with spinner
   - ✅ New deployment appears in list
   - ✅ Status shows "pending" or "building"

**Expected Behavior:**
- If you have Docker running: Deployment will progress through stages
- If you don't have Docker: Deployment will fail with error message (expected)

#### Test 9: View Deployment Logs

1. **Click** on a deployment in the list
2. **Verify**:
   - ✅ Deployment details page loads
   - ✅ Shows commit information (if available)
   - ✅ Shows deployment status
   - ✅ Build logs section visible

### Phase 4: API Testing (Using curl)

#### Test 10: Register via API

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "api-test@example.com",
    "password": "password123",
    "name": "API Test User"
  }'
```

**Expected Response:**
```json
{
  "token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "user": {
    "id": 2,
    "email": "api-test@example.com",
    "name": "API Test User",
    "role": "user"
  }
}
```

#### Test 11: Login via API

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Save the token from response for next tests!**

#### Test 12: Get Projects via API

```bash
# Replace YOUR_TOKEN with token from login
curl -X GET http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Expected Response:**
```json
{
  "projects": [
    {
      "id": 1,
      "name": "Test SvelteKit App",
      "framework": "sveltekit",
      ...
    }
  ],
  "total": 1
}
```

#### Test 13: Create Project via API

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "API Created Project",
    "git_url": "https://github.com/example/repo.git",
    "framework": "react",
    "git_branch": "main"
  }'
```

#### Test 14: Unauthorized Access

```bash
# Try without token
curl -X GET http://localhost:8080/api/v1/projects
```

**Expected Response:**
```json
{
  "error": "Missing authorization header"
}
```

**Status Code:** 401 Unauthorized ✅

### Phase 5: Error Handling Testing

#### Test 15: Database Persistence

1. **Stop both servers** (Ctrl+C)
2. **Restart servers**
3. **Login** with your test account
4. **Verify**: Your projects are still there ✅

This confirms SQLite database is working!

#### Test 16: Invalid Form Data

**Frontend Form Validation:**

1. **Go to**: New Project page
2. **Leave name empty**, click "Create Project"
3. **Expected**: "Please fill in all required fields" error

4. **Enter name** but leave Git URL empty
5. **Expected**: Required field validation error

#### Test 17: XSS Protection

1. **Create project** with name: `<script>alert('xss')</script>`
2. **View projects** list
3. **Verify**: Script tag is displayed as text, not executed ✅

## 🧪 Testing with Postman

### Import Collection

Create a new Postman collection with these requests:

**1. Register**
```
POST http://localhost:8080/api/v1/auth/register
Body (JSON):
{
  "email": "postman@example.com",
  "password": "password123",
  "name": "Postman User"
}
```

**2. Login**
```
POST http://localhost:8080/api/v1/auth/login
Body (JSON):
{
  "email": "postman@example.com",
  "password": "password123"
}

Save response token to environment variable!
```

**3. Get Projects**
```
GET http://localhost:8080/api/v1/projects
Headers:
Authorization: Bearer {{token}}
```

**4. Create Project**
```
POST http://localhost:8080/api/v1/projects
Headers:
Authorization: Bearer {{token}}
Body (JSON):
{
  "name": "Postman Project",
  "git_url": "https://github.com/example/repo.git",
  "framework": "vue"
}
```

## 🐛 Common Issues & Fixes

### Issue 1: Backend won't start

**Error:** `bind: address already in use`

**Fix:**
```bash
# Find what's using port 8080
# Windows
netstat -ano | findstr :8080

# Linux/Mac
lsof -i :8080

# Kill the process or change port in .env
PORT=8081
```

### Issue 2: Frontend shows "Failed to fetch"

**Check:**
1. Backend is running on port 8080
2. `frontend/.env` has correct API URL:
   ```
   VITE_API_URL=http://localhost:8080/api/v1
   ```

**Test backend:**
```bash
curl http://localhost:8080/health
```

### Issue 3: Login fails with no error

**Open Browser DevTools (F12):**
1. Go to **Console** tab
2. Look for errors
3. Go to **Network** tab
4. Try logging in again
5. Check the login request for error details

### Issue 4: Database locked

**Fix:**
```bash
# Stop backend
# Delete database
rm backend/data/vps-panel.db

# Restart backend
cd backend
go run cmd/server/main.go
```

## ✅ Testing Checklist

### Authentication
- [ ] Register new user
- [ ] Login with correct credentials
- [ ] Login with wrong password fails
- [ ] Logout works
- [ ] Token persists after page refresh
- [ ] Token expires correctly

### Projects
- [ ] Create new project
- [ ] View projects list
- [ ] Search/filter projects
- [ ] View project details
- [ ] Update project (settings)
- [ ] Delete project

### Deployments
- [ ] Trigger manual deployment
- [ ] View deployment list
- [ ] View deployment logs
- [ ] Cancel deployment (if in progress)
- [ ] Deployment status updates

### UI/UX
- [ ] Responsive on mobile
- [ ] Responsive on tablet
- [ ] Responsive on desktop
- [ ] Loading states show correctly
- [ ] Error messages are clear
- [ ] Success messages appear
- [ ] Form validation works

### API
- [ ] All endpoints return correct status codes
- [ ] Unauthorized requests blocked
- [ ] Invalid data rejected
- [ ] CORS headers present
- [ ] Response format consistent

### Security
- [ ] Passwords are hashed
- [ ] JWT tokens work correctly
- [ ] Protected routes require auth
- [ ] XSS protection working
- [ ] SQL injection prevented (GORM handles this)

## 🎯 Next Steps

### Add Automated Tests

**Backend (Go):**
```bash
# Create test file: backend/internal/api/handlers/auth_test.go

cd backend
go test ./...
```

**Frontend (Vitest):**
```bash
cd frontend
npm install -D vitest @testing-library/svelte
npm run test
```

### End-to-End Testing

Install Playwright:
```bash
cd frontend
npm install -D @playwright/test
npx playwright install
```

### Load Testing

Use `hey` or `ab`:
```bash
# Install hey
go install github.com/rakyll/hey@latest

# Test login endpoint
hey -n 1000 -c 10 -m POST \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  http://localhost:8080/api/v1/auth/login
```

## 📊 Test Results Template

Keep track of your testing:

```
✅ PASSED: User registration
✅ PASSED: User login
✅ PASSED: Project creation
✅ PASSED: Project list view
⚠️  PARTIAL: Deployment (requires Docker)
❌ FAILED: Webhook handling (needs setup)

Notes:
- All authentication flows working correctly
- Project management fully functional
- Deployment tested manually with Docker
- Webhooks need Git provider configuration
```

## 🎉 Success Criteria

Your VPS Panel is working correctly when:

1. ✅ You can register and login
2. ✅ You can create and view projects
3. ✅ Projects are saved to database
4. ✅ UI is responsive and user-friendly
5. ✅ API returns correct responses
6. ✅ Errors are handled gracefully
7. ✅ Data persists after restart

**You're ready for production when all automated tests pass!**

Happy Testing! 🚀
