# Harness Spec Pipeline 完整流程图

> 规范驱动全流程自动化：从需求到交付的端到端流程 · 最后更新 2026-08-14

---

## 概述

本文档展示一个需求从接收到最终交付的**全流程自动化**流程（由 `harness-spec-pipeline.js` 编排）。每个阶段：
- **自动编排**：Workflow 依次驱动各阶段（dispatch → 需求分析 → 评审 → 架构 → Proto → 编码 → 归档）
- **HITL 暂停**：每阶段末暂停等用户拍板（`need_input` → Owner 问用户 → `resumeFromRunId` 续跑）
- **编码流水线**：阶段 5 委托 `harness-pipeline.js`（Generator → QA → Review）

> **双层流水线**：全流程 = `harness-spec-pipeline.js`；编码流水线 = `harness-pipeline.js`（阶段 5 内部）。

---

## 案例需求

**需求**: 在社区枢纽服务新增"紧急联络人"功能，涉及后端 API + 前端管理页面。

**涉及服务**:
- `services/community-hub-service` (后端 API + RPC)
- `web/pc` (管理后台前端)

**技术栈**:
- Proto 定义（需新增消息类型）
- Go 后端实现
- Vue 3 前端实现

---

## 完整流程概览（spec-pipeline 自动编排）

```
用户输入需求（Workflow args: {change, task}）
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 0: dispatch 工作量分级（自动判定）                         │
│ ├─ 判定: S/M/L（按 dispatch.md 信号表 A-H）                    │
│ ├─ S/M → 短路阶段 1-4，直接阶段 5（轻量编码）                  │
│ ├─ L → 走完整 OpenSpec 全流程                                 │
│ └─ 产出: request.md（分级 + 路由 + 涉及服务）                  │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 1: 需求分析（自动 + HITL）                               │
│ ├─ 澄清: brainstorming 产出问题清单 → ⏸️ stage1_clarify 等用户拍板│
│ ├─ 分析: requirement-analyst 产出 proposal + specs            │
│ └─ 门禁: 追溯表全✅ + 无占位符                                 │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 2: 需求评审（自动 + HITL）                               │
│ ├─ 3 视角并行 (coverage/structure/clarity) → 投票 2/3         │
│ ├─ 0/3 → 自动回退阶段 1（≤3 轮）→ 超限 stage2_escalate 升级人工│
│ └─ 通过 → ⏸️ stage2_done 等用户裁决                            │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 3: 架构设计（自动 + HITL）                               │
│ ├─ architecture-designer 产出 design + tasks                  │
│ ├─ 门禁: 零占位符 + TDD 步骤                                  │
│ └─ ⏸️ stage3_done 等用户确认服务归属 + Proto 清单              │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 4: Proto 变更（HITL）                                    │
│ ├─ 解析 tasks.md「全局/Proto」段 → ⏸️ stage4_proto             │
│ └─ Owner 执行 make ci（lint + breaking + generate）→ resume   │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 5: 编码+测试（HITL 委托）                                │
│ ├─ ⏸️ stage5_dispatch: Owner 并行启动 N×harness-pipeline.js    │
│ │   └─ 每服务: Generator → QA(15项) → (Debug) → Review        │
│ ├─ 聚合: 全部 PASS → ⏸️ stage5_done 按置信度确认               │
│ └─ 编码流水线 = harness-pipeline.js（Generator→QA→Review 循环） │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 6: 集成归档（自动 + HITL）                               │
│ ├─ 门禁: 全链路 build/vet + QA/Review 归档 impl/              │
│ ├─ 生成 summary.md + 更新 INDEX.md                            │
│ └─ ⏸️ stage6_done 最终交付确认 → 批准归档 → pass              │
└───────────────────────────────────────────────────────────────┘
    ↓
最终交付（status: pass）
```

---

## HITL 暂停点（6 个，每阶段用户参与）

| 暂停点 | 阶段 | 用户确认内容 |
|--------|------|------------|
| `stage1_clarify` | 1 | brainstorming 澄清问题拍板（边界/方案/安全） |
| `stage2_done` / `stage2_escalate` | 2 | 评审裁决（进入 / 回退 / 放宽阈值 / 终止） |
| `stage3_done` | 3 | 服务归属 + Proto 清单确认 |
| `stage4_proto` | 4 | Owner 执行 make ci 后确认 |
| `stage5_dispatch` / `stage5_done` | 5 | 委托启动 N×Pipeline + 编码确认 |
| `stage6_done` | 6 | 最终交付归档确认 |

## resume 机制

每个暂停点返回 `need_input`（含 `ctx`）。Owner resume：

```
Workflow({scriptPath: "harness-spec-pipeline.js", args: {
  change, task, resumeFromRunId,
  resumeState: <上次返回的 ctx 原样传入>,
  resumeWith: { decisions: { <对暂停点问题的决策> } }
}})
```

详情见 owner-agent.md「如何 resume」。

---

## 失败路由（自动回退）

| 失败类型 | 回退目标 |
|---------|---------|
| 方案不可行 / 用户否决 | 阶段 1（重新分析） |
| 评审 ≤1/3 | 阶段 1（≤3 轮，超限升级人工） |
| 设计不合理 | 阶段 1 |
| Proto ci 失败 | 阶段 4 重试 |
| 编码某服务 FAIL | 阶段 5 重跑该服务（≤3 轮） |
| 集成门禁失败 | 阶段 6 修复重试 |
