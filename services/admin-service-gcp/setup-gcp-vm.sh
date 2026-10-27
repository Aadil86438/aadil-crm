#!/bin/bash
# ==============================================================================
# GCP VM (Ubuntu/Debian) — INSTANT SETUP SCRIPT FOR ADMIN MICROSERVICE
# ==============================================================================
# Uses pre-compiled Linux binary for 5-second instant launch (0% CPU stress)
# ==============================================================================

set -e

echo "🐳 [1/3] Ensuring Docker is installed..."
if ! command -v docker &> /dev/null; then
  sudo apt-get update -y
  sudo apt-get install -y ca-certificates curl gnupg lsb-release git
  sudo mkdir -p /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg --yes || true
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null || true
  sudo apt-get update -y || true
  sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin || true
  sudo usermod -aG docker $USER || true
fi

echo "⚙️ [2/3] Checking environment file (.env)..."
if [ ! -f ".env" ]; then
  if [ -f ".env.example" ]; then
    cp .env.example .env
  fi
fi

echo "🚀 [3/3] Launching pre-compiled admin-service-gcp and Redis..."
sudo docker compose down --remove-orphans || true
sudo docker compose up -d --build

echo ""
echo "✅ SUCCESS! admin-service-gcp is now running on GCP!"
echo "📡 Health Check: curl http://localhost:8081/health"
sudo docker compose ps
