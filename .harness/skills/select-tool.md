# select-tool

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
  ├─ 单文件、已知原因、改动 < 10 行？
  │     条件：1-2 个文件、原因明确、不需理解业务逻辑
  │     示例：字段名修正、路径修正、配置调整、类型修正
  │     → 直接 Edit，然后 build 验证
  │     → 完成后记录到 .harness/memory/（如踩坑）
  │
  ├─ 单服务、改动明确？
  │     → 使用 dispatch Skill 派发 Dev Agent 到该服务目录
  │     → 验证：build 通过
  │
  ├─ 单服务、新功能或大改动（>3 文件或需理解业务）？
  │     → 先使用 requirement-analysis Skill 产出 proposal.md
  │     → 再使用 architect-design Skill 产出 design.md + tasks.md（如涉及设计决策）
  │     → 派发 Dev Agent → QA Agent → Reviewer
  │     → 验证：build + test + review
  │
  ├─ 跨服务（后端+前端、多个微服务）？
  │     → OpenSpec: proposal → design → tasks
  │     → Proto 变更由全局 Claude 处理（修改 api-proto/ → make generate）
  │     → 并行派发 Dev Agent 到各服务
  │     → QA 验证各服务
  │     → Reviewer 全局审查
  │     → 集成验证
  │
  └─ 批量修复、已知问题清单、重复性任务（>5 项）？
        → 写 fix_plan.md
        → 启动 Ralph 自主循环逐项执行
        → 熔断器保护 + 自动重试
```

## 工具速查

| 工具 | 适用规模 | 触发条件 |
|------|:---:|------|
| **直接 Edit** | 极小 | 单文件 <10 行、Bug修复、配置修正 |
| **Dev Agent** | 小 | 单服务多文件、前端页面、API 端点 |
| **OpenSpec** | 中 | 新功能、跨层改动、需设计评审 |
| **Workflow** | 大 | 跨服务并行、前后端同时开发 |
| **Ralph 循环** | 批量 | 已知清单 >5 项、需自主迭代 |

## 反例（常见误判）

| 需求 | ❌ 错误选择 | ✅ 正确选择 | 原因 |
|------|-----------|-----------|------|
| 新增登录页面 | 直接 Edit | Dev Agent | 涉及多文件、需理解业务 |
| 创建新服务 | Dev Agent | Workflow | 跨多层、需 Proto + 脚手架 |
| 改一个字段名 | Dev Agent | 直接 Edit | 单文件 <10 行 |
| 批量迁移错误码 | 逐个 Edit | Ralph 循环 | 重复性任务 >5 项 |
| 问"怎么配置 etcd" | 启动 Agent | 直接回答 | 纯咨询 |

## 下一步

决策完成后，根据选择的结果：
- **直接 Edit** → 开始修改，build 验证
- **Dev Agent** → 调用 `dispatch` Skill
- **OpenSpec** → 调用 `requirement-analysis` Skill → `architect-design` Skill
- **Workflow** → 编写并行 Workflow 脚本
- **Ralph** → 写 fix_plan.md → 启动 Ralph 循环

## 关联资源

- 派发 Skill：`.harness/skills/dispatch.md`
- 需求分析 Skill：`.harness/skills/requirement-analysis.md`
- 架构设计 Skill：`.harness/skills/architect-design.md`
- Harness Pipeline：`.harness/workflows/harness-pipeline.js`
- Ralph 配置：`.ralphrc`
