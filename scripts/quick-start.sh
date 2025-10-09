#!/bin/bash
# Quick start script - starts both frontend and backend

echo "🚀 Starting VPS Panel..."
echo ""

# Check if go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.23+"
    exit 1
fi

# Check if node is installed
if ! command -v node &> /dev/null; then
    echo "❌ Node.js is not installed. Please install Node.js 20+"
    exit 1
fi

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Copying from .env.example..."
    cp .env .env
    echo "✅ Created .env file"
    echo "⚠️  Please edit .env and set JWT_SECRET and WEBHOOK_SECRET"
    echo ""
fi

# Check if frontend/.env exists
if [ ! -f frontend/.env ]; then
    echo "⚠️  frontend/.env not found. Copying from frontend/.env.example..."
    cp frontend/.env frontend/.env
    echo "✅ Created frontend/.env file"
    echo ""
fi

echo "Starting backend..."
cd backend
go run cmd/server/main.go &
BACKEND_PID=$!

echo "Waiting for backend to start..."
sleep 3

echo ""
echo "Starting frontend..."
cd ../frontend
npm run dev &
FRONTEND_PID=$!

echo ""
echo "================================"
echo "✅ VPS Panel is starting!"
echo "================================"
echo ""
echo "Frontend: http://localhost:5173"
echo "Backend:  http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop both servers"
echo ""

# Wait for Ctrl+C
trap "kill $BACKEND_PID $FRONTEND_PID; exit" INT
wait
