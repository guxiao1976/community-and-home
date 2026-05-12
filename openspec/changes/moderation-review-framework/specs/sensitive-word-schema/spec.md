## MODIFIED Requirements

### Requirement: 敏感词表结构
`md_sensitive_word` 表 SHALL 扩展以下字段：`word_type`(TINYINT, 1=黑名单/2=白名单, 默认1)、`pinyin_expanded`(TINYINT(1), 是否已生成谐音变体, 默认0)、`submission_status`(TINYINT, 默认0)、`submission_type`(TINYINT, 默认NULL)、`change_snapshot`(JSON, 默认NULL)、`submitter_id`(BIGINT, 默认NULL)、`submit_time`(DATETIME, 默认NULL)、`reviewer_id`(BIGINT, 默认NULL)、`review_time`(DATETIME, 默认NULL)、`review_notes`(VARCHAR(500), 默认NULL)、`delete_time`(DATETIME, 默认NULL)。

#### Scenario: 新增黑名单词
- **WHEN** 插入一条 word_type=1 的记录
- **THEN** 该词被 moderation 服务的 AC 自动机加载为黑名单词

#### Scenario: 新增白名单词
- **WHEN** 插入一条 word_type=2 的记录
- **THEN** 该词被 moderation 服务的白名单 AC 自动机加载为豁免词

#### Scenario: 向后兼容
- **WHEN** 已有记录不包含 word_type 字段
- **THEN** 默认值 1（黑名单）保证行为不变

### Requirement: Severity 语义
severity 字段语义 SHALL 调整为：1=High（高风险，直接拦截）、2=Medium（中风险，进入模型审核）、3=Low（低风险，进入模型审核）。

#### Scenario: 高风险直接拦截
- **WHEN** 词的 severity=1 被命中
- **THEN** AC 层直接拒绝，不进入模型层

#### Scenario: 中低风险模型审核
- **WHEN** 词的 severity=2 或 3 被命中
- **THEN** 进入小模型/大模型审核流程
