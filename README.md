# test-k8s — branch `multi-polyglot-full`

Pilot **L4B polyglot full** — React + Go + Node + .NET + Python trong 1 repo.

| Service | Stack | Build | Public |
|---------|-------|-------|--------|
| `web` | React (Vite) | Dockerfile | ✓ `/` |
| `api` | Go | Dockerfile | ✓ `/api` |
| `node` | Node.js | Buildpack | internal |
| `dotnet` | .NET 8 | Dockerfile | internal |
| `worker` | Python | Buildpack | internal |

## Gọi thử

- Web: `GET /api/fleet` — fleet view 5 service
- Web: `GET /api/polyglot` — Go gateway gọi Node + .NET + Python worker

## Console

1. Branch `multi-polyglot-full`
2. Sync `.platform/services.yaml`
3. Sync workflow GitHub → push

## Local

```bash
# API
cd backend && APP_GREETING=local go run ./cmd/server

# Node
cd backend-node && PORT=8081 node index.js

# React
cd frontend && npm install && npm run dev
```

Branch khác: `multi-polyglot` (Go+Python), `multi-n-service`, `multi-service`.
