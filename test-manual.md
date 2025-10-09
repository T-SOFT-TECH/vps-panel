# Manual Testing Checklist - VPS Panel

Follow this step-by-step guide to test your VPS Panel.

## 🎬 Prerequisites

**Before testing, ensure:**
- [ ] Backend is running on port 8080
- [ ] Frontend is running on port 5173
- [ ] Browser DevTools open (F12) - Console tab visible

## 🚀 Step 1: Start the Application (2 min)

### Terminal 1 - Backend
```bash
cd backend
go run cmd/server/main.go
```

**Wait for:**
```
✅ Database initialized successfully
🚀 Server starting on 0.0.0.0:8080
```

### Terminal 2 - Frontend
```bash
cd frontend
npm run dev
```

**Wait for:**
```
✅ Local:   http://localhost:5173/
```

### Verify Backend Health
```bash
curl http://localhost:8080/health
```

**Expected:** `{"status":"ok","service":"vps-panel-api"}`

---

## 🔐 Step 2: Test Authentication (3 min)

### Test 2.1: Registration

1. **Open:** http://localhost:5173
2. **Click:** "Register here" link
3. **Fill form:**
   ```
   Name: John Doe
   Email: john@example.com
   Password: password123
   Confirm Password: password123
   ```
4. **Click:** "Create account"

**✅ Expected:**
- Redirected to dashboard
- See "John Doe" in navbar (top right)
- See "0" in all stat cards
- No errors in console

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 2.2: Logout

1. **Click:** "Sign out" button (top right)

**✅ Expected:**
- Redirected to /login
- Navbar shows "Sign in" button
- No user info visible

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 2.3: Login

1. **Fill form:**
   ```
   Email: john@example.com
   Password: password123
   ```
2. **Click:** "Sign in"

**✅ Expected:**
- Redirected to dashboard
- Logged in as "John Doe"
- Stats show (still zeros)

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 2.4: Invalid Login

1. **Logout**
2. **Try login with:**
   ```
   Email: john@example.com
   Password: wrongpassword
   ```

**✅ Expected:**
- Red error alert: "Invalid credentials"
- Stays on login page

**📝 Test Result:** ☐ Pass ☐ Fail

---

## 📦 Step 3: Test Project Management (5 min)

### Test 3.1: Create First Project

1. **Click:** "New Project" button (top right)
2. **Fill form:**

   **Basic Information:**
   ```
   Name: My SvelteKit Blog
   Description: A personal blog built with SvelteKit
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

   **Build Configuration (leave defaults):**
   ```
   Install Command: npm install
   Build Command: npm run build
   Output Directory: build
   Node Version: Node.js 20
   ```

   **Port Configuration:**
   ```
   Frontend Port: 3000
   ```

   **Auto Deploy:** ✅ Checked

3. **Click:** "Create Project"

**✅ Expected:**
- Green success message
- Redirected to project detail page
- Project name shown
- Status badge: "pending"
- "No deployments yet" message

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 3.2: Create Second Project

1. **Click:** "Projects" in navbar
2. **Click:** "New Project"
3. **Fill form:**
   ```
   Name: React Dashboard
   Repository URL: https://github.com/facebook/react
   Branch: main
   Framework: React
   ```
4. **Click:** "Create Project"

**✅ Expected:**
- Successfully created
- Redirected to new project page

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 3.3: View Projects List

1. **Click:** "Projects" in navbar

**✅ Expected:**
- See 2 project cards
- "My SvelteKit Blog" card
- "React Dashboard" card
- Each shows framework badge
- Each shows status badge

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 3.4: Search Projects

1. **On projects page**
2. **Type in search:** "svelte"

**✅ Expected:**
- Only "My SvelteKit Blog" shown
- "React Dashboard" hidden

3. **Clear search**
4. **Type:** "react"

**✅ Expected:**
- Only "React Dashboard" shown

5. **Type:** "nonexistent"

**✅ Expected:**
- "No projects found" message
- "Try adjusting your search" text

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 3.5: View Project Details

1. **Clear search**
2. **Click:** "My SvelteKit Blog" card

**✅ Expected:**
- Project name in header
- Description shown
- "pending" status badge
- Deploy button visible
- Right sidebar shows:
  - Framework: sveltekit
  - Branch: main
  - Auto Deploy: Enabled
- Repository link clickable
- "No deployments yet" in deployment section

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 3.6: Return to Dashboard

1. **Click:** "Dashboard" in navbar (or logo)

**✅ Expected:**
- Stats updated:
  - Total Projects: 2
  - Active Projects: 0
  - Deployments: 0
- "Recent Projects" section shows both projects
- Each project clickable

**📝 Test Result:** ☐ Pass ☐ Fail

---

## 🚀 Step 4: Test Deployment (2 min)

### Test 4.1: Trigger Deployment

1. **Go to:** "My SvelteKit Blog" project
2. **Click:** "Deploy" button

**✅ Expected:**
- Button changes to "Deploying..." with spinner
- New deployment appears in list
- Status shows "pending" or "building"

**Note:** If you don't have Docker installed, deployment will fail. This is expected!

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 4.2: View Deployment

1. **Click:** on the deployment in the list

**✅ Expected:**
- Deployment details page loads
- Shows deployment ID
- Shows status badge
- Commit information section (may be empty)
- Build logs section

**If deployment failed (no Docker):**
- Status: "failed"
- Error message visible
- This is OK for testing!

**📝 Test Result:** ☐ Pass ☐ Fail

---

## 🎨 Step 5: Test UI/UX (3 min)

### Test 5.1: Responsive Design

**Desktop (current view):**
- [ ] Navbar shows all links
- [ ] Stats in 4 columns
- [ ] Projects in 3 columns

**Resize browser to tablet (768px):**
- [ ] Stats in 2 columns
- [ ] Projects in 2 columns
- [ ] Navbar still readable

**Resize to mobile (375px):**
- [ ] Hamburger menu appears
- [ ] Stats in 1 column
- [ ] Projects in 1 column
- [ ] All text readable

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 5.2: Form Validation

1. **Go to:** New Project page
2. **Leave name empty**
3. **Click:** "Create Project"

**✅ Expected:**
- Red error: "Please fill in all required fields"
- Form not submitted

4. **Fill name:** "Test"
5. **Leave Git URL empty**
6. **Click:** "Create Project"

**✅ Expected:**
- HTML5 validation message
- Form not submitted

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 5.3: Loading States

1. **Logout**
2. **Login again**
3. **Watch carefully:**

**✅ Expected:**
- Login button shows "Signing in..." with spinner
- Button disabled during login
- Smooth transition to dashboard

**📝 Test Result:** ☐ Pass ☐ Fail

---

## 🔒 Step 6: Test Security (2 min)

### Test 6.1: Protected Routes

1. **Logout**
2. **Try to visit:** http://localhost:5173/projects

**✅ Expected:**
- Redirected to /login
- Not able to view projects without login

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 6.2: Session Persistence

1. **Login**
2. **Refresh page (F5)**

**✅ Expected:**
- Still logged in
- No redirect to login
- User data still visible

3. **Close browser tab**
4. **Open new tab:** http://localhost:5173

**✅ Expected:**
- Still logged in (token in localStorage)

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 6.3: XSS Protection

1. **Create project with name:**
   ```
   <script>alert('XSS')</script>
   ```
2. **View projects list**

**✅ Expected:**
- Script tag shown as text
- No alert popup
- Text displayed safely

**📝 Test Result:** ☐ Pass ☐ Fail

---

## 💾 Step 7: Test Data Persistence (1 min)

### Test 7.1: Database Persistence

1. **Stop both servers** (Ctrl+C in both terminals)
2. **Wait 5 seconds**
3. **Restart both servers**
4. **Login** with same credentials

**✅ Expected:**
- Can login successfully
- All projects still there
- All data intact

**📝 Test Result:** ☐ Pass ☐ Fail

---

## 🧪 Step 8: Test API with Browser DevTools (2 min)

### Test 8.1: Inspect Network Requests

1. **Open DevTools** (F12)
2. **Go to:** Network tab
3. **Login**
4. **Find:** Login request

**✅ Check:**
- Request URL: `http://localhost:8080/api/v1/auth/login`
- Method: POST
- Status: 200
- Response contains: token, user

**📝 Test Result:** ☐ Pass ☐ Fail

---

### Test 8.2: Check Authorization Headers

1. **In Network tab**
2. **Click:** "Projects" in navbar
3. **Find:** Projects request
4. **Click:** on request
5. **Go to:** Headers tab

**✅ Check:**
- Request has `Authorization: Bearer [token]`
- Status: 200

**📝 Test Result:** ☐ Pass ☐ Fail

---

## 🗑️ Step 9: Test Delete Functionality (1 min)

### Test 9.1: Delete Project

1. **Go to:** "React Dashboard" project
2. **Click:** "Delete" button
3. **Modal appears**

**✅ Expected:**
- Confirmation modal shown
- Warning message about deletion
- "Cancel" and "Delete Project" buttons

4. **Click:** "Cancel"

**✅ Expected:**
- Modal closes
- Project still exists

5. **Click:** "Delete" again
6. **Click:** "Delete Project"

**✅ Expected:**
- Redirected to projects list
- "React Dashboard" no longer in list
- Only 1 project remains

**📝 Test Result:** ☐ Pass ☐ Fail

---

## 📊 Final Verification

**Check Dashboard Stats:**
- Total Projects: 1
- Active Projects: 0
- Deployments: 1 (or more if you deployed multiple times)

**Check Console (F12):**
- No errors in console
- No warnings about failed requests

**Check Database:**
```bash
# Backend terminal should show:
✅ Database initialized successfully
# No errors about schema
```

---

## ✅ Test Summary

Fill in your results:

### Authentication
- [ ] Registration works
- [ ] Login works
- [ ] Logout works
- [ ] Invalid login rejected
- [ ] Session persists

### Projects
- [ ] Create project works
- [ ] View projects list works
- [ ] Search projects works
- [ ] View project details works
- [ ] Delete project works

### Deployments
- [ ] Trigger deployment works
- [ ] View deployment works
- [ ] Deployment logs visible

### UI/UX
- [ ] Responsive on mobile
- [ ] Responsive on tablet
- [ ] Form validation works
- [ ] Loading states work
- [ ] Error messages clear

### Security
- [ ] Protected routes work
- [ ] Session persistence works
- [ ] XSS protection works

### Data
- [ ] Database persists after restart
- [ ] All CRUD operations work

---

## 🎯 Overall Result

**Total Tests:** 23

**Passed:** _____ / 23

**Failed:** _____ / 23

**Success Rate:** _____ %

---

## 🎉 Next Steps

If all tests passed:
- ✅ Your VPS Panel is working perfectly!
- 📚 Read the [Architecture docs](./ARCHITECTURE.md)
- 🚀 Try deploying to production
- 🎨 Customize the UI to your liking

If some tests failed:
- 🐛 Check [TESTING.md](./TESTING.md) troubleshooting section
- 💬 Review console errors in browser DevTools
- 📝 Check backend terminal for error messages

---

**Testing completed on:** _______________

**Tested by:** _______________

**Notes:**
```
[Add any observations or issues encountered]
```
