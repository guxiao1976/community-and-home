---
name: context-budget
description: 上下文预算门禁——阶段边界跑 /context，按 50/70/80 三线决策 compact / handoff；会话卫生（大文件不直读、验证不内联、成果必落盘）。任何 L 级长任务、任何长会话执行前必须遵守。
---

# Context Budget — 上下文预算门禁

> 设计依据：`.harness/docs/context-management.md`。核心认知：**compact 本身需要上下文空间**，必须远低于满时触发。
> 进度要求：长任务执行期必须启动进度心跳（`progress-heartbeat.sh`），否则用户无法区分"运行中"与"卡死"。

## 触发条件

- 任何 **L 级（跨服务）任务**开始前
- 任何**长会话**（预期 >30 分钟 / 涉及多阶段）
- 会话中每次切换阶段前

## 检查点（固定 gate，每个都必须执行 /context 并按表决策）

| # | 检查点 | 时机 |
|---|--------|------|
| 1 | 阶段开始前 | 每个新阶段（dispatch/设计/编码/验证/收尾） |
| 2 | 阶段结束后 | 每个阶段产出落盘后 |
| 3 | 大文件读取前 | 任何 >50KB 文件直读前（`tasks.md`/`design.md` 等） |
| 4 | QA 循环后 | 每次 QA 门禁运行后（QA 是输出大户） |

## 决策表

| 填充率 | 状态 | 动作 |
|--------|------|------|
| ≤50% | 🟢 健康 | 正常继续 |
| 50-70% | 🟡 关注 | **阶段边界强制 /compact**；60%+ 禁止开启新阶段（除非先 compact） |
| 70-80% | 🔴 强制 | **立即 /compact**（无论是否断点）；compact 后仍 >55% → 写 handoff + 停 |
| >80% | ⛔ 硬停 | **不 compact**（可能失败）；写 handoff + 结束会话，`claude -r <session-id>` 新会话续 |

## 会话卫生硬规则（L 级任务强制）

1. **主会话禁止直读 >50KB 文件**：`tasks.md`/`design.md` 等 → 只读相关小节（`sed -n` 片段）或交子 Agent 读。
2. **主会话禁止内联跑验证**（Task 6.x 迁移/E2E）→ 派独立子 Agent，回传验证结果摘要。
3. **主会话禁止内联做 QA FAIL 闭环**（TDD 证据复现等）→ 派子 Agent，回传证据文件路径 + 结论。
4. **成果必落盘**：每阶段产出写 `summary.md` 固定小节（当前阶段/已验收产出/上下文快照/剩余任务）。
5. **进度心跳必开**：长任务开始时 `Monitor` 跑 `progress-heartbeat.sh`；每步骤边界更新 `$CLAUDE_JOB_DIR/tmp/progress.txt`。

## 执行步骤

```text
1. 启动进度心跳（若未开）：
   Monitor({ command: "bash .harness/scripts/progress-heartbeat.sh 300", persistent: true })
2. 进入每个检查点：运行 /context（或 claude -p "/context"），读取填充率。
3. 按决策表处置：
   - compact：在阶段边界执行 /compact（先确认已验收产出已落盘）。
   - handoff：写 summary.md 交接小节 → 结束会话 → 告知用户用 claude -r 续。
4. 每步骤边界：echo "当前: <做什么>" > $CLAUDE_JOB_DIR/tmp/progress.txt
```

## 禁止

- ❌ 填充率未知时开启新阶段（必须先 /context）
- ❌ >70% 仍继续长工具序列（先 compact 或 handoff）
- ❌ 主会话直读 >50KB 文件 / 内联跑验证 / 内联 QA 闭环
- ❌ 无进度心跳的长任务（用户无法判断死活）
