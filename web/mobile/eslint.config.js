// ESLint flat config — Uni-app (Vue3) + TypeScript（2026-08-17 接入）
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
    name: 'uni-app/globals',
    files: ['**/*.{ts,tsx,vue}'],
    languageOptions: {
      globals: {
        uni: 'readonly',
        wx: 'readonly',
        plus: 'readonly',
        getApp: 'readonly',
        getCurrentPages: 'readonly',
      },
    },
  },
  {
    name: 'app/files-to-lint',
    files: ['**/*.{ts,tsx,vue}'],
    rules: {
      '@typescript-eslint/no-explicit-any': 'warn',
      'vue/multi-word-component-names': 'warn',
    },
  },
  {
    name: 'app/files-to-ignore',
    ignores: [
      'dist/**',
      'unpackage/**',
      'coverage/**',
      'node_modules/**',
      'src/**/*.d.ts',
    ],
  }
)
