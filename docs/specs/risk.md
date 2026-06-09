## 基本信息
- 名称：合规风控中心
- 功能：高频拦截与事后处罚。事前拦截违规发布，事后扣减信用分。

## 数据模型

- **MySQL (`db_risk`)**:
    - `risk_event_log`: `id`, `user_id`, `risk_type(porn/politics)`, `content_snapshot`, `penalty_points`
    - `risk_strategy_config`: `id`, `strategy_code(post_rate_limit)`, `strategy_value(JSON)`

## 接口清单

- `CheckContent(CheckContentReq) returns (CheckContentResp)`
- `GetRiskStrategy(GetStrategyReq) returns (StrategyResp)`
- `ReportRiskEvent(ReportRiskEventReq) returns (BaseResp)`

## 核心逻辑流

1. **CheckContent**:
    - **gRPC 同步调用**。接收文本/图片，执行敏感词匹配/AI审核。
    - 若违规：返回 `is_safe = false`，**同时**在本地异步记录日志并发送 MQ 事件。
    - 若安全：返回 `is_safe = true`，放行。
2. **ReportRiskEvent (内部处罚触发)**:
    - 插入 `risk_event_log` 记录。
    - **Outbox 模式**：在同一事务写入 `outbox_messages`，Event Type 为 `topic_risk_credit_deduct`，触发 User-Service 扣分。
