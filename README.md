# test-k8s

App mẫu test end-to-end với Platform Console (GitHub Actions → Harbor/GHCR → deploy K8s).

## Contract env (Phase 2)

Dev khai báo **key bắt buộc** trong repo; **giá trị** khai trên Console (không commit secret).

| File | Biến | Scope Console |
|------|------|----------------|
| `.platform/build.yaml` | `BUILD_LABEL` | **Khi build image** |
| `.platform/runtime.yaml` | `APP_GREETING` | **Khi app chạy (Pod)** |

Platform đọc contract từ GitHub → readiness banner → chặn pipeline nếu thiếu trên Console.

## Biến môi trường

| Biến | Bắt buộc | Nơi khai | Ghi chú |
|------|----------|----------|---------|
| `BUILD_LABEL` | Có | Console → build image | `ARG` Dockerfile, hiện trên `/health` |
| `APP_GREETING` | Có | Console → Pod | Runtime, hiện trên `/` |
| `PORT` | Không | Pod (tuỳ chọn) | Mặc định `8080` |
| `GIT_SHA`, `GIT_REF` | — | Platform tự inject | Mỗi lần build |

### Local (runtime only)

```bash
export APP_GREETING=hello-local
export BUILD_LABEL=local-build
go run ./cmd/server
```

```bash
docker build --build-arg BUILD_LABEL=local-docker --build-arg GIT_SHA=dev --build-arg GIT_REF=main -t test-k8s:local .
docker run --rm -p 8080:8080 -e APP_GREETING=hello-docker test-k8s:local
```

### Demo trên Platform (`research-labs`)

1. **Cấu hình app (dev)**  
   - Pod: `APP_GREETING`  
   - Build image: `BUILD_LABEL`  
   - **Đồng bộ workflow GitHub** sau khi đổi contract/Dockerfile

### Test Buildpack (không xóa `Dockerfile` trên `main`)

Platform quét **branch đang deploy**, không phải máy local.

1. Dùng branch [`buildpack-test`](https://github.com/huuthienit97/test-k8s/tree/buildpack-test) — trên branch này **không có** `Dockerfile` (file vẫn còn trên `main`).
2. Console → **Deploy / Git** → đổi **Branch** = `buildpack-test` → **Kết nối repo & bật auto-deploy** (hoặc sync workflow).
3. Badge phải hiện **Buildpack** (không phải Docker).
4. Push lên `buildpack-test` để chạy pipeline.

Quay lại test Docker: branch `main` + sync workflow lại.

2. **Thiếu env** → Actions fail ở bước *Kiểm tra cấu hình env* (422)

3. **Đủ env** → build → push Harbor → deploy → `/health` có `build_label`, `/` có greeting

## Endpoints

- `GET /health` → JSON: `status`, `version`, `git_sha`, `git_ref`, `build_label`
- `GET /` → plain text: `APP_GREETING`, `BUILD_LABEL`, git info
