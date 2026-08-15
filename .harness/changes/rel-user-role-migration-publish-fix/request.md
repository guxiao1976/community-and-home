# Request — rel_user_role 生命周期列迁移 + 移动端寻失列表路径/静默错误修复

## 用户原始需求（任务转述，P0+P1 缺口）

**P0 — rel_user_role 生命周期列缺迁移（permission-service + docs/specs/migration.sql）**

- permission-service 的 `rel_user_role` 表在 Go 模型（`services/permission-service/model/rel.go`）中依赖 `status`/`verified_at`/`expires_at` 三列，且 `Insert`/`BatchInsertUserRoles`（rel.go L149/160/231）显式写入；但仓库内没有任何迁移创建它们——`migration/` 只有 001（uk_user_role_scope）与 002（sys_role platforms）；全仓唯一建表是 `docs/specs/migration.sql`（rel_user_role 段），缺三列且为 AUTO_INCREMENT/created_time 旧结构。
- 结果：从零建库报 `MySQL 1054 Unknown column 'status'`，权限/发布链路在干净环境无法初始化；`init_permissions.sql:238` 的 rel_user_role 种子 INSERT（引用 status 列）同样依赖此修复。

**P1 — 移动端寻失列表路径不匹配 + 静默错误（web/mobile）**

- 后端 community-hub-service 注册寻失路由为 `GET /api/community/lostfound`（无连字符），但移动端 `web/mobile/src/api/community.ts:156` 调用 `GET /api/community/lost-found`（带连字符）→ 列表请求 404 且被 `notice.vue` 的 `catch { /* silent */ }` 静默吞掉 → 列表页永远空白无报错。
- `fetchLostFound`/`fetchNotices`/`fetchContacts` 三处均为静默 catch，同类隐患。

## 分级信息

- 分级: **M**
- 路由: Pipeline（需求评审 → architect-design → Generator → QA → Reviewer）
- 涉及: permission-service（P0 迁移）/ docs/specs（P0 权威 DDL）/ web/mobile（P1 前端）
- 后端 community-hub-service：**零变更**（P1 修复方向为「改前端对齐后端」）

## 需求要点（用户已拍板，D1-D11 见 proposal.md §已确认的设计决策）

| # | 要点 | 用户拍板结论 |
|---|------|------|
| D1 | 003 迁移幂等 | information_schema 探测列存在后 IF 执行（沿用 001/002 写法） |
| D2 | 存量回填 | 回填 status=2（已认证），verified_at/expires_at 置 NULL；不回填 0 |
| D3 | 文档修正范围 | 索引名对齐 uk_user_role_scope + 验证 init_permissions.sql 从零建库可执行 |
| D4 | 兄弟表 | 只修 rel_user_role，兄弟表不一致记入待办不动 |
| D5 | 迁移执行器 | 只加 003 文件，不补执行器（沿用 001/002 手动执行 + harness-checks 验证） |
| D6 | P1 路径方向 | 改前端 community.ts: lost-found → lostfound，后端不动 |
| D7 | 发布寻失入口 | 先不做（另开后续任务） |
| D8 | 浏览『全部→』页 / 详情页 | 本次不做 |
| D9 | 静默错误处理 | 三处统一 toast 提示 + 控制台日志，不 rethrow |
| D10 | 发布通知入口 | 不在本次范围，另开后续任务 |
| D11 | 发布入口 UX 门禁 | 本次无发布入口，随发布功能延后；越权由后端 AssertPublishScope 兜底 |

## 上轮计划评审 REVISION 驱动调整（本版本已吸收，详见 traceability）

1. 回填机制唯一化：**回填由 ADD COLUMN 的 ALTER 默认值在 guard 分支内自动完成，003 不含 guard 外 UPDATE**（REVISION #1/#6/#16，方案 b）。
2. **rel_user_role.id 保留 AUTO_INCREMENT**，取消「去自增改雪花 ID」——映射表 id 全仓代码从不消费，且 rel.go Insert/BatchInsert 与 init_permissions.sql:238 种子 INSERT 均省略 id，去自增会导致严格模式 1364 / INSERT IGNORE 静默丢种子行（REVISION #15，方案 a）。
3. 其余 REVISION 修订：migration.sql status 列 `NOT NULL DEFAULT 2` 契约入 spec、init_permissions.sql 验证失败边界界定、tasks.md 补齐 11 个 Requirement→task 映射、前端门禁入验收标准、路由引用统一为全路径、.change.yaml 去 community-hub-service 零变更服务。
