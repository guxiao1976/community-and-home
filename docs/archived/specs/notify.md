## 项目背景

系统统一多通道消息下发中台。统一管理短信、APP 推送、小程序订阅消息、站内信等多渠道的消息发送，对上屏蔽各渠道 SDK 差异，对业务方提供统一的消息投递入口。遵循"一次投递，可靠送达，多渠道收敛"原则。

## 数据模型

- **MySQL (`db_notify`)**:
    - `notify_template`: `id`, `template_code`, `channel(sms/push/miniapp/inbox)`, `title`, `body(Vars模板)`, `vars_schema(JSON)`, `status(enabled/disabled)`, `audit_id(第三方模板审核ID，短信必填)`
    - `notify_record`: `id`, `user_id`, `channel`, `template_code`, `title`, `content(渲染后)`, `status(pending/sent/failed)`, `biz_id(幂等键)`, `third_party_msg_id`, `send_at`, `fail_reason`
    - `user_notify_preference`: `id`, `user_id`, `channel`, `enabled(1/0)`, `dnd_start_time(HH:mm)`, `dnd_end_time(HH:mm)`
- **Redis**:
    - 发送频率控制: `notify:rate:{user_id}:{channel}:{window}` = `count`, TTL=窗口长度
    - 全局日上限: `notify:daily:{user_id}:{channel}:{date}` = `count`, TTL=当天剩余秒数
    - 偏好缓存: `notify:pref:{user_id}` = JSON，首次查询后缓存，偏好变更时失效

## 接口清单

- `SendNotification(SendNotifyReq) returns (SendNotifyResp)` —— 业务方统一投递入口（直接 gRPC 调用，非 MQ 消费接口）
- `GetNotifyHistory(GetNotifyHistoryReq) returns (NotifyHistoryResp)` —— 查询用户历史消息列表
- `UpdateNotifyPreference(UpdatePreferenceReq) returns (BaseResp)` —— 更新用户通知偏好（渠道开关、免打扰）
- `GetNotifyPreference(GetNotifyPreferenceReq) returns (NotifyPreferenceResp)` —— 查询用户通知偏好
- `RegisterTemplate(RegisterTemplateReq) returns (BaseResp)` —— 注册/更新消息模板（管理后台接口）

## 核心逻辑流

### 1. SendNotification（可靠下发主流程）

**业务方调用方式**：业务服务通过 Outbox 模式写入本地消息表，Worker 投递到 `topic_notify_send`，Notify-Svc 消费 MQ 后执行以下流程。

1. **幂等校验**：根据 `biz_id + template_code` 查询 `notify_record` 是否已存在相同记录。存在则跳过，返回已有的 `notify_record_id`。
2. **模板渲染**：查询 `notify_template`，校验模板状态为 enabled。用请求中的 `vars` 填充模板占位符，生成最终消息标题和正文。
3. **偏好过滤**：
    - 查询 `notify:pref:{user_id}` 缓存，未命中则查 MySQL 并回填缓存。
    - 如果用户关闭了该渠道通知 → 仅写站内信（inbox），跳过即时推送。
    - 如果当前时间在 `dnd_start_time ~ dnd_end_time` 区间内 → 延迟至免打扰结束后投递，或降级为仅站内信。
4. **频率控制**：
    - 短窗口限流: `INCR notify:rate:{user_id}:{channel}:{window}`，若超过阈值则延迟 60s 重试。
    - 日上限: `INCR notify:daily:{user_id}:{channel}:{date}`，超限则写入 fail_record 并触发告警，不重试。
5. **渠道路由**：
    - **短信**：调用阿里云/腾讯云 SMS SDK 下发，获取第三方 `msg_id`。
    - **APP Push**：调用极光推送/个推/厂商通道（华为/小米/OPPO/VIVO）。
    - **小程序订阅消息**：调用微信 `sendSubscribeMessage` 接口。
    - **站内信**：直接写入 `notify_record`，由前端轮询或 WebSocket 推送未读计数。
6. **结果记录**：插入/更新 `notify_record`（status = sent/failed）。记录 `third_party_msg_id` 或 `fail_reason`。
7. **失败重试**：
    - 网络超时/第三方服务不可用 → 指数退避重试（1min → 5min → 15min），最多 3 次。
    - 3 次后仍失败 → 写入死信表 `notify_dlq`，触发告警（钉钉/企微通知运维）。

### 2. GetNotifyHistory（用户消息列表）

1. 分页查询 `notify_record`，按 `send_at DESC` 排序，游标分页。
2. 支持按 `channel`、`status` 筛选。
3. 站内信返回全文；短信/Push/小程序消息由于内容已在推送时送达，此处仅返回摘要和发送时间。

### 3. UpdateNotifyPreference / GetNotifyPreference

1. Update：接收 `user_id`, `channel`, `enabled`, `dnd_start_time`, `dnd_end_time`。Upsert `user_notify_preference` 并删除 Redis 缓存 `notify:pref:{user_id}`。
2. Get：优先查 `notify:pref:{user_id}` 缓存，未命中查 MySQL 并回填（TTL 1 小时）。

### 4. RegisterTemplate（模板管理）

1. 管理后台注册或更新消息模板。
2. 若 channel 为 sms（短信），需支持上传第三方（阿里云/腾讯云）审核通过的模板 ID，填入 `audit_id`。
3. 模板变量用 `{var_name}` 占位符，`vars_schema` 定义变量名、类型、是否必填、示例值。

```json
{
  "vars_schema": {
    "user_name": {"type": "string", "required": true, "example": "张三"},
    "community_name": {"type": "string", "required": true, "example": "阳光花园小区"},
    "approve_result": {"type": "string", "required": true, "example": "通过"}
  }
}
```

## 与各模块协作

| 触发场景 | 触发方 | 投递方式 | template_code | 示例 |
|---------|--------|---------|--------------|------|
| 实名认证审批通过/驳回 | Approve-Svc | Outbox → MQ | `identity_approved` / `identity_rejected` | "您好{user_name}，您在{community_name}的实名认证已{approve_result}" |
| 信用分扣除 | Risk-Svc | Outbox → MQ | `credit_deducted` | "您的信用分因{reason}被扣除{points}分，当前信用分{current_score}" |
| 新设备登录告警 | Auth-Svc | Outbox → MQ | `new_device_login` | "您的账号于{time}在{location}通过{device}登录，如非本人操作请立即修改密码" |
| 角色变更通知 | Permission-Svc | Outbox → MQ | `role_changed` | "您已被{operator}设为{role_name}，权限范围：{scope_desc}" |
| 审批待办提醒 | Approve-Svc | Outbox → MQ | `approval_pending` | "您有一条来自{applicant_name}的{biz_type_name}审批待处理" |

## 可靠投递保障

```
业务服务                         Notify-Svc
   │                                │
   │ ① 写入业务数据 + outbox_event    │
   │    (同一 DB 事务)                │
   │                                │
   │ ② Worker 投递 MQ                │
   │─────────topic_notify──────────▶│
   │                                │ ③ 消费，幂等校验
   │                                │ ④ 模板渲染 + 偏好过滤
   │                                │ ⑤ 频率控制
   │                                │ ⑥ 渠道下发
   │                                │ ⑦ 写 notify_record
   │                                │
   │                        失败 → 延迟重试 (最多3次)
   │                        最终失败 → DLQ + 告警
```
