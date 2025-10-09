#!/bin/bash

#############################################################
# VPS Panel - Deployment Script
#
# This script deploys the VPS Panel application after
# running install.sh
#
# Usage: ./deploy.sh [--skip-build]
#############################################################

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
INSTALL_DIR="/opt/vps-panel"
DATA_DIR="/var/lib/vps-panel"
LOG_DIR="/var/log/vps-panel"
PANEL_USER="vps-panel"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Logging
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
error_exit() { log_error "$1"; exit 1; }

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error_exit "This script must be run as root. Please use: sudo ./deploy.sh"
    fi
}

# Check if installation was completed
check_installation() {
    log_info "Checking if VPS Panel is installed..."

    if [[ ! -d "$INSTALL_DIR" ]]; then
        error_exit "VPS Panel is not installed. Please run install.sh first."
    fi

    if ! id "$PANEL_USER" &>/dev/null; then
        error_exit "Panel user not found. Please run install.sh first."
    fi

    log_success "Installation verified"
}

# Copy application files
copy_files() {
    log_info "Copying application files..."

    # Stop service if running
    if systemctl is-active --quiet vps-panel; then
        log_info "Stopping vps-panel service..."
        systemctl stop vps-panel
    fi

    # Backup existing installation
    if [[ -d "$INSTALL_DIR/backend" ]]; then
        log_info "Creating backup..."
        BACKUP_DIR="$INSTALL_DIR/backup-$(date +%Y%m%d-%H%M%S)"
        mkdir -p "$BACKUP_DIR"
        cp -r "$INSTALL_DIR"/{backend,frontend,vps-panel} "$BACKUP_DIR/" 2>/dev/null || true
        log_success "Backup created at: $BACKUP_DIR"
    fi

    # Copy backend
    log_info "Copying backend files..."
    rm -rf "$INSTALL_DIR/backend"
    cp -r "$SCRIPT_DIR/backend" "$INSTALL_DIR/"

    # Copy frontend build (if exists)
    if [[ -d "$SCRIPT_DIR/frontend" ]]; then
        log_info "Copying frontend files..."
        rm -rf "$INSTALL_DIR/frontend"
        cp -r "$SCRIPT_DIR/frontend" "$INSTALL_DIR/"
    fi

    # Set permissions
    chown -R $PANEL_USER:$PANEL_USER "$INSTALL_DIR"

    log_success "Files copied successfully"
}

# Build backend
build_backend() {
    if [[ "$1" == "--skip-build" ]]; then
        log_warning "Skipping backend build"
        return
    fi

    log_info "Building backend application..."

    cd "$INSTALL_DIR/backend"

    # Initialize Go modules if needed
    if [[ ! -f "go.mod" ]]; then
        log_info "Initializing Go modules..."
        sudo -u $PANEL_USER go mod init github.com/vps-panel/backend
    fi

    # Download dependencies
    log_info "Downloading Go dependencies..."
    sudo -u $PANEL_USER go mod download
    sudo -u $PANEL_USER go mod tidy

    # Build the application
    log_info "Compiling application..."
    sudo -u $PANEL_USER go build -o "$INSTALL_DIR/vps-panel" ./cmd/server

    # Make executable
    chmod +x "$INSTALL_DIR/vps-panel"

    log_success "Backend built successfully"
}

# Build frontend (optional)
build_frontend() {
    if [[ ! -d "$INSTALL_DIR/frontend" ]]; then
        log_warning "Frontend directory not found, skipping frontend build"
        return
    fi

    if [[ "$1" == "--skip-build" ]]; then
        log_warning "Skipping frontend build"
        return
    fi

    log_info "Building frontend application..."

    cd "$INSTALL_DIR/frontend"

    # Install dependencies
    log_info "Installing npm dependencies..."
    sudo -u $PANEL_USER npm install

    # Build frontend
    log_info "Building frontend..."
    sudo -u $PANEL_USER npm run build

    log_success "Frontend built successfully"
}

# Update configuration
update_config() {
    log_info "Updating configuration..."

    # Create config if it doesn't exist
    if [[ ! -f "$DATA_DIR/config.json" ]]; then
        log_info "Creating default configuration..."
        cat > "$DATA_DIR/config.json" << EOF
{
  "database_path": "$DATA_DIR/database/vps-panel.db",
  "projects_dir": "$DATA_DIR/projects",
  "port": 3456,
  "jwt_secret": "$(openssl rand -hex 32)",
  "caddy_config_path": "$DATA_DIR/caddy",
  "caddy_reload_cmd": "systemctl reload caddy"
}
EOF
        chown $PANEL_USER:$PANEL_USER "$DATA_DIR/config.json"
    fi

    # Create symlink to config
    ln -sf "$DATA_DIR/config.json" "$INSTALL_DIR/config.json"

    log_success "Configuration updated"
}

# Run database migrations
run_migrations() {
    log_info "Running database migrations..."

    # The application will auto-migrate on first run
    # This is just a placeholder for future manual migrations

    log_success "Database ready"
}

# Start services
start_services() {
    log_info "Starting VPS Panel services..."

    # Reload systemd
    systemctl daemon-reload

    # Enable and start the service
    systemctl enable vps-panel
    systemctl restart vps-panel

    # Wait for service to start
    sleep 3

    # Check if service is running
    if systemctl is-active --quiet vps-panel; then
        log_success "VPS Panel service started successfully"
    else
        log_error "Failed to start VPS Panel service"
        log_info "Check logs with: journalctl -u vps-panel -f"
        exit 1
    fi

    # Restart Caddy
    systemctl restart caddy
    log_success "Caddy restarted"
}

# Print status
print_status() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                        ║${NC}"
    echo -e "${GREEN}║  VPS Panel Deployed Successfully! 🚀                  ║${NC}"
    echo -e "${GREEN}║                                                        ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""

    # Get server IP
    SERVER_IP=$(hostname -I | awk '{print $1}')

    log_info "Service Status:"
    systemctl status vps-panel --no-pager | grep "Active:" | sed 's/^/  /'
    echo ""

    log_info "Access Panel:"
    echo "  • Local: http://localhost:3456"
    echo "  • Network: http://$SERVER_IP"
    echo ""

    log_info "Useful Commands:"
    echo "  • Check status: systemctl status vps-panel"
    echo "  • View logs: tail -f $LOG_DIR/panel.log"
    echo "  • Restart: systemctl restart vps-panel"
    echo "  • Stop: systemctl stop vps-panel"
    echo ""

    log_info "Next Steps:"
    echo "  1. Access the panel in your browser"
    echo "  2. Create your first admin account"
    echo "  3. Deploy your first project!"
    echo ""

    # Show recent logs
    log_info "Recent Logs:"
    tail -n 10 "$LOG_DIR/panel.log" 2>/dev/null | sed 's/^/  /' || echo "  No logs yet"
    echo ""
}

# Main deployment flow
main() {
    echo -e "${BLUE}"
    cat << "EOF"
╦  ╦╔═╗╔═╗  ╔═╗╔═╗╔╗╔╔═╗╦    ╔╦╗╔═╗╔═╗╦  ╔═╗╦ ╦
╚╗╔╝╠═╝╚═╗  ╠═╝╠═╣║║║║╣ ║     ║║║╣ ╠═╝║  ║ ║╚╦╝
 ╚╝ ╩  ╚═╝  ╩  ╩ ╩╝╚╝╚═╝╩═╝  ═╩╝╚═╝╩  ╩═╝╚═╝ ╩
EOF
    echo -e "${NC}"
    echo "VPS Panel Deployment Script"
    echo "============================="
    echo ""

    check_root
    check_installation

    log_info "Starting deployment..."
    echo ""

    copy_files
    build_backend "$@"
    build_frontend "$@"
    update_config
    run_migrations
    start_services

    print_status
}

# Run main
main "$@"
