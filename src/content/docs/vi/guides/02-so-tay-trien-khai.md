---
title: 2. Sổ tay triển khai
description: Cách làm chi tiết từng giai đoạn — tác vụ, công cụ, ví dụ và tiêu chí hoàn thành.
---

Lộ trình chia thành sáu giai đoạn theo nguyên tắc "nhìn thấy trước — siết chặt sau". Mỗi giai đoạn tạo ra giá trị độc lập.

## Bản đồ lộ trình tổng thể

| GĐ | Tên giai đoạn | Đầu ra cốt lõi | Phụ thuộc |
|---|---|---|---|
| 1 | Khảo sát & lập bản đồ dữ liệu | Bản đồ PII + ma trận phân loại | Không |
| 2 | Lớp audit tập trung | Mọi truy cập PII có dấu vết bất biến | GĐ 1 |
| 3 | Mã hóa & masking | Dữ liệu vô nghĩa khi bị lấy trộm | GĐ 1 |
| 4 | Kiểm soát truy cập | Quyền tối thiểu, mặc định từ chối | GĐ 2, 3 |
| 5 | Giám sát & phát hiện | Cảnh báo realtime, diễn tập | GĐ 2, 4 |
| 6 | Tuân thủ & vận hành | Bằng chứng tuân thủ NĐ 13/2023 | GĐ 2–5 |

## Giai đoạn 1 — Khảo sát & lập bản đồ dữ liệu

**Mục tiêu:** biết chính xác PII đang nằm ở đâu trước khi bảo vệ.

1. Lập danh sách toàn bộ kho lưu trữ: database, file chia sẻ, log, backup, dữ liệu gửi đối tác.
2. Chạy quét phát hiện PII tự động để tìm trường chứa tên, SĐT, email, địa chỉ, số định danh.
3. Phân loại từng trường theo mức nhạy cảm và quyết định biện pháp: mã hóa, masking, hay cả hai.
4. Vẽ sơ đồ luồng dữ liệu: PII đi vào từ đâu, hệ thống nào đọc, chảy ra những điểm nào.
5. Đối chiếu nghĩa vụ theo Nghị định 13/2023.
6. Lập sổ đăng ký xử lý dữ liệu (data processing register).

```text
| Hệ thống | Bảng/Cột      | Loại PII | Mức       | Biện pháp        |
|----------|---------------|----------|-----------|------------------|
| CRM      | kh.sdt        | SĐT      | Nhạy cảm  | Mã hóa + masking |
| OMS      | don.dia_chi   | Địa chỉ  | Nhạy cảm  | Mã hóa           |
| CRM      | kh.email      | Email    | Nhạy cảm  | Mã hóa + index   |
```

:::tip[Hoàn thành khi]
Có bản đồ PII đầy đủ, ma trận phân loại, sơ đồ luồng dữ liệu và sổ đăng ký xử lý — được nghiệp vụ và kỹ thuật cùng xác nhận.
:::

## Giai đoạn 2 — Lớp Audit tập trung (ưu tiên cao nhất)

**Mục tiêu:** đạt khả năng nhìn thấy ngay lập tức. Đây là giai đoạn mang lại giá trị lớn nhất cho mục tiêu chống rò rỉ.

1. Thiết kế lược đồ bản ghi audit: thời gian, danh tính, hành động, định danh bản ghi, mục đích, kết quả.
2. Chọn cơ chế chống sửa đổi: append-only + hash-chaining.
3. Đặt nơi lưu audit tách biệt với DB nghiệp vụ.
4. Tích hợp điểm ghi audit tại tầng ứng dụng / data access layer.
5. Gắn việc khai mục đích vào yêu cầu đọc.
6. Dựng dashboard quan sát và cảnh báo cơ bản.
7. Kiểm chứng tính toàn vẹn (chain verify).

```sql
CREATE TABLE pii_audit (
  id           BIGSERIAL PRIMARY KEY,
  ts           TIMESTAMPTZ NOT NULL DEFAULT now(),
  actor        TEXT NOT NULL,
  action       TEXT NOT NULL,
  subject_ref  TEXT NOT NULL,
  field        TEXT,
  purpose      TEXT NOT NULL,
  result       TEXT NOT NULL,
  prev_hash    TEXT,
  row_hash     TEXT NOT NULL
);
-- chỉ cho phép INSERT; thu hồi UPDATE/DELETE với mọi role app
```

:::tip[Hoàn thành khi]
Mọi đường đọc PII chính đều ghi audit; chuỗi hash xác minh được; dashboard hiển thị truy cập realtime và phát cảnh báo khi bất thường.
:::

## Giai đoạn 3 — Mã hóa & Masking ở lớp dữ liệu

**Mục tiêu:** khiến dữ liệu vô nghĩa nếu bị lấy trái phép, và hạn chế nhân sự nhìn thấy PII đầy đủ.

1. Bật TDE để mã hóa toàn bộ dữ liệu ở mức đĩa.
2. Triển khai quản lý khóa chuyên dụng (KMS/HSM hoặc HashiCorp Vault), tách khỏi dữ liệu và quyền DBA.
3. Áp dụng mã hóa cấp cột cho trường nhạy cảm; dùng blind index cho trường cần tra cứu chính xác.
4. Định nghĩa chính sách Dynamic Data Masking theo vai trò.
5. Loại bỏ PII khỏi log, thông báo lỗi, URL.
6. Thử nghiệm trên staging, đo hiệu năng, rồi triển khai production theo nhóm bảng.

```sql
-- Lưu giá trị mã hóa + cột băm có khóa để tra cứu chính xác
phone_enc    BYTEA   -- ciphertext
phone_bidx   TEXT    -- = HMAC(key, normalize(phone))
-- Khi tìm theo SĐT:
WHERE phone_bidx = hmac_lookup('0901234567')
```

:::caution[Giới hạn cần biết]
Blind index chỉ hỗ trợ so khớp chính xác, không làm được tìm kiếm mờ hay khoảng giá trị. Nếu cần tìm kiếm phức tạp trên dữ liệu mã hóa, đó là lúc cân nhắc searchable encryption / vault đầy đủ.
:::

## Giai đoạn 4 — Kiểm soát truy cập

**Mục tiêu:** chuyển từ "ai cũng đọc được" sang "quyền tối thiểu, mặc định từ chối".

1. Thiết kế mô hình RBAC dựa trên vai trò công việc thực tế.
2. Áp nguyên tắc least-privilege và mặc định từ chối.
3. Bật Row-Level Security giới hạn phạm vi dòng.
4. Bắt buộc khai mục đích cho mọi truy cập PII.
5. Áp phê duyệt "bốn mắt" cho thao tác rủi ro cao.
6. Thu hồi các quyền rộng đang tồn tại.

| Vai trò | Trường được đọc | Phạm vi (RLS) | Bốn mắt? |
|---|---|---|---|
| CSKH | Tên, SĐT (che một phần) | Khách được phân công | Khi đọc hàng loạt |
| Vận hành đơn | Tên, địa chỉ, SĐT | Đơn trong ca | Khi xuất file |
| Marketing | Chỉ blind index / phân khúc | Toàn bộ (ẩn danh) | Luôn (xuất tệp) |
| Quản trị dữ liệu | Theo yêu cầu, có mục đích | Toàn bộ | Luôn |

## Giai đoạn 5 — Giám sát, phát hiện bất thường & diễn tập

**Mục tiêu:** phát hiện sớm và phản ứng nhanh.

1. Phân tích audit để thiết lập baseline truy cập bình thường.
2. Định nghĩa quy tắc phát hiện bất thường.
3. Kết nối cảnh báo vào kênh vận hành (SIEM/Slack/on-call).
4. Xây dựng quy trình ứng phó sự cố.
5. Diễn tập định kỳ với kịch bản rò rỉ giả lập.
6. Kiểm tra khả năng truy nguồn một bản ghi.

:::tip[Kết quả đáng giá nhất]
Thời gian điều tra một nghi vấn rò rỉ rút từ nhiều tuần xuống còn vài phút — vì đã có sẵn actor, thời điểm và mục đích trong audit.
:::

## Giai đoạn 6 — Tuân thủ, kiểm toán & vận hành liên tục

**Mục tiêu:** biến nỗ lực kỹ thuật thành bằng chứng tuân thủ, duy trì lâu dài.

1. Tổng hợp bằng chứng tuân thủ Nghị định 13/2023.
2. Thiết lập quy trình xử lý yêu cầu chủ thể dữ liệu (DSAR).
3. Mời pentest bên thứ ba.
4. Rà soát quyền truy cập định kỳ (access review).
5. Vận hành vòng đời khóa mã hóa (xoay/thu hồi).
6. Cập nhật bản đồ dữ liệu khi có trường/hệ thống mới.
