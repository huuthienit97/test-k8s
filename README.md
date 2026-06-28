# test-k8s — branch `multi-n-service`

Monorepo pilot **Giai đoạn 3 (L3)**: 1 repo → **3 image** (`api` + `web` + `worker`) → worker **internal** (không Ingress), service discovery qua `SVC_API_URL`.

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
