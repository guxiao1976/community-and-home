# Content Post Attachment Security Capability Specification

## Purpose

定义通用图文发布附件的安全契约（file-service，复用原 notice 已确认设计 REUSE:notice-D4/D5/D11/D23/D24）：白名单（png/jpg/jpeg/gif/pdf/doc/docx）与禁止清单（exe/bat/sh/cmd/com/msi/apk/js/vbs/ps1/py/pl/php 及 zip/rar 扩展名），单文件 ≤10MB（硬上限），两层校验（GetUploadUrl 快速拒绝 + ConfirmUpload magic-bytes 内容嗅探回读），doc 按 OLE2/CFB + `WordDocument` 流、docx 按 ZIP + `word/document.xml` 部件识别放行，其他 OLE2/OOXML 子类型与通用 zip/rar 内容拒绝（REVISION——正文措辞与场景消歧「zip/rar 扩展名层拒绝 vs docx 内容层特判」）；错误码 070004（类型不支持）/070005（大小超限）登记（复用 notice D11，不重编号 70001-70003）；附件引用校验载体 = 扩展 FileInfo（file_type + confirmed，D14）+ 复用 GetFileUrl 链路；**单帖总量上限（≤10 个/≤50MB）单源在 REQ-CPB-6，本 capability 引用不重复声明（REVISION）**。涉及 file-service（实现）、api-proto（错误码 + FileInfo 扩展）、community-hub-service（attachment_count/file_type 记录 + 绑定校验引用 REQ-CPB-6）。

## Requirements

### Requirement: REQ-CAS-1 — 附件类型白名单与禁止清单（070004，复用 notice D11，措辞消歧 REVISION）

The file-service upload flow SHALL accept only files whose extension is in the whitelist {png, jpg, jpeg, gif, pdf, doc, docx}, and SHALL reject files with extensions in the forbidden set {exe, bat, sh, cmd, com, msi, apk, js, vbs, ps1, py, pl, php}. **Archive handling SHALL be stated precisely (REVISION — resolves the docx=ZIP tension): at the extension layer the system SHALL reject the `.zip` and `.rar` extensions outright; at the content layer (ConfirmUpload) the system SHALL accept docx (OOXML: a ZIP container carrying the `word/document.xml` part) and doc (OLE2/CFB with a `WordDocument` stream) as the only special-cased containers, and SHALL reject all other zip/rar-format content (generic archives, OOXML subtypes xlsx/pptx, OLE2 subtypes msi/xls/ppt, etc.)** — i.e. "reject archives" applies to the extension layer for `.zip`/`.rar` and to any zip/rar content that is not a valid docx, never to the docx container itself. Extension matching SHALL be case-insensitive. Rejection SHALL use error code 070004 不支持的文件类型 (int32 = 70004, constant `ErrCodeUnsupportedFileType`). The codes 070004/070005 are NEW integer slots (70004/70005) and do not collide with the existing 70001-70003 (文件不存在 / 文件访问被拒绝 / 文件操作失败); `ErrCodeFileOperationFailed` (70003) SHALL NOT be renumbered (REUSE:notice-D11). The file.proto header error-code comment block SHALL be aligned to the actual constants.

#### Scenario: 白名单类型放行
- **GIVEN** an authenticated user requesting an upload URL for "notice.pdf" (file_type pdf, size ≤10MB)
- **WHEN** the user calls GetUploadUrl
- **THEN** the system returns a presigned upload URL

#### Scenario: 可执行/脚本类型拒绝（070004）
- **GIVEN** a user requesting an upload URL for "setup.exe" or "run.sh" or "payload.js"
- **WHEN** the user calls GetUploadUrl
- **THEN** the system rejects with error 070004 不支持的文件类型 and no upload URL is issued

#### Scenario: .zip/.rar 扩展名全部拒绝（070004）
- **GIVEN** a user requesting an upload URL for "backup.zip" or "archive.rar"
- **WHEN** the user calls GetUploadUrl
- **THEN** the system rejects with 070004 (zip/rar 扩展名层一律不接收)；仅在 ConfirmUpload 内容层按 docx OOXML 内容签名特判放行，扩展名层不接收

#### Scenario: 无扩展名/点文件被拒（070004）
- **GIVEN** a user requesting an upload URL for a file with no extension or a dot-file (e.g. "README")
- **WHEN** the user calls GetUploadUrl
- **THEN** the system rejects with 070004 不支持的文件类型

### Requirement: REQ-CAS-2 — 单文件 ≤10MB 硬上限（070005）

The file-service upload flow SHALL reject any single file whose size exceeds 10MB (10MB itself SHALL be accepted), using error code 070005 文件大小超限 (int32 = 70005, constant `ErrCodeFileSizeExceeded`). The 10MB hard cap SHALL NOT be relaxed by any entity_type override (REUSE:notice-D4/REQ-AS-4 invariant).

#### Scenario: 单文件 ≤10MB 放行
- **GIVEN** a user requesting an upload URL for a 10MB file (or under)
- **WHEN** the user calls GetUploadUrl
- **THEN** the system returns a presigned upload URL (10MB 边界放行)

#### Scenario: 单文件 >10MB 拒绝（070005）
- **GIVEN** a user requesting an upload URL for a 10.5MB file
- **WHEN** the user calls GetUploadUrl
- **THEN** the system rejects with error 070005 文件大小超限 and no upload URL is issued

### Requirement: REQ-CAS-3 — 两层校验（GetUploadUrl 快速拒绝 + ConfirmUpload magic-bytes 回读）

The upload flow SHALL apply two validation layers (REUSE:notice-D5): L1 GetUploadUrl rejects by declared extension + declared size (fast rejection); L2 ConfirmUpload reads back the actual MinIO object and sniffs its content by magic bytes (not extension / not client metadata): png `89 50 4E 47`, jpg `FF D8 FF`, gif `47 49 46 38`, pdf `%PDF`, doc = OLE2/CFB (D0 CF 11 E0) AND containing a `WordDocument` stream, docx = ZIP (PK) AND containing the `word/document.xml` part. Other OLE2 subtypes (msi/xls/ppt) and OOXML subtypes (xlsx/pptx) and generic zip/rar content SHALL be rejected with 070004 (REVISION — docx is the sole zip-content carve-out). The sniffed type SHALL be mapped to a canonical whitelist extension and SHALL match the declared extension before the file is accepted. On success, the File record SHALL persist `file_type` (sniffed canonical extension) and `confirmed=true` (REUSE:notice-D24).

#### Scenario: 改名 png 上传 exe 被 ConfirmUpload 拦截（070004）
- **GIVEN** a user uploads a file declared as "image.png" whose content is actually a Windows executable
- **WHEN** ConfirmUpload reads back the object and sniffs the magic bytes
- **THEN** the sniffed type (executable) does not map to the whitelist and the upload is rejected with 070004 (magic-bytes 回读拦截，非扩展名/元数据)

#### Scenario: doc 按 OLE2/CFB + WordDocument 流放行
- **GIVEN** a user uploads a genuine ".doc" file whose content is OLE2/CFB and contains a `WordDocument` stream
- **WHEN** ConfirmUpload reads back and sniffs the container
- **THEN** the sniffed type maps to `doc` and matches the declared extension; the upload is accepted and file_type="doc" is recorded

#### Scenario: msi/xls/ppt 改 .doc 被拒（070004）
- **GIVEN** a user uploads a file declared as "payload.doc" whose content is an OLE2 container that is actually an MSI/XLS/PPT (OLE2 subtype without a WordDocument stream)
- **WHEN** ConfirmUpload sniffs the container
- **THEN** the sniffed type is not `doc` (no WordDocument stream) and the upload is rejected with 070004 (封堵同容器子类型绕过)

#### Scenario: docx 按 ZIP + word/document.xml 放行
- **GIVEN** a user uploads a genuine ".docx" file whose content is a ZIP container containing `word/document.xml`
- **WHEN** ConfirmUpload sniffs the container
- **THEN** the sniffed type maps to `docx` and matches the declared extension; the upload is accepted and file_type="docx" is recorded

#### Scenario: 通用 zip/rar 内容改 .docx 被拒（070004）
- **GIVEN** a user uploads a file declared as "payload.docx" whose content is a generic ZIP archive without the `word/document.xml` part (or a rar archive)
- **WHEN** ConfirmUpload sniffs the container
- **THEN** the content is not a valid docx (no word/document.xml) and the upload is rejected with 070004 (docx is the sole zip-content carve-out)

### Requirement: REQ-CAS-4 — 白名单全局基线 + 按 entity_type 可扩展（硬上限不变量）

The whitelist and size limits SHALL form a global baseline with per-entity_type overrides permitted only to be STRICTER (smaller whitelist / smaller size), never weaker: the 10MB hard cap and the forbidden set SHALL NOT be relaxed for any entity_type (REUSE:notice-D4/REQ-AS-4 invariant). content_posts attachments SHALL be governed by the global baseline; a future entity_type for content posts MAY tighten but not loosen these limits.

#### Scenario: content_posts 附件按全局基线校验（正向基线生效）
- **GIVEN** a content post attachment upload with no entity_type override (governed by the global baseline whitelist and 10MB cap)
- **WHEN** the user requests an upload URL and confirms the file
- **THEN** the upload is validated against the global baseline (whitelist {png,jpg,jpeg,gif,pdf,doc,docx} + 10MB cap); a whitelist in-scope file within the cap is accepted and non-whitelist files are rejected with 070004

#### Scenario: 基线外自定义实体不可放宽
- **GIVEN** an entity_type override that attempts to permit "exe" or raise the size cap above 10MB
- **WHEN** the system evaluates the override
- **THEN** the override is rejected (10MB 硬上限与禁止集不可弱化，不变量)

### Requirement: REQ-CAS-5 — file_id/file_type 记录（引用载体 = 扩展 FileInfo，总量上限引用 REQ-CPB-6 REVISION）

The system SHALL persist per-attachment metadata from the file-service FileInfo contract (REUSE:notice-D24): on CreateContentPost / draft attachment-set change (REQ-CPB-9), community-hub-service SHALL read `file_type` back from `FileInfo.file_type` and record `file_id` (the attachment_ids carrier) into `content_post_attachments` (D14), and SHALL set `content_posts.attachment_count` to the number of bound attachments (D15). **The single-post aggregate caps (≤10 attachments and ≤50MB total, rejected with 080005) are owned by REQ-CPB-6 — this Requirement references them and does not re-declare the binding validation** (REVISION, single-source). The FileInfo contract SHALL carry the new fields `file_type` (field 11) and `confirmed` (field 12) added non-breakingly (REUSE:notice REQ-AS-7). The attachment count is the review-completeness comparison basis (REQ-CPB-8).

#### Scenario: 附件元数据完整落库（file_type 自 FileInfo 回读）
- **GIVEN** a user confirming a file via the upload flow whose FileInfo returns file_type="pdf" and confirmed=true
- **WHEN** the user binds the file to a content post
- **THEN** the content_post_attachments row records file_type="pdf" (from FileInfo, not the client), file_id=the bound id, and the post's attachment_count increments

#### Scenario: 单帖附件总量上限由 REQ-CPB-6 强制（绑定场景引用）
- **GIVEN** a user submitting CreateContentPost (or a draft attachment-set change) with attachment_ids whose count exceeds 10, or whose bound files' total size exceeds 50MB (verified from FileInfo.file_size)
- **WHEN** the binding validation (REQ-CPB-6) runs
- **THEN** the request is rejected with 080005 参数无效（附件超限）and no post/attachment change is created (REUSE:notice-D23, enforced at bind time, single-source REQ-CPB-6)

#### Scenario: 旧 FileInfo 契约（无 file_type/confirmed）的兼容处理
- **GIVEN** a pre-existing File record whose extended fields are absent (no file_type/confirmed snapshot, legacy row not back-filled)
- **WHEN** a bind attempts to read file_type/confirmed from it
- **THEN** the read path handles the absent fields gracefully (treated as not-confirmed / no file_type rather than crashing; consistent with the D2 legacy-data-migration compatibility)

## 服务职责边界

- **file-service**: 白名单/禁止集/单文件 10MB 两层校验（L1 快速拒绝 + L2 magic-bytes 回读）；doc/docx 容器签名识别（docx 为唯一 zip 内容特判）；FileInfo 扩展 file_type/confirmed；错误码 070004/070005；单帖总量上限的尺寸数据源（FileInfo.file_size）
- **api-proto**: file/v1 FileInfo 新增 file_type(11)/confirmed(12)（非破坏）；file/v1 头注释错误码块对齐 + 070004/070005 登记（REUSE:notice-D11）
- **community-hub-service**: CreateContentPost/draft 附件变更的绑定校验（GetFileUrl 读扩展 FileInfo：confirmed + user_id 归属 + file_type 回读，单帖 ≤10 个/≤50MB 总量上限 080005 — 单源 REQ-CPB-6）；content_post_attachments 记录 file_id/file_type；attachment_count 落库/重算（D15/REQ-CPB-9）
