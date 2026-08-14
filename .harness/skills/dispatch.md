# Dispatch Skill — 开发任务统一入口 + 工作量分级路由

> **统一入口**：所有开发任务（实现/修复/开发/新增/修改等）必须先经本 Skill 入口，完成 **工作量自动分级（S/M/L）** 后路由到对应执行机制。
> CLAUDE.md 硬性约束 #7：禁止绕过本入口直接开发。

## 服务名映射（中英文均可）

> **权威数据源**：`.harness/registry/services.json`（由 `build-service-registry.sh` 自动扫描 `services/` 生成，含 `name` / `displayName` / `hasApi` / `hasRpc`）。**Step 1 解析服务名时，先读该文件**，以 `name`（英文目录名）与 `displayName`（中文名）做权威匹配；下表仅补充自然语言别名/简称，服务名以 registry 为准。

```bash
# Step 1 解析参数前，读取服务注册表（权威来源）
cat .harness/registry/services.json
```

| 中文别名（自然语言解析用） | 服务名（registry.name） |
|--------|-----------|
| 用户服务 / 用户 | `user-service` |
| 认证服务 / 认证 / 鉴权 | `auth-service` |
| 权限服务 / 权限 | `permission-service` |
| 文件服务 / 文件 | `file-service` |
| AI服务 / AI模型 / 模型服务 | `ai-model-service` |
| 主数据服务 / 主数据 | `master-data-service` |
| 审核服务 / 内容审核 / 审核 | `moderation-service` |
| 社区枢纽服务 / 社区 / 枢纽 | `community-hub-service` |
| 监控服务 / 监控 | `monitoring-service` |

> **新增服务**：运行 `bash .harness/scripts/build-service-registry.sh` 更新 registry/services.json；如该服务有中文别名需求，在上表补一行。禁止硬编码 registry 中不存在的服务名。

## 触发条件

用户说以下任意自然语言时调用此技能（中英文均可）：
- "派发 <服务名> 实现/修复/开发 <任务>"
- "启动 <服务名> 子 Claude"
- "让 <服务名> 去 <任务>"
- "实现/开发/新增/修复/修改/优化/重构/处理/做 <任务>"
- "/dispatch <服务名> <task>"

> **判断是否开发任务**：涉及代码逻辑/配置/文档改动 → 开发任务，进本 Skill 分级；纯问答/查询 → 直接回答，不适用本 Skill。

## 两种模式

### 模式一：全流程管线（默认，按工作量分级路由）

默认行为。进入后先执行 **Step 2 工作量分级**，按 S/M/L 路由：

| 分级 | 执行方式 | QA | Review |
|------|---------|:--:|:--:|
| **S（轻量）** | `harness-pipeline.js`（轻量，`workload:"S"`） | ✅ 18项 | ❌ 跳过 |
| **M（单服务）** | `harness-pipeline.js`（全流程） | ✅ 18项 | 按 taskType |
| **L（跨服务）** | `harness-spec-pipeline.js`（全流程 0-6，每阶段 HITL） | ✅ 每服务 | ✅ 每服务 3视角 |

> **路由口径**（2026-08-14 统一）：S/M 直连 `harness-pipeline.js`，L 走 `harness-spec-pipeline.js`（与 Step 4 一致）。spec-pipeline 的 S/M 短路仅作为「绕过 Owner 直接调」的兜底路径。

QA 阶段默认只扫描本次 git diff 变更的文件。用户说以下词时切换为全量扫描：
- **"全量" / "全量扫描" / "完整检查" / "全面检查"** → QA 使用 `--full`

管线会自动循环：Generator → QA → Review，任一失败回到 Generator 修复重来，最多 3 轮。全部通过后通知用户。

### 模式二：仅开发（快速，需显式触发，覆盖自动分级）

跳过 QA 和 Review，仅派发开发任务。仅在用户明确要求跳过时使用。**这是唯一允许无 QA 的路径。**

触发词：**"快速" / "仅开发" / "跳过审查" / "不用审查"**

行为：仅派发 Generator Agent，完成后告知用户"你可以说'审查 <服务名>'进入下一步"

> **派发方式约束**：`Agent` 工具直接派发**仅用于模式二（快速）**。S/M 级一律走 `Workflow` 管线；L 级走 OpenSpec 子 Agent 序列。禁止用 Agent 直派替代管线。

## 流程

### Step 1: 解析参数

从用户输入中提取：
- `service_name` — 服务名（如 `ai-model-service`、`user-service`、`auth-service`）
- `task` — 任务描述

### Step 2: 工作量分级评估（自动，禁止跳过）

派发前先判定工作量。分级是后续所有路由的唯一依据。先做纯文案判定（2.1），再做 S/M/L 评估（2.2）。

#### 2.1 纯文案/配置判定（跳过 Pipeline 特例）

任务**只**涉及以下内容 → 路由=跳过Pipeline，不进 S/M/L 分级：
- 纯文案/注释/日志文案/README/CHANGELOG 措辞
- 配置文件值/环境变量默认值/yml/yaml/json 值（不涉及代码逻辑）
- 不需要编译验证的改动

命中即：直接 Edit + 对应构建验证，完成后照常记录，不进派发。

#### 2.2 评估信号表（逐项自动判断）

| # | 信号 | 判定方式（客观可查） | S(轻量) | M(单服务) | L(跨服务) |
|---|------|--------------------|:---:|:---:|:---:|
| A | 涉及服务数 | 用 registry 对照任务文本，数 `services/*` + `web/*` 目录 | 1 | 1 | ≥2 |
| B | 涉及 `api-proto/` | 任务含"proto/接口/gRPC/契约/字段"等词 | 否 | 否 | 是 |
| C | 涉及 `common/` | 任务含"common/共享/公共库"等词 | 否 | 否 | 是 |
| D | 预估文件数 | 实现涉及文件数（新功能按逻辑复杂度估算） | ≤1 | 2–5 | >5 |
| E | 预估行数 | 核心逻辑改动量 | ≤20 | 20–200 | >200 |
| F | 新增公开 API | 新 handler / endpoint / gRPC 方法 / 前端路由 | 否 | 否 | 是 |
| G | 架构决策 | 数据模型 / 表结构 / 服务拆分 / 依赖方向 / 技术选型 | 否 | 否 | 是 |
| H | 需求清晰度 | 是否需澄清、是否有多种实现 | 清晰 | 清晰/轻微模糊 | 模糊/多解 |

#### 2.3 判定规则

- **任一 L 信号命中**（A≥2、B=是、C=是、D>5、E>200、F=是、G=是、H=模糊）→ **L（跨服务全流水线）**
- **全部满足 S 条件**（A=1 且 B=否 且 C=否 且 D≤1 且 E≤20 且 F=否 且 G=否 且 H=清晰）→ **S（轻量）**
- **其余 → M（单服务流水线）**

> S 条件与 owner-agent §路径选择「轻量 Pipeline」完全一致（单文件≤20行、不涉及 Proto/common、不新增公开 API），两处共用同一套条件，避免逻辑漂移。

#### 2.4 分级输出格式（首条响应必输出）

```
## 工作量分级
- 分级: S / M / L
- 命中信号: A=单服务 B=否 C=否 D=1文件 E=≤20行 F=否 G=否 H=清晰
- 理由: <一句话，如"单服务+单文件≤20行+不涉及Proto/common+不新增API">
- 路由: 轻量Pipeline / Pipeline / OpenSpec→N×Pipeline
- QA: ✅18项 | Review: 跳过 / 3视角
- 涉及服务: <列表>
```

### Step 3: 验证服务存在

确认 `services/<service_name>/CLAUDE.md` 存在。如果不存在，告知用户该服务尚未配置子 Claude。
（分级在验证之前完成——服务未配置时仍可给出 L 级提示。）

### Step 4: 按分级路由派发

**S 级（轻量 Pipeline）**：
```javascript
Workflow({ scriptPath: ".harness/workflows/harness-pipeline.js",
           args: { serviceName: "<中文名>", serviceDir: "services/<dir>", task: "<任务>", workload: "S" } })
```

**M 级（Pipeline，默认全流程）**：
```javascript
Workflow({ scriptPath: ".harness/workflows/harness-pipeline.js",
           args: { serviceName: "<中文名>", serviceDir: "services/<dir>", task: "<任务>" } })
```

**L 级（跨服务 → spec-pipeline 全流程自动化）**：
启动 `harness-spec-pipeline.js` 全流程 Workflow（规范驱动）：自动走 路径选择→需求分析→需求评审→架构设计→Proto→编码→集成归档，每阶段末 HITL 暂停等用户确认后 `resumeFromRunId` 续跑。阶段 5 编码内部 HITL 委托 Owner 启动 N×Workflow（复用 harness-pipeline.js）。详见 owner-agent §4「全流程自动化（spec-pipeline）」。

**跳过级（纯文案/配置）**：直接 Edit + build 验证，不进派发。

**快速（模式二，用户显式覆盖）**：仅用 `Agent` 工具派发 Generator（prompt 见下文"构造派发提示词"节）。

### Step 5: 通知用户

告知用户：
- 分级结果与路由（S/M/L → 对应执行方式）
- Workflow/Agent 已启动到哪个服务
- 它在做什么任务
- 完成后会自动通知

## 构造派发提示词（模式二"快速"专用）

派发给子 Agent 的 prompt **必须包含以下标准结构**：

```
你是 <service-name> 的开发 Agent。

## 启动上下文（必须先读，顺序重要）
1. 阅读 services/<service-name>/CLAUDE.md — 理解角色定位、关键规则、全局公约、常用命令
2. 阅读 services/<service-name>/docs/design.md — 理解数据模型、业务流程、接口设计（如存在）
3. 阅读 services/<service-name>/CHANGELOG.md — 了解近期变更历史和已知待办
4. 读取 `.harness/knowledge/memory/MEMORY.md`（全局经验索引），根据任务关键词精读相关记忆文件
5. 读取 `services/<service-name>/.harness/knowledge/memory/MEMORY.md`（服务特有经验，如果存在）
6. 如果存在 services/<service-name>/_review.md — 阅读审查报告，修复其中的问题（修复任务时）

## 全局公约提醒
- Proto 变更必须在 api-proto/ 中操作，修改后告知用户切换到全局 Claude 执行 make generate
- 服务间通信仅通过 gRPC（etcd 服务发现），禁止直连其他服务数据库
- 所有 int64 ID 字段在 Proto 中加 [jstype = JS_STRING]，REST API 中加 json:",string"
- 不修改 common/ 和 api-proto/（需要全局 Claude 评估影响）
- **提交前必须运行 `bash .harness/skills/qa/scripts/harness-checks.sh --service <服务目录名>`，有 FAIL 则不可提交**

## 任务
<用户的任务描述，保持原文>

## 完成标准
- 代码通过 go build ./...（Go 服务）
- 代码通过 go test ./...（如有测试）
- **运行 `bash .harness/skills/qa/scripts/harness-checks.sh --service <服务目录名>` 全部 PASS（无 FAIL）**
- 更新 services/<service-name>/CHANGELOG.md — 记录做了什么、为什么、影响范围
- 如果涉及 Proto 变更 → 告知用户切换到全局 Claude，不要自己修改 api-proto/
- 如果涉及 common/ 变更 → 告知用户（需要全局评估影响）
- 不修改 .claude/、reviewers/ 目录（那是 Harness 基础设施）

## 产出
- 代码变更（通过 Edit/Write/Bash）
- CHANGELOG.md 更新
- 简要总结做了什么
```

## 审查闭环（修复任务专用）

如果任务是"修复 _review.md 中的问题"，修复完成后提醒用户：

> 修复完成。需要我派发 Reviewer 重新审查吗？

## 示例

用户说：**"修复 ai-model-service 的 2 个 CRITICAL"**（来自 _review.md）

先输出分级，再派发：

```
## 工作量分级
- 分级: M
- 命中信号: A=单服务 B=否 C=否 D=2文件 E=40行 F=否 G=否 H=清晰
- 理由: 单服务多文件小改动，非 S 非 L
- 路由: Pipeline
- QA: ✅18项 | Review: ✅3视角
- 涉及服务: ai-model-service
```

```
Workflow({ scriptPath: ".harness/workflows/harness-pipeline.js",
           args: { serviceName: "ai-model-service", serviceDir: "services/ai-model-service",
                   task: "修复 _review.md 中列出的 2 个 CRITICAL 问题：1. checkmodelhealthlogic.go — API Key 未解密即使用；2. deletemodelconfiglogic.go — 软删除返回 success=false" } })
```

## 自动派发指令格式

当 Harness Loop 以 `--auto-dispatch` 模式运行时，会输出 `[DISPATCH]` 指令行：

```
[DISPATCH] id=<task-id> service=<english-name> label=<chinese-label> dir=services/<dir> task=<title> [workload=<S|M|L>]
```

SessionStart agent 应解析每条 `[DISPATCH]` 指令，**先按工作量分级再启动对应 Pipeline**：

- 指令行显式带 `workload=` 字段（来自任务 frontmatter）→ 优先后者
- 否则按默认分级表：

| `[DISPATCH]` type | 默认分级 | 理由 |
|------|:---:|------|
| `chore` / `debt` | **S** | 机械性/已知修复模式，QA 18 项即可，Review 可跳过 |
| `bug` / `feature`（P0） | **M** | 高危或新功能，保守走全流程，安全优先 |

启动对应 Pipeline：
```
Workflow({ scriptPath: ".harness/workflows/harness-pipeline.js",
           args: { serviceName: "<label>", serviceDir: "<dir>", task: "<title>",
                   workload: "<S|M，仅当分级为S时传>" } })
```

派发完成后，更新任务状态：
```bash
bash .harness/scripts/harness-tasks.sh status --id <task-id> --status review
```

自动派发仅适用于 P0 + source: qa|review|sensor|github + triage: auto-fixable 的任务。
source: human 或 triage: needs-human 的任务不自动派发，需等待人工确认。
