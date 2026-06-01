---
title: 1. Đề xuất kiến trúc
description: Định vị giải pháp bảo vệ PII theo hướng kết hợp DB-native + Masking và lớp Audit/Access-Control tập trung.
---

## Tóm tắt điều hành

Tài liệu này trình bày một hướng tiếp cận thực dụng để bảo vệ dữ liệu cá nhân (PII) của khách hàng, thiết kế cho bối cảnh một doanh nghiệp bán lẻ/TMĐT với hơn một triệu bản ghi khách hàng và mục tiêu trọng tâm là chống rò rỉ dữ liệu cũng như truy được nguồn khi sự cố xảy ra.

Khác với phương án xây dựng một PII Vault hoàn chỉnh (tốn kém, rủi ro cao, thời gian dài) hay mua một sản phẩm thương mại (chi phí license, phụ thuộc nhà cung cấp), hướng kết hợp này tận dụng **các tính năng bảo mật sẵn có trong cơ sở dữ liệu** (mã hóa, masking, kiểm soát truy cập) và bổ sung **một lớp audit và access-control tập trung** để giải quyết đúng điểm đau cốt lõi: hiện nay không ai biết ai đã đọc dữ liệu của khách nào, khi nào và vì sao.

:::tip[Triết lý nền tảng]
"Bạn không thể bịt một lỗ rò mà bạn không nhìn thấy." Khoản đầu tư đầu tiên và quan trọng nhất là **khả năng nhìn thấy** — ghi vết mọi truy cập vào PII một cách bất biến.
:::

:::note[Phạm vi]
Đây là giải pháp giảm thiểu rủi ro theo chiều sâu (defense-in-depth), KHÔNG phải giải pháp bảo vệ dữ liệu khi đang xử lý (data in-use) ở mức mật mã như searchable encryption. Nếu về sau cần năng lực đó, hướng kết hợp này vẫn là nền móng tốt để tiến lên vault đầy đủ mà không phải làm lại từ đầu.
:::

## Vấn đề cần giải quyết

Ba triệu chứng quan sát được trong hiện trạng đều quy về một gốc rễ chung:

- **Dữ liệu để trần và phân tán:** PII (họ tên, số điện thoại, email, địa chỉ) được lưu plaintext, rải rác trên nhiều hệ thống — CRM, DB đơn hàng, log, file Excel xuất tay, backup, service đối tác.
- **Không có nhật ký truy cập:** Khi xảy ra rò rỉ, không có bằng chứng để xác định nguồn. Mỗi lần điều tra là một lần "mò trong bóng tối".
- **Không kiểm soát quyền theo mục đích:** Quá nhiều người và dịch vụ đọc được gần như mọi thứ; không ghi nhận vì sao một bản ghi bị truy cập.

**Hệ quả pháp lý:** Nghị định 13/2023/NĐ-CP về bảo vệ dữ liệu cá nhân yêu cầu xử lý dữ liệu đúng mục đích, có kiểm soát và truy vết. Việc thiếu nhật ký và kiểm soát truy cập khiến doanh nghiệp khó chứng minh tuân thủ khi bị thanh tra.

## Tổng quan kiến trúc hướng kết hợp

Giải pháp gồm hai trụ cột bổ sung cho nhau.

### Trụ cột A — DB-native + Masking (lớp dữ liệu)

Tận dụng các tính năng bảo mật tích hợp sẵn trong cơ sở dữ liệu hiện đại:

- **TDE (Transparent Data Encryption):** mã hóa toàn bộ dữ liệu khi lưu trên đĩa.
- **Column / Field-level Encryption:** mã hóa riêng từng cột nhạy cảm với khóa quản lý tách biệt.
- **Dynamic Data Masking:** che dữ liệu theo vai trò khi truy vấn.
- **Row-Level Security (RLS):** giới hạn mỗi vai trò chỉ đọc được những dòng thuộc phạm vi của mình.

### Trụ cột B — Lớp Audit & Access-Control tập trung

Đây là phần tạo khác biệt lớn nhất. Mọi truy vấn chạm vào PII đều được ghi nhận bởi một lớp tập trung, để lại dấu vết bất biến:

- **Cổng truy cập / Data Access Layer:** ứng dụng đọc PII qua một service chung thay vì truy vấn trực tiếp bảng nhạy cảm.
- **Bắt buộc khai mục đích (purpose binding):** mỗi yêu cầu đọc PII phải kèm lý do.
- **Audit log bất biến (append-only, hash-chained):** mỗi truy cập ghi một dòng: ai · gì · khi nào · vì sao · kết quả.
- **RBAC + least-privilege, mặc định từ chối:** thao tác nhạy cảm yêu cầu phê duyệt "bốn mắt".
- **Phát hiện bất thường & cảnh báo realtime.**

## Vì sao chọn hướng kết hợp này

| Tiêu chí | Tự xây Vault | Mua sản phẩm | Hướng kết hợp |
|---|---|---|---|
| Chi phí ban đầu | Cao | Trung bình–cao | Thấp |
| Thời gian triển khai | Chậm (12+ tháng) | Trung bình | Nhanh (theo quý) |
| Rủi ro lỗi mật mã | Cao (tự gánh) | Thấp | Thấp |
| Chống rò rỉ & truy nguồn | Có | Có | Có (trọng tâm) |
| Bảo vệ data in-use | Có (nếu đúng) | Mạnh | Hạn chế |
| Phụ thuộc nhà cung cấp | Không | Cao | Thấp |
| Data residency (VN) | Tự kiểm soát | Cần xác nhận | Tự kiểm soát |

**Kết luận định vị:** với mục tiêu số một là chống rò rỉ và truy nguồn, hướng kết hợp đạt được phần lớn giá trị với chi phí và rủi ro thấp nhất, đồng thời giữ toàn quyền kiểm soát dữ liệu trong nước.

## Lộ trình tóm tắt theo giai đoạn

| Giai đoạn | Trọng tâm | Kết quả chính |
|---|---|---|
| GĐ 1 | Khảo sát & lập bản đồ dữ liệu | Bản đồ PII, ma trận phân loại |
| GĐ 2 | Lớp audit tập trung (ưu tiên cao nhất) | Mọi truy cập PII để lại dấu vết bất biến |
| GĐ 3 | Mã hóa (TDE + cột) & masking | Lộ DB/backup không còn lộ dữ liệu thật |
| GĐ 4 | Siết kiểm soát truy cập | Quyền tối thiểu, mặc định từ chối |
| GĐ 5 | Giám sát, phát hiện bất thường | Phát hiện sớm; điều tra còn vài phút |
| GĐ 6 | Tuân thủ, kiểm toán, vận hành | Bằng chứng tuân thủ NĐ 13/2023 |
