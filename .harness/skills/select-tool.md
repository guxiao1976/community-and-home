# select-tool

> **绕过入口直接改代码已废弃**，所有代码改动统一走 `dispatch` Skill（S/M/L 分级路由）。本 Skill 只负责判定"是不是开发任务 / 是否批量"。

## 触发条件

收到任何开发需求时，**在动手之前**必须先执行本 Skill 判断用哪种工具。

触发词：任何开发需求、Bug修复、功能新增、批量改动、代码变更、需求实现

## 决策流程

按以下决策树逐层判断，选择第一个匹配的分支：

```
收到开发需求
  │
  ├─ 纯问答/咨询/解释？
  │     → 直接回答，不调用任何工具
  │
  ├─ 批量修复、已知问题清单、重复性任务（>5 项）？
  │     → 写 fix_plan.md
  │     → 启动 Ralph 自主循环逐项执行
  │     → 每一项仍经 dispatch 分级路由（S 轻量 / M 全流程），Ralph 只做调度不绕过门禁
  │
  └─ 任何代码/配置/文档改动（其余所有）？
        → 调用 dispatch Skill（统一入口，CLAUDE.md 约束 #7）
        → dispatch 自动做 S/M/L 工作量分级并路由：
            S      → 轻量（spec-pipeline 短路到阶段 5 编码：QA 15项，跳过 Review）
            M      → 全流程（spec-pipeline：QA 15项 + Review 按 taskType）
            L      → spec-pipeline 全流程自动化（0-6 阶段，每阶段 HITL）
            跳过级  → 纯文案/配置 → 直接 Edit + build 验证
        → 用户显式"快速/仅开发/跳过审查" → dispatch 模式二（仅开发，无 QA）
```

## 工具速查

| 工具 | 适用规模 | 触发条件 |
|------|:---:|------|
| **dispatch（统一入口）** | 全部 | 任何代码/配置/文档改动，内部 S/M/L 分级路由 |
| **spec-pipeline** | S/M/L 级 | dispatch 按分级启动 `harness-spec-pipeline.js`（全流程自动化，阶段 5 编码委托 harness-pipeline） |
| **Ralph 循环** | 批量 | 已知清单 >5 项、需自主迭代（每项仍分级） |

## 反例（常见误判）

| 需求 | ❌ 错误选择 | ✅ 正确选择 | 原因 |
|------|-----------|-----------|------|
| 改一个字段名 | 绕过入口直接 Edit | dispatch（分级为 S） | 所有代码改动统一走入口，S 级仍需 QA |
| 新增登录页面 | 绕过入口派发 Agent | dispatch（分级为 M） | 涉及多文件、需理解业务 |
| 创建新服务 | 绕过入口 | dispatch（分级为 L） | 跨多层、需 Proto + 脚手架 |
| 批量迁移错误码 | 逐个 Edit | Ralph 循环 | 重复性任务 >5 项（每项仍分级） |
| 问"怎么配置 etcd" | 启动 Agent | 直接回答 | 纯咨询 |

## 下一步

决策完成后，根据分级结果：
- **dispatch（S/M/L 级）** → `Workflow harness-spec-pipeline.js`（全流程自动化；S 级短路阶段 1-4 直接阶段 5 编码，L 级走完整 0-6）
- **跳过级（纯文案/配置）** → 直接 Edit + build 验证
- **Ralph** → 写 fix_plan.md → 启动 Ralph 循环（每项仍走 dispatch 分级）

## 关联资源

- 统一入口 / 分级路由：`.harness/skills/dispatch.md`
- 全流程自动化：`.harness/workflows/harness-spec-pipeline.js`（阶段 5 编码委托 harness-pipeline.js）
- Ralph 配置：`.ralphrc`
