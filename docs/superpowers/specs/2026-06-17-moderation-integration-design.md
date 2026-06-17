# 内容审核全链路集成 — 设计文档

## 概述

将社区平台所有内容发布点接入 moderation-service 的合规审核管线（AC 引擎 + 大模型），实现机器审核→人工审核的完整闭环，并开发人工审核管理界面。

## 动机

- **现状问题**：community-hub-service（通知/寻失）、user-service（昵称/认证）的发布内容均未接入审核，存在合规风险
- **审核闭环缺失**：`SubmitReview` RPC 为 stub，`mod_audit_log` 虽预留人审字段但无实际流程
- **前端空白**：仅有一个管线配置测试页，无人工审核工作界面
- **目标**：所有用户生成内容均经过审核管线，高风险/不确定内容自动推送人工审核

## 关键决策

| # | 决策 | 选项 |
|---|------|------|
| 1 | 审核时序 | **异步模式**：内容先入库标记"待审核"，后台管线异步执行 |
| 2 | 消息机制 | **Redis List**：LPUSH/BRPOP，复用现有 Redis 中间件 |
| 3 | 板块划分 | **source_type 枚举**：`notice` / `lost_found` / `certification` / `nickname` |
| 4 | 图文审核 | **分别审核+合并结果**：文本→CheckText，图片→CheckImage（预留），取最严结论 |
| 5 | 数据库 | **内容表加审核字段**：`moderation_status` + `moderation_time`，`mod_audit_log` 存审核过程 |

---

## 一、整体架构与数据流

### 1.1 异步审核流程

```
用户发布内容
    │
    ▼
[业务服务] (community-hub / user-service)
    │  1. 内容落库（moderation_status=0 待审核）
    │  2. 写入 mod_audit_log（review_status=0 待机器审核）
    │  3. LPUSH moderation:task:queue <task_json>
    │  4. 返回给前端（内容标记"审核中"）
    │
    ▼
[moderation-service 消费者] (goroutine，随 RPC 启动)
    │  BRPOP moderation:task:queue 阻塞监听
    │  解析任务 → 调用管线
    │
    ├── 文本 → CheckText 管线 (AC→大模型)
    ├── 图片 → CheckImage (预留)
    │
    ▼
[审核结果回写]
    │  pass=true  + need_review=false → 内容表 moderation_status=1 (机器通过)
    │  pass=false + need_review=true  → 内容表 moderation_status=2 (机器不通过)
    │                                    mod_audit_log.review_status=0 (入人审队列)
    │
    ▼
[人工审核]
    │  人审界面 → 点击通过/不通过
    │  → 内容表 moderation_status=3/4
    │  → mod_audit_log.review_status=1/2
    │  → 前端展示最终状态
```

### 1.2 Redis 任务消息格式

```json
{
  "task_id": "snowflake_id",
  "source_type": "notice|lost_found|certification|nickname",
  "source_id": "内容ID",
  "action": "create|update",
  "items": [
    {"type": "text", "content": "标题+正文", "field": "content"},
    {"type": "image", "content": "图片URL", "field": "attachment"}
  ]
}
```

### 1.3 涉及服务

| 服务 | 改动范围 |
|------|---------|
| `community-hub-service` | 发布/编辑通知、寻失时：加审核状态字段、写 mod_audit_log、发 Redis 任务 |
| `user-service` | 创建用户(nickname)、提交认证时：同上 |
| `moderation-service` | 新增 Redis 消费者、实现 SubmitReview RPC、新增人审列表/详情 API |
| `web/pc` | 新增人工审核界面、现有内容列表增加审核状态列 |

---

## 二、数据库修改

### 2.1 各内容表新增审核字段

**notices 表**（community-hub-service / `community_db`）：

```sql
ALTER TABLE notices ADD COLUMN moderation_status TINYINT NOT NULL DEFAULT 0 
  COMMENT '0=待审核 1=机器通过 2=机器不通过 3=人审通过 4=人审不通过';
ALTER TABLE notices ADD COLUMN moderation_time DATETIME NULL COMMENT '审核时间';
```

**lost_found_items 表**（community-hub-service / `community_db`）：

```sql
ALTER TABLE lost_found_items ADD COLUMN moderation_status TINYINT NOT NULL DEFAULT 0 
  COMMENT '0=待审核 1=机器通过 2=机器不通过 3=人审通过 4=人审不通过';
ALTER TABLE lost_found_items ADD COLUMN moderation_time DATETIME NULL COMMENT '审核时间';
```

**users 表**（user-service / `user_db`）：

```sql
ALTER TABLE users ADD COLUMN nickname_moderation_status TINYINT NOT NULL DEFAULT 0 
  COMMENT '0=待审核 1=机器通过 2=机器不通过 3=人审通过 4=人审不通过';
```

**certifications 表**（user-service / `user_db`）：

```sql
ALTER TABLE certifications ADD COLUMN moderation_status TINYINT NOT NULL DEFAULT 0 
  COMMENT '0=待审核 1=机器通过 2=机器不通过 3=人审通过 4=人审不通过';
ALTER TABLE certifications ADD COLUMN moderation_time DATETIME NULL COMMENT '审核时间';
```

### 2.2 mod_audit_log 表（已有，明确 source_type 枚举）

| source_type | 含义 | 对应表 | 对应服务 |
|-------------|------|--------|---------|
| `notice` | 通知公告 | notices | community-hub-service |
| `lost_found` | 寻失互助 | lost_found_items | community-hub-service |
| `certification` | 房主认证 | certifications | user-service |
| `nickname` | 用户昵称 | users | user-service |

`mod_audit_log` 现有字段完全满足需求（`review_status`、`reviewer_id`、`review_time`、`need_review`），无需新增列。

### 2.3 mod_audit_log 写入机制

业务服务不能直写 moderation_db（违反服务边界），需通过 gRPC 调用 moderation-service 来写审核日志。

**Proto 新增 RPC**（`api/moderation/v1/moderation.proto`）：

```protobuf
rpc CreateAuditLog(CreateAuditLogRequest) returns (CreateAuditLogResponse);
```

```protobuf
message CreateAuditLogRequest {
    string content_type = 1;     // text / image
    string content_summary = 2;  // 内容摘要（≤100 字符）
    string risk_level = 3;       // 初始：low
    bool pass = 4;               // 初始：false（待审核）
    string reason = 5;           // 初始：空
    string check_layer = 6;      // 初始：空（审核后回填）
    string matched_items = 7;    // 初始：空（审核后回填）
    int64 user_id = 8;           // 发布人 ID
    string source_type = 9;      // notice / lost_found / certification / nickname
    int64 source_id = 10;        // 对应内容表主键
    bool need_review = 11;       // 初始：false
}

message CreateAuditLogResponse {
    common.v1.BaseResp base = 1;
    int64 id = 2;                // 新创建的 audit_log id
}
```

调用链：业务服务 → `moderationClient.CreateAuditLog()` → 返回 `audit_log_id` → 写入 Redis 任务（带上 `audit_log_id`）→ 消费者审核完成后 UPDATE 该记录。

---

##三、后端修改

### 3.1 community-hub-service

**（1）Notice 发布/编辑接入审核**

`rpc/internal/logic/notice/createnoticelogic.go` 流程变更：

```
原流程：校验 → 生成ID → INSERT → 返回ID
新流程：校验 → 生成ID → INSERT (moderation_status=0)
        → moderationClient.CreateAuditLog(source_type=notice, ...) → 获得 audit_log_id
        → LPUSH Redis moderation:task:queue {task_json 含 audit_log_id} → 返回 notice_id
```

`updatenoticelogic.go` 同理，编辑后重新发起审核（moderation_status 重置为 0）。

**（2）LostFound 发布接入审核**

`createlostfoundlogic.go`：同上模式。文本字段（Title+Description）走 CheckText，ImageUrls 走 CheckImage（预留），合并结果。

**（3）新增 moderation-service gRPC 客户端**

`rpc/internal/svc/servicecontext.go` 中注入 `ModerationRpc`，用于写入 `mod_audit_log` 和接收审核结果回调。

**（4）新增 UpdateModerationStatus RPC 方法**

Proto 新增一个RPC，供 moderation-service 消费者审核完成后回调更新内容表的 `moderation_status`：

```protobuf
// 新增到 api/community/v1/community.proto
rpc UpdateNoticeModerationStatus(UpdateModerationStatusRequest) returns (UpdateModerationStatusResponse);
rpc UpdateLostFoundModerationStatus(UpdateModerationStatusRequest) returns (UpdateModerationStatusResponse);
```

### 3.2 user-service

**（5）CreateUser 昵称审核**

`create_user_logic.go`：nickname 不为空时，创建用户后发 Redis 审核任务。审核结果回写到 `users.nickname_moderation_status`。

**（6）SubmitCertification 审核**

`submit_certification_logic.go`：提交认证材料后，RealName（文本）+ DocumentUrls（附件）发 Redis 审核任务。审核结果回写到 `certifications.moderation_status`。

**（7）新增 moderation-service gRPC 客户端 + UpdateModerationStatus RPC**

同 community-hub-service 模式。

### 3.3 moderation-service（核心改动）

**（8）新增 Redis 任务消费者**

目录结构：
```
rpc/internal/consumer/
  ├── task_consumer.go    # BRPOP 阻塞监听 + 并发控制
  └── task_handler.go     # 解析任务 → 调管线 → 回写结果
```

启动方式：在 `rpc/moderation.go` 中启动 goroutine，随 gRPC 服务一起运行。

核心逻辑：
```go
// task_consumer.go
func (c *TaskConsumer) Run(ctx context.Context) {
    for i := 0; i < c.workerCount; i++ {
        go c.worker(ctx, i)
    }
}

func (c *TaskConsumer) worker(ctx context.Context, id int) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            result, err := c.redis.BRPop(ctx, 5*time.Second, "moderation:task:queue").Result()
            if err != nil { continue }
            c.handleTask(ctx, []byte(result[1]))
        }
    }
}
```

**审核失败重试**：失败任务 LPUSH 到 `moderation:task:retry`，最多重试 3 次。

**消费者配置**（`rpc/etc/moderation.yaml` 新增）：
```yaml
Consumer:
  WorkerCount: 3          # 并发消费者数
  PollTimeout: 5          # BRPOP 超时秒数
  MaxRetry: 3             # 最大重试次数
```

**（9）审核结果回写**

消费者审核完成后，通过 gRPC 回调业务服务的 `UpdateModerationStatus` RPC，更新内容表的 `moderation_status`，同时 UPDATE `mod_audit_log` 的审核结果字段。

**（10）实现 SubmitReview RPC**

`rpc/internal/logic/submitreviewlogic.go` 从 stub 改为：

```
1. 查询 mod_audit_log（校验 audit_log_id 存在 + review_status=0）
2. UPDATE mod_audit_log SET review_status/reviewer_id/review_time/review_notes
3. 回调业务服务更新内容表 moderation_status (3=人审通过/4=人审不通过)
4. 返回结果
```

**（11）新增人工审核 REST API**

```
GET  /api/moderation/review/list    — 人审列表（分页+筛选）
GET  /api/moderation/review/detail  — 人审详情（内容正文+附件+审核历史）
POST /api/moderation/review/submit  — 提交人审决定
```

API 设计：

**GET /api/moderation/review/list**

请求参数：
| 参数 | 类型 | 说明 |
|------|------|------|
| source_type | string? | 板块筛选，空=全部 |
| review_status | int? | 0=未审核, 1=已通过, 2=已不通过 |
| page | int | 页码，默认 1 |
| page_size | int | 每页条数，默认 20 |

响应：
```json
{
  "code": 0,
  "data": {
    "list": [{
      "id": "audit_log_id",
      "source_type": "notice",
      "source_id": "123",
      "content_summary": "前100字摘要",
      "risk_level": "high",
      "pass": false,
      "review_status": 0,
      "created_time": "2026-06-17T10:00:00Z"
    }],
    "total": 50,
    "page": 1,
    "page_size": 20
  }
}
```

**GET /api/moderation/review/detail**

请求参数：`?id=<audit_log_id>`

响应包含：内容完整正文（从业务服务查询）、附件URL列表、审核管线各层命中详情（matched_items JSON）、机器审核结论。

**POST /api/moderation/review/submit**

请求体：
```json
{
  "audit_log_id": "123",
  "review_status": 1,
  "review_notes": "内容合规"
}
```

### 3.4 图片审核预留

- Redis 任务中 `type=image` 时调用 `ImageEngine.Check()`（已有骨架）
- `CheckImage` RPC 已实现，管线暂未接入生产，预留接口供后续完善
- 图片审核结果合并到整体审核结论：任一图片不通过 → 整体不通过

---

## 四、前端修改

### 4.1 新增人工审核界面

**路由**：`/moderation/review`，菜单名"人工审核"，归属"内容审核"菜单组。

**筛选区**（顶部）：

```
┌──────────────────────────────────────────────────────────────┐
│  板块：[全部 ▼] [通知公告] [房主认证] [寻失互助]              │
│  审核状态：○ 全部  ● 未审核  ○ 已通过  ○ 已不通过            │
│  [搜索]                                                      │
└──────────────────────────────────────────────────────────────┘
```

- 板块默认"全部"（不传 source_type）
- 审核状态单选（未审核/已通过/已不通过），不选时默认"未审核"
- 可选时间范围筛选

**审核列表**：

```
┌────┬──────────┬──────────┬──────────┬──────┬────────┐
│ ID │ 板块     │ 内容摘要 │ 风险等级 │ 状态 │ 操作   │
├────┼──────────┼──────────┼──────────┼──────┼────────┤
│ 01 │ 通知公告 │ xxx小区… │ 高       │ 待审 │ [详情] │
│ 02 │ 房主认证 │ 张三…    │ 中       │ 待审 │ [详情] │
│ 03 │ 寻失互助 │ 丢失…    │ 低       │ 待审 │ [详情] │
└────┴──────────┴──────────┴──────────┴──────┴────────┘
                         [分页器]
```

**详情抽屉**（点击"详情"打开右侧抽屉）：

展示：
- 内容完整正文
- 附件/图片预览（如有）
- 机器审核结果（管线各层命中详情：AC引擎命中词、大模型分析结论、置信度）
- 审核历史时间线
- 底部操作按钮：**[通过]** **[不通过]** + 审核备注输入框

### 4.2 内容列表增加审核状态列

现有内容管理页面（通知列表、寻失列表、认证列表）增加"审核状态"列：

| 状态值 | 显示 | 标签色 |
|--------|------|--------|
| 0 | 待审核 | 灰色 |
| 1 | 机器通过 | 绿色 |
| 2 | 机器不通过 | 红色 |
| 3 | 人审通过 | 蓝色 |
| 4 | 人审不通过 | 红色 |

### 4.3 前端文件变更清单

| 操作 | 文件 | 说明 |
|------|------|------|
| 新增 | `views/moderation/ManualReview.vue` | 人工审核主页面 |
| 新增 | `components/moderation/ReviewFilter.vue` | 顶部筛选面板 |
| 新增 | `components/moderation/ReviewTable.vue` | 审核列表表格 |
| 新增 | `components/moderation/ReviewDetailDrawer.vue` | 详情抽屉（内容+审核历史+操作） |
| 修改 | `api/moderation.ts` | 新增 `listReview`/`detailReview`/`submitReview` API |
| 修改 | `config/modules/moderation.config.ts` | 新增菜单项+路由 `/moderation/review` |
| 修改 | `common/types/moderation.d.ts` | 新增人审相关 TypeScript 类型 |
| 视情况修改 | 各内容列表页 | 增加审核状态列（属于各模块子 Claude 范围） |

---

## 五、实施顺序

```
Phase 1: 基础设施
  ├── 1.1 各内容表 DDL 变更（4 张表加 moderation_status/moderation_time 列）
  ├── 1.2 moderation-service Redis 消费者实现
  └── 1.3 Proto 新增 UpdateModerationStatus RPC + SubmitReview 实现

Phase 2: 业务接入
  ├── 2.1 community-hub-service 通知发布/编辑接入审核
  ├── 2.2 community-hub-service 寻失发布接入审核
  ├── 2.3 user-service 昵称接入审核
  └── 2.4 user-service 认证材料接入审核

Phase 3: 前端
  ├── 3.1 人工审核界面（新页面）
  ├── 3.2 审核 REST API（list/detail/submit）
  └── 3.3 内容列表审核状态列

Phase 4: 收尾
  ├── 4.1 图片审核接口预留
  └── 4.2 端到端验证
```

---

## 六、风险与注意事项

| 风险 | 缓解措施 |
|------|---------|
| Redis 宕机导致审核任务丢失 | 消费者启动时检查残留 pending 记录补偿；任务带 task_id 可去重 |
| 大模型超时（2-5s） | 消费者独立 context 超时（60s），超时标记 need_review=true 入人审 |
| 业务服务回调失败 | 重试 3 次，最终失败写日志+告警，人工兜底 |
| 现有通知/寻失历史数据无审核状态 | migration 默认值 0（待审核），存量数据统一标记为 1（机器通过）或保持 0 |
| gRPC 4MB 上限（已知坑） | 审核内容控制在 10000 字符内，图片不通过 gRPC 传（传 URL） |
