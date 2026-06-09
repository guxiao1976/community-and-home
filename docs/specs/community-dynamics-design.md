# 小区动态 — 完整前后端设计方案

> **状态：待审核** | 2026-06-06

## 一、总体架构

```
┌─────────────────────────────────────────────────┐
│              Uni-app 移动端 (web/mobile/)         │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐            │
│  │公告信息│ │邻里互动│ │我的家庭│ │ 我的  │  4 Tab   │
│  └──┬───┘ └──────┘ └──────┘ └──────┘            │
│     │                                            │
│     ├─ 通知公告 (Notice)                          │
│     ├─ 便民联络 (Contact)      → REST API         │
│     └─ 寻失互助 (LostFound)                       │
└─────────────────────┬───────────────────────────┘
                      │ gRPC
┌─────────────────────▼───────────────────────────┐
│        community-dynamics-service (新建)          │
│        github.com/guxiao1976/community-dynamics   │
│                                                   │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐          │
│  │  Notice  │ │ Contact  │ │LostFound │  gRPC    │
│  │  Service │ │ Service  │ │ Service  │  Server   │
│  └────┬─────┘ └────┬─────┘ └────┬─────┘          │
│       └────────────┼────────────┘                 │
│              ┌─────▼─────┐                        │
│              │   MySQL   │  community_dynamics_db │
│              └───────────┘                        │
└─────────────────────────────────────────────────┘
```

### 新服务命名

| 项 | 值 |
|---|-----|
| 中文名 | 小区动态服务 |
| 目录 | `services/community-dynamics-service/` |
| Go Module | `github.com/guxiao1976/community-dynamics` |
| Proto | `api-proto/api/community/v1/community.proto` |
| 端口 | gRPC 8087 / API 8887（待确认未占用端口） |
| 数据库 | `community_dynamics_db` |

### 是否应该新建服务？

| 理由 | 说明 |
|------|------|
| ✅ 独立业务域 | 通知发布、便民联络、寻失互助是社区特有的内容运营场景，与现有服务（用户/认证/主数据）无重叠 |
| ✅ 独立扩展 | 通知和寻失互助涉及 UGC 内容，后续可能接入审核、推送，独立服务不影响其他服务 |
| ✅ 数据隔离 | 社区动态内容独立数据库，与核心业务数据物理隔离 |
| ❌ 运维成本 | 多一个服务需要维护（可接受，与其他 7 个服务一致） |

**结论：新建 `community-dynamics-service` 是合理的。**

---

## 二、移动端页面设计

### TabBar 重构

```json
// pages.json tabBar.list
[
  { "pagePath": "pages/notice/notice",   "text": "公告信息" },
  { "pagePath": "pages/interact/interact", "text": "邻里互动" },
  { "pagePath": "pages/family/family",   "text": "我的家庭" },
  { "pagePath": "pages/mine/mine",       "text": "我的"     }
]
```

> 当前标签页 tabBar 最多 5 个，4 个在限制之内。

### 公告信息页 (`pages/notice/notice.vue`) — 核心页面

```
┌──────────────────────────────────────┐
│  公告信息                             │  ← 页面标题
├──────────────────────────────────────┤
│                                      │
│  ┌── 通知公告 ─────────── [更多 ▸] ──┐│
│  │                                   ││
│  │ 📢 小区停水通知  [物业]            ││  ← 角色标签
│  │    2026-06-06 14:30               ││
│  │ ─────────────────────────────    ││
│  │ 📢 业主大会通知  [业委会]          ││
│  │    2026-06-05 10:00               ││
│  │ ─────────────────────────────    ││
│  │ 📢 网格员巡查通知  [网格员]        ││
│  │    2026-06-04 09:00               ││
│  │                                   ││
│  │  最多 3 条，不足显示"暂无通知"      ││
│  └───────────────────────────────────┘│
│                                      │
│  ┌── 便民联络 ───────────────────────┐│
│  │                                   ││
│  │ 供水维修    📞 12345678           ││
│  │             📞 87654321           ││
│  │ 电力维修    📞 95598              ││
│  │ 燃气维修    📞 96777              ││
│  │ 联通网络    📞 10010              ││
│  │ 移动网络    📞 10086              ││
│  │ 电信网络    📞 10000              ││
│  │ 小区民警    📞 110 (转) xxx警官   ││
│  │                                   ││
│  │  每项 1-2 个号码，点击拨号          ││
│  └───────────────────────────────────┘│
│                                      │
│  ┌── 寻失互助 ───────────────────────┐│
│  │                                   ││
│  │ ┌────┐  ┌────┐  ┌────┐           ││
│  │ │图片│  │图片│  │图片│           ││
│  │ │寻物│  │招领│  │寻物│           ││
│  │ │钥匙│  │手机│  │钱包│           ││
│  │ └────┘  └────┘  └────┘           ││
│  │                                   ││
│  │  最新 3 条，图片+简短文字           ││
│  └───────────────────────────────────┘│
│                                      │
└──────────────────────────────────────┘
```

### 通知详情页 (`pages/notice-detail/notice-detail.vue`)

```
┌──────────────────────────────────────┐
│  ← 返回    通知详情                   │
├──────────────────────────────────────┤
│                                      │
│  小区停水通知  [物业]                  │  ← 标题 + 角色标签
│                                      │
│  发布时间：2026-06-06 14:30           │
│  发布单位：XX 小区物业管理处           │
│                                      │
│  ──────────────────────────────────  │
│                                      │
│  各位业主：                           │
│                                       │
│  因管道维修需要，将于 6 月 8 日        │
│  8:00-18:00 停水...                   │  ← 正文
│                                      │
│  ──────────────────────────────────  │
│                                      │
│  附件：                               │
│  📎 停水通知.pdf                   │  ← 可点击下载
│  📎 影响范围图.jpg                  │
│                                      │
└──────────────────────────────────────┘
```

---

## 三、数据模型

### 表结构（MySQL `community_dynamics_db`）

#### `notices` — 通知公告

```sql
CREATE TABLE notices (
    id          BIGINT PRIMARY KEY COMMENT 'Snowflake ID',
    community_id BIGINT NOT NULL COMMENT '所属小区ID',
    title       VARCHAR(200) NOT NULL COMMENT '通知标题',
    content     TEXT NOT NULL COMMENT '正文（富文本）',
    role        VARCHAR(20) NOT NULL COMMENT '发布角色: community/committee/property/grid_officer',
    publisher   VARCHAR(100) NOT NULL COMMENT '发布单位/人名称',
    publisher_id BIGINT DEFAULT NULL COMMENT '发布人用户ID（可选）',
    is_pinned   TINYINT DEFAULT 0 COMMENT '是否置顶',
    published_at DATETIME NOT NULL COMMENT '发布时间',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at  DATETIME DEFAULT NULL,

    INDEX idx_community (community_id, deleted_at),
    INDEX idx_role (community_id, role, deleted_at),
    INDEX idx_published (community_id, published_at DESC, deleted_at)
);
```

#### `notice_attachments` — 通知附件

```sql
CREATE TABLE notice_attachments (
    id          BIGINT PRIMARY KEY COMMENT 'Snowflake ID',
    notice_id   BIGINT NOT NULL COMMENT '关联通知ID',
    file_name   VARCHAR(200) NOT NULL COMMENT '文件名',
    file_url    VARCHAR(500) NOT NULL COMMENT 'MinIO 存储路径',
    file_size   BIGINT DEFAULT 0 COMMENT '文件大小（字节）',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_notice (notice_id)
);
```

#### `community_contacts` — 便民联络

```sql
CREATE TABLE community_contacts (
    id          BIGINT PRIMARY KEY COMMENT 'Snowflake ID',
    community_id BIGINT NOT NULL COMMENT '所属小区ID',
    category    VARCHAR(30) NOT NULL COMMENT '类别: water/electricity/gas/unicom/mobile/telecom/police',
    name        VARCHAR(100) NOT NULL COMMENT '联络名称',
    phone       VARCHAR(20) NOT NULL COMMENT '电话号码',
    sort_order  INT DEFAULT 0 COMMENT '排序',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_community (community_id)
);
```

#### `lost_found_items` — 寻失互助

```sql
CREATE TABLE lost_found_items (
    id            BIGINT PRIMARY KEY COMMENT 'Snowflake ID',
    community_id  BIGINT NOT NULL COMMENT '所属小区ID',
    type          VARCHAR(10) NOT NULL COMMENT '类型: lost=寻物, found=招领',
    title         VARCHAR(200) NOT NULL COMMENT '标题',
    description   TEXT COMMENT '详细描述',
    image_urls    JSON DEFAULT NULL COMMENT '图片URL数组 ["url1","url2"]',
    contact_phone VARCHAR(20) COMMENT '联系电话',
    status        VARCHAR(20) DEFAULT 'active' COMMENT '状态: active=进行中, resolved=已解决',
    publisher_id  BIGINT NOT NULL COMMENT '发布人用户ID',
    created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at    DATETIME DEFAULT NULL,

    INDEX idx_community_type (community_id, type, status, deleted_at),
    INDEX idx_created (community_id, created_at DESC)
);
```

### 角色枚举

```go
// 通知发布角色
const (
    RoleCommunity    = "community"     // 社区
    RoleCommittee    = "committee"     // 业委会
    RoleProperty     = "property"      // 物业
    RoleGridOfficer  = "grid_officer"  // 网格员
)
```

### 联络类别枚举

```go
const (
    ContactWater      = "water"       // 供水维修
    ContactElectricity = "electricity" // 电力维修
    ContactGas        = "gas"          // 燃气维修
    ContactUnicom     = "unicom"       // 联通网络
    ContactMobile     = "mobile"       // 移动网络
    ContactTelecom    = "telecom"      // 电信网络
    ContactPolice     = "police"       // 小区民警
)
```

---

## 四、gRPC 接口设计

### Proto 文件：`api-proto/api/community/v1/community.proto`

```protobuf
syntax = "proto3";
package api.community.v1;

import "api/common/v1/common.proto";
import "buf/validate/validate.proto";

// ==================== Notice ====================

service NoticeService {
  // 发布通知
  rpc CreateNotice(CreateNoticeRequest) returns (CreateNoticeResponse);
  // 通知列表（含分页）
  rpc ListNotices(ListNoticesRequest) returns (ListNoticesResponse);
  // 通知详情（含附件）
  rpc GetNotice(GetNoticeRequest) returns (GetNoticeResponse);
  // 更新通知
  rpc UpdateNotice(UpdateNoticeRequest) returns (UpdateNoticeResponse);
  // 删除通知（软删除）
  rpc DeleteNotice(DeleteNoticeRequest) returns (DeleteNoticeResponse);
}

message Notice {
  int64  id          = 1 [jstype = JS_STRING];
  int64  community_id = 2 [jstype = JS_STRING];
  string title       = 3;
  string content     = 4;
  string role        = 5;  // community/committee/property/grid_officer
  string publisher   = 6;
  int64  publisher_id = 7 [jstype = JS_STRING];
  bool   is_pinned   = 8;
  int64  published_at = 9;
  int64  created_at   = 10;
  repeated Attachment attachments = 11;
}

message Attachment {
  int64  id        = 1 [jstype = JS_STRING];
  string file_name = 2;
  string file_url  = 3;
  int64  file_size = 4;
}

message CreateNoticeRequest { ... }
message CreateNoticeResponse {
  api.common.v1.BaseResp base = 1;
  int64 id = 2 [jstype = JS_STRING];
}

message ListNoticesRequest {
  int64  community_id = 1 [jstype = JS_STRING];
  string role         = 2;  // 可选筛选
  int32  page         = 3;
  int32  page_size    = 4;
}

message ListNoticesResponse {
  api.common.v1.BaseResp base = 1;
  repeated Notice notices     = 2;
  int64 total                 = 3 [jstype = JS_STRING];
}

message GetNoticeRequest  { int64 id = 1 [jstype = JS_STRING]; }
message GetNoticeResponse {
  api.common.v1.BaseResp base = 1;
  Notice notice               = 2;
}
// ... UpdateNotice, DeleteNotice 类似

// ==================== Contact ====================

service ContactService {
  // 获取小区便民联络列表
  rpc ListContacts(ListContactsRequest) returns (ListContactsResponse);
  // 管理员维护联络信息（批量增/改/删）
  rpc UpsertContacts(UpsertContactsRequest) returns (UpsertContactsResponse);
}

message Contact {
  int64  id          = 1 [jstype = JS_STRING];
  string category    = 2;  // water/electricity/gas/...
  string name        = 3;
  string phone       = 4;
  int32  sort_order  = 5;
}

message ListContactsRequest  { int64 community_id = 1 [jstype = JS_STRING]; }
message ListContactsResponse {
  api.common.v1.BaseResp base = 1;
  repeated Contact contacts   = 2;
}
// ... UpsertContacts 略

// ==================== Lost & Found ====================

service LostFoundService {
  rpc CreateLostFound(CreateLostFoundRequest) returns (CreateLostFoundResponse);
  rpc ListLostFound(ListLostFoundRequest)   returns (ListLostFoundResponse);
  rpc GetLostFound(GetLostFoundRequest)     returns (GetLostFoundResponse);
  rpc ResolveLostFound(ResolveLostFoundRequest) returns (ResolveLostFoundResponse);
}

message LostFoundItem {
  int64  id            = 1 [jstype = JS_STRING];
  int64  community_id  = 2 [jstype = JS_STRING];
  string type          = 3;  // lost / found
  string title         = 4;
  string description   = 5;
  repeated string image_urls = 6;
  string contact_phone = 7;
  string status        = 8;  // active / resolved
  int64  publisher_id  = 9 [jstype = JS_STRING];
  int64  created_at    = 10;
}

message ListLostFoundRequest {
  int64 community_id = 1 [jstype = JS_STRING];
  string type        = 2;  // 可选: lost / found
  int32  page        = 3;
  int32  page_size   = 4;
}

message ListLostFoundResponse {
  api.common.v1.BaseResp base = 1;
  repeated LostFoundItem items = 2;
  int64 total = 3 [jstype = JS_STRING];
}
// ... CreateLostFound, GetLostFound, ResolveLostFound 略
```

> 仅列出核心 message，完整 Request/Response 结构在编码阶段补充。

---

## 五、服务架构

### 目录结构

```
services/community-dynamics-service/
├── api/                          # REST 网关（代理 → gRPC）
│   ├── communitydynamics.go
│   └── internal/
│       ├── config/config.go
│       ├── svc/servicecontext.go
│       ├── handler/routes.go
│       ├── logic/                # notice/, contact/, lostfound/
│       └── types/types.go
├── rpc/                          # gRPC 服务
│   ├── communitydynamics.go
│   └── internal/
│       ├── config/config.go
│       ├── svc/servicecontext.go
│       ├── server/               # gRPC Server 实现
│       └── logic/                # notice/, contact/, lostfound/
├── model/                        # GORM 模型
│   ├── notice.go
│   ├── notice_attachment.go
│   ├── community_contact.go
│   └── lost_found_item.go
├── docs/design.md
├── CHANGELOG.md
└── CLAUDE.md
```

### 服务依赖

```
community-dynamics-service
  ├─ 依赖: api-proto (community/v1, common/v1)
  ├─ 依赖: community-common/v2 (configx, responsex, snowflake)
  ├─ 调用方: 无（独立服务，被前端直接调用）
  └─ 数据库: MySQL community_dynamics_db
```

### 端口分配

| 层 | 端口 | etcd Key |
|----|------|----------|
| API (REST) | 8887 | —（不注册 etcd） |
| RPC (gRPC) | 8087 | `community-dynamics.rpc` |

---

## 六、前端移动端变更

### 新增页面

| 页面路径 | 说明 |
|----------|------|
| `pages/notice/notice.vue` | 公告信息页（通知+联络+寻失） |
| `pages/notice-detail/notice-detail.vue` | 通知详情页 |
| `pages/interact/interact.vue` | 邻里互动（占位） |
| `pages/family/family.vue` | 我的家庭（占位） |

### 新增 API 模块

```
src/api/community.ts              # 小区动态 API
  getNoticeList(communityId)      → 通知列表（前3条）
  getMoreNotices(communityId)     → 通知列表（更多）
  getNoticeDetail(id)             → 通知详情+附件
  getContacts(communityId)        → 便民联络列表
  getLostFoundList(communityId)   → 寻失互助（前3条）
```

### 变更文件

| 文件 | 变更 |
|------|------|
| `src/pages.json` | TabBar: 3 → 4 个，新增 notice/interact/family 路由 |
| `src/pages/mine/mine.vue` | 不变（已在第4位） |
| `src/pages/index/index.vue` | 删除（被 notice 替代） |
| `src/pages/discover/discover.vue` | 删除（被 interact 替代） |

---

## 七、待确认问题

1. **第四个 Tab 是"我的"还是保留当前设计？** — 当前设计中 4 个 Tab 是 公告信息 / 邻里互动 / 我的家庭 / 我的。"邻里互动"和"我的家庭"本期只做占位。

2. **便民联络数据由谁维护？** — 建议由后端管理后台维护（PC 端加一个管理页面），移动端只读。每小区一套联络数据。

3. **寻失互助是否需要审核？** — 建议发布后先进入待审核状态，由 moderation-service 审核通过后展示，避免恶意内容。或者本期先不做审核，直接发布。

4. **用户如何归属于某个小区？** — 当前 User 模型有 `community_id` 吗？需要确认。如果没有，需要通过其他方式获取当前用户的社区 ID（可能从 User 的 scope 或者 association 表）。

5. **通知发布权限** — 只有特定角色（社区/网格员/业委会/物业）可以发布通知。这个权限如何校验？（本期可以先通过 PC 管理后台发布，移动端只读。）

---

## 八、开发顺序

```
Phase 1: 后端基础
  1. 创建 api-proto 定义 (community/v1)
  2. 搭建 community-dynamics-service 脚手架
  3. 实现 Notice CRUD（gRPC + REST）
  4. 实现 Contact 列表（gRPC + REST）
  5. 实现 LostFound CRUD（gRPC + REST）

Phase 2: 前端
  6. TabBar 四按钮改造
  7. 公告信息页（通知+联络+寻失）
  8. 通知详情页

Phase 3: 完善
  9. PC 管理后台：通知发布
  10. PC 管理后台：便民联络维护
  11. 邻里互动 / 我的家庭（后续迭代）
```
