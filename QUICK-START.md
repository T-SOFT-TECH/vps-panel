# VPS Panel - Quick Start Guide

## 🚀 One-Command Installation

### For Fresh Ubuntu/Debian VPS:

```bash
curl -fsSL https://raw.githubusercontent.com/yourusername/vps-panel/main/install.sh | sudo bash
```

Or download and run locally:

```bash
wget https://raw.githubusercontent.com/yourusername/vps-panel/main/install.sh
chmod +x install.sh
sudo ./install.sh
```

## 📋 Installation Process

### Step 1: System Check
The script will:
- ✅ Verify you're running as root
- ✅ Detect your OS (Ubuntu/Debian)
- ✅ Check RAM, CPU, and disk space

### Step 2: Domain Configuration

You'll be prompted to configure your access domain:

```
╔════════════════════════════════════════════════════════╗
║             Domain Configuration                       ║
╚════════════════════════════════════════════════════════╝

Configure the domain where you'll access VPS Panel

  Options:
    1. Enter a domain (e.g., panel.yourdomain.com) - Recommended for HTTPS
    2. Use server IP - HTTP only (not recommended for production)

  Detected Server IP: 123.45.67.89

Enter domain (or press Enter to use IP):
```

#### Option A: Use a Domain (Recommended)

**Enter your domain**: `panel.yourdomain.com`

The script will:
1. Validate the domain format
2. Show your server IP
3. Remind you to configure DNS
4. Wait for you to set up DNS before continuing

**Before continuing**, add an A record in your DNS:
```
Type: A
Name: panel (or your subdomain)
Value: 123.45.67.89 (your VPS IP)
TTL: 300
```

**Benefits:**
- ✅ Automatic HTTPS via Let's Encrypt
- ✅ Professional SSL certificate
- ✅ Secure access
- ✅ Custom branding

#### Option B: Use IP Address

**Press Enter** to use IP

The script will:
- Configure Caddy for HTTP only
- Use your server's IP address

**Note:** This is fine for testing but not recommended for production.

### Step 3: Automated Installation

The script automatically installs:

1. **System Updates** (~2 minutes)
2. **Docker** (~3 minutes)
3. **Go 1.23** (~1 minute)
4. **Node.js 20** (~2 minutes)
5. **SQLite** (~30 seconds)
6. **Caddy** (~1 minute)
7. **System Configuration** (~1 minute)

**Total time**: ~10-15 minutes

### Step 4: Deploy Application

After installation completes:

```bash
# Option 1: Clone from GitHub
cd /opt/vps-panel
sudo git clone https://github.com/yourusername/vps-panel.git .
sudo ./deploy.sh

# Option 2: Upload your code
scp -r vps-panel/* user@your-vps:/opt/vps-panel/
ssh user@your-vps
cd /opt/vps-panel
sudo ./deploy.sh
```

## 🎯 Post-Installation

### Access Your Panel

**With Domain:**
```
https://panel.yourdomain.com
```
- Automatic HTTPS ✅
- SSL Certificate ✅
- Secure connection ✅

**With IP:**
```
http://123.45.67.89
```
- HTTP only
- Not secure for production

### Create Admin Account

1. Open the panel URL in your browser
2. Click "Register"
3. Create your admin account
4. Start deploying projects!

### Deploy Your First App

1. **Connect Git Provider**
   - Go to Settings → Git Providers
   - Connect GitHub/GitLab/Gitea

2. **Create New Project**
   - Click "New Project"
   - Select a repository
   - Auto-detection will configure:
     - Framework (SvelteKit, Next.js, React, etc.)
     - Build commands
     - Output directory

3. **Deploy**
   - Click "Create & Deploy"
   - Watch real-time build logs
   - Your app is live!

4. **Add Custom Domain** (Optional)
   - Open your project
   - Go to Domains tab
   - Add your domain
   - Configure DNS A record
   - SSL certificate auto-generated

## 🔧 Management Commands

### Service Control

```bash
# Check status
sudo systemctl status vps-panel

# Start/stop/restart
sudo systemctl start vps-panel
sudo systemctl stop vps-panel
sudo systemctl restart vps-panel

# View logs
sudo tail -f /var/log/vps-panel/panel.log

# View systemd logs
sudo journalctl -u vps-panel -f
```

### Caddy Management

```bash
# Check status
sudo systemctl status caddy

# Reload configuration
sudo systemctl reload caddy

# View Caddyfile
sudo cat /etc/caddy/Caddyfile

# Check SSL certificates
sudo caddy list-modules
```

### Update Panel

```bash
cd /opt/vps-panel
sudo git pull
sudo ./deploy.sh
```

## 🆘 Troubleshooting

### DNS Not Resolving

**Problem:** Can't access panel via domain

**Solution:**
```bash
# Check DNS propagation
nslookup panel.yourdomain.com

# Verify Caddy configuration
sudo caddy validate --config /etc/caddy/Caddyfile

# Check Caddy logs
sudo journalctl -u caddy -f
```

### Service Won't Start

**Problem:** VPS Panel service fails to start

**Solution:**
```bash
# Check logs
sudo journalctl -u vps-panel -n 50

# Verify binary exists
ls -lh /opt/vps-panel/vps-panel

# Check permissions
sudo chown -R vps-panel:vps-panel /opt/vps-panel

# Rebuild
cd /opt/vps-panel
sudo ./deploy.sh
```

### SSL Certificate Issues

**Problem:** Can't get HTTPS certificate

**Causes:**
- DNS not propagated (wait 5-10 minutes)
- Firewall blocking port 443
- Domain not pointing to correct IP

**Solution:**
```bash
# Check firewall
sudo ufw status

# Open HTTPS port
sudo ufw allow 443/tcp

# Check Caddy logs
sudo journalctl -u caddy -n 100

# Force certificate renewal
sudo caddy reload --config /etc/caddy/Caddyfile
```

### Docker Issues

**Problem:** Deployments fail

**Solution:**
```bash
# Check Docker service
sudo systemctl status docker

# Test Docker
sudo docker run hello-world

# Check vps-panel user in docker group
groups vps-panel

# Add user to docker group
sudo usermod -aG docker vps-panel
sudo systemctl restart vps-panel
```

## 📞 Getting Help

### Logs to Check

1. **Application Logs**
   ```bash
   sudo tail -f /var/log/vps-panel/panel.log
   sudo tail -f /var/log/vps-panel/panel-error.log
   ```

2. **Systemd Logs**
   ```bash
   sudo journalctl -u vps-panel -f
   sudo journalctl -u caddy -f
   sudo journalctl -u docker -f
   ```

3. **Docker Logs**
   ```bash
   sudo docker logs <container-name>
   ```

### Configuration Files

- Panel Config: `/var/lib/vps-panel/config.json`
- Caddy Config: `/etc/caddy/Caddyfile`
- Systemd Service: `/etc/systemd/system/vps-panel.service`

### System Information

```bash
# Check installation
sudo systemctl status vps-panel caddy docker

# Check versions
docker --version
go version
node --version

# Check disk space
df -h

# Check memory
free -h
```

## 🎉 Success!

You're now ready to deploy applications like Vercel/Netlify, but on your own infrastructure!

**Next Steps:**
1. Secure your admin account
2. Connect your Git providers
3. Deploy your first project
4. Configure webhooks for auto-deployment
5. Add custom domains for your apps

---

**Need help?** Check [INSTALLATION.md](INSTALLATION.md) for detailed documentation.
