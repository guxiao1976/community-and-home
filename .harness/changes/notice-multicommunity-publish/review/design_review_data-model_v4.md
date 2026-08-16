# Design Review — notice-multicommunity-publish（data-model + interface-proto 视角，第 4 轮复审）

**审查模式**: 模式一.5 设计评审 · **视角**: data-model（数据模型/服务归属）+ interface-proto（接口契约/Proto 破坏性/鉴权/错误码）
**审查对象**: `.harness/changes/notice-multicommunity-publish/design.md` + `tasks.md`（07:19 修订版）
**对照基准**: v3 评审问题清单 + 6 个 spec（notice-publish/notice-read/attachment-security…）+ `.change.yaml` + 磁盘真相（community-hub migration 001/002、model/notice.go、model/notice_attachment.go、helper.go、createnoticelogic.go、getnoticelogic.go、listnoticeslogic.go、community.proto、file.proto、file-service migration 001、getfileurllogic.go、common/pkg/minio）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 3
- **VERDICT: APPROVED**（v3 MUST FIX——附件 file_id 载体——已闭环；本轮无阻塞问题）

---

## 一、上一轮（v3）MUST FIX 修复核验（磁盘证据确认）

| 项 | v3 问题 | 落地证据 | 状态 |
|---|---------|---------|:---:|
| M1 | S4 附件重生缺 file_id 持久化载体（`notice_attachments` 无 file_id 列，GetNotice 无法重生预签名 URL → 旧通知附件死链） | migration 003（design §数据模型 + Task 1.1）新增 `notice_attachments.file_id BIGINT DEFAULT 0`；Task 1.3 `NoticeAttachment.FileId int64` 模型字段 + Insert 含 file_id；Task 1.7 CreateNotice 落库把 `attachment_ids`(file_id) 写入 file_id；Task 1.8 GetNotice 按 file_id 调 GetFileUrl 重生 + 兼容期 file_id=0/NULL 回退 stored file_url；Task 0.4 CHANGELOG 登记 file_url 短期预签名语义；ADR「附件 file_url 持久化」已更迭。磁盘确认 `notice_attachments`（migration 001）原无 file_id/file_type，前提准确。 | ✅ 已修复 |

---

## 二、本轮新发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|---------|------|------|
| S1 | tasks.md Task 1.3 vs helper.go `toProtoNotice` | **模型类型变更的共享转换器未显式列任务**：Task 1.3 将 `Notice.CommunityId` 改 `*int64`、`PublishedAt` 改 `sql.NullTime`，但 `rpc/internal/logic/notice/helper.go` 的 `toProtoNotice`（getnoticelogic/listnoticeslogic 共用）现写 `CommunityId: n.CommunityId`、`PublishedAt: n.PublishedAt.Unix()`——前者变指针不匹配 int64，后者 `sql.NullTime` 无 `.Unix()` 方法，Task 1.3 后**编译即断**；Task 1.7 仅声明「roleToString 保留」，未点名 toProtoNotice。且新语义「响应 community_id = 请求小区（不读弃用列）」，converter 的 CommunityId 来源也必须改为 scope 派生，非仅解引用。 | Task 1.3 或 Task 1.8/1.9 显式补一项：更新 `toProtoNotice`（`CommunityId` 由请求/scope 注入、`PublishedAt` 用 `sql.NullTime` null 感知取值），并注明 ListNotices/GetNotice 的 community_id 响应覆盖逻辑，避免执行方漏改导致 build 反复回退或响应带 NULL。 |
| S2 | tasks.md Task 1.5/1.12 + design §DeleteNotice | **撤回「单事务」的实现机制未指定**：`SoftDelete`（notices）+ `NoticeScopeModel.DeleteByNoticeId` 若各自用 `m.conn.ExecCtx`（非共享 session），中间失败会留半态——scope 行已删而 notices 未软删（通知不可见却未真正撤回），或软删成功而 scope 行残留（孤儿关联，REQ-NP-5「together / one operation」不成立）。 | design/Task 1.12 钉死：逻辑层用 `conn.Transact(func(session) error)` 传共享 session 给两个 model 方法（或 model 层提供 `WithdrawTx`），并补一条「中途失败 → 全部回滚、无孤儿 scope 行」的失败注入测试。 |
| S3 | tasks.md Task 1.4/1.9/1.10 | **JOIN 双 community_id 列投影陷阱**：`FindListByCommunity`/`FindMarquee` 走 `notices JOIN notice_scope`，两表均含 `community_id`（notices 弃用列新行为 NULL、scope 列为目标小区）。若 SQL 用 `select *` 或未限定列，sqlx 按列名扫描到 `Notice.CommunityId` 时取到哪一列取决于列序（可能取到弃用 NULL），破坏「响应 community_id = 请求小区」（REQ-NR-1 场景）契约。 | Task 1.4/1.9 显式规定 JOIN 投影：`notices.*` + `notice_scope.community_id`（别名或右表限定），使模型 CommunityId = scope 派生值；Task 1.9 RED 加断言「多小区通知在 C2 列表返回 community_id=C2」。 |
| S4 | design §CreateNotice 落库 + migration 003（notice_attachments.file_url） | **`notice_attachments.file_url VARCHAR(500)` 首次落预签名 URL，有截断风险**：现代码无 notice_attachments 写入路径（磁盘确认 createnoticelogic 不插附件），本变更首次把 `GetFileUrl` 返回的 MinIO 预签名 URL（含 `X-Amz-Algorithm/Credential/Date/Expires/Signature` 等参数，object key 含 19 位 Snowflake + 19 位纳秒 + ≤255 文件名）存进 500 字符列；文件名较长时 URL 可超 500 → MySQL 严格模式 `ERROR 1406 Data too long` → 附件发布整事务回滚。 | 二选一：将 `file_url` 扩为 `TEXT`/`VARCHAR(1024)`（兼容存量）；或（更优，file_id 已为权威重生载体）新行只存 file_id、不再落 URL 快照，file_url 仅保留给 file_id=0 存量回退。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I1 | 多角色用户仅存单一 role（选择顺序 grid_worker>community_admin>committee）——`notices.role` 单值列既有局限，ListNotices 按另一角色过滤不会命中该通知；属既有语义延续，建议 design 一句话标注。 |
| I2 | 存量附件（file_id=0）回退 stored file_url 可能已过期——design 已注「尽力而为」；建议 BACKLOG 回填迁移同步回填 notice_attachments.file_id（或声明旧附件不可恢复），与 Q1 一并挂账。 |
| I3 | file-service `002_file_guard.sql` 需与 001 一致 `USE file_db;`（磁盘确认 001 用 file_db），Task 2.5 落库时注意 DB 上下文。 |

---

## 三、核验清单（磁盘证据）

- [x] 迁移前提与实际 DDL 一致：notices.community_id `BIGINT NOT NULL`、published_at `DATETIME NOT NULL`、notice_attachments 无 file_type/file_id（migration 001）；moderation_status 0-4 语义（migration 002）与 design 谓词一致 ✅
- [x] **v3 M1 闭环**：`notice_attachments.file_id` 加列 + model FileId + CreateNotice 落库 + GetNotice 重生 + 存量回退，design/Task/ADR 四层齐备 ✅
- [x] notice_scope 满足 REQ-NP-2：复合 PK(notice_id,community_id) = uk 唯一约束 + `idx_scope_community(community_id 左)`，双列 NOT NULL；物理删除语义符合编码规范 §3.1（纯关联表仅 created_at）✅
- [x] role 存储链一致：grid_worker(RBAC) → NOTICE_ROLE_GRID_OFFICER(proto) → `"grid_officer"`(DB，roleToString)，ListNotices 过滤走同一 string ✅
- [x] 模型 NULL 处理模式已在库内（`ModerationTime sql.NullTime`、`PublisherId *int64` 既有），`*int64`/`sql.NullTime` 落地可行 ✅
- [x] Proto 全部兼容新增且字段号空闲：CreateNoticeRequest 8/9、GetNoticeRequest 2、NoticeAttachment 5、FileInfo 11/12（file.proto 现状 1-10）；全 int64 带 JS_STRING ✅
- [x] GetNotice/GetMarqueeNotices community_id 必填的语义破坏已登记 Task 0.4 CHANGELOG + 消费面（web/pc 无消费、web/mobile 本变更内升级）核实收敛 ✅
- [x] 附件校验载体成立：GetFileUrl 返回 download_url + FileInfo（file.proto GetFileUrlResponse），CreateNotice 校验与 GetNotice 重生一次 RPC 双满足 ✅
- [x] 任务粒度：单任务不跨服务、不混合「模型+逻辑+前端」三类（Task 2.5 迁移+模型合并为既审 README 已接受的粒度说明，Task 5.1 独立运维验证兜底）、测试 1~10 个 ✅

## 报告自检
1. MUST FIX：0（v3 MUST 已闭环，本轮无阻塞）✅
2. 4 条 SHOULD 均定位到具体 Task/章节并附可落地修复建议 ✅
3. data-model + interface-proto 两维度覆盖：迁移/索引/时间字段/软删除/file_id 载体/谓词/角色派生/附件 URL 宽度/JOIN 投影/Proto 字段号/破坏性/错误码/鉴权/幂等 ✅

---
VERDICT: APPROVED
---
