# 通用图文发布组件设计（content_posts + 内容级审核 + Kafka）

> 日期：2026-08-16
> 范围：把「通知模块」升级为通用图文发布组件（content_posts），支持未来板块（通知/维修保修等），引入 Kafka 内容审核链路 + 内容级审核（正文+附件各自状态）
> 前置：B 方案（用户拍板）——暂停通知模块流水线，重构为通用化一次到位
> 背景：原通知模块设计见 `2026-08-15-notice-module-design.md`（通知专属 notices 表），本设计将其通用化为 content_posts

## 一、目标

1. **通用图文发布**：一次发布一个 `content_posts.id`，下辖一段文字 + 若干附件（图片/pdf/word），对应 `section_code` 板块
2. **内容级审核**：正文 + 每个附件各自审核状态，**全部通过才统一展示**（一个有问题则不展示）
3. **Kafka 审核链路**：发布时打包 JSON 推 Kafka，moderation-service 扩展消费者（文字先关键字后大模型，图片/pdf 走大模型）
4. **审核完整性判定**：主表 `attachment_count` 计数，已审核附件数 == 计数 → 审核完整
5. **多小区归属**：复用 scope 关联表模式（post_id + community_id）
6. **本期 status 默认 approved**（审核消费者后期开发，本期功能先跑通）

## 二、数据模型

### 2.1 content_posts（通用发布主表，原 notices 迁入）

```sql
content_posts
├── id              BIGINT PK       -- 一次发布唯一 ID（雪花）
├── section_code    VARCHAR(30)     -- 板块：notice=通知 / repair=维修保修 / ...
├── text            TEXT            -- 一段文字（图文发布的核心）
├── publisher_id    BIGINT          -- 发布者（JWT 派生）
├── publisher       VARCHAR(100)    -- 展示名（请求体）
├── role            VARCHAR(20)     -- 发布角色
├── status          TINYINT         -- 整体审核状态：0=pending 1=approved 2=rejected 3=withdrawn；本期默认 approved
├── attachment_count INT            -- 附件计数（审核完整性判定：已审附件数 == 此数 → 审核完整）
├── published_at    DATETIME NULL   -- 审核通过时设置（D27）
├── moderation_status / moderation_time  -- 兼容期保留（原审核列，逐步过渡到 status+附件级）
├── created_at / updated_at / deleted_at
```

**迁移**：`ALTER TABLE notices RENAME TO content_posts` + 加 section_code/attachment_count/status + 弃用 community_id（走 scope 关联）+ published_at 去 NOT NULL

### 2.2 content_post_scope（多小区关联，复用 notice_scope 模式）

```sql
content_post_scope
├── post_id      BIGINT NOT NULL   -- → content_posts.id
├── community_id BIGINT NOT NULL   -- → md_residential_area.id（小区或村）
├── created_at
└── PK (post_id, community_id) + KEY idx_scope_community (community_id, post_id)
```

### 2.3 content_post_attachments（附件，含独立审核状态）

```sql
content_post_attachments
├── id          BIGINT PK
├── post_id     BIGINT NOT NULL    -- 即用户说的 main_id（关联 content_posts）
├── file_id     BIGINT             -- 重生预签名 URL 载体（S4）
├── file_name / file_url / file_size
├── file_type   VARCHAR(20)        -- 白名单类型
├── review_status TINYINT          -- 附件级审核：0=pending 1=approved 2=rejected；默认 approved（本期）
└── created_at
```

## 三、审核机制（内容级 + Kafka）

### 3.1 审核状态模型

```
content_posts.status          -- 整体（正文审核结果）
content_post_attachments.review_status -- 每个附件各自审核
```

**审核完整性判定**：`count(attachments WHERE review_status=approved) == content_posts.attachment_count` 且 正文 status=approved → 整体可展示；任一附件 rejected → 整体不展示（status=rejected）

### 3.2 Kafka 审核链路（新增）

```
发布 CreateContentPost
  → 打包 JSON：{ post_id, text, attachment_ids:[{file_id,file_type}], section_code }
  → 推 Kafka topic: content-review
  → moderation-service 扩展消费者（后期开发，本期只定契约+推送）
       ├── text：先关键字过滤 → 再大模型审核
       ├── image/pdf：走大模型接口（图片内容/文档内容审查）
       └── 结果回写：正文→content_posts.status，附件→review_status
  → 全部通过 → 前端可见
```

- **Kafka 安装**：docker-compose 新增 kafka（+zookeeper 或 KRaft 模式），本期一并安装
- **现有 Redis List（moderation:task:queue）**：过渡期保留，正式切 Kafka 后移除

### 3.3 本期范围（审核消费者后期开发）

- 本期 `status` 默认 `approved`、附件 `review_status` 默认 `approved`（功能先跑通，无消费者也可见）
- Kafka 推送 + 契约本期实现；**消费者程序后期单独开发**
- 本期验收：发布→推送 Kafka 消息→（无消费者）→内容直接可见

## 四、接口契约

### 4.1 Proto（community/v1）

- `CreateNoticeRequest` → 通用化 `CreateContentPostRequest`（或保留 CreateNotice 名但加 section_code）
  - `section_code` 新增
  - `text`（原 content）、`community_ids`/`division_id`（多小区，D23-D29 已设计）
  - `attachment_ids`（文件 ID）
- 响应/列表/详情：`ContentPost` + `attachments[]`（含 review_status）
- `GetPublishPermission` / `GetMarqueeNotices` 保留（通知板块用）

### 4.2 Kafka 消息契约（新，content-review topic）

```json
{
  "post_id": 123,
  "section_code": "notice",
  "text": "正文内容",
  "publisher_id": 456,
  "attachments": [
    {"file_id": 789, "file_type": "pdf", "review_status": 0}
  ]
}
```

## 五、服务归属

| 能力 | 服务 |
|---|---|
| content_posts 数据模型 + CRUD + 多小区 scope + 附件绑定 + Kafka 推送 | community-hub-service |
| Kafka 消费者 + 关键字/大模型审核 + 结果回写 | moderation-service（扩展） |
| 附件白名单/大小/magic-bytes | file-service |
| 权限（发布角色 scope + can_publish） | permission-service |
| 小区/板块展开 | master-data-service |
| 通用组件（ContentPostList/Detail/Publish） | web/mobile |

## 六、可复用组件

| 组件 | 职责 |
|---|---|
| `ContentPostPublish` | 通用发布（文字+附件+板块+范围+审核状态展示） |
| `ContentPostList` | 按板块列表 |
| `ContentPostDetail` | 详情+附件（含审核状态） |
| 附件安全校验器 | file-service 白名单+大小+magic-bytes（已设计） |

## 七、验收标准

- [ ] content_posts 通用化（notices 迁入 + section_code + attachment_count + status）
- [ ] 多小区 scope 关联（网格员多选/社区管理员选社区展开）
- [ ] 内容级审核：正文 status + 附件 review_status 各自独立
- [ ] 审核完整性：已审附件数 == attachment_count 才可展示；任一 rejected 不展示
- [ ] Kafka 推送：发布打包 JSON 推 content-review topic
- [ ] 本期 status 默认 approved（消费者后期开发）
- [ ] 附件白名单/大小/magic-bytes（复用已设计）
- [ ] 板块化：notice 为首个板块，未来 repair 等
