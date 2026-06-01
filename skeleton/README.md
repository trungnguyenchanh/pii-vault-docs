# Skeleton code — Trang Admin theo dõi & phân tích Log PII

Bộ mã khởi đầu (skeleton) thực tế cho trang admin, gồm ba phần:

```
skeleton/
├── db/          script SQL tạo bảng/index/view cho admin
├── api-go/      Go Admin API (router, middleware, handler, repo, self-audit)
└── web-react/   React SPA (Vite + TS, router, API client, trang F1–F4)
```

> **Đây là skeleton, không phải bản chạy đầy đủ.** Các chỗ cần hiện thực được đánh dấu
> `TODO(<mã task>)`. Handler API trả `501 Not Implemented` ở phần nghiệp vụ thật, để
> đội dev điền truy vấn DB và logic. Frontend gọi API và hiển thị trạng thái rỗng/lỗi gọn gàng.

## 1. Database

```bash
psql "$DATABASE_URL" -f skeleton/db/001_admin_schema.sql
```

Tạo các bảng `alert`, `incident`, `approval_request`, `admin_audit`, các index lọc log
và materialized view `mv_access_hourly` cho dashboard. Yêu cầu đã có bảng `pii_audit` (M6).

## 2. Go Admin API

```bash
cd skeleton/api-go
cp .env.example .env        # điền DATABASE_URL; để JWKS_URL trống để chạy dev mode
go mod tidy
go run ./cmd/adminapi       # http://localhost:8080
# Kiểm tra: curl localhost:8080/healthz
```

Ở dev mode (JWKS_URL trống), API chấp nhận header `X-Debug-User` / `X-Debug-Roles`
thay cho JWT — **chỉ dùng khi phát triển cục bộ**.

## 3. React Admin Web

```bash
cd skeleton/web-react
cp .env.example .env.local
npm install
npm run dev                 # http://localhost:5173 (proxy /api -> :8080)
```

## Thứ tự gợi ý khi hiện thực

1. **DB**: chạy script, kiểm tra index/materialized view.
2. **API repo**: điền các hàm trong `internal/repo/repo.go` (đánh dấu `TODO`).
3. **API auth**: hiện thực xác minh JWT/JWKS trong `internal/auth/auth.go`.
4. **API self-audit**: nối `internal/audit` vào INSERT bảng `admin_audit`.
5. **FE**: nối OIDC thật trong `src/api/client.ts`, bỏ header debug.

## Nguyên tắc bảo mật (không được vi phạm)

- Không endpoint/màn hình nào trả về **giá trị PII thật** — chỉ `pii_ref`, tên trường, mục đích.
- RBAC ở frontend chỉ để ẩn/hiện; **quyền thật do API quyết định**.
- Mọi thao tác ghi của admin phải sinh **self-audit**.
- Phê duyệt bốn mắt: người duyệt phải **khác** người yêu cầu (API kiểm tra, FE chỉ hỗ trợ UX).
- Bảng `admin_audit` và `pii_audit` là **append-only** (thu hồi UPDATE/DELETE ở DB).
