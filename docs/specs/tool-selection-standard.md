# AI 开发团队工具选择标准

> 配套文档：[AI 软件开发团队设计方案](./ai-dev-team-design.md)
> 
> 适用范围：全局 Claude 实例，每次收到开发需求时参考

---

## 一、判断流程（每次需求到达时的决策树）

```
收到开发需求
  │
  ├─ 纯问答？咨询建议？
  │     → 我直接回答，不用工具
  │
  ├─ Bug 修复（单文件、已知原因、改动<10行）？
  │     → 我直接 Edit（如：字段名修正、路径修正、配置调整）
  │     → 验证：build 通过
  │     → 记录：写入对应服务的 .harness/memory/
  │
  ├─ 单服务、单文件、改动明确？
  │     → 派发 1 个 Dev Agent 到该服务目录
  │     → 验证：build 通过
  │
  ├─ 单服务、多文件（>3）或需理解业务逻辑？
  │     → OpenSpec 轻量版：先写 proposal.md（需求+影响+任务）
  │     → 派发 1 个 Dev Agent，逐项执行
  │     → 验证：build + 功能测试
  │
  ├─ 单服务、新功能/大改动？
  │     → OpenSpec 完整版：proposal → design → tasks
  │     → 派发 Dev Agent → QA Agent → Reviewer
  │     → 验证：build + test + review
  │
  ├─ 跨服务（后端+前端、多个微服务）？
  │     → OpenSpec 完整版
  │     → Proto 变更：我（全局 Claude）处理
  │     → Workflow 并行：Dev(svc1) ∥ Dev(svc2) ∥ Dev(web)
  │     → QA(svc1) ∥ QA(svc2) ∥ QA(web)
  │     → Reviewer（全局）
  │     → 集成验证（我）
  │
  └─ 批量修复、已知问题清单、重复性任务？
        → Ralph 自主循环
        → 写 fix_plan.md → Ralph 逐项执行
        → 熔断器保护 + 自动重试
```

---

## 二、工具矩阵

| 工具 | 触发条件 | 投入 | 产出 | 适用规模 |
|------|---------|------|------|:---:|
| **我直接改** | 单文件<10行、Bug修复、配置、路径修正 | 1分钟 | 代码改动 | 极小 |
| **Dev Agent** | 单服务多文件、前端页面、API端点 | 2-5分钟 | 代码+CHANGELOG | 小 |
| **OpenSpec** | 新功能、跨层改动、需设计评审 | 5-15分钟（设计） | proposal+design+tasks+specs | 中 |
| **Workflow** | 跨服务并行、前后端同时开发 | 5-15分钟（执行） | 多服务并行交付 | 大 |
| **Ralph 循环** | 批量修复、已知清单、需自主迭代 | 30分钟+ | 逐项完成清单 | 批量 |

---

## 三、各工具启用判据

### 我直接改 → 不派发

条件（全部满足）:
- 改动在 1-2 个文件内
- 每文件改动 < 10 行
- 原因明确（字段名错误、路径错误、类型不匹配）
- 不需要理解业务逻辑

示例:
- ✅ 修 crypto.ts 字段名 public_key → publicKey
- ✅ 改 pages.json 路由路径
- ✅ 修正 AES_KEY 格式
- ❌ 新增登录页面（需 Agent）
- ❌ 创建 community-hub-service（需 Workflow）

### Dev Agent → 单服务多文件

条件:
- 改动限定在 1 个服务目录内
- 涉及 3+ 文件
- 需要理解该服务的业务逻辑
- 不需要跨服务协调

派发方式:
```
Agent(subagent_type="general-purpose", 
      prompt="详细任务描述 + 文件路径 + 规范要求")
```

验证: `go build ./...` / `npm run build:h5`

### OpenSpec → 先设计再开发

条件（任一满足）:
- 新功能、新增数据库表
- 涉及前后端两层
- 需要设计评审
- 团队其他成员需要理解设计

产出目录:
```
openspec/changes/<name>/
  ├── proposal.md    ← 为什么做、影响哪些服务
  ├── design.md      ← 技术决策、数据模型、API契约
  ├── tasks.md       ← 按服务拆分的任务清单
  └── specs/         ← 验收场景
```

何时跳过: 修复性改动、已有明确 specs 参照、纯前端页面开发

### Workflow → 跨服务并行

条件（任一满足）:
- 涉及 2+ 独立服务
- 后端+前端需要并行开发
- Proto 变更 + 服务实现 + 前端对接

核心函数:
- `parallel([fn1, fn2])` → 多个 Agent 同时执行
- `pipeline(items, stage1, stage2)` → 逐阶段流水线

### Ralph → 批量自主执行

条件（全部满足）:
- 有明确的 fix_plan.md 任务清单
- 每个任务独立可验证（build/test）
- 需要多次迭代循环
- 任务数 > 5 个

不适用: 需求模糊、任务高度耦合、需要人类实时审核

---

## 四、完整流程速查表

| 需求类型 | 阶段1 | 阶段2 | 阶段3 | 阶段4 | 阶段5 |
|---------|:---:|:---:|:---:|:---:|:---:|
| **修 Bug** | — | — | 我 Edit | 我 build | 写入 memory |
| **小功能**（单服务） | — | — | Dev Agent | 我 build | CHANGELOG |
| **中功能**（前后端） | OpenSpec proposal | OpenSpec design | Dev×2 (parallel) | QA×2 | Reviewer |
| **大功能**（新服务） | OpenSpec proposal | design + Proto(我) | Dev×N + Frontend | QA×N | Reviewer + 集成 |
| **批量修复** | 写 fix_plan.md | Ralph 循环 | 我验收 | — | — |
