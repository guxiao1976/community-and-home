# Tasks: rel-user-role-migration-publish-fix — rel_user_role 生命周期列迁移 + 移动端寻失列表路径/静默错误修复

> **对执行 Agent 的指令**：每个 Task 独立可测，精确到文件路径，零占位符。含逻辑代码的 Task 按 TDD（先写测试→看失败→看通过）。
> **依赖顺序**：P0 迁移（2.1 → 2.2 三态库验证）→ P0 文档修正（1.1，可先行）→ P0 从零建库验证（2.3，依赖 1.1 + 2.1）→ P1 前端（3.1 路径 → 3.2 错误处理 → 3.3 门禁）→ 待办登记（4.1/4.2，可并行）。
> **硬性约束**：permission-service 提交前 `bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service`；web/mobile 提交前 `cd web/mobile && npm run test:unit` + `npm run type-check`。
> **本变更无 Proto/api-proto 变更**（`proto_change_required=false`）；`docs/specs/migration.sql` 为全仓权威建表文档，由全局 Claude 修改（Task 1.1），不分发子 Claude。

---

## 全局 / Proto（由全局 Claude 执行）

本变更 **proto_change_required=false**，无 api-proto 变更。`docs/specs/migration.sql` 为全仓权威从零建库 DDL，由全局协调层修改。

### Task 1.1: P0 — 修正 migration.sql §3.4 rel_user_role 段（REQ-P0-5）
- **修改**: `docs/specs/migration.sql`（§3.4 `rel_user_role` 建表段，约 L224-236）
- [ ] 补齐三列（与 REQ-P0-3 契约严格一致，**NOT NULL DEFAULT 2**，非 `DEFAULT 0`、非无默认）：
  - `status INT NOT NULL DEFAULT 2 COMMENT '个体角色生命周期: 0=未认证 1=待审 2=已认证 3=已驳回 4=已过期'`
  - `verified_at DATETIME NULL COMMENT '个体认证通过时间'`
  - `expires_at DATETIME NULL COMMENT '个体角色到期时间, NULL=永久'`
- [ ] **id 保留 AUTO_INCREMENT**（REVISION #15 方案 a）：`id BIGINT NOT NULL AUTO_INCREMENT` 不变，不改为雪花 ID
- [ ] `created_time` → `created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`
- [ ] `UNIQUE INDEX idx_user_role_scope` → `uk_user_role_scope`；全文档 grep 无 `idx_user_role_scope` 残留（REQ-P0-5 Scenario 2）
- [ ] 兄弟表（`rel_role_permission` 等）不触碰（REQ-P0-7）
- [ ] TDD 不适用（文档 DDL）；验证由 Task 2.3 从零建库覆盖

---

## permission-service

### Task 2.1: P0 — 编写 003_add_role_lifecycle.sql（REQ-P0-1/2/3/4）
- **创建**: `services/permission-service/migration/003_add_role_lifecycle.sql`
- **修改**: `services/permission-service/CHANGELOG.md`
- [ ] 沿用 001/002 的 information_schema guard 写法（`SET @col := (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='rel_user_role' AND COLUMN_NAME=...)` + `IF(...)` + `PREPARE/EXECUTE`），**逐列 guard**（status/verified_at/expires_at 各一段）
- [ ] 列定义（REQ-P0-3）：
  - `ALTER TABLE rel_user_role ADD COLUMN status INT NOT NULL DEFAULT 2 COMMENT '...'`
  - `ALTER TABLE rel_user_role ADD COLUMN verified_at DATETIME NULL COMMENT '...'`
  - `ALTER TABLE rel_user_role ADD COLUMN expires_at DATETIME NULL COMMENT '...'`
- [ ] **零 guard 外 UPDATE**：003 不含对 `rel_user_role` 的 UPDATE 语句；存量回填由 ALTER DEFAULT 在补列当次自动完成（REQ-P0-4）——`ADD COLUMN status INT NOT NULL DEFAULT 2` 使存量行置 2，`verified_at`/`expires_at DATETIME NULL` 使存量行置 NULL
- [ ] 末尾 SELECT 验证三列存在（仿 002 的验证段，输出 ✅ PASS/❌ FAIL）
- [ ] **changelog 只登记 003（q3/D15）**：CHANGELOG.md 登记 `003_add_role_lifecycle.sql`，不重复登记 001/002
- [ ] TDD 不适用（DDL 脚本）；执行验证由 Task 2.2 覆盖（SEE [[migration-must-execute]]——迁移三步闭环：写→提交→执行）

### Task 2.2: P0 — 三态库执行 003 并验证幂等（REQ-P0-2/4/6）
- **目录**: `services/permission-service/migration/`
- [ ] **三态库构造（D14/q2）**：临时构造三种库状态各执行一次 003——(a) 旧 migration.sql 从零库（缺三列 + `created_time`）、(b) 生产 live 库（已含三列 + `created_at`，对齐 rel.go L248 注释）、(c) 修正后 migration.sql 从零库（已含三列）——验证补列 vs no-op（REQ-P0-6 Scenario 2）
- [ ] **库状态唯一化（D12）**：(a) 旧从零库缺三列、时间为 `created_time` → 003 补三列、**不 rename created_time→created_at**；(b) 生产 live 库已含三列（历史 1054 源于不存在的 `assign_time` 列而非这三列，SEE [[need-human-findbyroleid-assign_time]]）→ 003 no-op、不回填
- [ ] 对 (a) 执行 003：无 1054/Duplicate column；`DESCRIBE rel_user_role` 确认三列存在（SEE [[migration-must-execute]]）
- [ ] 验证存量回填：对 (a) `SELECT status, verified_at, expires_at FROM rel_user_role` 存量行全部 `status=2 / NULL / NULL`；对 (b) 执行 003 后存量数据不变（no-op 不回填）
- [ ] **幂等重跑**：再次执行 003 → 无报错、表结构不变、存量数据不被重复改写（guard 跳过 ADD 与回填）
- [ ] **显式值不被改写**：在已含三列 + 显式 status=0/4 行的库上执行 003 → 显式值不被改写（REQ-P0-4 边界）
- [ ] 部分列缺失库执行 003 → 只补缺失列（逐列 guard，REQ-P0-2 边界）

### Task 2.3: P0 — 从零建库验证 init_permissions.sql（REQ-P0-6，仅验证不改 docs/specs）
- **目录**: `services/permission-service/scripts/`（只读引用 `docs/specs/migration.sql` 与 `services/permission-service/migration/003_add_role_lifecycle.sql` 作为输入，**不修改 docs/specs**）
- [ ] 用 Task 1.1 修正后 migration.sql 从零建库 → 按序执行 `init_permissions.sql` → 成功，无 `MySQL 1054 Unknown column 'status'`，`:238` 的 rel_user_role INSERT 成功（REQ-P0-6 Scenario 1）
- [ ] **省略 id 的写入验证**：仿 rel.go `Insert`（省略 id，仅 user_id/role_id/scope_type/scope_id/status/verified_at/expires_at）执行 → 成功，无 MySQL 1364 / id=0 主键冲突（REQ-P0-5 Scenario 3）
- [ ] **从零库重跑 003（REVISION #1 补：让验收标准（c）有执行载体）**：在从零库（含 init_permissions.sql 已灌入的种子行）上**显式执行一次 003** → guard 探测三列已存在、跳过 ADD 与回填、无报错、表结构与存量种子数据不变（REQ-P0-2「从零库安全通过」）
- [ ] **失败边界（不改 docs/specs）**：若 init_permissions.sql 暴露与 rel_user_role 三列无关的失败 → **上报 Owner**：源于本变更 rel_user_role DDL 修正副作用的，回路由 **Task 1.1（全局 Claude）** 修复 `docs/specs/migration.sql`；否则由全局登记待办（BACKLOG）。本 Task **不直接修改 `docs/specs/migration.sql`**、不擅自扩大迁移范围（REQ-P0-6 Scenario 4）
- [ ] 顺序约束记录：migration.sql → init_permissions.sql（记入 [[cross-service-seed-deployment-order]]）

### Task 2.4: P0 — permission-service 门禁（REQ-P0 验收）
- **目录**: `services/permission-service/`
- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过
- [ ] `bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service` 通过（003 迁移 + 文档修正不引入 Go 代码回归）

---

## 前端 web/mobile

### Task 3.1: P1 — community.ts 寻失列表路径对齐（REQ-P1-PATH-1/2）
- **修改**: `web/mobile/src/api/community.ts`（L156）
- [ ] `getLostFoundList` 请求路径 `/api/community/lost-found` → `/api/community/lostfound`
- [ ] 全仓 grep 确认无 `/api/community/lost-found`（带连字符）调用残留（REQ-P1-PATH-1 边界）
- [ ] **后端不动**（REQ-P1-PATH-2）：community-hub-service `services/community-hub-service/api/internal/handler/routes.go` 零改动
- [ ] TDD 不适用（单路径字符串修改）；验证由 Task 3.3 门禁覆盖

### Task 3.2: P1 — notice.vue 三处静默 catch 消除（REQ-P1-ERR-1/2）
- **修改**: `web/mobile/src/pages/notice/notice.vue`
- **修改**: `web/mobile/src/pages/notice/notice.spec.ts`
- [ ] `fetchLostFound`/`fetchNotices`/`fetchContacts` 三处 `catch { /* silent */ }` → catch 内：`console.error`（错误对象 + 区块标识）+ `uni.showToast`（区块相关失败提示，如「寻失加载失败」/「通知加载失败」/「联络加载失败」，`icon: 'none'`）；**不 rethrow**（REQ-P1-ERR-2）
- [ ] 成功路径不弹错误 toast（REQ-P1-ERR-1）
- [ ] **RED**: notice.spec.ts 补用例（复用既有 `vi.mock('@/api/community')`，`mockRejectedValue`）：
  - `getLostFoundList` 失败 → `uni.showToast` 被调用（含「寻失」提示）+ `console.error` 被调用
  - `getNotices`/`getContacts` 失败同理（三区块各自断言）
  - **三请求并发全部失败（REVISION #2 补：核心触发场景）**：三个 API 全 mock 失败 → `uni.showToast` 至少调用一次 + `console.error` 恰好三次 + 页面不崩（loadAll 正常结束、loading 复位）
  - 局部失败（fetchNotices 失败 + 其余成功）→ 成功区块数据仍渲染（Promise.all 不被阻断）
  - 成功场景 → 无错误 toast
  - `cd web/mobile && npm run test:unit` → 看 FAIL（新用例）
- [ ] **GREEN**: 实现 → 看 PASS
- [ ] **REFACTOR**: 保持绿

### Task 3.3: P1 — 前端门禁（REQ-P1 验收）
- **目录**: `web/mobile/`
- [ ] `npm run test:unit` 全绿（含 notice.spec.ts 新增失败路径用例）
- [ ] `npm run type-check` 通过（vue-tsc 无类型错误）

---

## 待办登记（BACKLOG，可并行，无代码变更）

### Task 4.1: P0 — 兄弟表不一致登记待办（REQ-P0-7）
- **修改**: `.harness/tasks/BACKLOG.md`
- [ ] 登记 `rel_role_permission` 等兄弟表 AUTO_INCREMENT/created_time 结构不一致项为待办（P2/P3），本次不落库
- [ ] 确认本次 diff 不含兄弟表 DDL 改动（REQ-P0-7 回归防护）

### Task 4.2: P1 — 其余静默点登记待办（REQ-P1-ERR-3，D13/q1）
- **修改**: `.harness/tasks/BACKLOG.md`
- [ ] 登记 `stores/community.ts`、`join-community.vue`、`onCommunitySwitch` 非 10015 分支等静默吞错点为独立后续任务（P1/P2），本次不改动这些文件
- [ ] 确认本次 diff 不含 `stores/community.ts`/`join-community.vue`/`onCommunitySwitch` 非 10015 分支的改动（REQ-P1-ERR-3 回归防护）
