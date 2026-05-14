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
    port: 3000,
    proxy: {
      '/api/identity': {
        target: 'http://172.31.39.71:8888',
        changeOrigin: true,
        agent: false
      },
      '/api/masterdata': {
        target: 'http://172.31.39.71:8889',
        changeOrigin: true,
        agent: false
      },
      '/api/moderation': {
        target: 'http://172.31.39.71:8890',
        changeOrigin: true,
        agent: false
      },
      '/api/v1': {
        target: 'http://127.0.0.1:8891',
        changeOrigin: true,
        agent: false
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
