## Context

主数据服务采用两阶段软删除机制：用户提交删除请求 → 管理员审核通过后设置 `delete_time = now()`。所有正常查询通过 `WHERE delete_time IS NULL` 排除已删除记录。现有恢复逻辑嵌入在审核流程中（拒绝删除请求时从 `change_snapshot` 恢复），但审核通过后数据已标记为删除，没有独立的管理界面可以浏览和恢复。

后端使用 go-zero 框架，前端使用 Vue 3 + Element Plus + TypeScript。

## Goals / Non-Goals

**Goals:**
- 提供独立的后端 API 查询已软删除记录和执行恢复操作
- 提供与审核中心风格一致的删除数据恢复页面
- 支持按实体类型筛选浏览和一键恢复

**Non-Goals:**
- 批量恢复功能（后续按需添加）
- 永久删除（物理删除）功能
- 恢复操作的审核流程（直接恢复，无需审批）

## Decisions

### 1. 恢复方式：直接置空 delete_time

选择直接将 `delete_time` 置为 NULL 来恢复记录，而非走审核流程。

**理由**：删除数据恢复页面本身是管理员操作，且恢复是低风险操作（可再次删除），无需额外的审批环节。

**备选方案**：通过审核流程恢复 — 增加了不必要的复杂度。

### 2. API 设计：统一查询接口

使用统一接口 `GET /api/masterdata/deleted-items` + `entity_type` 筛选参数，而非为每种实体类型单独建接口。

**理由**：4 种实体类型的已删除记录结构相似，统一接口简化前端调用。返回通用字段（id、entity_type、name、code、delete_time）足以满足列表展示需求。

### 3. 恢复接口路径：`POST /api/masterdata/deleted-items/{entity_type}/{id}/restore`

将 entity_type 放在路径中而非请求体中，与现有审批接口 `POST /api/masterdata/approval/{entity_type}/{id}/review` 风格保持一致。

### 4. 前端页面：复用审核中心统计卡片样式

直接复用审核中心的 `statCards` + `stat-card` 样式模式，保持 UI 一致性。

## Risks / Trade-offs

- **[数据一致性]** 恢复后记录的 `submission_status` 可能仍为待删除状态 → 恢复时同时将 `submission_status` 重置为 `2`（已批准）
- **[性能]** 统一查询已删除记录需 UNION 多表 → 数据量不大时可接受，后续可通过缓存优化
