#!/bin/bash

#############################################################
# VPS Panel - One-Command Installation Script
#
# Usage: curl -fsSL https://raw.githubusercontent.com/yourusername/vps-panel/main/install.sh | bash
# Or: wget -qO- https://raw.githubusercontent.com/yourusername/vps-panel/main/install.sh | bash
#
# System Requirements:
# - Ubuntu 20.04/22.04/24.04 or Debian 11/12
# - 2GB RAM minimum (4GB recommended)
# - 2 CPU cores minimum
# - 20GB disk space minimum
#############################################################

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/vps-panel"
DATA_DIR="/var/lib/vps-panel"
LOG_DIR="/var/log/vps-panel"
PANEL_USER="vps-panel"
PANEL_PORT="3456"
PANEL_DOMAIN=""  # Will be set during installation
GITHUB_REPO="yourusername/vps-panel" # Update this with your actual repo

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Error handler
error_exit() {
    log_error "$1"
    exit 1
}

# Banner
print_banner() {
    echo -e "${BLUE}"
    cat << "EOF"
╦  ╦╔═╗╔═╗  ╔═╗╔═╗╔╗╔╔═╗╦
╚╗╔╝╠═╝╚═╗  ╠═╝╠═╣║║║║╣ ║
 ╚╝ ╩  ╚═╝  ╩  ╩ ╩╝╚╝╚═╝╩═╝
EOF
    echo -e "${NC}"
    echo "VPS Panel Installation Script"
    echo "=============================="
    echo ""
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        error_exit "This script must be run as root. Please use: sudo bash install.sh"
    fi
}

# Detect OS
detect_os() {
    log_info "Detecting operating system..."

    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
        log_success "Detected: $PRETTY_NAME"
    else
        error_exit "Cannot detect OS. This script supports Ubuntu 20.04+ and Debian 11+"
    fi

    # Validate OS
    case $OS in
        ubuntu)
            if [[ ! "$OS_VERSION" =~ ^(20.04|22.04|24.04)$ ]]; then
                log_warning "Ubuntu $OS_VERSION detected. Recommended versions: 20.04, 22.04, 24.04"
            fi
            ;;
        debian)
            if [[ ! "$OS_VERSION" =~ ^(11|12)$ ]]; then
                log_warning "Debian $OS_VERSION detected. Recommended versions: 11, 12"
            fi
            ;;
        *)
            error_exit "Unsupported OS: $OS. This script supports Ubuntu and Debian only."
            ;;
    esac
}

# Check system requirements
check_requirements() {
    log_info "Checking system requirements..."

    # Check RAM
    TOTAL_RAM=$(free -m | awk '/^Mem:/{print $2}')
    if [[ $TOTAL_RAM -lt 2000 ]]; then
        log_warning "Low RAM detected: ${TOTAL_RAM}MB. Minimum 2GB recommended."
    else
        log_success "RAM: ${TOTAL_RAM}MB"
    fi

    # Check CPU
    CPU_CORES=$(nproc)
    if [[ $CPU_CORES -lt 2 ]]; then
        log_warning "Low CPU cores: $CPU_CORES. Minimum 2 cores recommended."
    else
        log_success "CPU Cores: $CPU_CORES"
    fi

    # Check disk space
    AVAILABLE_DISK=$(df -BG / | awk 'NR==2 {print $4}' | sed 's/G//')
    if [[ $AVAILABLE_DISK -lt 20 ]]; then
        log_warning "Low disk space: ${AVAILABLE_DISK}GB. Minimum 20GB recommended."
    else
        log_success "Available Disk: ${AVAILABLE_DISK}GB"
    fi
}

# Prompt for domain configuration
prompt_domain() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║             Domain Configuration                       ║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
    log_info "Configure the domain where you'll access VPS Panel"
    echo ""
    echo "  Options:"
    echo "    1. Enter a domain (e.g., panel.yourdomain.com) - Recommended for HTTPS"
    echo "    2. Use server IP - HTTP only (not recommended for production)"
    echo ""

    # Get server IP
    SERVER_IP=$(hostname -I | awk '{print $1}')
    if [[ -z "$SERVER_IP" ]]; then
        SERVER_IP=$(curl -s ifconfig.me)
    fi

    echo -e "  ${YELLOW}Detected Server IP: $SERVER_IP${NC}"
    echo ""

    while true; do
        read -p "Enter domain (or press Enter to use IP): " domain_input

        # If empty, use IP
        if [[ -z "$domain_input" ]]; then
            PANEL_DOMAIN="$SERVER_IP"
            log_warning "Using IP address: $PANEL_DOMAIN (HTTP only, no automatic HTTPS)"
            break
        fi

        # Validate domain format
        if [[ "$domain_input" =~ ^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$ ]]; then
            PANEL_DOMAIN="$domain_input"
            log_success "Domain configured: $PANEL_DOMAIN"
            echo ""
            log_info "Important DNS Setup:"
            echo "  → Add an A record for '$PANEL_DOMAIN' pointing to: $SERVER_IP"
            echo "  → Caddy will automatically obtain SSL certificate from Let's Encrypt"
            echo ""
            read -p "Press Enter to continue after DNS is configured (or Ctrl+C to cancel)..."
            break
        else
            log_error "Invalid domain format. Please enter a valid domain name."
        fi
    done

    echo ""
}

# Update system packages
update_system() {
    log_info "Updating system packages..."
    apt-get update -qq
    apt-get upgrade -y -qq
    apt-get install -y -qq curl wget git jq ca-certificates gnupg lsb-release
    log_success "System packages updated"
}

# Install Docker
install_docker() {
    if command -v docker &> /dev/null; then
        log_info "Docker is already installed ($(docker --version))"
        return
    fi

    log_info "Installing Docker..."

    # Add Docker's official GPG key
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/$OS/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg

    # Add the repository to Apt sources
    echo \
        "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$OS \
        $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

    # Install Docker Engine
    apt-get update -qq
    apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

    # Start and enable Docker
    systemctl start docker
    systemctl enable docker

    log_success "Docker installed: $(docker --version)"
}

# Install Go
install_go() {
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        log_info "Go is already installed ($GO_VERSION)"
        return
    fi

    log_info "Installing Go 1.23..."

    GO_VERSION="1.23.0"
    wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
    rm -rf /usr/local/go
    tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
    rm go${GO_VERSION}.linux-amd64.tar.gz

    # Add to PATH
    echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    export PATH=$PATH:/usr/local/go/bin

    log_success "Go installed: $(go version)"
}

# Install Node.js
install_nodejs() {
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node --version)
        log_info "Node.js is already installed ($NODE_VERSION)"
        return
    fi

    log_info "Installing Node.js 20..."

    # Install NodeSource repository
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
    apt-get install -y -qq nodejs

    log_success "Node.js installed: $(node --version)"
    log_success "npm installed: $(npm --version)"
}

# Install SQLite
install_sqlite() {
    if command -v sqlite3 &> /dev/null; then
        log_info "SQLite is already installed ($(sqlite3 --version))"
        return
    fi

    log_info "Installing SQLite..."
    apt-get install -y -qq sqlite3 libsqlite3-dev
    log_success "SQLite installed: $(sqlite3 --version)"
}

# Install Caddy
install_caddy() {
    if command -v caddy &> /dev/null; then
        log_info "Caddy is already installed ($(caddy version))"
        return
    fi

    log_info "Installing Caddy web server..."

    # Install Caddy
    apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
    apt-get update -qq
    apt-get install -y -qq caddy

    log_success "Caddy installed: $(caddy version)"
}

# Create panel user
create_panel_user() {
    if id "$PANEL_USER" &>/dev/null; then
        log_info "User '$PANEL_USER' already exists"
        return
    fi

    log_info "Creating system user '$PANEL_USER'..."
    useradd -r -m -d /home/$PANEL_USER -s /bin/bash $PANEL_USER
    usermod -aG docker $PANEL_USER
    log_success "User '$PANEL_USER' created"
}

# Create directories
create_directories() {
    log_info "Creating application directories..."

    mkdir -p $INSTALL_DIR
    mkdir -p $DATA_DIR/database
    mkdir -p $DATA_DIR/projects
    mkdir -p $LOG_DIR

    # Set permissions
    chown -R $PANEL_USER:$PANEL_USER $INSTALL_DIR
    chown -R $PANEL_USER:$PANEL_USER $DATA_DIR
    chown -R $PANEL_USER:$PANEL_USER $LOG_DIR

    log_success "Directories created"
}

# Download and build application
install_application() {
    log_info "Installing VPS Panel application..."

    cd $INSTALL_DIR

    # Clone repository (for now, we'll copy from local, but in production this would clone from GitHub)
    # git clone https://github.com/$GITHUB_REPO.git .

    # For now, we'll create a placeholder
    log_warning "Application installation will be completed after source code is available"
    log_info "Please manually copy your application files to $INSTALL_DIR"
}

# Create systemd service
create_systemd_service() {
    log_info "Creating systemd service..."

    cat > /etc/systemd/system/vps-panel.service << EOF
[Unit]
Description=VPS Panel - Web Hosting Control Panel
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
User=$PANEL_USER
Group=$PANEL_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/vps-panel
Restart=always
RestartSec=10
StandardOutput=append:$LOG_DIR/panel.log
StandardError=append:$LOG_DIR/panel-error.log

Environment="GIN_MODE=release"
Environment="DATABASE_PATH=$DATA_DIR/database/vps-panel.db"
Environment="PROJECTS_DIR=$DATA_DIR/projects"
Environment="PORT=$PANEL_PORT"

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    log_success "Systemd service created"
}

# Configure Caddy
configure_caddy() {
    log_info "Configuring Caddy reverse proxy..."

    # Create Caddyfile based on domain configuration
    if [[ "$PANEL_DOMAIN" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        # IP address - use HTTP only
        cat > /etc/caddy/Caddyfile << EOF
# VPS Panel - Main Interface (HTTP only - using IP)
http://$PANEL_DOMAIN {
    reverse_proxy localhost:$PANEL_PORT
}

# Also listen on port 80 for any host
:80 {
    reverse_proxy localhost:$PANEL_PORT
}

# Deployed applications will be configured here
import $DATA_DIR/caddy/*.caddy
EOF
        log_warning "Configured for HTTP only (using IP address)"
    else
        # Domain name - enable automatic HTTPS
        cat > /etc/caddy/Caddyfile << EOF
# VPS Panel - Main Interface (Auto HTTPS)
$PANEL_DOMAIN {
    reverse_proxy localhost:$PANEL_PORT
}

# Deployed applications will be configured here
import $DATA_DIR/caddy/*.caddy
EOF
        log_success "Configured with automatic HTTPS for: $PANEL_DOMAIN"
    fi

    # Create directory for dynamic configs
    mkdir -p $DATA_DIR/caddy
    chown -R $PANEL_USER:$PANEL_USER $DATA_DIR/caddy

    # Reload Caddy
    systemctl restart caddy
    systemctl enable caddy

    log_success "Caddy configured and started"
}

# Configure firewall
configure_firewall() {
    log_info "Configuring firewall..."

    if command -v ufw &> /dev/null; then
        ufw --force enable
        ufw allow 22/tcp comment 'SSH'
        ufw allow 80/tcp comment 'HTTP'
        ufw allow 443/tcp comment 'HTTPS'
        ufw allow $PANEL_PORT/tcp comment 'VPS Panel'
        log_success "UFW firewall configured"
    else
        log_warning "UFW not found. Please configure firewall manually."
    fi
}

# Create initial configuration
create_config() {
    log_info "Creating application configuration..."

    cat > $DATA_DIR/config.json << EOF
{
  "database_path": "$DATA_DIR/database/vps-panel.db",
  "projects_dir": "$DATA_DIR/projects",
  "port": $PANEL_PORT,
  "panel_domain": "$PANEL_DOMAIN",
  "jwt_secret": "$(openssl rand -hex 32)",
  "caddy_config_path": "$DATA_DIR/caddy",
  "caddy_reload_cmd": "systemctl reload caddy"
}
EOF

    chown $PANEL_USER:$PANEL_USER $DATA_DIR/config.json
    log_success "Configuration created"
}

# Print completion message
print_completion() {
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                                                        ║${NC}"
    echo -e "${GREEN}║  VPS Panel Installation Complete! 🎉                  ║${NC}"
    echo -e "${GREEN}║                                                        ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
    log_info "Installation Summary:"
    echo "  • Install Directory: $INSTALL_DIR"
    echo "  • Data Directory: $DATA_DIR"
    echo "  • Log Directory: $LOG_DIR"
    echo "  • Panel Port: $PANEL_PORT"
    echo "  • Panel User: $PANEL_USER"
    echo ""
    log_info "Installed Components:"
    echo "  • Docker: $(docker --version)"
    echo "  • Go: $(go version | awk '{print $3}')"
    echo "  • Node.js: $(node --version)"
    echo "  • npm: $(npm --version)"
    echo "  • Caddy: $(caddy version | head -n1)"
    echo "  • SQLite: $(sqlite3 --version | awk '{print $1}')"
    echo ""
    log_info "Next Steps:"
    echo "  1. Copy your application files to: $INSTALL_DIR"
    echo "  2. Build the application:"
    echo "     cd $INSTALL_DIR/backend"
    echo "     go build -o $INSTALL_DIR/vps-panel ./cmd/server"
    echo ""
    echo "  3. Start the service:"
    echo "     systemctl start vps-panel"
    echo "     systemctl enable vps-panel"
    echo ""
    echo "  4. Check service status:"
    echo "     systemctl status vps-panel"
    echo ""
    echo "  5. View logs:"
    echo "     tail -f $LOG_DIR/panel.log"
    echo ""

    # Show access URL based on configuration
    if [[ "$PANEL_DOMAIN" =~ ^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_info "Access the panel at:"
        echo "  • http://$PANEL_DOMAIN"
        echo ""
        log_warning "You're using an IP address. For HTTPS, configure a domain:"
        echo "  1. Update DNS: Add A record pointing to $PANEL_DOMAIN"
        echo "  2. Edit /etc/caddy/Caddyfile with your domain"
        echo "  3. Reload Caddy: systemctl reload caddy"
    else
        log_info "Access the panel at:"
        echo "  • https://$PANEL_DOMAIN (Automatic HTTPS)"
        echo "  • http://$PANEL_DOMAIN (will redirect to HTTPS)"
        echo ""
        log_success "Caddy will automatically obtain SSL certificate from Let's Encrypt!"
        echo ""
        log_info "Note: Make sure your DNS A record is configured:"
        echo "  Domain: $PANEL_DOMAIN → IP: $(hostname -I | awk '{print $1}')"
    fi
    echo ""
}

# Main installation flow
main() {
    print_banner
    check_root
    detect_os
    check_requirements
    prompt_domain

    log_info "Starting installation..."
    echo ""

    update_system
    install_docker
    install_go
    install_nodejs
    install_sqlite
    install_caddy
    create_panel_user
    create_directories
    create_config
    create_systemd_service
    configure_caddy
    configure_firewall

    print_completion
}

# Run main function
main "$@"
