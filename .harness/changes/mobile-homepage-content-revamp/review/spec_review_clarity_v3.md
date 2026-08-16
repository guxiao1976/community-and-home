# Plan Review — mobile-homepage-content-revamp（清晰可执行视角）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL 唯一解释、Scenario 落地性、术语一致）
**审查版本**: P1.3 · fallback:r2:rc1（与 v1/v2 轮次哈希不同，按磁盘最新内容独立重审；r2 轮修订 r2-1~r2-6 已入 spec）

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 4

## 上轮（v2）问题修复验证
| v2 项 | 结论 |
|------|------|
| MUST #1（REQ-NDP-2 `file_type ∈ image/*` 与 wire/NDP-4 自相矛盾，字面实现即图片预览不可用） | ✅ 已修复（r2-1/r2-3/r2-4）：REQ-NDP-2 分发谓词统一为扩展名白名单 `{png, jpg, jpeg, gif}`，REQ-NDP-3/4 同口径；REQ-NDP-2/3/4 三处判定一致，wire 事实（嗅探扩展名非 MIME）已显式声明。 |
| SHOULD #1（REQ-NTW-1 场景 THEN「20-day-old 非置顶帖」自相矛盾） | 🟡 部分修复：场景 GIVEN 已补「3 older in-window notices」，THEN 主句（置顶 20 天帖在前 → 最新非置顶 → 截断 3 条）清晰唯一；但尾句仍残留「the 1-hour-old notice is shown before the **20-day-old non-pinned** ones」——GIVEN 中 20 天前帖是置顶帖、不存在「20-day-old 非置顶」，表述依旧矛盾（语义可还原为「较旧非置顶」）。 |
| SHOULD #2（REQ-NTW-5 published_at→created_at 回退在首页/列表为死规则；REQ-NDP-1 详情页 NULL 时间未定义） | 🟡 未修复：REQ-NTW-5 仍声明回退且作用域为「首页通知区 + 列表页」（该两处已排除 NULL published_at，回退永不触发）；详情页按 id 取、不过窗口，**可**展示 NULL published_at 的通知，而 REQ-NDP-1「publish time (published_at)」对 NULL 未定义 → 实现者对详情页 NULL 时间无唯一解释。 |
| SHOULD #3（REQ-NTW-2「non-numeric」对 int32 不成立） | ✅ 已修复（r2-5）：措辞改为「≤0 or >365 拒绝，080005」，并显式说明非数字由 REST 网关解析层拒绝。 |
| INFO #2（since_days 字段名/取值未钉死） | ✅ 已修复：REQ-NTW-2 钉死 `since_days`（int32，1..365，additive 字段 6）。 |

## 交叉校验（Spec ↔ 磁盘真实契约，逐条核过）
- `file_type` 实际落库值 = **magic-bytes 嗅探规范扩展名**（`guard/magic.go:30-55` SniffType 返回 `png|jpg|gif|pdf|doc|docx`；`confirmuploadlogic.go:68` `FileType: sniffedExt`）。**注意：JPEG 一律规范化为 `"jpg"`，`"jpeg"` 永不出现在 wire**——REQ-NDP-2 白名单含 `jpeg` 为无害超集，但「the whitelist that file-service persists」的表述不精确（见 SHOULD #3）。
- `ListContentPostsRequest` 现字段 1-5（`api-proto/api/community/v1/community.proto:115-120`，community_id/role/section_code/page/page_size），`since_days` additive 字段 6 可行；page_size 上限 100，REQ-NTW-4 场景 page_size=50 合法。
- REST 透传链路（r2-2）：`api/internal/types.ListContentPostsReq` form 字段 → `api/internal/logic/notice/listcontentpostslogic.go` → RPC `rpc/internal/logic/notice/listcontentpostslogic.go`——两文件补入 .change.yaml revises，REQ-NTW-2 服务职责边界已完整描述，无断环。
- `NoticeAttachment` 现仅 `id/file_name/file_url/file_size`（community.ts:10-15），缺 `file_id/file_type`——REQ-NDP-4 成立。
- `notice-browse.vue:110` 现请求 `getNoticeList(cid, 1, 50)`（page_size=50，与 REQ-NTW-4 场景一致）；`:112` 存在客户端 3 个月过滤（`90*24*3600`）——REQ-NTW-4「移除客户端过滤」成立。
- 首页 `getNoticeList` 默认 `page_size=3`（community.ts:122）——REQ-NTW-1 `page_size=3` 成立。
- 空态文案一致：暂无通知公告 / 暂无联络信息 / 功能开发中 / 互助功能开发中。

## 发现

### 🔴 MUST FIX
| # | 文件:章节 | 问题 | 修复建议 |
|---|------|------|------|
| （无） | | | |

### 🟡 SHOULD FIX
| # | 文件:章节 | 问题 | 建议 |
|---|------|------|------|
| 1 | specs/notice-time-window/spec.md → REQ-NTW-1 场景「窗口内置顶帖与更早非置顶帖并存时置顶优先」(line 35) | 尾句「the 1-hour-old notice is shown before the **20-day-old non-pinned** ones」残留矛盾：GIVEN 中 20 天前帖为**置顶**帖，无「20-day-old 非置顶帖」；与 v2 SHOULD #1 同族，未彻底修净。 | 尾句改写为「the 1-hour-old non-pinned notice is shown before the other, older in-window non-pinned notices」，与 SHALL 排序（置顶优先 → published_at 倒序 → 截断 3）逐字对齐。 |
| 2 | specs/notice-time-window/spec.md → REQ-NTW-5 (line 121,128-131) + specs/notice-detail-preview/spec.md → REQ-NDP-1 (line 15) | REQ-NTW-5 的 published_at→created_at 回退作用域限「首页+列表」（该两处已排除 NULL published_at，回退为死规则）；而**详情页**按 id 取、不过窗口，可展示 NULL published_at 通知，REQ-NDP-1 对「publish time」NULL 时显示未定义 → 详情页 NULL 时间无唯一解释（v2 SHOULD #2 未修复）。 | 在 REQ-NDP-1 补一条「published_at 为 NULL 时详情页时间显示回退 created_at（与 REQ-NTW-5 卡片口径一致）」；或在 REQ-NTW-5 显式声明该回退同时覆盖详情页展示（卡片组件 + 详情页共用同一时间格式化）。 |
| 3 | specs/notice-detail-preview/spec.md → REQ-NDP-2 (line 34, 修订注 line 9) | REQ-NDP-2 声称 `{png, jpg, jpeg, gif}` 是「the lowercase image-extension whitelist that file-service persists via magic-byte sniffing」——但 file-service 嗅探落库值恒为 `{png, jpg, gif, pdf, doc, docx}`，JPEG 一律规范化为 `jpg`，`jpeg` **永不出现在 wire**。白名单含 jpeg 功能无害（超集），但「file-service persists 的白名单」表述不精确，且与 REQ-NDP-4 场景 `file_type: "jpg"` 并存会让实现者疑惑 jpeg 是否真会出现。 | 二选一：(a) 白名单收敛为 `{png, jpg, gif}` 并注明「jpeg 内容在 wire 规范化为 jpg」；(b) 保留 `jpeg` 但把表述改为「`{png, jpg, jpeg, gif}`（jpeg 兼容历史/文件名扩展，wire 实际为 jpg）」——使「spec 声称的 wire 白名单」与真实落库值逐字一致。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| 1 | specs/notice-time-window/spec.md → REQ-NTW-6：content_posts 索引迁移编号仍「设计阶段定」（proposal/.change.yaml 亦一致）；REQ-CLP-2 已钉死 004 为 community_contacts，设计阶段须确保索引迁移编号 ≠ 004（建议 005），避免同变更两迁移撞号。 |
| 2 | specs/notice-time-window/spec.md → REQ-NTW-4 (line 85)：SHALL 写「固定 page_size」未钉具体值，值仅出现在场景（50）与 D11（「如 50」）；现状 notice-browse 即请求 50，建议 SHALL 直接钉死 page_size=50 消歧。 |
| 3 | specs/homepage-layout/spec.md → REQ-HL-1 (line 11-12)：占位区块「点击不导航」已定义，但未定义点击是否弹 toast；REQ-FE-3 对 3 个入口占位定义了「功能开发中」toast。建议为邻里互助占位明确「点击弹『互助功能开发中』toast + 不跳转」（v2 INFO #3 未修复）。 |
| 4 | specs/notice-detail-preview/spec.md → REQ-NDP-2/3：图片/文档分支失败文案未统一（NDP-2 无文案示例、NDP-3 有「附件打开失败」）；建议共用同一失败文案常量（v2 INFO #4 未修复）。 |

---
VERDICT: APPROVED（清晰可执行视角：无 MUST FIX；r2-1~r2-6 已解决上轮 MUST #1 与 SHOULD #3，残余 3 项 SHOULD 均为边缘表述/边界缺口，不阻塞进入设计阶段，建议在设计中同步修正）
---
