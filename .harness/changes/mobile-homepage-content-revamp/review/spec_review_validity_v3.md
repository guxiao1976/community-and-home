# Plan Review — mobile-homepage-content-revamp（业务有效性视角）

**审查维度**: 业务自洽 / 非功能 / 合规 / 架构冲突 / 技术债 / 依赖风险
**审查版本**: P1.3 fallback:r2:rc1（与历史 r0/r1 哈希不同 → 按磁盘最新内容独立重新审查，未沿用旧轮结论）
**审查对象**: `proposal.md` + `specs/{notice-time-window,notice-detail-preview,function-entries,contact-list-page,homepage-layout}/spec.md`
**审查者**: 业务有效性 Reviewer（代码级验证：api-proto / REST 层 / RPC 层 / model / migration / file-service magic / web-mobile 页面）

## 摘要

- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 2

本轮（r2:rc1）对磁盘最新内容做代码级独立复核：前两轮 validity 的 MUST FIX（v1 MF-1 附件预览主路径、v1 MF-2 窗口索引、v1 MF-3 跑马灯后端无效变更、v2 MF-1 `file_type image/*` MIME 谓词）与 v2 SHOULD FIX（SF-1 整单失败边界、SF-2 created_at 回退）**已修复且方向与既有 wire 一致**；本轮 r2 修订（r2-1~r2-6）经逐条代码验证属实且修复正确。未发现新的 MUST FIX。**VERDICT: APPROVED**。

## 上一轮 MUST/SHOULD 修复验证（本轮代码级独立复核）

| 上轮 # | 内容 | 验证结果 |
|--------|------|---------|
| v1 MF-1 | 附件预览主路径直连 file-service REST 被拒（所有权） | ✅ 已修复。REQ-NDP-2/3 主路径改为消费详情响应重生 `file_url`；代码核实 `rpc/internal/logic/notice/getcontentpostlogic.go` toProtoAttachments 经 RPC GetFileUrl（无所有权限制）重生、file_id=0 兼容期回退 stored URL；file-service REST GetFileUrl 强制所有权（`rpcResp.File.UserId != userId`）属实 |
| v1 MF-2 | 30 天窗口过滤索引/性能未评估 | ✅ 已修复。REQ-NTW-6 新增；代码核实 idx_published 经 003 RENAME 随 notices→content_posts 迁移，但 `content_posts.community_id` 已 MODIFY 为弃用 NULL 列、无法服务 scope JOIN 后的 published_at 过滤/排序——spec 主张属实；新索引 + EXPLAIN 验收点已写入 |
| v1 MF-3 | GetMarqueeNotices 15→30 天无效变更 | ✅ 已修复。REVISION #6 删除后端常量变更；代码核实 `notice.vue` marqueeText 为 computed 自 `notices`（getNoticeList→ListContentPosts 派生），移动端无 GetMarqueeNotices 消费方 |
| v2 MF-1 | REQ-NDP-2 `file_type ∈ image/*`（MIME）与 wire 相悖 | ✅ 已修复。REQ-NDP-2/3/4 统一为扩展名白名单 `{png, jpg, jpeg, gif}`；代码核实 file-service `internal/guard/magic.go` SniffType 返回规范小写扩展名（png/jpg/gif/pdf/doc/docx），非 MIME，`image/*` 在 wire 不存在——修复方向正确 |
| v2 SF-1 | 附件重生失败模式（整单失败 vs 逐附件降级） | ✅ 已修复。r2-6 明确「GetFileUrl 失败 → GetContentPost 读整单失败 → REST 透错 → REQ-NDP-1 详情加载失败态」；代码核实 toProtoAttachments 对任一附件 GetFileUrl 失败返回 (nil, err)——属实 |
| v2 SF-2 | REQ-NTW-5 created_at 回退在窗口页不可达 | 🟡 未修复（非阻塞，见 SHOULD FIX 2） |

## 本轮 r2 修订逐条代码验证

| r2 # | 内容 | 验证结果 |
|------|------|---------|
| r2-2 | `since_days` REST 层透传 | ✅ 属实。`api/internal/types/types.go` ListContentPostsReq 当前仅有 community_id/role/section_code/page/page_size（无 since_days）；`api/internal/logic/notice/listcontentpostslogic.go` 构造 RPC 请求未透传 since_days——仅改 RPC/proto 会致移动端 30 天窗口静默失效的论断成立，补入 revises 方向正确 |
| r2-5 | REQ-NTW-2 校验措辞去掉 non-numeric | ✅ 属实。int32 wire 值恒为数字，非数字由 REST 网关解析层拒绝；服务端仅校验数值范围，口径一致 |
| r2-6 | 附件重生整单失败边界归入 REQ-NDP-1 | ✅ 属实。见上表 v2 SF-1 |
| r2-1/3/4 | file_type 扩展名白名单统一 | ✅ 属实地修复；唯一残留为白名单含 `jpeg`（file-service SniffType 从不产出该值），见 SHOULD FIX 1 |

## 发现

### 🔴 MUST FIX

无。

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 1 | `specs/notice-detail-preview/spec.md` REQ-NDP-2/3/4 | 图片白名单 `{png, jpg, jpeg, gif}` 含 `jpeg`，但 file-service `SniffType`（`internal/guard/magic.go`）对 JPEG 一律返回规范扩展名 `jpg`，wire 上不存在 `jpeg` 值——`jpeg` 为死条目；spec 同时声称该白名单即「file-service 嗅探落库的白名单」，与事实轻微不符。业务影响为零（file-service 产出的全部图片扩展名 png/jpg/gif 均在白名单内，图片预览路径不失效），属规范一致性问题 | 将白名单收敛为 file-service 真实规范扩展名 `{png, jpg, gif}`，或将 `jpeg` 显式标注为防御性条目（如「兼容历史脏数据」），避免实现者误以为 file-service 可能产出 jpeg |
| 2 | `specs/notice-time-window/spec.md` REQ-NTW-5 场景 2 | created_at 回退场景在窗口页（首页/列表）不可达：REQ-NTW-1/4 已排除 published_at 为 NULL 的行出窗口，窗口页渲染的卡片全部 published_at 有值；「published_at NULL 但 created_at 有值 → 卡片显示 created_at」仅详情页/深链可达。给实现者「窗口页可能出现 NULL published_at 卡片」的错误预期（v2 SF-2，仍未修复） | 将 REQ-NTW-5 时间回退收敛到「详情页渲染（深链场景）」，窗口页场景删除或标注为防御性回退、不可触发 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | `community_contacts`（001 + 004）表缺 `deleted_at` 列，偏离项目编码规范 §3.1「全链路 created_at/updated_at/deleted_at」。属 001 既有技术债，004 为保证「001 单源」必须对齐（补 deleted_at 反而违反 REVISION 对齐约束）；建议记入技术债条目，后续如需软删单独处理。代码核实：CommunityContactModel.FindByCommunityId 查询 `where community_id = ? order by sort_order asc`，不引用 deleted_at，004 对齐 DDL 后读路径安全 |
| 2 | 「便民联络做实」落地后为空态上线（migration 004 不预置种子，D4），运行库 community_contacts 为空 → 新联络列表页首屏「暂无联络信息」。属用户拍板行为，建议在 design/上线说明标注「运营需预置联络数据」，避免「做实入口点开是空页」被误判为缺陷（v2 INFO 1 复述） |

## 关键代码级核验清单（本轮）

- ✅ `api-proto ListContentPostsRequest` 现字段 1-5（community_id/role/section_code/page/page_size），`since_days`=字段 6 additive，与「additive 字段 6」一致
- ✅ `FindListByCommunity`：JOIN content_post_scope 按 community_id 过滤 + status=approved + is_pinned desc/published_at desc；窗口谓词可加于 WHERE，NULL published_at 天然被排除（NULL 比较恒 false），与 REQ-NTW-1/2 边界一致
- ✅ migration 001 `community_contacts` DDL 列（id/community_id/category/name/phone/sort_order/created_at/updated_at + idx_community）与 REQ-CLP-2 完全一致
- ✅ file-service `SniffType` 扩展名集合 {png, jpg, gif, pdf, doc, docx}，无 MIME、无 jpeg/webp
- ✅ `notice.vue` 跑马灯由首页通知列表派生（marqueeText computed from notices）；`notice-browse.vue` 当前「请求 50 + 客户端 3 个月过滤」属实，将被服务端窗口替代
- ✅ `web/mobile/src/api/community.ts` NoticeAttachment 仅 id/file_name/file_url/file_size，缺 file_id/file_type（REQ-NDP-4 依据属实）
- ✅ community-hub 业务错误码 080005（参数无效）存在
- ✅ 当前 `notice-detail.vue` onDownload 对全部附件一律 downloadFile+openDocument（无图片预览）——REQ-NDP-2 为新增能力，legacy/缺失 file_type 走文档分支与现状一致，无回归

## 问题跟踪表

| # | 状态 | 备注 |
|---|------|------|
| 上轮 v1 MF-1/2/3、v2 MF-1 | 已修复 | 本轮代码级验证通过（附件预览主路径 / 窗口索引 / 跑马灯 / file_type 谓词） |
| 上轮 v2 SF-1 | 已修复 | 整单失败边界 r2-6 已解决 |
| 上轮 v2 SF-2 | 待修复（非阻塞） | REQ-NTW-5 created_at 回退收敛到详情页（本轮 SHOULD FIX 2） |
| 本轮 SHOULD FIX 1 | 待修复（非阻塞） | file_type 白名单收敛到 file-service 真实扩展名 |
| 本轮 INFO 1 | 记技术债 | community_contacts 缺 deleted_at（001 既有债务） |

---
VERDICT: APPROVED
---
