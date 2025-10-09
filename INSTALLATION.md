# VPS Panel - Installation Guide

Complete guide for installing VPS Panel on a fresh Ubuntu/Debian server.

## 📋 System Requirements

### Minimum Requirements
- **OS**: Ubuntu 20.04/22.04/24.04 or Debian 11/12
- **RAM**: 2GB (4GB recommended)
- **CPU**: 2 cores (4 cores recommended)
- **Disk**: 20GB (50GB+ recommended)
- **Architecture**: x86_64 / AMD64

### Recommended Specifications
- **RAM**: 8GB+
- **CPU**: 4+ cores
- **Disk**: 256GB SSD
- **Network**: Stable internet connection

## 🚀 One-Command Installation

For a fresh Ubuntu/Debian server, run:

```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/vps-panel/main/install.sh | sudo bash
```

Or using wget:

```bash
wget -qO- https://raw.githubusercontent.com/yourusername/vps-panel/main/install.sh | sudo bash
```

This will install all dependencies and configure the system.

## 📦 What Gets Installed

The installation script automatically installs:

- ✅ **Docker** (latest stable version)
- ✅ **Go 1.23** (programming language for backend)
- ✅ **Node.js 20** (for building frontend)
- ✅ **SQLite** (database)
- ✅ **Caddy** (reverse proxy with automatic HTTPS)
- ✅ **System user and permissions**
- ✅ **Systemd services**
- ✅ **Firewall rules**

## 📁 Directory Structure

After installation, the following directories are created:

```
/opt/vps-panel/          # Application files
├── backend/             # Go backend source
├── frontend/            # SvelteKit frontend
├── vps-panel            # Compiled binary
└── config.json          # Configuration symlink

/var/lib/vps-panel/      # Data directory
├── database/            # SQLite database
├── projects/            # Cloned repositories
├── caddy/               # Caddy configurations
└── config.json          # Main configuration

/var/log/vps-panel/      # Log files
├── panel.log            # Application logs
└── panel-error.log      # Error logs
```

## 🔧 Manual Installation Steps

If you prefer manual installation or need to customize:

### Step 1: Update System

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y curl wget git
```

### Step 2: Install Docker

```bash
# Add Docker repository
curl -fsSL https://get.docker.com | sudo sh

# Start Docker
sudo systemctl start docker
sudo systemctl enable docker
```

### Step 3: Install Go

```bash
# Download and install Go 1.23
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee -a /etc/profile
source /etc/profile
```

### Step 4: Install Node.js

```bash
# Install Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo bash -
sudo apt install -y nodejs
```

### Step 5: Install SQLite

```bash
sudo apt install -y sqlite3 libsqlite3-dev
```

### Step 6: Install Caddy

```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install -y caddy
```

### Step 7: Create System User

```bash
sudo useradd -r -m -d /home/vps-panel -s /bin/bash vps-panel
sudo usermod -aG docker vps-panel
```

### Step 8: Create Directories

```bash
sudo mkdir -p /opt/vps-panel
sudo mkdir -p /var/lib/vps-panel/{database,projects,caddy}
sudo mkdir -p /var/log/vps-panel
sudo chown -R vps-panel:vps-panel /opt/vps-panel /var/lib/vps-panel /var/log/vps-panel
```

## 📥 Deploying the Application

After installation, deploy the application:

### Option 1: Using Git (Recommended)

```bash
# Clone repository
cd /opt/vps-panel
sudo -u vps-panel git clone https://github.com/yourusername/vps-panel.git .

# Run deployment script
sudo ./deploy.sh
```

### Option 2: Manual Upload

1. **Upload your code** to `/opt/vps-panel`

```bash
# From your local machine
scp -r vps-panel/* user@your-server-ip:/tmp/vps-panel/

# On the server
sudo mv /tmp/vps-panel/* /opt/vps-panel/
sudo chown -R vps-panel:vps-panel /opt/vps-panel
```

2. **Run deployment script**

```bash
cd /opt/vps-panel
sudo ./deploy.sh
```

### Option 3: Build Manually

```bash
# Build backend
cd /opt/vps-panel/backend
sudo -u vps-panel go mod download
sudo -u vps-panel go build -o /opt/vps-panel/vps-panel ./cmd/server

# Build frontend (optional)
cd /opt/vps-panel/frontend
sudo -u vps-panel npm install
sudo -u vps-panel npm run build

# Start service
sudo systemctl start vps-panel
sudo systemctl enable vps-panel
```

## 🔐 Configuration

### Basic Configuration

Edit `/var/lib/vps-panel/config.json`:

```json
{
  "database_path": "/var/lib/vps-panel/database/vps-panel.db",
  "projects_dir": "/var/lib/vps-panel/projects",
  "port": 3456,
  "jwt_secret": "your-secret-key-here",
  "caddy_config_path": "/var/lib/vps-panel/caddy",
  "caddy_reload_cmd": "systemctl reload caddy"
}
```

### Environment Variables

You can also configure via systemd service environment variables:

```bash
sudo systemctl edit vps-panel
```

Add:

```ini
[Service]
Environment="PORT=3456"
Environment="DATABASE_PATH=/var/lib/vps-panel/database/vps-panel.db"
```

## 🌐 Setting Up Domain & HTTPS

### Configure Caddy for HTTPS

Edit `/etc/caddy/Caddyfile`:

```caddy
# Replace with your domain
panel.yourdomain.com {
    reverse_proxy localhost:3456
}

# Import dynamic configs for deployed apps
import /var/lib/vps-panel/caddy/*.caddy
```

Reload Caddy:

```bash
sudo systemctl reload caddy
```

Caddy will automatically obtain SSL certificates from Let's Encrypt!

## 🔥 Firewall Configuration

### Using UFW (Ubuntu)

```bash
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS
sudo ufw allow 3456/tcp  # VPS Panel (if direct access needed)
sudo ufw enable
```

### Using iptables

```bash
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 3456 -j ACCEPT
```

## 📊 Service Management

### Start/Stop/Restart

```bash
# Start
sudo systemctl start vps-panel

# Stop
sudo systemctl stop vps-panel

# Restart
sudo systemctl restart vps-panel

# Enable auto-start on boot
sudo systemctl enable vps-panel

# Check status
sudo systemctl status vps-panel
```

### View Logs

```bash
# Application logs
tail -f /var/log/vps-panel/panel.log

# Error logs
tail -f /var/log/vps-panel/panel-error.log

# Systemd journal
journalctl -u vps-panel -f

# Last 100 lines
journalctl -u vps-panel -n 100
```

## 🧪 Testing Installation

### Check Services

```bash
# Check Docker
docker --version
docker ps

# Check Go
go version

# Check Node.js
node --version
npm --version

# Check VPS Panel
systemctl status vps-panel

# Check Caddy
systemctl status caddy
```

### Test Web Interface

```bash
# Get server IP
hostname -I

# Access panel
curl http://localhost:3456
# or
curl http://YOUR_SERVER_IP
```

## 🔄 Updating

To update the VPS Panel:

```bash
cd /opt/vps-panel

# Pull latest changes
sudo -u vps-panel git pull

# Redeploy
sudo ./deploy.sh
```

## 🐛 Troubleshooting

### Service Won't Start

```bash
# Check logs
journalctl -u vps-panel -n 50 --no-pager

# Check binary
ls -lh /opt/vps-panel/vps-panel

# Check permissions
ls -ld /opt/vps-panel /var/lib/vps-panel

# Rebuild
cd /opt/vps-panel
sudo ./deploy.sh
```

### Docker Issues

```bash
# Check Docker daemon
systemctl status docker

# Check Docker socket
ls -lh /var/run/docker.sock

# Test Docker
docker run hello-world

# Check user in docker group
groups vps-panel
```

### Database Issues

```bash
# Check database file
ls -lh /var/lib/vps-panel/database/vps-panel.db

# Check permissions
sudo -u vps-panel sqlite3 /var/lib/vps-panel/database/vps-panel.db ".tables"
```

### Port Already in Use

```bash
# Check what's using port 3456
sudo lsof -i :3456

# Change port in config
sudo nano /var/lib/vps-panel/config.json

# Restart service
sudo systemctl restart vps-panel
```

## 🗑️ Uninstallation

To completely remove VPS Panel:

```bash
# Stop and disable service
sudo systemctl stop vps-panel
sudo systemctl disable vps-panel

# Remove service file
sudo rm /etc/systemd/system/vps-panel.service
sudo systemctl daemon-reload

# Remove application files
sudo rm -rf /opt/vps-panel

# Remove data (CAUTION: This deletes all projects and databases!)
sudo rm -rf /var/lib/vps-panel

# Remove logs
sudo rm -rf /var/log/vps-panel

# Remove user
sudo userdel -r vps-panel

# Optionally remove dependencies
sudo apt remove --purge -y docker-ce docker-ce-cli containerd.io caddy
sudo rm -rf /usr/local/go
```

## 📞 Support

For issues and questions:
- GitHub Issues: https://github.com/yourusername/vps-panel/issues
- Documentation: https://docs.yourpanel.com
- Community: https://discord.gg/yourpanel

## 📄 License

[Your License Here]
