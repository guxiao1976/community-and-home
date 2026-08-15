# P0 rel_user_role 生命周期列补齐 Specification

## Purpose

补齐 rel_user_role 表生命周期三列（status/verified_at/expires_at）的幂等迁移与权威从零建表脚本，消除从零建库时 MySQL 1054 Unknown column 'status' 导致的权限/发布链路初始化阻断，同时保证存量授权的「有 grant 即活跃」语义在迁移后不失效，且迁移对迁移后新行与显式生命周期值零改写。

## Requirements

### Requirement: REQ-P0-1 交付 003 迁移文件

permission-service SHALL ship a migration file at `services/permission-service/migration/003_add_role_lifecycle.sql` that follows the idempotent information_schema guard style used by `001_scope_three_state.sql` and `002_add_role_platforms.sql`, guarding each of the three lifecycle columns individually.

#### Scenario: 交付物沿用既有幂等写法（正向）

- **GIVEN** `migration/` 目录已存在 `001_scope_three_state.sql` 与 `002_add_role_platforms.sql`
- **WHEN** 本变更完成
- **THEN** `003_add_role_lifecycle.sql` 存在，且对每个生命周期列使用 information_schema.COLUMNS 探测列存在后 `IF` 执行 `ALTER TABLE ... ADD COLUMN`，与 001/002 同款 guard 模式（PREPARE/EXECUTE）

#### Scenario: 未沿用 guard 则验收失败（边界）

- **GIVEN** 003 使用裸 `ALTER TABLE ... ADD COLUMN` 而非 information_schema guard
- **WHEN** 在已含三列的库上执行该迁移
- **THEN** 报 Duplicate column 错误，验收不通过（guard 是本变更的幂等硬性要求）

### Requirement: REQ-P0-2 三列幂等添加

The migration SHALL add columns `status`, `verified_at`, and `expires_at` to table `rel_user_role` such that execution succeeds without error on a legacy live database (columns missing) and on an already-migrated or freshly-created database (columns present); a re-run SHALL NOT alter the table structure or the data already present. The migration SHALL add exactly these three columns and nothing else — it SHALL NOT rename or introduce `created_at` on an existing database (specifically SHALL NOT perform a `created_time` → `created_at` column migration); the `created_time` (legacy DDL) vs `created_at` (Go model) difference on non-fresh databases is accepted because `RelUserRole.CreatedTime` (`db:"created_at"`) is never consumed by any code, and fresh builds get `created_at` from the corrected `docs/specs/migration.sql`.

#### Scenario: 旧结构 live 库正常补列（正向）

- **GIVEN** 旧结构 live 库 = 生产库实际结构（对齐 `rel.go` L248 注释：live 库仅有 `id`/`user_id`/`role_id`/`scope_type`/`scope_id`/`status`/`verified_at`/`expires_at`/`created_at`），即 `created_at` 已存在、仅缺 `status`/`verified_at`/`expires_at` 三列
- **WHEN** 执行 003
- **THEN** 三列被添加，无 MySQL 1054 或 Duplicate column 报错；补列后列结构与 Go 模型 `RelUserRole` 的 db tag 兼容（`status` int64 / `verified_at` sql.NullTime / `expires_at` sql.NullTime；`created_at` 已存在，与 `db:"created_at"` 匹配）

#### Scenario: 幂等重跑（列已存在）跳过补列与回填

- **GIVEN** 003 已成功执行过，三列已存在
- **WHEN** 再次执行 003
- **THEN** guard 探测到列已存在，**同时跳过 ADD COLUMN 与存量回填**，执行不报错，表结构不变，存量数据不被重复改写

#### Scenario: 从零库（新结构已含三列）安全通过

- **GIVEN** 数据库由修正后的 `docs/specs/migration.sql` 从零建立，`rel_user_role` 已含三列
- **WHEN** 执行 003
- **THEN** guard 跳过 ADD 与回填，003 安全通过，无错误

#### Scenario: 部分列缺失（边界输入）

- **GIVEN** 仅 `status` 缺失，`verified_at`/`expires_at` 已存在
- **WHEN** 执行 003
- **THEN** 只补 `status`（逐列 guard），已有列不重复添加，最终三列齐全

#### Scenario: 旧 migration.sql 建库（created_time 旧结构）仅补列不重命名（边界）

- **GIVEN** 库由修正前的 `docs/specs/migration.sql` 建立，`rel_user_role` 使用 `created_time` 列、缺三列（即 request.md 描述的「缺三列且为 AUTO_INCREMENT/created_time 旧结构」）
- **WHEN** 执行 003
- **THEN** 003 仅补 `status`/`verified_at`/`expires_at` 三列，**不做 `created_time` → `created_at` 重命名**；`created_at` 缺失差异被显式接受——`RelUserRole.CreatedTime`（`db:"created_at"`）无任何代码消费（`helpers.go` 仅读 `SysRole`/`SysPermission` 的 `CreatedTime`，见 L143/L177，rel_user_role 的 `CreatedTime` 零值无害）；从零建库的 `created_at` 由 Task 1.1 修正后的 migration.sql 保证，实现者不得借 003 顺手改 `created_time`

### Requirement: REQ-P0-3 列定义语义与模型一致

The added column definitions SHALL preserve the semantic contract consumed by `rel_user_role` reads and writes: `status` SHALL be `INT NOT NULL DEFAULT 2` (non-null integer whose default does not silently deactivate a grant written without an explicit value — default `2`, not `0`); `verified_at` and `expires_at` SHALL be nullable `DATETIME` (`DATETIME NULL`) where NULL respectively means "not yet verified" and "permanent (never expires)".

#### Scenario: 默认值不静默失效授权（正向）

- **GIVEN** 003 添加的 `status` 列定义为 `INT NOT NULL DEFAULT 2`
- **WHEN** 一个不带 `status` 的外部 INSERT 写入 `rel_user_role`
- **THEN** 该行默认状态为 `2`，与「有 grant 即活跃」语义一致，不因缺失显式值而在 status∈{0,1,2} 活跃判定下意外失效

#### Scenario: expires_at NULL = 永久（正向）

- **GIVEN** 某行 `expires_at` 为 NULL
- **WHEN** `FindActiveByUserId` / `FindActiveRolesByUserId` 执行过期判定
- **THEN** 该行视为未过期（与既有 SQL 谓词 `expires_at IS NULL OR expires_at > NOW()` 一致）

#### Scenario: 默认值与 Go 模型显式写入无冲突（边界）

- **GIVEN** 运行时代理 `Insert`/`BatchInsertUserRoles` 总是显式写入 `status`/`verified_at`/`expires_at`（rel.go L149/160/231）
- **WHEN** 上述 INSERT 执行
- **THEN** 列 DEFAULT 仅在外部 INSERT 缺列时兜底，Go 代理写入不受 DEFAULT 影响（DEFAULT 不改变显式值）

### Requirement: REQ-P0-4 存量行回填（唯一机制，零 guard 外 UPDATE）

The migration SHALL backfill every row present at the moment the columns are first added to `status=2` (verified), `verified_at=NULL`, `expires_at=NULL`, preserving the legacy "grant exists = active" semantics, and SHALL NOT backfill to `status=0`. Backfill SHALL be effectuated solely by the column definitions applied inside the guard branch: `status INT NOT NULL DEFAULT 2` makes MySQL set existing rows to `2` at `ADD COLUMN` time, and `verified_at`/`expires_at DATETIME NULL` makes existing rows `NULL`. The migration SHALL NOT contain any `UPDATE` statement on `rel_user_role` that runs outside this guard branch; consequently rows written after migration (including explicit `status`/`verified_at`/`expires_at` values) are never rewritten, and pre-existing explicit lifecycle values on databases where the columns already exist are never overwritten.

#### Scenario: 存量行统一回填为已认证（正向）

- **GIVEN** 迁移执行前 `rel_user_role` 存在存量行（三列缺失，无 status 值）
- **WHEN** 003 首次执行（guard 探测列缺失 → 补列）
- **THEN** 存量行经 ALTER DEFAULT 变为 `status=2, verified_at=NULL, expires_at=NULL`，存量授权在 `FindActiveByUserId`（status=2 严格判定）下继续生效，不静默失效

#### Scenario: 回填不触碰迁移后新行（边界）

- **GIVEN** 003 已执行，之后经 `Insert`/`BatchInsertUserRoles` 显式写入新行（含明确 status=0/4、verified_at/expires_at）
- **WHEN** 003 再次执行或数据被检查
- **THEN** 新行的显式生命周期值不被回填改写（003 无 guard 外运行的 UPDATE，重跑命中 guard 即跳过）

#### Scenario: 已存在显式 status 的存量行不被改写（边界）

- **GIVEN** 库已含三列（生产库恢复/半迁移环境），且存量行带显式 `status=0`（未认证）或 `status=4`（已过期）等值
- **WHEN** 执行 003
- **THEN** guard 探测列已存在 → 跳过 ADD 与回填，存量行的显式 status/verified_at/expires_at 不被改写（未认证 grant 不被静默提升为已认证，过期 grant 不被重置为永久）

#### Scenario: 回填语义与 sys_admin 先例一致（正向）

- **GIVEN** `init_permissions.sql:238-239` 的 `sys_admin(0,8,'global',0,2)` 先例为 status=2 + verified_at 缺省
- **WHEN** 003 对存量行回填
- **THEN** 回填结果与既有种子先例语义一致（status=2、verified_at=NULL），无新的语义分裂

### Requirement: REQ-P0-5 权威建表脚本修正

The authoritative from-scratch DDL `docs/specs/migration.sql` (§3.4 `rel_user_role`) SHALL be corrected so that: (a) the table includes the three lifecycle columns with the exact definitions `status INT NOT NULL DEFAULT 2`, `verified_at DATETIME NULL`, `expires_at DATETIME NULL`; (b) `id` REMAINS `BIGINT NOT NULL AUTO_INCREMENT` (the mapping-table id is never consumed by code and both `rel.go` inserts and the `init_permissions.sql:238` seed INSERT omit it, so removing AUTO_INCREMENT would break runtime and seed writes); (c) the time column is named `created_at` (`DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`); (d) the unique index is named `uk_user_role_scope`.

#### Scenario: 修正后 DDL 与新结构一致（正向）

- **GIVEN** `docs/specs/migration.sql` 是跨服务从零建库的权威脚本
- **WHEN** 检查其 §3.4 `rel_user_role` 建表段
- **THEN** 含 `status INT NOT NULL DEFAULT 2` / `verified_at DATETIME NULL` / `expires_at DATETIME NULL` 三列；`id` 为 `BIGINT NOT NULL AUTO_INCREMENT`；时间列为 `created_at`（有默认值）；唯一索引名为 `uk_user_role_scope`

#### Scenario: 索引名与 001 对齐（无残留）

- **GIVEN** 001 已在 live 库建立唯一索引 `uk_user_role_scope`
- **WHEN** 从零建库走修正后 migration.sql
- **THEN** 唯一索引名同为 `uk_user_role_scope`，文档中无 `idx_user_role_scope` 残留（命名一致）

#### Scenario: id 保留自增后省略 id 的写入可执行（边界）

- **GIVEN** `rel.go` 的 `Insert`/`BatchInsertUserRoles` 与 `init_permissions.sql:238` 种子 INSERT 均省略 `id`
- **WHEN** 在修复后从零库上执行这些写入
- **THEN** 写入成功（id 由 AUTO_INCREMENT 生成），无 MySQL 1364（Field 'id' doesn't have a default value）或 id=0 主键冲突

### Requirement: REQ-P0-6 从零建库流程可执行

The from-scratch database build sequence (apply `docs/specs/migration.sql`, then `init_permissions.sql`) SHALL complete without MySQL error 1054, specifically for the `rel_user_role` INSERT at `init_permissions.sql:238` that references the `status` column. If verification surfaces a failure in `init_permissions.sql` unrelated to the `rel_user_role` lifecycle columns, that failure is out of scope unless it is a direct consequence of this change's DDL correction: failures traceable to this change's `rel_user_role` DDL SHALL be fixed in `docs/specs/migration.sql`; other failures SHALL be recorded as backlog, not silently absorbed or fixed by expanding this change's migration scope.

#### Scenario: 从零建库完整流程无 1054（正向）

- **GIVEN** 数据库由修正后 migration.sql 从零建立，`rel_user_role` 含 `status` 列
- **WHEN** 按序执行 `init_permissions.sql`（含 :238 的 `INSERT INTO rel_user_role (user_id, role_id, scope_type, scope_id, status) VALUES ...`）
- **THEN** 执行成功，无 `MySQL 1054 Unknown column 'status'`

#### Scenario: 建表/种子顺序约束（权限场景）

- **GIVEN** 从零建库
- **WHEN** 先执行 `init_permissions.sql`、后执行 migration.sql
- **THEN** 因 `rel_user_role` 未建表而失败；部署编排 SHALL 固定为 migration.sql → init_permissions.sql 顺序（记入 [[cross-service-seed-deployment-order]] 约束），而非依赖 003 兜底

#### Scenario: 超出 rel_user_role.status 的失败边界（边界）

- **GIVEN** 从零建库流程中 `init_permissions.sql` 暴露了与 `rel_user_role` 三列无关的失败（如其他表/其他种子缺失）
- **WHEN** 验证该失败
- **THEN** 若失败源于本变更 rel_user_role DDL 修正的副作用 → 在 migration.sql 侧修；否则登记待办（`.harness/tasks/BACKLOG.md`），不擅自扩大本次迁移范围

### Requirement: REQ-P0-7 兄弟表保持不动

The change SHALL NOT modify the schema of sibling tables (e.g. `rel_role_permission` and other tables with the AUTO_INCREMENT/`created_time` legacy structure); their inconsistencies SHALL be recorded as backlog only.

#### Scenario: 兄弟表差异仅登记待办（正向）

- **GIVEN** `rel_role_permission` 等兄弟表仍为 AUTO_INCREMENT/`created_time` 结构
- **WHEN** 本变更完成
- **THEN** 这些表不被改动，差异已登记到任务待办（`.harness/tasks/BACKLOG.md` 或既有任务系统），供后续变更处理

#### Scenario: 变更引入兄弟表改动（回归防护）

- **GIVEN** `rel_user_role` 三列已处理完毕
- **WHEN** 评审 diff 时发现 `rel_role_permission` 等兄弟表被顺带修改
- **THEN** 该改动被拒绝（超出 D4 范围），兄弟表差异仅允许登记待办、不允许在本次变更中落库
