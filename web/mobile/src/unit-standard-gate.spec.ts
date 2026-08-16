import { describe, expect, it } from 'vitest';
import { spawnSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import { dirname, join } from 'node:path';

/**
 * 回归测试：QA 门禁 harness-checks-frontend.sh 的 unit_standard 检查
 * （长度/字号一律 rem）不得误报两类合法内容：
 *   1) 守卫/单测测试文件（*.spec.ts / *.test.ts）内的用例名、正则、断言文本；
 *   2) 多行块注释（斜杠加星号起、星号加斜杠止）的续行——续行内含闭注释符，
 *      而行首没有开注释符。
 *
 * 修复背景（2026-08-16）：check_unit_standard 曾以 grep --include='*.ts' 扫描，
 * 未排除测试文件，且注释排除仅匹配行内含开注释符的行；导致 unit-system.spec.ts
 * 的用例名/正则/断言与 App.vue 块注释续行被误报为 rpx/px 违规
 * （详见 _qa.md 根因分析）。
 *
 * 递归守卫：检查脚本的 unit_test 步骤会再次调用 vitest。本用例向被调用的
 * 检查脚本注入 HARNESS_RECURSE=1，使由该脚本触发的内层 vitest 中本用例直接
 * 放行，避免"检查脚本 → vitest → 本用例 → 检查脚本"无限递归。
 */
const HARNESS_RECURSE = process.env.HARNESS_RECURSE === '1';

/** 自 cwd 向上定位 QA 检查脚本（对 cwd 变化鲁棒） */
function findCheckScript(): string {
  let dir = process.cwd();
  for (;;) {
    const candidate = join(dir, '.harness', 'skills', 'qa', 'scripts', 'harness-checks-frontend.sh');
    if (existsSync(candidate)) return candidate;
    const parent = dirname(dir);
    if (parent === dir) break;
    dir = parent;
  }
  throw new Error('harness-checks-frontend.sh 未找到（应位于仓库 .harness/skills/qa/scripts/）');
}

describe('QA 门禁 unit_standard（rem only）', () => {
  it(
    '不误报测试文件与块注释续行（回归 2026-08-16）',
    () => {
      if (HARNESS_RECURSE) {
        return; // 由检查脚本触发的内层 vitest：直接放行，防递归
      }

      const script = findCheckScript();
      const res = spawnSync('bash', [script, '--service', 'mobile', '--json'], {
        encoding: 'utf8',
        timeout: 120_000,
        env: { ...process.env, HARNESS_RECURSE: '1' },
      });

      // 脚本 stdout 为纯 JSON（进度输出在 stderr）；从首个 { 起解析以容错
      const stdout = res.stdout ?? '';
      const start = stdout.indexOf('{');
      expect(start, `检查脚本无 JSON 输出（exit=${res.status}）`).toBeGreaterThanOrEqual(0);
      const report = JSON.parse(stdout.slice(start)) as {
        results: Array<{ check: string; status: string; detail?: string }>;
      };

      const unit = report.results.find((r) => r.check === 'unit_standard');
      expect(unit, 'unit_standard 结果缺失').toBeTruthy();
      expect(unit!.status, `unit_standard detail: ${unit!.detail ?? '(none)'}`).toBe('PASS');
    },
    120_000, // 检查脚本含 build/type-check/嵌套 vitest，耗时约 15s
  );
});
