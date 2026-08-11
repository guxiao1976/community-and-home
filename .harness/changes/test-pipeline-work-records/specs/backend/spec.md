# Backend Specification — 测试流水线工作记录模块

## Purpose

为测试流水线验证提供工作记录管理功能的后端实现。该模块归属 master-data-service，提供工作事项和工作内容记录的 CRUD REST API，支持日期范围查询和关联查询。数据使用 `pipeline_test_` 前缀隔离，不涉及其他服务调用。

## Requirements

---

### Requirement: 数据库表结构设计

The system SHALL create two database tables with `pipeline_test_` prefix for data isolation.

#### Scenario: 工作事项表结构
- **GIVEN** masterdata_db 数据库
- **WHEN** 执行 migration 脚本
- **THEN** 创建表 `pipeline_test_work_items`，包含字段：
  - `id` BIGINT PRIMARY KEY — Snowflake ID
  - `name` VARCHAR(100) NOT NULL — 工作事项名称
  - `sort_order` INT DEFAULT 0 — 排序号（升序）
  - `status` TINYINT DEFAULT 1 — 状态（0=禁用 1=启用）
  - `created_by` BIGINT — 创建人 ID
  - `created_time` DATETIME — 创建时间
  - `updated_time` DATETIME — 更新时间
  - `delete_time` DATETIME NULL — 软删除时间
  - INDEX `idx_status` (status, delete_time)
  - INDEX `idx_sort_order` (sort_order)

#### Scenario: 工作记录表结构
- **GIVEN** masterdata_db 数据库
- **WHEN** 执行 migration 脚本
- **THEN** 创建表 `pipeline_test_work_records`，包含字段：
  - `id` BIGINT PRIMARY KEY — Snowflake ID
  - `work_item_id` BIGINT NOT NULL — 关联工作事项 ID
  - `work_date` DATE NOT NULL — 工作日期
  - `description` TEXT NOT NULL — 工作内容描述
  - `duration_hours` DECIMAL(5,2) NOT NULL — 工作时长（小时，如 0.5）
  - `created_by` BIGINT — 创建人 ID
  - `created_time` DATETIME — 创建时间
  - `updated_time` DATETIME — 更新时间
  - `delete_time` DATETIME NULL — 软删除时间
  - INDEX `idx_work_date` (work_date, delete_time)
  - INDEX `idx_work_item` (work_item_id, delete_time)
  - CONSTRAINT `CHECK (duration_hours > 0 AND duration_hours <= 24)`

#### Scenario: 外键约束检查
- **GIVEN** 两张表已创建
- **WHEN** 开发者尝试在 `pipeline_test_work_records` 插入不存在的 `work_item_id`
- **THEN** 应用层校验失败（不使用数据库外键，避免软删除冲突），返回错误 "工作事项不存在或已禁用"

---

### Requirement: 工作事项 CRUD API

The system SHALL provide REST API endpoints for work items management under `/api/masterdata/pipeline-test/work-items`.

#### Scenario: 创建工作事项（正常）
- **GIVEN** 用户已登录
- **WHEN** POST `/api/masterdata/pipeline-test/work-items` with body:
  ```json
  {
    "name": "需求分析",
    "sort_order": 10,
    "status": 1
  }
  ```
- **THEN** 返回 200，响应体：
  ```json
  {
    "code": 0,
    "msg": "success",
    "data": {
      "id": "1234567890123456789",  // Snowflake ID as string
      "name": "需求分析",
      "sort_order": 10,
      "status": 1,
      "created_time": "2026-06-22T21:00:00+08:00",
      "updated_time": "2026-06-22T21:00:00+08:00"
    }
  }
  ```
- **AND** 数据库插入一条记录，`created_by` 从 JWT 提取

#### Scenario: 创建工作事项（名称为空）
- **GIVEN** 用户已登录
- **WHEN** POST `/api/masterdata/pipeline-test/work-items` with body `{"name": "", "status": 1}`
- **THEN** 返回 400，响应体：
  ```json
  {
    "code": 99001,
    "msg": "参数错误: 工作事项名称不能为空",
    "data": null
  }
  ```

#### Scenario: 创建工作事项（名称重复）
- **GIVEN** 已存在名称为 "需求分析" 的工作事项（未软删除）
- **WHEN** POST `/api/masterdata/pipeline-test/work-items` with body `{"name": "需求分析", "status": 1}`
- **THEN** 返回 400，响应体：
  ```json
  {
    "code": 99001,
    "msg": "参数错误: 工作事项名称已存在",
    "data": null
  }
  ```

#### Scenario: 查询工作事项列表
- **GIVEN** 数据库中有 3 条工作事项（1 条已软删除）
- **WHEN** GET `/api/masterdata/pipeline-test/work-items?page=1&page_size=10`
- **THEN** 返回 200，响应体包含 2 条记录（不含软删除），按 `sort_order` ASC 排序

#### Scenario: 查询工作事项列表（包含禁用项）
- **GIVEN** 数据库中有 2 条启用、1 条禁用工作事项
- **WHEN** GET `/api/masterdata/pipeline-test/work-items?page=1&page_size=10&include_disabled=true`
- **THEN** 返回 200，响应体包含 3 条记录

#### Scenario: 查询工作事项列表（仅启用项用于下拉列表）
- **GIVEN** 数据库中有 2 条启用、1 条禁用工作事项
- **WHEN** GET `/api/masterdata/pipeline-test/work-items?status=1`
- **THEN** 返回 200，响应体仅包含 2 条启用记录

#### Scenario: 查询单个工作事项
- **GIVEN** 存在 ID 为 123 的工作事项
- **WHEN** GET `/api/masterdata/pipeline-test/work-items/123`
- **THEN** 返回 200，响应体包含该工作事项完整信息

#### Scenario: 查询不存在的工作事项
- **GIVEN** 不存在 ID 为 999 的工作事项
- **WHEN** GET `/api/masterdata/pipeline-test/work-items/999`
- **THEN** 返回 404，响应体：
  ```json
  {
    "code": 30404,
    "msg": "工作事项不存在",
    "data": null
  }
  ```

#### Scenario: 更新工作事项
- **GIVEN** 存在 ID 为 123 的工作事项
- **WHEN** PUT `/api/masterdata/pipeline-test/work-items/123` with body:
  ```json
  {
    "name": "需求分析（更新）",
    "sort_order": 20,
    "status": 1
  }
  ```
- **THEN** 返回 200，响应体包含更新后的数据，`updated_time` 已更新

#### Scenario: 删除工作事项（软删除）
- **GIVEN** 存在 ID 为 123 的工作事项
- **WHEN** DELETE `/api/masterdata/pipeline-test/work-items/123`
- **THEN** 返回 200，数据库 `delete_time` 字段更新为当前时间
- **AND** 后续查询列表不再返回该记录

#### Scenario: 删除工作事项（存在关联记录）
- **GIVEN** 工作事项 ID 123 有 5 条未删除的工作记录
- **WHEN** DELETE `/api/masterdata/pipeline-test/work-items/123`
- **THEN** 返回 400，响应体：
  ```json
  {
    "code": 99001,
    "msg": "该工作事项存在关联记录，无法删除",
    "data": null
  }
  ```

---

### Requirement: 工作记录 CRUD API

The system SHALL provide REST API endpoints for work records management under `/api/masterdata/pipeline-test/work-records`.

#### Scenario: 创建工作记录（正常）
- **GIVEN** 存在启用的工作事项 ID 123
- **WHEN** POST `/api/masterdata/pipeline-test/work-records` with body:
  ```json
  {
    "work_item_id": "123",
    "work_date": "2026-06-22",
    "description": "完成用户模块需求分析，产出 proposal 和 spec",
    "duration_hours": 2.5
  }
  ```
- **THEN** 返回 200，响应体：
  ```json
  {
    "code": 0,
    "msg": "success",
    "data": {
      "id": "1234567890123456790",
      "work_item_id": "123",
      "work_item_name": "需求分析",  // 冗余字段，便于列表展示
      "work_date": "2026-06-22",
      "description": "完成用户模块需求分析，产出 proposal 和 spec",
      "duration_hours": 2.5,
      "created_time": "2026-06-22T21:00:00+08:00",
      "updated_time": "2026-06-22T21:00:00+08:00"
    }
  }
  ```

#### Scenario: 创建工作记录（工作事项不存在）
- **GIVEN** 不存在 ID 为 999 的工作事项
- **WHEN** POST `/api/masterdata/pipeline-test/work-records` with body `{"work_item_id": "999", "work_date": "2026-06-22", "description": "test", "duration_hours": 1.0}`
- **THEN** 返回 400，响应体：
  ```json
  {
    "code": 99001,
    "msg": "工作事项不存在或已禁用",
    "data": null
  }
  ```

#### Scenario: 创建工作记录（工作时长非法）
- **GIVEN** 存在启用的工作事项 ID 123
- **WHEN** POST `/api/masterdata/pipeline-test/work-records` with body `{"work_item_id": "123", "work_date": "2026-06-22", "description": "test", "duration_hours": -1.0}`
- **THEN** 返回 400，响应体：
  ```json
  {
    "code": 99001,
    "msg": "参数错误: 工作时长必须大于 0 且不超过 24 小时",
    "data": null
  }
  ```

#### Scenario: 创建工作记录（描述为空）
- **GIVEN** 存在启用的工作事项 ID 123
- **WHEN** POST `/api/masterdata/pipeline-test/work-records` with body `{"work_item_id": "123", "work_date": "2026-06-22", "description": "", "duration_hours": 1.0}`
- **THEN** 返回 400，响应体：
  ```json
  {
    "code": 99001,
    "msg": "参数错误: 工作内容描述不能为空",
    "data": null
  }
  ```

#### Scenario: 查询工作记录列表（日期范围筛选）
- **GIVEN** 数据库中有 2026-06-20 ~ 2026-06-25 的工作记录各 2 条
- **WHEN** GET `/api/masterdata/pipeline-test/work-records?start_date=2026-06-22&end_date=2026-06-23&page=1&page_size=10`
- **THEN** 返回 200，响应体包含 4 条记录（6-22 和 6-23 的），按 `work_date` DESC, `created_time` DESC 排序

#### Scenario: 查询工作记录列表（无日期筛选）
- **GIVEN** 数据库中有 10 条工作记录
- **WHEN** GET `/api/masterdata/pipeline-test/work-records?page=1&page_size=10`
- **THEN** 返回 200，响应体包含 10 条记录（默认返回最近 30 天）

#### Scenario: 查询工作记录列表（按工作事项筛选）
- **GIVEN** 工作事项 ID 123 有 5 条记录，ID 456 有 3 条记录
- **WHEN** GET `/api/masterdata/pipeline-test/work-records?work_item_id=123&page=1&page_size=10`
- **THEN** 返回 200，响应体仅包含 5 条工作事项 123 的记录

#### Scenario: 查询单个工作记录
- **GIVEN** 存在 ID 为 789 的工作记录
- **WHEN** GET `/api/masterdata/pipeline-test/work-records/789`
- **THEN** 返回 200，响应体包含该记录完整信息及关联的 `work_item_name`

#### Scenario: 更新工作记录
- **GIVEN** 存在 ID 为 789 的工作记录
- **WHEN** PUT `/api/masterdata/pipeline-test/work-records/789` with body:
  ```json
  {
    "work_item_id": "123",
    "work_date": "2026-06-22",
    "description": "更新后的描述",
    "duration_hours": 3.0
  }
  ```
- **THEN** 返回 200，响应体包含更新后的数据

#### Scenario: 删除工作记录
- **GIVEN** 存在 ID 为 789 的工作记录
- **WHEN** DELETE `/api/masterdata/pipeline-test/work-records/789`
- **THEN** 返回 200，数据库 `delete_time` 字段更新为当前时间

---

### Requirement: API 响应格式

The system SHALL follow single-layer response wrapping pattern as defined in [[api-response-single-wrap]].

#### Scenario: Logic 层返回纯业务数据
- **GIVEN** 任何 Logic 函数
- **WHEN** 执行成功
- **THEN** 返回纯业务数据结构（struct/pointer），不包含 `BaseResponse`

#### Scenario: Handler 层统一包装
- **GIVEN** Logic 返回数据或错误
- **WHEN** Handler 调用 `response.Success(w, data)` 或 `response.Error(w, err)`
- **THEN** 响应体为单层包装：
  ```json
  {
    "code": 0,
    "msg": "success",
    "data": <业务数据>
  }
  ```
- **AND** 不出现双层嵌套 `{code:0, data: {code:0, data:...}}`

---

### Requirement: Snowflake ID 序列化

The system SHALL serialize all int64 ID fields as JSON strings as defined in [[proto-jstype]].

#### Scenario: API 返回 ID 字段
- **GIVEN** 任何包含 `id`, `work_item_id`, `created_by` 等 int64 字段的响应
- **WHEN** 序列化为 JSON
- **THEN** 所有 int64 ID 字段使用 `json:",string"` 标签
- **AND** JSON 输出为字符串格式：`"id": "1234567890123456789"`

#### Scenario: 前端发送 ID 参数
- **GIVEN** 前端发送 `{"work_item_id": "123"}` (string)
- **WHEN** 后端反序列化
- **THEN** 正确解析为 int64 类型 123
- **AND** 不出现精度丢失

---

### Requirement: 错误处理

The system SHALL use community-common/v2 errx package for error code management.

#### Scenario: 参数校验错误
- **GIVEN** 客户端发送非法参数（必填项为空、类型错误、范围越界）
- **WHEN** 后端校验
- **THEN** 返回 HTTP 400，错误码 99001，msg 包含具体字段和原因

#### Scenario: 资源不存在
- **GIVEN** 客户端请求不存在的资源 ID
- **WHEN** 后端查询数据库
- **THEN** 返回 HTTP 404，错误码 30404，msg 为 "xxx不存在"

#### Scenario: 业务规则冲突
- **GIVEN** 客户端操作违反业务规则（如删除有关联记录的工作事项）
- **WHEN** 后端校验
- **THEN** 返回 HTTP 400，错误码 99001，msg 包含冲突原因

#### Scenario: 服务器内部错误
- **GIVEN** 数据库连接失败或其他内部错误
- **WHEN** 后端捕获异常
- **THEN** 返回 HTTP 500，错误码 99999，msg 为通用错误信息（不暴露内部细节）
- **AND** 错误详情记录到日志

---

### Requirement: 数据库事务

The system SHALL use database transactions for data consistency.

#### Scenario: 创建工作记录
- **GIVEN** 需要插入 work_record 并更新统计信息
- **WHEN** 执行创建操作
- **THEN** 两个操作在同一事务中
- **AND** 任一步骤失败则全部回滚

#### Scenario: 删除工作事项
- **GIVEN** 需要检查关联记录并软删除
- **WHEN** 执行删除操作
- **THEN** 检查和删除在同一事务中（避免并发问题）

---

### Requirement: 日志记录

The system SHALL log key operations for pipeline evaluation.

#### Scenario: API 请求日志
- **GIVEN** 任何 API 请求
- **WHEN** 请求到达
- **THEN** 记录日志：method, path, user_id, request_id, 开始时间

#### Scenario: API 响应日志
- **GIVEN** API 请求处理完成
- **WHEN** 返回响应
- **THEN** 记录日志：status_code, error_code (如果失败), 耗时

#### Scenario: 业务操作日志
- **GIVEN** CRUD 操作执行
- **WHEN** 操作成功或失败
- **THEN** 记录日志：操作类型, 资源 ID, user_id, 结果

---

## API Endpoint Summary

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/masterdata/pipeline-test/work-items` | 创建工作事项 |
| GET | `/api/masterdata/pipeline-test/work-items` | 查询工作事项列表（支持分页、状态筛选） |
| GET | `/api/masterdata/pipeline-test/work-items/:id` | 查询单个工作事项 |
| PUT | `/api/masterdata/pipeline-test/work-items/:id` | 更新工作事项 |
| DELETE | `/api/masterdata/pipeline-test/work-items/:id` | 删除工作事项（软删除） |
| POST | `/api/masterdata/pipeline-test/work-records` | 创建工作记录 |
| GET | `/api/masterdata/pipeline-test/work-records` | 查询工作记录列表（支持分页、日期范围、工作事项筛选） |
| GET | `/api/masterdata/pipeline-test/work-records/:id` | 查询单个工作记录 |
| PUT | `/api/masterdata/pipeline-test/work-records/:id` | 更新工作记录 |
| DELETE | `/api/masterdata/pipeline-test/work-records/:id` | 删除工作记录（软删除） |

## Data Model Summary

### PipelineTestWorkItem
```go
type PipelineTestWorkItem struct {
    Id          int64      `json:"id,string"`
    Name        string     `json:"name"`
    SortOrder   int        `json:"sort_order"`
    Status      int        `json:"status"`  // 0=禁用 1=启用
    CreatedBy   int64      `json:"created_by,string"`
    CreatedTime time.Time  `json:"created_time"`
    UpdatedTime time.Time  `json:"updated_time"`
    DeleteTime  *time.Time `json:"delete_time,omitempty"`
}
```

### PipelineTestWorkRecord
```go
type PipelineTestWorkRecord struct {
    Id            int64      `json:"id,string"`
    WorkItemId    int64      `json:"work_item_id,string"`
    WorkItemName  string     `json:"work_item_name"`  // 冗余，JOIN 查询
    WorkDate      string     `json:"work_date"`       // YYYY-MM-DD
    Description   string     `json:"description"`
    DurationHours float64    `json:"duration_hours"`
    CreatedBy     int64      `json:"created_by,string"`
    CreatedTime   time.Time  `json:"created_time"`
    UpdatedTime   time.Time  `json:"updated_time"`
    DeleteTime    *time.Time `json:"delete_time,omitempty"`
}
```

## Non-Functional Requirements

### Performance
- The system SHALL respond to GET requests within 200ms for datasets < 1000 records
- The system SHALL support pagination with page_size up to 100

### Security
- The system SHALL extract user_id from JWT token for created_by field
- The system SHALL validate all input parameters before database operations

### Data Integrity
- The system SHALL use soft delete (delete_time) instead of physical deletion
- The system SHALL prevent deletion of work items with active records

### Testability
- The system SHALL use `pipeline_test_` prefix for easy cleanup
- The system SHALL log all operations with request_id for tracing
