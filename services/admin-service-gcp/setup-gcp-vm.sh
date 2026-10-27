#!/bin/bash
# ==============================================================================
# GCP VM (Ubuntu/Debian) — 1-CLICK SETUP SCRIPT FOR ADMIN MICROSERVICE
# ==============================================================================
# Run this on your GCP VM after SSH login:
#   chmod +x setup-gcp-vm.sh && ./setup-gcp-vm.sh
# ==============================================================================

set -e

echo "🧠 [1/5] Enabling 2GB Swap Memory (ensures smooth build on 1GB RAM e2-micro)..."
if [ ! -f "/swapfile" ]; then
  sudo fallocate -l 2G /swapfile || sudo dd if=/dev/zero of=/swapfile bs=1M count=2048
  sudo chmod 600 /swapfile
  sudo mkswap /swapfile
  sudo swapon /swapfile
  echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab || true
  echo "✅ 2GB Swap enabled successfully!"
else
  sudo swapon /swapfile 2>/dev/null || true
fi

echo "🚀 [2/5] Updating package lists..."
sudo apt-get update -y
sudo apt-get install -y ca-certificates curl gnupg lsb-release git

echo "🐳 [3/5] Installing Docker & Docker Compose plugin..."
if ! command -v docker &> /dev/null; then
  sudo mkdir -p /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg --yes || \
  curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg --yes
  
  echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$(lsb_release -is | tr '[:upper:]' '[:lower:]') \
    $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

  sudo apt-get update -y
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
  sudo usermod -aG docker $USER
fi

echo "⚙️ [4/5] Preparing environment file (.env)..."
if [ ! -f ".env" ]; then
  if [ -f ".env.example" ]; then
    cp .env.example .env
    echo "⚠️ Created .env from .env.example."
  fi
fi

echo "🚀 [5/5] Building and launching admin-service-gcp containers..."
sudo docker compose down --remove-orphans || true
sudo docker compose build --no-cache
sudo docker compose up -d

echo ""
echo "✅ SUCCESS! admin-service-gcp is now running on GCP!"
echo "📡 Health Check: curl http://localhost:8081/health"
sudo docker compose ps
