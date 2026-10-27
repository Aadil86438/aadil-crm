#!/bin/bash
# ==============================================================================
# GCP VM (Ubuntu/Debian) — 1-CLICK SETUP SCRIPT FOR ADMIN MICROSERVICE
# ==============================================================================
# Run this on your GCP VM after SSH login:
#   chmod +x setup-gcp-vm.sh && ./setup-gcp-vm.sh
# ==============================================================================

set -e

echo "🚀 [1/4] Updating package lists..."
sudo apt-get update -y
sudo apt-get install -y ca-certificates curl gnupg lsb-release git

echo "🐳 [2/4] Installing Docker & Docker Compose plugin..."
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

echo "⚙️ [3/4] Preparing environment file (.env)..."
if [ ! -f ".env" ]; then
  if [ -f ".env.example" ]; then
    cp .env.example .env
    echo "⚠️ Created .env from .env.example. Please update your AWS IP and INTERNAL_SERVICE_KEY in .env!"
  fi
fi

echo "🚀 [4/4] Starting admin-service-gcp containers..."
sudo docker compose down --remove-orphans || true
sudo docker compose up -d --build

echo ""
echo "✅ SUCCESS! admin-service-gcp is now running on GCP!"
echo "📡 Health Check: curl http://localhost:8081/health"
sudo docker compose ps
