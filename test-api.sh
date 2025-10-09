#!/bin/bash
# VPS Panel - API Testing Script

API_URL="http://localhost:8080/api/v1"
TOKEN=""

echo "🧪 VPS Panel API Test Suite"
echo "============================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
PASSED=0
FAILED=0

# Helper function to test endpoint
test_endpoint() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local auth=$5

    echo -n "Testing: $name ... "

    if [ "$auth" = "true" ]; then
        response=$(curl -s -X $method "$API_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $TOKEN" \
            -d "$data")
    else
        response=$(curl -s -X $method "$API_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data")
    fi

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ PASSED${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        ((FAILED++))
        return 1
    fi
}

# Test 1: Health Check
echo "1️⃣  Health Check"
echo "----------------"
response=$(curl -s http://localhost:8080/health)
if echo "$response" | grep -q "ok"; then
    echo -e "${GREEN}✓ Backend is healthy${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Backend health check failed${NC}"
    echo "Make sure backend is running on port 8080"
    exit 1
fi
echo ""

# Test 2: Register
echo "2️⃣  Authentication"
echo "----------------"
EMAIL="test-$(date +%s)@example.com"
echo "Registering user: $EMAIL"

register_data="{\"email\":\"$EMAIL\",\"password\":\"password123\",\"name\":\"Test User\"}"
response=$(curl -s -X POST "$API_URL/auth/register" \
    -H "Content-Type: application/json" \
    -d "$register_data")

if echo "$response" | grep -q "token"; then
    echo -e "${GREEN}✓ Registration successful${NC}"
    TOKEN=$(echo "$response" | grep -o '"token":"[^"]*' | cut -d'"' -f4)
    echo "Token received: ${TOKEN:0:20}..."
    ((PASSED++))
else
    echo -e "${RED}✗ Registration failed${NC}"
    echo "Response: $response"
    ((FAILED++))
fi
echo ""

# Test 3: Login
echo "3️⃣  Login"
echo "--------"
login_data="{\"email\":\"$EMAIL\",\"password\":\"password123\"}"
response=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "$login_data")

if echo "$response" | grep -q "token"; then
    echo -e "${GREEN}✓ Login successful${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Login failed${NC}"
    ((FAILED++))
fi
echo ""

# Test 4: Get current user
echo "4️⃣  Get Current User"
echo "------------------"
response=$(curl -s -X GET "$API_URL/users/me" \
    -H "Authorization: Bearer $TOKEN")

if echo "$response" | grep -q "$EMAIL"; then
    echo -e "${GREEN}✓ User info retrieved${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Failed to get user info${NC}"
    ((FAILED++))
fi
echo ""

# Test 5: Create project
echo "5️⃣  Create Project"
echo "----------------"
project_data="{
    \"name\":\"Test Project\",
    \"description\":\"Automated test project\",
    \"git_url\":\"https://github.com/sveltejs/kit\",
    \"git_branch\":\"main\",
    \"framework\":\"sveltekit\",
    \"baas_type\":\"\",
    \"auto_deploy\":false
}"

response=$(curl -s -X POST "$API_URL/projects" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "$project_data")

if echo "$response" | grep -q "Test Project"; then
    echo -e "${GREEN}✓ Project created${NC}"
    PROJECT_ID=$(echo "$response" | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
    echo "Project ID: $PROJECT_ID"
    ((PASSED++))
else
    echo -e "${RED}✗ Failed to create project${NC}"
    echo "Response: $response"
    ((FAILED++))
fi
echo ""

# Test 6: Get all projects
echo "6️⃣  Get All Projects"
echo "------------------"
response=$(curl -s -X GET "$API_URL/projects" \
    -H "Authorization: Bearer $TOKEN")

if echo "$response" | grep -q "projects"; then
    echo -e "${GREEN}✓ Projects retrieved${NC}"
    project_count=$(echo "$response" | grep -o '"total":[0-9]*' | cut -d':' -f2)
    echo "Total projects: $project_count"
    ((PASSED++))
else
    echo -e "${RED}✗ Failed to get projects${NC}"
    ((FAILED++))
fi
echo ""

# Test 7: Get project by ID
if [ -n "$PROJECT_ID" ]; then
    echo "7️⃣  Get Project by ID"
    echo "-------------------"
    response=$(curl -s -X GET "$API_URL/projects/$PROJECT_ID" \
        -H "Authorization: Bearer $TOKEN")

    if echo "$response" | grep -q "Test Project"; then
        echo -e "${GREEN}✓ Project retrieved by ID${NC}"
        ((PASSED++))
    else
        echo -e "${RED}✗ Failed to get project${NC}"
        ((FAILED++))
    fi
    echo ""
fi

# Test 8: Unauthorized access
echo "8️⃣  Unauthorized Access Test"
echo "-------------------------"
response=$(curl -s -w "\n%{http_code}" -X GET "$API_URL/projects")
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✓ Unauthorized request blocked (401)${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Security issue: Got status $http_code instead of 401${NC}"
    ((FAILED++))
fi
echo ""

# Test 9: Invalid login
echo "9️⃣  Invalid Login Test"
echo "--------------------"
invalid_data="{\"email\":\"$EMAIL\",\"password\":\"wrongpassword\"}"
response=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -d "$invalid_data")
http_code=$(echo "$response" | tail -n1)

if [ "$http_code" = "401" ]; then
    echo -e "${GREEN}✓ Invalid credentials rejected${NC}"
    ((PASSED++))
else
    echo -e "${RED}✗ Security issue: Wrong password accepted${NC}"
    ((FAILED++))
fi
echo ""

# Summary
echo "================================"
echo "📊 Test Results Summary"
echo "================================"
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Failed: ${RED}$FAILED${NC}"
echo "Total:  $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed${NC}"
    exit 1
fi
