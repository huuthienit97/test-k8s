# test-k8s — branch `multi-service`

Monorepo pilot **Giai đoạn 2**: 1 repo → 2 image (`api` + `web`) → 2 Deployment → Ingress `/api` + `/`.

## Cấu trúc

```
backend/     → image research-labs/api
frontend/    → image research-labs/web
.platform/    → env contract (BUILD_LABEL, APP_GREETING)
```

## Console (research-labs)

1. Tab **Deploy / Git** → **Multi-service (api + web)** → Lưu (template `backend/` + `frontend/`)
2. Branch = `multi-service`
3. **Kết nối repo & bật auto-deploy** (sync workflow mới — 2 bước build)
4. Push branch này

Ingress: `https://<domain-dev>/` → web, `https://<domain-dev>/api/health` → api.

## Env contract

| Biến | Scope |
|------|--------|
| `BUILD_LABEL` | Build (`.platform/build.yaml`) |
| `APP_GREETING` | Runtime (`.platform/runtime.yaml`) |

## Local

```bash
# API
cd backend && APP_GREETING=hello-local BUILD_LABEL=local go run ./cmd/server

# Web (cần proxy /api → localhost:8080 hoặc test sau deploy)
cd frontend && docker build -t test-web . && docker run -p 8081:8080 test-web
```

Các branch khác: `main` (single Go), `buildpack-node`, `buildpack-python`.
