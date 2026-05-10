## ADDED Requirements

### Requirement: 提交记录表结构
系统 SHALL 新建 `md_submission_record` 表，包含以下字段：
- `id` BIGINT 自增主键
- `entity_type` VARCHAR(50) NOT NULL，实体类型（administrative_division / residential_area / configuration / sensitive_word）
- `entity_id` BIGINT NOT NULL，实体 ID（冗余存储，实体删除后记录保留）
- `entity_name` VARCHAR(200)，实体名称（冗余存储，用于删除后展示）
- `entity_code` VARCHAR(100)，实体代码（冗余存储，可为 NULL）
- `submission_type` TINYINT NOT NULL，操作类型（1=新增, 2=修改, 3=删除）
- `submitter_id` BIGINT NOT NULL，提交人 ID
- `submit_time` TIMESTAMP NOT NULL，提交时间
- `reviewer_id` BIGINT NULL，审核人 ID
- `review_time` TIMESTAMP NULL，审核时间
- `review_result` TINYINT NOT NULL DEFAULT 0，审核结果（0=待审核, 1=通过, 2=拒绝, 3=已撤回）
- `review_notes` VARCHAR(500) NULL，审核备注
- `created_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

表 SHALL 包含索引：`idx_submitter` (submitter_id), `idx_reviewer` (reviewer_id), `idx_entity` (entity_type, entity_id), `idx_result` (review_result)。

#### Scenario: 建表成功
- **WHEN** 执行建表 DDL
- **THEN** 表创建成功，索引正确建立

### Requirement: 提交时写入记录
当任意实体的提交审核逻辑执行时，系统 SHALL 在 `md_submission_record` 中插入一条记录，`review_result=0`（待审核），同时写入 `entity_name` 和 `entity_code`。

#### Scenario: 行政区划提交审核
- **WHEN** 调用 submitDivision API 提交 id=100 的行政区划（name="测试社区", code="370502001"）
- **THEN** 插入一条记录：entity_type=administrative_division, entity_id=100, entity_name="测试社区", entity_code="370502001", submission_type 对应实际类型, review_result=0

### Requirement: 审核通过时更新记录
审核通过时，系统 SHALL 更新对应提交记录的 `review_result=1, reviewer_id, review_time`。

#### Scenario: 审核通过
- **WHEN** 审核人 id=5 通过 entity_id=100 的审核
- **THEN** 该条记录的 review_result=1, reviewer_id=5, review_time 为当前时间

### Requirement: 审核拒绝时更新记录
审核拒绝时，系统 SHALL 更新对应提交记录的 `review_result=2, reviewer_id, review_time, review_notes`。即使实体被物理删除，记录 MUST 保留。

#### Scenario: 新增被拒绝后记录仍存在
- **WHEN** 新增的 entity_id=200 被拒绝并物理删除
- **THEN** md_submission_record 中 entity_id=200 的记录 review_result=2，记录仍然存在

### Requirement: 撤回时更新记录
撤回提交时，系统 SHALL 更新对应提交记录的 `review_result=3`。

#### Scenario: 撤回提交
- **WHEN** 提交人撤回 entity_id=300 的提交
- **THEN** 该条记录的 review_result=3

### Requirement: 覆盖所有主数据实体
提交记录写入逻辑 SHALL 覆盖以下 4 种实体的提交/审核/撤回操作：administrative_division、residential_area、configuration、sensitive_word。

#### Scenario: 住宅小区审核通过
- **WHEN** 审核人通过一条 residential_area 的审核
- **THEN** md_submission_record 中对应记录 review_result=1

#### Scenario: 敏感词审核拒绝
- **WHEN** 审核人拒绝一条 sensitive_word 的审核
- **THEN** md_submission_record 中对应记录 review_result=2, review_notes 已填充
