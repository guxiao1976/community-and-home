# Plan Review — mobile-homepage-content-revamp（业务有效性视角）

**审查维度**: 业务自洽 / 非功能 / 合规 / 架构冲突 / 技术债 / 依赖风险
**审查版本**: P1.3 fallback:r1:rc1（与历史 r0:rc1 哈希不同 → 按磁盘最新内容独立重新审查，未沿用旧轮结论）
**审查对象**: `proposal.md` + `specs/{notice-time-window,notice-detail-preview,function-entries,contact-list-page,homepage-layout}/spec.md`
**审查者**: 业务有效性 Reviewer（代码级验证：backend logic / model / migration / proto / REST wire / file-service / mobile 页面）

## 摘要

- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 2

上一轮（r0:rc1）validity 的 3 项 MUST FIX（MF-1 附件预览、MF-2 窗口索引、MF-3 跑马灯后端变更）与 2 项 SHOULD FIX（SF-1 附件类型扩展、SF-2 列表排序措辞）经代码级验证**全部修复且方向正确**。本轮重新审查磁盘最新内容，发现 **1 项新增 MUST FIX**：REQ-NDP-2 图片分发键 `file_type ∈ image/*` 与实际 wire 格式（嗅探扩展名，如 `jpg`/`png`，非 MIME）相悖，且与 REQ-NDP-4 场景（`file_type: "jpg"`）自相矛盾——按字面实现将导致图片全屏预览路径对全部附件失效。

## 上一轮 MUST/SHOULD 修复验证（P1.3 独立复核）

| 上轮 # | 内容 | 验证结果 |
|--------|------|---------|
| MF-1 | 附件预览主路径重定向到详情响应重生 file_url | ✅ 已修复。`rpc/internal/logic/notice/getcontentpostlogic.go` toProtoAttachments 经 RPC GetFileUrl（无所有权限制）重生 file_url，file_id=0 兼容期回退 stored URL；file-service REST GetFileUrl 确认强制所有权（`rpcResp.File.UserId != userId → permission denied`）。REQ-NDP-2/3 已改为消费响应 file_url、前端不直连 REST。方向与既有 wire 一致。 |
| MF-2 | 30 天窗口过滤索引/性能 | ✅ 已修复。REQ-NTW-6 新增。迁移 003 确认 `content_posts.community_id` 已 MODIFY 为弃用 NULL 列，idx_published(community_id, published_at DESC, deleted_at) 无法服务 scope JOIN 后的 published_at 过滤/排序——spec 主张属实；索引补充 + EXPLAIN 验收点已写入 REQ-NTW-6 与 proposal 验收标准。 |
| MF-3 | 跑马灯后端 GetMarqueeNotices 无效变更 | ✅ 已修复。`notice.vue` marqueeText 为 `computed` 自 `notices`（getNoticeList→ListContentPosts）派生，移动端无 GetMarqueeNotices 消费方；REQ-NTW-3 已改为同源派生、GetMarqueeNotices 不修改（REVISION #6）。 |
| SF-1 | 前端 NoticeAttachment 扩展 file_id/file_type | ✅ 已修复。REQ-NDP-4 显式声明；`web/mobile/src/api/community.ts` NoticeAttachment 当前仅 id/file_name/file_url/file_size（无 file_id/file_type）——spec 主张属实。 |
| SF-2 | 列表页排序措辞统一（置顶优先） | ✅ 已修复。REQ-NTW-4 明确「置顶优先(is_pinned DESC) + published_at DESC」，与 REQ-NTW-1/后端 `is_pinned desc, published_at desc` 契约一致。 |

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | `specs/notice-detail-preview/spec.md` REQ-NDP-2 / REQ-NDP-4 | **图片/文档分发键与实际 wire 格式相悖且 spec 内部自相矛盾**：REQ-NDP-2 规范文本与场景写「`file_type` ∈ image/*」（MIME 前缀语义），REQ-NDP-4 场景写 `file_type: "jpg"`。实际 wire 中 `file_type` 是 file-service 嗅探落库的**规范扩展名**（迁移 003 注释「文件类型（扩展名，自 FileInfo 回读）」；`file-service/rpc/internal/logic/file/helper.go`「返回嗅探的规范扩展名（用于落库 file_type）」；confirmupload 测试断言 `FileType = "png"`/`"jpg"`），不是 MIME。若实现者按 REQ-NDP-2 字面写 `file_type.startsWith('image/')`，该判断对所有附件恒 false → 图片附件永远走文档打开路径，「图片全屏预览」主路径对全部附件失效（业务不可用）。 | 将 REQ-NDP-2 规范文本与场景统一为「file_type 为 file-service 白名单扩展名（如 jpg/jpeg/png/gif/webp）」，图片分发按**扩展名集合**（小写）判定而非 MIME 前缀；同时同步 REQ-NDP-3/REQ-NDP-4 措辞（`file_type: "jpg"` 即正确示例，仅需把 REQ-NDP-2 的 `image/*` 改为扩展名集合）。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| 1 | `specs/notice-detail-preview/spec.md` REQ-NDP-2/3 | 附件「file_url 不可用」场景假设**逐附件**降级，但后端实际失败模式是**整单失败**：`getcontentpostlogic.go` toProtoAttachments 对任一附件 GetFileUrl RPC 失败（文件被删/file-service 不可用）返回 (nil, err)，GetContentPost 整体报错 → REST 透错 → 详情页显示「通知不存在」（现有行为，非本次变更引入）。因此 REQ-NDP-2/3 的「file_url 为空/加载失败 → 逐附件明确提示」分支在 file-service 故障时不可达，且对一个存在的通知显示「通知不存在」具误导性。 | 在 REQ-NDP 服务职责边界/风险处注明：附件重生失败时详情读整单失败，前端按详情加载失败态处理（提示明确、不静默）；REQ-NDP-2/3 的「file_url 空值」场景限定为「响应已返回但 file_url 为空（legacy 无重生可能）」的逐附件降级，其余整单失败归入 REQ-NDP-1 详情加载失败态。 |
| 2 | `specs/notice-time-window/spec.md` REQ-NTW-5 | 场景 2「published_at 为 NULL 但 created_at 有值 → 卡片显示 created_at」在窗口页（首页/列表）**不可达**：REQ-NTW-1/4 已排除 published_at 为 NULL 的行出窗口，窗口页渲染的卡片全部 published_at 有值；created_at 回退仅在详情页/深链可达。当前措辞给实现者一个「窗口内可能出现 NULL published_at 卡片」的错误预期。 | 将 REQ-NTW-5 时间回退收敛到「详情页渲染（深链场景）」，窗口页（首页/列表）场景删除或标注为防御性回退、不可触发。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | 「便民联络做实」落地后**空态上线**：migration 004 不预置种子（D4），运行库 community_contacts 为空 → 新联络列表页首屏「暂无联络信息」。属用户拍板行为，但建议在 design/上线说明标注「运营需预置联络数据」，避免「做实入口点开是空页」被误判为缺陷。 |
| 2 | proposal §影响范围「详情页当前直接以存储 file_url 打开附件」措辞略不准：现状 `notice-detail.vue` onDownload 已消费详情响应 `att.file_url`（后端已重生），本次实际变更 = 图片/文档分发 + 显式错误处理 + file_type/file_id 类型扩展。方向正确，仅建议措辞修正避免实现者误解（勿将「重生 file_url 直链」误删）。 |

## 问题跟踪表

| # | 状态 | 备注 |
|---|------|------|
| MF-1（本轮） | 待修复 | REQ-NDP-2 file_type「image/*」→ 扩展名集合，对齐 REQ-NDP-4 |
| 上轮 MF-1/2/3 | 已修复 | 本轮代码级验证通过（附件预览 / 窗口索引 / 跑马灯） |
| 上轮 SF-1/2 | 已修复 | 本轮验证通过 |
| SF-1（本轮） | 待修复 | REQ-NDP 错误模式对齐（整单失败 vs 逐附件降级） |
| SF-2（本轮） | 待修复 | REQ-NTW-5 created_at 回退收敛到详情页 |

---
VERDICT: REVISION
---
