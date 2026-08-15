# Design: rel_user_role 生命周期列迁移 + 移动端寻失列表路径/静默错误修复

> 变更标识：`rel-user-role-migration-publish-fix`（分级 M，路由 Pipeline）
> 设计权威基线：proposal.md §已确认的设计决策（D1-D15）+ request.md REVISION 驱动调整段

## 需求追溯矩阵（spec → design，防需求遗漏/设计蔓延）

| 需求 ID | 需求内容摘要 | 对应设计章节 | 覆盖状态 |
|---------|-------------|-------------|:---:|
| REQ-P0-1 | 交付 003 迁移文件（information_schema guard 逐列探测） | §数据模型（003 迁移）+ §ADR | ✅ |
| REQ-P0-2 | 三列幂等添加（三态库补列 vs no-op，不 rename created_time） | §数据模型（003 迁移）+ §业务流程 | ✅ |
| REQ-P0-3 | 列定义语义与模型一致（status DEFAULT 2 / DATETIME NULL） | §数据模型（列定义契约） | ✅ |
| REQ-P0-4 | 存量行回填（唯一机制：ALTER DEFAULT，零 guard 外 UPDATE） | §数据模型（回填机制） | ✅ |
| REQ-P0-5 | 权威建表脚本修正（migration.sql §3.4） | §数据模型（migration.sql 修正） | ✅ |
| REQ-P0-6 | 从零建库流程可执行（migration.sql → init_permissions.sql 无 1054） | §业务流程（从零建库验证） | ✅ |
| REQ-P0-7 | 兄弟表保持不动（仅登记待办） | §数据模型（不做清单）+ §ADR | ✅ |
| REQ-P1-PATH-1 | 前端请求对齐后端路由 `/api/community/lostfound` | §接口设计（前端 API 对齐）+ §业务流程 | ✅ |
| REQ-P1-PATH-2 | 后端路由保持不动（只改前端） | §服务归属决策 + §接口设计 | ✅ |
| REQ-P1-ERR-1 | 三处 fetch 失败不再静默（toast + console.error） | §接口设计（前端错误处理） | ✅ |
| REQ-P1-ERR-2 | 失败不 rethrow、不阻断并发加载 | §接口设计（前端错误处理） | ✅ |
| REQ-P1-ERR-3 | 其余静默点仅登记待办不动 | §数据模型（不做清单）+ §ADR | ✅ |

> 所有 spec 正式需求点 100% 覆盖。本设计无 spec 依据之外的设计蔓延——全部设计内容均可追溯至 D1-D15 决策或 REQ 条目。

## 服务归属决策

| 功能 | 归属服务 | 理由 |
|------|---------|------|
| 003 幂等迁移（补三列 + ALTER-DEFAULT 回填） | permission-service | 拥有 `rel_user_role` 表数据与 `model/rel.go` 模型；迁移文件落在其 `migration/` 目录（D1/D2/D5） |
| 权威建表脚本修正（migration.sql §3.4） | docs/specs（全局 Claude） | 全仓唯一跨服务从零建库权威 DDL，非单服务产物；由全局协调层修改，不分发子 Claude（D3） |
| init_permissions.sql 从零建库可执行性验证 | permission-service | 只验证不改 SQL；验证其 `:238` 种子 INSERT 依赖 rel_user_role 三列（D3） |
| 移动端寻失路径对齐（community.ts） | web/mobile | 前端调用层，路径字符串修正，后端零变更（D6） |
| 移动端三处静默 catch 消除（notice.vue） | web/mobile | 前端页面错误处理，toast + console.error（D9） |
| 兄弟表不一致 / 其余静默点登记待办 | .harness/tasks/BACKLOG.md（全局） | 仅登记不落库（D4/D13） |

**归属存疑项**：无。community-hub-service 后端路由 `/api/community/lostfound` 已注册且为修复前置契约（D6 明确「后端不动」），不产生灰色地带。

## 数据模型

### 变更类型：对既有表 `rel_user_role` 补列（无新增表）

**目标表**：`rel_user_role`（用户-角色关联表，含数据范围三态）。Go 模型 `model/rel.go` 的 `RelUserRole` 已依赖 `status`/`verified_at`/`expires_at` 三列（`db:"status"` / `db:"verified_at"` / `db:"expires_at"`），但仓库内无迁移创建它们——本变更为「对齐模型与真实表结构」，非新增字段语义。

### 新增列（003 迁移 + migration.sql 修正，两处定义严格一致）

```sql
ALTER TABLE rel_user_role ADD COLUMN status INT NOT NULL DEFAULT 2
  COMMENT '个体角色生命周期: 0=未认证 1=待审 2=已认证 3=已驳回 4=已过期';
ALTER TABLE rel_user_role ADD COLUMN verified_at DATETIME NULL COMMENT '个体认证通过时间';
ALTER TABLE rel_user_role ADD COLUMN expires_at DATETIME NULL COMMENT '个体角色到期时间, NULL=永久';
```

**字段约束补充**：

| 字段 | 类型/精度 | 非空/唯一 | 语义契约 | 与 Go 模型对齐 |
|------|-----------|-----------|----------|----------------|
| `status` | INT | NOT NULL DEFAULT 2 | 0=未认证 1=待审 2=已认证 3=已驳回 4=已过期；DEFAULT 2 使「不带 status 的外部 INSERT」默认已认证，不静默失效 grant | `Status int64`（显式写入，DEFAULT 仅兜底） |
| `verified_at` | DATETIME | NULL | NULL = 未认证（未通过）；认证通过时间 | `VerifiedAt sql.NullTime` |
| `expires_at` | DATETIME | NULL | NULL = 永久（永不过期），与 SQL 谓词 `expires_at IS NULL OR expires_at > NOW()` 一致 | `ExpiresAt sql.NullTime` |

**表引擎/字符集**：沿用 `rel_user_role` 既有 InnoDB/utf8mb4（003 仅 ADD COLUMN，不重建表）。

**软删除**：`rel_user_role` 无 `deleted_at`（硬删除映射表），003 不引入；遵循编码规范 §5.1 的三列 `created_at`/`updated_at`/`deleted_at` 仅适用于新表，此处为补列不扩改时间列结构。

### 回填机制（唯一机制，REQ-P0-4 / D2 / REVISION #1/#6/#16）

- **回填值**：`status=2`（已认证）、`verified_at=NULL`、`expires_at=NULL`，保留「有 grant 即活跃」旧语义；不回填 status=0。
- **回填方式**：**由 ADD COLUMN 的列定义在 guard 分支内自动完成**——`ADD COLUMN status INT NOT NULL DEFAULT 2` 使存量行 ALTER 时即置 2；`ADD COLUMN verified_at/expires_at DATETIME NULL` 使存量行置 NULL。003 **不含 guard 外运行的 UPDATE**。
- **幂等保证**：重跑命中 guard（列已存在）→ 同时跳过 ADD 与回填；迁移后 Insert 写入的新行（含 status=0/4 显式值）天然不被触碰。
- **适用环境唯一化（D12）**：回填仅发生在「缺三列且有存量行」的旧从零库；生产 live 库已含三列 → 003 no-op 不回填。

### migration.sql §3.4 修正（REQ-P0-5 / D3）

```sql
CREATE TABLE IF NOT EXISTS `rel_user_role` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,             -- 保留自增（REVISION #15 方案 a）
    `user_id` BIGINT NOT NULL COMMENT '用户ID',
    `role_id` BIGINT NOT NULL COMMENT '角色ID',
    `scope_type` VARCHAR(50) NOT NULL DEFAULT 'community' COMMENT '...',
    `scope_id` BIGINT NOT NULL COMMENT '...',
    `status` INT NOT NULL DEFAULT 2 COMMENT '个体角色生命周期...',
    `verified_at` DATETIME NULL COMMENT '个体认证通过时间',
    `expires_at` DATETIME NULL COMMENT '个体角色到期时间, NULL=永久',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- created_time → created_at
    PRIMARY KEY (`id`),
    UNIQUE INDEX `uk_user_role_scope` (`user_id`, `role_id`, `scope_type`, `scope_id`),  -- idx → uk（对齐 001）
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_role_id` (`role_id`),
    INDEX `idx_scope` (`scope_type`, `scope_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户-角色关联表（含数据范围）';
```

- `id` **保留 AUTO_INCREMENT**：映射表 id 全仓代码从不消费（无雪花语义），且 rel.go `Insert`/`BatchInsertUserRoles` 与 `init_permissions.sql:238` 种子 INSERT 均省略 id、依赖自增；去自增会导致严格模式 MySQL 1364 / INSERT IGNORE 静默丢种子行。
- `created_time` → `created_at`：`RelUserRole.CreatedTime`（`db:"created_at"`）无任何代码消费，零值无害；旧库的 `created_time` 不 rename（003 只补三列，不动时间列）。

### 索引设计

无新增索引。唯一索引名 `idx_user_role_scope` → `uk_user_role_scope` 仅影响从零建库 DDL 文本，与 001 已建的 uk 索引一致（live 库已有 uk，003 不重复执行索引逻辑）。触发记忆：[[unique-index-migration-dup-precheck]]（本次仅改名对齐既有 uk，无新增唯一索引，无重复数据风险）。

## 接口设计

本变更**无新接口、无 gRPC/Proto 变更**。仅涉及既有前端 API 调用的路径修正与错误处理语义。

### web/mobile `getLostFoundList`（路径对齐，REQ-P1-PATH-1）

- **变更**：`web/mobile/src/api/community.ts:156` 请求路径 `GET /api/community/lost-found` → `GET /api/community/lostfound`（对齐后端 `community-hub-service` 已注册路由，routes.go List 注册于 60-62 行）。
- **后端契约**：`GET /api/community/lostfound`（无连字符）已存在，本次零改动（REQ-P1-PATH-2）。
- **错误码**：无新增错误码；失败反馈由前端 toast + console 兜底。

### web/mobile `notice.vue` 三处 fetch 错误处理（REQ-P1-ERR-1/2）

- **变更**：`fetchNotices`/`fetchContacts`/`fetchLostFound` 三处 `catch { /* silent */ }` → catch 内 `console.error`（错误对象 + 区块标识）+ `uni.showToast`（区块相关失败提示，`icon: 'none'`）；**不 rethrow**（保持 `Promise.all` 并发加载互不阻断）。
- **并发收敛唯一解释**：uni-app `showToast` 单实例替换语义使并发全失败时可见 toast 收敛为最后一次调用（至少一个可见），每次失败各一次 `console.error`，被覆盖不算违反。
- **成功路径**：不弹错误 toast、无 `console.error`。

## 业务流程

### P0 迁移执行流程（正常路径）

1. 部署/初始化：按 `docs/specs/migration.sql` → `init_permissions.sql` 顺序建库（顺序固定，见 [[cross-service-seed-deployment-order]]）。
2. 执行 003：information_schema 逐列探测 → 缺列则 ADD COLUMN（含 ALTER-DEFAULT 回填），已存在则跳过；末尾 SELECT 验证三列存在。
3. 验证：`DESCRIBE rel_user_role` 确认三列存在，存量行 `status=2/NULL/NULL`。

### P0 三态库行为（REQ-P0-6 三态验证）

| 库状态 | 列状态 | 003 行为 |
|--------|--------|----------|
| (a) 旧 migration.sql 从零库 | 缺三列 + created_time | 补三列 + 回填 status=2，不 rename created_time |
| (b) 生产 live 库 | 已含三列 + created_at（rel.go L248） | guard 跳过 ADD 与回填，no-op |
| (c) 修正后从零库 | 已含三列 | guard 跳过，no-op（从零建库后重跑 003 亦 no-op） |

### 异常/失败路径

- **幂等重跑**：003 再次执行 → guard 探测列已存在 → 跳过 ADD 与回填，无报错，表结构与存量数据不变。
- **部分列缺失**：逐列 guard，只补缺失列，已有列不重复添加。
- **并发双跑竞态（低）**：information_schema guard 非原子；迁移按 001→002→003 串行执行，guard 保证「顺序重跑」安全（001/002 同款写法，不在本变更保障范围）。
- **init_permissions.sql 暴露无关失败**：与本变更 rel_user_role DDL 修正无因果 → 由全局登记待办，不扩大迁移范围；若失败源于本变更 DDL 修正的副作用 → 回路由全局 Claude 在 Task 1.1 修 `docs/specs/migration.sql`，不分发子 Claude（REQ-P0-6 Scenario 4）。

### P1 前端加载流程

- **正常**：`loadAll` 并发 `Promise.all([fetchNotices, fetchContacts, fetchLostFound])`，路径对齐后 `fetchLostFound` 命中正确路由，数据渲染。
- **失败**：单个 fetch 失败 → toast + console.error，不 rethrow，其余区块正常渲染；并发全失败 → 至少一个可见 toast + 三次 console.error，页面不崩、loading 复位。

### 跨服务一致性

本变更**无跨服务数据写入**（不新增/修改业务数据流转），无需 Saga/事务消息/对账。涉及的唯一跨组件顺序约束是「migration.sql → init_permissions.sql」部署编排顺序，记入 [[cross-service-seed-deployment-order]]。

## Proto 变更

| 文件 | 变更类型 | 破坏性(是/否) | 说明 |
|------|:---:|:---:|------|
| （无） | — | 否 | `proto_change_required=false`，本变更不涉及 api-proto |

> 无破坏性变更需评估。

## 安全考虑

- **存量授权静默失效防护（高）**：回填 status=2（非 0）保留「有 grant 即活跃」语义，避免存量授权在 `FindActiveByUserId`（status=2 严格判定）下静默失效。触发 [[auto-grant-unverified-grant-confers-scope-level0]]——status∈{0,1,2} 活跃判定下，回填 2 不改变能力分层。
- **越权防护（高，已化解）**：003 无 guard 外 UPDATE，避免把 Insert 显式写入的 status=0（未认证自动授权）/status=4（已过期）行静默改写为 2 → level-0 → level-2 能力越权。
- **路径错误可见性**：前端路径对齐 + 静默 catch 消除，使路径错误/网络错误对用户可见（toast）+ 可排查（console），杜绝「空白无报错」排障黑洞。触发 [[verify-api-before-calling]]。
- **无 PII/明文日志风险**：console.error 输出错误对象（Error/网络错误），不落库、不写敏感字段。

## 记忆引用（设计阶段预防性注入，Step 1.5 产出）

| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| [[migration-must-execute]] | 数据模型（003 迁移） | 迁移三步闭环（写→提交→执行）；003 末尾 SELECT 验证列存在，Task 0.2 实际执行 + `DESCRIBE` 验证，不重蹈「提交未执行 → 1054」 |
| [[need-human-findbyroleid-assign_time]] | 数据模型（列契约） | 历史 1054 源于臆想的 `assign_time` 列而非三列；live 库已含三列（rel.go L248），003 对 live 库为 no-op，列定义严格对齐 `RelUserRole` db tag |
| [[auto-grant-unverified-grant-confers-scope-level0]] | 数据模型（回填语义） | 回填 status=2 而非 0，保留「有 grant 即活跃」；status∈{0,1,2} 活跃判定下，回填不改写显式 status=0/4，防能力越权 |
| [[verify-api-before-calling]] | 接口设计（前端路径对齐） | 调用前验证路由存在：`/api/community/lost-found`（未注册）→ `/api/community/lostfound`（已注册）；消除静默 catch |
| [[cross-service-seed-deployment-order]] | 业务流程（部署编排） | 从零建库顺序固定 migration.sql → init_permissions.sql，顺序颠倒记为部署编排约束 |
| [[unique-index-migration-dup-precheck]] | 数据模型（索引名对齐） | migration.sql `idx_user_role_scope` → `uk_user_role_scope` 仅对齐既有 uk 命名，非新增唯一索引，无重复数据阻塞风险 |

**不适用记忆**（主动排除，供 reviewer 确认非遗漏）：
- [[insert-ignore-swallows-errors]] — 本变更不改 Insert 逻辑（InsertIgnore 已存在且语义正确），仅补列使其不 1054；无新的 INSERT IGNORE 假成功风险引入。
- [[grpc-only-comms]] — 本变更无服务间通信、无 gRPC 调用新增，不涉及。

## 非功能设计（精简 checklist）

- [x] 可靠性：003 幂等 guard（顺序重跑安全）；前端失败不 rethrow（并发不阻断、页面不崩）；无数据回滚需求（DDL 幂等）
- [x] 性能：无分页/缓存/限流需求；003 为一次性 DDL，无运行时性能影响（标注「无」）
- [x] 可观测性：前端三处失败 `console.error`（错误对象 + 区块标识）供排查；后端迁移 SELECT 验证段输出 PASS/FAIL；无新增告警阈值
- [x] 无显式要求项已标注「无」（缓存/限流/大查询优化）

## 关键设计决策与权衡（ADR，轻量）

| 决策点 | 备选方案 | 最终选型 | 取舍理由 | 未采用方案原因 |
|-------|---------|---------|---------|--------------|
| 存量回填机制 | A. guard 外显式 UPDATE；B. ADD COLUMN 的 ALTER DEFAULT 在 guard 内自动回填 | B | 幂等天然成立、迁移后新行与显式值零改写、无「重跑二次回填」歧义 | A 会静默改写 status=0/4 显式行，且需额外幂等判定 |
| rel_user_role.id 是否改雪花 ID | A. 改 BIGINT Snowflake；B. 保留 AUTO_INCREMENT | B | 映射表 id 全仓代码不消费，rel.go 与种子 INSERT 均省略 id 依赖自增 | A 会在严格模式 1364 / INSERT IGNORE 静默丢种子行 |
| P1 修复方向 | A. 改后端路由加连字符别名；B. 改前端对齐后端 | B | 后端路由已注册且为契约，前端 404 属调用方错误；单点字符串修改零后端风险 | A 引入双路径维护成本与路由污染 |
| 迁移执行器 | A. 新增自动执行器；B. 沿用 001/002 手动执行 + harness-checks | B | 范围最小化（D5），迁移文件幂等，手动串行执行足够 | A 超出本变更范围、引入新基础设施 |
