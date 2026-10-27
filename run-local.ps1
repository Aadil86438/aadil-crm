param (
    [switch]$Check,
    [switch]$Stop,
    [switch]$Redis,
    [switch]$Docker
)

$RootDir = Split-Path -Parent $MyInvocation.MyCommand.Path

if ($Docker) {
    Write-Host "Launching complete stack using Docker Compose..." -ForegroundColor Cyan
    docker compose -f "$RootDir\docker-compose.local.yml" up --build
    exit 0
}

if ($Stop) {
    Write-Host "Stopping local Go and Node servers..." -ForegroundColor Red
    Get-Process -Name "main", "core-service-aws", "admin-service-gcp" -ErrorAction SilentlyContinue | Stop-Process -Force
    Write-Host "Stopped background Go processes." -ForegroundColor Green
    exit 0
}

if ($Check) {
    Write-Host ""
    Write-Host "Checking local services status..." -ForegroundColor Cyan
    
    # 1. PostgreSQL
    try {
        $pgService = Get-Service postgresql* -ErrorAction SilentlyContinue
        if ($pgService -and $pgService.Status -eq 'Running') {
            Write-Host "[OK] PostgreSQL Service is running." -ForegroundColor Green
        } else {
            Write-Host "[WARN] PostgreSQL Service is not running." -ForegroundColor Yellow
        }
    } catch {
        Write-Host "[WARN] Unable to check PostgreSQL service." -ForegroundColor Yellow
    }

    # 2. Core Service (Port 8080)
    try {
        $resCore = Invoke-RestMethod -Uri "http://localhost:8080/health" -TimeoutSec 3 -ErrorAction Stop
        Write-Host "[OK] Core Service AWS (:8080) is healthy." -ForegroundColor Green
    } catch {
        Write-Host "[OFFLINE] Core Service AWS (:8080) not responding." -ForegroundColor Red
    }

    # 3. Admin Service (Port 8081)
    try {
        $resAdmin = Invoke-RestMethod -Uri "http://localhost:8081/health" -TimeoutSec 3 -ErrorAction Stop
        Write-Host "[OK] Admin Service GCP (:8081) is healthy." -ForegroundColor Green
    } catch {
        Write-Host "[OFFLINE] Admin Service GCP (:8081) not responding." -ForegroundColor Red
    }

    # 4. Frontend (Port 3000)
    try {
        $resFront = Invoke-WebRequest -Uri "http://localhost:3000" -TimeoutSec 3 -ErrorAction Stop
        Write-Host "[OK] Frontend (:3000) is running." -ForegroundColor Green
    } catch {
        Write-Host "[OFFLINE] Frontend (:3000) not responding." -ForegroundColor Red
    }

    Write-Host ""
    exit 0
}

if ($Redis) {
    Write-Host "Starting local Redis container on port 6379..." -ForegroundColor Cyan
    docker run -d --name crm-redis -p 6379:6379 redis:7-alpine 2>$null
    if ($LASTEXITCODE -ne 0) {
        docker start crm-redis 2>$null
    }
    Write-Host "Redis container ready on port 6379." -ForegroundColor Green
    exit 0
}

Write-Host "========================================================" -ForegroundColor Cyan
Write-Host " Launching Zoho CRM Local Development Stack             " -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  * Core Service AWS  : http://localhost:8080" -ForegroundColor White
Write-Host "  * Admin Service GCP : http://localhost:8081" -ForegroundColor White
Write-Host "  * Frontend App      : http://localhost:3000" -ForegroundColor White
Write-Host ""

# Verify PostgreSQL
$pg = Get-Service postgresql* -ErrorAction SilentlyContinue | Where-Object { $_.Status -eq 'Running' }
if (-not $pg) {
    Write-Host "Warning: PostgreSQL service is not running. Please start postgresql service." -ForegroundColor Yellow
}

# 1. Start Core Service AWS
Write-Host "Starting [1/3] Core Service AWS on port 8080..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$RootDir\services\core-service-aws'; Write-Host 'Running Core Service AWS on http://localhost:8080...' -ForegroundColor Cyan; go run main.go"

Start-Sleep -Seconds 2

# 2. Start Admin Service GCP
Write-Host "Starting [2/3] Admin Service GCP on port 8081..." -ForegroundColor Green
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$RootDir\services\admin-service-gcp'; Write-Host 'Running Admin Service GCP on http://localhost:8081...' -ForegroundColor Green; go run main.go"

Start-Sleep -Seconds 1

# 3. Start Frontend
Write-Host "Starting [3/3] Frontend Dev Server on port 3000..." -ForegroundColor Yellow
Start-Process powershell -ArgumentList "-NoExit", "-Command", "Set-Location '$RootDir\frontend'; Write-Host 'Running Frontend on http://localhost:3000...' -ForegroundColor Yellow; npm run serve"

Write-Host ""
Write-Host "All services launched in separate windows!" -ForegroundColor Green
Write-Host "Open your browser to: http://localhost:3000" -ForegroundColor White
Write-Host "To check health, run: .\run-local.ps1 -Check" -ForegroundColor Gray
Write-Host ""
