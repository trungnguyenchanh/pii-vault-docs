---
title: 6. Trang Admin theo dõi & phân tích Log
description: Phân rã chi tiết xây dựng trang admin (React SPA + Go API) để xem log, dashboard, cảnh báo/điều tra và phê duyệt bốn mắt.
---

Trang Admin là giao diện vận hành cho lớp Audit (M6) và Giám sát (M7): xem log, phân tích thống kê, xử lý cảnh báo, điều tra sự cố, và quản lý quyền/phê duyệt bốn mắt. Stack: **React SPA + Go API**.

:::caution[Nguyên tắc cốt lõi]
Trang admin KHÔNG truy vấn trực tiếp PII thật và KHÔNG hiển thị giá trị PII. Nó chỉ làm việc với metadata truy cập (audit) và cảnh báo. Mọi thao tác admin đều được ghi self-audit.
:::

## Kiến trúc thành phần

```text
[React SPA Admin]  --HTTPS/JSON-->  [Go Admin API]
      |  (JWT từ IdP/SSO, role=admin/dpo/security)
      |                                   |
      |                                   +--> [M6 Audit store (read-only)]
      |                                   +--> [M7 Detection (alerts)]
      |                                   +--> [M5 Access Control (roles, 4-eyes)]
      v                                   +--> ghi self-audit mọi thao tác
[Recharts/ECharts]                        (Go API là tầng duy nhất chạm dữ liệu)
```

## Bốn nhóm chức năng

| Mã | Nhóm | Tóm tắt |
|---|---|---|
| F1 | Xem & tìm kiếm log | Bảng log, lọc, phân trang keyset, chi tiết, verify-chain |
| F2 | Dashboard thống kê | Biểu đồ access theo thời gian, top actor, tỉ lệ DENY |
| F3 | Cảnh báo & điều tra | Danh sách alert, timeline subject, gom incident, khoanh vùng |
| F4 | Quyền & bốn mắt | Quản lý vai trò, duyệt/từ chối yêu cầu bốn mắt |

## Phân rã cho Database (DB)

| Task | Nội dung | Ưu tiên |
|---|---|---|
| DB-ADM-01 | Index lọc log: (ts), (actor,ts), (subject_ref,ts), (result,ts) | P0 |
| DB-ADM-02 | Materialized view tổng hợp theo giờ/ngày cho dashboard | P1 |
| DB-ADM-03 | Bảng `alert` (M7) + index theo status/severity/ts | P0 |
| DB-ADM-04 | Bảng `incident` gom nhiều alert | P1 |
| DB-ADM-05 | Bảng `approval_request` cho luồng bốn mắt | P0 |
| DB-ADM-06 | Bảng `admin_audit` (self-audit) | P0 |
| DB-ADM-07 | Read-replica/credentials chỉ-đọc cho Admin API | P1 |
| DB-ADM-08 | Tối ưu keyset pagination ở > 100 triệu dòng | P1 |

```sql
TABLE alert (
  alert_id    UUID PRIMARY KEY,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  rule        TEXT NOT NULL,       -- BULK_READ|OFF_HOURS|...
  severity    TEXT NOT NULL,       -- LOW|MEDIUM|HIGH
  actor       TEXT NOT NULL,
  window_from TIMESTAMPTZ, window_to TIMESTAMPTZ,
  evidence    JSONB,               -- danh sách audit seq liên quan
  status      TEXT NOT NULL DEFAULT 'open',
  assignee    TEXT, incident_id UUID NULL
);
CREATE INDEX ON alert(status, severity, created_at DESC);
```

:::tip[Hiệu năng]
Dùng keyset pagination (`WHERE seq < :cursor ORDER BY seq DESC LIMIT n`) thay vì OFFSET. Dashboard đọc từ materialized view, không quét bảng log gốc mỗi lần.
:::

## Phân rã cho API (Go)

Gợi ý: chi-router hoặc gin, pgx/sqlc, JWT middleware xác thực qua JWKS. Go API là tầng duy nhất chạm dữ liệu.

| Task | Nội dung | Ưu tiên |
|---|---|---|
| API-ADM-01 | Khung service: router, middleware, cấu hình | P0 |
| API-ADM-02 | Middleware authn (JWT/JWKS) + RBAC admin | P0 |
| API-ADM-03 | `GET /logs`: lọc + keyset pagination | P0 |
| API-ADM-05 | `POST /logs/verify-chain` | P1 |
| API-ADM-06 | `GET /stats/*` cho dashboard | P0 |
| API-ADM-07 | `GET/PATCH /alerts` | P0 |
| API-ADM-08 | `GET /subjects/{ref}/timeline` | P0 |
| API-ADM-10 | `POST /actors/{id}/revoke` (gọi M5, cần 4-eyes) | P1 |
| API-ADM-11 | `GET/POST /approvals` (bốn mắt) | P0 |
| API-ADM-13 | Ghi self-audit mọi thao tác ghi | P0 |

```text
GET /api/v1/logs?actor=&result=&from=&to=&cursor=<seq>&limit=50
200 -> { items:[{seq,ts,actor,action,subject_ref,field,purpose,result}],
         next_cursor:<seq|null> }

POST /api/v1/logs/verify-chain { from_seq, to_seq }
200 -> { ok:true } | { ok:false, broken_at:<seq> }

GET /api/v1/stats/access?from=&to=&bucket=hour|day
GET /api/v1/alerts?status=&severity=&cursor=&limit=
PATCH /api/v1/alerts/{id} { status, assignee }
GET /api/v1/subjects/{ref}/timeline
POST /api/v1/approvals/{id}/approve   // approver != requester
```

Phân tầng mã nguồn gợi ý:

```text
/cmd/adminapi/main.go    -- khởi tạo, router
/internal/auth           -- JWT/JWKS middleware, RBAC
/internal/handler        -- HTTP handlers F1..F4
/internal/service        -- nghiệp vụ
/internal/repo           -- truy vấn DB (pgx), read-only audit
/internal/audit          -- self-audit (hash-chain)
/internal/model          -- DTO/response
```

## Phân rã cho Frontend (React)

Gợi ý: Vite + TypeScript, React Router, TanStack Query, Recharts, một UI lib (shadcn/ui hoặc Ant Design). Xác thực OIDC.

| Task | Nội dung | Ưu tiên |
|---|---|---|
| FE-ADM-01 | Khung dự án: Vite+TS, router, layout | P0 |
| FE-ADM-02 | Đăng nhập SSO/OIDC + refresh token | P0 |
| FE-ADM-03 | API client + xử lý 401/403 tập trung | P0 |
| FE-ADM-04 | Trang Logs: bảng + lọc + keyset pagination | P0 |
| FE-ADM-06 | Dashboard: biểu đồ access theo thời gian | P0 |
| FE-ADM-08 | Trang Alerts: danh sách + đổi trạng thái | P0 |
| FE-ADM-09 | Màn điều tra: timeline subject + incident | P1 |
| FE-ADM-11 | Trang Approvals: duyệt/từ chối bốn mắt | P0 |
| FE-ADM-13 | Phân quyền hiển thị theo vai trò (RBAC UI) | P0 |

**Bản đồ màn hình:**

| Màn hình | Đường dẫn | Thành phần chính |
|---|---|---|
| Dashboard | `/` | Thẻ tổng quan, biểu đồ access, top actor |
| Nhật ký | `/logs` | Bộ lọc, bảng log, drawer chi tiết |
| Cảnh báo | `/alerts` | Danh sách alert, badge severity |
| Điều tra | `/incidents/:id` | Timeline subject, danh sách alert |
| Phê duyệt | `/approvals` | Hàng chờ bốn mắt, nút duyệt/từ chối |
| Phân quyền | `/roles` | Ma trận role × field × action |

:::caution[Lưu ý then chốt]
- Không hiển thị PII thật — chỉ `pii_ref`, tên trường, mục đích.
- RBAC ở UI chỉ là lớp phụ; quyền thật do API quyết định (không tin client).
- Nút Approve bị vô hiệu nếu người duyệt trùng người yêu cầu (bốn mắt).
- Phân trang dùng keyset (`next_cursor`), không dùng OFFSET.
:::

## Skeleton code

Repo kèm sẵn bộ skeleton code thực tế trong thư mục `skeleton/`:

- `skeleton/api-go/` — khung Go Admin API (router, middleware, handler rỗng, repo, self-audit).
- `skeleton/web-react/` — khung React SPA (Vite + TS, router, API client, các trang F1–F4).
- `skeleton/db/` — script SQL tạo bảng/index/view.

Xem `skeleton/README.md` để chạy thử cục bộ.

## Thứ tự thực hiện theo sprint

| Sprint | Nội dung |
|---|---|
| S1 | DB index/view + khung Go API + authn/RBAC + khung React |
| S2 | F1 Xem/tìm log |
| S3 | F2 Dashboard |
| S4 | F3 Cảnh báo & điều tra |
| S5 | F4 Quyền & bốn mắt + self-audit |
| S6 | Hardening, export, realtime, kiểm thử & nghiệm thu |

## Tiêu chí nghiệm thu (DoD)

- Tìm kiếm log đúng theo mọi bộ lọc; keyset pagination ổn định ở quy mô lớn.
- `verify-chain` phát hiện đúng khi một bản ghi audit bị sửa.
- Dashboard khớp số liệu với truy vấn kiểm chứng trên DB.
- Cảnh báo đổi trạng thái, gom incident, dựng timeline subject được.
- Nút Approve bị chặn khi người duyệt trùng người yêu cầu.
- Mọi thao tác ghi của admin xuất hiện trong `admin_audit`.
- Không endpoint/màn hình nào lộ giá trị PII thật.
- API trả 403 khi vai trò không đủ quyền (kể cả khi UI đã ẩn nút).
