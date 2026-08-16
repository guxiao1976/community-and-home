# Proposal: 通用图文发布组件重构（content_posts 通用化 + 内容级审核 + Kafka）

> **优先级**: P1 · **改动规模**: 大（L） · **影响风险**: 高
> **核心风险点**: 跨 6 个服务协同（community-hub / moderation / file / permission / master-data / api-proto）+ 通知模块重构为通用发布组件（`notices` 表 RENAME 为 `content_posts` + section_code 板块化 + status 全生命周期枚举 + attachment_count + published_at 去 NOT NULL + 弃 community_id 走 scope）+ 引入 Kafka 基建（docker-compose 单节点 KRaft，本期安装）+ 审核链路从 Redis 切换到 Kafka topic（content-review）+ 内容级审核状态模型（正文 + 附件各自独立状态 + attachment_count 完整性判定）+ 两阶段发布状态机（draft/submitted 存储映射 + submit 隐式通过）
> **变更类型**: modify（改造既有「通知模块 notices 表」为通用 `content_posts`；原行为→新行为 diff 见 §影响范围；**原 notice-multicommunity-publish 变更被本变更取代/存档**——B 方案用户拍板：暂停通知流水线，重构为通用化一次到位）
> **设计文档**: `docs/superpowers/specs/2026-08-16-content-post-design.md`（用户已拍板，本次 REVISION 已同步 §2.1/§3.1/§4.2 消除漂移）+ 已确认决策包（Q1-Q10，映射 D1-D18）+ REVISION 新增决策 D19-D22
> **前置**: 本变更复用原 `notice-multicommunity-publish` 已确认设计决策（D1-D32：多小区发布 / 附件安全 / GetPublishPermission / GetMarqueeNotices / published_at 审核锚定 / 错误码消歧），但**推翻其中 D26（剔除 property_admin 移动端发布）**——本变更 property_admin 保留本小区发布权（D6）。原 notice 变更存档，本变更为独立 L 级。

## 为什么做

当前通知模块是**通知专属**的：`notices` 表字段（title/content/community_id 单值）绑死通知语义，只能服务「通知」一个板块；每个新板块（维修保修等）都要复制一套「表 + 发布 + 浏览 + 附件」链路；审核是通知级整体审核（moderation_status 单字段），无法对正文和附件分别做内容级审核；审核链路走 Redis List（`moderation:task:queue`），无消息持久化/回放/分区能力，消费者扩展（文字关键字+大模型、图片/pdf 大模型）缺乏标准消息基础设施支撑。

本变更把通知模块升级为**通用图文发布组件**：一次发布一个 `content_posts.id`，下辖一段文字 + 若干附件，对应 `section_code` 板块（notice=通知 / repair=维修保修 / …），为未来所有图文板块提供统一「发布/浏览/详情/附件安全/内容级审核」底座；引入 Kafka 作为内容审核消息链路（content-review topic），停用 content_posts 的 Redis 审核队列；实现内容级审核状态模型（正文 status + 附件 review_status 各自独立，attachment_count 做完整性判定——全部通过才统一展示，一个有问题则不展示）；引入两阶段发布状态机（draft 草稿可编辑 → submitted 提交后不可编辑但可删 → 审核）。让未来任何板块都能「一次发布、板块复用、内容级审核」。

## 做什么

1. **数据模型通用化（community-hub-service）**：`notices` RENAME 为 `content_posts`，字段升级为：保留 `title`/`text`（原 content）/`published_at`/`publisher_id`/`is_pinned`/`role`/`publisher` 全字段（Q1 title-retention + REVISION：is_pinned 为跑马灯置顶载体、role/publisher 为发布角色/展示名，均 RENAME 物理保留列），新增 `section_code`（板块：notice/repair/…）、`status`（**全生命周期 + 审核结果枚举：0=draft / 1=submitted / 2=approved / 3=rejected / 4=withdrawn，REVISION**）、`attachment_count`（附件计数，审核完整性判定载体）；`published_at` 去 NOT NULL（D1 迁移语义，审核锚定；**本期 submit 即置 NOW() 隐式通过，REVISION**）；弃用 `community_id`（兼容期保留列，范围关联单源走 content_post_scope）；保留 `moderation_status`/`moderation_time` 兼容期（逐步过渡到 status + 附件级）
2. **content_post_scope 多小区关联（新表，复用 notice_scope 模式）**：`post_id + community_id` 双 NOT NULL + 复合 PK + community_id 读索引；**存量不迁（Q2 legacy-data-migration：仅新数据走 scope）**
3. **content_post_attachments（改自 notice_attachments）**：加 `post_id`（关联 content_posts）+ `review_status`（附件级审核：0=pending/1=approved/2=rejected，本期默认 approved）+ `file_id`（预签名 URL 重生载体，D14/S4）+ `file_type`（白名单校验类型）
4. **内容级审核状态模型 + 审核完整性判定（D15）**：`content_posts.status`（正文）+ `content_post_attachments.review_status`（每附件）；`count(attachments WHERE review_status=approved) == attachment_count` 且 正文 status=approved → 整体可展示；任一附件 rejected → 整体不展示（**读路径谓词隐藏，status 由审核流写入，读路径不 mutate，REVISION 与设计文档 §3.1 对齐**）
5. **Kafka 基建 + content-review topic（D17）**：docker-compose 新增 Kafka（**单节点 KRaft 模式 + 数据卷持久化，D8/Q8**）；发布 submitted 后打包 JSON 推 `content-review` topic；**契约单源 REQ-CPM-2 且带可再生 file_url + version（D7/Q7 REVISION）**；**停用 content_posts 的 Redis `moderation:task:queue` 推送，只推 Kafka（D3/Q3：lostfound/user 等其他来源仍走 Redis）**；**推送 at-least-once（落库待推标记 + 定时重推，推送失败=审核盲区风险显式登记 + pending-push 可观测指标，D20 REVISION）**
6. **moderation-service 扩展消费者（后期开发，本期只定契约 + 推送，D16）**：文字先关键字后大模型、图片/pdf 走大模型；结果回写正文→content_posts.status、附件→review_status；**本期不实现消费者，submit 即隐式通过置 status=approved + published_at=NOW()（无消费者也可见，REVISION）**；**同步改 moderation-service Redis 消费者对 content_posts 不再回调 NoticeService（跳过 source_type="notice"，D4/Q4 REVISION）**
7. **两阶段发布（D9/Q9 edit-recount-push，REVISION）**：`draft` 草稿可编辑（可删）→ `submitted` 提交后不可编辑但可删 → 审核；**存储映射 = content_posts.status 枚举（draft=0/submitted=1/approved=2/rejected=3/withdrawn=4）；Create 入口 status 字段（draft 默认 / submitted 立即提交）；draft→submitted 由 UpdateContentPost.status=submitted 触发（无独立 Submit RPC）；本期无消费者 → submit 即隐式通过置 status=approved + published_at=NOW()**
8. **通用 RPC 契约（api-proto，D4/Q4 proto-backcompat）**：`CreateNotice` 等**直接改名**为 `CreateContentPost`/`ListContentPosts`/`GetContentPost`/`DeleteContentPost`/`UpdateContentPost`（通用化一次到位，不做兼容别名）；`CreateContentPostRequest` 加 entry `status`（draft/submitted）与 `section_code`；`UpdateContentPostRequest` 承载 draft 编辑 + submit 动作 + is_pinned；`ContentPost` + `ContentPostAttachment`（含 review_status）；多小区 `community_ids`/`division_id` 复用 notice 契约设计；**`UpdateNoticeModerationStatus` 回调 RPC 随 NoticeService 改名移除（D21，本期无回调路径）**
9. **本期只改后端（D10/Q10 frontend-wiring-scope）**：前端各板块展示差异化后续单独做，不接线通用组件
10. **物业管理员保留本小区发布权（D6/Q6 property-admin-exclusion）**：用通用组件发布，**推翻原 notice 设计剔除 property_admin 的 D26**；发布角色集 = grid_worker（多小区）/ community_admin（选社区展开）/ property_admin（本小区）/ committee（本小区），业主只读
11. **本期新实现 GetPublishPermission + GetMarqueeNotices（D5/Q5 publish-perm-marquee-scope）**：代码中本不存在，新实现（复用原 notice 设计契约）
12. **权限种子补齐（REVISION）**：property_admin 保留 421 + grid_worker 授 421 + **撤销 owner/tenant 的 (1,421)/(5,421) 绑定（保留 435/436）** + **421 的 min_verf_level 由现有 0 提升到 2**（行为变更，init_permissions.sql:201-202/252-253）

## 决策日志（Step 2，Step 8 追溯唯一依据）

> 本变更决策 ID 从 D1 重新编号（与内容通用化相关）；**复用的原 notice-multicommunity-publish 决策（D1-D32）以「REUSE:notice-D<n>」标注**，具体语义见原 change `.harness/changes/notice-multicommunity-publish/`。REVISION 新增/修订的决策以 ★ 标注。

| ID | 决策内容（结论） | 依据 |
|----|-----------------|------|
| D1 | **content_posts 保留 title/text/published_at/publisher_id 全字段（Q1 title-retention）**；`content` 更名 `text`，`title` 保留 | Q1 用户拍板 + 设计文档 §2.1 |
| D2 | **存量不迁（Q2 legacy-data-migration）：仅新数据走 content_post_scope**，存量 notices 行不迁移，兼容期保留列 community_id 不写入 | Q2 用户拍板 + 设计文档 §2.1/§2.2 |
| D3 | **停 Redis 只推 Kafka（content_posts）**（Q3/Q8）：content_posts 发布不再 LPUSH `moderation:task:queue`，改推 Kafka content-review topic；**lostfound/user 等其他来源仍走 Redis** | redis-kafka-dual-write 用户拍板 |
| D4 | **Proto 直接改名（Q4 proto-backcompat）**：CreateContentPost 等通用化改名，不做兼容别名；**同步改 moderation-service Redis 消费者对 content_posts 不再回调 NoticeService**（content_posts 不走该回调路径） | Q4 用户拍板 |
| D5 | **本期实现 GetPublishPermission + GetMarqueeNotices（Q5 publish-perm-marquee-scope）**：代码中本不存在，新实现（复用原 notice 设计契约 REQ-PP-1/REQ-NR-3） | Q5 用户拍板 |
| D6 | **物业管理员保留本小区发布权，用通用组件发布（Q6 property-admin-exclusion，推翻原 notice D26 剔除设计）**；发布角色集 = {grid_worker 多小区, community_admin 选社区展开, property_admin 本小区, committee 本小区}，业主只读 | Q6 用户拍板 |
| D7 | **Kafka 契约带可再生 file_url（Q7 kafka-contract-file-url）**：消费者直接拉取附件内容（消费者经 file_id 重生预签名 URL） | Q7 用户拍板 |
| D8 | **Kafka 基建 = docker-compose 单节点 KRaft 模式 + 数据卷持久化（Q8 kafka-infra-form）** | Q8 用户拍板 |
| D9 ★ | **两阶段发布（Q9 edit-recount-push）：draft 草稿可编辑 → submitted 提交后不可编辑但可删 → 审核**。REVISION 落定存储与触发：content_posts.status 枚举含 draft=0/submitted=1/approved=2/rejected=3/withdrawn=4；CreateContentPostRequest.entry status（draft 默认 / submitted 立即提交）；draft→submitted 由 UpdateContentPostRequest.status=submitted 触发（**无独立 Submit RPC**）；本期无消费者 → submit 即隐式通过置 status=approved + published_at=NOW() | Q9 用户拍板 + REVISION 反馈 1/6 |
| D10 | **本期只改后端（Q10 frontend-wiring-scope）：前端各板块展示差异化后续单独做，不接线通用组件** | Q10 用户拍板 |
| D11 | **section_code 板块化**：content_posts 新增 `section_code`（VARCHAR(30)），notice 为首个板块，未来 repair 等 | 设计文档 §1/§2.1 |
| D12 ★ | **content_posts 结构**：保留 title/text/published_at/publisher_id/**is_pinned/role/publisher** + section_code + status + attachment_count + published_at 去 NOT NULL + 弃 community_id 走 scope；保留 moderation_status/moderation_time 兼容期。**role 由 RBAC→发布角色映射派生、publisher 展示名取用户真实档案（禁请求体信任，堵展示名伪造向量）** | 设计文档 §2.1 + REVISION 反馈 4/8/12 |
| D13 | **content_post_scope 多小区关联**（post_id + community_id，复用 notice_scope 模式）；存量不迁（D2） | 设计文档 §2.2 |
| D14 | **content_post_attachments 加 post_id + review_status + file_id + file_type**（附件级审核状态 + 预签名 URL 重生载体） | 设计文档 §2.3 + REUSE:notice-D24 |
| D15 | **attachment_count 审核完整性判定**：已审附件数 == 计数 且 正文通过 → 展示；任一附件 rejected → 不展示（读路径谓词隐藏，status 由审核流写入，读路径不 mutate） | 设计文档 §3.1 + REVISION 结构-6 |
| D16 ★ | **审核消费者后期开发**：本期只实现 Kafka 推送 + 契约，**submit 即隐式通过置 status=approved + published_at=NOW()**（无消费者也可见，替代「status 默认 approved」的含混表述）；消费者上线后按 D27 覆盖 status/published_at | 设计文档 §3.3 + REVISION 反馈 2/7/9 |
| D17 | **Kafka 引入**：docker-compose 单节点 KRaft（D8）+ `content-review` topic；契约带 file_url（D7） | 设计文档 §3.2 + 用户拍板 |
| D18 | **moderation-service 扩展消费者**：文字先关键字后大模型、图片/pdf 走大模型（后期开发，本期只定契约 + 推送） | 设计文档 §3.2 |
| D19 ★ | **UpdateContentPost（draft 编辑 + submit 动作 + is_pinned 置位）+ DeleteContentPost（撤回，soft delete + status=withdrawn）独立契约**；attachment_count 每次附件集合变更同事务重算（不变量，提交时冻结）；submitted/approved 不可编辑（080005 仅 draft 可编辑）；非发布者删除 080002 | REVISION 反馈 5/8/10 + 结构-7/覆盖率 SF-4 |
| D20 ★ | **Kafka 推送 at-least-once**（落库待推标记 + 定时扫描重推，本期实现）；推送失败不阻塞发布（status=approved 可见）但登记「推送失败=该帖永不审核」显式业务风险 + pending-push 可观测指标 | REVISION 反馈 11 |
| D21 ★ | **UpdateNoticeModerationStatus 回调 RPC 随 NoticeService 改名移除**（本期无回调路径；Redis 消费者对 content_posts 不再回调，精确跳过判定 source_type="notice"） | REVISION 覆盖率 I-3 / 清晰度 SF-4 |
| D22 ★ | **权限种子变更补齐**：撤销 rel_role_permission (1,421)/(5,421)（保留 435/436）+ 421 min_verf_level 由现有 0 提升到 2（行为变更） | REVISION 反馈 3 |
| REUSE:notice-D1 | notices 弃用 community_id → content_posts 弃 community_id（兼容期保留列不写入，范围关联单源 content_post_scope） | 复用 notice 设计（用户确认） |
| REUSE:notice-D24 | 附件引用校验载体 = 扩展 FileInfo（confirmed + user_id + file_type）+ 复用 GetFileUrl 链路；file_id 为预签名 URL 重生载体 | 复用 notice 设计（用户确认） |
| REUSE:notice-D23/D27/D30 | 单通知附件 ≤10 个 且 ≤50MB（总量上限，单源 REQ-CPB-6）；published_at 审核锚定（消费者上线后按 D27 覆盖）；published_at 去 NOT NULL 迁移 | 复用 notice 设计（用户确认） |
| REUSE:notice-D19/D25 | 撤回复用 Delete（仅发布者本人，全局生效，附件保留）；Create 后端不幂等（前端防重） | 复用 notice 设计（用户确认） |
| REUSE:notice-D29/D31 | 错误码对齐 + 080005/080006 消歧（目标级解析失败统一 080006；division 展开为空 080005） | 复用 notice 设计（用户确认） |
| REUSE:notice-D3/D12 | GetPublishPermission（can_publish + 可发布角色）、GetMarqueeNotices（≤10 条置顶+15 天）本期新实现（D5）；division→community 授权集解析行为结论化（REVISION，fail-closed 080006，design 阶段单测验证 HOW） | 复用 notice 设计（用户确认）+ REVISION 清晰度-2 |

## 影响范围

| 服务 | 变更类型 | 说明（原行为 → 新行为） |
|------|:---:|------|
| api-proto | 修改 | community/v1：`CreateNotice` 等**直接改名**为 `CreateContentPost`/`ListContentPosts`/`GetContentPost`/`DeleteContentPost`/`UpdateContentPost`（D4）；**`UpdateNoticeModerationStatus` 回调 RPC 移除（D21）**；`Notice`→`ContentPost`（+section_code/status 全生命周期枚举/attachment_count/published_at 语义）、`NoticeAttachment`→`ContentPostAttachment`（+review_status/file_id/file_type）；`CreateContentPostRequest`（+entry status 字段 + section_code、content→text、community_ids/division_id 多小区复用 notice 契约）；`UpdateContentPostRequest`（draft 编辑字段 + submit 动作 + is_pinned）；`GetPublishPermission`/`GetMarqueeNotices` RPC 新契约（D5）；CHANGELOG 登记。**file/v1：`FileInfo` 新增 file_type + confirmed（复用 notice D24）**；错误码 070004/070005 已登记（复用 notice D11） |
| community-hub-service | 修改 | **migration：notices RENAME content_posts + section_code + status 全生命周期枚举 + attachment_count + published_at 去 NOT NULL + 弃 community_id（D12）+ is_pinned/role/publisher 保留列**；新增 content_post_scope 表（D13）；content_post_attachments 加 post_id/review_status/file_id/file_type（D14）；发布 CreateContentPost（多小区 scope + 附件绑定 + 单事务 + section_code + 入口状态 draft/submitted）+ 两阶段 draft/submitted 状态机（D9）+ **UpdateContentPost（draft 编辑 + attachment_count 重算 + is_pinned + submit 动作）+ DeleteContentPost（撤回）** + **Kafka 推送 content-review（at-least-once 待推标记 + 定时重推，D3/D17/D20）** + submit 隐式通过 status=approved + published_at=NOW()（D16）；读 List/Get/GetMarqueeNotices（scope 过滤 + 审核完整性判定 D15）；GetPublishPermission 实现（D5/D6）；`publisher_id`/`role`/`publisher` 从 JWT 实际身份/真实档案派生；**停 content_posts 的 Redis 队列推送（D3）** |
| moderation-service | 修改 | **Redis 消费者对 content_posts 不再回调 NoticeService（跳过 source_type="notice"，D4）+ UpdateNoticeModerationStatus RPC 移除（D21）**；content_posts 走 Kafka content-review topic，本期不实现消费者（D16/D18）；lostfound/user 等其他来源的 Redis 审核流不变（D3） |
| file-service | 修改 | 附件安全（白名单 + 10MB + magic-bytes 两层校验 + 总量上限）复用 notice 设计（REUSE:notice-D4/D5/D23/D24；docx 为唯一 zip 内容特判，REVISION）；FileInfo 扩展 file_type/confirmed（D14 载体）；错误码 070004/070005（已登记） |
| permission-service | 修改 | 种子变更：**property_admin 保留 421（本小区发布权，D6，推翻 notice D26 剔除）+ grid_worker 授 421 + 撤销 owner/tenant 的 (1,421)/(5,421)（保留 435/436）+ 421 置 min_verf_level=2（0→2 行为变更，D22）**；读码 422/423/424/426/427/428 扩展绑定（复用 notice REQ-PP-4 契约）；division→community 授权集解析（行为结论化，design gate 验 HOW，REVISION） |
| master-data-service | 只读复用 | GetResidentialAreasByDivision（division→小区展开，status=1）+ ResolveScopeAncestors（祖先链）供社区管理员发布展开（复用 notice 设计） |
| docker-compose | 修改 | **新增 Kafka 服务（单节点 KRaft 模式 + 数据卷持久化，D8/Q8）** + content-review topic + retention 覆盖消费者上线空窗（D17/REVISION） |
| web/mobile | 不做 | Q10 已拍板：本期只改后端，前端各板块展示差异化后续单独做，不接线通用组件 |
| web/pc | 不做 | Q10 已拍板 |

## 风险评估

- **高：notices 表 RENAME + 通用化重构的破坏性** — `notices` 物理重命名为 `content_posts`，字段 content→text，Proto 直接改名（D4），任何存量调用方（web/mobile 现无 notice 接线，已核实）与被本变更取代的 notice-multicommunity-publish 变更（存档）需同步。缓解：原 notice 变更存档、前端本期不接线（D10）；RENAME 迁移先于功能上线；已核实 web/pc 无 notice 读消费方、web/mobile 本期不改
- **高：Kafka 基建引入的运维与契约风险** — docker-compose 新增单节点 KRaft Kafka，容器网络（172.19.0.0/24）需分配 IP、健康检查、数据卷持久化；content-review topic 契约（含 file_url + version）一旦发布即稳定。缓解：Kafka 本期一并安装（D8）；topic 契约带可再生 file_url（D7）+ version（REVISION）让消费者不依赖存储侧 URL；发布即推（无消费者也不阻塞，submit 隐式通过 D16）
- **高：内容级审核状态模型的正确性** — status（正文）+ review_status（每附件）+ attachment_count（完整性判定）三者一致性：已审附件数==计数 且 正文通过才展示，任一 rejected 不展示（D15）；attachment_count 每次附件集合变更同事务重算（提交时冻结，D19）；读路径谓词不 mutate status（REVISION 与设计文档对齐）。缓解：D15/D19 显式契约 + 读路径统一完整性谓词
- **高：两阶段发布状态机（draft/submitted）与 Kafka 推送时机** — draft 草稿可编辑（不推 Kafka），submitted 提交后不可编辑但可删且触发 Kafka 推送；本期无消费者 → submit 即隐式通过置 status=approved + published_at=NOW()（REVISION 消除 published_at 恒 NULL 矛盾）。缓解：D9/D16 契约 + REQ-CPB 状态机场景穷举（draft→submit→approved、draft 编辑、submitted 禁编辑、submitted 可删、draft 可删、非发布者删除被拒）
- **中：Kafka 推送 at-least-once 的审核盲区** — 推送失败若静默丢弃，未来消费者上线后帖子永无审核记录却一直可见（REVISION 反馈 11）。缓解：D20 落库待推标记 + 定时重推（本期实现）+ pending-push 可观测指标 + 业务风险显式登记（推送失败=该帖永不审核），BACKLOG 回填消费者上线前对 pending-push 积压的处置
- **中：Redis → Kafka 双轨切换的过渡一致性** — content_posts 停 Redis 队列（D3），lostfound/user 等其他来源仍走 Redis；moderation-service Redis 消费者对 content_posts 跳过 source_type="notice"（D4）。缓解：D3/D4 显式契约 + moderation-service 职责边界声明
- **中：权限种子行为变更的既有仓库冲突** — **撤销 owner/tenant 的 421 绑定 + 421 min_verf_level 0→2 是行为变更**（init_permissions.sql:201-202/252-253），与 notice 变更的 D26 回收方向不同（本变更保留 property_admin 421）。缓解：REQ-CPP-3 显式列出完整种子变更集 + 场景（撤销后 level-2 业主直调创建被拒 080002）；冲突预检已跑（check-change-conflict.sh：无变更冲突）
- **中：存量通知迁移后不可见（业务可用性回归，已拍板）** — D2 存量不迁 → 上线瞬间存量通知从列表/详情/跑马灯全部消失（无 content_post_scope 行，读路径不返回）。缓解：proposal/REQ-CPB-1 显式登记影响面 + BACKLOG 回填「存量通知迁移回填」项；本期功能先跑通新数据
- **中：错误码 / 跨服务字段一致性回归** — 080005/080006 消歧（REUSE:notice-D31）、070004/070005 已登记（REUSE:notice-D11）；沿用 `[[snake-camel-field-mismatch]]` 记忆；community_ids/division_id 等 int64 显式 `[jstype=JS_STRING]`；QA 自动检查
- **中：与进行中变更服务重叠** — 冲突预检（check-change-conflict.sh）检出冲突：与 **notice-multicommunity-publish** 16 处 C1/C2 重叠（**属取代关系，预期**：B 方案本变更取代/存档原通知变更，revises 已列原 change 文档）；与 rel-user-role-migration-publish-fix（permission-service 服务重叠 + init_permissions.sql C2 文件重叠）、spec-pipeline-e2e-l（api-proto/permission-service）、test-pipeline-work-records（master-data-service）为**与归档 notice 变更一致的既有重叠**（原 notice proposal 已登记同批重叠）。缓解：proto 全走新增字段/RPC 兼容路径 + 直接改名（D4，前端本期不改故破坏面收敛）；种子变更仅动 421 关联 + property_admin 保留 + grid_worker 授 421 + 撤销 (1,421)/(5,421) + 置 min_verf_level=2；合并前各变更走 `make ci` breaking-check 交叉验证

## 不做清单（Won't have — 本轮明确不实现）

- **不做前端接线（D10/Q10）**：web/mobile 本期不改，前端各板块展示差异化后续单独做，不接线通用组件
- **不做 Kafka 审核消费者实现（D16/D18）**：moderation-service 扩展消费者（文字先关键字后大模型、图片/pdf 走大模型）后期开发，本期只定契约 + 推送 + at-least-once 重推；submit 隐式通过（无消费者也可见）
- **不做存量数据迁移（D2/Q2）**：存量 notices 行不迁移，仅新数据走 content_post_scope；存量 attachment review_status/file_id 不回填；「存量通知迁移后不可见」影响登记 BACKLOG
- **不做 Proto 兼容别名（D4）**：CreateNotice 等直接改名，不做旧名兼容转发（一次性破坏，前端本期不改故破坏面收敛）
- **不做 PC 端**（Q10）
- **不做小区/村层级统筹（task-016）**：不阻塞本变更
- **不做附件病毒扫描 / 图片处理 / CDN**：magic-bytes 嗅探仅用于类型白名单判定（复用 notice 边界；docx 为唯一 zip 内容特判）
- **不做批量操作**：发布/撤回一次一条
- **不做缓存层**：读路径靠 content_post_scope 索引兜底
- **不做 Redis 审核队列的移除**：`moderation:task:queue` 对 content_posts 停推（D3），但 lostfound/user 等其他来源仍走 Redis，队列机制保留；不做物理清理
- **不做 moderation_service Redis 消费者的删除**：仅「对 content_posts 不再回调 NoticeService」（D4），lostfound/user 消费逻辑保留
- **不做事务性 outbox 中间件**：Kafka 补偿以落库待推标记 + 定时重推实现（D20），不引入独立 outbox 组件

## 验收标准

- [ ] migration 后：`notices` RENAME 为 `content_posts`，含 title/text/published_at/publisher_id/is_pinned/role/publisher/section_code/status/attachment_count；published_at 可空（去 NOT NULL）；community_id 弃用保留列；content_post_scope 新表（post_id+community_id 双 NOT NULL + 复合 PK + 读索引）；content_post_attachments 含 post_id/review_status/file_id/file_type
- [ ] 两阶段发布：Create entry draft（可编辑可删）→ UpdateContentPost.status=submitted 提交（不可编辑但可删）→ submit 隐式通过 status=approved + published_at=NOW()；submitted/approved 编辑被拒（080005）、非发布者删除被拒（080002）
- [ ] 发布 CreateContentPost（section_code=notice，多小区 community_ids 或 division_id，entry draft/submitted）：正文 + 附件单事务落库；Kafka content-review topic 收到符合 REQ-CPM-2 契约（含 file_url + version）的消息；**不再 LPUSH Redis moderation:task:queue**（content_posts）；Kafka 不可用时发布不阻塞但进入 pending-push 由重推扫描补投
- [ ] 附件集合变更（Create/draft 编辑）attachment_count 同事务重算并在提交时冻结（无陈旧计数致帖永不可见）
- [ ] 本期 submit 隐式通过 status=approved、附件 review_status 默认 approved → 无消费者也可见
- [ ] 审核完整性判定：正文 approved 且 已审附件数==attachment_count 且 无 rejected → 展示；任一附件 rejected → 不展示（读路径统一谓词，不 mutate status）
- [ ] 物业管理员 property_admin 保留 421，可发布本小区（GetPublishPermission can_publish=true 含 property_admin 角色，D6）；grid_worker 多小区、community_admin 选社区展开、committee 本小区；**撤销后 owner/tenant 直调创建被拒 080002（含 level-2 业主，D22）**
- [ ] GetPublishPermission / GetMarqueeNotices 新 RPC 可用（D5）；跑马灯 ≤10 条置顶优先 + 15 天窗口非空（本期 published_at submit 即置值）
- [ ] 附件白名单外拒绝（070004）、单文件 >10MB 拒绝（070005）、zip/rar 扩展名拒绝（docx 内容层特判放行）；magic-bytes 回读拦截改名上传；doc/docx 容器签名放行；单帖附件 ≤10 个 且 ≤50MB（单源 REQ-CPB-6）
- [ ] moderation-service Redis 消费者对 content_posts 跳过 source_type="notice" 不再回调 NoticeService（D4/D21）；lostfound/user 审核流不变（D3）
- [ ] docker-compose 启动 Kafka（单节点 KRaft + 数据卷持久化），content-review topic 可投递（D8/D17）
