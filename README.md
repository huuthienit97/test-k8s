# test-k8s

App mẫu test end-to-end với Platform Console (GitHub Actions → GHCR → deploy K8s).

## Biến môi trường (demo `.env` không lên Git)

App **bắt buộc** `APP_GREETING` lúc chạy. Không có trong repo — khai báo trên Console.

| Biến | Bắt buộc | Ghi chú |
|------|----------|---------|
| `APP_GREETING` | Có | Câu chào hiển thị trên `/` |
| `PORT` | Không | Mặc định `8080` |

### Local

```bash
cp .env.example .env
export $(grep -v '^#' .env | xargs)
go run ./cmd/server
```

Hoặc:

```bash
APP_GREETING=hello-local go run ./cmd/server
```

### Demo trên Platform (deployGHCR)

1. **Push code** (chưa khai báo env trên Console)  
   → Build GitHub OK, deploy OK, **pod CrashLoopBackOff** (log: `thiếu biến môi trường APP_GREETING`).

2. **Console → Project → Env vars → Dev** → thêm:
   - Key: `APP_GREETING`
   - Value: `Xin chào từ Platform!` (tuỳ ý)

3. **Cách A:** Bấm lưu (tự sync + restart pod) — pod Running, mở domain thấy greeting.  
   **Cách B:** Re-run workflow GitHub Actions — deploy lại với env đã có → thành công.

## Endpoints

- `GET /health` → JSON `status`, `version`
- `GET /` → plain text + `APP_GREETING`

## Container

```bash
docker build -t test-k8s:local .
docker run --rm -p 8080:8080 -e APP_GREETING=hello-docker test-k8s:local
```
