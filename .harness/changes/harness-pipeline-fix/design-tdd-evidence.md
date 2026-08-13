# 修复目标 #4 设计：TDD 测试有效性机械强制（替代 RED 证据捕获）

> 归属：`harness-pipeline-fix` 修复目标 #4「TDD 证据强制捕获」
> 状态：设计定稿（待评审 → 修复 → 回归）
> 日期：2026-08-13

## 0. 结论先行

**「RED 证据捕获」修了 4 次都没闭环，因为它在追一个脆弱代理，而不是本体。**

真正的诉求是「**测试能不能抓到 bug（测试有效性）**」，RED 证据只是它的一个时间点代理。业界对「测试有效性」的机械验证标准答案是 **变异测试（mutation testing）**。本方案：**用变异测试替代 RED 证据作为测试有效性门禁，并把 RED 证据要求按「有无逻辑」分诊收窄**。

---

## 1. 问题诊断

### 1.1 现状

- 测试质量机械验证只有 `coverage` 阈值（quality-gates.yml：<30% error，<40% warning）。
- `tdd-red-evidence-requires-fail-excerpt`（must-follow 记忆）要求 RED 阶段留存真实 FAIL 输出摘录，**已 4 次复发**（master-data T2.2 / restore / web-mobile T5.1 / access-control permission-service）。
- 修复目标 #4「TDD 证据强制捕获」在 HANDOFF-WAVE1 已尝试（generator.js 加「RED→GREEN」铁律），但**未闭环**。

### 1.2 根因

RED 证据是一个 **时间点产物**（代码缺失时那个失败的瞬间），而：

1. **AI Generator 的工作方式**：一个批次写完实现+测试，`go test` 首跑即 GREEN，RED 瞬间天然被跳过。改提示词压不住（4 次复发证明）。
2. **「自觉」不是门禁**：规则靠 Generator 自觉留存，而自动化管线里只有「机械可查的状态」才是门禁。
3. **QA 只能事后查**：QA 能 grep 出「没有 FAIL 摘录」并判 FAIL，但没法让 Generator「曾经产生」那段 FAIL 输出。

**核心区分**：机械门禁能查「静态状态」（文件里有没有某字段、命令退出码），查不了「过程瞬间」（测试曾经失败过吗）。RED 证据属于后者。

---

## 2. 行业最佳实践：变异测试

变异测试把「测试有效性」从时间点产物变成**静态可查的机械验证**：

- 对代码注入变异（`==`→`!=`、`+`→`-`、`if x`→`if true`…），每个变异生成一个突变体。
- 对每个突变体跑测试：**测试失败 = 抓住了变异（好）；测试仍通过 = 没抓住（该分支没被真正测到）**。
- 变异存活率 = 未抓住的变异 / 总变异。存活率高 = 测试弱。

Go 标准工具：`github.com/zimmski/go-mutesting`。

### 与 RED 证据的等价性

| 维度 | RED 证据（现状） | 变异测试（最佳实践） |
|------|----------------|---------------------|
| 机制 | 靠 Generator 自觉留存失败瞬间 | 工具确定性注入变异、跑测试 |
| 可机械强制 | ❌（时间点，查不到） | ✅（退出码，CI 友好） |
| 强度 | 只证明「曾经失败过」 | 证明「每个关键分支都被测到」 |
| 覆盖 | 编译错误 + 断言 | 逻辑分支（`==`/`+`/条件翻转） |

---

## 3. 整改方案（分层）

### P0-1 分诊：RED 证据只要求「有逻辑的函数」

现状 QA 一刀切，把 `SysRole.Platforms` 这种「struct 加字段 + SQL 加列」也要求 RED 证据（tasks.md 自己写了「无独立逻辑，仅字段映射」）。正确分诊：

| 变更类型 | 例子 | 要求 |
|---------|------|------|
| 纯字段映射 | struct 加字段、SQL 加列、proto 透出、seed | 只要求 build/test 绿，**不要求 RED/变异** |
| 有逻辑的函数 | splitPlatforms、每户≤6 计数、端判定、祖先链解析 | 变异测试门禁 |

**落地**：generator.js prompt + qa.js 判据都按此分诊；tasks.md 已有「无独立逻辑」标注，QA 需据此跳过。

### P0-2 变异测试替代 RED 证据（治本）

- `harness-checks.sh` 加 `check_mutation_testing`：
  - **范围**：只对本次 diff 里「有逻辑的函数」所在 package 跑 go-mutesting（非全服务，避免太慢）。
  - **判据**：diff 函数相关的变异存活率 > 阈值（建议 20%）→ FAIL，报告具体存活突变体。
  - 新增的「字段映射」变更跳过。
- 这一步**完全机械**，不再依赖 Generator 自觉。
- `tdd-red-evidence-requires-fail-excerpt` 记忆降级或改写为「变异测试门禁」语义（`severity: must-follow` → 保留，但触发条件改为「有逻辑函数变更且变异存活率超标」）。

**实现注意**：go-mutesting 以 package 为单位变异，需过滤到 diff 涉及的函数；成本高时先做「changed package 全量变异」再逐步细化。

### P1 若保留 TDD 过程语义：checkpoint 化

若仍需「写测试先行」的 TDD 语义，把自觉变成硬关卡（与 git_hygiene check #17 同理）：

- Generator 阶段**强制分两步**：① 只写测试 → `go test` → FAIL 输出写入 `_tdd_evidence.md`；② 再写实现 → GREEN。
- QA **机械检查**：`_tdd_evidence.md` 存在且含真实 FAIL 文本（grep `undefined:` / `expected: … actual: …`），否则 FAIL。

### P2 回溯兜底：机械捕获 RED

对历史「无 RED 证据」的遗留，可脚本化补录，不靠 AI 重跑：

```
git stash 实现（保留测试）→ go test（必然 FAIL）→ 捕获输出 → git stash pop → go test（GREEN）
```

作为 pipeline 兜底步骤。

---

## 4. 落地任务清单

| # | 任务 | 文件 | 依赖 |
|---|------|------|------|
| T1 | QA 分诊：字段映射变更跳过 RED/变异判据 | `qa.js`、`harness-checks.sh` TDD 检查 | 无 |
| T2 | generator 分诊：有逻辑/字段映射在任务里标注 | `generator.js` | 无 |
| T3 | 引入 go-mutesting + 加 `check_mutation_testing` | `harness-checks.sh`、`go.mod` | T1 |
| T4 | 变异存活率阈值 + 报告存活突变体 | `check_mutation_testing` | T3 |
| T5 | 改写 `tdd-red-evidence-requires-fail-excerpt` 记忆为变异测试语义 | memory 文件 | T3 |
| T6 | （可选 P1）`_tdd_evidence.md` 机械门禁 | `harness-checks.sh` | 无 |
| T7 | （可选 P2）stash/pop 回溯捕获脚本 | `scripts/` | 无 |

## 5. 验收标准

- [ ] 有逻辑函数变更 → 变异测试自动运行，存活率超标 → FAIL（机械、可复现）
- [ ] 字段映射变更 → 不再被要求 RED 证据（QA 跳过）
- [ ] 小任务回归：单服务「有逻辑」改动，管线 ≤2 轮收敛，变异测试 PASS
- [ ] `tdd-red-evidence-requires-fail-excerpt` 不再因「无 FAIL 摘录」空转复发

## 6. 与现状的关系

- 不推翻 HANDOFF-WAVE1 已修的 6 项（generator 去 worktree、QA 审 diff 等），只补 #4 的「证据强制」这一环。
- `check_mutation_testing` 作为第 18 项机械检查，与既有 17 项并列，全部静态可查、退出码可判。
