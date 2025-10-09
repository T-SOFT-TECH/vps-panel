# VPS Panel - API Testing Script (PowerShell)
# Run this on Windows

$API_URL = "http://localhost:8080/api/v1"
$TOKEN = ""
$PASSED = 0
$FAILED = 0

Write-Host "🧪 VPS Panel API Test Suite" -ForegroundColor Cyan
Write-Host "=============================" -ForegroundColor Cyan
Write-Host ""

# Test 1: Health Check
Write-Host "1️⃣  Health Check" -ForegroundColor Yellow
Write-Host "----------------"
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
    if ($response.status -eq "ok") {
        Write-Host "✓ Backend is healthy" -ForegroundColor Green
        $PASSED++
    }
} catch {
    Write-Host "✗ Backend health check failed" -ForegroundColor Red
    Write-Host "Make sure backend is running on port 8080"
    exit 1
}
Write-Host ""

# Test 2: Register
Write-Host "2️⃣  Authentication" -ForegroundColor Yellow
Write-Host "----------------"
$EMAIL = "test-$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"
Write-Host "Registering user: $EMAIL"

$registerBody = @{
    email = $EMAIL
    password = "password123"
    name = "Test User"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$API_URL/auth/register" -Method Post -Body $registerBody -ContentType "application/json"
    Write-Host "✓ Registration successful" -ForegroundColor Green
    $TOKEN = $response.token
    Write-Host "Token received: $($TOKEN.Substring(0, [Math]::Min(20, $TOKEN.Length)))..."
    $PASSED++
} catch {
    Write-Host "✗ Registration failed" -ForegroundColor Red
    Write-Host $_.Exception.Message
    $FAILED++
}
Write-Host ""

# Test 3: Login
Write-Host "3️⃣  Login" -ForegroundColor Yellow
Write-Host "--------"
$loginBody = @{
    email = $EMAIL
    password = "password123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$API_URL/auth/login" -Method Post -Body $loginBody -ContentType "application/json"
    Write-Host "✓ Login successful" -ForegroundColor Green
    $PASSED++
} catch {
    Write-Host "✗ Login failed" -ForegroundColor Red
    $FAILED++
}
Write-Host ""

# Test 4: Get current user
Write-Host "4️⃣  Get Current User" -ForegroundColor Yellow
Write-Host "------------------"
$headers = @{
    Authorization = "Bearer $TOKEN"
}

try {
    $response = Invoke-RestMethod -Uri "$API_URL/users/me" -Method Get -Headers $headers
    if ($response.email -eq $EMAIL) {
        Write-Host "✓ User info retrieved" -ForegroundColor Green
        $PASSED++
    }
} catch {
    Write-Host "✗ Failed to get user info" -ForegroundColor Red
    $FAILED++
}
Write-Host ""

# Test 5: Create project
Write-Host "5️⃣  Create Project" -ForegroundColor Yellow
Write-Host "----------------"
$projectBody = @{
    name = "Test Project"
    description = "Automated test project"
    git_url = "https://github.com/sveltejs/kit"
    git_branch = "main"
    framework = "sveltekit"
    baas_type = ""
    auto_deploy = $false
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$API_URL/projects" -Method Post -Body $projectBody -Headers $headers -ContentType "application/json"
    Write-Host "✓ Project created" -ForegroundColor Green
    $PROJECT_ID = $response.id
    Write-Host "Project ID: $PROJECT_ID"
    $PASSED++
} catch {
    Write-Host "✗ Failed to create project" -ForegroundColor Red
    Write-Host $_.Exception.Message
    $FAILED++
}
Write-Host ""

# Test 6: Get all projects
Write-Host "6️⃣  Get All Projects" -ForegroundColor Yellow
Write-Host "------------------"
try {
    $response = Invoke-RestMethod -Uri "$API_URL/projects" -Method Get -Headers $headers
    Write-Host "✓ Projects retrieved" -ForegroundColor Green
    Write-Host "Total projects: $($response.total)"
    $PASSED++
} catch {
    Write-Host "✗ Failed to get projects" -ForegroundColor Red
    $FAILED++
}
Write-Host ""

# Test 7: Get project by ID
if ($PROJECT_ID) {
    Write-Host "7️⃣  Get Project by ID" -ForegroundColor Yellow
    Write-Host "-------------------"
    try {
        $response = Invoke-RestMethod -Uri "$API_URL/projects/$PROJECT_ID" -Method Get -Headers $headers
        Write-Host "✓ Project retrieved by ID" -ForegroundColor Green
        $PASSED++
    } catch {
        Write-Host "✗ Failed to get project" -ForegroundColor Red
        $FAILED++
    }
    Write-Host ""
}

# Test 8: Unauthorized access
Write-Host "8️⃣  Unauthorized Access Test" -ForegroundColor Yellow
Write-Host "-------------------------"
try {
    $response = Invoke-RestMethod -Uri "$API_URL/projects" -Method Get
    Write-Host "✗ Security issue: Unauthorized request not blocked" -ForegroundColor Red
    $FAILED++
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "✓ Unauthorized request blocked (401)" -ForegroundColor Green
        $PASSED++
    } else {
        Write-Host "✗ Unexpected status code" -ForegroundColor Red
        $FAILED++
    }
}
Write-Host ""

# Test 9: Invalid login
Write-Host "9️⃣  Invalid Login Test" -ForegroundColor Yellow
Write-Host "--------------------"
$invalidLogin = @{
    email = $EMAIL
    password = "wrongpassword"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "$API_URL/auth/login" -Method Post -Body $invalidLogin -ContentType "application/json"
    Write-Host "✗ Security issue: Wrong password accepted" -ForegroundColor Red
    $FAILED++
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "✓ Invalid credentials rejected" -ForegroundColor Green
        $PASSED++
    }
}
Write-Host ""

# Summary
Write-Host "================================" -ForegroundColor Cyan
Write-Host "📊 Test Results Summary" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Passed: " -NoNewline
Write-Host "$PASSED" -ForegroundColor Green
Write-Host "Failed: " -NoNewline
Write-Host "$FAILED" -ForegroundColor Red
Write-Host "Total:  $($PASSED + $FAILED)"
Write-Host ""

if ($FAILED -eq 0) {
    Write-Host "🎉 All tests passed!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "❌ Some tests failed" -ForegroundColor Red
    exit 1
}
