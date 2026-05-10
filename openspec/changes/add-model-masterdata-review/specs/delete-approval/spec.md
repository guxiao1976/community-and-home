## ADDED Requirements

### Requirement: 行政区划删除走审批
md_administrative_division 的删除操作 MUST 标记为待删除（submission_status=4），而非直接软删除。

#### Scenario: 删除行政区划
- **WHEN** 用户请求删除一条行政区划记录（非已删除、非已处于待删除状态）
- **THEN** 系统 SHALL 设置 `submission_status = 4`（待删除）、`submission_type = 3`（删除），不设置 delete_time

#### Scenario: 待删除的行政区划审批通过
- **WHEN** 审核人通过一条 submission_status=4 的行政区划记录
- **THEN** 系统 SHALL 设置 `delete_time` 为当前时间，`submission_status` 保持 4

#### Scenario: 待删除的行政区划审批拒绝
- **WHEN** 审核人拒绝一条 submission_status=4 的行政区划记录
- **THEN** 系统 SHALL 恢复 `submission_status = 3`（已拒绝），`submission_type` 清空为 NULL

### Requirement: 系统配置删除走审批
md_configuration 的删除操作 MUST 标记为待删除（submission_status=4），而非直接软删除。

#### Scenario: 删除系统配置
- **WHEN** 用户请求删除一条系统配置记录（非已删除、非已处于待删除状态）
- **THEN** 系统 SHALL 设置 `submission_status = 4`（待删除）、`submission_type = 3`（删除），不设置 delete_time

#### Scenario: 待删除的系统配置审批通过
- **WHEN** 审核人通过一条 submission_status=4 的系统配置记录
- **THEN** 系统 SHALL 设置 `delete_time` 为当前时间，`submission_status` 保持 4

#### Scenario: 待删除的系统配置审批拒绝
- **WHEN** 审核人拒绝一条 submission_status=4 的系统配置记录
- **THEN** 系统 SHALL 恢复 `submission_status = 3`（已拒绝），`submission_type` 清空为 NULL

### Requirement: 敏感词删除走审批
md_sensitive_word 的删除操作 MUST 标记为待删除（submission_status=4），而非直接软删除。

#### Scenario: 删除敏感词
- **WHEN** 用户请求删除一条敏感词记录（非已删除、非已处于待删除状态）
- **THEN** 系统 SHALL 设置 `submission_status = 4`（待删除）、`submission_type = 3`（删除），不设置 delete_time

#### Scenario: 待删除的敏感词审批通过
- **WHEN** 审核人通过一条 submission_status=4 的敏感词记录
- **THEN** 系统 SHALL 设置 `delete_time` 为当前时间，`submission_status` 保持 4

#### Scenario: 待删除的敏感词审批拒绝
- **WHEN** 审核人拒绝一条 submission_status=4 的敏感词记录
- **THEN** 系统 SHALL 恢复 `submission_status = 3`（已拒绝），`submission_type` 清空为 NULL

### Requirement: 待删除记录默认不出现在列表中
所有主数据表的默认列表查询 MUST 排除 submission_status=4（待删除）的记录。

#### Scenario: 默认查询排除待删除
- **WHEN** 前端请求列表数据且未指定 submission_status 过滤条件
- **THEN** 系统 SHALL 查询结果中不包含 submission_status=4 的记录

#### Scenario: 主动查看待删除记录
- **WHEN** 前端请求列表数据且指定 submission_status=4
- **THEN** 系统 SHALL 返回 submission_status=4 的记录
