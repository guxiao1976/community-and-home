# Notice Attachment Security Capability Specification

## Purpose

定义通知附件上传的安全契约（file-service）：白名单（png/jpg/jpeg/gif/pdf/doc/docx）与禁止清单（exe/bat/sh/cmd/com/msi/apk/js/vbs/ps1/py/pl/php 及所有 zip/rar），单文件 ≤10MB（硬上限）**且单通知附件总量 ≤10 个 且 总大小 ≤50MB（D23）**，采用两层校验（Q5 已拍板：GetUploadUrl 快速拒绝 + ConfirmUpload 回读实际对象校验），第二层以 **magic-bytes 内容嗅探**判定真实类型（非扩展名）。**容器类型识别已修正**（REVISION-3/REVISION-5 闭环）：doc 按 OLE2 Compound File Binary（魔数 D0 CF 11 E0）**且内含 `WordDocument` 流**识别，docx 按 ZIP+OOXML 容器**且内含 `word/document.xml` 部件**识别，两者分别按各自内容签名放行；其他 OLE2/OOXML 子类型（msi/xls/ppt/xlsx/pptx）与通用 zip/rar 一律 070004 拒绝。**错误码冲突已处置**（REVISION-4 闭环）：新增登记 070004（类型不支持，int32=70004）/070005（大小超限，int32=70005），既有 70001-70003 **不重编号**（ErrCodeFileOperationFailed 保持 70003），file.proto 头注释错误码块对齐实际常量——杜绝同整数双语义。**附件引用校验载体 = 扩展 FileInfo（新增 file_type + confirmed 字段，D24）+ 复用 GetFileUrl 链路**：community-hub 经 GetFileUrl(file_id) 校验存在/confirmed/归属，file_type 从该契约回读（非客户端回传）。白名单为全局基线并按 entity_type 可扩展（Q4 已拍板），10MB 为不可放宽硬上限、禁止集不可弱化、总量上限（D23）不可放宽。涉及 file-service（实现）、api-proto（错误码登记 + FileInfo.file_type/confirmed + NoticeAttachment.file_type）、community-hub-service（notice_attachments.file_type 记录 + CreateNotice 总量上限校验）。

## Requirements

### Requirement: REQ-AS-1 — 附件类型白名单与禁止清单（070004 新增登记）

The file-service upload flow SHALL accept only files whose extension is in the whitelist {png, jpg, jpeg, gif, pdf, doc, docx}, and SHALL reject files with extensions in the forbidden set {exe, bat, sh, cmd, com, msi, apk, js, vbs, ps1, py, pl, php} and SHALL reject all archive files (zip, rar) without exception. Extension matching SHALL be case-insensitive. Rejection SHALL use error code 070004 不支持的文件类型 (int32 = 70004, constant `ErrCodeUnsupportedFileType`). **REVISION-4 冲突处置**：070004/070005 are NEW integer slots (70004/70005) and do not collide with the existing 70001-70003 (文件不存在 / 文件访问被拒绝 / 文件操作失败); the previously-planned label "070003" is NOT used for type-unsupported because int32 70003 is already `ErrCodeFileOperationFailed` (used by getfileurllogic/deletefilelogic) — `ErrCodeFileOperationFailed` SHALL NOT be renumbered. The file.proto header error-code comment block SHALL be aligned to the actual constants: 070001 文件不存在, 070002 文件访问被拒绝, 070003 文件操作失败, 070004 文件类型不支持, 070005 文件大小超限 (the current header's "070002 上传失败 / 070003 文件类型不支持" drift SHALL be corrected). The new codes SHALL be registered in api-proto file/v1 header comment + CHANGELOG + file-service errcode constant.

#### Scenario: 白名单类型放行
- **GIVEN** an authenticated user requesting an upload URL for "notice.pdf" (file_type pdf, size ≤10MB)
- **WHEN** the user calls GetUploadUrl
- **THEN** the system returns a presigned upload URL

#### Scenario: 可执行/脚本类型拒绝（070004）
- **GIVEN** a user requesting an upload URL for "setup.exe" or "run.sh" or "payload.js"
- **WHEN** the user calls GetUploadUrl
- **THEN** the system rejects with error 070004 不支持的文件类型 and no upload URL is issued

#### Scenario: 压缩包全部拒绝（070004）
- **GIVEN** a user requesting an upload URL for "backup.zip" or "archive.rar"
- **WHEN** the user calls GetUploadUrl
- **THEN** the system rejects with 070004 (zip/rar 一律不接收; zip/rar 是禁止集，仅在 ConfirmUpload 层按 docx OOXML 内容签名特判放行，扩展名层不接收)

#### Scenario: 大小写绕过无效
- **GIVEN** a user requesting an upload URL for "VIRUS.EXE" or "malware.ExE"
- **WHEN** the user calls GetUploadUrl
- **THEN** the system rejects with 070004 (匹配大小写不敏感，不能通过改扩展名大小写绕过)

#### Scenario: 无扩展名 / 点文件拒绝
- **GIVEN** a user requesting an upload URL for a file with no extension (e.g., "README") or a dot-file ("config.ini")
- **WHEN** the user calls GetUploadUrl
- **THEN** the system rejects with 070004 (白名单语义下无扩展名/未知扩展名一律拒绝；不接收）

### Requirement: REQ-AS-2 — 单文件大小上限 10MB（070005 新增登记）

The upload flow SHALL reject any single file whose size exceeds 10MB (10 * 1024 * 1024 bytes). A file exactly at the 10MB boundary SHALL be accepted. Rejection SHALL use error code 070005 文件大小超限 (int32 = 70005, constant `ErrCodeFileSizeExceeded`, NEW slot — no collision with 70001-70003, see REQ-AS-1).

#### Scenario: 超限拒绝（070005）
- **GIVEN** a user requesting an upload URL for a file declared as 12MB
- **WHEN** the user calls GetUploadUrl
- **THEN** the system rejects with error 070005 文件大小超限

#### Scenario: 边界放行
- **GIVEN** a user requesting an upload URL for a file of exactly 10MB
- **WHEN** the user calls GetUploadUrl
- **THEN** the system accepts and issues an upload URL

### Requirement: REQ-AS-3 — 两层校验（GetUploadUrl 快速拒绝 + ConfirmUpload magic-bytes 回读）

The upload flow SHALL validate file type and size in two layers: (1) at GetUploadUrl, reject based on the declared file name extension and declared file size (fast reject before any bytes are uploaded); (2) at ConfirmUpload, re-read the actual stored object from object storage and detect its REAL content type by magic-bytes (content sniffing) — not by the object key extension and not by client-supplied Content-Type metadata (both are client-controllable). The real type recognized by magic bytes SHALL be mapped to the whitelist extension set, and the file SHALL be accepted only when the declared extension and the sniffed type agree (after mapping sniffed type → canonical whitelist extension). A declared-whitelisted filename carrying a different actual payload SHALL be rejected with 070004 (type is the primary security boundary; if size also exceeds, 070005 SHALL be the secondary code). **Container-type recognition SHALL distinguish the two Word containers and reject other same-container subtypes** (REVISION-3 + REVISION-4): `.doc` (Word 97-2003 binary) is an **OLE2 Compound File Binary** — recognized by the CFB magic `D0 CF 11 E0 A1 B1 1A E1` **AND the presence of a `WordDocument` stream inside the OLE2 container** (the Word-document-specific storage stream; merely matching the CFB header is NOT sufficient, because `.msi` (Windows Installer), `.xls`, `.ppt` are also OLE2 containers with the same CFB header — a `.msi` renamed to `.doc` SHALL be rejected with 070004, NOT accepted as doc); `.docx` is a **ZIP container with an OOXML Word profile** — recognized by the ZIP signature (`PK` header) **AND the presence of a `word/document.xml` part inside the ZIP archive** (merely containing `[Content_Types].xml` is NOT sufficient, because `.xlsx`/`.pptx` also contain it — a `.xlsx` renamed to `.docx` SHALL be rejected with 070004). Other OLE2 subtypes (msi/xls/ppt) and other OOXML subtypes (xlsx/pptx) SHALL be rejected with 070004 (no whitelist mapping). A generic zip/rar archive SHALL be rejected with 070004 (no whitelist mapping). pdf SHALL be recognized by its `%PDF` signature; images by their standard magic bytes (e.g., PNG `89 50 4E 47`, JPEG `FF D8 FF`, GIF `47 49 46 38`).

#### Scenario: 声明与真实不一致被拒（改名 png 上传 exe）
- **GIVEN** a client that obtained an upload URL for "photo.png" (declared image/png, 1MB)
- **WHEN** the client actually uploads an executable object (magic bytes = PE `MZ`, 25MB) and then calls ConfirmUpload
- **THEN** the system reads back the stored object, sniffs the magic bytes, detects the real type (executable) does not match the declared png, rejects the confirmation with 070004 (type mismatch is the primary code; the size exceed of 070005 is secondary) and the file is not registered for the notice

#### Scenario: 真实 .doc（OLE2/CFB + WordDocument 流）放行（REVISION-3 更新）
- **GIVEN** a client uploading a real Word 97-2003 `.doc` file whose magic bytes are the OLE2 CFB signature `D0 CF 11 E0 ...` **and whose OLE2 container contains a `WordDocument` stream**
- **WHEN** the client calls ConfirmUpload
- **THEN** the system sniffs the OLE2/CFB signature AND verifies the `WordDocument` stream is present (Word-specific), maps it to the whitelist extension doc (NOT treated as generic ZIP), and accepts the file

#### Scenario: .msi 改名 .doc 上传 → 拒绝（REVISION-4 新增，防 OLE2 子类型绕过）
- **GIVEN** a client that obtained an upload URL for "document.doc" but actually uploads a Windows Installer `.msi` file (an OLE2 container with the same CFB magic `D0 CF 11 E0 ...` but NO `WordDocument` stream)
- **WHEN** the client calls ConfirmUpload
- **THEN** the system detects the OLE2 CFB header but finds no `WordDocument` stream (`.msi` is an OLE2 subtype, not Word), rejects with 070004 (`.msi` is in the forbidden list; merely matching the CFB header is not sufficient)

#### Scenario: 真实 .docx（ZIP + word/document.xml）放行（REVISION-3 更新）
- **GIVEN** a client uploading a real .docx file (a ZIP container whose OOXML profile contains the `word/document.xml` part)
- **WHEN** the client calls ConfirmUpload
- **THEN** the system sniffs the ZIP container, verifies the `word/document.xml` part is present (Word-specific OOXML marker), maps it to the whitelist extension docx, and accepts the file

#### Scenario: .xlsx 改名 .docx 上传 → 拒绝（REVISION-4 新增，防 OOXML 子类型绕过）
- **GIVEN** a client that obtained an upload URL for "sheet.docx" but actually uploads an Excel `.xlsx` file (a ZIP container containing `[Content_Types].xml` and `xl/workbook.xml`, but NO `word/document.xml`)
- **WHEN** the client calls ConfirmUpload
- **THEN** the system detects the ZIP container + OOXML marker but finds no `word/document.xml` part (`.xlsx` is an OOXML subtype, not Word), rejects with 070004 (merely containing `[Content_Types].xml` is not sufficient)

#### Scenario: 通用 zip 容器被拒（无法映射白名单）
- **GIVEN** a client uploading a generic .zip archive whose content signature does not match docx OOXML profile
- **WHEN** the client calls ConfirmUpload
- **THEN** the system rejects with 070004 (generic archive not in whitelist; only OOXML-signed ZIP containers map to docx, and only OLE2/CFB-signed containers map to doc)

#### Scenario: 正常确认成功
- **GIVEN** a client that uploaded a real 1MB PNG whose magic bytes match the declared type and size
- **WHEN** the client calls ConfirmUpload with the object_key
- **THEN** the system validates the stored object, records the file metadata, and returns the file info (含 id)

### Requirement: REQ-AS-4 — 全局基线 + 按 entity_type 可扩展（10MB 硬上限）

The whitelist and the 10MB size limit SHALL be a global baseline applied to all upload flows by default. The system SHALL allow per-`entity_type` extension/refinement (e.g., a stricter subset for avatar uploads, or a broader document set for a future entity type) without weakening the global baseline. The forbidden set (exe/sh/zip/rar, etc.) SHALL remain rejected in all overrides. The 10MB limit SHALL be a hard upper bound that an override cannot widen. Existing entity types (e.g., avatar, verification, lostfound, contacts/notice) SHALL continue to function; the baseline SHALL NOT regress uploads that previously succeeded for whitelisted types under 10MB.

#### Scenario: 基线默认生效
- **GIVEN** an upload request without any entity_type-specific override
- **WHEN** GetUploadUrl is called
- **THEN** the global baseline (whitelist + 10MB) is applied

#### Scenario: entity_type 覆盖（不弱化基线、不放松 10MB/禁止集）
- **GIVEN** an upload request with entity_type whose override adds an extra allowed type
- **WHEN** GetUploadUrl is called
- **THEN** the override is applied on top of the global baseline; the forbidden set (exe/sh/zip/rar etc.) SHALL remain rejected in all overrides, and the 10MB hard cap SHALL NOT be widened

#### Scenario: 既有 entity_type 上传不回归
- **GIVEN** existing upload flows for entity types such as avatar, verification, lostfound, and contacts using whitelisted types (png/jpg/pdf) under 10MB
- **WHEN** the baseline validation ships
- **THEN** those uploads continue to succeed; only previously-unsupported types/sizes are newly rejected (regression boundary verified for each existing entity_type)

#### Scenario: override 试图放宽 10MB / 放行禁止集被拒（不变式不弱化）
- **GIVEN** an entity_type override that tries to raise the single-file limit above 10MB, or to add a forbidden type (exe/sh/zip/rar) to the allowed set
- **WHEN** GetUploadUrl is called with that override
- **THEN** the baseline invariants hold: the widened size limit is rejected (070005) and the forbidden extension is rejected (070004); an override SHALL NOT weaken the 10MB hard cap nor the forbidden set (REQ-AS-4 invariants enforced)

### Requirement: REQ-AS-5 — notice_attachments.file_type 记录（载体 = 扩展 FileInfo，D24）

The community-hub-service SHALL persist a `file_type` field on `notice_attachments` (and expose it in the `NoticeAttachment` proto message) as the recorded basis of the whitelist validation for each attachment. The recorded file_type SHALL be read back from the extended `FileInfo.file_type` returned by the file-service `GetFileUrl` (or `ListFiles`) contract — i.e., the type that passed the ConfirmUpload magic-bytes validation — and SHALL NOT be taken from client-supplied request fields. If the FileInfo lacks a file_type value (e.g., legacy data), the attachment MAY be recorded with an empty file_type, and the missing value SHALL NOT break the read path.

#### Scenario: 附件携带 file_type（自 FileInfo 回读）
- **GIVEN** a notice with one attachment confirmed as PDF, whose extended FileInfo.file_type is "pdf" (returned via GetFileUrl)
- **WHEN** the notice is created and its detail is later fetched
- **THEN** the attachment's file_type is "pdf" (read back from the validated contract, not from a client-supplied value) and is present in the NoticeAttachment response

#### Scenario: 附件元数据缺失
- **GIVEN** a notice whose attachment lacks a file_type value (e.g., legacy data)
- **WHEN** the notice detail is fetched
- **THEN** the response still includes the attachment with file_type empty; the missing value does not break the read path

### Requirement: REQ-AS-6 — 单通知附件总量上限（≤10 个 且 总大小 ≤50MB，D23）

The system SHALL cap the total attachments of a single notice: a notice SHALL reference at most 10 attachments, and the sum of their file sizes SHALL NOT exceed 50MB (50 * 1024 * 1024 bytes). These aggregate caps SHALL be enforced at CreateNotice binding time (community-hub-service) using the file sizes read back from the extended `FileInfo.file_size` (via GetFileUrl), not from client-declared sizes. A CreateNotice whose `attachment_ids` count exceeds 10, or whose bound files' total size exceeds 50MB, SHALL be rejected with 080005 参数无效（附件超限）and no notice SHALL be created. The 10MB single-file cap (REQ-AS-2) is independent and continues to apply at upload time; the aggregate caps are in addition to it.

#### Scenario: 数量超限被拒（080005）
- **GIVEN** a user who confirmed 12 files and submits CreateNotice with attachment_ids listing all 12
- **WHEN** the system validates the attachment references
- **THEN** the request is rejected with 080005 参数无效（附件数量超限，>10）and no notice is created

#### Scenario: 总大小超限被拒（080005）
- **GIVEN** a user who confirmed 3 files of 20MB each (each under the 10MB single-file cap, total 60MB)
- **WHEN** the user submits CreateNotice with those 3 attachment_ids
- **THEN** the sum (60MB) exceeds the 50MB aggregate cap and the request is rejected with 080005 参数无效（附件总大小超限）and no notice is created

#### Scenario: 恰好在总量边界内放行
- **GIVEN** a user who confirmed 10 files whose total size is exactly 50MB
- **WHEN** the user submits CreateNotice with those attachment_ids
- **THEN** the request passes the aggregate-cap validation (count ≤10 AND total ≤50MB, boundaries inclusive) and the notice is created

### Requirement: REQ-AS-7 — FileInfo 扩展 file_type + confirmed 字段（D24）

The file-service `FileInfo` proto message SHALL be extended with two new fields: `file_type` (the whitelist-canonical type validated at ConfirmUpload magic-bytes layer) and `confirmed` (a boolean indicating the upload flow completed successfully — the file object exists and its metadata was recorded). These fields SHALL be populated by file-service and returned through the existing `GetFileUrl` (and `ListFiles`) contract so that consuming services can verify attachment references without a new RPC. Field evolution SHALL be non-breaking (new field numbers, e.g., file_type=11, confirmed=12).

#### Scenario: GetFileUrl 返回扩展字段
- **GIVEN** a confirmed uploaded PDF whose ConfirmUpload validated the magic bytes as pdf
- **WHEN** community-hub-service invokes GetFileUrl(file_id) during CreateNotice attachment validation
- **THEN** the returned FileInfo carries file_type="pdf" and confirmed=true

#### Scenario: 未确认文件返回 confirmed=false
- **GIVEN** a file whose GetUploadUrl was issued but ConfirmUpload was never called (no object recorded as complete)
- **WHEN** a consuming service invokes GetFileUrl(file_id)
- **THEN** the returned FileInfo has confirmed=false; the consumer SHALL reject a binding that references an unconfirmed file (080005, see REQ-NP-6)

## 服务职责边界

- **file-service**: 白名单/禁止清单/10MB 校验、两层校验（GetUploadUrl 快速拒绝 + ConfirmUpload magic-bytes 回读）、doc OLE2/CFB+WordDocument 流 + docx OOXML word/document.xml 容器识别、通用校验器沉淀；**扩展 FileInfo（file_type + confirmed，D24）**；**新增登记错误码 70004/70005（070004/070005）**；70001-70003 不重编号
- **api-proto**: file/v1 头注释错误码块对齐实际常量并登记 070004/070005 + FileInfo 扩展 file_type/confirmed + CHANGELOG；community/v1 NoticeAttachment 增加 file_type
- **community-hub-service**: notice_attachments.file_type 写入与返回（载体 = GetFileUrl 扩展 FileInfo）；CreateNotice 单通知总量上限校验（≤10 个 且 ≤50MB，080005）
- **web/mobile**: 前端可做一致的预校验（白名单 + 单文件大小 + 单通知总量）以提升体验，但最终安全边界在后端（见 notice-mobile REQ-NM-6）
