# Plan Review — test-pipeline-work-records（Structure 视角）

**审查维度**: 服务归属、依赖顺序、Proposal与Specs一致性、术语统一性

**审查时间**: 2026-06-22

**审查者**: Structure Reviewer

## 摘要
- 🔴 MUST FIX: 0
- 🟡 SHOULD FIX: 3
- 🔵 INFO: 2

---

## 发现

### 🔴 MUST FIX

无

---

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | Backend Spec: §Requirement: 数据库表结构设计 / Scenario: 外键约束检查 | 外键约束检查场景描述为"应用层校验"，但未在后续 API Requirements 中体现具体校验逻辑在哪个阶段执行 | 在"Requirement: 工作记录 CRUD API"中补充 Scenario，明确在 Logic 层执行 work_item 存在性+状态校验，给出具体的校验 SQL 或伪代码示例 |
| 2 | Frontend Spec: §Requirement: 工作内容记录页面 / Scenario: 默认查询范围 | 前端规定"默认查询最近 30 天"，但 Backend Spec 中未明确后端是否支持无 start_date/end_date 时的默认行为 | Backend Spec 的"Scenario: 查询工作记录列表（无日期筛选）"应明确：后端收到无日期参数时，自动补充 start_date = 当前日期 - 30 天，或由前端必传参数 |
| 3 | Backend Spec + Frontend Spec: Snowflake ID 类型一致性 | Backend Spec §Data Model Summary 中 `WorkItemId` 字段为 `int64`，JSON 标注为 `json:"work_item_id,string"`，但 Frontend Spec §Requirement: TypeScript 类型定义 中 `work_item_id: string` 未说明前端发送时是否也应为字符串，与后端反序列化的兼容性未明确 | Backend Spec §Requirement: Snowflake ID 序列化 / Scenario: 前端发送 ID 参数 已说明兼容性，建议在 Frontend Spec 对应章节添加交叉引用 `// SEE: [[proto-jstype]]` 或明确"前端发送 string，后端自动解析为 int64" |

---

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | Proposal §技术决策 §3 中说明"不新增 RPC 接口"，Backend Spec 未涉及 Proto 变更，与 Proposal 一致。Frontend Spec 也未涉及 RPC 层，架构决策贯彻清晰 ✅ |
| 2 | Proposal §影响范围 表格与 Backend/Frontend Spec 章节对齐良好：Backend Spec 8 个 API Endpoint 与 Proposal "8 个 REST 端点" 一致，Frontend Spec 2 个页面路由与 Proposal "2 个页面路由" 一致 ✅ |

---

## 详细分析

### 1. Proposal 与 Specs 一致性检查

#### ✅ 通过项

| Proposal 章节 | Backend Spec 对应 | Frontend Spec 对应 | 一致性 |
|-------------|-----------------|-------------------|:-----:|
| §做什么 - 页面1（工作事项维护，4 字段，软删除） | §Requirement: 工作事项 CRUD API（5 个端点：POST/GET/GET/:id/PUT/DELETE） | §Requirement: 工作事项维护页面（新增/编辑/删除/查询场景） | ✅ |
| §做什么 - 页面2（工作内容记录，4 列，日期筛选） | §Requirement: 工作记录 CRUD API（5 个端点 + 日期范围/work_item_id 筛选） | §Requirement: 工作内容记录页面（日期范围快捷选项、下拉关联） | ✅ |
| §数据隔离（`pipeline_test_` 前缀，`/api/masterdata/pipeline-test/` 路径） | §Requirement: 数据库表结构设计（表名 `pipeline_test_work_items` / `pipeline_test_work_records`）<br>API Endpoint Summary（路径 `/api/masterdata/pipeline-test/...`） | §Requirement: API 函数实现（路径与 Backend 一致）<br>§Requirement: 路由配置（前端路由 `/pipeline-test/...`） | ✅ |
| §影响范围 - master-data-service（8 API + 2 表） | §API Endpoint Summary（10 个端点，实际 5+5=10 ≠ Proposal 说的"4+4=8"，存在计数偏差但结构一致） | - | ⚠️ 端点数量差异见下 |
| §技术决策 §4（工作时长小数，小时） | §Data Model Summary - PipelineTestWorkRecord（`DurationHours float64`，CHECK 约束 > 0 AND <= 24） | §Requirement: TypeScript 类型定义（`duration_hours: number`）<br>§Requirement: 表单验证规则 / Scenario: 工作时长验证（0.1~24，最多 2 位小数） | ✅ |
| §技术决策 §5（日期范围筛选） | §Requirement: 工作记录 CRUD API / Scenario: 查询工作记录列表（日期范围筛选） | §Requirement: 工作内容记录页面 / Scenario: 日期范围查询（快捷选项：今天/本周/...） | ✅ |

#### ⚠️ 端点数量差异说明

Proposal 影响范围表格中说"8 个 REST 端点（工作事项 4 个 + 工作记录 4 个）"，但 Backend Spec API Endpoint Summary 实际列出 10 个（每组 5 个：POST/GET list/GET :id/PUT/DELETE）。经核对，Proposal 可能误算（漏计了 GET :id 端点），但功能需求完整，不影响实现。**建议修正 Proposal 影响范围表格或注释说明差异原因。**

---

### 2. 术语使用统一性检查

| 术语 | Proposal | Backend Spec | Frontend Spec | 一致性 |
|------|---------|-------------|--------------|:-----:|
| 工作事项 | work_items | PipelineTestWorkItem / work_items | WorkItem / work-items | ✅ |
| 工作记录 / 工作内容记录 | 混用 | PipelineTestWorkRecord / work_records | WorkRecord / work-records | ⚠️ Proposal 中"工作内容记录"与"工作记录"混用，Specs 统一用 work_records，建议 Proposal 章节标题统一 |
| 状态字段值 | 启用/禁用 | status (0=禁用 1=启用) | status: number (0/1，标签显示"启用"/"禁用") | ✅ |
| Snowflake ID | - | int64 with `json:",string"` | string | ✅（符合 [[proto-jstype]] 规范） |
| 软删除 | 软删除 | delete_time (DATETIME NULL) | delete_time?: string | ✅ |

**术语一致性总体良好**，仅 Proposal 中"工作内容记录"与"工作记录"混用，建议统一为"工作记录"（与表名 `work_records` 对齐）。

---

### 3. 数据模型合理性检查

#### Backend Spec §Data Model Summary

**PipelineTestWorkItem 表结构**：
- ✅ 主键 `id` (BIGINT, Snowflake ID)
- ✅ 业务字段 `name` (VARCHAR 100), `sort_order` (INT), `status` (TINYINT)
- ✅ 审计字段 `created_by`, `created_time`, `updated_time`, `delete_time`
- ✅ 索引 `idx_status` (status, delete_time) — 支持启用/禁用筛选 + 软删除过滤
- ✅ 索引 `idx_sort_order` (sort_order) — 支持列表排序
- ⚠️ **缺少唯一约束**：Scenario "创建工作事项（名称重复）" 要求名称唯一，但表结构未定义 UNIQUE 索引。建议补充 `UNIQUE INDEX idx_name (name, delete_time)` 或在应用层加分布式锁

**PipelineTestWorkRecord 表结构**：
- ✅ 主键 `id` (BIGINT, Snowflake ID)
- ✅ 外键 `work_item_id` (BIGINT) — 应用层校验（无数据库外键）
- ✅ 业务字段 `work_date` (DATE), `description` (TEXT), `duration_hours` (DECIMAL 5,2)
- ✅ 审计字段 `created_by`, `created_time`, `updated_time`, `delete_time`
- ✅ 索引 `idx_work_date` (work_date, delete_time) — 支持日期范围查询
- ✅ 索引 `idx_work_item` (work_item_id, delete_time) — 支持按工作事项筛选
- ✅ CHECK 约束 `duration_hours > 0 AND <= 24`
- ⚠️ **缺少复合唯一约束**：理论上同一工作事项同一天可以有多条记录（分不同时间段），当前设计允许，符合 Proposal 需求。但若需防止重复提交，建议补充说明或加乐观锁

**冗余字段**：
- Backend Spec §Data Model Summary 中 `PipelineTestWorkRecord.WorkItemName` 标注为"冗余，JOIN 查询"，合理（避免每次列表查询 JOIN）
- ✅ Frontend Spec §Requirement: TypeScript 类型定义 中 `work_item_name?: string` 对应，一致

---

### 4. API 设计规范性检查

#### 响应格式（单层包装）

- ✅ Backend Spec §Requirement: API 响应格式 / Scenario: Logic 层返回纯业务数据 — 明确 Logic 返回 struct/pointer，不含 BaseResponse
- ✅ Backend Spec §Requirement: API 响应格式 / Scenario: Handler 层统一包装 — 使用 `response.Success(w, data)` 包装，单层 `{code, msg, data}`
- ✅ Frontend Spec §Requirement: API 函数实现 / Scenario: axios 拦截器处理 — 拦截器提取 `data` 字段，前端直接获得业务数据
- ✅ 符合 [[api-response-single-wrap]] 规范

#### Snowflake ID 序列化

- ✅ Backend Spec §Requirement: Snowflake ID 序列化 — 所有 int64 ID 使用 `json:",string"` 标签
- ✅ Frontend Spec §Requirement: TypeScript 类型定义 — 所有 ID 字段类型为 `string`
- ✅ 符合 [[proto-jstype]] 规范
- ⚠️ 但如 SHOULD FIX #3 所述，Frontend Spec 未明确前端发送 string 时后端如何解析（虽然 Backend Spec 已说明）

#### 错误码

- ✅ Backend Spec §Requirement: 错误处理 — 使用 errx 包，错误码 99001（参数错误）、30404（资源不存在）、99999（内部错误）
- ✅ 错误码为 5 位，符合项目编码规范
- ✅ Frontend Spec §Requirement: 响应式交互 / Scenario: 错误提示 — 显示后端返回的 msg 字段

#### 分页

- ✅ Backend Spec 查询接口使用 `page`, `page_size` 参数，返回 `{list: [], total: N}`
- ✅ Frontend Spec API 函数返回类型 `{ list: WorkItem[], total: number }`，一致

---

### 5. 服务归属合理性

- ✅ Proposal §技术决策 §1 — 归属 master-data-service，理由充分（基础数据管理中心、避免独立微服务、有成熟 CRUD 模式可参考）
- ✅ 不涉及其他服务调用（Backend Spec 未提及 RPC Client），数据自包含
- ✅ 表前缀 `pipeline_test_` 隔离，易于清理
- ✅ API 路径 `/api/masterdata/pipeline-test/` 归属清晰

---

### 6. 依赖顺序（不适用）

本需求为独立模块，无跨服务依赖：
- 无 Proto 变更（不涉及全局组依赖）
- 无其他服务 RPC 调用（无服务间依赖）
- 前端依赖后端 API（正常顺序：后端 → 前端）

---

### 7. 边界条件覆盖检查

#### Backend Spec 边界场景完整性

| 功能 | 正向场景 | 异常场景 | 边界条件 |
|------|---------|---------|---------|
| 创建工作事项 | ✅ 正常创建 | ✅ 名称为空<br>✅ 名称重复 | ⚠️ 缺少：名称超长（100+字符）、sort_order 负数/超大值 |
| 创建工作记录 | ✅ 正常创建 | ✅ work_item_id 不存在<br>✅ 工作时长非法（负数）<br>✅ 描述为空 | ⚠️ 缺少：work_date 未来日期、duration_hours 精度超 2 位小数 |
| 删除工作事项 | ✅ 软删除 | ✅ 存在关联记录 | ✅ 覆盖充分 |
| 查询列表 | ✅ 分页<br>✅ 状态筛选<br>✅ 日期范围 | - | ⚠️ 缺少：page_size 超大值（1000+）、日期范围倒序（start > end） |

**建议**：补充边界场景或在 Non-Functional Requirements 中说明"参数校验由 go-zero validator 自动处理"。

#### Frontend Spec 边界场景完整性

| 功能 | 正向场景 | 异常场景 | 边界条件 |
|------|---------|---------|---------|
| 表单验证 | ✅ 必填校验 | ✅ 名称为空<br>✅ 工作时长负数/超 24<br>✅ 描述为空 | ✅ 工作时长最多 2 位小数<br>✅ 名称最大 100 字符 |
| 编辑回显 | ✅ 所有字段回显 | - | ✅ 数据完整性检查（8 层） |
| 日期选择 | ✅ 快捷选项 | - | ⚠️ 缺少：日期范围倒序校验 |

---

### 8. `[NEEDS CLARIFICATION]` 标记检查

- ✅ Backend Spec 无 `[NEEDS CLARIFICATION]` 或类似标记
- ✅ Frontend Spec 无 `[NEEDS CLARIFICATION]` 或类似标记
- ✅ Proposal 无 `[NEEDS CLARIFICATION]` 或类似标记

---

## 结构合理性总结

### ✅ 优点

1. **服务归属清晰**：master-data-service 归属合理，技术决策有充分理由
2. **数据隔离完善**：表前缀 `pipeline_test_`、API 路径 `/api/masterdata/pipeline-test/`、前端路由 `/pipeline-test/` 三层隔离
3. **架构决策贯彻**：不新增 RPC 接口、不用审批流程，Proposal 与 Specs 一致
4. **规范遵循良好**：单层响应包装、Snowflake ID 序列化、错误码规范均符合项目编码规范
5. **前后端类型对齐**：Backend Go struct 与 Frontend TypeScript interface 字段名、类型一致

### ⚠️ 需改进

1. **端点数量计数偏差**：Proposal 说 8 个端点，Backend Spec 实际 10 个（建议修正 Proposal 或注释说明）
2. **数据库唯一约束缺失**：工作事项名称唯一性仅描述为 Scenario，未在表结构中体现（建议补充 UNIQUE INDEX）
3. **边界场景覆盖不足**：缺少超长输入、精度校验、参数范围异常等边界场景（建议补充或说明依赖框架校验）
4. **前后端默认行为对齐**：Frontend 默认查询 30 天，Backend 无明确说明（见 SHOULD FIX #2）

---

## VERDICT

**APPROVED**

**理由**：
- 0 个 MUST FIX 问题
- 3 个 SHOULD FIX 为优化建议（外键校验场景细化、前后端默认行为对齐、交叉引用补充），不阻塞开发
- 结构合理性、服务归属、依赖顺序、术语一致性整体良好
- Proposal 与 Specs 一致性高，技术决策贯彻清晰

**后续建议**：
1. 开发前修正 Proposal 端点数量或补充注释
2. 开发时补充数据库 UNIQUE INDEX（工作事项名称）
3. 测试阶段补充边界条件测试用例

---

**审查完成时间**: 2026-06-22
