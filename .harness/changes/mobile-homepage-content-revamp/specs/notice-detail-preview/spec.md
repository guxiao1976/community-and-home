# Notice Detail & Attachment Preview Capability Specification

## Purpose

通知详情页完整展示通知的标题、发布单位（role）、发布时间、内容与附件列表，并让附件点击预览统一走 file-service 能力：图片附件经 `uni.previewImage` 全屏预览，文档附件经下载后 `uni.openDocument` 打开。附件预览主路径直接消费详情响应中由 community-hub 服务端重生的 file-service 预签名 URL（`file_url`），前端不直连 file-service REST。目标是消除当前详情页「直接以存储 file_url 打开附件」的脆弱链路（存储 file_url 对新帖为占位空串），让附件预览与 file-service 生命周期一致。涉及 web/mobile（详情页 + 附件类型扩展）与 community-hub-service（GetContentPost 服务端重生 file_url，复用既有能力，无新增后端变更）与 file-service（GetFileUrl 签名 URL，复用，无契约变更）。

> **REVISION 修订记录**：首轮修订解决评审 MUST #4——附件预览主路径由「前端以 file_id 直连 file-service REST 取签名 URL、file_url 仅回退」改为「直接消费详情响应中已由 community-hub 服务端重生（GetContentPost 内 RPC GetFileUrl）的 file_url」：file-service REST `GET /api/files/:id` 强制文件所有权（`rpcResp.File.UserId != userId → permission denied`），通知附件由发布方上传、查看者非所有者 → 原主路径必然被拒；且新帖附件落库 file_url 为占位空串，回退路径亦无可用 URL → 两路皆败。实际 wire 中 GetContentPost 已在服务端经 RPC GetFileUrl（无所有权限制）把 file_url 重生为有效预签名 URL，REST 透出——详情响应 file_url 本身就是权威新鲜签名 URL。另解决 SF-1：REQ-NDP-4 显式声明前端附件类型须扩展 file_id/file_type。
>
> **第 2 轮 REVISION（本轮，r2-1/r2-3/r2-4 = 评审 coverage/clarity/validity 同轮 MUST）**：统一图片/文档分发谓词并修正与真实 wire 的偏差——wire 的 `file_type` 是 file-service magic-bytes 嗅探落库的**规范小写扩展名**（迁移 003 注释「文件类型（扩展名，自 FileInfo 回读）」；`guard/magic.go` SniffType 返回 "png"/"jpg"/"gif"/"pdf"/"doc"/"docx"），**不是 MIME**，wire 上不存在 `image/*` 值。REQ-NDP-2 的 `file_type ∈ image/*`（MIME 前缀读法）与 REQ-NDP-4 的 `file_type: "jpg"` 自相矛盾；按字面 `startsWith('image/')` 实现会致所有图片附件被当文档处理、图片全屏预览对全部附件失效。**已修复**：REQ-NDP-2 图片分发谓词固定为扩展名白名单 `file_type ∈ {png, jpg, jpeg, gif}`（与 file-service 白名单对齐），其余（pdf/doc/docx 或缺失/无法识别）视为文档走 REQ-NDP-3，REQ-NDP-2/3/4 三处判定口径统一。另解决同轮 validity 发现（r2-6）：附件 `file_url` 服务端重生失败（file-service 不可用/文件已删）时 GetContentPost 读**整单失败**（toProtoAttachments 对任一附件 GetFileUrl RPC 失败返回 (nil, err)）→ REST 透错 → 详情页按 REQ-NDP-1 详情加载失败态处理；REQ-NDP-2/3 的「file_url 为空」降级场景仅限「响应已返回但 file_url 为空（legacy 无重生可能）」的逐附件情况，其余整单失败归入 REQ-NDP-1。

## Requirements

### Requirement: REQ-NDP-1 — 详情页完整展示标题 / 发布单位 / 发布时间 / 内容 / 附件

The notice detail page SHALL display all of the following for a notice: title, publish role (发布单位), publish time (`published_at`), content, and the attachment list (when attachments exist). When a notice has no attachments, the page SHALL NOT render the attachments section (no empty attachment header shown).

#### Scenario: 完整展示含附件
- **GIVEN** a review-complete notice with title, role, published_at, content, and 2 attachments
- **WHEN** a user opens the notice detail page
- **THEN** the page displays title, role (发布单位), publish time, content, and the 2 attachments with file name and size

#### Scenario: 无附件时不渲染附件区
- **GIVEN** a notice with `attachment_count == 0`
- **WHEN** a user opens the notice detail page
- **THEN** the page displays title/role/time/content and does not render the attachments section

#### Scenario: 详情加载失败明确提示
- **GIVEN** the notice detail API fails or the notice does not exist / is out of scope
- **WHEN** a user opens the notice detail page
- **THEN** the page shows the failure/not-exists state explicitly (加载失败 / 通知不存在), not a silently blank page

### Requirement: REQ-NDP-2 — 图片附件点击全屏预览（消费详情响应重生 file_url）

When a user taps an image attachment — defined as attachment `file_type` ∈ {png, jpg, jpeg, gif}, the lowercase image-extension whitelist that file-service persists via magic-byte sniffing (REVISION r2-1/r2-3/r2-4: the wire `file_type` is the sniffed canonical extension, NOT a MIME type, so `image/*` values never occur on the wire) — the system SHALL preview the image full-screen via `uni.previewImage` using the attachment `file_url` carried in the detail response — the file-service presigned URL already regenerated server-side by community-hub (GetContentPost → RPC GetFileUrl). The preview SHALL NOT invoke document-opening APIs. When the detail response has returned successfully and carries the attachment but its `file_url` is empty (e.g. legacy row with no regeneration possible) or the image fails to load during preview, the system SHALL surface an explicit preview-failure message (禁止静默吞错) and SHALL NOT fall through to a document opener. (REVISION #4: main path consumes the response `file_url`; the frontend does not call file-service REST `GET /api/files/:id`, which enforces file ownership and would deny a viewer who is not the attachment's uploader. REVISION r2-6: a file-service-level regeneration failure makes the whole detail read fail and is handled as the REQ-NDP-1 detail load-failure state, not this per-attachment branch.)

#### Scenario: 图片附件全屏预览
- **GIVEN** a notice detail whose response carries an image attachment (`file_type` ∈ {png, jpg, jpeg, gif}, e.g. `file_type: "png"`) with a non-empty regenerated `file_url`
- **WHEN** the user taps the image attachment
- **THEN** the system opens a full-screen `uni.previewImage` preview using that `file_url` (image dispatch by lowercase extension membership, never by `image/*` prefix)

#### Scenario: 图片附件 file_url 不可用时明确提示（逐附件降级，响应已返回）
- **GIVEN** a detail response that returned successfully and carries an image attachment, but the attachment's `file_url` is empty (legacy row where server regeneration produced no URL) or the image fails to load during the preview
- **WHEN** the user taps the image attachment
- **THEN** the system shows an explicit preview-failure message (禁止静默吞错) and does not fall through to a document opener (REVISION r2-6: a whole-order regeneration failure — file-service unavailable / file deleted — surfaces as the REQ-NDP-1 detail load-failure state, not this per-attachment branch)

### Requirement: REQ-NDP-3 — 文档附件经详情响应重生 file_url 打开，前端不直连 file-service

When a user taps a non-image attachment — a document extension `file_type` ∈ {pdf, doc, docx}, or a missing/unrecognized `file_type` NOT in the image whitelist {png, jpg, jpeg, gif} — the system SHALL use the attachment `file_url` carried in the detail response (the file-service presigned URL regenerated server-side by community-hub GetContentPost) as the download source: download it (e.g. via `uni.downloadFile`) and open the document via `uni.openDocument`. The frontend SHALL NOT call file-service REST `GET /api/files/:id` to re-resolve a URL (REVISION #4: that REST endpoint enforces file ownership and rejects a viewer who is not the uploader; the detail response `file_url` is already the authoritative fresh presigned URL). When the detail response has returned successfully and carries the attachment but its `file_url` is empty (legacy row with no regeneration) or the download/open fails, the system SHALL surface an explicit error (如「附件打开失败」, 禁止静默吞错). (REVISION r2-6: a file-service-level regeneration failure makes the whole detail read fail and is handled as the REQ-NDP-1 detail load-failure state, not this per-attachment branch.)

#### Scenario: 文档附件以详情响应签名 URL 打开
- **GIVEN** a notice detail whose response carries a document attachment with a non-empty regenerated `file_url` (presigned URL from file-service)
- **WHEN** the user taps the document attachment
- **THEN** the system downloads the file from that `file_url` and opens the document via `uni.openDocument`; no direct file-service REST call is made by the frontend

#### Scenario: 文档附件 file_url 不可用时明确报错（逐附件降级，响应已返回）
- **GIVEN** a document attachment whose response `file_url` is empty (legacy row where regeneration could not produce a URL) or whose download fails
- **WHEN** the user taps the document attachment
- **THEN** the system shows an explicit error message (如「附件打开失败」) rather than silently doing nothing

#### Scenario: 附件点击目标类型无法识别时安全处理
- **GIVEN** a non-image attachment whose `file_type` is missing or unrecognized (NOT in the image whitelist {png, jpg, jpeg, gif} and not a known document extension pdf/doc/docx)
- **WHEN** the user taps the attachment
- **THEN** the system treats it as a document (download from response `file_url` + openDocument) and never attempts a raw in-page navigation that bypasses file-service

### Requirement: REQ-NDP-4 — 前端附件类型须扩展 file_id / file_type 字段

The mobile frontend `NoticeAttachment` type SHALL include `file_id` and `file_type` fields (snake_case aligned with the wire: `file_id`/`file_type`), in addition to the existing `id`/`file_name`/`file_url`/`file_size`. This makes the detail/list wire fields consumable by the preview logic (REVISION #4 + validity SF-1: the current frontend type only carries id/file_name/file_url/file_size, so the image-vs-document dispatch on `file_type` and the server regeneration key `file_id` have no field to consume).

#### Scenario: 前端类型解析 wire 的 file_type 用于图片/文档分发
- **GIVEN** a detail response whose attachment includes `file_type: "jpg"` (a lowercase image-whitelist extension, matching file-service's persisted wire value — not a MIME value) and `file_id: "1001"`
- **WHEN** the frontend maps the response to `NoticeAttachment`
- **THEN** `file_type` and `file_id` are present on the mapped object, and the tap handler dispatches to `uni.previewImage` (file_type ∈ {png, jpg, jpeg, gif}) vs document-open (non-image, REQ-NDP-3) accordingly

#### Scenario: wire 缺失 file_type / file_id 时类型字段保持缺省且不崩溃（边界）
- **GIVEN** a detail response whose attachment carries only id/file_name/file_url/file_size (no `file_type`/`file_id`, e.g. legacy wire)
- **WHEN** the frontend maps the response to `NoticeAttachment`
- **THEN** the mapped object has `file_type`/`file_id` absent (undefined/empty), the tap handler treats the attachment as a document (REQ-NDP-3 fallback), and no runtime crash occurs (type fields are optional-safe)

## 服务职责边界

- **web/mobile**: 详情页渲染标题/role/发布时间/内容/附件列表；附件点击分发按 `file_type` 扩展名白名单判定——`file_type ∈ {png, jpg, jpeg, gif}` → `uni.previewImage` 全屏预览，其余（pdf/doc/docx 或缺失/无法识别）→ 下载详情响应 `file_url` 后 `openDocument`（REQ-NDP-2/3/4 同一谓词，r2-1/r2-3/r2-4）；前端不直连 file-service REST（所有权限制）；`NoticeAttachment` 类型扩展 `file_id`/`file_type`（REQ-NDP-4）；加载/预览失败明确提示。附件 `file_url` 服务端重生失败（file-service 不可用/文件已删）→ 详情读整单失败 → 按 REQ-NDP-1 详情加载失败态处理（r2-6），逐附件降级仅限「响应已返回但 file_url 为空」的 legacy 情况
- **community-hub-service**: GetContentPost 读路径已把附件 `file_url` 重生为 file-service 预签名 URL（`file_id` 权威重生载体、兼容期 `file_id=0` 回退 stored URL），REST 透出；对任一附件 GetFileUrl RPC 失败时读整单失败（r2-6）——本能力无新增后端变更（复用既有 wire）
- **file-service**: GetFileUrl（RPC，无所有权限制）为 community-hub 服务端重生提供预签名 URL；REST `GET /api/files/:id` 强制所有权，前端不调用（无契约变更，复用）
