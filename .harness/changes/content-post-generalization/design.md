# Design: 通用图文发布组件重构（content_posts 通用化 + 内容级审核 + Kafka）

> 变更名: `content-post-generalization` · 规模 L · 优先级 P1 · 类型 modify
> 输入: proposal.md + 5 个 spec（content-post-publish / content-post-read / content-post-moderation / content-post-permission / content-post-attachment-security）
> 权威参考: `docs/superpowers/specs/2026-08-16-content-post-design.md`（用户已拍板）+ 存档 `notice-multicommunity-publish` 设计（复用 D1-D32，推翻 D26）+ 决策 D1-D22
> 说明: 本变更取代/存档 `notice-multicommunity-publish`（B 方案，通知流水线暂停，通用化一次到位）
>
> **REVISION v3（本轮，对照设计评审 data-model v2 / interface-proto v2 逐条修订）**：
> - **R1 [MUST] Design Gate 前提 grounded**：`community_div` scope_type 在权限数据模型**不存在**（rel.go:77-82 仅 global/""/community/building/unit/grid；apply_role_logic.go:15,66 将 community_admin 绑定为 scope_type='community'、scope_id=communityId；UserRoles.vue scopeType 无 division；rbac-design.md:38 scope 域无 division；全仓 community_div 仅作 md_residential_area.community_div_id 列）→ 按评审推荐**方案 2**：`ResolveAdminDivision` 改经既有 `scope_type='community'` grant 派生管辖 division（`GetResidentialArea(communityId).community_div_id`）；permission-service `resolvePublishScope` 改为**社区管理员角色感知展开**（社区管理员 community grant 展开为其 division 下全部 approved 小区子树，owner/tenant/committee/property_admin/grid_worker 语义不变）——不再依赖不存在的 community_div grant。§Design Gate 与 Task 1.7/3.1 相应重写。
> - **R2 [MUST] 破坏面评估修正 + REST wire 兼容决策**：如实登记 web/mobile 三处活跃通知消费方（pages.json tabbar `pages/notice/notice` + notice-detail/notice-browse 注册页；api/community.ts `getNoticeList` 读 `res.notices`、`getNoticeDetail(id)` 读 `res.notice` 且不传 community_id；notice.vue:336-337 / notice-browse.vue:44,110-113 / notice-detail.vue:103 实际调用）→ 采用评审**选项 (a) 保留 JSON wire 标签**：REST 响应 wire 字段名保持 `notices`/`notice`/`content`（Go 类型改名，REST 路径已保持 /notices），RPC/proto/DB 通用化改名 `text`；REST 层映射 proto Text → wire `content`。`GetContentPost` 强制 community_id 与移动端 `getNoticeDetail(id)`（无 community_id）冲突 → **REST 详情兼容回退**（v3 起为 **scope 反查 + 逐小区 FilterAllowed 任一允许即放行**，见下方 REVISION v4 R2 修订）。§Proto 破坏性影响评估重写。
> - **R3 [SHOULD] proto 头注释 080002 语义扩展 + 保留 080004**；`UpdateContentPostRequest.status` 注释标注「action」vs `ContentPost.status`「state」。
> - **R4 [SHOULD] Migration 003 一次性 RENAME 勿重跑**（MySQL 无 `RENAME TABLE IF EXISTS`）；Task 1.1/6.1 注明。
> - **R5 [SHOULD] publisher 真实档案查询接线缺口**：community-hub servicecontext 现无 UserClient（仅 Moderation/Permission/MasterData/SysConfig）→ 新增 Task 1.9 接入 `userv1.UserServiceClient` + `UserRpc` 配置（user-service `GetUsersByIds` 已存在）。
>
> **REVISION v4（本轮，对照设计评审 data-model v3 / interface-proto v3 闭合 design↔tasks 漂移）**：
> - **V4-R2 [MUST] 详情 community_id 兼容回退改 scope 反查（取消 grant 唯一假设）**：R2 原「用户活跃 community grant 唯一 → 使用；多/无 → 080005」对多小区用户（grid_worker 为多小区角色、多房产业主可持多条 community grant）是行为收窄 → 改按 `content_post_scope` 反查该帖所属小区集 + 逐小区 `FilterAllowed` 任一允许即放行（与现网 getnoticelogic.go 反查 LIMITED 语义一致）；RPC 层保持 community_id 严格必填，回退只落 REST 薄代理层。§GetContentPost 与 Task 1.14/1.23 相应重写（`ResolveReadableCommunityForCompat`）。
> - **V4-1 [MUST] UpdateContentPost 授权分流**：消除「作者校验先行使 is_pinned 操作者置顶不可达」——§UpdateContentPost (a) 内容编辑路径先作者校验 080002 / (b) 仅 is_pinned 路径跳过作者校验改验 `PublishRolesFrom` + `AssertCommunitiesScope`。
> - **V4-3 [SHOULD] 080004 头注释标签修正**：`寻失记录不存在（LostFoundService 仍用，CodeLostFoundMiss types.go:19）`——不固化「便民联络不存在/ContactService 仍用」旧标签（R3 措辞更新）。
> - **V4-4 [SHOULD] Kafka Producer.Push 触发点/顺序显式落位**：Task 1.10/1.11「事务提交成功后调用 Producer.Push（先提交后推送，提交失败不推送）」。
>
> **REVISION v5（本轮，对照设计评审 data-model v4 / interface-proto v4 逐条修订）**：
> - **V5-1 [MUST] UpdateContentPost 字段 presence 语义显式化 + is_pinned-only 持久化独立化**（data-model v4 M1 + interface-proto v4 MUST 1）：`title`/`text`/`section_code`/`is_pinned` 改 proto3 `optional`（生成 `*string`/`*bool`，presence 可判，仓内先例 user-service/master-data/file-service）；`community_ids`/`attachment_ids` 为「全量替换集」，以 bool 标志 `has_scope_change`/`has_attachment_change` 区分「未携带=不改」与「空数组=清空」（repeated 无法用 `optional`，评审 MUST 1(b)）——分支判定以 presence/标志位为准，禁止 value 非空启发式（取消置顶 `*false`、清空全部附件 `attachment_count→0` 均可确定实现）；(b) 仅 is_pinned 路径持久化统一走新增 `UpdateIsPinned` 独立列更新（不碰 title/text/section_code），**禁止复用 `UpdateContent` 传空 title/text（会清空已发布帖正文，数据丢失）**；(a) draft 编辑声明内容字段**全量替换语义 + title/text 非空不变量**（presence+空串 → 080005），「仅改附件/scope」时正文列不动。
> - **V5-2 [SHOULD] 详情兼容回退错误码消歧**（data-model v4 S1）：帖有 scope 小区但全部不可读 → **080001**（与 RPC 层 scope 外统一 080001 一致，不泄露）；仅「帖无任何 scope 小区（数据异常）」→ 080005。design §GetContentPost 与 Task 1.14 措辞统一。
> - **V5-3 [SHOULD] Migration 003 部分失败恢复指引**（data-model v4 S2）：003 中途失败先 `RENAME TABLE content_posts TO notices` 对齐状态再修重跑；已 RENAME 但缺新列的中间态**禁止直接重跑完整脚本**。
> - **V5-4 [SHOULD] UpdateContentPostRequest 预期破坏清单补 `content`(4)→`text`(4) 改名**（interface-proto v4 SHOULD 1）：buf WIRE_JSON 类 FIELD_SAME_NAME/FIELD_SAME_JSON_NAME 预期 fail（proto JSON 名改 `text` 属有意，wire 兼容由 REST 层 `json:"content"` 承担）。
> - **V5-5 [INFO] 契约收敛**：Create/Update 请求 `attachment_ids` 统一 `repeated int64 [jstype=JS_STRING]`（对齐 `ContentPostAttachment.file_id`，interface v4 INFO 1）；Task 0.1 显式列齐四个响应消息字段（INFO 2）；Kafka 契约注明「后续消费者按 `post_id` 幂等去重」（并发 submit 双推由 at-least-once 容忍，INFO I3）；role 筛选映射收敛单一来源 helper.go（data-model v4 I2）；`FindListByCommunity` 完整性子查询注明走 `idx_notice(post_id)`（I4）。

---

## 需求追溯矩阵（spec → design）

| 需求 ID | 需求内容摘要 | 对应设计章节 | 覆盖状态 |
|---------|-------------|-------------|:---:|
| REQ-CPB-1 | notices RENAME content_posts + title/text/is_pinned/role/publisher 保留 + section_code/status(0-4)/attachment_count + published_at/community_id 去 NOT NULL + 存量不迁 + status 显式写 | §数据模型 / §迁移 | ✅ |
| REQ-CPB-2 | content_post_scope 双 NOT NULL + 复合 PK + idx_scope_community 读索引 | §数据模型 / §迁移 | ✅ |
| REQ-CPB-3 | content_post_attachments 加 post_id/review_status/file_id/file_type（file_id 重生载体） | §数据模型 / §迁移 | ✅ |
| REQ-CPB-4 | 两阶段状态机（draft 可编辑 → submitted 提交不可编辑可删 → 隐式通过 approved + published_at=NOW()；后端不幂等） | §CreateContentPost / §UpdateContentPost | ✅ |
| REQ-CPB-5 | CreateContentPost 契约（section_code + entry status + community_ids，无 division_id 入参，后端展开 community_admin 唯一管辖 division，≤100 展开快照计量，080003/080005/080006 消歧，JWT 派生 publisher_id/role/publisher） | §CreateContentPost / §Design Gate | ✅ |
| REQ-CPB-6 | 附件绑定校验（GetFileUrl 扩展 FileInfo：confirmed+user_id 归属+file_type 回读；≤10 个/≤50MB 单源；080005） | §CreateContentPost / §附件绑定 | ✅ |
| REQ-CPB-7 | Kafka content-review 推送（停 Redis 只推 Kafka，契约单源 REQ-CPM-2，at-least-once 待推标记 + 定时重推，推送失败不阻塞但登记 + 可观测） | §Kafka 推送 | ✅ |
| REQ-CPB-8 | attachment_count 审核完整性判定（正文 approved + 已审附件数==计数 → 展示；任一 rejected 隐藏；读路径不 mutate status） | §审核完整性判定 | ✅ |
| REQ-CPB-9 | UpdateContentPost（draft 编辑 + attachment_count 同事务重算 + scope 复校验 + is_pinned 置位 + submit 动作 + 非 draft 编辑 080005） | §UpdateContentPost | ✅ |
| REQ-CPB-10 | DeleteContentPost（撤回：仅发布者 080002、软删 + status=withdrawn、scope/附件保留、不推 Kafka） | §DeleteContentPost | ✅ |
| REQ-CPR-1 | 列表按 scope 过滤 + section_code 筛选 + role 过滤（notice 兼容语义）+ community_id 取请求小区 + published_at DESC（NULLS LAST 防御） | §ListContentPosts | ✅ |
| REQ-CPR-2 | 详情 community_id 必填（080005）+ scope 外/不存在/未完整统一 080001 + 附件 file_id 重生 URL；**REST 层对既有移动端无 community_id 调用做兼容回退（R2）** | §GetContentPost / §REST wire 兼容 | ✅ |
| REQ-CPR-3 | GetMarqueeNotices（≤10、置顶优先、15 天含端点、仅审核通过且完整、items=id+title、080005） | §GetMarqueeNotices | ✅ |
| REQ-CPM-1 | Kafka 基建（单节点 KRaft + 数据卷 + content-review topic + retention 覆盖消费者空窗 + 全栈启动可达） | §Kafka 基建 | ✅ |
| REQ-CPM-2 | content-review 消息契约（单源：version/post_id/section_code/text/publisher_id/attachments[{file_id,file_type,review_status,file_url}]；file_url 可再生） | §Kafka 推送契约 | ✅ |
| REQ-CPM-3 | content_posts 停 Redis moderation:task:queue（lostfound/user 仍走 Redis，机制保留） | §Kafka 推送 / §moderation-service | ✅ |
| REQ-CPM-4 | moderation Redis 消费者跳过 source_type="notice"（不回调 NoticeService）+ UpdateNoticeModerationStatus RPC 移除 | §moderation-service / §Proto | ✅ |
| REQ-CPM-5 | 审核消费者后期开发（本期只定契约+推送；submit 即隐式通过 status=approved + published_at=NOW()） | §本期范围 | ✅ |
| REQ-CPP-1 | GetPublishPermission（can_publish + 可发布角色含 property_admin，level-2 判定经 GetUserRoles） | §GetPublishPermission | ✅ |
| REQ-CPP-2 | 各角色发布范围（property_admin 本小区 / grid_worker 多小区 / community_admin **经既有 community grant 派生唯一 division 展开**（R1） / committee 本小区 / 业主只读）+ division→community 行为结论化 | §服务归属 / §Design Gate | ✅ |
| REQ-CPP-3 | 权限种子（property_admin 保留 421 + grid_worker 授 421 + 撤销 (1,421)/(5,421) 保留 435/436 + 421 min_verf_level 0→2 + 读码 422 扩展/423/424/426 + 写码 427/428） | §权限种子 | ✅ |
| REQ-CAS-1 | 白名单 {png,jpg,jpeg,gif,pdf,doc,docx} + 禁止集 + zip/rar 扩展名拒绝 + 070004 | §附件安全 | ✅ |
| REQ-CAS-2 | 单文件 ≤10MB 硬上限 + 070005 | §附件安全 | ✅ |
| REQ-CAS-3 | 两层校验（L1 扩展名+大小快速拒绝 + L2 magic-bytes 回读；doc OLE2/CFB+WordDocument、docx ZIP+word/document.xml 特判） | §附件安全 | ✅ |
| REQ-CAS-4 | 全局基线 + 按 entity_type 可扩展（10MB 硬上限与禁止集不可弱化） | §附件安全 | ✅ |
| REQ-CAS-5 | file_id/file_type 记录（自 FileInfo 回读）；单帖总量上限单源 REQ-CPB-6 | §附件绑定 | ✅ |

> 设计蔓延声明：§权限种子中的读码 422/423/424/426 + 写码 427/428 的「全部移动端角色」绑定（422 扩展、423/424/426/427/428 新增）沿袭存档 notice REQ-PP-4 契约，属通用化后读写端点权限矩阵的必要闭环（fail-closed 下缺 def → 403），非新增需求——标注供 reviewer 确认。§Design Gate 的 community_admin division 展开判据经既有 community grant 派生（R1 替代原存档 notice design gate REV-17 的 community_div grant 假设），供 community_admin 唯一 division 展开落位。

---

## 服务归属决策

| 功能 | 归属服务 | 理由 |
|------|---------|------|
| content_posts 表 RENAME + 字段演化 + content_post_scope + content_post_attachments + Kafka 待推列 | community-hub-service | 数据归属（content_posts 域） |
| CreateContentPost（多小区 scope + 附件绑定 + 单事务 + 入口状态 draft/submitted + division 展开） | community-hub-service | 数据归属 + 业务域 |
| UpdateContentPost（draft 编辑 + attachment_count 重算 + is_pinned + submit）/ DeleteContentPost（撤回） | community-hub-service | 数据归属 + 状态机 |
| ListContentPosts / GetContentPost / GetMarqueeNotices（scope 过滤 + 审核完整性谓词） | community-hub-service | 读路径数据归属 |
| Kafka content-review 推送（at-least-once 待推标记 + 定时重推） | community-hub-service | 发布侧推送 |
| GetPublishPermission 判定 | community-hub-service（判定逻辑）+ permission-service（GetUserRoles 角色状态权威） | 入口显隐属业务面；角色状态数据在 permission-service |
| publisher 展示名真实档案查询 | community-hub-service（经 user-service `GetUsersByIds` RPC，Task 1.9 接线） | 身份档案数据在 user-service（跨服务 gRPC，禁请求体信任） |
| Kafka 基建（docker-compose 单节点 KRaft + content-review topic） | docker-compose（全局基建） | 基础设施归属 |
| Redis 消费者跳过 source_type="notice" + NoticeServiceClient 接线移除 | moderation-service | 消费者归属 |
| 附件白名单/10MB/magic-bytes 两层校验 + FileInfo 扩展 + 070004/070005 | file-service | 文件数据归属 |
| `resolvePublishScope` 社区管理员角色感知展开（community grant → division 子树） | permission-service（AssertPublishScope 判据） | 授权判据权威 + rel_user_role 数据归属 |
| GetResidentialAreasByDivision 展开 / GetResidentialArea（community_div_id 派生） / ResolveScopeAncestors 祖先链 | master-data-service | 只读复用（已存在） |
| 前端 web/pc | 不做 | Q10 已拍板：本期只改后端 |
| 前端 web/mobile | 不做（**但 REST wire 兼容保证既有消费方不破坏**，R2） | Q10 已拍板 + 本期已核实存在活跃通知消费方（pages.json tabbar / notice.vue / notice-browse.vue / notice-detail.vue）→ 采用 wire 兼容（§REST wire 兼容），非「前端本期不改」简单带过 |

**归属存疑项（已裁决）**：
- **division→community 授权落位（R1，grounded）**：代码库无 `community_div` scope_type grant（rel.go:77-82 / apply_role_logic.go:15,66 / rbac-design.md:38），原 design 的「`resolvePublishScope` 收集 community ∪ community_div 双 scope_type」是对不存在数据做变更 → **废弃**。推荐并采纳：`ResolveAdminDivision`（community-hub）经 community_admin 的既有 `scope_type='community'` grant（scope_id=communityId）→ masterdata `GetResidentialArea(communityId).community_div_id` 得管辖 division；permission-service 新增社区管理员角色感知 `resolvePublishScope`（社区管理员的 community grant 展开为其 division 下全部 approved 小区，其余角色 community grant 语义不变）。证据见 §Design Gate。
- **Kafka 客户端库选型**：仓库无既有 Kafka 依赖（common/pkg 无 kafka，go.mod 无 kafka 依赖）→ 推荐 `github.com/segmentio/kafka-go`（KRaft 单节点友好、producer 支持 acks=all + Writer 语义，契合 at-least-once）。备选 sarama。见 §ADR。

---

## 数据模型

### 迁移文件：community-hub-service `migration/003_content_posts_generalize.sql`

> **⚠️ 003 为一次性 RENAME 迁移（R4）**：MySQL 无 `RENAME TABLE IF EXISTS`，本迁移**仅执行一次，勿重跑**；重跑报错为预期（与 001/002 幂等风格不同属有意为之）。三态库执行见 Task 6.1。
> **⚠️ 部分失败恢复（V5，评审 data-model v4 S2）**：003 为**单脚本内串联**（RENAME → 多次 ALTER → CREATE TABLE → MODIFY）。若中途某条 ALTER 失败，表已 RENAME 但后续语句未执行，处于「已 RENAME、缺新列」的**半完成态**——**禁止直接重跑完整脚本**（RENAME 已发生会报「表不存在/重名」）。恢复指引：先 `RENAME TABLE content_posts TO notices`（或按失败语句前状态手动对齐表结构）回到可重入状态，再修复后重跑完整 003。

```sql
USE community_hub_db;

-- 1. notices → content_posts RENAME + content→text + section_code/status/attachment_count（REQ-CPB-1）
RENAME TABLE notices TO content_posts;
ALTER TABLE content_posts
    CHANGE COLUMN content `text` TEXT NOT NULL COMMENT '一段文字（图文发布核心，原 content，D1；反引号包裹防解析歧义，评审 I7；REST wire 仍以 content 键输出，见 §REST wire 兼容）',
    ADD COLUMN section_code VARCHAR(30) NOT NULL DEFAULT 'notice' COMMENT '板块：notice=通知/repair=维修保修/...（D11）',
    ADD COLUMN status TINYINT NOT NULL DEFAULT 0 COMMENT '全生命周期+审核结果：0=draft 1=submitted 2=approved 3=rejected 4=withdrawn（REVISION，权威契约）',
    ADD COLUMN attachment_count INT NOT NULL DEFAULT 0 COMMENT '附件计数（审核完整性判定载体，D15）';

-- 2. published_at / community_id 去 NOT NULL（REQ-CPB-1；D1 迁移语义，审核锚定；弃用列不写入）
ALTER TABLE content_posts MODIFY published_at DATETIME DEFAULT NULL COMMENT '审核锚定：本期 submit 即置 NOW()（隐式通过 D16）；消费者上线后按审核结果覆盖（D27）';
ALTER TABLE content_posts MODIFY community_id BIGINT DEFAULT NULL COMMENT '弃用：范围关联单源 content_post_scope（兼容期保留列，不写入，REUSE:notice-D1）';

-- 3. Kafka at-least-once 待推标记（D20：落库待推 + 定时重推；推送失败不阻塞发布但登记 + 可观测）
ALTER TABLE content_posts
    ADD COLUMN kafka_push_status TINYINT NOT NULL DEFAULT 0 COMMENT '0=无待推 1=pending-push 2=已推(ack)',
    ADD COLUMN kafka_push_retries INT NOT NULL DEFAULT 0 COMMENT '重推次数',
    ADD COLUMN kafka_push_last_error VARCHAR(500) DEFAULT NULL COMMENT '最近一次推送错误摘要（可观测）',
    ADD COLUMN kafka_pushed_at DATETIME DEFAULT NULL COMMENT '成功推送时间';

-- 4. content_post_scope 多小区关联（REQ-CPB-2，复用 notice_scope 模式）
CREATE TABLE IF NOT EXISTS content_post_scope (
    post_id      BIGINT NOT NULL COMMENT 'content_posts.id',
    community_id BIGINT NOT NULL COMMENT '目标小区（md_residential_area.id，代表小区或村）',
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (post_id, community_id),              -- 唯一约束（同一小区只一条）
    KEY idx_scope_community (community_id, post_id)   -- 读路径：列表/跑马灯按 community_id 先过滤
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='内容帖-小区范围关联（多小区发布单源；纯关联表仅 created_at——显式偏离编码规范 §3.1 的 updated_at/deleted_at，理由见 §字段约束补充，评审 S1/I3）';

-- 5. notice_attachments → content_post_attachments + post_id/review_status/file_id/file_type（REQ-CPB-3）
RENAME TABLE notice_attachments TO content_post_attachments;
ALTER TABLE content_post_attachments
    CHANGE COLUMN notice_id post_id BIGINT NOT NULL COMMENT '关联 content_posts.id（post_id 全链一致，无 main_id 别名）',
    ADD COLUMN review_status TINYINT NOT NULL DEFAULT 1 COMMENT '附件级审核：0=pending 1=approved 2=rejected（本期默认 approved，D14）',
    ADD COLUMN file_id BIGINT DEFAULT 0 COMMENT 'file-service 文件ID（重生预签名 URL 的权威载体，D14/REUSE:notice-D24）；兼容期存量行 0，读路径回退 stored file_url',
    ADD COLUMN file_type VARCHAR(20) DEFAULT NULL COMMENT '白名单校验通过的文件类型（扩展名，自 FileInfo 回读）';
ALTER TABLE content_post_attachments MODIFY file_url VARCHAR(1024) NOT NULL COMMENT '存量回退用 stored URL（新行写占位空串 ''，file_id 为权威重生载体；加宽防 ERROR 1406）';
```

**字段约束补充**：
- **列保留**：`title`/`text`/`published_at`/`publisher_id`/`is_pinned`/`role`/`publisher` 全字段经 RENAME 物理保留（REQ-CPB-1，D1/Q1 + REVISION）；`role`（VARCHAR(20) NOT NULL，发布角色，RBAC→映射派生）/`publisher`（VARCHAR(100) NOT NULL，展示名，取用户真实档案禁请求体信任）/`is_pinned`（TINYINT DEFAULT 0，跑马灯置顶载体）三列是 RENAME 契约的一部分（REVISION REQ-CPR-3 依赖）。
- **status 枚举（REVISION，权威契约）**：`0=draft / 1=submitted / 2=approved / 3=rejected / 4=withdrawn`——重定义 notice 时代枚举（0=pending/1=approved/2=rejected/3=withdrawn）；存量行无 scope 不可见，无冲突。**INSERT 显式写 status**（draft 入口写 0；submitted 入口/隐式通过写 2，D16）。
- **attachment_count**：每次附件集合变更（Create / draft 编辑）同事务重算，submit 时冻结（REVISION，D19 不变量）；审核完整性判定的比较基（D15）。
- **community_id 去 NOT NULL 是硬性门禁**（REVISION-9/coverage MUST-1）：新发布路径 community_id 置 NULL（scope 单源），若未迁移 INSERT 会被 NOT NULL 拒绝 → 迁移必须先于功能上线（与 published_at 去 NOT NULL 同类门禁）。
- **Kafka 待推列**（D20）：`kafka_push_status` 0/1/2 三态 + `kafka_push_retries` + `kafka_push_last_error`（可观测）+ `kafka_pushed_at`；重推扫描器（定时任务）只挑 `status=1` 重推直到 ack 或 quarantined；pending-push 计数经日志/指标暴露。
- **保留列**：`moderation_status`/`moderation_time` 兼容期保留（逐步过渡到 status + 附件级，D12）；`idx_community`/`idx_published` 兼容期保留并标记 deprecated（新读路径走 content_post_scope 索引，后续清理迁移再删）。
- **表引擎/字符集**：InnoDB / utf8mb4_unicode_ci（与现有表一致）；软删除：content_posts 保持 `deleted_at` 逻辑删除（撤回置 deleted_at + status=withdrawn），content_post_scope 为**物理保留**（撤回不删行，帖的撤回由主表软删表达，REQ-CPB-10）。
- **数据增长**：content_post_scope 行数 ≈ **累计发布帖数** × 平均目标小区数（≤100/条）——**撤回帖 scope 行永久保留（REQ-CPB-10，非物理删除），量级仍小（百级 × ≤100）、索引兜底无归档需求**（评审 S1 显式结论）。
- **content_post_scope 为不可变纯关联表（评审 S1 显式偏离登记）**：仅 `created_at`，无 `updated_at/deleted_at`（偏离编码规范 §3.1 时间三件套）——行不可变（同一小区同帖唯一），撤回由主表 `deleted_at` 表达（REQ-CPB-10 保留 scope 行），draft 改写走 delete+reinsert 物理替换，无逻辑更新需求；注释措辞由「物理删除语义」澄清为「纯关联表无软删列；撤回保留由主表软删表达；改写物理替换」（评审 I3）。

### 索引设计

| 索引 | 表 | 用途 |
|------|-----|------|
| `uk_post_community (post_id, community_id)` = 复合 PK | content_post_scope | 唯一约束 + 按帖查范围（同一小区只一条） |
| `idx_scope_community (community_id, post_id)` | content_post_scope | 列表/跑马灯按小区读路径（左列 community_id） |
| `idx_notice (post_id)`（RENAME 后自动改指 post_id） | content_post_attachments | 详情附件查询 |
| `idx_community` / `idx_published`（保留，deprecated） | content_posts | 兼容期旧代码；新读路径不走 |

---

## 审核完整性判定（REQ-CPB-8，读路径单源谓词）

**单一谓词 `IsReviewComplete`**（正文 `status == approved(2)` 且 `count(attachments WHERE review_status=approved) == attachment_count`）为**列表 / 详情 / 跑马灯**读查询共用（与存档 notice `IsModerationPassed` 同构，本变更为 status + 附件级双条件）：

- **读路径**：`FindListByCommunity` / `FindOneReviewComplete` / `FindMarquee` 一律加该谓词（`status = 2` AND 附件 approved 计数 == attachment_count），**任一附件 rejected → 谓词不成立 → 不展示**（谓词隐藏，读路径不 mutate status，REVISION 与设计文档 §3.1 对齐）。
- **status=rejected(3) 仅由审核流（后期消费者 D18）回写**，读路径绝不写 status。
- **attachment_count 冻结**（D19）：submit 时冻结；draft 编辑时同事务重算——无陈旧计数致帖永不可见（REQ-CPB-9 场景）。
- **无附件帖恒完整**：count(approved)=0 == attachment_count=0 恒成立。
- **本期自洽**：submit 即置 status=approved + 附件 review_status 默认 approved → 谓词立即满足，无消费者也可见（D16/REQ-CPM-5）。
- **防御回退**：读路径遇 file_id=0/NULL 附件行按 review_status 计入完整计数并回退 stored file_url（存量帖本期不可达，防御性代码不参与可观测行为，REQ-CPB-3 场景）。

---

## 接口设计

> 所有 int64 ID 字段 `[jstype=JS_STRING]`（硬性约束 #3，SEE [[proto-jstype]]）；status/review_status 用 **int32** + 文档化值语义（与既有 `UpdateModerationStatusRequest.moderation_status` int32 模式一致；DB 枚举为权威契约，避免枚举 0=UNSPECIFIED 偏移映射错误）；**entry_status / Update.status 同样 int32 + 文档化值（0=draft / 1=submitted），REST / Proto / DB 三侧同号，从根因消除「REST 1=submitted ↔ proto 1=DRAFT」跨层枚举错位（对应设计评审 M2）**；错误码 5 位 `XXYYY`。

### 服务与 RPC 契约（D4 直接改名，一次到位）

`NoticeService` → **`ContentPostService`**，`CreateNotice`/`ListNotices`/`GetNotice`/`UpdateNotice`/`DeleteNotice` **直接改名**为 `CreateContentPost`/`ListContentPosts`/`GetContentPost`/`UpdateContentPost`/`DeleteContentPost`（不做兼容别名，D4；**REST 响应 wire 形状保留 + REST 路径保持 /notices → 既有 web/mobile 消费方不破坏，见 §REST wire 兼容（R2）**）；**`UpdateNoticeModerationStatus` 回调 RPC 移除**（D21——本期无回调路径；`UpdateModerationStatusRequest/Response` 消息**保留**，LostFoundService 仍用）；**新增** `GetPublishPermission` / `GetMarqueeNotices` RPC（D5/Q5，代码中本不存在，新实现）。

### community-hub-service.CreateContentPost（重写）

- **输入**：`section_code`（注册板块白名单）、`title`、`text`（原 content）、`entry_status`（**int32 值语义：0=draft 默认 / 1=submitted 立即提交**——REST types.go 与 DB status 0/1 **同号**，禁止裸「透传」的枚举错位，对应评审 M2）、`repeated community_ids`（int64 JS_STRING，≥1）、`repeated attachment_ids`（**int64 JS_STRING = file_id，接口 v4 INFO 1 对齐 `ContentPostAttachment.file_id`——同一 ID 单一载体内型，REST 层 `[]string` 由代理转换**）。**无 division_id 请求字段**（REVISION-10/A2——社区管理员唯一管辖 division 由后端自 scope 派生展开）；**无 role/publisher_id/publisher 请求字段**（服务端从 JWT/RBAC/真实档案派生，禁请求体信任）。
- **校验顺序**（任一步失败整体拒绝，无部分写）：
  1. section_code ∈ 注册板块集（本期 `notice`）→ 否则 080005；title/text 非空 → 否则 080005。
  2. scope：`community_ids` 去重；空（全去重后）→ 080005；**展开后快照 >100 → 080003**（REVISION：按展开后长度计量）。
  3. **社区管理员自动展开（R1，grounded）**：`GetUserRoles(user_id)` → 取 role_code=community_admin 的 `scope_type='community'` 且 `scope_id!=0` grant（scope_id 列表）→ 逐条 masterdata `GetResidentialArea(scope_id).community_div_id` → 收集 distinct division 集；**空 → 080005（非 admin 走 community_ids 直传路径）；>1 个 distinct division → 080005（「唯一管辖」契约守卫，fail-closed，评审 I4 语义保留）** → 唯一 division 经 masterdata `GetResidentialAreasByDivision(division, status=1)` 展开为 approved 小区集（展开为空 → 080005；>100 → 080003）。其他角色不接受 division 语义（请求本无 division_id 字段，故无「非 admin 传 division」路径；grid_worker/committee/property_admin 直传 community_ids）。
  4. 附件绑定（REQ-CPB-6，单源）：逐 `attachment_id` 调 file `GetFileUrl(file_id)` → `FileInfo.confirmed==true` 且 `user_id==JWT`；数量 ≤10 且 `Σ file_size ≤ 50MB` → 否则 **080005**（附件引用无效/超限）；`file_type`/`file_name`/`file_size`/`download_url` 从 FileInfo 回读。
  5. 数据权限：**单次批量** `AssertPublishScope(user, targets)`（一次携带全部 target ScopeRef；任一越权/不可解析（目标不存在）→ 060007→**080006** 一次映射整体拒绝，fail-closed，D31/REUSE:notice-D31；**community_admin 目标小区由其 division 子树授权覆盖，见 §Design Gate**）。
- **鉴权**：功能权限 `community:notice:create-api`(421) + `min_verf_level=2`（REQ-CPP-3 行为变更）由 REST PermMiddleware 强制（先于 scope 校验，080002）；写路径角色状态门槛 = level-2（status=2 且 verified_at NOT NULL 且未过期）。
- **幂等**：**不幂等**（REUSE:notice-D25，无幂等键；重复提交产生重复帖；防重由前端提交中禁用承担——本期前端不接线，客户端语义由契约声明）。
- **落库**（单事务，all-or-nothing）：content_posts（`community_id=NULL`、`published_at`：draft=NULL / submitted=NOW()、`status`：draft=0 / submitted=2（隐式通过）、`attachment_count`=附件数、`section_code`、`title`/`text`、`is_pinned=0`、`role`=RBAC→发布角色映射字符串、`publisher`=真实档案展示名（user-service `GetUsersByIds`，Task 1.9 接线）、`publisher_id`=JWT）+ content_post_scope 批量插入 + content_post_attachments 批量插入（`post_id`、`file_name`/`file_size`/`file_type` 自 FileInfo 回读、`file_id`=attachment_ids、`file_url`=占位空串 `''`、`review_status=approved(1)` 默认）。
- **role 派生**（REVISION REQ-CPB-5）：JWT 仅含 user_id，实际角色显式调 permission `GetUserRoles(user_id)` 解析 → RBAC→发布角色映射（grid_worker→grid_officer、community_admin→community、committee→committee、property_admin→property；DB role 列存映射后字符串）；多角色取授权 421 的发布角色优先（顺序 grid_worker > community_admin > committee > property_admin）。`publisher` 展示名取用户真实档案（user-service `GetUsersByIds`），**禁请求体信任**（堵 createnoticelogic.go `Publisher: in.Publisher` 伪造向量，REVISION REQ-CPB-5）。
- **Kafka 推送（REVISION，D20）**：`submitted` 入口（立即提交）或 `UpdateContentPost.status=submitted` 提交时，同事务内置 `kafka_push_status=1`（待推标记），随后生产者推 content-review 消息（契约 REQ-CPM-2）；**推送成功置 2 + kafka_pushed_at；失败置 1 保留 + 记录错误摘要，不阻塞发布**（status 已 approved，可见）；重推扫描器定时补投。**不再 LPUSH Redis `moderation:task:queue`**（D3，content_posts 只推 Kafka）。
- **错误码**：080002（功能权限，功能层先于 scope）、080003（目标数超限）、080005（请求形状：空范围/section_code 非法/title·text 空/附件引用无效或超限/community_admin division 解析异常）、080006（目标级解析失败或越权，D31）。
- **输出**：`id`。

### community-hub-service.UpdateContentPost（新建，REVISION V5）

- **输入（presence 语义，V5 权威定义——评审 interface v4 MUST 1）**：`id`（必填）+ 各字段**显式声明 presence**，分支判定以 **presence/标志位** 为准（非 value 非空启发式）：
  - **`optional string title` / `optional string text` / `optional string section_code`**（proto3 `optional` → Go `*string`，带 presence）：**指针非 nil = 携带变更**；缺失（nil）= 不改。**携带空串 = 清空，但 title/text 携带空串 → 080005**（title/text 非空不变量，REQ-CPB-5，创建与 draft 编辑同规则——评审 MUST 1(c) 显式规定「编辑不得置空」）。
  - **`optional bool is_pinned`**（proto3 `optional` → Go `*bool`）：presence 即**双向置位**——`*true`=置顶、`*false`=取消置顶（REQ-CPB-9(f) 双向置位，取消置顶确定性可达——评审 MUST 1(a)）；缺失 = 不改。
  - **`repeated int64 community_ids`（全量替换集）+ `bool has_scope_change` 标志**：`has_scope_change==true` → community_ids 为**全量替换 scope 集**（去重后 ≥1，复跑 `AssertPublishScope`；**空集 → 080005**，帖必须 ≥1 scope 小区）；`false` → scope 不改。repeated 空数组/缺失与「不改」的区分由标志位承担（repeated 无法用 `optional`——评审 MUST 1(b)）。
  - **`repeated int64 attachment_ids`（全量替换集）+ `bool has_attachment_change` 标志**：`has_attachment_change==true` → attachment_ids 为**全量替换附件集**（**空集 = 清空全部附件 → 同事务重算 `attachment_count=0`**，D19 不变量可归零——评审 MUST 1(b)）；`false` → 附件不改。
  - **`int32 status`**（**0=无提交动作（编辑路径，默认）/ 1=submit（提交动作）**——与 REST types.go 同号（REST 1↔proto 1=submitted），RPC 校验：1→submit、0→编辑（仅 draft 可编辑）、其他值→080005；审核结果仅审核流写入；**proto 注释标注「action」语义，区别于 `ContentPost.status` 的「state」语义（评审 I2）**）。
- **授权分流（REVISION-11/评审 M1 + V5 修订，消除作者校验 vs 操作者置顶互斥）**：UpdateContentPost 先 `FindOne(id)` → 080001，再按请求形状分流：
  - **(a) 内容/附件/scope 编辑路径**（`Title!=nil` 或 `Text!=nil` 或 `SectionCode!=nil` 或 `HasScopeChange==true` 或 `HasAttachmentChange==true`，或 `status==1`）→ **先作者校验**（`publisher_id == JWT user_id`，非发布者 → **080002**），再按 draft/非 draft 走 080005（仅 draft 可内容编辑）。
  - **(b) 仅 is_pinned 路径**（`IsPinned!=nil` 且 Title/Text/SectionCode 均 nil、HasScopeChange/HasAttachmentChange 均 false、status==0）→ **跳过作者校验**，改验操作者授权：draft 帖 → 发布者即可；submitted/approved 帖 → `PublishRolesFrom` 非空 且 `AssertCommunitiesScope`（数据范围覆盖帖小区），满足即放行（**非作者操作者不 080002**），scope 不覆盖 → **080006**。**置顶（`*true`）与取消置顶（`*false`）均走此路径**。
- **is_pinned 持久化（V5 修订——评审 data-model v4 M1 修复）**：置顶/取消置顶一律走**独立列更新 `UpdateIsPinned(ctx, id, isPinned)`**（`UPDATE content_posts SET is_pinned=? WHERE id=? AND deleted_at IS NULL`，仅写 is_pinned 列，**不碰 title/text/section_code**）——**禁止复用 `UpdateContent`（同时写 title/text/section_code/is_pinned）传空 title/text，会把已 approved 帖正文清空（发布后内容数据丢失）**。draft 与 submitted/approved 帖均如此。
- **draft 编辑**（仅 status==draft(0)，走 (a) 分支）：
  - **内容字段全量替换语义（V5 显式声明——评审 MUST 1(c)）**：`Title`/`Text`/`SectionCode` 任一 presence → 对应列按请求值覆盖（presence 即「本次携带完整现值」）；未携带列保持现值。**附件/scope-only 变更（Title/Text/SectionCode 全 nil）经字段级更新，正文列不动**——不得复用 `UpdateContent` 传空 title/text 覆盖正文。
  - 附件集合变更（`has_attachment_change==true`）→ **同事务重算 `attachment_count`**（新绑定数，空集=0）+ 复跑完整绑定校验（REQ-CPB-6：confirmed + user_id + file_type 回读 + ≤10/≤50MB）→ 080005 超限整体拒绝；新附件 review_status 默认 approved。
  - scope 变更（`has_scope_change==true`）→ 复跑 `AssertPublishScope` 于新目标集（任一越权 → 080006，all-or-nothing）+ 重写 content_post_scope 行。
  - is_pinned（请求含 is_pinned）→ 经 `UpdateIsPinned` 置位（draft 发布者即可，走 (b) 分支语义）；submitted/approved 帖由持发布角色的授权操作者（数据范围覆盖帖小区）置位（REVISION REQ-CPR-3 依赖 REQ-CPB-9(f)）。
- **submit 动作**（`status==1` on draft，走 (a) 分支）：校验最终态 → 推 Kafka content-review（REQ-CPB-7）→ **同事务置 `status=approved(2)` + `published_at=NOW()`**（隐式通过 D16，原子）→ 帖不可再内容编辑。**submit 路径不再 LPUSH Redis `moderation:task:queue`（D3——updatenoticelogic.go 原 `CreateAuditLog` + `LpushCtx("moderation:task:queue")` 块整体移除，只推 Kafka，评审 M3）**。**并发重复 submit（两请求同读 draft）可双推 content-review——at-least-once 容忍重复投递，后续消费者按 `post_id` 幂等去重（评审 INFO I3，见 §Kafka 推送契约），不设防重锁**。
- **非 draft 不可编辑**：`status != draft` 的任何**内容/附件/scope 变更**（非 is_pinned-only）→ **080005（仅 draft 可编辑）**；`attachment_count`/scope 不变（all-or-nothing）。
- **错误码**：080001（不存在）、080002（**(a) 内容编辑路径**的非发布者）、080005（仅 draft 可内容编辑 / status 非法值 / 附件超限 / **title·text 携带空串** / **scope 空集**）、080006（**(b) is_pinned 路径**操作者 scope 不覆盖）。

### community-hub-service.DeleteContentPost（撤回，REVISION）

- **输入**：`id`。
- **流程**：`FindOne(id)` 未找到 → 080001；**作者校验**：`publisher_id == JWT user_id`，否则 **080002**（仅发布者本人，REUSE:notice-D19）；**单事务**：`conn.Transact` 传共享 session——`SoftDelete`（置 deleted_at）+ `status=withdrawn(4)`；**content_post_scope 行与附件行全部保留**（帖的撤回由主表软删+withdrawn 表达，scope 共享单帖行，REQ-CPB-10）；**不推 Kafka**（撤回非审核提交）。
- **可用状态**：draft/submitted/approved 均可删（REVISION——draft 可删、submitted 不可编辑但可删）；withdrawn/rejected 已不可见无需删。
- **错误码**：080001、080002。
- **鉴权**：REST 权限码 `427 DELETE:/api/community/notices/:id`（新增，绑定全部移动端角色）由 PermMiddleware 强制；真正越权判定交 080002 作者校验。

### community-hub-service.ListContentPosts（改造）

- **输入**：`community_id`（单值）、`role`（可选筛选，notice 兼容语义——按发布者 role 列筛选）、`section_code`（可选板块筛选）、`page`/`page_size`（≤100，默认 10）。
- **role 枚举→DB 列映射单一来源（V5，评审 data-model v4 I2）**：读侧 `ContentPostRole` → DB role 列映射与写侧 `PublishRoleToString`（Task 1.8，RBAC code→DB role 列）**收敛到 helper.go 单一函数**（如 `ContentPostRoleToString`），防两份产生同一字符串集合的映射漂移。
- **流程**：`FilterAllowed(userID, community_id)`（GLOBAL 放行 / LIMITED IN / EMPTY 空列表）→ `FindListByCommunity` JOIN content_post_scope（`scope.community_id=?` + **`IsReviewComplete` 谓词（REQ-CPB-8）** + deleted_at IS NULL + 可选 section_code/role）→ `order by is_pinned desc, published_at desc`（本期 published_at 恒非空，防御 NULLS LAST）→ 分页。**JOIN 投影**：显式 `content_posts.*` + `content_post_scope.community_id`（右表限定，防 `select *` 双 community_id 列按列序取到弃用 NULL）。
- **输出**：`ContentPost[]`（**community_id = 请求小区**，scope 匹配行派生）+ `total`。
- **鉴权**：REST 权限码 `422 GET:/api/community/notices`（**扩展绑定全部移动端角色**）由 PermMiddleware 强制；越权读返回空列表不泄露。
- **错误码**：无（越权读空列表）。

### community-hub-service.GetContentPost（重写，含 REST 兼容回退 R2）

- **输入**：`id` + **`community_id`（RPC 契约必填，缺失/空 → 080005）**。
- **REST 绑定**：`GetContentPostReq.CommunityId` 用 **`form:"community_id"`** 标签（GET query，勿用 json/path——go-zero httpx.Parse 对 GET 走 ParseForm，json 标签被 skip → 恒 0 → 每次详情 080005 全挂）。
- **REST 兼容回退（R2 修订，评审 M2，取消 grant 唯一假设）**：移动端 `getNoticeDetail(id)`（notice-detail.vue:103 / api/community.ts:133-136）**不传 community_id**。REST handler `GET /notices/:id` 在 query 缺 `community_id` 时，**按 content_post_scope 反查该帖所属小区集**，对每小区执行 `FilterAllowed(userID, community_id)`（与现网 getnoticelogic.go 反查 `community_id → FilterAllowed` 的 LIMITED 语义一致，任意小区允许即放行）——**不依赖用户 grant 恰好唯一**（多小区 grid_worker / 多房产业主详情页迁移后仍可用，不 080005）；**scope 反查后错误码消歧（V5，评审 data-model v4 S1）**：帖有 scope 小区但全部不可读 → **080001**（与 RPC 层 scope 外统一 080001 一致，不泄露）；仅「帖无任何 scope 小区（数据异常）」→ 080005。**RPC 层 GetContentPost 保持 community_id 必填不变**（新消费方走严格契约），兼容回退只落在 REST 薄代理层。本期移动端不改也保持可用。
- **流程**：`FindOneReviewComplete(id)`（谓词 REQ-CPB-8）未找到 → 080001；`content_post_scope` 匹配 `(id, community_id)` 缺失 → 080001；`FilterAllowed(userID, community_id)` false → 080001（scope 外/不存在/未完整统一 080001，不泄露——**含原 GetNotice 读路径 scope 外 080006 拒绝路径，统一映射 080001，避免与写路径 080006 混淆，评审 INFO 1**）。
- **输出**：`ContentPost`（community_id=请求小区）+ `attachments[]`（id/file_name/file_url/file_size/file_type/**review_status**/file_id）。
- **附件 file_url 重生**（REUSE:notice S4）：每个附件按 `content_post_attachments.file_id` 调 file `GetFileUrl(file_id)` 重生预签名 URL（file_type/file_name/file_size 同源回读）；兼容期 file_id=0/NULL 行回退 stored file_url（防御，存量帖本期不可达）。
- **错误码**：080001、080005。

### community-hub-service.GetMarqueeNotices（新建，D5/Q5）

- **输入**：`community_id`（必填；缺失/空/0 → 080005）。REST 绑定 `form:"community_id"`。
- **流程**：`FilterAllowed` → `FindMarquee(community_id, since=now-15*24h, limit=10)`：JOIN scope、**`IsReviewComplete` 谓词**、`published_at >= since`（含端点）、`order by is_pinned desc, published_at desc`、`limit 10`。
- **输出**：`ContentPostMarqueeItem[]{id, title}` ≤10 条（跑马灯数据，不承载正文；命名对齐 ContentPostService，评审 S2）。
- **板块固定说明（评审 INFO 1）**：`GetMarqueeNotices` **无 section_code 入参、固定 notice 板块**——跑马灯仅通知场景，与通用化命名并列存在；通用化后未来其他板块跑马灯另设契约，本期不扩展。
- **鉴权**：REST 权限码 `423 GET:/api/community/notices/marquee`（新增，绑定全部移动端角色）。
- **错误码**：080005；空态空列表。

### community-hub-service.GetPublishPermission（新建，D5/Q5）

- **输入**：空（身份经 JWT metadata，`UserIDFromCtx`）。
- **流程**：permission `GetUserRoles(user_id)`；level-2 判定：`role.Code ∈ {grid_worker, community_admin, property_admin, committee}`（D6——property_admin 保留）且 `status==2` 且 `verified_at>0`（NOT NULL）且 `expires_at==0 OR expires_at>now`（基于 RPC 输出，禁止直读 rel_user_role）。任一满足 → `can_publish=true` + 追加映射后的发布角色（grid_worker→CONTENT_POST_ROLE_GRID_OFFICER、community_admin→CONTENT_POST_ROLE_COMMUNITY、committee→CONTENT_POST_ROLE_COMMITTEE、property_admin→CONTENT_POST_ROLE_PROPERTY）。
- **输出**：`can_publish bool` + `publishable_roles []ContentPostRole`；owner/tenant/merchant/sys_admin → can_publish=false。
- **鉴权**：任意已认证用户；REST 权限码 `424 GET:/api/community/notices/publish-permission`（新增，绑定全部移动端角色）。
- **错误码**：无（未登录由认证中间件 UNAUTHENTICATED → 前端视为 false）。

### REST wire 兼容性决策（R2，评审 MUST 修复——破坏面如实评估 + 显式兼容裁决）

> 破坏面评估修正：**web/mobile 存在活跃通知消费方**（评审实测）——(a) `web/mobile/src/pages.json` 中 `pages/notice/notice` 是 **tabbar 主页面**（L91-94），`notice-detail`（L34）/`notice-browse`（L66）均为注册页；(b) `web/mobile/src/api/community.ts:119-136` `getNoticeList` 读 `res.notices`、`getNoticeDetail(id)` 读 `res.notice`（**不传 community_id**）；(c) `notice.vue:336-337` / `notice-browse.vue:44,110-113` / `notice-detail.vue:103` 实际调用。proposal §风险评估「web/mobile 现无 notice 接线，已核实」**与代码现状不符**，本设计不再以「前端本期不改」作破坏缓解。

**裁决（评审选项 a——最小破坏面 + 与 Q10 自洽）**：
- **REST 响应 wire 字段名保持既有**：`ListContentPostsResp.notices`（数组）/ `GetContentPostResp.notice`（单对象）/ 帖体 `content`（正文）/ `role`/`publisher`/`publisher_id`/`is_pinned`/`published_at`/`created_at`/`updated_at`/`community_id`/`attachments[]`（`id/file_name/file_url/file_size`）——**全部保持现有 JSON 键**；Go 类型可改名（`NoticeInfo`→`ContentPostInfo` 等），但 JSON tag 保持旧键。**REST 路径保持 `/api/community/notices`**（权限码 422-428 path 一致）。
- **新增字段走新键（additive）**：`section_code`/`status`/`attachment_count`（帖体）+ `file_type`/`file_id`/`review_status`（附件）——移动端 TS 接口无对应字段，多出的键被忽略，无破坏。
- **正文键名映射（proto/DB 与 wire 分轨）**：DB 列 `text`、proto 字段 `text`、**REST wire 键 `content`**——REST API 层显式映射 `proto.Text → json:"content"`（ADR 登记：RPC 契约通用化、REST 外部契约稳定，分轨是本期 Q10 下破坏面最小化决策）。
- **详情 community_id 兼容**：见 §GetContentPost「REST 兼容回退」。
- **移动端「本期不改」语义澄清**：不是「无消费方」，而是「消费方 wire 契约保持兼容 → 无需改」；未来前端展示差异化接线时按 ContentPost 契约升级，另立任务。

### 跨服务一致性

- **Create/Update 单事务**：content_posts + content_post_scope + content_post_attachments 同一 MySQL 事务 all-or-nothing（本地事务，无跨服务写）。
- **Kafka 推送 at-least-once（D20）**：提交事务落 `kafka_push_status=1` → 生产者推 content-review → ack 置 2；失败保留 1 + 重推扫描器补投（定时）。推送失败**不阻塞发布**（status=approved 可见），但显式登记业务风险（推送失败=该帖永不审核，未来消费者上线后可见）——pending-push 可观测指标 + 日志。**非事务性 outbox**：补偿以落库待推标记 + 定时重推实现，不引入独立 outbox 组件（Won't have）。
- **附件校验**：经 file `GetFileUrl` 只读（无跨服务写），FileInfo 为校验载体（REUSE:notice-D24）。
- **division 展开**：masterdata `GetResidentialArea`/`GetResidentialAreasByDivision` 只读快照（发布时固定，后续 division 成员变化不影响已发布帖，REQ-CPP-2）；community_admin 的 division 由既有 community grant 派生（R1）。
- **publisher 展示名**：user-service `GetUsersByIds` 只读查询（Task 1.9 接线）。
- **审核链路**：本期无消费者（D18）；submit 隐式通过 status=approved + published_at=NOW()（D16），附件 review_status 默认 approved → 读路径谓词立即成立，内容直接可见。

---

## Kafka 基建（REQ-CPM-1，docker-compose）

- **单节点 KRaft 模式**（D8/Q8）：无 ZooKeeper；`process.roles=broker,controller`；app-network 固定 IP（**172.19.0.8**，网段 172.19.0.0/24，避开既有 0.2-0.7）；`./data/kafka-data` 数据卷持久化；healthcheck + `depends_on` 接线（community-hub 等服务在 Kafka 就绪后再启）。
- **content-review topic**（D17）：自动/启动时创建，partition=1、replication=1；**retention 覆盖消费者上线空窗**（本期无消费者，消息须存活至消费者存在或待推重推补投；retention 为配置项，设计约定「推送消息须存活到消费者存在或 pending-push 重推」）。
- **可达性**：Kafka 容器内稳定 advertised listener（`kafka:9092`，compose 网络内可解析）；community-hub 生产者配置 brokers 指向该地址。
- **验证**：全栈启动（`docker compose up` + `scripts/start.sh`）后 Kafka 可投递 content-review；容器重启后数据卷持久（broker 健康）。
- **客户端库**：`github.com/segmentio/kafka-go`（ADR）。

### Kafka 推送契约（REQ-CPM-2，单源）

```json
{
  "version": 1,
  "post_id": "123",            // int64 string
  "section_code": "notice",
  "text": "正文内容",
  "publisher_id": "456",       // int64 string
  "attachments": [
    { "file_id": "789", "file_type": "pdf", "review_status": 1, "file_url": "https://...presigned..." }
  ]
}
```

- `version`：契约版本（int32，本期 1；变更即 bump，供未来消费者协商，REVISION）。
- `file_url`：**可再生预签名 URL**（D7）——消费者直接拉取附件内容；经 `GetFileUrl(file_id)` 再生，非永久链接依赖（URL 过期后消费者再生）。
- `attachments[].review_status`：推送时刻为审核前默认值（本期 approved），**非审核结论**（快照）。
- **单源声明**：REQ-CPM-2 为唯一权威 payload 定义；REQ-CPB-7 引用之，不重复枚举字段（REVISION 消除双源漂移）。
- 无附件帖推 `attachments: []`（空数组非 null）。
- **消费者按 `post_id` 幂等去重（评审 INFO I3）**：并发重复 submit / 重推扫描（Task 1.19）可产生**同 post_id 重复消息**——at-least-once 语义容忍重复投递，**后续消费者（D18）须以 `post_id` 为幂等键去重**（本期推送方不承担，属消费者契约要求）。

---

## 附件安全（file-service，REQ-CAS-1/2/3/4/5）

- **白名单** {png, jpg, jpeg, gif, pdf, doc, docx}；**禁止集** {exe, bat, sh, cmd, com, msi, apk, js, vbs, ps1, py, pl, php}；**zip/rar 扩展名层全部拒绝**（070004）；大小写不敏感；无扩展名/点文件拒绝。
- **单文件 ≤10MB 硬上限**（070005，=10MB 放行）；**10MB 与禁止集不可被 entity_type override 弱化**（REQ-CAS-4 不变量）。
- **两层校验**：L1 GetUploadUrl 扩展名+大小快速拒绝；L2 ConfirmUpload **magic-bytes 回读**实际 MinIO 对象——png `89 50 4E 47`、jpg `FF D8 FF`、gif `47 49 46 38`、pdf `%PDF`、**doc = OLE2/CFB（D0 CF 11 E0 A1 B1 1A E1）且含 `WordDocument` 流**、**docx = ZIP（PK）+ 含 `word/document.xml` 部件**；其他 OLE2（msi/xls/ppt）与 OOXML（xlsx/pptx）与通用 zip/rar 内容 → 070004（docx 为唯一 zip 内容特判，REVISION 消歧）。
- **落库**：`file_type`（嗅探规范扩展名）+ `confirmed=true` 写入 File 记录（REUSE:notice-D24）；FileInfo 扩展 `file_type`(11)/`confirmed`(12)（非破坏，REQ-CAS-5/REUSE:notice REQ-AS-7）。
- **错误码**：`070004 不支持的文件类型`（int32=70004，`ErrCodeUnsupportedFileType`）/ `070005 文件大小超限`（int32=70005，`ErrCodeFileSizeExceeded`）——**新整数位**，不重编号既有 70001-70003（`ErrCodeFileOperationFailed` 保持 70003，REUSE:notice-D11）；file.proto 头注释错误码块对齐。
- **单帖总量上限（≤10 个/≤50MB，080005）单源在 REQ-CPB-6**（REVISION，REQ-CAS-5 引用不重复声明）。

---

## Design Gate（division→community 授权落位，R1 重写——grounded 于既有 community grant）

**前提核实（评审 MUST，已实测确认）**：`community_div` **不是**权限数据模型中的 scope_type——
- `services/permission-service/model/rel.go:77-82` scope_type 常量 = {`global`, `""`(empty), `community`, `building`, `unit`, `grid`}，**无 `community_div`**；
- `services/user-service/rpc/internal/logic/user/apply_role_logic.go:15,49-67,86-91` 将 owner/tenant/committee/grid_worker/property_admin/**community_admin** 一律绑定 `scope_type='community'`、`scope_id=communityId`（一个 `md_residential_area.id`，非 division）；
- `web/pc/src/views/users/UserRoles.vue` scopeType 类型 = `'community' | 'building' | 'unit' | 'grid'`，无 division 选项；`docs/specs/rbac-design.md:38` scope 域 = community/building/unit/grid/global；
- 全仓无任何 seed/SQL/分配流写入 `scope_type='community_div'`（`community_div` 仅作 masterdata `md_residential_area.community_div_id` 列，非授权 scope_type）。

**裁决（R1）**：废弃原 design 的「`resolvePublishScope` 收集 community ∪ community_div 双 scope_type」——那是针对不存在的 grant 类型做变更。改为**两层 grounded 机制**：

1. **community-hub `ResolveAdminDivision`（Task 1.7）**：取 community_admin 的 `scope_type='community'` 且 `scope_id!=0` **且 `URStatus==2`（level-2 等价：status==2 且 verified_at NOT NULL 且未过期，评审 M2 状态过滤——已过期/已驳回 grant 不计入，防其与另一 level-2 发布角色组合时越权驱动 division 展开）** 的 grant 列表 → 逐条 masterdata `GetResidentialArea(scope_id).community_div_id` → distinct division 集；**空 → 080005；>1 个 distinct division → 080005**（「唯一管辖 division」守卫，评审 I4 语义保留）→ 返回唯一 division D1。
2. **permission-service `resolvePublishScope`（Task 3.1，社区管理员角色感知）**：在 `resolveUserScope` 基础上新增发布专用解析——先收集 `scope_type='community'` grant（同现状，ids 基线）；**若用户持有 community_admin 角色**，对每个 community grant 的 scope_id 经 masterdata `GetResidentialArea` → division → `GetResidentialAreasByDivision(division, status=1)` 展开为 approved 小区子树，**并入 ids**；非 community_admin 角色（owner/tenant/committee/property_admin/grid_worker）**语义完全不变**（精确小区授权）。`targetCovered`（祖先链 ∩ ids）逻辑不变——展开后 division 下每个小区 T 的祖先链含自身 ∈ ids → 覆盖成立。
3. **目标集与授权集自洽**：community-hub 展开的 targets=[C1..Cn] 与 permission-service 为 community_admin 展开的 ids=[C1..Cn] 一致 → 逐 target 校验通过；任一展开目标不在授权集（如 division 含未 approved 小区而 GetResidentialAreasByDivision status=1 已过滤，或数据异常）→ 060007→080006 整体拒绝（fail-closed，无部分快照）。

**共享 blast radius 声明**：`AssertPublishScope` 是共享 RPC，还被 lostfound 创建、contacts upsert 调用。`resolvePublishScope` 角色感知展开生效后，community_admin 的 community grant 将连带放行 lostfound/contacts 写到其 division 下小区——大概率是修复（division 管辖权本就应覆盖其全部小区，现社区管理员仅能写自己那一小区实为权限过窄），但属共享语义变更，**必须显式声明并经回归验证**（Task 3.1 门禁场景 5）。

**验证门禁（编码前必须通过，Task 3.1/6.1）**：
1. **现状核实（Task 6.1，Owner 运维验证）**：真实 DB `SELECT DISTINCT scope_type FROM rel_user_role` 确认无 `community_div` 行（验证前提 grounded）。
2. community_admin 持 `scope_type='community'` grant（scope_id=C_admin，C_admin ∈ division D1）→ 发布到 D1 内另一小区 C1（`GetResidentialArea(C_admin).community_div_id=D1` → 展开含 C1）→ allowed；
3. 发布到 D1 外小区 C2 → 060007（denied）；
4. 目标小区不存在（ResolveScopeAncestors found=false）→ denied（安全拒绝未知节点）；
5. 非 community_admin（grid_worker/committee/property_admin/owner）行为不回归（精确小区授权，division 不展开）；
6. **共享调用方回归**：community_admin 持 community grant（C_admin）时（a）lostfound 发布到 C1（同 division）→ allowed、（b）contacts upsert 到 C1 → allowed；owner 发布到 C1 外小区仍 denied。

> 该 Task 标注 `SEE: [[grpc-timeout-layers]]`（AssertPublishScope 内嵌 ResolveScopeAncestors 三层超时对齐）。

---

## 权限种子（permission-service，REQ-CPP-3 REVISION）

> **fail-closed 语义（磁盘核实 checkpermissionlogic.go:55-59）**：移动端每个 content-post REST 端点必须有 sys_permission type=3 权限码并绑定角色，否则全体 403、RPC 契约不可达。

### 写路径权限

| 权限码 | 端点 | code | min_verf_level | 角色绑定 |
|--------|------|------|:---:|------|
| 421 | POST:/api/community/notices（创建） | community:notice:create-api | **0→2**（行为变更） | **property_admin(2) 保留**（D6，推翻 notice D26 回收）；**grid_worker(4) 新增授**；**owner(1)/tenant(5) 撤销**（DELETE (1,421)/(5,421)，保留 435/436）；community_admin(3)/committee(6) 保留 |
| 427 | DELETE:/api/community/notices/:id（撤回） | community:notice:delete-api | 0 | 全部移动端角色（真正越权判定交 080002 作者校验） |
| 428 | PUT:/api/community/notices/:id（编辑） | community:notice:update-api | 0 | 全部移动端角色 |

### 读路径权限

| 权限码 | 端点 | code | min_verf_level | 角色绑定 |
|--------|------|------|:---:|------|
| 422 | GET:/api/community/notices（列表） | community:notice:read-list-api | 0 | 现 (9,1,5) **扩展为全部移动端角色**（补 grid_worker/community_admin/committee/merchant/sys_admin） |
| 423 | GET:/api/community/notices/marquee（跑马灯） | community:notice:read-marquee-api | 0 | 全部移动端角色（新增） |
| 424 | GET:/api/community/notices/publish-permission | community:notice:publish-permission-api | 0 | 全部移动端角色（新增） |
| 426 | GET:/api/community/notices/:id（详情） | community:notice:read-detail-api | 0 | 全部移动端角色（新增——现无任何码，fail-closed 下全体 403） |

> **「全部移动端角色」**：{owner 1 / community_admin 3 / grid_worker 4 / tenant 5 / committee 6 / merchant 7 / sys_admin 8 / registered_user 9}。property_admin(2) `platforms='pc'` 不绑移动端读码（本变更仅改 421 保留）。
> **parent_id**：423/424/426 → 410（community:read）；427/428 → 420（community:notice）；422 保持 → 410。
> **REST 路径保持 `/api/community/notices`**（本期不改，R2 wire 兼容 + D10 前端不接线；权限码 path 与实际 REST 路由一致，SEE [[permission-seed-api-path-must-match-routes]]；RPC/proto/model 通用化改名为 ContentPost，REST 薄代理路径/响应键保持——ADR）。
> **幻影 435**：`community:lostfound:create-api` 无 sys_permission 行，仅 (1,435)/(5,435) 绑定引用——本变更不动 435/436（保留 owner/tenant 绑定）。
> **property_admin 发布权限不对称（评审 SHOULD #4，有意为之）**：property_admin 绑 421（create）但不绑 427/428（update/delete 绑「全部移动端角色」——property_admin `platforms='pc'` 不在移动端角色集）——PC 本期不接线，property_admin 的编辑/撤回走后续 PC 接线，创建后的操作由 080002 作者校验兜底；该不对称已在 rbac-design.md §6.5 注明。
> **080002 语义（评审 S1，跨端点重载）**：Create 功能权限层「无发布权限」/ Update·Delete「非帖作者」复用 080002（继承 notice 时代 DeleteNotice 即用 080002 表作者校验，REUSE:notice-D19）；proto 头注释扩展为「080002 — 无发布权限 / 非帖作者（功能权限层先于 scope 校验；Update/Delete 为作者归属校验）」。
> **rbac-design.md §6.5 矩阵补登**：421 发布角色集变更 + 422 扩展 + 423/424/426/427/428 新增 + property_admin 421↔427/428 不对称注明。

---

## 本期范围（审核消费者后期开发，D18）

- **实现**：Kafka 生产者推送（契约 REQ-CPM-2）+ at-least-once 待推标记 + 定时重推（D20）+ 停 Redis 队列推送（D3）+ submit 隐式通过（status=approved + published_at=NOW()，D16）。
- **不做**：moderation-service Kafka content-review 消费者（文字先关键字后大模型、图片/pdf 走大模型，结果回写正文→status、附件→review_status）——后期开发（D18）；本期只定契约 + 推送，submit 隐式通过使内容无消费者也可见。
- **不做前端接线（Q10，R2 澄清）**：web/mobile 既有通知消费方**保持 wire 兼容**（§REST wire 兼容），不做展示差异化接线；未来各板块展示差异化单独做。

---

## Proto 变更

### community/v1/community.proto

| 变更 | 类型 | 破坏性 | 说明 |
|------|:---:|:---:|------|
| 头注释错误码块对齐 | 注释 | 否 | 080001/080002/080003/080005/080006 语义与实际一致；**080004（寻失记录不存在，LostFoundService 仍用——`CodeLostFoundMiss` types.go:19，唯一使用方 lostfound；contact 逻辑无 080004 引用）保留不动**（评审 SHOULD 2，修正旧标签「便民联络不存在/ContactService 仍用」）；**080002 注释扩展为「无发布权限 / 非帖作者」**（评审 S1） |
| `NoticeService` → `ContentPostService`；5 个 RPC 直接改名 | 改名 | **是** | D4 直接改名一次到位，不做兼容别名；**破坏面评估见下方（R2——web/mobile 消费方 wire 兼容，非「无消费方」）** |
| `UpdateNoticeModerationStatus` RPC **移除** | 删除 | **是** | D21 本期无回调路径；`UpdateModerationStatusRequest/Response` 消息保留（LostFoundService 仍用） |
| `Notice` → `ContentPost`（text 语义 + section_code/status/attachment_count；role 保留为 ContentPostRole） | 改名+扩字段 | 是 | **保留既有字段号 1-12（`content`→`text` 用 4、`role`→`ContentPostRole` 用 5，title=3/publisher=6/created_at=10/updated_at=11 不动）；新增字段一律追加新号：`section_code=13` / `status=14` / `attachment_count=15`**（评审 M1——原 section_code(3)/status(10)/attachment_count(11) 与 title/created_at/updated_at 冲突、role 误标 6）；REVISION status int32 值语义 0-4 |
| `NoticeAttachment` → `ContentPostAttachment`（+file_type/file_id/review_status） | 改名+扩字段 | 是 | |
| `CreateContentPostRequest`（section_code/entry_status/text/community_ids/attachment_ids；**无 division_id 入参**） | 新契约 | 是 | REVISION-10/A2；**entry_status int32 0=draft/1=submitted，REST/Proto/DB 同号（评审 M2 去歧义，删 ContentPostEntryStatus 枚举）**；`attachment_ids` 统一 `repeated int64 [jstype=JS_STRING]`（对齐 `ContentPostAttachment.file_id`，接口 v4 INFO 1）；新契约消息全新字段号（不沿旧 CreateNoticeRequest 字段号，评审 I6） |
| `UpdateContentPostRequest`（draft 编辑 + submit 动作 + is_pinned，**V5 presence 语义显式化**） | 新契约 | 是 | **`title`/`text`/`section_code`/`is_pinned` 用 proto3 `optional`（Go `*string`/`*bool`，presence 可判）；`community_ids`/`attachment_ids` 为「全量替换集」，以 `has_scope_change`/`has_attachment_change` bool 标志区分「未携带=不改」与「空数组=清空」**（repeated 无法用 `optional`——评审 interface v4 MUST 1）；**`status` int32 0=无提交动作(编辑)/1=submit，与 REST 同号（评审 M2）；proto 注释标注「action」区别于 `ContentPost.status` 的「state」（评审 I2）**；置顶/取消置顶双向 + 清空附件/scope 语义见 §UpdateContentPost（V5） |
| `CreateContentPostResponse`/`GetContentPostResponse`/`UpdateContentPostResponse`/`DeleteContentPostResponse` 显式字段 | 新契约 | 是 | `CreateContentPostResponse{base=1, id=2}` / `GetContentPostResponse{base=1, ContentPost content_post=2}`（REST wire 键 `notice` 由 API 层映射，与移动端 `res.notice` 一致）/ `UpdateContentPostResponse{base=1}` / `DeleteContentPostResponse{base=1}`（对齐既有 Notice 响应，接口 v4 INFO 2） |
| 新增 `GetPublishPermission` RPC | 新增 RPC | 否 | D5 can_publish + publishable_roles |
| 新增 `GetMarqueeNotices` RPC + `ContentPostMarqueeItem` | 新增 RPC | 否 | D5 ≤10 条；命名对齐 ContentPostService（评审 S2）；板块固定 notice（评审 INFO 1） |
| `enum NoticeRole` → `ContentPostRole`（值不变） | 改名 | 是 | 契约统一；**buf breaking-check 预期 fail 项（FIELD_SAME_TYPE 类）登记 Task 0.1 人工核对清单**（评审 INFO 3） |

### file/v1/file.proto

| 变更 | 类型 | 破坏性 | 说明 |
|------|:---:|:---:|------|
| 头注释错误码块对齐 | 注释 | 否 | 070001-070005 语义对齐（REUSE:notice-D11） |
| `FileInfo` +`file_type`(11, string) | 新增字段 | 否 | 白名单规范类型 |
| `FileInfo` +`confirmed`(12, bool) | 新增字段 | 否 | 上传流程完成标志 |

> **破坏性影响评估（community.proto，R2 重写）**：D4 直接改名 + D21 移除回调是一次性破坏。**已如实核实消费方**：
> - **web/mobile（活跃，wire 兼容不破坏）**：`pages.json` tabbar `pages/notice/notice` + `notice-detail`/`notice-browse` 注册页；`api/community.ts:119-136` 读 `res.notices`/`res.notice`（不传 community_id）；`notice.vue:336-337`/`notice-browse.vue:44,110-113`/`notice-detail.vue:103` 实际调用。**缓解：REST wire 键保持（`notices`/`notice`/`content`）+ REST 路径保持 `/notices` + 详情 community_id 兼容回退（§REST wire 兼容）**——RPC/proto 层破坏不传导到移动端运行期契约。
> - **web/pc**：无 notice 消费方（grep 为空）——断言属实，无破坏。
> - **moderation-service**：同步移除 `NoticeServiceClient` 接线（Task 4.1）——同仓同步，无运行期破坏。
> - wire 上 file.proto 全部为兼容新增（新字段号）。合并前各变更走 `make ci` breaking-check 交叉验证（与进行中变更无冲突，check-change-conflict.sh 已跑）。

---

## 安全考虑

- **附件两层校验**（REQ-CAS-3）：L1 扩展名+大小快速拒绝 + L2 magic-bytes 回读；doc/docx 容器签名特判；msi/xls/ppt 改 doc、xlsx/pptx 改 docx、通用 zip/rar 改 docx → 070004（封堵同容器子类型绕过）。
- **附件引用信任边界**（REUSE:notice-D24）：community-hub 经 `GetFileUrl(file_id)` 读扩展 FileInfo（confirmed + user_id 归属 + file_type 回读），不信任客户端回传类型。
- **身份信任边界**（REVISION）：`publisher_id`/`role`/`publisher` 均以 JWT/RBAC/真实档案为准，请求体伪造被纠正（REQ-CPB-5 场景——堵 createnoticelogic.go `Publisher: in.Publisher` 展示名伪造向量）。
- **读路径隐私**：GetContentPost scope 外/不存在/未完整统一 080001（不泄露跨小区/未审内容）；越权 List 空列表；跑马灯仅审核通过且完整。
- **审核完整性谓词一致性**（REVISION）：`IsReviewComplete` 单一谓词承载列表/详情/跑马灯；读路径不 mutate status；status=rejected 仅审核流写入。
- **目标级解析失败 fail-closed**（REUSE:notice-D31）：不存在的 community_id 与越权同 080006。
- **权限回收回归**（REQ-CPP-3）：owner/tenant 撤销 421（保留 435/436）+ 421 min_verf_level 0→2（堵 level-0 未认证业主直调创建；SEE [[auto-grant-unverified-grant-confers-scope-level0]]）；property_admin 保留 421（D6）。
- **RPC 身份**：`UserIDFromCtx` 盲信入站 metadata，前提 RPC 回环绑定（既有）；新 RPC 同约束（SEE [[rpc-identity-spoofing-loopback-isolation]]）。
- **Kafka 消息安全**：content-review 消息含可再生 file_url（预签名，短期窗口），不落永久链接；消息本身无敏感明文扩展（text 为待审内容，消费者侧合规）。
- **division 授权派生防越权**（R1）：community_admin 的 division 仅由其**既有 community grant** 派生（非任意 division 指定）；>1 个 distinct division fail-closed 080005；展开目标与授权集一致校验兜底（§Design Gate）。

---

## 记忆引用（Step 1.5 注入，slug 均已自验存在）

| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| [[proto-jstype]] | Proto | community_ids/attachment file_id/publisher_id/post_id 均 int64 + `[jstype=JS_STRING]` |
| [[grpc-only-comms]] | 接口设计 | 附件校验/角色查询/scope 校验/division 展开/档案查询全经 gRPC，禁止直读 rel_user_role/uploaded_file/user_base |
| [[migration-must-execute]] | 数据模型 | 003 RENAME + 去 NOT NULL 迁移提交后必须执行；community_id/published_at 去 NOT NULL 先于功能上线（REQ-CPB-1 门禁）；**003 一次性勿重跑（R4）** |
| [[error-code-collision-and-namespace-alignment]] | 错误码 | 070004/070005 新整数位不重编号 70003；080005/080006 消歧；080002 跨端点语义扩展（R3） |
| [[error-code-literal-bypasses-qa-gate]] | 错误码 | 全量 errx 命名常量，禁裸数字 |
| [[permission-seed-api-path-must-match-routes]] | 权限种子 | 423/424/426/427/428 path 与 REST 路由逐一对照（REST 路径保持 /notices）；421 path 保持 |
| [[is-system-no-permission-shortcut]] | 权限种子/Design Gate | sys_admin 全权限经 rel_role_permission；AssertPublishScope 无短路 |
| [[auto-grant-unverified-grant-confers-scope-level0]] | 权限种子 | 421 min_verf_level 0→2 收窄未认证发布；与 GetPublishPermission level-2 一致 |
| [[insert-ignore-swallows-errors]] | 权限种子 | 撤销 (1,421)/(5,421) 必须显式 DELETE；grid_worker 授 421 用 INSERT IGNORE |
| [[grpc-timeout-layers]] | Design Gate/GetPublishPermission | AssertPublishScope(≤500ms) 内嵌 ResolveScopeAncestors；GetUserRoles/GetResidentialArea 超时对齐 |
| [[rpc-identity-spoofing-loopback-isolation]] | 安全 | UserIDFromCtx 盲信 metadata 的回环前提；新 RPC 同约束 |
| [[moderation-status-write-without-read-gating]] | 审核完整性判定 | 读路径恒应用 `IsReviewComplete` 谓词，杜绝「status 写 approved 但读路径不过滤」 |
| [[restore-compensation-zero-time]] | 数据模型 | published_at 用 `sql.NullTime`，严禁 `time.Time{}` 零值写 DATETIME |
| [[best-effort-compensation-must-log]] | Kafka 推送 | 推送失败（补偿路径）错误摘要落库 + 日志，不静默丢弃（D20） |
| [[grpc-max-msg-size-sensitive-words]] | Kafka 契约（未来消费者） | 未来消费者经 Kafka 拉取正文/附件 URL，不受 gRPC 4MB 限制；契约含 file_url 直接拉取 |
| [[tdd-red-evidence-requires-fail-excerpt]] | 任务 TDD | 含逻辑任务 RED 证据须含实际 FAIL 摘录 |
| [[cross-service-seed-deployment-order]] | 部署 | community-hub 003 迁移 + file 002 迁移 + permission 种子纳入部署编排（迁移先于上线） |
| [[api-required-field-marked-optional]] | 接口设计 | GetContentPost.community_id RPC 必填勿误标 optional；REST 兼容回退只落薄代理层（R2） |
| **proto3 `optional` presence 模式（V5，仓内先例 prose 记录——user-service `*string json:"...,optional"`、master-data `*int64`、file-service `*int64/*string`、permission.proto/user.proto 大量 `optional` 字段）** | 接口设计 | UpdateContentPost 的 title/text/section_code/is_pinned 用 proto3 `optional` 生成 pointer，presence 判定分支；repeated（community_ids/attachment_ids）以 bool 标志 `has_scope_change`/`has_attachment_change` 区分「未携带=不改」与「空数组=清空」——评审 interface v4 MUST 1 修复（无独立 slug，prose 记录） |
| [[go-deprecated-directive-not-test-comment]] | Proto | community_id deprecated 是机器可读指令，不当作占位注释 |
| [[monorepo-port-management]] | Kafka 基建 | Kafka 端口（9092/9093）与既有服务端口无冲突；docker-compose 网络分配稳定 IP |
| [[notfound-cache-sentinel-vs-transient-error]] | 不适用 | 无新缓存层（不做缓存，Q 决策）——排除 |
| [[redis-cache-soft-delete]] | 不适用 | 本变更不动 Redis 索引缓存——排除 |
| [[unique-index-migration-dup-precheck]] | 不适用 | content_post_scope 为**新表**无存量重复；唯一约束写路径去重兜底——排除 |
| `snake-camel-field-mismatch`（用户 auto-memory，非 harness 记忆库） | API 层 | TS 字段名与 Go snake_case 对齐（section_code/attachment_count/review_status 等）；**REST wire 键 `content` 与 Go `Text`/proto `text` 分轨是有意为之（R2 wire 兼容），非字段名不匹配**；以 prose 引用避免 slug 校验 MISS |
| **跨层枚举数值对齐（评审 M2 根因，prose 记录防再犯）** | 接口设计 | entry_status / Update.status 采用 REST / Proto / DB 三侧同号 int32（0=draft / 1=submitted），从根因消除「REST 枚举数值 ≠ proto enum 数值」的裸透传隐患；新增 pitfall 建议沉淀（「REST 枚举数值 ≠ proto enum 数值，禁止裸透传」）——以 prose 记录避免 slug 校验 MISS |
| **wire 兼容分轨（R2，prose 记录）** | 接口设计 | REST 外部契约稳定（notices/notice/content）+ RPC/proto/DB 通用化（text）分轨是 Q10 破坏面最小化决策；详情 community_id 兼容回退只落 REST 薄代理层，RPC 保持严格必填 |

**不适用记忆**（主动排除）：见上表末 4 行 + `[[grpc-max-msg-size-sensitive-words]]` 仅作未来消费者注记（本期消费者未实现）。

---

## 非功能设计

- **可靠性**：Create/Update 单事务 all-or-nothing；AssertPublishScope/GetUserRoles/GetFileUrl/GetResidentialArea/GetUsersByIds 依赖失败 fail-closed（080002/080005/080006 或传输错误）；Kafka 推送 at-least-once（待推标记 + 定时重推，推送失败不阻塞发布但登记 + 可观测，D20）；Delete 撤回全局生效（单事务）；UpdateContentPost 幂等性由状态机保证（draft 编辑/提交同事务）。
- **性能**：列表/跑马灯走 `idx_scope_community` 索引兜底（无缓存，沿用现有策略）；跑马灯 limit 10 + 15 天窗口；附件校验逐条 GetFileUrl（≤10 条 × 1 RPC）；单次发布目标 ≤100（080003）；Kafka 生产者批量/异步不阻塞发布；community_admin division 派生 RPC 数 = community grant 数 × 2（GetResidentialArea + GetResidentialAreasByDivision，典型 1-2 条 grant）。
- **可观测性**：pending-push 计数指标 + `kafka_push_last_error` 落库 + 日志；发布/提交/撤回核心路径 Infof；070004/070005 拒绝计数可经日志聚合；错误码全量 errx 常量。
- **移动端兼容回归（R2）**：REST wire 键 + 路径 + 详情 community_id 回退三者组合须在 Task 6.2 冒烟回归（移动端现行 notice 消费方在迁移后仍可用）。
- **显式「无」**：无新增缓存层；无幂等键（后端不幂等）；无批量操作；无 CDN/病毒扫描（magic-bytes 仅类型判定）；无事务性 outbox 中间件（D20 落库待推标记 + 定时重推替代）。

---

## 关键设计决策与权衡（ADR，轻量）

| 决策点 | 备选方案 | 最终选型 | 取舍理由 | 未采用方案原因 |
|-------|---------|---------|---------|--------------|
| notices 表处置 | 保留 notice 语义 + 平铺新板块 / **RENAME 通用化 content_posts** | RENAME 通用化（D1/D4/D12） | 一次到位服务未来板块；B 方案用户拍板 | 平铺则每板块复制一套表+发布+审核链路 |
| Proto 契约演进 | 兼容别名（旧名转发）/ **直接改名** | 直接改名（D4） | RPC 契约通用化一次到位；**REST wire 键保持让移动端消费方不破坏（R2）** | 兼容别名增加维护面，与「一次到位」冲突 |
| **REST wire 形状（R2）** | 改名 notices→posts/notice→post/content→text / **保留 notices/notice/content 键（Go 类型改名）** | 保留 wire 键（评审选项 a） | 移动端 tabbar/浏览/详情现行消费方零改动；与 Q10「只改后端」自洽 | 改名则移动端运行期静默破坏（业务不可用） |
| **详情 community_id（R2）** | 移动端补传小区 / **REST 层兼容回退（scope 反查 + 逐小区 FilterAllowed 任一允许即放行，v4 修订）** | REST 薄代理兼容回退（`ResolveReadableCommunityForCompat`） | RPC 保持严格必填（新消费方），移动端无改动；多小区用户（grid_worker/多房产业主）不 080005（v4 取消 grant 唯一假设）；与现网 getnoticelogic.go 反查 LIMITED 语义一致 | 移动端补传需改前端（违 Q10）；grant 唯一派生对多小区用户行为收窄（v3 评审 MUST 2 否决） |
| status 枚举 wire 表示 | proto enum（0=UNSPECIFIED 偏移）/ **int32 + 文档化值** | int32（0=draft...4=withdrawn；entry_status/Update.status 0=draft/1=submitted） | DB 枚举为权威契约，避免枚举偏移映射错误；与既有 moderation_status int32 模式一致；**entry_status/Update.status 三侧同号消除跨层枚举错位（评审 M2）** | enum 偏移引入 off-by-one 类映射 bug（REST 1=submitted ↔ proto 1=DRAFT 即此类） |
| **UpdateContentPost 字段 presence（V5）** | 值类型 + value 非空启发式 / **proto3 `optional`（标量）+ bool 标志（repeated）** | optional + 标志位（仓内先例一致） | presence 可判：**取消置顶（is_pinned=`*false`）、清空全部附件（has_attachment_change=true+空集 → attachment_count=0）、清空 scope（has_scope_change=true+空集 → 080005）确定性可达**，D19 不变量不被绕开；(b) is_pinned-only 分支判定可靠 | 值类型无法区分「未携带」与「false/空串/空数组」，分支判定留给实现者 → design 未定义的两种行为（评审 interface v4 MUST 1 否决） |
| **community_admin 管辖 division 来源（R1）** | 新造 community_div scope_type grant（评审选项 1）/ **经既有 community grant 派生（GetResidentialArea → community_div_id）** | 既有 community grant 派生（评审选项 2） | 权限模型 grounded（rel.go/apply_role_logic/UserRoles.vue 均无 community_div）；零种子/分配流/前端改动 | 新建 scope_type 需 model 常量 + 分配路径 + 迁移 + 前端 scopeType 扩展，重且与现状冲突 |
| community_admin 发布范围 | 请求带 division_id / **后端自 scope 派生唯一 division 展开** | 后端派生展开（REVISION-10/A2） | 前端不选 division；权限最小面（admin 只管辖一 division） | 请求带 division_id 引入前端选择面与伪造面 |
| Kafka 客户端库 | sarama / **segmentio/kafka-go** | segmentio/kafka-go | KRaft 单节点友好；Writer 语义契合 at-least-once；零历史包袱 | sarama 更底层，配置面大 |
| Kafka 推送补偿 | 独立 outbox 组件 / **落库待推标记 + 定时重推** | 落库待推标记 + 定时重推（D20） | 无跨服务事务；推送失败不阻塞发布且可观测 | outbox 组件增加基建复杂度（Won't have） |
| Redis 队列 | content_posts 仍推 Redis / **停 Redis 只推 Kafka** | 停 Redis 只推 Kafka（D3） | 内容级审核需消息持久化/回放/分区能力；lostfound/user 仍走 Redis 双轨过渡 | 继续 Redis 无法支撑多消费者与回放 |
| 本期审核态 | status 默认 approved 含混 / **submit 即隐式通过 approved + published_at=NOW()** | submit 置 approved + NOW()（D16 REVISION） | 消除「published_at 恒 NULL」矛盾；无消费者也可见 | status 默认 approved 表述含混，跑马灯锚点失序 |
| property_admin 发布权 | 剔除（notice D26）/ **保留本小区** | 保留（D6，推翻 D26） | 物业需发布本小区通知；与 B 方案通用化方向一致 | 剔除与物业发布产品需求冲突 |
| REST 路径 | 改名 /api/community/content-posts / **保持 /api/community/notices** | 保持 /notices（本期） | 权限码 422-428 path 一致；R2 wire 兼容 + 移动端零改动 | 改名需同步改权限码 path，churn 大且本期无收益 |
| moderation 回调 | 保留 UpdateNoticeModerationStatus / **移除 + Redis 消费者跳过** | 移除 + 跳过 source_type="notice"（D4/D21） | content_posts 不走 Redis 回调路径；删除死路径 | 保留则 dead code 与契约漂移 |

---

## 待办回填（BACKLOG）

- 「存量通知迁移回填」项（D2 存量不迁 → 上线瞬间存量通知从列表/详情/跑马灯消失，已登记 proposal 风险节）。
- 「消费者上线前 pending-push 积压处置」（D20：未来消费者上线前对 pending-push 积压的处置策略）。
- Kafka content-review 消费者实现（D18：文字先关键字后大模型、图片/pdf 走大模型、结果回写）。
- 「web/mobile 未来展示差异化接线」：本期 wire 兼容保运行期可用，不接线通用组件；未来按 ContentPost 契约升级前端展示（D10/Q10 自然演进，非本期欠债）。
