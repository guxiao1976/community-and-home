# Dispatch Skill — 派发服务子 Claude

## 服务名映射（中英文均可）

| 中文名 | 英文目录名 |
|--------|-----------|
| 用户服务 / 用户 | `user-service` |
| 认证服务 / 认证 / 鉴权 | `auth-service` |
| 权限服务 / 权限 | `permission-service` |
| 文件服务 / 文件 | `file-service` |
| AI服务 / AI模型 / 模型服务 | `ai-model-service` |
| 主数据服务 / 主数据 | `master-data-service` |
| 审核服务 / 内容审核 / 审核 | `moderation-service` |

## 触发条件

用户说以下任意自然语言时调用此技能（中英文均可）：
- "派发 <服务名> 实现/修复/开发 <任务>"
- "启动 <服务名> 子 Claude"
- "让 <服务名> 去 <任务>"
- "/dispatch <服务名> <task>"

## 两种模式

### 模式一：全流程管线（默认）
默认行为，派发后自动走完 开发→测试→审查 全流程。

触发条件：**用户说"派发"时，默认进入此模式**（无需额外触发词）

行为：使用 Workflow 工具启动 `harness-pipeline` 工作流：
```
Workflow({ scriptPath: ".claude/workflows/harness-pipeline.js",
           args: { serviceName: "<中文名>", serviceDir: "services/<dir>", task: "<任务>" } })
```

QA 阶段默认只扫描本次 git diff 变更的文件。用户说以下词时切换为全量扫描：
- **"全量" / "全量扫描" / "完整检查" / "全面检查"** → QA 使用 `--full`

管线会自动循环：Generator → QA → Review，任一失败回到 Generator 修复重来，最多 3 轮。全部通过后通知用户。

### 模式二：仅开发（快速，需显式触发）
跳过 QA 和 Review，仅派发开发任务。仅在用户明确要求跳过时使用。

触发词：**"快速" / "仅开发" / "跳过审查" / "不用审查"**

行为：仅派发 Generator Agent，完成后告知用户"你可以说'审查 <服务名>'进入下一步"

## 流程

### Step 1: 解析参数

从用户输入中提取：
- `service_name` — 服务名（如 `ai-model-service`、`user-service`、`auth-service`）
- `task` — 任务描述

### Step 2: 验证服务存在

确认 `services/<service_name>/CLAUDE.md` 存在。如果不存在，告知用户该服务尚未配置子 Claude。

### Step 3: 构造派发提示词

派发给子 Agent 的 prompt **必须包含以下标准结构**：

```
你是 <service-name> 的开发 Agent。

## 启动上下文（必须先读，顺序重要）
1. 阅读 services/<service-name>/CLAUDE.md — 理解角色定位、关键规则、全局公约、常用命令
2. 阅读 services/<service-name>/docs/design.md — 理解数据模型、业务流程、接口设计（如存在）
3. 阅读 services/<service-name>/CHANGELOG.md — 了解近期变更历史和已知待办
4. 读取 `.claude/memory/MEMORY.md`（全局经验索引），根据任务关键词精读相关记忆文件
5. 读取 `services/<service-name>/.claude/memory/MEMORY.md`（服务特有经验，如果存在）
6. 如果存在 services/<service-name>/_review.md — 阅读审查报告，修复其中的问题（修复任务时）

## 全局公约提醒
- Proto 变更必须在 api-proto/ 中操作，修改后告知用户切换到全局 Claude 执行 make generate
- 服务间通信仅通过 gRPC（etcd 服务发现），禁止直连其他服务数据库
- 所有 int64 ID 字段在 Proto 中加 [jstype = JS_STRING]，REST API 中加 json:",string"
- 不修改 common/ 和 api-proto/（需要全局 Claude 评估影响）
- **提交前必须运行 `bash scripts/harness-checks.sh --service <服务目录名>`，有 FAIL 则不可提交**

## 任务
<用户的任务描述，保持原文>

## 完成标准
- 代码通过 go build ./...（Go 服务）
- 代码通过 go test ./...（如有测试）
- **运行 `bash scripts/harness-checks.sh --service <服务目录名>` 全部 PASS（无 FAIL）**
- 更新 services/<service-name>/CHANGELOG.md — 记录做了什么、为什么、影响范围
- 如果涉及 Proto 变更 → 告知用户切换到全局 Claude，不要自己修改 api-proto/
- 如果涉及 common/ 变更 → 告知用户（需要全局评估影响）
- 不修改 .claude/、reviewers/ 目录（那是 Harness 基础设施）

## 产出
- 代码变更（通过 Edit/Write/Bash）
- CHANGELOG.md 更新
- 简要总结做了什么
```

### Step 4: 派发

使用 `Agent` 工具派发：
- `subagent_type`: `"general-purpose"`
- `description`: `"<service-name>: <简短任务描述>"`
- `prompt`: 上面构造的完整提示词
- `run_in_background`: `true`（让 Agent 后台执行，完成后通知用户）

### Step 5: 通知用户

告知用户：
- Agent 已派发到哪个服务
- 它在做什么任务
- 完成后会自动通知

## 审查闭环（修复任务专用）

如果任务是"修复 _review.md 中的问题"，修复完成后提醒用户：

> 修复完成。需要我派发 Reviewer 重新审查吗？

## 示例

用户说：**"派发 ai-model-service 修复 _review.md 中的 2 个 CRITICAL"**

派发 prompt：
```
你是 ai-model-service 的开发 Agent。

## 启动上下文（必须先读，顺序重要）
1. 阅读 services/ai-model-service/CLAUDE.md
2. 阅读 services/ai-model-service/docs/design.md
3. 阅读 services/ai-model-service/CHANGELOG.md
4. 读取 `.claude/memory/MEMORY.md`（全局经验索引），根据任务关键词精读相关记忆文件
5. 读取 `services/ai-model-service/.claude/memory/MEMORY.md`（服务特有经验，如果存在）
6. 阅读 services/ai-model-service/_review.md — 这是审查报告，修复其中的 CRITICAL

## 全局公约提醒
- Proto 变更必须在 api-proto/ 中操作，告知用户切换到全局 Claude
- 服务间通信仅通过 gRPC
- 不修改 common/ 和 api-proto/

## 任务
修复 _review.md 中列出的 2 个 CRITICAL 问题：
1. checkmodelhealthlogic.go — API Key 未解密即使用
2. deletemodelconfiglogic.go — 软删除返回 success=false

## 完成标准
- go build ./... 通过
- go test ./... 通过
- CHANGELOG.md 已更新
```
