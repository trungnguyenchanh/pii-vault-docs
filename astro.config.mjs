// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// https://astro.build/config
export default defineConfig({
  // Khi deploy lên Cloudflare Pages với domain riêng, đổi site cho đúng.
  // Có thể để trống nếu dùng *.pages.dev mặc định.
  site: 'https://pii-vault-docs.pages.dev',
  integrations: [
    starlight({
      title: 'PII Vault Docs',
      description:
        'Tài liệu hệ thống bảo vệ dữ liệu PII — hướng kết hợp DB-native + Masking & lớp Audit/Access-Control tập trung.',
      // Song ngữ: tiếng Việt là mặc định (root locale, phục vụ tại /),
      // tiếng Anh là bản dịch (phục vụ tại /en/).
      defaultLocale: 'root',
      locales: {
        root: { label: 'Tiếng Việt', lang: 'vi' },
        en: { label: 'English', lang: 'en' },
      },
      social: {
        github: 'https://github.com/your-org/pii-vault-docs',
      },
      sidebar: [
        {
          label: 'Bắt đầu',
          translations: { en: 'Getting Started' },
          items: [
            { slug: 'index' },
          ],
        },
        {
          label: 'Tài liệu chính',
          translations: { en: 'Core Documents' },
          items: [
            { slug: 'guides/01-de-xuat-kien-truc' },
            { slug: 'guides/02-so-tay-trien-khai' },
            { slug: 'guides/03-dac-ta-nghiep-vu' },
            { slug: 'guides/04-thiet-ke-ky-thuat' },
            { slug: 'guides/05-phan-tich-he-thong' },
          ],
        },
      ],
    }),
  ],
});
