# test-k8s — branch `buildpack-node`

App **Node.js** mẫu cho Buildpack (không có `Dockerfile`). Dùng với Platform Console project `research-labs`.

## Contract env

| File | Biến | Scope |
|------|------|--------|
| `.platform/build.yaml` | `BUILD_LABEL` | Khi build image |
| `.platform/runtime.yaml` | `APP_GREETING` | Khi Pod chạy |

## Test trên Console

1. **Deploy / Git** → Branch = `buildpack-node`
2. **Đồng bộ workflow GitHub** (badge phải hiện **Buildpack** / Node)
3. Push lên `buildpack-node` để chạy pipeline

## Local

```bash
export APP_GREETING=hello-local
export BUILD_LABEL=local
npm start
# curl http://localhost:8080/health
```

## Endpoints

- `GET /health` — JSON (`stack`: `node`)
- `GET /` — plain text

Các branch buildpack khác: `buildpack-test` (Go), `buildpack-python` (Python), `main` (Docker + Go).
