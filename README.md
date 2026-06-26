# test-k8s — branch `buildpack-python`

App **Python (Flask + Gunicorn)** mẫu cho Buildpack (không có `Dockerfile`). Dùng với Platform Console project `research-labs`.

## Contract env

| File | Biến | Scope |
|------|------|--------|
| `.platform/build.yaml` | `BUILD_LABEL` | Khi build image |
| `.platform/runtime.yaml` | `APP_GREETING` | Khi Pod chạy |

## Test trên Console

1. **Deploy / Git** → Branch = `buildpack-python`
2. **Đồng bộ workflow GitHub** (badge **Buildpack**)
3. Push lên `buildpack-python` để chạy pipeline

Paketo nhận stack qua `requirements.txt` + `Procfile`.

## Local

```bash
python3 -m venv .venv && source .venv/bin/activate
pip install -r requirements.txt
export APP_GREETING=hello-local
export PORT=8080
gunicorn app:app --bind 0.0.0.0:$PORT
```

## Endpoints

- `GET /health` — JSON (`stack`: `python`)
- `GET /` — plain text

Các branch buildpack khác: `buildpack-test` (Go), `buildpack-node` (Node), `main` (Docker + Go).
