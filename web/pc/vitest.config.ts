import { defineConfig } from 'vitest/config';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath } from 'node:url';

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: ['./tests/setup.ts'],
    exclude: ['**/node_modules/**', '**/tests/e2e/**'],
    coverage: {
      // 覆盖率量化（2026-08-17 接入，对齐 web/mobile）。
      provider: 'v8',
      include: ['src/**/*.{ts,vue}'],
      exclude: ['src/**/*.test.ts', 'src/**/*.spec.ts', 'src/**/*.d.ts'],
      // 阈值设「地板」：当前实测基线 Stmts 4.2 / Branch 3.2 / Funcs 2.7 / Lines 4.4（pc 存量单测极少，views 大块 0%）。
      // 低于基线留缓冲防回退即 FAIL；随测试增长应同步上调（对齐 mobile 做法：基线 62% → 及格线 58）。
      thresholds: {
        statements: 3,
        branches: 2,
        functions: 2,
        lines: 3,
      },
    },
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('../common', import.meta.url))
    }
  }
});
