## 项目背景

系统统一操作审计中台。以非侵入方式旁路采集全平台所有服务的敏感操作行为，提供统一的存储、查询、导出和防篡改验证能力。审计中心**不参与业务流程**，仅做"事后追溯"——它与风控中心形成互补：风控往前看（事前预警+事中阻断），审计往后看（事后追溯+合规举证）。

建立独立审计中心的三个刚性需求：
- **合规底线**：《个人信息保护法》要求对所有涉及个人敏感信息的操作保留审计记录，保存期限≥3年，且具备防篡改能力
- **安全防线**：内部越权行为（物业管理员私查居民身份证、员工批量导出数据）需要通过审计日志追溯举证
- **运维基线**：跨服务排障时，统一审计日志是还原问题现场的最可靠手段

## 数据模型

### MySQL (`db_audit`) —— 热数据，近 90 天

- `audit_event`: `id`, `trace_id`, `event_time`, `operator_id`, `operator_name`, `operator_type(staff/resident/admin/system)`, `operator_ip`, `resource_type`, `resource_id`, `action(CREATE/READ/UPDATE/DELETE/EXPORT/LOGIN/LOGOUT)`, `api_path`, `result(SUCCESS/FAILED/DENIED)`, `error_code`, `detail(人类可读描述)`, `service_name`, `request_params_hash(SHA256关键参数)`, `created_at`
    - 按天分表：`audit_event_20260529`
    - 复合索引：`(operator_id, event_time)`, `(resource_type, resource_id)`, `(action, result)`

### ClickHouse —— 温数据，90 天 ~ 1 年

- `audit_event`：与 MySQL 同 Schema，ReplacingMergeTree 引擎，按月分区。支持运营人员秒级多维度筛选查询。

### 对象存储 (OSS/S3) —— 冷数据，1 年 ~ 3 年+

- Parquet 格式按月归档，附带 `manifest.json`（含 Merkle Tree 根哈希），支持防篡改验证。

## 接口清单

- (MQ 消费) `ConsumeAuditEvent` —— 消费各服务投递的审计事件，持久化存储
- `QueryAuditLog(QueryAuditLogReq) returns (QueryAuditLogResp)` —— 运营/安全人员多维度查询审计日志
- `ExportAuditLog(ExportAuditLogReq) returns (ExportAuditLogResp)` —— 合规导出，异步生成下载链接
- `VerifyAuditIntegrity(VerifyIntegrityReq) returns (VerifyIntegrityResp)` —— 验证指定时间范围的审计日志哈希链完整性

## 核心逻辑流

### 1. 事件采集（非侵入旁路）

各业务服务通过 **gRPC Unary Interceptor** 统一埋点，自动捕获取决于 Proto 标注的审计级别（见下文"审计标注规范"章节）。Interceptor 拦截请求：

```
Interceptor 拦截请求
  │
  ├── ① 提取上下文 (从 gRPC Metadata)
  │     - trace_id: key "x-trace-id"
  │     - operator_id: key "x-user-id" (API Gateway 验证 AT 后注入)
  │     - operator_type: key "x-user-type"
  │     - operator_name: key "x-user-name"
  │     - operator_ip: peer IP 或 "x-forwarded-for"
  │
  ├── ② 执行实际业务逻辑 (handler)
  │
  ├── ③ 判断是否需要审计
  │     - 读取 Proto 中标注的 @audit 级别
  │     - NONE → 跳过
  │     - SENSITIVE → 只审计写操作和敏感字段读取
  │     - ALL → 所有操作（含普通读）均审计
  │
  └── ④ 异步投递 (不阻塞主流程)
        - 构造 AuditEvent 消息
        - 投递到 Kafka/Pulsar topic_audit_event
        - 投递失败 → 写入本地 outbox_audit 表 (Worker 补投)
```

### 2. 事件存储（分层归档）

```
MQ Consumer (audit-svc)
  │
  ├── ① 写入 MySQL (热层，90 天)
  │     按天分表，批量 insert
  │
  ├── ② 同步写入 ClickHouse (温层，90天 ~ 1年)
  │     ReplacingMergeTree，按月分区
  │
  └── ③ 归档任务 (每日凌晨 03:00)
        - 扫描 MySQL 中 event_time > 90天 的数据
        - 导出为 Parquet 文件
        - 计算 Merkle Tree 根哈希，写入 manifest.json
        - 上传至对象存储
        - 删除 MySQL 中已归档数据
        - ClickHouse 保留至 1 年后删除
```

### 3. QueryAuditLog（多维度查询）

1. 接收筛选条件：`operator_id`, `operator_type`, `resource_type`, `resource_id`, `action`, `result`, `service_name`, `start_time`, `end_time`。均为可选字段。
2. 路由存储层：
    - `event_time` 在 90 天内 → 查 MySQL
    - 90 天 ~ 1 年 → 查 ClickHouse
    - 超过 1 年 → 提示"超出在线查询范围，请使用 ExportAuditLog 进行合规导出"
3. 默认按 `event_time DESC` 排序。
4. 游标分页：`next_token` 编码了上次查询的最后一条记录位置。
5. **脱敏处理**：查询结果中 `operator_ip` 打码后三位，`detail` 字段若涉及身份证号/手机号则做部分打码。

### 4. ExportAuditLog（合规导出）

1. 接收筛选条件和导出格式（CSV / JSON / Parquet）。
2. 创建异步导出任务，返回 `task_id`。
3. 后台 Worker 从 ClickHouse 或对象存储拉取数据，拼装为文件。
4. 文件上传至 OSS 私有 Bucket，生成预签名下载链接（有效期 1 小时）。
5. 通过 **Notify-Svc** 推送下载链接给操作者（站内信 + 短信）。
6. **导出操作本身写入 audit_event**。

### 5. VerifyAuditIntegrity（防篡改验证）

1. 接收时间范围（如 `2026-05`），从对象存储拉取对应的 Parquet 文件和 `manifest.json`。
2. 按 manifest 中记录的算法重建 Merkle Tree，计算根哈希。
3. 与 manifest 中的根哈希对比：
    - 一致 → 返回 `pass`
    - 不一致 → 返回 `fail`，触发安全告警
    - 文件缺失 → 返回 `missing`，标记异常
4. 返回验证结果摘要。

## 审计标注规范

各服务在 Proto 文件中通过注释标注接口的审计级别。Interceptor 在运行时读取 Method Descriptor 的注释决定是否采集：

```protobuf
// @audit: SENSITIVE  — 仅写操作 + 敏感字段读取时审计
rpc GetUserProfile(GetUserReq) returns (UserProfileResp);

// @audit: ALL  — 所有操作（含普通读）均审计
rpc AssignRole(AssignRoleReq) returns (BaseResp);

// 无标注  — 等同于 NONE，不审计
rpc GetCommunityList(GetCommunityListReq) returns (CommunityListResp);
```

## 与各模块协作

### 各模块需审计的操作明细

| 服务 | 操作 | @audit 级别 | 原因 |
|------|------|:---:|------|
| **User-Svc** | GetUserProfile（查实名/手机/身份证） | SENSITIVE | 涉及个人敏感信息，个保法合规要求 |
| **User-Svc** | CreateUser、SubmitIdentity、UpdateIdentityStatus | ALL | 实名状态的任何变更必须可追溯 |
| **User-Svc** | DeductCredit | ALL | 信用分涉及用户权益，变更必须留痕 |
| **Auth-Svc** | Login（记录成功/失败/异地） | SENSITIVE | 登录安全基线——登录失败也是关键安全信号 |
| **Auth-Svc** | Logout、KickDevice | ALL | 会话管理操作 |
| **Permission-Svc** | AssignRole、RevokeRole | ALL | 权限变更是安全红线，100% 审计 |
| **Permission-Svc** | CheckPermission（仅记录 DENIED） | SENSITIVE | 只记录被拒绝的鉴权请求，用于发现越权探测行为 |
| **Approve-Svc** | CreateApplication、ApproveTask | ALL | 审批流全程留痕，防止审批舞弊 |
| **Risk-Svc** | CheckContent（违规拦截）、ReportRiskEvent | ALL | 风控决策可追溯，用于申诉复核 |
| **Notify-Svc** | SendNotification | ALL | 消息推送记录，用于客诉回查 |
| **Audit-Svc** | ExportAuditLog | ALL | 导出审计日志本身是敏感操作，防止审计日志被内部泄露 |

### 服务协作架构图

```
┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐
│ User-Svc │ │ Auth-Svc │ │Perm-Svc  │ │Approve-  │ │ Risk-Svc │ │Notify-Svc│
│          │ │          │ │          │ │Svc       │ │          │ │          │
└────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘ └────┬─────┘
     │            │            │            │            │            │
     │    gRPC Unary Interceptor (审计埋点)                              │
     │   • 读取 @audit 标注 → 决定是否采集                               │
     │   • 自动捕获: trace_id, user_id, api_path, result                │
     │   • 普通读自动跳过，敏感操作/写操作异步投递 MQ                       │
     │            │            │            │            │            │
     └────────────┼────────────┼────────────┼────────────┼────────────┘
                  │            │            │            │
                  └────────────┼────────────┼────────────┘
                               │            │
                     Kafka/Pulsar topic_audit_event
                               │
                               ▼
                  ┌───────────────────────────────────┐
                  │          Audit-Svc                 │
                  │                                    │
                  │  ┌──────────────────────────────┐ │
                  │  │ Consumer: 消费 → 写入 → 归档   │ │
                  │  └──────────────────────────────┘ │
                  │                                    │
                  │  ┌──────────────────────────────┐ │
                  │  │ MySQL (90天)  按天分表        │ │
                  │  │ 热数据，支撑在线秒级查询       │ │
                  │  └──────────────────────────────┘ │
                  │  ┌──────────────────────────────┐ │
                  │  │ ClickHouse (1年)  按月分区    │ │
                  │  │ 温数据，运营多维分析          │ │
                  │  └──────────────────────────────┘ │
                  │  ┌──────────────────────────────┐ │
                  │  │ OSS (3年+)  Parquet + 哈希链  │ │
                  │  │ 冷数据，合规归档 + 防篡改     │ │
                  │  └──────────────────────────────┘ │
                  │                                    │
                  │  提供: QueryAuditLog               │
                  │       ExportAuditLog               │
                  │       VerifyAuditIntegrity          │
                  └───────────────────────────────────┘
```

## 非功能约束

| 约束项 | 目标值 | 说明 |
|--------|:---:|------|
| 采集延迟 (P99) | < 1s | 从 Interceptor 投递到 MQ 到 MySQL 写入完成的端到端延迟 |
| 在线查询延迟 (P95) | < 2s | 90 天热数据，任意筛选条件组合 |
| ClickHouse 查询延迟 (P95) | < 5s | 90天~1年温数据，多维度聚合查询 |
| 归档完整性 | 99.99% | MySQL → OSS 归档不丢失不重复 |
| 导出任务完成 (P95) | < 10min | 1 年数据全量导出 |
| 日写入量 | ~100 万条 | 按日均 10 万活跃用户估算 |
| 存储容量 | ~100 GB/年 | 单条约 300B，含索引冗余 |
| 数据留存 | 热 90天 / 温 1年 / 冷 ≥3年 | 满足个保法最低 3 年要求 |
