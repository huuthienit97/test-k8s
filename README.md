# test-k8s

App mẫu tối giản để test end-to-end với Platform Console.

## Local run

```bash
go run ./cmd/server
```

Mặc định chạy ở `http://localhost:8080`.

## Endpoints

- `GET /health` -> `{\"status\":\"ok\"}`
- `GET /` -> plain text

## Container

```bash
docker build -t test-k8s:local .
docker run --rm -p 8080:8080 test-k8s:local
```
