import { defineConfig } from 'vitest/config';
import vue from '@vitejs/plugin-vue';
import { fileURLToPath, URL } from 'node:url';

// Dedicated vitest config — deliberately does NOT load the uni() plugin
// (the uni build plugin is not needed for unit tests and complicates SFC
// resolution). .vue files are compiled by @vitejs/plugin-vue.
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
      '@common': fileURLToPath(new URL('../common', import.meta.url)),
    },
  },
  test: {
    environment: 'happy-dom',
    setupFiles: ['./vitest.setup.ts'],
    include: ['src/**/*.spec.ts'],
    coverage: {
      // 覆盖率量化（2026-08-17 接入）。此前 v8/istanbul 报「Something removed coverage/*.tmp」bug——
      // 根因是 unit-standard-gate.spec.ts spawnSync 调用 harness-checks-frontend.sh，内层 vitest 再跑 --coverage
      // 与当前进程冲突删 .tmp；已通过 HARNESS_RECURSE=1 守卫（内层不跑 coverage）修复。若复发，检查该守卫。
      provider: 'v8',
      include: ['src/**/*.{ts,vue}'],
      exclude: ['src/**/*.spec.ts', 'src/**/*.d.ts'],
      // 覆盖率门禁：低于阈值 FAIL。当前基线（2026-08-17 含全 src）：Stmts 62.4 / Branch 54.7 / Funcs 59.8 / Lines 62.7。
      // 阈值设「及格线」低于基线留缓冲，防存量拖累新代码被拒；新逻辑函数须有测试（TDD 已强制）。
      thresholds: {
        statements: 58,
        branches: 50,
        functions: 55,
        lines: 58,
      },
    },
  },
});
