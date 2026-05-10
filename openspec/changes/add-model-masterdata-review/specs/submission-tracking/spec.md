## ADDED Requirements

### Requirement: submission_type 字段记录操作类型
4 张主数据表（md_residential_area、md_administrative_division、md_configuration、md_sensitive_word）MUST 新增 `submission_type TINYINT NULL` 列，取值：1=新增, 2=修改, 3=删除。

#### Scenario: 创建记录时设置操作类型
- **WHEN** 用户新建任意主数据记录
- **THEN** 系统 SHALL 设置 `submission_type = 1`（新增）

#### Scenario: 修改记录时设置操作类型
- **WHEN** 用户修改任意主数据记录
- **THEN** 系统 SHALL 设置 `submission_type = 2`（修改）

#### Scenario: 删除记录时设置操作类型
- **WHEN** 用户删除任意主数据记录
- **THEN** 系统 SHALL 设置 `submission_type = 3`（删除）

### Requirement: change_snapshot 字段存储修改前快照
4 张主数据表 MUST 新增 `change_snapshot TEXT NULL` 列，存储修改前记录的 JSON 快照。

#### Scenario: 修改前保存快照
- **WHEN** 用户修改主数据记录
- **THEN** 系统 SHALL 在应用修改之前，将当前记录的业务字段序列化为 JSON 存入 `change_snapshot`，仅保留业务字段（排除 id、created_time、updated_time、submission_status、submission_type、submitter_id、submit_time、reviewer_id、review_time、review_notes）

#### Scenario: 新建时无快照
- **WHEN** 用户新建主数据记录
- **THEN** `change_snapshot` SHALL 为 NULL

#### Scenario: 删除时无快照
- **WHEN** 用户删除主数据记录（标记待删除）
- **THEN** `change_snapshot` SHALL 保持原值不变

### Requirement: 修改提交被拒绝后恢复原始数据
当修改操作（submission_type=2）被审核拒绝时，MUST 从 change_snapshot 恢复原始数据。

#### Scenario: 拒绝修改恢复原数据
- **WHEN** 审核人拒绝一条 submission_type=2 的待审记录
- **THEN** 系统 SHALL 从 `change_snapshot` JSON 反序列化并覆盖当前记录的业务字段，`submission_status` 设为 3（已拒绝），`submission_type` 清空为 NULL

#### Scenario: 拒绝非修改操作不恢复
- **WHEN** 审核人拒绝一条 submission_type=1（新增）或 submission_type=3（删除）的待审记录
- **THEN** 系统 SHALL NOT 从 change_snapshot 恢复数据

### Requirement: 审核人字段补齐
md_administrative_division、md_configuration、md_sensitive_word MUST 新增 `submitter_id BIGINT NULL`、`submit_time TIMESTAMP NULL`、`reviewer_id BIGINT NULL`、`review_time TIMESTAMP NULL`、`review_notes VARCHAR(500) NULL` 字段，与 md_residential_area 保持一致。

#### Scenario: 提交时记录提交人
- **WHEN** 用户提交任意主数据记录进行审核
- **THEN** 系统 SHALL 设置 `submitter_id` 为当前用户 ID，`submit_time` 为当前时间

#### Scenario: 审核时记录审核人
- **WHEN** 审核人通过或拒绝任意主数据记录
- **THEN** 系统 SHALL 设置 `reviewer_id` 为当前审核人 ID，`review_time` 为当前时间，`review_notes` 为审核备注

### Requirement: 敏感词表补齐 delete_time
md_sensitive_word MUST 新增 `delete_time DATETIME NULL` 列用于软删除。

#### Scenario: 敏感词软删除
- **WHEN** 敏感词的删除审批通过
- **THEN** 系统 SHALL 设置 `delete_time` 为当前时间，查询时过滤 `delete_time IS NOT NULL` 的记录
