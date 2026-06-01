---
title: 5. Phân tích hệ thống & plan task
description: Phân rã công việc theo vai trò (BA, Dev, DB, QC, DevOps, Security) để lên plan task.
---

## Quy ước

- **FR-xx** = yêu cầu chức năng; **NFR-xx** = phi chức năng; **US-xx** = user story.
- Mã task gợi ý: `[VaiTrò]-[Module]-[số]`, ví dụ `DEV-M2-01`, `DB-M6-02`.
- Ưu tiên: **P0** (bắt buộc, chặn) · **P1** (cao) · **P2** (nên có).

## Đối tượng đọc

| Vai trò | Dùng để | Phần đọc chính |
|---|---|---|
| BA | Chốt yêu cầu, user story, AC | Yêu cầu, BA |
| Tech / Dev | Bóc task hiện thực module, API | Dev, API |
| DB | Schema, migration, RLS | DB |
| QC | Test plan, ca kiểm thử | QC |
| DevOps | Hạ tầng, CI/CD, HA, KMS | DevOps |
| Security | Mô hình đe dọa, kiểm soát | Security |

## Yêu cầu chức năng (FR)

| Mã | Mô tả | Module |
|---|---|---|
| FR-01 | Tạo subject và sinh pii_ref | M1 |
| FR-02 | Đọc PII qua cổng kèm purpose | M2,M3,M4 |
| FR-03 | Ghi/cập nhật PII, mã hóa khi lưu | M2,M3 |
| FR-04 | Tra cứu chính xác (blind index) | M3 |
| FR-05 | Mã hóa từng trường; xoay khóa | M3 |
| FR-06 | Che dữ liệu theo vai trò | M4 |
| FR-07 | RBAC + RLS + kiểm tra purpose | M5 |
| FR-08 | Phê duyệt bốn mắt | M5 |
| FR-09 | Ghi audit bất biến | M6 |
| FR-10 | Xác minh toàn vẹn chuỗi audit | M6 |
| FR-11 | Phát hiện bất thường & cảnh báo | M7 |
| FR-12 | Xử lý DSAR theo NĐ 13/2023 | M8 |
| FR-13 | Xóa bằng crypto-shred | M3,M8 |
| FR-14 | Kết xuất bằng chứng tuân thủ | M6,M8 |

## Yêu cầu phi chức năng (NFR)

| Mã | Mô tả | Tiêu chí đo |
|---|---|---|
| NFR-01 | Hiệu năng cổng | p99 độ trễ thêm < ~40ms |
| NFR-02 | Quy mô dữ liệu | Ổn ở > 1 triệu subject |
| NFR-03 | Tính sẵn sàng | HA; có RTO/RPO |
| NFR-04 | Toàn vẹn audit | Không mất bản ghi; phát hiện sửa |
| NFR-05 | Bảo mật khóa | Khóa tách khỏi dữ liệu & DBA |
| NFR-06 | Data residency | Trong phạm vi NĐ 13/2023 |
| NFR-08 | Truy nguồn | Dựng lịch sử 1 bản ghi trong vài phút |

## Phân rã theo vai trò

### Business Analyst (BA)

| Task | Nội dung | Ưu tiên |
|---|---|---|
| BA-01 | Danh mục purpose hợp lệ | P0 |
| BA-02 | Vai trò & ma trận quyền | P0 |
| BA-03 | Quy tắc masking | P1 |
| BA-04 | Quy tắc bốn mắt | P1 |
| BA-05 | Luồng DSAR & thời hạn NĐ 13/2023 | P0 |
| BA-06 | Tiêu chí nghiệm thu (AC) cho FR | P0 |

### Tech / Developer

| Task | Nội dung | Ưu tiên |
|---|---|---|
| DEV-M6-01 | Service audit: Append + hash-chain | P0 |
| DEV-M2-01 | Khung cổng M2: authn, định tuyến | P0 |
| DEV-M2-02 | Reveal/Store/Update/Lookup | P0 |
| DEV-M3-01 | Envelope encryption, AES-GCM | P0 |
| DEV-M3-02 | Blind index + chuẩn hóa | P1 |
| DEV-M5-01 | Authorize (RBAC+purpose), default-deny | P0 |
| DEV-M5-02 | RLS hook + four-eyes workflow | P1 |
| DEV-M4-01 | Engine masking theo policy | P1 |
| DEV-M7-01 | Detector + cảnh báo | P1 |
| DEV-M8-01 | DSAR orchestrator + ExportEvidence | P1 |
| DEV-MIG-01 | Công cụ dual-write + backfill | P1 |

:::tip[Thứ tự gợi ý]
Bắt đầu DEV-M6-01 và DEV-M2-01 trước — tạo khả năng nhìn thấy sớm; rồi M3 → M5 → M4 → M7 → M8.
:::

### Database (DB)

| Task | Nội dung | Ưu tiên |
|---|---|---|
| DB-M1-01 | Schema subject, subject_field + index | P0 |
| DB-M6-01 | Schema pii_audit; thu hồi UPDATE/DELETE | P0 |
| DB-M6-02 | Tách instance/credentials kho audit | P0 |
| DB-M5-01 | Schema role, role_grant, purpose | P0 |
| DB-M5-02 | Chính sách RLS theo claims | P1 |
| DB-ENC-01 | Bật TDE; kiểm thử hiệu năng cột mã hóa | P0 |
| DB-PERF-01 | Tối ưu index ở >1 triệu bản ghi | P1 |
| DB-MIG-01 | Migration + rollback + đối soát | P1 |

### Quality Control (QC)

| Mã | Kịch bản | Kỳ vọng |
|---|---|---|
| TC-AUTH-1 | Gọi reveal thiếu purpose | 403 DENY + audit |
| TC-RBAC-1 | Vai trò không có quyền đọc | 403 DENY |
| TC-RLS-1 | Đọc subject ngoài phạm vi | 403 DENY |
| TC-4EYES-1 | Bulk reveal chưa duyệt | 202 PENDING |
| TC-MASK-1 | CSKH đọc SĐT | Trả giá trị đã che |
| TC-AUD-1 | Sửa 1 dòng audit | VerifyChain báo broken |
| TC-SHRED-1 | DSAR erase | Giải mã sau đó thất bại |
| TC-PERF-1 | 1000 reveal/giây | p99 trong ngưỡng NFR-01 |

### DevOps / SRE

| Task | Nội dung | Ưu tiên |
|---|---|---|
| OPS-01 | Dựng KMS/HSM hoặc Vault (residency VN) | P0 |
| OPS-02 | Hạ tầng HA cho M2 | P0 |
| OPS-03 | Cụm DB + kho audit tách biệt | P0 |
| OPS-04 | CI/CD + quét bảo mật | P1 |
| OPS-05 | Quản lý secret/mTLS, xoay tự động | P0 |
| OPS-07 | Backup mã hóa & diễn tập khôi phục | P1 |
| OPS-08 | Kiểm thử hành vi fail-closed | P0 |

### Security

| Task | Nội dung | Ưu tiên |
|---|---|---|
| SEC-01 | Mô hình đe dọa toàn hệ | P0 |
| SEC-02 | Duyệt thiết kế mã hóa (nonce, KEK/DEK) | P0 |
| SEC-03 | Duyệt mô hình phân quyền & bốn mắt | P0 |
| SEC-04 | Hardening cổng (input, rate-limit, SSRF) | P1 |
| SEC-05 | Chính sách rò rỉ PII khỏi log/lỗi/URL | P0 |
| SEC-06 | Kế hoạch pentest bên thứ ba | P1 |
| SEC-07 | Quy trình ứng phó & thông báo sự cố | P1 |

## Gợi ý sắp xếp plan theo giai đoạn

| GĐ | Cụm công việc chính | Vai trò chủ lực |
|---|---|---|
| GĐ 1 | Khảo sát, BA chốt purpose/vai trò, DB schema nền | BA, DB |
| GĐ 2 | Audit (M6) + khung cổng (M2) | Dev, DB, DevOps |
| GĐ 3 | Mã hóa (M3), masking (M4), KMS, migration | Dev, DB, DevOps |
| GĐ 4 | Access control (M5), RLS, bốn mắt | Dev, Security, BA |
| GĐ 5 | Giám sát (M7), cảnh báo, diễn tập | Dev, DevOps, Security |
| GĐ 6 | DSAR (M8), tuân thủ, pentest, access review | BA, Security, DPO |

:::tip[Cột mốc kiểm soát]
Sau mỗi giai đoạn, QC nghiệm thu theo tiêu chí và Security ký duyệt các hạng mục P0 trước khi sang giai đoạn sau.
:::
