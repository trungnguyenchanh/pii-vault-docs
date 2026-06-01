---
title: 4. Thiết kế kỹ thuật chi tiết
description: Lược đồ dữ liệu, hợp đồng API và quyết định thiết kế cho tám module M1–M8.
---

## Kiến trúc tổng thể

### Ngăn xếp công nghệ tham chiếu

- **CSDL nghiệp vụ & vault:** PostgreSQL (TDE, pgcrypto, RLS gốc). Có thể thay bằng SQL Server (Always Encrypted) hoặc Oracle (TDE + Data Redaction).
- **Quản lý khóa:** HashiCorp Vault tự vận hành hoặc KMS/HSM cloud có region tại Việt Nam.
- **Cổng truy cập (M2):** service gRPC/REST stateless, scale ngang sau load balancer.
- **Audit & giám sát:** kho append-only riêng (PostgreSQL tách biệt hoặc OpenSearch); SIEM.
- **Định danh:** IdP/SSO phát JWT cho người dùng, mTLS cho service-to-service.

### Nguyên tắc xuyên suốt

- Mặc định từ chối (default-deny).
- Tách biệt trách nhiệm: dữ liệu, khóa, audit ở ba ranh giới tin cậy khác nhau.
- Một điểm vào: mọi truy cập PII qua cổng M2.
- Ghi vết toàn bộ: không thao tác PII nào không sinh audit.
- Fail-closed cho bảo mật.

```text
[Ứng dụng: Web/App/CRM/OMS]
        | (chỉ lưu pii_ref)
        v   JWT/mTLS + purpose
[M2 Cổng truy cập PII] --(kiểm tra)--> [M5 Access Control]
        | reveal/store
        v
[M3 Mã hóa & KMS] <---> [Vault: khóa]
        |  ciphertext + blind index
        v
[CSDL Vault: TDE]      [M4 Masking áp khi trả về]
        +--> [M6 Audit log bất biến] --> [M7 Giám sát/Cảnh báo]
        +--> [M8 Tuân thủ/DSAR]
```

## M1 — Subject Management

```sql
TABLE subject (
  pii_ref      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  status       TEXT NOT NULL DEFAULT 'active',  -- active|merged|shredded
  merged_into  UUID NULL REFERENCES subject(pii_ref),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

TABLE subject_field (
  pii_ref      UUID REFERENCES subject(pii_ref),
  field        TEXT NOT NULL,        -- phone|email|address|fullname
  value_enc    BYTEA NOT NULL,       -- AES-256-GCM ciphertext
  value_bidx   TEXT NULL,            -- HMAC blind index
  dek_id       TEXT NOT NULL,        -- tham chiếu khóa trong Vault
  PRIMARY KEY (pii_ref, field)
);
CREATE INDEX ON subject_field(value_bidx);
```

**Quyết định:** `pii_ref` dùng UUIDv4 (không đoán được). Tách `subject_field` theo dòng để mỗi trường dùng DEK riêng và xoay khóa độc lập.

## M2 — PII Access Gateway

```text
AuthContext { identity, roles[], auth_method(JWT|mTLS) }
RevealField(pii_ref, field, purpose) -> { value | masked_value, audit_id }
StoreSubject(fields{}, purpose)      -> { pii_ref, audit_id }
UpdateSubject(pii_ref, patch{}, purpose) -> { ok, audit_id }
LookupByIndex(field, value, purpose) -> { pii_ref | null, audit_id }
BulkReveal(pii_refs[], field, purpose)  -> { request_id, PENDING_APPROVAL }
```

**Trình tự RevealField:** xác thực → M5.Authorize → (DENY/PENDING/ALLOW) → M3.Decrypt → M4.Mask → M6 ghi audit → trả kết quả.

:::tip[Hiệu năng]
Cache quyết định phân quyền (M5) theo (role, action, field); KHÔNG cache plaintext PII. Mục tiêu p99 độ trễ cổng < vài chục ms.
:::

## M3 — Encryption & Key Management

```text
KEK (trong Vault/HSM, không xuất ra)
  └─ wrap/unwrap ─> DEK (theo subject/field)
                      └─ AES-256-GCM mã hóa value
value_enc    = AES_GCM(DEK, nonce, value)
value_bidx   = HMAC-SHA256(IndexKey, normalize(value))
```

**Quyết định:** AES-256-GCM (AEAD), nonce duy nhất mỗi lần. DEK theo subject/field thu hẹp blast radius. Xoay khóa chỉ cần re-wrap DEK (không mã hóa lại toàn bộ).

:::caution[Cảnh báo bảo mật]
Blind index để lộ việc hai bản ghi có cùng giá trị. Chỉ tạo cho trường cần tra cứu; cân nhắc thêm muối theo trường để giảm rủi ro phân tích tần suất.
:::

## M4 — Dynamic Masking

```sql
TABLE mask_policy (
  role TEXT, field TEXT,
  strategy TEXT,  -- FULL|PARTIAL|HIDE
  PRIMARY KEY (role, field)
);
-- Mặc định khi không có dòng: HIDE
-- phone 0901234567 -> 09****4567 ; email an@mail.com -> a***@mail.com
```

**Quyết định:** cấu hình bằng dữ liệu (không hard-code). Nhiều vai trò → áp chiến lược ít lộ nhất.

## M5 — Access Control

```sql
TABLE role_grant (role_id TEXT, field TEXT, action TEXT,
  PRIMARY KEY (role_id, field, action));
TABLE purpose_catalog (purpose TEXT PRIMARY KEY, active BOOL);
TABLE fourEyes_rule (action TEXT, condition TEXT);
```

**Thuật toán Authorize:** kiểm purpose hợp lệ → kiểm role_grant (default-deny) → áp RLS phạm vi dòng → đối chiếu four-eyes → ALLOW. Mọi nhánh trả lý do để M6 ghi audit.

## M6 — Immutable Audit Log

```sql
TABLE pii_audit (
  seq BIGSERIAL PRIMARY KEY,
  ts TIMESTAMPTZ DEFAULT now(),
  actor TEXT, action TEXT, subject_ref UUID,
  field TEXT, purpose TEXT, result TEXT,
  meta JSONB, prev_hash TEXT, row_hash TEXT NOT NULL
);
row_hash = SHA256(prev_hash || ts || actor || action ||
                  subject_ref || field || purpose || result)
```

**Cơ chế bất biến:** thu hồi UPDATE/DELETE; hash-chain mỗi INSERT; tách instance/credentials; job định kỳ `VerifyChain()`.

:::caution[Toàn vẹn]
Nếu ghi audit lỗi thì thao tác PII nên fail-closed (không cho qua mà không ghi vết).
:::

## M7 — Monitoring & Detection

| Luật | Điều kiện ví dụ | Mức |
|---|---|---|
| Đọc hàng loạt bất thường | > N subject/phút bởi 1 actor | Cao |
| Truy cập ngoài giờ | READ ngoài 7h–21h | Trung bình |
| Tỉ lệ DENY cao | > X% DENY trong cửa sổ | Trung bình |
| Truy cập rải rác | 1 actor chạm nhiều subject lạ | Cao |
| Xuất dữ liệu | EXPORT bất kỳ | Cao |

## M8 — Compliance & DSAR

```sql
TABLE consent (pii_ref UUID, purpose TEXT, granted BOOL, ts TIMESTAMPTZ,
  PRIMARY KEY(pii_ref, purpose));
TABLE dsar_request (request_id UUID PRIMARY KEY, pii_ref UUID,
  type TEXT, status TEXT, due_date DATE);
TABLE retention_policy (field TEXT, ttl_days INT);
```

**Quy trình ERASE:** tiếp nhận & xác minh → M5 phê duyệt bốn mắt → M1.ShredSubject → M3.ShredKey → đánh dấu shredded + M6 ghi audit → sinh xác nhận xóa.

## Phụ thuộc giữa các module

| Module | Phụ thuộc | Hiện thực ở GĐ |
|---|---|---|
| M6 Audit | Độc lập (nền tảng) | GĐ 2 — trước tiên |
| M2 Cổng | M5, M3, M6 | GĐ 2 khung → GĐ 3 đủ |
| M3 Mã hóa | Vault/KMS | GĐ 3 |
| M1 Subject | M3, M6 | GĐ 1 nền → GĐ 3 |
| M4 Masking | M5 | GĐ 3 |
| M5 Access | IdP/SSO, M6 | GĐ 4 |
| M7 Giám sát | M6 | GĐ 5 |
| M8 Tuân thủ | M1, M3, M5, M6 | GĐ 6 |
