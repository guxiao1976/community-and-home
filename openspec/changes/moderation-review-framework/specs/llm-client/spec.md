## ADDED Requirements

### Requirement: LLM 客户端接口定义
系统 SHALL 定义 `LLMClient` 接口，包含 `CheckText` 和 `CheckImage` 方法，返回统一的结构化结果。

#### Scenario: 接口定义
- **WHEN** 引擎层需要调用模型审核
- **THEN** 可通过 `LLMClient` 接口调用，不依赖具体实现

### Requirement: Ollama 客户端预留
系统 SHALL 提供 `OllamaClient` 结构体实现 `LLMClient` 接口，方法体当前返回 `ErrNotImplemented`。

#### Scenario: 调用空实现
- **WHEN** 调用 `OllamaClient.CheckText(ctx, content, contentType)`
- **THEN** 返回 `ErrNotImplemented` 错误

### Requirement: 远端大模型客户端预留
系统 SHALL 提供 `RemoteLLMClient` 结构体实现 `LLMClient` 接口，方法体当前返回 `ErrNotImplemented`。

#### Scenario: 调用空实现
- **WHEN** 调用 `RemoteLLMClient.CheckText(ctx, content, contentType)`
- **THEN** 返回 `ErrNotImplemented` 错误

### Requirement: CheckResult 结构
`CheckResult` SHALL 包含 `Compliant`(bool)、`Confidence`(float64 0-1)、`Reason`(string)、`Category`(string) 字段。

#### Scenario: 结构完整性
- **WHEN** 模型返回审核结果
- **THEN** 结果包含合规性判断、置信度、原因和类别信息
