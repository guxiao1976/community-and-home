# Review Skill — 派发代码审查

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
- "审查 <服务名>" — 默认只扫描本次 git diff 变更文件
- "全量审查 <服务名>" / "全面审查 <服务名>" — 全量扫描（`--full`）
- "review <服务名>"
- "检查 <服务名> 的代码"
- "重新审查 <服务名>"
- "/review <服务名>"

## 流程

### Step 1: 解析参数

从用户输入中提取 `service_name`。

### Step 2: 派发审查 Agent

使用 `Agent` 工具派发独立审查 Agent：

- `subagent_type`: `"general-purpose"`
- `description`: `"Review <service-name>"`
- `prompt`:

```
你是一个 Code Reviewer Agent。

## 角色定义（必须先读）
阅读 /home/jiaoxh/my-project/community-home/reviewers/code-reviewer/CLAUDE.md — 这是你的角色定义、审查规则和产出格式。

## 审查目标
审查 services/<service-name>/ 的代码变更。

## 审查步骤
1. 阅读根 CLAUDE.md（全局硬规则、Proto 管理规范）
2. 阅读目标服务 CLAUDE.md（角色、关键规则）
3. 阅读目标服务 docs/design.md（数据模型、业务流程）
4. 阅读目标服务 CHANGELOG.md（近期变更）
5. 获取变更内容（git diff main...HEAD 或审查全部代码）
6. 按 8 个维度逐文件审查
7. 写入 _review.md 到目标服务目录
8. 输出 VERDICT

## 约束
- 只读权限：仅使用 Read、Grep、Glob、Bash（go build、go vet、go test、git diff 等）
- 严禁 Write、Edit
- 审查报告写入 services/<service-name>/_review.md
```

### Step 3: 通知用户

告知审查 Agent 已派发，完成后会输出 VERDICT。

## 审查结果处理

审查完成后：
- `VERDICT: PASS` → 告知用户可以合并
- `VERDICT: FAIL` → 告知用户有 CRITICAL 问题，询问是否需要派发修复

## 示例

```
用户: "审查 ai-model-service"
→ 全局 Claude 派发 Reviewer Agent
→ 产出 services/ai-model-service/_review.md
→ VERDICT: PASS 或 FAIL
```
