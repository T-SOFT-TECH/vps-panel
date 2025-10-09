#!/bin/bash

echo "=== Setting up Ubuntu 24.04 for VPS Panel Development ==="

# Fix DNS in WSL2
echo ""
echo "Step 1: Fixing DNS configuration..."
sudo rm -f /etc/resolv.conf
echo "nameserver 8.8.8.8" | sudo tee /etc/resolv.conf
echo "nameserver 8.8.4.4" | sudo tee -a /etc/resolv.conf
sudo chattr +i /etc/resolv.conf  # Make it immutable

# Test DNS
echo ""
echo "Testing DNS resolution..."
nslookup google.com

# Update package lists
echo ""
echo "Step 2: Updating package lists..."
sudo apt update

# Install Docker
echo ""
echo "Step 3: Installing Docker..."
sudo apt install -y ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Set up the Docker repository
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker Engine
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# Start Docker service
sudo service docker start

# Add current user to docker group
sudo usermod -aG docker $USER

# Install Go
echo ""
echo "Step 4: Installing Go..."
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
rm go1.23.0.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Install Node.js (using NodeSource)
echo ""
echo "Step 5: Installing Node.js 20..."
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# Install build essentials
echo ""
echo "Step 6: Installing build tools..."
sudo apt install -y build-essential git sqlite3

echo ""
echo "=== Setup Complete! ==="
echo ""
echo "Installed versions:"
docker --version
go version
node --version
npm --version

echo ""
echo "Next steps:"
echo "1. Exit this terminal and open a new one (to apply group changes)"
echo "2. Run: cd /mnt/c/Users/TSOFT/Documents/Websites/vps-panel"
echo "3. Build and test the application"
