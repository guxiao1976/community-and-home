# Git 仓库治理规范

> 主仓库 + 所有子模块的 git 命名、分支、子模块统一规范。
> 归属：全局架构协调层（CLAUDE.md 职责「全局规范定义 + 架构决策」）。
> 相关：`.harness/rules/工程结构.md`（仓库→目录映射）、`harness-checks.sh` check #17（git 卫生兜底）。

## 1. 默认分支

- 主仓库与所有子模块的默认分支统一为 `main`。
- 禁止 `main` 与 `master` 并存——两者会形成无共同祖先的孤儿分叉线，无法正常合并（2026-08 曾因此清理 4 条孤儿线）。

## 2. 分支命名

| 类型 | 命名 | 说明 |
|------|------|------|
| 功能 | `feature/<描述>` | 新功能 |
| 修复 | `fix/<描述>` | bug 修复 |
| 杂项 | `chore/<描述>` | 基建/清理 |
| 临时 | `worktree-wf_*` | harness worktree 临时分支 |

- 分支合并/关闭后即删，禁止 stale 分支堆积。
- `worktree-wf_*` 是 EnterWorktree/ExitWorktree 的临时分支，退出时必须删除（残留由 check #17 兜底）。

## 3. 仓库命名

- 仓库名：`community-<name>`（如 `community-user`、`community-ai-model`）。
- 目录名：`services/<name>-service`（如 `services/user-service`）。
- Go module：`github.com/guxiao1976/community-<name>`，与仓库名一致。

## 4. 子模块治理

- 主仓库 index 里的 gitlink（mode 160000）必须在 `.gitmodules` 登记，一一对应。
- `.gitmodules` 的 `url` 必须与子模块实际 remote 一致。
- 子模块默认分支同样为 `main`。

## 5. 已知偏差（待收敛）

| 位置 | 现状 | 应为 |
|------|------|------|
| master-data 仓库名 | `community-masterdata` | `community-master-data` |
| master-data module | `community-master-data-service` | `community-master-data` |
| moderation module | `community-moderation-service` | `community-moderation` |

> 收敛需改 go.mod import + 子模块 URL + go.work，涉及跨服务改动，单独排期，不在本规范落地时强行改。
>
> 另：`工程结构.md` 隐含所有服务都是独立仓库，但当前仅 `common` + `ai-model/master-data/moderation` 是子模块，其余 6 个服务直接跟踪在主仓库。此结构差异待确认后收敛。

## 6. 机械化兜底

- `harness-checks.sh` check #17 检测：gitlink 无 `.gitmodules` 条目、孤儿 `worktree-wf_*` 分支。
