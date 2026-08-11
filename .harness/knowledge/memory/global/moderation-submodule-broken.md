---
triggers: ["submodule", "worktree", ".gitmodules", "moderation-service", "master-data-service"]
type: pitfall
severity: should-follow
service: all
status: active
created: 2026-06-15
updated: 2026-08-09
apply_count: 0
---
`services/moderation-service`、`services/master-data-service`、`common` 是 git submodule（mode 160000），`.gitmodules` 文件已补全（2026-06-15, commit ff59ef6），包含三个 submodule 的 URL：

- `common` → `git@github.com:guxiao1976/community-common.git`
- `services/moderation-service` → `git@github.com:guxiao1976/community-moderation.git`
- `services/master-data-service` → `git@github.com:guxiao1976/community-masterdata.git`

**已解决。** worktree 现已可用。新建 worktree 后运行 `git submodule update --init --recursive` 即可拉取所有子模块。
