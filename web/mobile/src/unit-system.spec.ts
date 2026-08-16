import { describe, expect, it } from 'vitest';
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';

// 单位体系守卫测试：全站长度/字号使用 rem（根字号固定 16px，非响应式），
// 换算锚定 375px 设计稿：1rpx=0.5px → N rpx → N/32 rem；N px → N/16 rem。
// 详见 App.vue 全局样式注释与 uni.scss。
// vitest 以项目根（web/mobile）为 cwd，src 目录即 process.cwd()/src。
const SRC = join(process.cwd(), 'src');

function collectFiles(dir: string, exts: string[]): string[] {
  const out: string[] = [];
  for (const name of readdirSync(dir)) {
    const full = join(dir, name);
    const st = statSync(full);
    if (st.isDirectory()) out.push(...collectFiles(full, exts));
    else if (exts.some((e) => name.endsWith(e))) out.push(full);
  }
  return out;
}

/** 去除块/行注释（避免注释里提到的单位误命中） */
function stripComments(src: string): string {
  return src
    .replace(/\/\*[\s\S]*?\*\//g, '')
    .replace(/^\s*\/\/.*$/gm, '');
}

/** 去除 env(...) 内部（fallback 0px 允许保留） */
function stripEnv(src: string): string {
  return src.replace(/env\([^)]*\)/g, '');
}

function stripHtmlRootFont(src: string): string {
  return src.replace(/html\s*\{\s*font-size:\s*16px;?\s*\}/, '');
}

describe('单位体系：rpx/px → rem（根字号 16px，rem=rpx÷32、px÷16，锚定 375px）', () => {
  it('src 下所有 .vue/.scss 不再残留 rpx 单位', () => {
    const offenders: string[] = [];
    for (const f of collectFiles(SRC, ['.vue', '.scss'])) {
      const m = stripComments(readFileSync(f, 'utf8')).match(/\d+(?:\.\d+)?rpx/);
      if (m) offenders.push(`${f}: ${m[0]}`);
    }
    expect(offenders).toEqual([]);
  });

  it('src 下无长度 px 残留（仅允许 html 根字号 16px 与 env() fallback）', () => {
    const offenders: string[] = [];
    for (const f of collectFiles(SRC, ['.vue', '.scss'])) {
      let content = stripComments(readFileSync(f, 'utf8'));
      content = stripEnv(content);
      if (f.endsWith('App.vue')) content = stripHtmlRootFont(content);
      const m = content.match(/\d+(?:\.\d+)?px/);
      if (m) offenders.push(`${f}: ${m[0]}`);
    }
    expect(offenders).toEqual([]);
  });

  it('uni.scss 主题变量换算为 rem 值', () => {
    const scss = readFileSync(join(SRC, 'uni.scss'), 'utf8');
    const vars: Record<string, string> = {};
    for (const line of scss.split('\n')) {
      const m = line.match(/^\$([\w-]+):\s*([^;]+);/);
      if (m) vars[m[1]] = m[2].trim();
    }
    expect(vars['uni-border-radius']).toBe('0.5rem'); // 8px
    expect(vars['uni-border-radius-card']).toBe('0.375rem'); // 12rpx
    expect(vars['uni-border-radius-btn']).toBe('1.5rem'); // 48rpx
    expect(vars['uni-spacing-sm']).toBe('0.5rem'); // 8px
    expect(vars['uni-spacing-md']).toBe('1rem'); // 16px
    expect(vars['uni-spacing-lg']).toBe('1.5rem'); // 24px
    expect(vars['uni-spacing-xl']).toBe('2rem'); // 32px
    expect(vars['uni-font-size-xs']).toBe('0.625rem'); // 10px
    expect(vars['uni-font-size-sm']).toBe('0.75rem'); // 12px
    expect(vars['uni-font-size-base']).toBe('0.875rem'); // 14px
    expect(vars['uni-font-size-md']).toBe('1rem'); // 16px
    expect(vars['uni-font-size-lg']).toBe('1.125rem'); // 18px
    expect(vars['uni-font-size-xl']).toBe('1.25rem'); // 20px
    // env() 内 fallback 0px 保留（指令：不动 env()）
    expect(vars['app-safe-area-bottom']).toContain('env(safe-area-inset-bottom, 0px)');
  });

  it('App.vue 全局样式设置固定根字号 html { font-size: 16px }', () => {
    const app = readFileSync(join(SRC, 'App.vue'), 'utf8');
    expect(app).toMatch(/html\s*\{\s*font-size:\s*16px;?\s*\}/);
  });
});
