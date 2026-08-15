# Tasks: rel-user-role-migration-publish-fix — rel_user_role 生命周期列迁移 + 移动端寻失列表路径/静默错误修复

> **对执行 Agent 的指令**：每个 Task 独立可测，精确到文件路径，零占位符。逻辑代码按 TDD（先写测试→看失败→看通过）。
> **依赖顺序**：P0 迁移（0.1 → 0.2）→ P0 文档修正（1.1）→ P0 从零建库验证（1.2，依赖 0.1/1.1 完成）→ P0 待办登记（1.3，可并行）→ P1 前端（2.1 路径 → 2.2 错误处理 → 2.3 门禁）。
> **硬性约束**：P0 提交前 `bash .harness/skills/qa/scripts/harness-checks.sh --service permission-service`；P1 提交前 `cd web/mobile && npm run test:unit` + `npm run type-check`。
> **本变更无 Proto/api-proto 变更**（proto_change_required=false）；`docs/specs/migration.sql` 为全仓权威建表文档，由全局 Claude 修改（Task 1.1），不分发子 Claude。

---

## 全局 / Proto（占位说明）

本变更 **proto_change_required=false**，无 api-proto 变更。`docs/specs/migration.sql` 修正归全局 Claude（Task 1.1）。

### Task 0.1: P0 — 编写 003_add_role_lifecycle.sql（REQ-P0-1/2/3/4）
- **创建**: `services/permission-service/migration/003_add_role_lifecycle.sql`
- [ ] 沿用 001/002 的 information_schema guard 写法（`SET @col := (SELECT COUNT(*) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA=DATABASE() AND TABLE_NAME='rel_user_role' AND COLUMN_NAME=...)` + `PREPARE/EXECUTE`），**逐列 guard**
- [ ] 列定义（REQ-P0-3）：
  - `status INT NOT NULL DEFAULT 2 COMMENT '个体角色生命周期: 0=未认证 1=待审 2=已认证 3=已驳回 4=已过期'`
  - `verified_at DATETIME NULL COMMENT '个体认证通过时间'`
  - `expires_at DATETIME NULL COMMENT '个体角色到期时间, NULL=永久'`
- [ ] **零 guard 外 UPDATE**：003 不含对 rel_user_role 的 UPDATE 语句；存量回填由 ALTER DEFAULT 在补列当次自动完成（REQ-P0-4）——`ADD COLUMN status INT NOT NULL DEFAULT 2` 使存量行置 2，`ADD COLUMN verified_at DATETIME NULL`/`expires_at DATETIME NULL` 使存量行置 NULL
- [ ] 末尾 SELECT 验证三列存在（仿 002 的验证段，输出 ✅ PASS/❌ FAIL）
- [ ] TDD 不适用（DDL 脚本）；验证由 Task 0.2 实际执行

### Task 0.2: P0 — 执行 003 到 live 库并验证幂等（REQ-P0-2/4）
- **目录**: `services/permission-service/migration/`
- [ ] **「旧结构库」定义**：仅缺 `status`/`verified_at`/`expires_at` 三列、`created_at` 已存在（对齐 rel.go L248 注释所述生产库实际结构）；若环境是修正前 migration.sql 建的 created_time 旧库，003 也仅补三列、**不 rename created_time→created_at**（REQ-P0-2 边界，见 REVISION #3）
- [ ] 对旧结构库执行 003：无 1054/Duplicate column；`DESCRIBE rel_user_role` 确认三列存在（SEE [[migration-must-execute]]——迁移三步闭环：写→提交→执行）
- [ ] 验证存量回填：`SELECT status, verified_at, expires_at FROM rel_user_role` 存量行全部 `status=2 / NULL / NULL`
- [ ] **幂等重跑**：再次执行 003 → 无报错、表结构不变、存量数据不被重复改写（guard 跳过 ADD 与回填）
- [ ] **显式值不被改写**：在已含三列 + 显式 status=0/4 行的库上执行 003 → 显式值不被改写（REQ-P0-4 边界）
- [ ] 部分列缺失库执行 003 → 只补缺失列（逐列 guard）

---

## docs/specs

### Task 1.1: P0 — 修正 migration.sql §3.4 rel_user_role 段（REQ-P0-5）
- **修改**: `docs/specs/migration.sql`（§3.4 rel_user_role 建表段）
- [ ] 补齐三列：`status INT NOT NULL DEFAULT 2`、`verified_at DATETIME NULL`、`expires_at DATETIME NULL`（与 REQ-P0-3 契约一致，**NOT NULL DEFAULT 2**，非 `DEFAULT 0`、非无默认）
- [ ] **id 保留 AUTO_INCREMENT**（REVISION #15 方案 a）：`id BIGINT NOT NULL AUTO_INCREMENT` 不变，不改为雪花 ID
- [ ] `created_time` → `created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP`
- [ ] `UNIQUE INDEX idx_user_role_scope` → `uk_user_role_scope`；全文档 grep 无 `idx_user_role_scope` 残留
- [ ] 兄弟表（rel_role_permission 等）不触碰（REQ-P0-7）
- [ ] TDD 不适用（文档 DDL）；验证由 Task 1.2 从零建库覆盖

### Task 1.2: P0 — 从零建库验证 init_permissions.sql（REQ-P0-6）
- **目录**: `services/permission-service/scripts/` + `docs/specs/`
- [ ] 用 Task 1.1 修正后 migration.sql 从零建库 → 按序执行 `init_permissions.sql` → 成功，无 `MySQL 1054 Unknown column 'status'`，:238 的 rel_user_role INSERT 成功
- [ ] **省略 id 的写入验证**：仿 rel.go `Insert`（省略 id，仅 user_id/role_id/scope_type/scope_id/status/verified_at/expires_at）执行 → 成功，无 MySQL 1364 / id=0 主键冲突（REQ-P0-5 边界）
- [ ] **从零库重跑 003（REVISION #1 补：让验收标准（c）有执行载体）**：在从零库（含 init_permissions.sql 已灌入的种子行）上**显式执行一次 003** → guard 探测三列已存在、跳过 ADD 与回填、无报错、表结构与存量种子数据不变（REQ-P0-2「从零库安全通过」）
- [ ] **失败边界**：若 init_permissions.sql 暴露与 rel_user_role 三列无关的失败 → 源于本变更 DDL 修正的在 migration.sql 侧修；否则登记待办（BACKLOG），不扩范围（REQ-P0-6 Scenario 3）
- [ ] 顺序约束记录：migration.sql → init_permissions.sql（记入 [[cross-service-seed-deployment-order]]）

### Task 1.3: P0 — 兄弟表不一致登记待办（REQ-P0-7）
- **修改**: `.harness/tasks/BACKLOG.md`
- [ ] 登记 `rel_role_permission` 等兄弟表 AUTO_INCREMENT/created_time 结构不一致项为待办（P2/P3，来源 review/spec），本次不落库

---

## 前端 web/mobile

### Task 2.1: P1 — community.ts 寻失列表路径对齐（REQ-P1-PATH-1/2）
- **修改**: `web/mobile/src/api/community.ts`（L156）
- [ ] `getLostFoundList` 请求路径 `/api/community/lost-found` → `/api/community/lostfound`
- [ ] 全仓 grep 确认无 `/api/community/lost-found`（带连字符）调用残留（REQ-P1-PATH-1 边界）
- [ ] **后端不动**（REQ-P1-PATH-2）：community-hub-service `services/community-hub-service/api/internal/handler/routes.go` 零改动
- [ ] TDD 不适用（单路径字符串修改）；验证由 Task 2.3 门禁 + 既有单测覆盖

### Task 2.2: P1 — notice.vue 三处静默 catch 消除（REQ-P1-ERR-1/2）
- **修改**: `web/mobile/src/pages/notice/notice.vue`
- **修改**: `web/mobile/src/pages/notice/notice.spec.ts`
- [ ] `fetchLostFound`/`fetchNotices`/`fetchContacts` 三处 `catch { /* silent */ }` → catch 内：`console.error`（输出错误对象 + 区块标识）+ `uni.showToast`（区块相关失败提示，如「寻失加载失败」/「通知加载失败」/「联络加载失败」，icon: 'none'）；**不 rethrow**（REQ-P1-ERR-2）
- [ ] 成功路径不弹错误 toast（REQ-P1-ERR-1）
- [ ] **RED**: notice.spec.ts 补用例（复用既有 `vi.mock('@/api/community')`，`mockRejectedValue` 单测）：
  - `getLostFoundList` 失败 → `uni.showToast` 被调用（含「寻失」提示）+ `console.error` 被调用
  - `getNotices`/`getContacts` 失败同理（三区块各自断言）
  - **三请求并发全部失败（REVISION #2 补：核心触发场景）**：三个 API 全 mock 失败 → `uni.showToast` 至少调用一次 + `console.error` 恰好三次 + 页面不崩（loadAll 正常结束、loading 复位）
  - 局部失败（fetchNotices 失败 + 其余成功）→ 成功区块数据仍渲染（Promise.all 不被阻断）
  - 成功场景 → 无错误 toast
  - `cd web/mobile && npm run test:unit` → 看 FAIL（新用例）
- [ ] **GREEN**: 实现 → 看 PASS
- [ ] **REFACTOR**: 保持绿

### Task 2.3: P1 — 前端门禁（REQ-P1-ERR-1 验收）
- **目录**: `web/mobile/`
- [ ] `npm run test:unit` 全绿（含 notice.spec.ts 新增失败路径用例）
- [ ] `npm run type-check` 通过（vue-tsc 无类型错误）
- [ ] `npm run build:h5`（或等价前端构建）通过
