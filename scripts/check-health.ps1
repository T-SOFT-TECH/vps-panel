# Quick health check script for VPS Panel (PowerShell)

Write-Host "🏥 VPS Panel Health Check" -ForegroundColor Cyan
Write-Host "=========================" -ForegroundColor Cyan
Write-Host ""

# Check backend
Write-Host "Backend (port 8080): " -NoNewline
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -TimeoutSec 2
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Running" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Not running" -ForegroundColor Red
    Write-Host "   Start with: cd backend && go run cmd/server/main.go"
}

# Check frontend
Write-Host "Frontend (port 5173): " -NoNewline
try {
    $response = Invoke-WebRequest -Uri "http://localhost:5173" -UseBasicParsing -TimeoutSec 2
    if ($response.StatusCode -eq 200) {
        Write-Host "✅ Running" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Not running" -ForegroundColor Red
    Write-Host "   Start with: cd frontend && npm run dev"
}

Write-Host ""
Write-Host "Full URLs:"
Write-Host "  Frontend: http://localhost:5173"
Write-Host "  Backend:  http://localhost:8080"
Write-Host "  Health:   http://localhost:8080/health"
