// ESLint flat config — Vue3 + TypeScript（2026-08-17 接入）
// 门禁策略：harness-checks-frontend.sh check #10 对 diff 内新增文件跑 eslint，新违规拦截、存量 WARN。
import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import tseslint from 'typescript-eslint'
import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'

export default defineConfigWithVueTs(
  js.configs.recommended,
  tseslint.configs.recommended,
  pluginVue.configs['flat/essential'],
  vueTsConfigs.recommended,
  {
    name: 'app/files-to-lint',
    files: ['**/*.{ts,mts,tsx,vue}'],
    rules: {
      // 存量 98 处 as any 由 check_type_safety 单独 WARN 跟踪（目标 ≤10），ESLint 不再重复告警
      '@typescript-eslint/no-explicit-any': 'warn',
      // views/*.vue 单字组件名是项目惯例（List/Form/Detail…），不强制多字命名
      'vue/multi-word-component-names': 'warn',
    },
  },
  {
    name: 'app/files-to-ignore',
    ignores: [
      'dist/**',
      'coverage/**',
      'node_modules/**',
      'src/**/*.d.ts',
      'vitest.config.ts',
      'vite.config.ts',
      'playwright.config.ts',
    ],
  }
)
