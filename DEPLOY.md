# Hướng dẫn triển khai lên Cloudflare Pages

Tài liệu này mô tả hai cách deploy: qua giao diện Cloudflare (khuyến nghị, đơn giản nhất) và qua GitHub Actions với Wrangler.

## Bước 0 — Đẩy mã nguồn lên Git

```bash
cd pii-docs
git init
git add .
git commit -m "chore: initial PII Vault docs (Astro Starlight, bilingual)"
git branch -M main
git remote add origin https://github.com/<your-org>/pii-vault-docs.git
git push -u origin main
```

## Cách A — Kết nối Git trên Cloudflare Pages (khuyến nghị)

1. Đăng nhập Cloudflare Dashboard → **Workers & Pages** → **Create** → **Pages** → **Connect to Git**.
2. Chọn repo `pii-vault-docs` vừa đẩy lên.
3. Cấu hình build:
   - **Framework preset:** `Astro`
   - **Build command:** `npm run build`
   - **Build output directory:** `dist`
   - **Node version:** đặt biến môi trường `NODE_VERSION` = `20` nếu cần.
4. Bấm **Save and Deploy**. Cloudflare sẽ build và cấp một URL dạng `https://pii-vault-docs.pages.dev`.
5. Mỗi lần `git push` lên nhánh `main`, Cloudflare tự build lại (CI/CD sẵn có).

> Sau khi có domain chính thức, cập nhật trường `site` trong `astro.config.mjs` cho đúng để sitemap và canonical URL chuẩn.

## Cách B — Deploy bằng GitHub Actions + Wrangler

Phù hợp khi muốn kiểm soát pipeline trong repo. Đã có sẵn file `.github/workflows/deploy.yml`.

Cần tạo các secret trong GitHub repo (**Settings → Secrets and variables → Actions**):

- `CLOUDFLARE_API_TOKEN` — API token có quyền *Cloudflare Pages: Edit*.
- `CLOUDFLARE_ACCOUNT_ID` — Account ID lấy trong Cloudflare Dashboard.

Sau đó mỗi lần push lên `main`, workflow sẽ build và deploy bằng Wrangler.

### Tạo Cloudflare API Token

1. Cloudflare Dashboard → **My Profile** → **API Tokens** → **Create Token**.
2. Dùng template **Edit Cloudflare Workers** hoặc tạo custom với quyền **Account → Cloudflare Pages → Edit**.
3. Lưu token vào GitHub secret `CLOUDFLARE_API_TOKEN`.

## Kiểm tra build cục bộ trước khi deploy

```bash
npm install
npm run build
npm run preview
```

Nếu `npm run build` chạy sạch và `dist/` được tạo, deploy trên Cloudflare sẽ thành công tương tự.

## Xử lý sự cố thường gặp

- **Lỗi Node version:** đặt `NODE_VERSION=20` trong biến môi trường build của Pages.
- **Trang trắng / 404 trên route con:** đảm bảo **Build output directory** là `dist`, không phải thư mục khác.
- **Sitemap/URL sai:** kiểm tra trường `site` trong `astro.config.mjs`.
