import { describe, it, expect, beforeAll } from 'vitest';
import { readdirSync, readFileSync, existsSync, statSync } from 'node:fs';
import { join, dirname, resolve } from 'node:path';
import { homedir } from 'node:os';
import { fileURLToPath } from 'node:url';

// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 本文件 RED 阶段捕获真实 FAIL 摘录

/**
 * 记忆引用完整性回归测试（修复 standards-eng 多视角审查 CRITICAL）
 *
 * 背景：login/agreement/reg-pending 三处 SEE 标记引用了不存在的记忆 slug
 * （sms-code-persist-localstorage / frontend-cross-page-storage-contract /
 * cross-page-sensitive-temp-data-storage），被 M2 规则判定为 🔴 CRITICAL。
 * 本测试扫描 src/ 全部 SEE 标记引用的 slug，断言每个 slug 都能在
 * 项目记忆目录 .harness/knowledge/memory 下找到同名 .md 文件（递归），
 * 或在本机个人记忆目录 ~/.claude/projects/<repo>/memory 下找到，
 * 防止再次出现悬空引用。
 */

const __dirname = dirname(fileURLToPath(import.meta.url));
// 本文件位于 web/mobile/src/utils/memory-refs.spec.ts
// SRC_DIR = .. = web/mobile/src（扫描范围）
const SRC_DIR = resolve(__dirname, '..');
// REPO_ROOT = ../../../../ = 仓库根目录（src/utils → src → mobile → web → 仓库根）
const REPO_ROOT = resolve(__dirname, '../../../../');
const PROJECT_MEMORY_DIR = join(REPO_ROOT, '.harness/knowledge/memory');

const SPEC_SLUG = 'tdd-red-evidence-requires-fail-excerpt';

/**
 * 本机个人记忆目录：~/.claude/projects/<repo-fingerprint>/memory
 * fingerprint = 绝对仓库路径去掉前导 '/'，将 '/' 替换为 '-'，并加前导 '-'。
 * 例：/home/jiaoxh/my-project/community-and-home → -home-jiaoxh-my-project-community-and-home
 */
const REPO_FINGERPRINT = `-${REPO_ROOT.slice(1).replace(/\//g, '-')}`;
const PERSONAL_MEMORY_ROOT = join(homedir(), '.claude', 'projects', REPO_FINGERPRINT, 'memory');

/** 递归收集目录下所有文件绝对路径 */
function collectFiles(dir: string, acc: string[] = []): string[] {
  if (!existsSync(dir)) return acc;
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      collectFiles(full, acc);
    } else {
      acc.push(full);
    }
  }
  return acc;
}

/** 在项目记忆目录（递归）中查找 {slug}.md，存在返回 true */
function projectMemoryExists(slug: string): boolean {
  return collectFiles(PROJECT_MEMORY_DIR).some((f) => f.endsWith(`/${slug}.md`) || f.endsWith(`\\${slug}.md`));
}

/** 在个人记忆目录中查找 {slug}.md，存在返回 true */
function personalMemoryExists(slug: string): boolean {
  if (!existsSync(PERSONAL_MEMORY_ROOT)) return false;
  return existsSync(join(PERSONAL_MEMORY_ROOT, `${slug}.md`));
}

/** 解析单个 slug 是否存在对应记忆文件（项目记忆 ∪ 个人记忆） */
function memoryFileExists(slug: string): boolean {
  if (projectMemoryExists(slug)) return true;
  if (personalMemoryExists(slug)) return true;
  return false;
}

/** 提取 src/ 下所有 `SEE: [[slug]]` 引用的 slug 列表（去重） */
function collectReferencedSlugs(): string[] {
  const slugs = new Set<string>();
  const pattern = /\/\/\s*SEE:\s*\[\[([a-z0-9-]+)\]\]|\*\s*SEE:\s*\[\[([a-z0-9-]+)\]\]/g;
  for (const file of collectFiles(SRC_DIR)) {
    if (!/\.(ts|vue|tsx)$/.test(file)) continue;
    const content = readFileSync(file, 'utf-8');
    let m: RegExpExecArray | null;
    while ((m = pattern.exec(content)) !== null) {
      slugs.add((m[1] ?? m[2]) as string);
    }
  }
  return [...slugs];
}

describe('记忆引用完整性（SEE 标记 slug 无悬空引用）', () => {
  let referencedSlugs: string[];

  beforeAll(() => {
    referencedSlugs = collectReferencedSlugs();
  });

  it('src/ 至少存在一处 // SEE: 引用（测试自身能扫到）', () => {
    expect(referencedSlugs.length).toBeGreaterThan(0);
    expect(referencedSlugs).toContain(SPEC_SLUG);
  });

  it('每个被引用的记忆 slug 都能解析到项目或个人记忆文件', () => {
    const missing = referencedSlugs.filter((slug) => !memoryFileExists(slug));
    expect(missing).toEqual([]);
  });

  it('本次修复目标：3 个 smsCode 存储相关 slug 必须已存在', () => {
    const criticalSlugs = [
      'sms-code-persist-localstorage',
      'frontend-cross-page-storage-contract',
      'cross-page-sensitive-temp-data-storage',
    ];
    for (const slug of criticalSlugs) {
      expect(memoryFileExists(slug), `悬空记忆引用: [[${slug}]]`).toBe(true);
    }
  });
});
