# 🚀 Microservices Architecture — Employee Management System

## Folder Structure
```
services/
├── core-service-aws/      ← Runs on AWS EC2 (Port 8080)
│   ├── handlers/          ← Auth, Users, Tasks, Activities, Notes, Dashboard
│   ├── repositories/      ← PostgreSQL queries
│   ├── middleware/        ← Auth, CORS, Logger, InternalServiceAuth ⭐
│   ├── migrations/        ← DB migration SQL files
│   └── main.go
│
├── admin-service-gcp/     ← Runs on GCP e2-micro VM (Port 8081)
│   ├── handlers/          ← Admin PIN, Redis Inspector, K8s, Log Viewer
│   │   └── core_proxy_handler.go  ← Calls AWS via HTTP ⭐
│   ├── coreclient/        ← Inter-service HTTP client ⭐
│   │   └── client.go      ← X-Internal-Service-Key Auth
│   ├── middleware/        ← Auth + CORS
│   ├── routes/            ← Admin routes + cross-cloud proxy routes
│   └── main.go
│
└── docker-compose.yml     ← Test both microservices locally
```

## Running Locally
```bash
cd services
docker-compose up --build
```
- **core-service-aws** → http://localhost:8080
- **admin-service-gcp** → http://localhost:8081

## How Inter-Service Communication Works
```
Browser → admin-service-gcp (GCP :8081)
              |
              | GET /api/admin/core/stats
              ↓
         admin-service-gcp sends:
         GET http://core-service-aws:8080/internal/stats
         Header: X-Internal-Service-Key: <shared_secret>
              |
              ↓
         core-service-aws (AWS :8080) validates the key
         Returns data from PostgreSQL
              |
              ↓
         admin-service-gcp forwards response to browser
```

## Key Environment Variables
| Variable | core-service-aws | admin-service-gcp |
|---|---|---|
| `PORT` | `8080` | `8081` |
| `INTERNAL_SERVICE_KEY` | `set_same_value` | `set_same_value` |
| `CORE_SERVICE_URL` | — | `http://<AWS-IP>:8080` |
| `JWT_SECRET` | `set_same_value` | `set_same_value` |
