## ADDED Requirements

### Requirement: 文本审核引擎编排
系统 SHALL 实现文本审核引擎，按顺序执行：归一化 → AC 自动机匹配 → 白名单过滤 → 拆字检测 → 小模型(预留) → 大模型(预留)。

#### Scenario: AC 层直接拦截
- **WHEN** 文本包含 severity=1 的敏感词，且白名单未覆盖
- **THEN** 返回 `{Pass: false, RiskLevel: "high", CheckLayer: "ac_engine"}`

#### Scenario: AC 层放行
- **WHEN** 文本不包含任何敏感词，拆字检测也无命中
- **THEN** 返回 `{Pass: true, RiskLevel: "low", CheckLayer: "ac_engine"}`

#### Scenario: 小模型不可用降级
- **WHEN** AC 层检测到灰名单命中，小模型返回 ErrNotImplemented
- **THEN** 跳过小模型，直接调用大模型（大模型也不可用时放行并标记 need_review）

### Requirement: 图片审核引擎编排
系统 SHALL 实现图片审核引擎，按顺序执行：感知哈希比对 → 小模型(预留) → 大模型(预留)。

#### Scenario: 感知哈希命中
- **WHEN** 图片哈希与违规图库匹配（距离 <= 阈值）
- **THEN** 返回 `{Pass: false, RiskLevel: "high", CheckLayer: "image_hash"}`

#### Scenario: 模型不可用降级
- **WHEN** 感知哈希未命中，模型层返回 ErrNotImplemented
- **THEN** 放行并标记 `{NeedReview: true, RiskLevel: "medium"}`

### Requirement: 审核结果结构
审核结果 SHALL 包含 `Pass`(bool)、`RiskLevel`(high/medium/low)、`Reason`(string)、`NeedReview`(bool)、`Details`([]MatchDetail)。

#### Scenario: 结果完整性
- **WHEN** 审核完成
- **THEN** 结果包含通过状态、风险等级、原因、是否需要人工复审和命中详情列表

### Requirement: 灰名单判断逻辑
AC 层命中后 SHALL 根据以下规则分流：severity=1 直接拒绝；severity=2/3 或拆字检测命中 → 进入模型层。

#### Scenario: 高严重度直接拒绝
- **WHEN** AC 匹配到 severity=1 的词
- **THEN** 不调用模型，直接拒绝

#### Scenario: 低严重度进入模型
- **WHEN** AC 匹配到 severity=2 的词
- **THEN** 进入小模型判断流程
