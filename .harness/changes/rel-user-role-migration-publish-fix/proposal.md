# Proposal: rel_user_role 生命周期列迁移 + 移动端寻失列表路径/静默错误修复

## 为什么做

**P0（数据权限/发布链路的建库级缺口）**：permission-service 的 `rel_user_role` 表在 Go 模型（`services/permission-service/model/rel.go`）中依赖 `status`/`verified_at`/`expires_at` 三列，且 `Insert`/`BatchInsertUserRoles` 显式写入这三列；但仓库内**没有任何迁移创建它们**——`migration/` 只有 001（唯一索引 uk_user_role_scope）和 002（sys_role.platforms），全仓唯一建表脚本 `docs/specs/migration.sql`（§3.4 rel_user_role 段）缺这三列且为 AUTO_INCREMENT/created_time 旧结构。结果：**从零建库时 `rel_user_role` 缺列，任何授权/查询写 `status` 即报 `MySQL 1054 Unknown column 'status'`**，用户/权限链路在干净环境根本无法初始化。同时 `init_permissions.sql:238` 的 `INSERT ... (user_id, role_id, scope_type, scope_id, status)` 也引用 `status` 列，依赖此修复才能跑通从零建库流程。这直接命中 [[migration-must-execute]] 的经验：迁移文件必须提交后真正执行。

**P1（移动端发布/浏览链路的体验缺口）**：后端 community-hub-service 注册的寻失路由是 `GET /api/community/lostfound`（无连字符，注册于 `services/community-hub-service/api/internal/handler/routes.go` 的 lostfound 段，Create/List/Get/Resolve 四路由约 54-73 行，List 为 GET），但移动端 `web/mobile/src/api/community.ts:156` 调用的是 `GET /api/community/lost-found`（带连字符）。请求 404 且被 `notice.vue` 的 `catch { /* silent */ }` 静默吞掉 → **寻失列表页永远空白且无任何报错**。这直接命中 [[verify-api-before-calling]] 的经验：调用 API 前必须验证路由存在、禁止静默吞错。`fetchNotices`/`fetchContacts` 同样存在静默 catch，属于同类隐患。

用户价值：干净环境能正常建库初始化权限系统（P0）；移动端寻失列表能真正加载出数据、失败时有可见提示而非「空白无报错」（P1）。

## 做什么

一次门禁周期内交付（范围已由用户拍板，见「已确认的设计决策」）：

### P0 — rel_user_role 生命周期三列补齐

- **003 迁移**（D1）：新增 `services/permission-service/migration/003_add_role_lifecycle.sql`，沿用 001/002 的 information_schema 幂等 guard 写法（逐列探测列存在后 `IF` 执行），为 `rel_user_role` 添加 `status`/`verified_at`/`expires_at` 三列。live 库与从零库都能安全跑（重复执行不报错）。**「旧结构 live 库」定义**：仅缺 `status`/`verified_at`/`expires_at` 三列、`created_at` 已存在（对齐 rel.go L248 注释所述生产库实际结构）；003 **不在任何既有库上做 `created_time` → `created_at` 重命名**——该差异被接受（`RelUserRole.CreatedTime` 无任何代码消费，helpers.go 仅读 SysRole/Permission 的 CreatedTime，零值无害；从零建库的 `created_at` 由修正后 migration.sql 保证），REVISION #3 已用边界 Scenario 固化。
- **存量回填**（D2）：回填 `status=2`（已认证）、`verified_at=NULL`、`expires_at=NULL`——保留「有 grant 即活跃」的旧语义，避免存量授权在 `status=2` 严格判定下静默失效；**不回填 status=0**。回填唯一机制（REVISION #1/#6/#16）：**由 ADD COLUMN 的列定义在 guard 分支内自动完成**（`status INT NOT NULL DEFAULT 2` 使存量行 ALTER 时即置 2；`verified_at`/`expires_at DATETIME NULL` 使存量行置 NULL），003 **不含 guard 外运行的 UPDATE**；幂等重跑命中 guard=列已存在时同时跳过补列与回填，迁移后 Insert 写入的新行（含 status=0/4 显式值）不被触碰。
- **权威建表脚本修正**（D3）：同步修正 `docs/specs/migration.sql` §3.4 的 `rel_user_role` 段——补齐三列（`status INT NOT NULL DEFAULT 2` / `verified_at DATETIME NULL` / `expires_at DATETIME NULL`）；**id 保留 AUTO_INCREMENT**（REVISION #15 方案 a：映射表 id 全仓代码从不消费，且 rel.go 与 init_permissions.sql 种子均省略 id，去自增会破坏运行时代理写入与种子写入）；`created_time` → `created_at`；索引名 `idx_user_role_scope` → `uk_user_role_scope`（与 001 对齐）。
- **从零建库验证**（D3/D4）：验证 `init_permissions.sql:238` 在修复后的从零建库流程（`docs/specs/migration.sql` → `init_permissions.sql`）可执行，不再报 1054。

### P1 — 移动端寻失列表路径 + 静默错误处理

- **路径对齐**（D6）：`web/mobile/src/api/community.ts:156` 的 `GET /api/community/lost-found` → `GET /api/community/lostfound`，对齐后端注册路由；**后端 routes.go 不动**。
- **静默 catch 消除**（D9）：`web/mobile/src/pages/notice/notice.vue` 的 `fetchLostFound`/`fetchNotices`/`fetchContacts` 三处 `catch { /* silent */ }` 统一改为「toast 提示 + `console.error` 日志」，且**不 rethrow**（保持 `Promise.all` 并发加载不互相阻断）。

## 影响范围

| 服务 | 变更类型 | 说明 |
|------|:---:|------|
| permission-service | 迁移新增 | 新增 `migration/003_add_role_lifecycle.sql`（幂等补三列，回填由 ALTER DEFAULT 在 guard 内完成） |
| docs/specs | 文档修正 | `migration.sql` §3.4 rel_user_role 段：补三列 / id 保留自增 / created_at / 索引名对齐 uk_user_role_scope |
| permission-service | 验证（不改代码） | `init_permissions.sql` 从零建库可执行性验证（**不改 SQL**，仅验证；失败边界见 REQ-P0-6） |
| web/mobile | 前端修复 | `src/api/community.ts` 路径对齐；`src/pages/notice/notice.vue` 三处静默 catch → toast + console.error |
| community-hub-service | 无变更 | 后端路由 `/api/community/lostfound` 已存在，本次不动（P1 修复方向为「改前端对齐后端」） |

## 已确认的设计决策

用户已拍板（本 proposal 与 specs 以此为设计权威基线；原始需求对照见 `request.md`）：

| # | 决策 | 方案 |
|---|------|------|
| D1 | p0-migration-idempotency | 幂等 guard（information_schema 探测列存在后 IF 执行） |
| D2 | p0-backfill-status | 回填 status=2（已认证），verified_at/expires_at 置 NULL |
| D3 | p0-doc-fix-scope | 一并处理：索引名对齐 uk_user_role_scope + 验证 init_permissions.sql 从零建库可执行 |
| D4 | p0-sibling-table | 只修 rel_user_role，兄弟表不一致记入待办不动 |
| D5 | p0-runner-verify | 只加 003 文件，不补执行器 |
| D6 | p1-path-direction | 改前端 community.ts: `/api/community/lost-found` → `/api/community/lostfound`，后端不动 |
| D7 | p1-publish-form | 先不做（本次不开发发布寻失入口，另开后续任务） |
| D8 | p1-publish-extra-pages | 本次不开发（发布失物只是举例，浏览/详情页也不做） |
| D9 | p1-error-handling | 三处统一 toast 提示 + 控制台日志 |
| D10 | p1-notice-publish | 不在本次范围，另开后续任务 |
| D11 | p1-publish-ux-gate | 有当前小区即显示发布入口，越权由后端 AssertPublishScope 兜底（本次无发布入口，随发布功能延后） |

**上轮计划评审 REVISION 驱动调整（本版本已吸收）**：

**早轮已吸收**：
1. **回填机制唯一化**（REVISION #1/#6/#16）：存量回填**由 ADD COLUMN 的 ALTER 默认值在 guard 分支内自动完成**，003 不含 guard 外运行的 UPDATE——幂等重跑与迁移后新行天然不被触碰。原草稿「显式 UPDATE 回填」的两种实现歧义被消除。
2. **rel_user_role.id 保留 AUTO_INCREMENT**（REVISION #15 方案 a）：取消原草稿「去 AUTO_INCREMENT 改雪花 ID」。理由：该映射表 id 全仓代码从不消费，无雪花语义；rel.go Insert/BatchInsertUserRoles 与 init_permissions.sql:238 种子 INSERT 均省略 id、依赖自增，去自增会在严格模式报 MySQL 1364、非严格模式 id=0 触发主键冲突致 INSERT IGNORE 静默丢弃种子行，直接违背「不改 init_permissions.sql」约束。

**本轮（REVISION #1/#2/#3）已解决**：
3. **验收标准（c）补执行载体**（REVISION #1，已解决）：验收标准「003 在（c）从修复后 migration.sql 新建的库上执行一次不报错」原无对应 task——Task 0.2 只覆盖旧结构 live 库/幂等重跑/部分列缺失，Task 1.2 从零建库后仅跑 init_permissions.sql 与省略 id 写入，未显式执行 003。现 Task 1.2 追加一步：从零建库 + init_permissions.sql 后**再执行一次 003** → guard 探测三列已存在、跳过 ADD 与回填、无报错、表结构与存量种子数据不变。验收标准（c）明确映射到该 task。
4. **REQ-P1-ERR-1 唯一解释 + 并发全失败场景**（REVISION #2，已解决）：原「每个失败 SHALL 呈现用户可见 toast」与「showToast 单实例替换不堆叠」在并发全失败场景无单一解释，且 spec 无对应场景。现 REQ-P1-ERR-1 收敛为唯一解释——「失败时刻各触发区块 toast + console.error；并发失败时可见 toast 收敛为最后一次调用（单实例替换，至少一个可见 toast），每失败各一次 console.error，被覆盖不算违反」；REQ-P1-ERR-1 新增「三请求并发全部失败（toast 收敛）」场景、REQ-P1-ERR-2 新增「三请求并发全失败（核心触发场景）」场景。
5. **「旧结构」定义唯一化 + created_time 不重命名**（REVISION #3，已解决）：原 REQ-P0-2 Scenario 1 的「旧结构」是否含 created_time 无唯一解释。现明确「旧结构 live 库」= 仅缺三列、created_at 已存在（对齐 rel.go L248 注释）；003 不在 live 库做 created_time→created_at 重命名（RelUserRole.CreatedTime 零值无害），REQ-P0-2 新增边界 Scenario 固化该结论。

## 风险评估

- **存量授权静默失效（高，P0 决策已化解）**：若回填 status=0，则存量授权在 `FindActiveByUserId`（status=2 严格判定）下全部失效，用户数据范围/能力被静默剥夺。缓解：D2 明确回填 status=2 + verified_at/expires_at=NULL，保留「有 grant 即活跃」语义（与 `init_permissions.sql` 中 `sys_admin(0,8,'global',0,2)` 的 verified_at=NULL 先例一致）；该语义同时被 [[auto-grant-unverified-grant-confers-scope-level0]] 的 status∈{0,1,2} 活跃判定支持。
- **回填覆盖迁移后新行 / 显式 status 存量行（高，REVISION 驱动化解）**：无条件 UPDATE 回填会把 Insert 显式写入的 status=0（未认证自动授权）/status=4（已过期）行静默改写为 2 → level-0 → level-2 能力越权。缓解：003 不含 guard 外 UPDATE，回填仅由 ALTER DEFAULT 在补列当次自动完成；重跑与「列已存在 + 显式 status」的库均跳过，显式值不被改写（REQ-P0-4 边界场景固化）。
- **幂等 guard 的并发竞态（低）**：information_schema guard 对「严格并发双跑」非原子（两会话可能同时读到 guard=0 后都执行 ADD COLUMN 而报 Duplicate column）。缓解：迁移按 001→002→003 串行执行（人工/CI 有序），guard 保证「顺序重跑」安全；并发双跑不在本变更保障范围（001/002 同款写法）。
- **去自增引发的写入断裂（高，已化解）**：原草稿「id 去 AUTO_INCREMENT」会让省略 id 的运行时代理写入与种子写入断裂。缓解：REVISION #15 方案 a——rel_user_role.id 保留 AUTO_INCREMENT，仅对齐列/索引/时间列命名；新增 REQ-P0-6 边界场景验证省略 id 的 Insert 在修复后从零库可写。
- **唯一索引名对齐的影响（低）**：migration.sql 由 `idx_user_role_scope` 改 `uk_user_role_scope` 仅影响**从零建库**的 DDL 文本，与 001 已建的索引一致；live 库已有 001 的 uk 索引，不重复执行 003 的索引逻辑（003 只加列）。无破坏。
- **init_permissions.sql 依赖建表顺序（低）**：`init_permissions.sql:238` 的 INSERT 依赖 `rel_user_role` 先建表，从零流程必须按 migration.sql → init_permissions.sql 顺序执行；顺序颠倒按 [[cross-service-seed-deployment-order]] 记录为部署编排约束。
- **init_permissions.sql 可能暴露超出 rel_user_role.status 的失败（低）**：验证若发现与本变更 DDL 修正无关的其他建库失败，按 REQ-P0-6 边界登记待办，不擅自扩大迁移范围（见 REQ-P0-6 Scenario 3）。
- **前端路径修复后仍有业务错误（低）**：路径对齐只消 404，业务/网络错误仍可能发生——由 D9 的 toast + console 兜底提示，不再静默。
- **后端返回数据结构与前端类型不一致（低）**：本次仅修路径与错误处理，若后端 ListLostFound 响应字段与前端 `LostFoundItem` 不匹配属另一缺陷，不在本变更范围（spec 验收聚焦「请求到达正确路由」）。

## 不做清单（Won't have — 本轮明确不实现）

- **不改兄弟表**（D4）：`rel_role_permission` 等其他表的 AUTO_INCREMENT/created_time 结构不一致，仅记入待办，本次不动。
- **不补迁移执行器**（D5）：沿用 001/002 的手动执行 + harness-checks 验证模式，不新增自动执行器。
- **不改 rel_user_role.id 为雪花 ID**（REVISION #15 方案 a）：保留 AUTO_INCREMENT，仅对齐列/索引/时间列命名。
- **不做移动端发布寻失入口**（D7）：发布表单、发布交互不在本次范围，另开后续任务。
- **不做寻失浏览『全部→』页**（D8）：`notice.vue` 的「全部 →」入口不做跳转目标页。
- **不做寻失详情页**（D8）：`onLostFoundClick` 的「寻失详情开发中...」占位保留，不实现详情页。
- **不做移动端发布通知入口**（D10）：通知发布另开后续任务。
- **不做发布入口 UX 门禁**（D11）：「有当前小区即显示发布入口」随发布功能延后；越权由后端 `AssertPublishScope` 兜底（本次无发布入口故无需实现）。
- **不改后端**：community-hub-service `routes.go` 的 `/api/community/lostfound` 路由不动；不改 `init_permissions.sql` 本体（只验证其可执行性）。
- **不做 api-proto / common 变更**：本变更不涉及 proto，不改 `common/`。

## 验收标准

- [ ] 新增 `services/permission-service/migration/003_add_role_lifecycle.sql`，沿用 001/002 的 information_schema guard 写法（逐列探测）。
- [ ] 003 在（a）旧结构 live 库、（b）已执行过 003 的库、（c）从修复后 migration.sql 新建的库上各执行一次均不报错，且 rel_user_role 三列存在。（a/b 由 Task 0.2 执行；**c 由 Task 1.2 覆盖**：从零建库 + init_permissions.sql 后显式重跑 003 → guard 探测三列已存在、跳过 ADD 与回填、无报错、表结构与存量种子数据不变。）
- [ ] 003 首跑于旧结构库后，存量 rel_user_role 行全部 `status=2, verified_at=NULL, expires_at=NULL`；003 **无 guard 外 UPDATE**，幂等重跑不触碰迁移后 Insert 显式写入的 status 值（含 status=0/4）。
- [ ] `docs/specs/migration.sql` §3.4 rel_user_role 段含三列（`status INT NOT NULL DEFAULT 2` / `verified_at DATETIME NULL` / `expires_at DATETIME NULL`）、id 保留 AUTO_INCREMENT、`created_at` 时间列、唯一索引名 `uk_user_role_scope`，文档无 `idx_user_role_scope` 残留。
- [ ] 从修复后 migration.sql 建库 → 执行 init_permissions.sql 成功（含 :238 的 rel_user_role INSERT），无 MySQL 1054 Unknown column 'status'；省略 id 的运行时代理 Insert（仿 rel.go）可写入（无 1364/主键冲突）。
- [ ] `web/mobile/src/api/community.ts` 中寻失列表请求路径为 `/api/community/lostfound`，全仓无 `/api/community/lost-found` 调用残留。
- [ ] `notice.vue` 三处 fetch（notices/contacts/lostfound）失败时刻各触发区块 toast + `console.error`，不再静默；成功时不弹错误 toast；失败不阻断其他请求渲染；**并发全部失败时可见 toast 收敛为最后一次调用（至少一个可见 toast）+ 每失败一次 `console.error` + 页面不崩**（REQ-P1-ERR-1 唯一解释）。
- [ ] `harness-checks.sh --service permission-service` 通过（003 迁移 + 文档修正不引入 Go 代码回归）。
- [ ] **P1 前端门禁**：`cd web/mobile && npm run test:unit` 通过（notice.spec.ts 覆盖失败 toast + `console.error` + 不阻断并发），`npm run type-check` 通过（或等价前端门禁）。
- [ ] 兄弟表不一致项已登记待办（`.harness/tasks/BACKLOG.md` 或既有任务系统）。
