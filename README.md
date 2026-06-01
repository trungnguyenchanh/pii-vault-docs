# PII Vault Docs

Bộ tài liệu thiết kế và triển khai **hệ thống bảo vệ dữ liệu PII** theo hướng kết hợp *DB-native + Masking* và *lớp Audit/Access-Control tập trung*. Trang tài liệu song ngữ (Tiếng Việt / English) dựng bằng [Astro Starlight](https://starlight.astro.build), deploy lên [Cloudflare Pages](https://pages.cloudflare.com).

> _Documentation set for a PII data protection system using the hybrid DB-native + Masking and centralized Audit/Access-Control approach. Bilingual (VI/EN) docs built with Astro Starlight, deployed to Cloudflare Pages._

## Nội dung / Contents

| # | Tài liệu | Document |
|---|----------|----------|
| 1 | Đề xuất kiến trúc | Architecture Proposal |
| 2 | Sổ tay triển khai | Implementation Playbook |
| 3 | Đặc tả nghiệp vụ chức năng | Functional Specification |
| 4 | Thiết kế kỹ thuật chi tiết | Technical Design |
| 5 | Phân tích hệ thống & plan task | System Analysis & Task Plan |

## Yêu cầu / Requirements

- Node.js 18+ (khuyến nghị 20+)
- npm (hoặc pnpm/yarn)

## Phát triển cục bộ / Local development

```bash
npm install
npm run dev          # http://localhost:4321
```

## Build

```bash
npm run build        # xuất ra ./dist
npm run preview      # xem thử bản build
```

## Cấu trúc thư mục / Structure

```
.
├── astro.config.mjs           # cấu hình Starlight + i18n (vi/en)
├── package.json
├── public/                    # tài nguyên tĩnh
├── src/
│   ├── content.config.ts
│   └── content/docs/
│       ├── vi/                # bản tiếng Việt (mặc định)
│       │   ├── index.mdx
│       │   └── guides/        # 5 tài liệu chính
│       └── en/                # bản tiếng Anh
│           ├── index.mdx
│           └── guides/
└── .github/workflows/         # CI deploy (tùy chọn)
```

> Dự án dùng **Cloudflare Pages** (web tĩnh), không cần `wrangler.toml`. Toàn bộ
> cấu hình build đặt trong dashboard Pages — xem `DEPLOY.md`.

## Triển khai lên Cloudflare Pages / Deploy

Xem hướng dẫn chi tiết trong [`DEPLOY.md`](./DEPLOY.md).

Tóm tắt: trong Cloudflare Pages, tạo project mới kết nối với repo Git này và đặt:

- **Build command:** `npm run build`
- **Build output directory:** `dist`
- **Framework preset:** Astro

## Lưu ý / Note

Bộ tài liệu mang tính tham chiếu thiết kế, không phải tư vấn pháp lý. Việc đối chiếu Nghị định 13/2023/NĐ-CP nên có ý kiến bộ phận pháp chế.

## License

[MIT](./LICENSE)
