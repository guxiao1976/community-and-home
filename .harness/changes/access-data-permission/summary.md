# 数据权限核心（access-data-permission）— 全部 Wave 执行摘要

**创建时间**: 2026-08-12
**Owner**: 全局架构协调层
**状态**: ✅ 全部阶段完成（阶段① permission + 阶段② master-data + 阶段③ user + 阶段④ community-hub + 阶段⑤ web/mobile + 阶段⑥ 集成验收）

## 阶段追踪

| # | 阶段 | 状态 | 评审轮次 | 结论 | 备注 |
|---|------|:---:|:---:|------|------|
| 0 | 工具选择 | ✅ | — | OpenSpec（L 级） | |
| 1 | 需求分析 | ✅ | — | proposal.md + 5 specs | |
| 2 | 需求评审 | ✅ | /3 | 3/3 APPROVED | |
| 3 | 架构设计 | ✅ | — | design.md + tasks.md | |
| 4 | Proto 变更 | ✅ | — | commit 031f4e4 + 错误码 060007 契约同步 | Owner 执行 |
| 5 | 编码+测试 | ✅ | /3×N | permission/master-data/user/community-hub/web-mobile 全 PASS | Wave1=①②，Wave2=③④⑤ |
| 6 | 集成验证 | ✅ | — | T6.1 端到端矩阵全绿 + T6.2 收尾 | 见下方验收矩阵 |

## Wave 2（阶段③④⑤）执行摘要

| 阶段 | 服务 | 迭代 | QA | Review | 结论 |
|------|------|:---:|------|:---:|------|
| ③ | user-service (T3.1-T3.4) | 1 | 16 PASS / 71 测试(新增12) | 3/3 PASS (9 WARN) | ✅ |
| ④ | community-hub (T4.0-T4.8) | 3 | 16 PASS / 41 测试 | 3/3 PASS (10 WARN) | ✅ 修复 3 CRITICAL + 8 TDD 缺口 |
| ⑤ | web/mobile (T5.1) | 2 | 5 PASS/2 WARN / 17 测试 | 3/3 PASS (5 WARN) | ✅ |

## T6.1 端到端验收矩阵（阶段⑥）

| # | 验收项 | 结果 |
|---|--------|:---:|
| 1 | 注册用户(无小区)发布 ❌ | ✅ 99401 功能权限拒绝 |
| 2 | 加入小区(自有)→ owner+scope 出现 → 发布 ✅ | ✅ owner grant(community scope, status=0)创建 + 发布成功 |
| 3 | 未认证选举 ❌ / 认证后选举 ✅ | ✅ 能力分层:未认证 denied level-2,已认证 allowed |
| 4 | owner@A 发 B ❌ 080006 | ✅ 80006 目标小区超出数据范围 |
| 5 | 抓包改 publisher_id ❌ (JWT 生效) | ✅ 伪造 999999 落库为 JWT 身份 |
| 6 | 审核员(global) ✅ | ✅ 管理员发布成功 |
| 7 | 退出 B 后立刻发布 ❌ (缓存 DEL) | ✅ 退出后立即发布 080006 |
| 8 | 读列表按 scope 过滤;注册用户读不到内部内容 | ✅ owner 只见 A;registered_user 空列表 |
| 9 | moderation 回调放行 / 内容不存在拒绝 | ✅ 放行(status 更新) + 80001 拒绝 |

**验收发现并修复的 2 个真实缺陷**:
1. **permission 种子缺口**:owner/tenant 缺发布+读权限、contact upsert 权限缺失、选举权限未绑定 → `init_permissions.sql` 新增 4.8/4.8.1/4.8.2 段
2. **master-data scope 缓存 O(N²) 加载**:656k division OFFSET 分页需 27min → keyset 分页 + 启动预热 + 复合索引(migration 004),实测 4 秒

## 关键决策

| 日期 | 决策 | 原因 |
|------|------|------|
| 2026-08-12 | master-data T2.1 ResolveScopeAncestors 复用交接基线已有实现 | 交接文档 §2.1 已实现并提交，避免重复 |
| 2026-08-12 | permission 敏感权限(user:read/moderation)标 min_verf_level=2 | security-arch 评审 CRITICAL：未认证用户可越权读 PII |
| 2026-08-12 | AssertPublishScope 错误码 060007(独立于 060006) | 同码两义契约冲突，由 Owner 同步 proto |
| 2026-08-12 | scope 祖先缓存全量加载改 keyset 分页 + 启动预热 | 集成验收发现 656k 行 OFFSET 分页 27min → 超时 |
| 2026-08-12 | owner/tenant 补发布+读权限种子、选举权限绑定 | 集成验收发现功能权限层先于数据范围拒绝,080006 不可达 |

## QA 摘要

| 服务 | QA 轮次 | 测试数 | 覆盖率 | 结论 |
|------|:---:|:---:|:---:|------|
| permission-service | 3 | ~68 函数 | model 66% / rpc logic 61% | ✅ PASS |
| master-data-service | 2+ | 38 函数 | 覆盖达标(多数>90%) | ✅ PASS |
| user-service | 1 | 71 函数(新增12) | rpc logic/user 51.8% | ✅ PASS |
| community-hub-service | 3 | 41 函数(新增~15) | api util 86.4% / logic 40-79% | ✅ PASS |
| web/mobile | 2 | 17 测试 | 8 新增函数 7 覆盖 | ✅ PASS |

## Review 摘要

| 服务 | 审查轮次 | CRITICAL | WARNING | 结论 |
|------|:---:|:---:|:---:|------|
| permission-service | 1 | 4(assign_time/敏感权限越权/错误码/悬空引用) | 15 | ✅ 2/3 PASS + 4 CRITICAL 全修复 |
| master-data-service | 2 | 3(JWT认证/review失效/restore失效) | 若干 | ✅ 全部 CRITICAL 修复 |
| user-service | 1 | 0 | 9 | ✅ 3/3 PASS |
| community-hub-service | 3 | 3(Get-by-ID 读过滤/回环绑定×2) | 10 | ✅ 全部 CRITICAL 修复 + TDD 补齐 |
| web/mobile | 2 | 0 | 5 | ✅ 3/3 PASS |

## 例外 & 未解决问题

| 事项 | 严重度 | 处理方式 |
|------|:---:|------|
| moderation 读路径未按状态过滤 | WARN | 审核可见性门禁待后续 Wave(已在 CHANGELOG 注明) |
| 选举功能(committee:election:vote)无 API 路由 | — | 权限已绑定(600)，功能后续 Wave |
| master-data REST 无 JWT 历史遗留 | WARN | 管线已补 `rest.WithJwt`(10 路由组)；完整认证评估待后续 |
| ListPermissions/InvalidateUserCache 透传无直接单测 | WARN | QA WARNING，非阻塞，记 BACKLOG |
| community-hub scope 包 helper 无直接单测(行为级覆盖) | WARN | 仅经 RPC 逻辑测试覆盖，建议后续补隔离单测 |

## 产物索引

| 类型 | 路径 |
|------|------|
| request | `.harness/changes/access-data-permission/request.md` |
| proposal | `.harness/changes/access-data-permission/proposal.md` |
| specs | `.harness/changes/access-data-permission/specs/` |
| design | `.harness/changes/access-data-permission/design.md` |
| tasks | `.harness/changes/access-data-permission/tasks.md` |
| 需求评审 | `.harness/changes/access-data-permission/review/` |
| QA | `impl/{permission-service,master-data-service,user-service,community-hub-service,web-mobile}/_qa.md` |
| Review | `impl/{permission-service,user-service}/_review_*.md` |
| TDD 证据 | `impl/{master-data-service,web-mobile}/_tdd_evidence.md` |
| CHANGELOG | 各服务 `services/*/CHANGELOG.md` + `api-proto/CHANGELOG.md` |
| 交接 | `.harness/changes/access-data-permission/HANDOFF-NEXT-WAVE.md` |
| 新会话安排 | `.harness/changes/access-data-permission/SESSION-START-NEXT-WAVE.md` |
