#!/bin/bash
# VPS Panel Diagnostic and Fix Script
# Run this on your VPS: curl -sL https://raw.githubusercontent.com/T-SOFT-TECH/vps-panel/main/fix_panel.sh | bash

set -e

echo "=== VPS Panel Diagnostic & Fix Script ==="
echo ""

echo "1. Checking Frontend Service Status..."
if systemctl is-active --quiet vps-panel-frontend; then
    echo "✓ Frontend service is running"
    systemctl status vps-panel-frontend --no-pager
else
    echo "✗ Frontend service is NOT running"
    echo ""
    echo "2. Checking recent frontend logs..."
    journalctl -u vps-panel-frontend -n 30 --no-pager || true
    echo ""

    echo "3. Checking if frontend build exists..."
    if [ -d "/opt/vps-panel/frontend/build" ]; then
        echo "✓ Frontend build directory exists"
        ls -lh /opt/vps-panel/frontend/build/
    else
        echo "✗ Frontend build directory DOES NOT exist"
        echo ""
        echo "4. Building frontend..."
        cd /opt/vps-panel/frontend
        npm install
        npm run build
        echo "✓ Frontend built successfully"
    fi
    echo ""

    echo "5. Starting frontend service..."
    systemctl enable vps-panel-frontend
    systemctl restart vps-panel-frontend
    sleep 3

    if systemctl is-active --quiet vps-panel-frontend; then
        echo "✓ Frontend service started successfully"
    else
        echo "✗ Frontend service failed to start"
        echo "Check logs: journalctl -u vps-panel-frontend -n 50"
        exit 1
    fi
fi

echo ""
echo "6. Checking Backend Service Status..."
systemctl status vps-panel --no-pager | head -10

echo ""
echo "7. Checking Caddy Status..."
systemctl status caddy --no-pager | head -10

echo ""
echo "8. Checking Listening Ports..."
netstat -tlnp | grep -E ':(3456|3000|80|443)' || echo "No ports found"

echo ""
echo "=== Testing Panel Access ==="
echo "Checking http://localhost:3000..."
curl -s -o /dev/null -w "HTTP %{http_code}\n" http://localhost:3000 || echo "Frontend not responding"

echo "Checking http://localhost:3456..."
curl -s -o /dev/null -w "HTTP %{http_code}\n" http://localhost:3456/api/v1/auth/registration-status || echo "Backend not responding"

echo ""
echo "=== Diagnostic Complete ==="
echo ""
echo "If everything is working, access your panel at:"
echo "https://panel.tsoft-tech.dev"
