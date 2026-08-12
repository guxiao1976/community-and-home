# 数据权限核心（access-data-permission）— Wave 1 执行摘要

**创建时间**: 2026-08-12
**Owner**: 全局架构协调层
**状态**: 🟡 Wave 1 完成，阶段 3/4(user/community-hub)待后续 Wave

## 阶段追踪

| # | 阶段 | 状态 | 评审轮次 | 结论 | 备注 |
|---|------|:---:|:---:|------|------|
| 0 | 工具选择 | ✅ | — | OpenSpec（L 级） | |
| 1 | 需求分析 | ✅ | — | proposal.md + 5 specs | |
| 2 | 需求评审 | ✅ | /3 | 3/3 APPROVED | |
| 3 | 架构设计 | ✅ | — | design.md + tasks.md | |
| 4 | Proto 变更 | ✅ | — | commit 031f4e4 + 错误码 060007 契约同步 | Owner 执行 |
| 5 | 编码+测试 | ✅ | /3 | permission PASS / master-data 修复后 PASS | Wave 1 = 阶段① |
| 6 | 集成验证 | 🟡 | — | 门禁全绿，归档完成 | 跨服务集成(阶段6.1)待后续 |

## 关键决策

| 日期 | 决策 | 原因 |
|------|------|------|
| 2026-08-12 | master-data T2.1 ResolveScopeAncestors 复用交接基线已有实现 | 交接文档 §2.1 已实现并提交，避免重复 |
| 2026-08-12 | permission 敏感权限(user:read/moderation)标 min_verf_level=2 | security-arch 评审 CRITICAL：未认证用户可越权读 PII |
| 2026-08-12 | AssertPublishScope 错误码 060007(独立于 060006 角色编码已存在) | 同码两义契约冲突，由 Owner 同步 proto |
| 2026-08-12 | FindByRoleId 改用 select * 去除不存在 assign_time 列 | need_human 安全修复，缓存失效静默失败 |

## QA 摘要

| 服务 | QA 轮次 | 测试数 | 覆盖率 | 结论 |
|------|:---:|:---:|:---:|------|
| permission-service | 3 | ~68 函数 | model 66% / rpc logic 61% | ✅ PASS |
| master-data-service | 2+ | 38 函数 | 覆盖达标(多数>90%) | ✅ PASS |

## Review 摘要

| 服务 | 审查轮次 | CRITICAL | WARNING | 结论 |
|------|:---:|:---:|:---:|------|
| permission-service | 1 | 4(assign_time/敏感权限越权/错误码/悬空引用) | 15 | ✅ 2/3 PASS + 4 CRITICAL 全修复 |
| master-data-service | 2 | 3(JWT认证/review失效/restore失效) | 若干 | ✅ 全部 CRITICAL 修复 |

## 例外 & 未解决问题

| 事项 | 严重度 | 处理方式 |
|------|:---:|------|
| master-data REST 无 JWT 历史遗留 | WARN | 管线已补 `rest.WithJwt`(10 路由组)；完整认证评估待后续 |
| ListPermissions/InvalidateUserCache 透传无直接单测 | WARN | QA WARNING，非阻塞，记 BACKLOG |
| master-data 管线超轮次(TDD 证据占位符) | WARN | 手工补 RED 摘录(CHANGELOG + _tdd_evidence.md)后验收 |
| 阶段 3/4(user/community-hub)+ 阶段 5 前端 + 阶段 6 集成 | — | 后续 Wave 执行（见 HANDOFF-NEXT-WAVE.md + BACKLOG task-2026-08-12-nextwave） |

## 产物索引

| 类型 | 路径 |
|------|------|
| request | `.harness/changes/access-data-permission/request.md` |
| proposal | `.harness/changes/access-data-permission/proposal.md` |
| specs | `.harness/changes/access-data-permission/specs/` |
| design | `.harness/changes/access-data-permission/design.md` |
| tasks | `.harness/changes/access-data-permission/tasks.md` |
| 需求评审 | `.harness/changes/access-data-permission/review/` |
| QA | `.harness/changes/access-data-permission/impl/permission-service/_qa.md`、`impl/master-data-service/_qa.md` |
| Review | `.harness/changes/access-data-permission/impl/permission-service/_review_{design-biz,security-arch,standards-eng}.md` |
| TDD 证据 | `.harness/changes/access-data-permission/impl/master-data-service/_tdd_evidence.md` |
| CHANGELOG | `services/permission-service/CHANGELOG.md`、`services/master-data-service/CHANGELOG.md`、`api-proto/CHANGELOG.md` |
| 交接 | `.harness/changes/access-data-permission/HANDOFF-NEXT-WAVE.md` |
| 新会话安排 | `.harness/changes/access-data-permission/SESSION-START-NEXT-WAVE.md` |
