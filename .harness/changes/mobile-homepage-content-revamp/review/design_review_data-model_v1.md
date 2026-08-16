# Design Review — mobile-homepage-content-revamp（data-model + interface-proto 视角）

**审查维度**: 数据模型 / 服务归属 / 接口契约 / Proto 破坏性 / 依赖顺序
**审查对象**: `.harness/changes/mobile-homepage-content-revamp/design.md` + `tasks.md`
**审查者**: 设计评审 Agent（data-model 视角）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 4

## 验证方法

将 design.md 全部数据模型/接口/Proto/流程声明与实际代码、Proto、Migration、前端源码逐项对照核验（不审查代码实现质量，只审设计正确性）。

## 核验结论（逐项，均通过）

| 设计声明 | 实际核验 | 结果 |
|---------|---------|:---:|
| `ListContentPostsRequest` 现占字段 1-5（community_id/role/section_code/page/page_size），`since_days = 6` 为下一可用号 | `api-proto/api/community/v1/community.proto` L115-121 | ✅ |
| 非法 since_days 复用 `080005`（CodeInvalidParam=80005），不新增错误码 | `api/internal/types/types.go` L20；`rpc/internal/logic/scope/scope.go` L30 | ✅ |
| REST `GET /api/community/notices` / `GET /api/community/contacts` / `GET /api/community/notices/:id` 路由存在 | `api/internal/handler/routes.go` L21/57/42 | ✅ |
| REST `ListContentPostsLogic` 目前**未**检查 `resp.GetBase()`，须补 `responsex.ToError`（REVISION r2-2 结论成立） | `api/internal/logic/notice/listcontentpostslogic.go` L23-49（无 Base 检查）；`getcontentpostlogic.go` L60 为既有 ToError 模式；`common/pkg/responsex/grpc.go` ToError 成功返回 nil | ✅ |
| RPC `ListContentPosts` 走 scope.FilterAllowed + FindListByCommunity（读过滤 GLOBAL/LIMITED/EMPTY，越权空列表不泄露） | `rpc/internal/logic/notice/listcontentpostslogic.go` L33-46 | ✅ |
| GetContentPost 附件 file_url 经 RPC GetFileUrl 重生；任一附件重生失败 → 读整单失败（r2-6 结论成立） | `rpc/internal/logic/notice/getcontentpostlogic.go` L80-84、L95-118（toProtoAttachments 返回 nil,err） | ✅ |
| `community_contacts` DDL 与 001 / CommunityContactModel 完全对齐（8 列 + idx_community + InnoDB/utf8mb4） | `migration/001_initial.sql` L39-49 逐列比对一致；`model/community_contact.go` 字段一致 | ✅ |
| `community_contacts` 无 deleted_at 为既有显式偏离（硬删除 Upsert 语义） | `services/community-hub-service/docs/design.md` §设计决策 4 | ✅ |
| `content_posts.community_id` 为弃用 NULL 列，既有 `idx_published(community_id, published_at DESC, deleted_at)` 无法服务 scope JOIN 后过滤/排序（REVISION #5 结论成立） | `model/content_post.go` L33 `CommunityId *int64` 弃用；`migration/003` L22-23 置 NULL 不写入；001 L24 索引定义 | ✅ |
| `content_post_scope` 含 `idx_scope_community(community_id, post_id)` | `migration/003` L38 | ✅ |
| 模型 `FindListByCommunity` 改变参 `opts ...ContentPostListOption` 兼容既有调用方 | 现签名 `(ctx, communityId, sectionCode, role string, offset, limit int64)` L67——变参为兼容扩展 | ✅ |
| Contact proto id/community_id 带 `jstype=JS_STRING`（Snowflake 规范） | `community.proto` L222-223 | ✅ |
| ContentPostAttachment 含 `file_id=6 [jstype=JS_STRING]`、`file_type=5`，前端 NoticeAttachment 扩展字段可消费 | `community.proto` L93；`web/mobile/src/api/community.ts` L10-15（现缺 file_id/file_type，为 additive 扩展） | ✅ |
| 首页现状：内嵌联络网格（getContacts）、2 广告在联络下 + 1 广告在寻失下 | `notice.vue` L120-152/L198-210/L349 | ✅ |
| 列表页现状：客户端 3 个月过滤 + 翻页，需移除 | `notice-browse.vue` L112-114、L48-60 | ✅ |
| `GetMarqueeNotices` 移动端不消费、不修改 | `notice.vue` 无 getMarquee 导入 | ✅ |
| migration 004 对齐 .change.yaml 声明 | `.change.yaml` L20 | ✅ |
| file-service `GetFileUrl` RPC 重生预签名 URL（无所有权限制），前端不直连 `GET /api/files/:id` | `getcontentpostlogic.go` L100（RPC 侧调用）；REVISION #4 结论与 wire 一致 | ✅ |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX
| # | 文件:章节 | 问题 | 建议 |
|---|-----------|------|------|
| 1 | tasks.md Task 3.2 EXPLAIN 验收 | EXPLAIN SQL 为简化版（`WHERE scope.community_id=? AND status=2 AND published_at>=? AND published_at<=? ORDER BY is_pinned DESC, published_at DESC`），未含真实查询的 `deleted_at is null` + 附件完整性相关子查询；且验收判据「content_posts 无全表扫描」由 `idx_scope_community` 驱动 + PK 回表的 join 计划也能满足——即使新索引 `idx_status_pinned_published` 未被选中，EXPLAIN 仍可通过，新索引形同虚设，ADR-3「减少 filesort」意图不可验证 | 用真实生成 SQL（含 deleted_at 与附件子查询）跑 EXPLAIN，并增加判据：`possible_keys` 含 `idx_status_pinned_published`，或 `Extra` 无 `Using filesort` |
| 2 | design.md §接口设计 + REQ-NTW-2 | REQ-NTW-2 字面「有效范围 1..365，值 ≤0 或 >365 拒绝」与设计「`since_days==0` 缺省不过滤」冲突。设计按 0=不过滤 处理是**正确且必要**的（REST `form:"since_days,optional"` 缺省即 0，若 0 被拒则未传参的 PC 调用方全部 080005，破坏向后兼容），但 spec 措辞与设计不一致会致下游实现歧义 | 同步修订 REQ-NTW-2 措辞为「<0 或 >365 拒绝；缺省/显式 0 = 不过滤」，消除歧义 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | `design.md §记忆引用` 自验声明「文件均在 `.harness/knowledge/memory/`」不准确：`[[snake-camel-field-mismatch]]` 不在项目记忆树（仅存在于用户级 auto-memory）。引用本身有效（触发词与场景匹配），建议修正自验措辞 |
| 2 | 前端图片白名单 `{png, jpg, jpeg, gif}` 与 file-service `SniffType` 实际产出 `{png, jpg, gif, pdf, doc, docx}` 非完全对齐：`jpeg` 不会出现在 wire（.jpeg 上传被 magic 嗅探为 jpg，`guard/magic.go` L34 返回 "jpg"），白名单为无害超集。`design.md §安全考虑`「与 guard/magic.go 对齐」表述宜改为「对齐且含防御性超集条目」 |
| 3 | `.change.yaml` revises 登记了 `004_add_community_contacts.sql` 但未登记新增的 `005_content_posts_window_index.sql`，变更清单不完整 |
| 4 | design.md §数据模型 用 `status=approved`（语义）、tasks.md Task 3.2 EXPLAIN 用 `status=2`（数值，TINYINT 枚举）——二者等价，建议 005 索引注释统一标注 status 枚举（2=approved）避免歧义 |

## 服务归属 / 任务粒度检查（模式一.5 刚性校验）
- 单任务不跨服务：0.x=api-proto、1.x=community-hub、2.x=web/mobile、3.x=Owner 运维，边界清晰 ✅
- 数据模型变更（1.1/1.2 migration、1.3 model）、业务逻辑（1.4/1.5 logic）、前端（2.x）拆分独立 ✅
- Proto 变更独立成任务（0.1-0.3）、Migration 独立成任务（1.1/1.2）✅
- 单任务测试用例数 ≤5（1.4 明示 ≤5），未超 1~10 刚性上限 ✅
- 依赖顺序：Proto（0.x）→ community-hub（1.x）→ web/mobile（2.x）→ Owner 运维（3.x），1.5 依赖 0.1/0.3 生成代码已声明 ✅

---
VERDICT: APPROVED
---
