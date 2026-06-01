import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      // Chuyển tiếp API sang Go backend khi dev.
      '/api': 'http://localhost:8080',
    },
  },
});
