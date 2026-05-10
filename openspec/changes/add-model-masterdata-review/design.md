## Context

主数据服务（masterdata）管理 4 张 MySQL 表：`md_residential_area`、`md_administrative_division`、`md_configuration`、`md_sensitive_word`。当前状态：

- **住宅小区**：完整的提交→审核→通过/拒绝流程，删除走审批（submission_status=4），有 reviewer 字段
- **行政区划、系统配置、敏感词**：仅有提交功能（submit），无审核逻辑，无 reviewer 字段，删除为直接软删除
- 所有表无法区分操作类型（新增/修改/删除），修改时无变更前后对比
- 前端审核页面只有住宅小区有（`Review.vue`），无统一入口

技术栈：Go-zero 1.6+（goctl 代码生成）、MySQL 8.0+、Vue 3.4+ / TypeScript / Element Plus

## Goals / Non-Goals

**Goals:**
- 统一审核 API：一个路由组 `/api/masterdata/approval/*` 聚合 4 张表的待审数据
- 操作类型追踪：`submission_type` 字段区分新增/修改/删除
- 变更快照：`change_snapshot` JSON 字段存储修改前数据，审核时可对比
- 删除审批：4 个实体统一走审批删除（不再直接软删除）
- 前端统一审核中心：统计卡片 + 待审列表 + 变更对比抽屉 + 批量审批

**Non-Goals:**
- 不做审核人角色/权限区分（当前所有审核人共享同一权限）
- 不做审核工作流引擎（如多级审批、会签等）
- 不迁移现有住宅小区的 `Review.vue`（新审核中心替代，旧页面保留但不强制迁移）
- 不处理已存在的历史数据快照补录

## Decisions

### D1: 快照存储方案 — 各表 JSON TEXT 列

**选择**: 每张表加 `change_snapshot TEXT NULL` 列，存储修改前的完整记录 JSON。

**备选方案**: 独立 `md_change_history` 表，每次变更写入一行。

**理由**: 独立表更规范化，但增加 JOIN 复杂度和额外查询。JSON TEXT 列方案简单直接——快照仅在审核详情页按单条 ID 读取，不存在查询性能问题。现有代码库已有 `sql.NullString` 处理 TEXT 列的模式，改动最小。

### D2: 待审列表聚合方式 — 后端聚合端点

**选择**: 新增 `GET /approval/pending-items` 后端聚合端点，后端分别查询 4 张表并合并。

**备选方案**: 前端分别调 4 个列表 API 在客户端聚合。

**理由**: 后端聚合统一分页和排序，避免前端发 4 个请求再合并的复杂性。按 entity_type 过滤时只查对应一张表；不按 entity_type 过滤时查 4 张表合并。实际数据量（待审几十到几百条）不足以产生性能问题。

### D3: 统一审批端点 vs 分实体审批端点

**选择**: 统一端点 `POST /approval/:entity_type/:id/review`，后端按 entity_type 分发到对应 model。

**备选方案**: 每个实体保留各自的 review 端点。

**理由**: 统一端点让前端只需调一个 API，后端通过 switch-case 分发。住宅小区现有的 `reviewResidentialArea` 端点保留兼容，统一端点内部复用相同逻辑。

### D4: 拒绝修改时的数据恢复

**选择**: 拒绝修改操作时，从 `change_snapshot` JSON 反序列化恢复原始数据，`submission_status` 设为 3（已拒绝），`submission_type` 清空。

**理由**: 拒绝意味着"不认可这次修改"，应恢复原状让提交人重新编辑。恢复数据来自提交修改时保存的快照，保证一致性。

### D5: `submission_type` 值设计

| 值 | 含义 | 设置时机 |
|---|---|---|
| 1 | 新增 | Create 时设置 |
| 2 | 修改 | Update 时设置（同时存 change_snapshot） |
| 3 | 删除 | Delete 时设置（同时设 submission_status=4） |

字段类型 `TINYINT NULL`——NULL 表示历史数据或不需要追踪的类型。

### D6: 审核人字段补齐方案

**选择**: 行政区划、系统配置、敏感词 3 张表各加 `submitter_id`, `submit_time`, `reviewer_id`, `review_time`, `review_notes` 5 个字段，与住宅小区保持一致。

**理由**: 统一 4 张表的审核字段，使统一审批逻辑不需要特殊处理某个实体。

## Risks / Trade-offs

**[跨表分页性能]** → 当用户不按 entity_type 过滤时，需查询 4 张表合并排序。数据量在几百条级别时无问题。若未来数据量增长，可引入 `md_approval_queue` 物化视图表。当前不优化。

**[快照 JSON schema 漂移]** → 表结构变更后，旧的 change_snapshot JSON 字段可能与新 struct 不匹配。缓解：解析快照到 `map[string]interface{}` 而非强类型 struct，缺失字段显示为空而非报错。

**[goctl 重新生成覆盖]** → 运行 `goctl api go` 会覆盖 `routes.go`、`types.go`、handler 文件。缓解：先运行 goctl 生成骨架，再修改 logic 文件（logic 文件不会被覆盖）。已手写的 handler 改动（responsex.Response 包装）需在生成后重新应用。

**[现有数据 submission_type 为 NULL]** → 历史已通过的数据没有 submission_type 值。这些数据不会出现在待审列表（submission_status=2），不影响功能。迁移脚本中回填 `submission_type=1` 仅用于数据完整性。

## Migration Plan

1. 执行 SQL ALTER 语句添加新字段
2. 回填历史数据：已通过记录设 `submission_type=1`
3. 修改后端 model `_gen.go` 文件（struct + Insert/Update SQL）
4. 修改 `.api` 文件，运行 goctl 生成新骨架
5. 实现 approval logic
6. 修改现有 create/update/delete/submit logic
7. 构建部署后端
8. 实现前端审核中心
9. 验证完整流程

**回滚**: 新增字段均为 NULL 或有默认值，不影响现有功能。回滚只需重新部署旧二进制。

## Open Questions

（无）
