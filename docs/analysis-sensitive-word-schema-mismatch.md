# 敏感词表结构不一致分析报告

## 问题概述

MySQL 数据库中 `md_sensitive_word` 表的实际结构与 Go Model 定义存在不一致。

## 详细对比

### 1. MySQL Schema (masterdata_schema.sql)

```sql
CREATE TABLE IF NOT EXISTS md_sensitive_word (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'Word ID',
    word VARCHAR(100) NOT NULL COMMENT 'Sensitive word',
    category VARCHAR(50) NOT NULL COMMENT 'Category (e.g., political, violence)',
    severity TINYINT NOT NULL COMMENT '1=Low, 2=Medium, 3=High',
    action TINYINT NOT NULL COMMENT '1=Warn, 2=Block, 3=Review',
    status TINYINT NOT NULL DEFAULT 1 COMMENT '1=Active, 2=Inactive',
    created_by BIGINT NOT NULL COMMENT 'Creator user ID',
    created_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Creation timestamp',
    updated_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update timestamp',
    UNIQUE KEY uk_word (word),
    KEY idx_category (category),
    KEY idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Sensitive word list';
```

**字段数量**: 9 个字段

### 2. Go Model (mdSensitiveWordModel_gen.go)

```go
type MdSensitiveWord struct {
    Id               int64          `db:"id"`
    Word             string         `db:"word"`
    Category         string         `db:"category"`
    Severity         int64          `db:"severity"`
    Action           int64          `db:"action"`
    Status           int64          `db:"status"`
    SubmissionStatus int64          `db:"submission_status"`
    SubmissionType   sql.NullInt64  `db:"submission_type"`
    ChangeSnapshot   sql.NullString `db:"change_snapshot"`
    SubmitterId      sql.NullInt64  `db:"submitter_id"`
    SubmitTime       sql.NullTime   `db:"submit_time"`
    ReviewerId       sql.NullInt64  `db:"reviewer_id"`
    ReviewTime       sql.NullTime   `db:"review_time"`
    ReviewNotes      sql.NullString `db:"review_notes"`
    CreatedBy        int64          `db:"created_by"`
    CreatedTime      time.Time      `db:"created_time"`
    UpdatedTime      time.Time      `db:"updated_time"`
    DeleteTime       sql.NullTime   `db:"delete_time"`
}
```

**字段数量**: 18 个字段

### 3. 缺失字段清单

Model 中存在但 MySQL Schema 中缺失的字段：

1. `submission_status` - 提交状态 (0=待提交, 1=待审核, 2=已通过, 3=已拒绝, 4=待删除)
2. `submission_type` - 提交类型 (1=新增, 2=修改, 3=删除)
3. `change_snapshot` - 修改前快照 (JSON)
4. `submitter_id` - 提交人 ID
5. `submit_time` - 提交时间
6. `reviewer_id` - 审核人 ID
7. `review_time` - 审核时间
8. `review_notes` - 审核备注
9. `delete_time` - 软删除时间

## 业务逻辑分析

### 当前业务依赖的审批工作流

通过分析业务逻辑代码，发现敏感词模块已经实现了完整的审批工作流：

#### 1. 创建流程 (createSensitiveWordLogic.go)
- 创建时设置 `submission_status = 0` (待提交)
- 设置 `submission_type = 1` (新增)

#### 2. 更新流程 (updateSensitiveWordLogic.go)
- 更新前保存快照到 `change_snapshot`
- 设置 `submission_type = 2` (修改)
- 重置 `submission_status = 0` (待提交)

#### 3. 删除流程 (deleteSensitiveWordLogic.go)
- 设置 `submission_status = 4` (待删除)
- 设置 `submission_type = 3` (删除)

#### 4. 提交流程 (submitSensitiveWordLogic.go)
- 设置 `submission_status = 1` (待审核)
- 记录 `submitter_id` 和 `submit_time`
- 创建 `md_submission_record` 记录

#### 5. 审核流程 (reviewItemLogic.go)
- 审核通过: `submission_status = 2`，删除操作时设置 `delete_time`
- 审核拒绝: `submission_status = 3`，从 `change_snapshot` 恢复数据
- 记录 `reviewer_id`, `review_time`, `review_notes`

#### 6. 查询过滤
- `FindWithFilters`: 过滤 `submission_status != 4` 和 `delete_time is null`
- `FindPendingBySubmissionStatus`: 按审批状态查询
- `FindDeleted`: 查询软删除记录 (`delete_time is not null`)
- `Restore`: 恢复删除 (清除 `delete_time`)

### 与其他实体的一致性

查看其他主数据实体的表结构：

- `md_administrative_division` - **包含完整审批字段**
- `md_residential_area` - **包含完整审批字段**
- `md_configuration` - **包含完整审批字段**

所有主数据实体都遵循统一的审批工作流模式。

## 建议

### **推荐方案：修改 MySQL Schema，添加缺失字段**

#### 理由：

1. **业务逻辑已实现**: 代码中已经完整实现了审批工作流，包括创建、修改、删除、提交、审核、恢复等功能
2. **架构一致性**: 其他主数据实体 (administrative_division, residential_area, configuration) 都包含这些审批字段
3. **功能完整性**: 
   - 支持提交-审核工作流
   - 支持修改回滚 (change_snapshot)
   - 支持软删除和恢复
   - 支持审批记录追溯
4. **前端已适配**: API 定义中 `SensitiveWord` 类型已包含 `submission_status` 和 `submission_type` 字段
5. **最小改动**: 只需执行 ALTER TABLE 添加字段，无需修改业务逻辑代码

#### 风险：

- **低风险**: 添加字段为向后兼容操作，不影响现有数据
- 需要执行数据库迁移脚本

### 不推荐方案：简化 Model 去除审批字段

#### 理由：

1. **破坏已有功能**: 需要删除大量已实现的业务逻辑代码
2. **架构不一致**: 与其他主数据实体的设计模式不一致
3. **功能降级**: 失去审批、回滚、软删除等重要功能
4. **工作量大**: 需要修改多个文件的业务逻辑和 API 定义

## 实施方案

### 数据库迁移脚本

```sql
-- Migration: Add approval workflow fields to md_sensitive_word
-- Execute: mysql -u root -p masterdata_db < migration_add_sensitive_word_approval_fields.sql

USE masterdata_db;

-- Add approval workflow fields
ALTER TABLE md_sensitive_word
ADD COLUMN submission_status TINYINT NOT NULL DEFAULT 0 COMMENT '0=Draft, 1=Submitted, 2=Approved, 3=Rejected, 4=PendingDelete' AFTER status,
ADD COLUMN submission_type TINYINT NULL COMMENT '1=Create, 2=Update, 3=Delete' AFTER submission_status,
ADD COLUMN change_snapshot JSON NULL COMMENT 'Snapshot of values before modification' AFTER submission_type,
ADD COLUMN submitter_id BIGINT NULL COMMENT 'Submitter user ID' AFTER change_snapshot,
ADD COLUMN submit_time TIMESTAMP NULL COMMENT 'Submission timestamp' AFTER submitter_id,
ADD COLUMN reviewer_id BIGINT NULL COMMENT 'Reviewer user ID' AFTER submit_time,
ADD COLUMN review_time TIMESTAMP NULL COMMENT 'Review timestamp' AFTER reviewer_id,
ADD COLUMN review_notes VARCHAR(500) NULL COMMENT 'Review notes/rejection reason' AFTER review_time,
ADD COLUMN delete_time TIMESTAMP NULL COMMENT 'Soft delete timestamp' AFTER updated_time;

-- Add indexes for approval workflow
ALTER TABLE md_sensitive_word
ADD KEY idx_submission_status (submission_status),
ADD KEY idx_delete_time (delete_time);

-- Verify changes
SHOW COLUMNS FROM md_sensitive_word;
```

### 更新 Schema 文件

需要同步更新 `scripts/sql/masterdata_schema.sql` 中的表定义，使其与 Model 保持一致。

## Moderation 微服务业务匹配性分析

### 当前架构

根据代码分析，**不存在独立的 moderation 微服务**。敏感词功能被整合在 **masterdata 服务**中，包含两个层面：

#### 1. 管理层面 (API)
- **服务**: `masterdata-api` (HTTP REST API)
- **端口**: 8081
- **功能**: 敏感词的 CRUD 管理、审批工作流
- **用户**: 管理员通过 Web 后台管理

#### 2. 检测层面 (RPC) - **尚未实现**
- **服务**: `masterdata-rpc` (gRPC)
- **端口**: 8082
- **功能**: `CheckSensitiveWords` - 内容审核接口
- **用户**: 其他微服务调用进行内容过滤
- **状态**: ⚠️ **proto 文件中定义了接口，但实际代码中未实现**

### 业务场景对比

| 场景 | 需要的字段 | Model 是否支持 | Schema 是否支持 |
|------|-----------|---------------|----------------|
| **管理场景** (已实现) | | | |
| 创建敏感词 | word, category, severity, action, status | ✅ | ✅ |
| 提交审批 | submission_status, submission_type, submitter_id, submit_time | ✅ | ❌ |
| 审核通过/拒绝 | reviewer_id, review_time, review_notes | ✅ | ❌ |
| 修改回滚 | change_snapshot | ✅ | ❌ |
| 软删除/恢复 | delete_time | ✅ | ❌ |
| **检测场景** (未实现) | | | |
| 内容过滤 | word, category, severity, action, status | ✅ | ✅ |
| 只检测已审核通过的词 | submission_status = 2 | ✅ | ❌ |
| 排除已删除的词 | delete_time is null | ✅ | ❌ |

### 关键发现

**Model 的完整结构对两种业务场景都是必需的：**

1. **管理场景**（已实现）
   - 需要审批字段支持工作流
   - 需要软删除支持数据恢复
   - 需要快照支持修改回滚

2. **检测场景**（未来实现 `CheckSensitiveWords` RPC）
   - 需要 `submission_status = 2` 过滤：只检测已审核通过的敏感词
   - 需要 `delete_time is null` 过滤：排除已删除的敏感词
   - 需要 `status = 1` 过滤：只检测启用状态的敏感词

### RPC 接口实现建议

当实现 `CheckSensitiveWords` 接口时，查询逻辑应该是：

```go
// 只检测已审核通过、未删除、启用状态的敏感词
query := `
  SELECT word, category, severity, action 
  FROM md_sensitive_word 
  WHERE submission_status = 2    -- 已审核通过
    AND delete_time IS NULL      -- 未删除
    AND status = 1               -- 启用状态
`
```

**如果没有这些审批字段，RPC 接口将无法正确过滤敏感词，可能导致：**
- 检测到未审核通过的敏感词（误报）
- 检测到已删除的敏感词（误报）
- 无法区分草稿状态和生效状态

## 总结

**强烈建议采用方案一：修改 MySQL Schema 添加审批字段**

这是最合理的选择，因为：
- ✅ **管理场景**：业务逻辑已完整实现审批工作流，必须有这些字段
- ✅ **检测场景**：未来实现 `CheckSensitiveWords` RPC 时，需要这些字段过滤生效的敏感词
- ✅ **架构一致性**：保持与其他主数据实体的设计模式一致
- ✅ **实施简单**：只需执行 ALTER TABLE，无需修改代码
- ✅ **功能完整**：支持企业级审批需求和内容安全

### 当前不一致的原因

1. Schema 文件是早期版本，只考虑了基础字段
2. 开发过程中逐步添加了审批功能，但忘记同步更新 Schema
3. RPC 检测接口尚未实现，Schema 设计者可能未考虑检测场景的过滤需求

### 风险提示

**如果不添加审批字段，将导致：**
- ❌ 管理功能无法正常工作（提交、审核、回滚、恢复等功能会报错）
- ❌ 未来实现 RPC 检测接口时，无法正确过滤生效的敏感词
- ❌ 生产环境部署时出现字段缺失错误

建议尽快执行数据库迁移，避免生产环境部署时出现字段缺失错误。
