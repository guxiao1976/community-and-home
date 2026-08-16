# Design Review — notice-multicommunity-publish（data-model + interface-proto 视角，第 3 轮复审）

**审查维度**: 数据模型（字段/关系/Snowflake/时间/软删除）+ 接口契约（gRPC/Proto 自洽、破坏性标注、鉴权/错误码语义）
**审查对象**: `.harness/changes/notice-multicommunity-publish/design.md` + `tasks.md`（07:09 修订版）
**对照**: 6 个 spec + `.change.yaml` + 磁盘代码证据（community.proto/file.proto/permission.proto/masterdata.proto、migration 001/002、init_permissions.sql、model/notice.go、model/notice_attachment.go、types.go、GetFileUrlResponse、BACKLOG.md、GetResidentialAreasByDivision）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 3
- **VERDICT: REVISION**（本轮新发现 1 个 MUST FIX：S4 附件 file_url 重生方案缺少 file_id 持久化载体）

---

## 一、上一轮（v2）问题修复核验（逐条，代码证据确认）

### ✅ data-model v2 MUST（0 项）+ SHOULD S1-S4 落地核验

| 项 | v2 问题 | 落地证据 | 状态 |
|---|---------|---------|:---:|
| S1 | `isModerationPassed` 未导出 | design §可见性门禁 + Task 1.4 统一 `model.IsModerationPassed(status int64) bool` 导出；Task 1.5/1.13 复用 | ✅ |
| S2 | FindOneByCommunity 冗余 | Task 1.4 显式「不新增 FindOneByCommunity」，GetNotice 收敛 `FindOnePublished(id)` + notice_scope 匹配（Task 1.8 对齐） | ✅ |
| S3 | role 派生需 GetUserRoles | Task 1.7 + design §CreateNotice 显式补「调 permission GetUserRoles(user_id) 解析实际角色」+ 多角色选择顺序（grid_worker>community_admin>committee） | ✅ |
| S4 | file_url 短期预签名 | 已采纳「file_id 权威 + 详情读时重生」方向（design §GetNotice、Task 1.8、Task 0.4 CHANGELOG、ADR 行）——**但方向落地缺 file_id 持久化载体，见本轮 MUST 1** | ⚠️ 部分 |

### ✅ interface-proto v2 MUST 1（GetNotice community_id form 标签）+ SHOULD 2（blast radius）+ INFO 3/4

- Task 1.15 `GetNoticeReq.CommunityId` 改 `form:"community_id"` + 禁用 json 标签 + RED 断言 `?community_id=456` 绑定；Task 1.16 marquee 同用 form 标签；design §GetNotice/§GetMarqueeNotices 补 REST 绑定说明。✅
- Task 3.1 验证门禁追加 lostfound/contacts 调用方回归（community_div=D1 时两写路径 allowed）；design §Design Gate「共享 blast radius 声明」+ ADR 行。✅
- Task 4.1 改「新增 createNotice client」（community.ts 现无此函数，非扩既有，已核实）。✅
- Task 2.5 补「粒度说明」理由（迁移即模型列 DB 载体，Task 5.1 兜底）。✅

---

## 二、本轮新发现

### 🔴 MUST FIX（1）

| # | 文件:章节 | 问题 | 修复建议 |
|---|---------|------|---------|
| M1 | design.md §GetNotice「附件 file_url 语义（评审 S4 修订）」+ ADR「附件 file_url 持久化」+ tasks.md Task 1.8 | **S4「按 file_id 重生预签名 URL」无数据载体：`notice_attachments` 没有 `file_id` 列，GetNotice 无法重生 → 附件 7 天后变死链**。design 明确「详情读路径对每个附件按 `file_id` 调 `GetFileUrl(file_id)` 重生预签名 URL」，但磁盘证据确认：`notice_attachments` 表（migration 001: id/notice_id/file_name/file_url/file_size/created_at）无 file_id 列；设计 migration 003 仅 `ADD COLUMN file_type`，未加 file_id；`model/notice_attachment.go` 无 FileId 字段；proto `NoticeAttachment`（1-4 + 新增 5=file_type）无 file_id；CreateNotice 落库只存 file_name/file_url/file_size/file_type，`attachment_ids`(file_id) 校验后即丢弃。预签名 URL 默认 3600s、上限 7 天，超期后 stored file_url 失效；GetNotice 无 file_id 无法重生 → 新通知附件在详情页数周后不可下载，违反 REQ-NM-3（附件可打开/下载）。 | migration 003 为 `notice_attachments` 增 `file_id BIGINT NOT NULL`（存量兼容期可先 DEFAULT 0/NULL，新写强制）；model `NoticeAttachment` 增 `FileId int64`；CreateNotice 落库时把 `attachment_ids`（file_id）一并写入 `notice_attachments.file_id`；GetNotice 按存储的 file_id 重生预签名 URL。可选：proto `NoticeAttachment` 暴露 `file_id`（内部重生用可不暴露，但模型必须有）。 |

### 🟡 SHOULD FIX（1）

| # | 文件:章节 | 问题 | 建议 |
|---|---------|------|------|
| S1 | design.md §GetNotice / migration 003 | M1 加列后，**存量 `notice_attachments` 行 file_id 为 NULL**：Q1 已声明存量 notice 因无 notice_scope 行暂不可见（GetNotice→080001），但若后续回填迁移补 notice_scope 使旧通知可见，其附件 file_url 已过期且无 file_id，仍不可下载。 | 设计显式声明：file_id 列可空、GetNotice 遇 file_id NULL 时回退返回 stored file_url（尽力而为）或与 Q1 一并声明旧附件不可下载的兼容期；回填迁移（挂 BACKLOG）需同时回填 notice_attachments.file_id 或声明不可恢复。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I1 | `notice_scope` 复合 PK 满足了 REQ-NP-2 的「唯一约束」语义，但约束名不是 spec 字面的 `uk_notice_community`。若存在按约束名引用的工具/脚本需确认；建议迁移注释注明「复合 PK 即 uk_notice_community 唯一约束」避免执行方误以为缺唯一约束。 |
| I2 | GetNotice 重生附件 URL 为逐附件 N 次 GetFileUrl RPC（≤10），design 非功能已声明「附件校验逐条 GetFileUrl」，重生路径同理，量级可接受；建议在可观测性补充重生 RPC 失败降级日志。 |
| I3 | `restore-compensation-zero-time` 记忆引用正确：published_at 用 `sql.NullTime`（Task 1.3），`time.Time{}` 零值写 DATETIME 已防；确认无遗漏。 |

---

## 三、核验清单（磁盘证据）

- [x] Proto 字段号无冲突（community.proto CreateNoticeRequest 1-7 占用、8/9 空闲；GetNoticeRequest id=1、community_id=2 空闲；NoticeAttachment 1-4、file_type=5；FileInfo 1-10、file_type=11/confirmed=12）✅
- [x] 全部 int64 带 `[jstype=JS_STRING]`；community_id(1) deprecated 保留 wire 兼容 ✅
- [x] moderation_status 语义 1/2/3/4 = community.proto UpdateModerationStatusRequest 注释一致 ✅
- [x] 迁移编号正确（community-hub 现有 001/002 → 003；file 现有 001 → 002）；001 的 `community_id BIGINT NOT NULL`/`published_at DATETIME NOT NULL`/notice_attachments 无 file_type 与去 NOT NULL 前提一致 ✅
- [x] GetFileUrlResponse 含 download_url(2)+FileInfo(3)，D24「经 GetFileUrl 读扩展 FileInfo」契约成立 ✅
- [x] 错误码常量（file 70001/70002/70003 现状；community 080002/080003/080005/080006/080007；头注释漂移声明准确）✅
- [x] 种子证据：421 现绑 (2,3,6)+(1,5)、grid_worker(4) 未持 → 授 4 + 回收 2/1/5 正确；422 现绑 (9,1,5)、grid_worker 不在内 → 425 专用码排除发布者判定成立；421 min_verf_level 现状 0 → 置 2 前提成立 ✅
- [x] AssertPublishScope 共享 RPC（lostfound createlostfoundlogic / contacts upsertcontactslogic 调用）→ blast radius 声明 + 回归门禁已补 ✅
- [x] masterdata GetResidentialAreasByDivision 存在（community_div_id>0 + status=1）；division<=0 guard 必要性成立（default FindAll 分支）✅
- [x] GetUserRoles 输出含 status/verified_at/expires_at → level-2 判定可基于 RPC 输出 ✅
- [x] 任务粒度：单任务不跨服务、不混合「模型+逻辑+前端」三类、Proto/Migration 独立成任务、测试用例 1~10 个 ✅
- [x] **`notice_attachments` 全库无 file_id（grep community-hub-service 无 FileId/file_id 命中）→ M1 前提确认** ✅

## 报告自检
1. MUST FIX 定位到具体章节（design §GetNotice/ADR + Task 1.8 + migration 003）✅
2. MUST FIX 附可落地修复建议（加列 + 模型 + 落库 + 重生）✅
3. data-model + interface-proto 两维度均覆盖（file_id 持久化/迁移/索引/时间字段/软删除/谓词/角色派生/附件 URL/Proto 字段号/鉴权/错误码/批量校验/必填回归）✅

---
VERDICT: REVISION（M1 为 MUST FIX：S4 附件重生缺 file_id 载体，需架构设计师修订后复审）
---
