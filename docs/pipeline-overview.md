# Community-Home 开发流水线全景图

## 📐 总体架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                     Owner Agent（全局协调层）                          │
│                      纯编排器 + 质量把关                               │
├─────────────────────────────────────────────────────────────────────┤
│  路径选择 → 子Agent派发 → 产出验收 → Go/No-Go 裁决 → Proto管理 → 归档  │
└─────────────────────────────────────────────────────────────────────┘
                              ↓
    ┌─────────────┬──────────┴──────────┬─────────────┐
    ↓             ↓                      ↓             ↓
【需求分析】   【架构设计】          【编码管线】    【集成归档】
 子Agent       子Agent           Workflow × N      Owner内联
```

---

## 🎯 完整流程（OpenSpec 模式 - 跨服务大需求）

### 阶段 0：工具选择（路径判断）

**技术**：Owner Agent 内联判断
**文件**：`.harness/skills/select-tool.md`

**决策树**：
```
用户需求
  ├─ 代码量 ≤10 行？                    → 直接 Edit（快速修复）
  ├─ 单服务、无架构变更？                → Dev Agent（子服务 Claude）
  ├─ 跨服务、需架构设计？                → OpenSpec（6 阶段流程）
  └─ 前端视觉交互？                      → Ralph（前端设计工具）
```

**产出**：`.harness/changes/<change>/request.md`

---

### 阶段 1：需求分析

**技术栈**：
- **执行方式**：子 Agent（`subagent_type: "general-purpose"`）
- **指令文件**：`.harness/agents/subagents/requirement-analyst.md`
- **上下文注入**：
  - 项目规范：`项目编码规范.md`
  - 工程结构：`工程结构.md`
  - 业务流程：`.harness/knowledge/business-flows.md`
  - 历史记忆：`.harness/knowledge/memory/` (触发词匹配)

**流程**：
1. 理解需求背景
2. 拆解为原子需求
3. 每个需求写 Spec（影响范围、数据模型、接口变更）
4. 构建追溯表（需求 → Spec → 服务 → 文件）
5. Self-Review 检查清单

**产出**：
- `proposal.md` - 需求提案
- `specs/*/spec.md` - 各需求详细规格

**门禁**：
- ✅ 追溯表 100% 覆盖
- ✅ Self-Review 全 PASS

---

### 阶段 2：需求评审

**技术栈**：
- **执行方式**：3 个子 Agent 并行
- **指令文件**：`.harness/skills/review.md`（计划评审模式）
- **评审视角**：
  1. **Coverage Reviewer** - 完整性（需求是否全覆盖）
  2. **Structure Reviewer** - 结构性（Spec 是否清晰）
  3. **Clarity Reviewer** - 明确性（是否有歧义）

**流程**：
```javascript
parallel([
  agent("Coverage 评审", {phase: "需求评审", schema: REVIEW_SCHEMA}),
  agent("Structure 评审", {phase: "需求评审", schema: REVIEW_SCHEMA}),
  agent("Clarity 评审", {phase: "需求评审", schema: REVIEW_SCHEMA})
])
```

**产出**：
- `review/spec_review_coverage_v1.md`
- `review/spec_review_structure_v1.md`
- `review/spec_review_clarity_v1.md`

**门禁**：
- ✅ 2/3 APPROVED → 通过
- ❌ REVISION → 回到阶段 1（最多 3 轮）
- ❌ REJECTED → 终止

---

### 阶段 3：架构设计

**技术栈**：
- **执行方式**：子 Agent
- **指令文件**：`.harness/agents/subagents/architecture-designer.md`
- **上下文注入**：
  - Proposal + Specs（阶段 1 产出）
  - 系统架构：`.harness/knowledge/design.md`
  - 服务图谱：`services/*/docs/graph-context.md` (Neo4j 自动生成)
  - Proto 规范：`.harness/rules/Proto管理规范.md`

**设计原则**：
1. **记忆驱动设计**：查询 memory 索引，注入相关经验
2. **零占位符原则**：所有任务必须完全展开，禁止 "TODO" 占位
3. **TDD 强制**：每个任务必须包含测试步骤

**流程**：
1. 确定服务边界（哪些服务需要修改）
2. 设计数据模型（表结构、Redis key）
3. 设计接口（Proto 定义、REST API）
4. 拆分实现任务（每服务一个 task list）
5. 依赖分析（服务间调用关系）

**产出**：
- `design.md` - 架构设计文档
- `tasks.md` - 实现任务清单

**门禁**：
- ✅ 记忆注入完整
- ✅ 零占位符（100% 展开）
- ✅ TDD 步骤明确

---

### 阶段 4：Proto 变更（如需要）

**技术栈**：
- **执行方式**：Owner Agent 亲自执行（硬性规则）
- **工具链**：
  - `buf` - Proto lint + breaking change 检测
  - `protoc` + `protoc-gen-go` / `protoc-gen-go-grpc` - 代码生成
  - `protoc-gen-ts` - TypeScript 定义生成

**流程**：
```bash
cd api-proto
# 1. 修改 Proto 文件
vim proto/<service>/<version>/<message>.proto

# 2. 运行 CI 检查
make ci
  ├─ make lint           # buf lint（规范检查）
  ├─ make breaking-check # buf breaking（破坏性检查）
  └─ make generate       # 生成 Go + TS 代码
```

**门禁**：
- ✅ `buf lint` 全 PASS
- ✅ `buf breaking` 无破坏性变更（或经过 HITL 确认）
- ✅ 代码生成成功

**硬性约束**：
- Snowflake ID → `[jstype=JS_STRING]`
- 所有字段必须有注释
- 禁止 `repeated` 嵌套超过 2 层

---

### 阶段 5：编码 + 测试管线（核心）

**技术栈**：
- **执行方式**：N 个 Workflow 并行（每服务 1 个）
- **编排脚本**：`.harness/workflows/harness-pipeline.js`（1186 行，模块化）
- **模块组成**：
  - `generator.js` - TDD 编码 Agent
  - `qa.js` - 机械化检查
  - `review.js` - 3 视角代码审查
  - `debug.js` - 根因分析
  - `harness-pipeline-core.js` - 编排循环

#### 5.1 Generator Agent（TDD 驱动编码）

**技术**：
- **Memory 驱动**：查询 `.harness/knowledge/memory/.memory-index.json`
- **TDD 强制**：RED → GREEN → REFACTOR
- **上下文注入**：
  - 服务规范：`services/<name>/CLAUDE.md`
  - 图谱上下文：`services/<name>/docs/graph-context.md`
  - 相关记忆：触发词匹配的 memory 文件

**TDD 三段式**：
```
Phase 1 (RED):
  ✓ 写失败测试（assert 期望行为）
  ✓ 运行测试 → 红色（确认测试有效）

Phase 2 (GREEN):
  ✓ 实现最简代码
  ✓ 运行测试 → 绿色（功能通过）

Phase 3 (REFACTOR):
  ✓ 重构代码（消除重复、提升可读性）
  ✓ 运行测试 → 仍然绿色（确保不破坏）
```

**产出**：实现代码 + 测试代码

#### 5.2 QA Agent（机械化检查 - 15 项）

**技术**：
- **脚本**：`.harness/skills/qa/scripts/harness-checks.sh`
- **模式**：
  - **差分模式**（增量）：只检查变更文件（秒级）
  - **全量模式**（发布）：检查整个服务（分钟级）

**检查项清单**：

| # | 检查项 | 技术 | 阻塞级别 |
|---|--------|------|:-------:|
| 1 | Go 编译 | `go build` | 🔴 FAIL |
| 2 | 单元测试 | `go test` | 🔴 FAIL |
| 3 | 测试覆盖率 | `go test -cover` | 🟡 WARN (30%) |
| 4 | 0/0 假通过检测 | AST 分析 | 🟡 WARN |
| 5 | Logic 文件必须有测试 | 文件存在性检查 | 🔴 FAIL |
| 6 | Context 取消检查 | AST 分析 `context.TODO()` | 🟡 WARN |
| 7 | 事务回滚检查 | AST 分析 `defer tx.Rollback()` | 🟡 WARN |
| 8 | SQL 注入检测 | 字符串拼接模式匹配 | 🔴 FAIL |
| 9 | Proto 规范检查 | `buf lint` | 🔴 FAIL |
| 10 | 硬编码密钥检测 | 正则匹配 `password=` | 🔴 FAIL |
| 11 | 日志敏感信息 | 正则匹配 `log.*password` | 🟡 WARN |
| 12 | Snowflake ID 规范 | AST 分析 `json:",string"` | 🔴 FAIL |
| 13 | 知识图谱新鲜度 | 时间戳检查 | 🟡 WARN |
| 14 | Benchmark 回归 | 对比基线（>50% 退化） | 🔴 FAIL |
| 15 | API 冒烟测试 | `curl` 健康检查 | 🟢 INFO |

**门禁**：
- ✅ 全部 PASS → 下一步
- ❌ 任意 FAIL → 触发 Debug Agent

**产出**：`_qa.md` (JSON 格式可选)

#### 5.3 Debug Agent（根因分析）

**技术**：
- **指令模板**：`.harness/agents/prompts/templates/debug.md`
- **方法论**：Systematic Debugging（4 阶段）

**流程**：
```
Stage 1: 复现问题
  ✓ 复述 QA 失败现象
  ✓ 最小化复现步骤

Stage 2: 收集证据
  ✓ 读取错误日志
  ✓ 检查相关代码
  ✓ 分析调用链

Stage 3: 根因诊断
  ✓ 提出假设
  ✓ 验证假设
  ✓ 定位根本原因

Stage 4: 修复建议
  ✓ 提供修复方案
  ✓ 附加测试验证
```

**产出**：诊断报告 → 回到 Generator（最多 3 轮）

#### 5.4 Review Agent（3 视角并行）

**技术**：
- **指令模板**：`.harness/agents/prompts/templates/review.md`
- **执行方式**：3 个 Agent 并行

**评审视角**：

| 视角 | 关注点 | 维度 |
|------|--------|------|
| **安全架构** | Security + Architecture | 认证授权、注入防护、密钥管理、架构一致性 |
| **规范工程** | Conventions + Engineering | 编码规范、测试规范、错误处理、日志规范 |
| **设计业务** | Design + Business | API 设计、业务逻辑、数据模型、边界情况 |

**9 维度矩阵**：
1. **Security** - SQL 注入、XSS、CSRF、密钥泄露
2. **Architecture** - 服务边界、依赖方向、分层清晰
3. **Conventions** - 命名、格式、注释、目录结构
4. **Engineering** - 测试覆盖、错误处理、日志、性能
5. **Design** - API 一致性、数据模型合理性
6. **Business** - 业务逻辑正确性、边界处理
7. **Correctness** - 算法正确、并发安全
8. **Maintainability** - 可读性、可扩展性
9. **Testing** - 测试充分性、测试质量

**流程**：
```javascript
parallel([
  agent("安全架构 Reviewer", {schema: REVIEW_SCHEMA}),
  agent("规范工程 Reviewer", {schema: REVIEW_SCHEMA}),
  agent("设计业务 Reviewer", {schema: REVIEW_SCHEMA})
])
```

**门禁**：
- ✅ 2/3 APPROVED → 通过
- ❌ REVISION → 回到 Generator（最多 2 轮）

**产出**：
- `_review_security_arch_v1.md`
- `_review_conventions_eng_v1.md`
- `_review_design_business_v1.md`

#### 5.5 编码管线编排逻辑

```
Generator (TDD 编码)
    ↓
QA (15 项机械化检查)
    ├─ PASS → Review
    └─ FAIL → Debug → Generator (最多 3 轮)
        ↓
Review (3 视角并行)
    ├─ 2/3 APPROVED → 完成
    └─ REVISION → Generator (最多 2 轮)
```

**循环控制**：
- QA-Debug 最大 3 轮
- Review 最大 2 轮
- 超限 → 升权到 Owner 人工介入

---

### 阶段 6：集成验证 + 归档

**技术栈**：
- **执行方式**：Owner Agent 内联
- **验证内容**：
  - 全链路编译：`go build ./...`
  - 全链路测试：`go test ./...`
  - Proto 一致性：检查 api-proto 与服务代码

**归档流程**：
1. 移动 QA/Review 报告到 `.harness/changes/<change>/impl/<service>/`
2. 更新 `.harness/changes/INDEX.md`
3. 生成 `summary.md`（终稿）
4. 更新各服务 `CHANGELOG.md`
5. 如有新坑 → 写入 `.harness/knowledge/memory/`

**产出**：
- `summary.md` - 变更总结
- `INDEX.md` - 变更索引更新
- `CHANGELOG.md` - 各服务变更日志

---

## 🛡️ 质量保障体系（三层防护）

### Layer 1：机械化检查（秒级反馈）

**Pre-commit Hook**（本地）：
- Logic 文件必须有测试
- 运行变更包测试
- 代码格式检查
- 静态分析
- **Memory 索引自动构建**

**Harness QA（15 项）**：
- 完整的自动化检查矩阵
- 差分模式（快）+ 全量模式（全）
- JSON 输出可供 CI 集成

### Layer 2：CI/CD 自动化

**Go 后端 CI**（`.github/workflows/test.yml`）：
- 单元测试 + 覆盖率（30% 门禁）
- golangci-lint 静态分析
- Quality Gate（logic 文件测试强制）
- Codecov 集成

**Harness QA CI**（`.github/workflows/harness-qa-check.yml`）：
- 差分检测（仅检查变更服务）
- 15 项机械化检查
- PR 自动评论（失败详情）
- JSON 报告归档

**前端 CI**（`.github/workflows/frontend-ci.yml`）：
- TypeScript 类型检查
- Vitest 单元测试
- Vite 构建验证
- PC + Mobile 并行

**Docker 构建**（`.github/workflows/docker-deploy.yml`）：
- 9 个微服务并行构建
- 推送到 GitHub Container Registry
- 镜像标签策略（branch/tag/sha）
- 构建缓存优化

### Layer 3：AI 驱动审查

**3 视角并行审查**：
- 安全架构 Reviewer
- 规范工程 Reviewer
- 设计业务 Reviewer

**9 维度覆盖**：
- Security / Architecture / Conventions
- Engineering / Design / Business
- Correctness / Maintainability / Testing

**门禁**：2/3 APPROVED 才能合并

---

## 🧪 测试体系

### 测试金字塔

```
        ┌─────────┐
        │  E2E    │  ← Playwright (TODO)
        │  (少量)  │
     ┌──┴─────────┴──┐
     │  Integration   │  ← API 冒烟测试（QA 15）
     │  (适量)        │
  ┌──┴───────────────┴──┐
  │   Unit Tests        │  ← go test（强制）
  │   (大量)            │     TDD 驱动
  └─────────────────────┘
```

### 测试策略

**单元测试**（强制）：
- TDD 驱动：先写测试，再写实现
- Logic 文件 100% 有测试（pre-commit + CI 双重门禁）
- 覆盖率 30% 门禁（可提升）
- Table-driven tests 模式

**集成测试**：
- API 冒烟测试（QA 检查 15）
- gRPC 调用测试
- 数据库操作测试（事务回滚检查）

**E2E 测试**（规划中）：
- Playwright 前端 E2E
- 全链路业务流程验证

### 测试工具链

| 工具 | 用途 | 状态 |
|------|------|:----:|
| `go test` | 单元测试 | ✅ |
| `go test -cover` | 覆盖率统计 | ✅ |
| `go test -bench` | 性能基准 | ✅ |
| `vitest` | 前端单元测试 | ✅ |
| `playwright` | E2E 测试 | 📋 |
| AST 检查器 | 静态分析 | ✅ |

### 测试规范

**强制约束**：
1. `*_logic.go` 必须有 `*_logic_test.go`
2. 测试函数必须有实际断言（不能空函数）
3. 事务操作必须有 `defer tx.Rollback()`
4. Context 必须正确传递（禁止 `context.TODO()`）

**推荐实践**：
1. Table-driven tests
2. Mock 外部依赖
3. Benchmark 热点路径
4. 测试覆盖边界情况

---

## 🔧 支撑技术栈

### 知识管理

**Neo4j 知识图谱**：
- 自动扫描代码生成服务依赖图
- REST 路由、gRPC 接口、数据库表映射
- 实体血缘追踪（Proto → Go → DB）
- `graph-sync.sh` 自动刷新

**Memory 系统**：
- 倒排索引：触发词 → 记忆文件
- Frontmatter 元数据：severity / type / service
- `memory-index-build.sh` 自动构建
- Pre-commit 自动更新索引

**文档体系**：
```
.harness/
├── rules/              ← 硬性约束（很少变）
│   ├── 项目编码规范.md
│   ├── Proto管理规范.md
│   └── 工程结构.md
├── knowledge/          ← 项目知识（按需查询）
│   ├── INDEX.md
│   ├── design.md
│   ├── business-flows.md
│   └── memory/         ← 经验记忆（自动注入）
├── changes/            ← 变更追溯（完整审计链）
│   ├── INDEX.md
│   └── <change>/
│       ├── proposal.md
│       ├── design.md
│       ├── impl/
│       └── summary.md
└── tasks/              ← 任务管理
    ├── BACKLOG.md
    └── MAINTENANCE.md
```

### MCP 工具集成

**GitHub MCP**：
- Issues / PR 管理
- 代码搜索
- 跨仓库协调

**MySQL MCP**：
- 只读查询数据库
- 数据一致性验证
- Migration 检查

### 自动化脚本

| 脚本 | 功能 | 触发方式 |
|------|------|---------|
| `graph-sync.sh` | 同步知识图谱 | post-commit hook |
| `memory-index-build.sh` | 构建 memory 索引 | pre-commit hook |
| `harness-checks.sh` | 15 项机械化检查 | 手动 / CI |
| `harness-tasks.sh` | 任务管理 | 手动 / Loop |
| `install-git-hooks.sh` | 安装 Git hooks | 手动 |
| `build-pipeline.sh` | 构建 harness-pipeline.js | 手动 |

---

## 📊 流水线性能指标

### 反馈速度

| 阶段 | 耗时 | 方式 |
|------|------|------|
| Pre-commit | **5-15 秒** | 本地 Hook |
| QA 差分检查 | **10-30 秒** | 增量扫描 |
| QA 全量检查 | **2-5 分钟** | 完整验证 |
| Generator 编码 | **3-10 分钟** | AI 生成 |
| Review 审查 | **2-5 分钟** | 3 并行 |
| CI 构建 | **2-5 分钟** | GitHub Actions |

### 质量指标

| 指标 | 目标 | 当前 |
|------|:----:|:----:|
| 单元测试覆盖率 | 30% | ✅ 门禁 |
| Logic 文件测试率 | 100% | ✅ 强制 |
| CI 通过率 | >95% | 🎯 (pre-commit 后) |
| QA 检查准确率 | >99% | ✅ |
| Review 召回率 | >90% | 🎯 (3 视角) |

### 自动化程度

| 环节 | 自动化 | 人工介入 |
|------|:-----:|:-------:|
| 代码格式 | ✅ | - |
| 单元测试 | ✅ | - |
| 静态分析 | ✅ | - |
| QA 检查 | ✅ | - |
| 代码审查 | ✅ | 2/3 FAIL 时 |
| 部署 | 🔧 | 待配置 |

---

## 🎯 核心优势

### 1. 确定性验证

- **机械化检查**：15 项自动检查，0 漏检
- **门禁强制**：FAIL 无法绕过
- **追溯完整**：需求 → 设计 → 代码 → 测试 → 审查

### 2. AI + 自动化融合

- **AI 生成 + 机械验证**：Generator 编码 → QA 自动检查
- **AI 审查 + 人类决策**：3 视角并行 → 2/3 投票
- **记忆驱动**：历史经验自动注入，避免重复踩坑

### 3. 上下文隔离

- **子 Agent 独立**：需求分析、架构设计各自干净上下文
- **Workflow 隔离**：每服务独立编码管线
- **Owner 纯编排**：不参与具体工作，只做验收和裁决

### 4. 全流程可追溯

- **变更追溯链**：`.harness/changes/` 完整保留
- **审计日志**：每次 Review 版本递增（v1/v2/v3）
- **知识沉淀**：踩坑经验写入 memory，永久生效

---

## 📈 未来规划

### 短期（Q3 2026）
- [ ] 前端 E2E 测试（Playwright）
- [ ] 部署 pipeline 实际配置（K8s/Docker）
- [ ] 性能监控集成（Prometheus/Grafana）

### 中期（Q4 2026）
- [ ] 覆盖率提升到 60%
- [ ] 自动化回归测试套件
- [ ] API 契约测试（Pact）

### 长期（2027）
- [ ] 混沌工程（故障注入）
- [ ] 智能根因分析（ML 辅助）
- [ ] 自适应质量门禁（动态阈值）

---

**文档版本**：v2.0
**最后更新**：2026-07-11
**维护者**：Community-Home Dev Team
