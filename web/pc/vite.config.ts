import { defineConfig } from 'vite';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';
import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers';
import compression from 'vite-plugin-compression';
import { visualizer } from 'rollup-plugin-visualizer';

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      imports: ['vue', 'vue-router', 'pinia'],
      resolvers: [ElementPlusResolver()],
      dts: 'src/auto-imports.d.ts'
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts'
    }),
    compression({
      algorithm: 'gzip',
      ext: '.gz'
    }),
    visualizer({
      open: false,
      gzipSize: true,
      brotliSize: true
    })
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('../common', import.meta.url))
    }
  },
  server: {
    port: 3003,
    proxy: {
      // Dev mode: direct to Go service API ports
      // Production: APISIX :9080 handles routing + JWT
      '/api/auth': {
        target: 'http://127.0.0.1:8881',
        changeOrigin: true
      },
      '/api/users': {
        target: 'http://127.0.0.1:8882',
        changeOrigin: true
      },
      '/api/perm': {
        target: 'http://127.0.0.1:8883',
        changeOrigin: true
      },
      '/api/files': {
        target: 'http://127.0.0.1:8884',
        changeOrigin: true
      },
      '/api/property': {
        target: 'http://127.0.0.1:8882',
        changeOrigin: true
      },
      '/api/verifications': {
        target: 'http://127.0.0.1:8882',
        changeOrigin: true
      },
      '/api/masterdata': {
        target: 'http://127.0.0.1:8889',
        changeOrigin: true
      },
      '/api/monitoring': {
        target: 'http://127.0.0.1:8886',
        changeOrigin: true
      },
      '/api/v1': {
        target: 'http://127.0.0.1:8891',
        changeOrigin: true
      },
      '/api/moderation': {
        target: 'http://127.0.0.1:8890',
        changeOrigin: true
      }
    }
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: (id) => {
          if (id.includes('node_modules')) {
            if (id.includes('element-plus')) {
              return 'element-plus';
            }
            if (id.includes('@element-plus/icons-vue')) {
              return 'element-icons';
            }
            if (id.includes('vue') || id.includes('pinia') || id.includes('vue-router')) {
              return 'vue-vendor';
            }
          }
        }
      }
    },
    chunkSizeWarningLimit: 1000
  }
});
