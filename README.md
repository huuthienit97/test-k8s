# test-k8s — branch `feature/minio-files`

Monorepo demo: **api + web** + upload/list/xóa file qua **MinIO addon** (`S3_*`).

## Cấu trúc

```
backend/     → image …/api  (Go + AWS SDK S3)
frontend/    → image …/web  (static UI upload/list/delete)
.platform/   → runtime contract gồm S3_*
```

## API

| Method | Path | Mô tả |
|--------|------|--------|
| GET | `/api/health` | Health + `s3_ready` |
| GET | `/api/files` | List object trong bucket |
| POST | `/api/files` | Upload multipart field `file` (+ optional `key`) |
| DELETE | `/api/files?key=` | Xóa object |

## Console

1. Project bật **MinIO** addon (dev) → inject `S3_*`
2. Deploy branch **`feature/minio-files`**
3. Mở web → Upload / Xóa file

## Local

```bash
# Cần MinIO local hoặc copy external JSON từ Console
export APP_GREETING=hello
export S3_ENDPOINT=http://127.0.0.1:9000
export S3_ACCESS_KEY=...
export S3_SECRET_KEY=...
export S3_BUCKET=app
export S3_FORCE_PATH_STYLE=true

cd backend && go run ./cmd/server
```
