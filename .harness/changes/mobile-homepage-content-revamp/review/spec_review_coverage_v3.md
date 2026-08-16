# Plan Review — mobile-homepage-content-revamp（覆盖完整性视角）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别
**审查版本**: P1.3 fallback:r2:rc1（与上轮 r1:rc1 哈希不同——按磁盘最新内容独立重新审查，未沿用旧轮结论或缓存）
**审查对象**: proposal.md + specs/{notice-time-window, notice-detail-preview, function-entries, contact-list-page, homepage-layout}/spec.md + .change.yaml（request.md 不存在，以 .change.yaml 决策包 D1-D16 + proposal 验收标准 + out_of_scope 为原始需求对照）
**事实核验**: 已对照磁盘代码逐项核验 spec 依赖的事实锚点（见下），r2 修订（r2-1~r2-6）已全部在磁盘内容中确认落盘

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 4
- 上轮（r1）SHOULD FIX ×2 / INFO ×2 → r2 已修复 SHOULD ×2（图片谓词统一、REST 透传清单），INFO 部分遗留（非阻塞）

## 事实锚点核验（本轮重新核验，含 r2 新增锚点）

| 锚点 | 核验结果 |
|------|---------|
| `since_days` 需贯通 REST 层：types.go ListContentPostsReq（现无 since_days 字段）+ api listcontentpostslogic.go（RPC 请求字面量未透传）→ RPC → model FindListByCommunity（现无窗口谓词） | ✅ types.go:43-49 无 since_days；api logic:29-35 无透传；rpc logic:68 + model/content_post.go:178-197 无窗口谓词（r2-2 前提成立） |
| 附件 wire：ContentPostAttachment 含 file_id/file_type/file_url/file_size | ✅ api-proto community.proto:87-95 |
| 前端 NoticeAttachment 现仅 id/file_name/file_url/file_size，缺 file_id/file_type | ✅ community.ts:10-15（REQ-NDP-4 前提成立） |
| getNoticeList 现无 since_days 参数（communityId,page,pageSize） | ✅ community.ts:119-128（需加 since_days） |
| notice-browse.vue 现 page_size=50 + 客户端 3 个月过滤 | ✅ notice-browse.vue:110-115（REQ-NTW-4 前提成立） |
| 首页跑马灯由 notices 派生 + marquee-bar 含「更多 →」入口（→notice-browse，是列表页唯一首页入口） | ✅ notice.vue:49-55, 280-283（见 INFO 2） |
| 首页现内嵌联络拨号网格 + 广告位 ×2（联络下方）+ ×1（寻失下方） | ✅ notice.vue:92-150, 198-213（REQ-HL-3 前提成立） |
| 详情页现 `onDownload(att.file_url)` H5 window.open 直链 | ✅ notice-detail.vue:84-98（REQ-NDP-2/3 改造目标） |
| GetContentPost toProtoAttachments 对任一附件 GetFileUrl 失败返回 (nil, err) → 详情读整单失败 | ✅ getcontentpostlogic.go:99-102（r2-6 前提成立） |
| file-service 白名单：图片 {png,jpg,jpeg,gif} / 文档 {pdf,doc,docx}（嗅探扩展名，非 MIME） | ✅ file-service/internal/guard/whitelist.go:34-40（REQ-NDP-2/3 谓词与 wire 对齐） |
| migration 001 已定义 community_contacts（列与 CommunityContactModel 一致） | ✅ 001_initial.sql:39-49（REQ-CLP-2 前提成立） |
| proto ListContentPostsRequest 现有字段 1-5，字段 6 可做 additive since_days | ✅ community.proto:115-121（D10/r2-2 契约前提成立） |
| 非法窗口参数非数值在 REST 层：httpx.Parse 失败 → responsex.Response → errx.FromError nil → 回落 CodeInternalError（**非 080005**） | ✅ notice/handler.go:26-32 + common/pkg/responsex/response.go:31-42（见 INFO 1——r2-5 括号内声称「同参数无效错误」与实测不符） |

## 覆盖矩阵核验（决策点 D1-D16 → REQ）

| 决策点 | 覆盖 REQ | 结论 |
|--------|---------|------|
| D1 跑马灯 + 3 条卡片同源 30 天（REVISION #6） | REQ-NTW-1/3 | ✅ |
| D2 列表页同样 30 天过滤 | REQ-NTW-4 | ✅ |
| D3 4 功能入口（便民做实 + 3 占位 + 原网格移除） | REQ-FE-1/2/3 + REQ-CLP-1 | ✅ |
| D4 不预置联络种子 | REQ-CLP-2 | ✅ |
| D5 附件预览统一 file-service（REVISION #4） | REQ-NDP-2/3/4 | ✅ |
| D6/D7 广告集中 + 点击预留 | REQ-HL-3 | ✅ |
| D8/D9 邻里互助占位无后端无页面 | REQ-HL-1 | ✅ |
| D10 时间窗口后端强制 | REQ-NTW-2 | ✅ |
| D11 列表页数据量行为（REVISION #2） | REQ-NTW-4 | ✅ |
| D12 published_at NULL/未来边界（REVISION #3） | REQ-NTW-1/4 | ✅ |
| D13 索引/性能（REVISION #5） | REQ-NTW-6 | ✅ |
| D14 通知卡片视觉契约（REVISION #1） | REQ-NTW-5 | ✅ |
| D15 附件分发谓词 = file_type 扩展名白名单（r2-1/r2-3/r2-4） | REQ-NDP-2/3/4 | ✅ 三处谓词口径统一 |
| D16 REST 层透传 since_days + 附件重生整单失败边界（r2-2/r2-6） | REQ-NTW-2 + REQ-NDP-1/2/3 | ✅ |

- 验收标准（含 r2 新增「REST 透传验收」）→ 全部映射到 REQ
- out_of_scope 10 条 → 全部显式对齐（GetMarqueeNotices 不变、file-service REST 不直连、邻里互助无后端、广告硬编码、PC/发布侧不动等）
- 每个 REQ 均 ≥1 正向 + ≥1 异常 Scenario（逐 REQ 核对：NTW-1/2/3/4/5/6、NDP-1/2/3/4、FE-1/2/3、CLP-1/2、HL-1/2/3/4 全部达标）

## 上轮问题修复核验（问题跟踪表 r1 → r2）

| # | 上轮问题 | 状态 |
|---|---------|------|
| 1（SHOULD）REQ-NDP-2 图片谓词 `image/*` 与 REQ-NDP-4 `file_type:"jpg"` 不一致、与 wire 不符 | **已修复**——REQ-NDP-2/3/4 统一为扩展名白名单 {png,jpg,jpeg,gif}→图片、其余→文档（D15，r2-1/3/4）；已核验 file-service 白名单逐项吻合 |
| 2（SHOULD）since_days REST 层透传文件未列入 revises | **已修复**——.change.yaml revises 补入 `api/internal/types/types.go` + `api/internal/logic/notice/listcontentpostslogic.go`（r2-2），REQ-NTW-2 服务职责边界显式点明三层透传 |
| 3/4（INFO）NTW-5 时间回退不可达 + 跑马灯「更多 →」未声明 | 部分遗留（非阻塞，见本轮 INFO 2/4） |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | specs/notice-time-window/spec.md · REQ-NTW-1（首页通知区） | 首页通知区的**加载失败边界未定义**：REQ-NTW-4（列表页）与 REQ-NDP-1（详情页）均有显式「加载失败明确提示」场景，唯 REQ-NTW-1 只有空态/NULL/未来/置顶边界，无 API 失败场景。首页为多区块复合页，通知区是核心区块，现状代码 catch 后仅 toast「通知加载失败」（notice.vue:339-342）。覆盖完整性要求每 REQ 异常分支对齐，当前首页通知区缺网络/后端失败态定义。 | 在 REQ-NTW-1 增补「首页通知区加载失败」异常 Scenario：明确失败时展示加载失败提示（不静默、不渲染陈旧数据），并说明与区块空态/整体首页的关系。 |
| 2 | .change.yaml revises 清单 | 新建页 `web/mobile/src/pages/contact-list/contact-list.vue` 未列入 revises：revises 的「新建文件（非修订）」分类只列了 migration 004，未列 contact-list.vue。REQ-CLP-1/REQ-FE-2 与 proposal §影响范围均明确该新页，但若任务拆分仅从 revises 派生，新页创建可能漏任务（pages.json 已列、spec 已声明，风险为任务派生层面）。 | 在 .change.yaml revises 补充 `web/mobile/src/pages/contact-list/contact-list.vue`（新建文件），与 migration 004 同类注明，保证任务清单派生完整。 |

### 🔵 INFO

| # | 文件:行号/章节 | 建议 |
|---|-------------|------|
| 3 | specs/notice-time-window/spec.md · REQ-NTW-2（非法参数 Scenario 括号） | r2-5 括号声称「malformed non-numeric REST query value is rejected at the gateway parse layer with the same parameter-invalid error」——与实测不符：ListContentPostsHandler 对 httpx.Parse 失败走 `responsex.Response(w, nil, err)`（notice/handler.go:29-32），responsex 对非 errx 错误回落 `errx.CodeInternalError`（response.go:38-40），**并非 080005**。边界已覆盖（非数值 → 显式拒绝、不静默），但所述机制/错误码不准确，实现者按字面验证 080005 会得到不同结果。建议把括号措辞改为「parse 层显式拒绝（具体错误码以网关实测为准），服务端 080005 仅覆盖数值范围 ≤0/>365」。 |
| 4 | specs/homepage-layout/spec.md · REQ-HL-4（首页区块全序） | 跑马灯条「更多 →」（notice.vue:54，现为通知列表页唯一首页入口）在 REQ-HL-4（自述为「前端布局唯一依据」）与 REQ-NTW-3 中未声明保留与否：若实现者按 REQ-HL-4 字面「marquee bar + up-to-3 notice cards」移除该入口，REQ-NTW-4 定义的列表页将失去首页可达路径（卡片点击只进详情）。v2 INFO #4 遗留。建议 REQ-HL-4/REQ-NTW-3 显式一句：「跑马灯条保留『更多 →』入口，点击进入通知列表页」。 |
| 5 | specs/notice-time-window/spec.md · REQ-NTW-1（置顶优先 Scenario） | 置顶优先 Scenario 的 THEN 中「the 1-hour-old notice is shown before the 20-day-old non-pinned ones」与 GIVEN 不自洽（GIVEN 无 20 天前非置顶帖，仅「3 older in-window notices」未命名）。语义意图明确（置顶优先 + 倒序），但措辞易使实现者困惑。建议改写为「…then the newest non-pinned in-window notice, truncated to 3 total」直述。 |
| 6 | specs/notice-detail-preview/spec.md · REQ-NDP-1（详情页发布时间） | 详情页发布时间对 NULL published_at 的回退未定义：窗口过滤仅作用于首页/列表，详情页（FindOneReviewComplete）可展示 NULL-published_at 行，现状 formatTime 回退 created_at（notice-detail.vue:15）。REQ-NDP-1 只写「publish time」，未声明 NULL 回退。v2 INFO #3 遗留。建议 REQ-NDP-1 明确「发布时间为 NULL 时回退 created_at（与现状一致）」。 |

## 问题跟踪表
| # | 状态 |
|---|------|
| 1（首页通知区加载失败） | 待评估（SHOULD） |
| 2（contact-list.vue 未入 revises） | 待评估（SHOULD） |
| 3（r2-5 括号错误码） | 待评估（INFO） |
| 4（更多 → 入口） | 待评估（INFO，v2 遗留） |
| 5（置顶 Scenario 措辞） | 待评估（INFO） |
| 6（详情页时间回退） | 待评估（INFO，v2 遗留） |

---
VERDICT: APPROVED
---
