# test-k8s — branch `multi-polyglot`

Pilot **L4B polyglot**: Go (Docker) api + nginx web + **Python buildpack** worker internal.

| Service | Stack | Build |
|---------|-------|-------|
| api | Go | Dockerfile `backend/` |
| web | nginx | Dockerfile `frontend/` |
| worker | Python | Buildpack `worker/` (Procfile + requirements.txt) |

Branch trước: `multi-n-service` (worker Go Docker). Branch này đổi worker → Python buildpack.

Console: sync `.platform/services.yaml` → sync workflow → push.

## Cấu trúc

```
backend/     → api (public, Ingress /api)
frontend/    → web (public, Ingress /)
worker/      → worker (internal — ping API qua SVC_API_URL)
.platform/   → env contract
```

## Console (research-labs)

1. Layout **Multi-service** → 3 service: api, web, worker (`expose_ingress=false`)
2. Branch = `multi-n-service`
3. Sync workflow GitHub (3 bước build)
4. Push branch này

Worker: `GET /health`, `GET /status` (log ping tới api nội bộ).

## Env contract

| Biến | Scope |
|------|--------|
| `BUILD_LABEL` | Build |
| `APP_GREETING` | Runtime (api) |
| `SVC_API_URL` | Auto-inject platform (worker) |

Các branch khác: `multi-service` (2 service), `main` (single).
