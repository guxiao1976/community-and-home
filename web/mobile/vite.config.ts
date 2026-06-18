import { defineConfig } from 'vite';
import uni from '@dcloudio/vite-plugin-uni';
import { fileURLToPath, URL } from 'node:url';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [uni()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('../common', import.meta.url)),
    },
  },
  server: {
    port: 3004,
    host: '0.0.0.0',
    proxy: {
      // Dev mode: direct to Go service API ports
      // Production: APISIX :9080 handles routing + JWT
      '/api/auth': {
        target: 'http://127.0.0.1:8881',
        changeOrigin: true,
      },
      '/api/users': {
        target: 'http://127.0.0.1:8882',
        changeOrigin: true,
      },
      '/api/files': {
        target: 'http://127.0.0.1:8884',
        changeOrigin: true,
      },
      '/api/masterdata': {
        target: 'http://127.0.0.1:8889',
        changeOrigin: true,
      },
      '/api/moderation': {
        target: 'http://127.0.0.1:8890',
        changeOrigin: true,
      },
      '/api/community': {
        target: 'http://127.0.0.1:8887',
        changeOrigin: true,
      },
      '/api/v1': {
        target: 'http://127.0.0.1:8891',
        changeOrigin: true,
      },
    },
  },
});
