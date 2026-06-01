# Hướng dẫn triển khai lên Cloudflare Pages

> **Quan trọng — đây là dự án Cloudflare *Pages* (web tĩnh), KHÔNG phải Cloudflare *Workers*.**
> Nếu build log báo lỗi đòi `main = "src/index.ts"` hoặc `[assets] directory`, nghĩa là
> Cloudflare đang chạy nhầm `wrangler deploy` (chế độ Workers). Cách khắc phục ở mục
> "Xử lý sự cố" bên dưới.

Có hai cách deploy: qua giao diện Cloudflare (khuyến nghị) và qua GitHub Actions.

## Bước 0 — Đẩy mã nguồn lên Git

```bash
cd pii-docs
git remote add origin https://github.com/<your-org>/pii-vault-docs.git
git push -u origin main
```

## Cách A — Kết nối Git trên Cloudflare Pages (KHUYẾN NGHỊ)

1. Cloudflare Dashboard → **Workers & Pages** → **Create** → chọn tab **Pages**
   → **Connect to Git**. (Phải chọn đúng tab **Pages**, không phải Workers.)
2. Chọn repo `pii-vault-docs`.
3. Cấu hình build (Build settings):
   - **Framework preset:** `Astro`
   - **Build command:** `npm run build`
   - **Build output directory:** `dist`
4. (Tùy chọn) thêm biến môi trường: **Settings → Environment variables**
   - `NODE_VERSION` = `20`
5. **Save and Deploy**. Cloudflare cấp URL `https://pii-vault-docs.pages.dev`.
6. Mỗi lần `git push` lên `main`, Cloudflare tự build lại.

> Dự án này KHÔNG cần `wrangler.toml`. Với Cloudflare Pages dùng kết nối Git,
> toàn bộ cấu hình build nằm ở giao diện dashboard. Có một `wrangler.toml`
> cấu hình kiểu Worker là nguyên nhân chính gây lỗi deploy.

## Cách B — Deploy bằng GitHub Actions + Wrangler (Pages)

Đã có sẵn `.github/workflows/deploy.yml`. Workflow build ra `dist/` rồi gọi
`wrangler pages deploy` (đúng chế độ Pages, không phải Workers).

Tạo secret trong GitHub repo (**Settings → Secrets and variables → Actions**):

- `CLOUDFLARE_API_TOKEN` — token có quyền **Cloudflare Pages: Edit**.
- `CLOUDFLARE_ACCOUNT_ID` — Account ID trong Cloudflare Dashboard.

Lần đầu, tạo project Pages trống tên `pii-vault-docs` (hoặc để lệnh tự tạo):

```bash
npx wrangler pages project create pii-vault-docs --production-branch main
```

## Kiểm tra build cục bộ trước khi deploy

```bash
npm install
npm run build      # phải tạo ra thư mục dist/
npm run preview
```

Nếu `npm run build` chạy sạch và có `dist/`, deploy trên Cloudflare sẽ tương tự.

## Xử lý sự cố

### Lỗi: đòi `main = "src/index.ts"` hoặc `[assets] directory = "./dist"`
Cloudflare đang chạy **`wrangler deploy`** (Workers) thay vì **Pages**. Nguyên nhân
thường gặp và cách sửa:

1. **Có `wrangler.toml` kiểu Worker trong repo** → đã gỡ bỏ trong bản này. Đảm bảo
   repo của bạn KHÔNG còn `wrangler.toml`. Nếu muốn giữ, nội dung phải là Pages:
   ```toml
   name = "pii-vault-docs"
   pages_build_output_dir = "dist"
   ```
   và lệnh deploy phải là `wrangler pages deploy dist`, KHÔNG phải `wrangler deploy`.
2. **Tạo nhầm project Workers trên dashboard** → xóa và tạo lại bằng tab **Pages**
   → **Connect to Git** (Cách A).
3. **Build command sai** → phải là `npm run build`, output `dist`.

### Lỗi Node version
Đặt `NODE_VERSION=20` trong Environment variables của Pages.

### Trang trắng / 404 ở route con
Kiểm tra **Build output directory** = `dist`.

### Sitemap/canonical URL sai
Cập nhật trường `site` trong `astro.config.mjs` thành domain thật.
