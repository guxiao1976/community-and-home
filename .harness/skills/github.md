# GitHub Skill

## 触发条件

需要使用 GitHub（Issues、PR、代码搜索、仓库操作）时加载。触发词：`GitHub`、`PR`、`Issue`、`仓库`、`Pull Request`。

## 角色

你是 GitHub 工具的使用者。通过 MCP Server `github` 提供的工具与 GitHub 交互。

## 可用 MCP 工具

GitHub MCP Server (`@anthropic/mcp-server-github`) 提供以下工具：

| 工具 | 用途 | 何时用 |
|------|------|--------|
| `search_repositories` | 搜索仓库 | 查找项目相关仓库 |
| `get_repository` | 获取仓库详情 | 查看仓库信息 |
| `list_issues` | 列出 Issues | 扫描待处理问题 |
| `get_issue` | 获取单个 Issue | 读取 Issue 详情 |
| `create_issue` | 创建 Issue | 向仓库报告问题 |
| `update_issue` | 更新 Issue | 修改状态/标签/指派人 |
| `list_pull_requests` | 列出 PR | 扫描待审查 PR |
| `get_pull_request` | 获取 PR 详情 | 查看 PR 变更和状态 |
| `create_pull_request` | 创建 PR | 提交代码变更 |
| `merge_pull_request` | 合并 PR | 审查通过后合并 |
| `search_code` | 搜索代码 | 跨仓库查找代码 |
| `get_file_contents` | 读文件内容 | 查看仓库中的文件 |

## Harness 集成场景

### 场景 1：扫描 Issue → 写入 BACKLOG

```
1. 读取 ENABLED_LABELS（如 "bug,enhancement,debt"）
2. 对每个 label → list_issues(state: "open", labels: [label])
3. 对每个 Issue → 查 BACKLOG 是否已有对应 task（source_detail 含 issue URL）
4. 新 Issue → harness-tasks.sh create --source github --detail "issue: <url>"
5. 输出发现摘要
```

### 场景 2：Pipeline PASS → 开 PR

```
1. Generator 在 worktree 中完成修改
2. QA + Review 全部 PASS
3. 使用 create_pull_request 提交变更
4. PR 描述中附上 QA/Review 摘要
5. 更新对应 task: status → review（等待人工 merge）
```

### 场景 3：PR Review 反馈 → 修复任务

```
1. 扫描 open PR 的 review comments
2. 有 "changes requested" → 自动创建修复 task（P0, source: github）
3. 在 PR 下 comment 当前状态
```

## 约束

- Issue 创建需要在标题前加 `[AI]` 前缀，标识为 AI 自动创建
- PR 描述末尾附带标准 footer：`🤖 Generated with [Claude Code](https://claude.com/claude-code)`
- 不要 force push 到 main/master 分支
- 操作前确认仓库权限（GITHUB_TOKEN 的 scope）
