---
title: 3. Đặc tả nghiệp vụ chức năng
description: Tám module chức năng và các luồng nghiệp vụ end-to-end của hệ thống PII.
---

## Bản đồ module chức năng

| Mã | Module | Vai trò chính | Giai đoạn |
|---|---|---|---|
| M1 | Quản lý đối tượng dữ liệu | Tạo/quản lý subject & pii_ref | GĐ 1, 3 |
| M2 | Cổng truy cập PII | Điểm vào duy nhất đọc/ghi PII | GĐ 2, 3 |
| M3 | Mã hóa & quản lý khóa | Mã hóa cột, blind index, KMS | GĐ 3 |
| M4 | Masking động | Che dữ liệu theo vai trò | GĐ 3 |
| M5 | Kiểm soát truy cập | RBAC, RLS, purpose, bốn mắt | GĐ 4 |
| M6 | Nhật ký audit | Ghi vết bất biến mọi truy cập | GĐ 2 |
| M7 | Giám sát & cảnh báo | Phát hiện bất thường realtime | GĐ 5 |
| M8 | Tuân thủ & DSAR | Quyền chủ thể, bằng chứng | GĐ 6 |

## M1 — Quản lý đối tượng dữ liệu

**Mục đích:** quản lý vòng đời mỗi chủ thể dữ liệu và tham chiếu `pii_ref` dùng để thay thế PII trong ứng dụng.

**Chức năng:** tạo subject, liên kết tham chiếu, cập nhật PII, hợp nhất/tách, vô hiệu hóa/xóa (crypto-shred).

**Quy tắc nghiệp vụ:**
1. Mỗi subject có đúng một `pii_ref` ổn định; ứng dụng chỉ lưu `pii_ref`.
2. Mọi thao tác phải đi qua cổng M2 và ghi audit M6.
3. Xóa thực hiện bằng hủy khóa mã hóa (ciphertext vĩnh viễn vô nghĩa).

## M2 — Cổng truy cập PII

**Mục đích:** điểm vào duy nhất để đọc/ghi PII, thực thi chính sách bảo mật, mục đích và audit nhất quán.

**Chức năng:** Reveal (đọc), Store (ghi), Update, Lookup theo blind index, xác thực gọi.

```text
RevealField(pii_ref, field, purpose) -> { value | masked }
StoreSubject(pii_payload)            -> { pii_ref }
UpdateSubject(pii_ref, patch)        -> { ok }
LookupByIndex(field, value)          -> { pii_ref | null }
```

**Quy tắc:** mọi yêu cầu kèm mục đích hợp lệ; thiếu mục đích → từ chối. Mỗi lần gọi tạo đúng một bản ghi audit. Không bao giờ trả PII ra log/lỗi/URL.

## M3 — Mã hóa & Quản lý khóa

**Mục đích:** mã hóa giá trị PII, sinh blind index, quản lý vòng đời khóa qua KMS/HSM.

**Quy tắc:** khóa không nằm cùng nơi với dữ liệu và không thuộc quyền DBA. Mỗi subject/field dùng DEK riêng. Blind index chỉ phục vụ so khớp chính xác.

## M4 — Masking động

**Mục đích:** che một phần/toàn bộ PII theo vai trò khi trả về.

**Quy tắc:** áp sau khi M5 xác định vai trò, trước khi M2 trả kết quả. Vai trò mặc định nhận giá trị ẩn hoàn toàn (an toàn mặc định).

## M5 — Kiểm soát truy cập

**Mục đích:** quyết định ai được làm gì với dữ liệu nào, dựa trên vai trò, phạm vi dòng, mục đích, phê duyệt.

**Chức năng:** RBAC, RLS, kiểm tra mục đích, phê duyệt bốn mắt, quản lý vòng đời quyền (SSO/IdP).

**Quy tắc:** mặc định từ chối. Thao tác hàng loạt/nhạy cảm cần phê duyệt bốn mắt. Thu hồi quyền có hiệu lực tức thì.

## M6 — Nhật ký Audit bất biến

**Mục đích:** ghi bất biến mọi truy cập/thao tác trên PII, nền tảng cho truy nguồn và tuân thủ. **Module có giá trị cao nhất cho mục tiêu chống rò rỉ.**

**Quy tắc:** append-only, không sửa/xóa; hash-chaining; tách khỏi DB nghiệp vụ; không chứa PII thật, chỉ chứa `pii_ref` và tên trường.

## M7 — Giám sát & Phát hiện bất thường

**Mục đích:** phân tích audit realtime, phát hiện hành vi bất thường và cảnh báo.

**Quy tắc:** cảnh báo kèm đủ ngữ cảnh (actor, thời điểm, mục đích, số bản ghi). Ngưỡng tinh chỉnh theo baseline để giảm nhiễu.

## M8 — Tuân thủ & Quyền chủ thể (DSAR)

**Mục đích:** thực thi quyền chủ thể theo Nghị định 13/2023 và tổng hợp bằng chứng tuân thủ.

**Quy tắc:** xóa qua crypto-shred (M3) và ghi audit (M6). Thao tác DSAR nhạy cảm cần phê duyệt bốn mắt.

## Các luồng nghiệp vụ chính

### Luồng đọc PII có kiểm soát
1. Gọi cổng M2 với `pii_ref` + danh tính + mục đích.
2. M2 xác thực; M5 kiểm tra quyền, phạm vi, mục đích.
3. Nếu rủi ro cao → M5 yêu cầu phê duyệt bốn mắt.
4. M3 giải mã; M4 áp masking theo vai trò.
5. M2 trả kết quả; M6 ghi audit bất biến.
6. M7 phân tích; nếu bất thường → cảnh báo.

### Luồng xử lý yêu cầu xóa (DSAR)
1. M8 tiếp nhận, xác minh danh tính người yêu cầu.
2. M5 yêu cầu phê duyệt bốn mắt.
3. M3 thực hiện crypto-shred (hủy khóa).
4. M6 ghi audit; M8 lưu bằng chứng đã xử lý.

### Luồng điều tra sự cố
1. M7 phát cảnh báo (ví dụ đọc bất thường hàng trăm hồ sơ).
2. Đội ứng cứu truy vấn M6 theo actor/thời gian.
3. M5 thu hồi quyền tác nhân nghi vấn để khoanh vùng.
4. M8 tổng hợp hồ sơ và thực hiện nghĩa vụ thông báo nếu cần.

:::tip[Giá trị cốt lõi]
Nhờ M6 (audit bất biến) và M7 (giám sát), câu hỏi "rò rỉ từ đâu?" có câu trả lời với actor, thời điểm và mục đích — thay vì "không biết" như hiện trạng.
:::
