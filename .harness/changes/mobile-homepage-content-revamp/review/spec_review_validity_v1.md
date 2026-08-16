# Plan Review — mobile-homepage-content-revamp（业务有效性视角）

**审查维度**: 业务自洽 / 非功能 / 合规 / 架构冲突 / 技术债 / 依赖风险
**审查版本**: fallback:r0:rc1（spec 磁盘最新内容，独立审查）
**审查对象**: `proposal.md` + `specs/{notice-time-window,notice-detail-preview,function-entries,contact-list-page,homepage-layout}/spec.md`

## 摘要

- 🔴 CRITICAL(MUST FIX): 1 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 2

审查过程对 proposal 与 5 份 spec 的既有事实主张做了代码级验证（backend logic / model / migration / proto / REST wire / runtime DB），结论：大部分现状描述准确（缺 community_contacts 表属实、跑马灯 15 天常量属实、广告位位置属实、附件 REST wire 含 file_id/file_type 属实），但 **REQ-NDP-2/REQ-NDP-3 附件预览机制存在与既有 wire 语义相悖且对目标用户不可用的设计缺陷**，必须修订。

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | `specs/notice-detail-preview/spec.md` REQ-NDP-2 / REQ-NDP-3 | **附件预览主路径对目标用户必然失败，且与既有 wire 语义相悖（file_id 权威、file_url 回退 与事实相反）**。<br>证据链：<br>(a) spec 要求前端以附件 `file_id` 为权威键直连 file-service REST `GET /api/files/:id`（GetFileUrl）取签名 URL，`file_url` 仅回退。<br>(b) 但 file-service REST GetFileUrl **强制文件所有权**（`api/internal/logic/file/getfileurllogic.go`：`rpcResp.File.UserId != userId → "permission denied: not the file owner"`）。通知附件由发布方（社区管理员/业委会/物业）上传，查看通知的小区成员非文件所有者 → 主路径必然被拒。<br>(c) 新发布通知的附件落库 `file_url` 为**占位空串**（`rpc/internal/logic/notice/helper.go` bindAttachments `FileUrl: ""`）→ 回退路径也无可用 URL → 两路皆败 → 「附件打开失败」。（旧帖仅靠回退 stored file_url，恰是本次变更要消除的脆弱链路。）<br>(d) 实际 wire 中 community-hub `GetContentPost` RPC 读路径**已**经 `FileClient.GetFileUrl`（RPC 层无所有权限制）在服务端把 `file_url` 重生为有效预签名 URL（`rpc/internal/logic/notice/getcontentpostlogic.go` toProtoAttachments），REST 薄代理原样透出（`api/internal/logic/notice/helper.go`）。即移动端详情响应里的 `file_url` 本身就是权威新鲜签名 URL。 | 将 REQ-NDP-2/3 主路径改为**直接消费详情响应中已由 community-hub 服务端重生的 `file_url`**（file-service 生命周期一致、无所有权问题、零新增后端变更）；`file_id` 定位为服务端重生键（前端不直连 file-service）。若坚持前端直连 file-service 取 URL，则需**新增后端能力放开非所有者附件读**（与「file-service 复用、无契约变更」的承诺冲突，需同步修订 proposal 影响范围）。同时 REQ-NDP-3 场景「file_id 存在时以签名 URL 打开文档」当前无法通过，需按上述方式改写或删除。 |
| 2 | `specs/notice-time-window/spec.md` REQ-NTW-2 / `proposal.md` §影响范围 | **30 天窗口过滤的索引/性能未评估**。`FindListByCommunity` 驱动表为 `content_post_scope`（idx_scope_community），新增 `published_at >= ?` 过滤落在 `content_posts` 上，且排序仍按 `is_pinned desc, published_at desc`；001 的 `idx_published(community_id, published_at DESC, deleted_at)` 中 community_id 为弃用列（NULL），无法直接服务按 scope 过滤后的 published_at 排序/过滤。 | 设计阶段为 content_posts 增加面向 (status, published_at) 或经 scope JOIN 后可用的 published_at 索引，并验证 30 天窗口过滤 + 倒序排序不走全表扫描；在 spec/proposal 增加性能验收点。 |
| 3 | `specs/homepage-layout/spec.md` / `specs/notice-time-window/spec.md` REQ-NTW-3 / `proposal.md` D1 | **后端 GetMarqueeNotices 15→30 天变更为无效/冗余变更**。移动端首页跑马灯文案由首页通知列表（`getNoticeList` → ListContentPosts）派生（`notice.vue` marqueeText computed from notices），并未调用 GetMarqueeNotices；全仓未发现其他消费方（PC 无）。后端常量 15→30 的改动对移动端跑马灯无实际效果，纯属误导性工作量。 | 确认 GetMarqueeNotices 是否有其他消费方；若无，删除该后端变更或明确标注为「对齐备用 RPC 常量」，避免实现者误以为跑马灯由该 RPC 驱动。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 1 | `specs/notice-detail-preview/spec.md` REQ-NDP-2/3 | 依赖附件的 `file_id`/`file_type` 字段，但移动端 `api/community.ts` 的 `NoticeAttachment` 类型当前**只有** id/file_name/file_url/file_size，无 file_id/file_type（后端 REST wire 有，前端类型未扩展）。spec 未显式要求类型扩展。 | 在 REQ-NDP 能力中显式声明前端 `NoticeAttachment` 类型须补充 `file_id`/`file_type` 字段（snake_case 对齐），否则实现者无字段可消费。 |
| 2 | `specs/notice-time-window/spec.md` REQ-NTW-4 | 列表页排序只写「published_at 倒序」，未提置顶优先；与 REQ-NTW-1（窗口内置顶保留优先级）及后端 `is_pinned desc, published_at desc` 既有契约不完全一致。 | 统一措辞为「窗口内置顶优先、其余按 published_at 倒序」，与首页口径及后端契约一致。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | 各 spec 服务职责边界对「前端传 since_days 参数」的 REST 形参名（`since_days`）与错误码 080005 已明确，验证通过；建议在 proposal 影响范围同步标注 REST `ListContentPostsReq` 需新增 `since_days` form 参数映射。 |
| 2 | 首页「便民联络」由内嵌网格改为 4 功能入口之一 + 跳转联络列表页（D3 合并落地）属行为变更，proposal 已作补充解释并在评审提示中声明；业务自洽，建议在 design 阶段保留该说明避免实现走样。 |

## 问题跟踪表

| # | 状态 | 备注 |
|---|------|------|
| MF-1 | 待修复 | 附件预览机制重定向到既有重生 file_url |
| MF-2 | 待修复 | 30 天窗口过滤索引评估 |
| MF-3 | 待修复 | 跑马灯后端变更确认/删减 |
| SF-1 | 待修复 | 前端附件类型扩展声明 |
| SF-2 | 待修复 | 列表页排序措辞统一 |

---
VERDICT: REVISION
---
