#!/bin/bash
# Quick health check script for VPS Panel

echo "🏥 VPS Panel Health Check"
echo "========================="
echo ""

# Check backend
echo -n "Backend (port 8080): "
if curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Not running"
    echo "   Start with: cd backend && go run cmd/server/main.go"
fi

# Check frontend
echo -n "Frontend (port 5173): "
if curl -s http://localhost:5173 > /dev/null 2>&1; then
    echo "✅ Running"
else
    echo "❌ Not running"
    echo "   Start with: cd frontend && npm run dev"
fi

echo ""
echo "Full URLs:"
echo "  Frontend: http://localhost:5173"
echo "  Backend:  http://localhost:8080"
echo "  Health:   http://localhost:8080/health"
