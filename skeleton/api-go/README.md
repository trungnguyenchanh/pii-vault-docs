# PII Admin API (Go) — Skeleton

Khung Admin API cho trang theo dõi & phân tích log PII. **Đây là skeleton**: handler trả dữ liệu mẫu/`501 Not Implemented` ở chỗ cần hiện thực, để đội dev điền nghiệp vụ thật.

## Chạy thử

```bash
cp .env.example .env      # điền DATABASE_URL, JWKS_URL...
go mod tidy
go run ./cmd/adminapi
# API lắng nghe tại :8080
```

> Lưu ý: ở chế độ skeleton, middleware xác thực có thể chạy "dev mode" (bỏ qua xác minh JWKS) khi `JWKS_URL` rỗng — CHỈ dùng khi phát triển cục bộ.

## Cấu trúc

```
cmd/adminapi/main.go   khởi tạo, router, middleware
internal/config        nạp cấu hình từ env
internal/auth          JWT/JWKS middleware + RBAC
internal/handler       HTTP handlers (F1..F4)
internal/service       nghiệp vụ (gọi repo, dựng DTO)
internal/repo          truy vấn DB (pgx) — read-only audit
internal/audit         ghi self-audit (hash-chain)
internal/model         DTO/response structs
```

## Nhóm endpoint

- F1 `GET /api/v1/logs`, `GET /api/v1/logs/{seq}`, `POST /api/v1/logs/verify-chain`
- F2 `GET /api/v1/stats/access|top-actors|summary`
- F3 `GET/PATCH /api/v1/alerts`, `GET /api/v1/subjects/{ref}/timeline`, `POST /api/v1/incidents`, `POST /api/v1/actors/{id}/revoke`
- F4 `GET /api/v1/approvals`, `POST /api/v1/approvals/{id}/approve|reject`, `GET/PUT /api/v1/roles`

Mọi thao tác ghi đều gọi `audit.Record(...)` để self-audit.
