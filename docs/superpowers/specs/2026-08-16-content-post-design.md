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
6. **本期 submit 即隐式通过（status=approved + published_at=NOW()）**（审核消费者后期开发，本期功能先跑通）

## 二、数据模型

### 2.1 content_posts（通用发布主表，原 notices 迁入）

```sql
content_posts
├── id              BIGINT PK       -- 一次发布唯一 ID（雪花）
├── section_code    VARCHAR(30)     -- 板块：notice=通知 / repair=维修保修 / ...
├── text            TEXT            -- 一段文字（图文发布的核心）
├── publisher_id    BIGINT          -- 发布者（JWT 派生）
├── publisher       VARCHAR(100)    -- 展示名（取用户真实档案，禁请求体信任 — REVISION 堵伪造向量）
├── role            VARCHAR(20)     -- 发布角色（RBAC→发布角色映射派生）
├── is_pinned       TINYINT DEFAULT 0  -- 置顶（跑马灯 order by is_pinned desc；UpdateContentPost 置位）
├── status          TINYINT         -- 全生命周期+审核结果：0=draft 1=submitted 2=approved 3=rejected 4=withdrawn；本期 submit 即隐式通过置 approved（无消费者）
├── attachment_count INT            -- 附件计数（审核完整性判定：已审附件数 == 此数 → 审核完整；每次附件集合变更同事务重算，提交时冻结）
├── published_at    DATETIME NULL   -- 审核锚定：本期 submit 即置 NOW()（隐式通过）；消费者上线后按审核结果覆盖（D27）
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

**审核完整性判定**：`count(attachments WHERE review_status=approved) == content_posts.attachment_count` 且 正文 status=approved → 整体可展示；任一附件 rejected → 整体不展示。**语义澄清（REVISION）：不展示由读路径完整性谓词隐藏，读路径不 mutate content_posts.status；status=rejected 仅由审核流（后期消费者）回写。**

### 3.2 Kafka 审核链路（新增）

```
发布 CreateContentPost（entry draft/submitted；submitted 即触发推送）
  → 打包 JSON：{ version, post_id, section_code, text, publisher_id, attachments:[{file_id,file_type,review_status,file_url}] }
     （契约单源 REQ-CPM-2；file_url 为可再生预签名 URL，D7）
  → 推 Kafka topic: content-review（at-least-once：落库待推标记 + 定时重推，D20）
  → moderation-service 扩展消费者（后期开发，本期只定契约+推送）
       ├── text：先关键字过滤 → 再大模型审核
       ├── image/pdf：走大模型接口（图片内容/文档内容审查）
       └── 结果回写：正文→content_posts.status，附件→review_status
  → 全部通过 → 前端可见
```

- **Kafka 安装**：docker-compose 新增 kafka（单节点 KRaft 模式 + 数据卷持久化，D8/Q8），本期一并安装
- **现有 Redis List（moderation:task:queue）**：对 content_posts 停推（D3），lostfound/user 等其他来源仍走 Redis；队列机制保留，不做物理清理

### 3.3 本期范围（审核消费者后期开发）

- 本期 submit 即隐式通过：`status=approved(2)` + `published_at=NOW()`（无消费者也可见，REVISION——替代「status 默认 approved」含混表述）；附件 `review_status` 默认 `approved`
- Kafka 推送 + 契约本期实现（at-least-once 待推标记 + 定时重推）；**消费者程序后期单独开发**
- 本期验收：发布→推送 Kafka 消息→（无消费者）→内容直接可见

## 四、接口契约

### 4.1 Proto（community/v1）

- `CreateNoticeRequest` → 通用化 `CreateContentPostRequest`（或保留 CreateNotice 名但加 section_code）
  - `section_code` 新增
  - `text`（原 content）、`community_ids`（多小区；社区管理员无 division_id 入参，后端自动展开其唯一管辖社区下所有通过小区，A2）
  - `attachment_ids`（文件 ID）
- 响应/列表/详情：`ContentPost` + `attachments[]`（含 review_status）
- `GetPublishPermission` / `GetMarqueeNotices` 保留（通知板块用）

### 4.2 Kafka 消息契约（新，content-review topic；单源 REQ-CPM-2，REVISION 同步）

```json
{
  "version": 1,
  "post_id": 123,
  "section_code": "notice",
  "text": "正文内容",
  "publisher_id": 456,
  "attachments": [
    {"file_id": 789, "file_type": "pdf", "review_status": 0, "file_url": "https://.../presigned?x-id=..."}
  ]
}
```

- `version`：契约版本（变更即 bump，供消费者协商）
- `file_url`：可再生预签名 URL（消费者直接拉取附件内容；经 file_id 经 GetFileUrl 再生，非永久链接依赖，D7）
- `attachments[].review_status`：推送时刻为审核前默认值（本期 approved），非审核结论

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
