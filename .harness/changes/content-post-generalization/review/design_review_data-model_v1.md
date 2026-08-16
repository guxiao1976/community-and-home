# Design Review — content-post-generalization（data-model / interface-proto 视角）

**审查维度**: 数据模型 + 接口契约/Proto
**审查对象**: `.harness/changes/content-post-generalization/design.md` + `tasks.md`
**审查者**: Reviewer（data-model 视角，含 interface-proto）

## 摘要
- 🔴 MUST FIX: 2 / 🟡 SHOULD FIX: 2 / 🔵 INFO: 4

## 已核实的事实基准（对照代码库）

- 现有 `notices` 表（migration/001_initial.sql:10-25）字段：id / community_id BIGINT **NOT NULL** / title / content TEXT NOT NULL / role VARCHAR(20) NOT NULL / publisher VARCHAR(100) NOT NULL / publisher_id BIGINT DEFAULT NULL / is_pinned TINYINT DEFAULT 0 / published_at DATETIME **NOT NULL** / created_at / updated_at / deleted_at / idx_community / idx_published。design §数据模型 RENAME+去 NOT NULL+保留列的描述**与现状一致**（REQ-CPB-1 门禁成立）。
- `notice_attachments`（001:28-36）：id / notice_id / file_name / file_url VARCHAR(500) NOT NULL / file_size / idx_notice(notice_id)。RENAME 后 idx_notice 随 CHANGE COLUMN 自动改指 post_id —— 设计描述正确。
- `permission-service scope.go` `resolveUserScope(scopeType='community')` 只收集 `g.ScopeType == scopeType`，`community_div` grant 确实不会被收集 → Design Gate 判据断言**准确**（community_admin division 授权误拒成立）。
- moderation `task_handler.go` 已有 `case "notice":`（nil-client 跳过），Task 4.1 改为无条件跳过 + 移除 NoticeServiceClient 可行，与 proto 移除 `UpdateNoticeModerationStatus` 一致（`UpdateModerationStatusRequest/Response` 仍被 LostFoundService 复用）。
- file-service errcode.go 现有 70001/70002/70003，70004/70005 为新整数位，**无冲突**；`getuploadurllogic.go` 当前**无**扩展名/大小校验，Task 2.2 新增 L1 校验是真实增量。
- proto `Notice` 消息字段号：1=id 2=community_id 3=title 4=content 5=role 6=publisher 7=publisher_id 8=is_pinned 9=published_at 10=created_at 11=updated_at 12=attachments。**Task 0.1 给出的新字段号与之冲突（见 MUST FIX 1）。**

## 发现

### 🔴 MUST FIX

#### M1（interface-proto）Task 0.1 `Notice→ContentPost` 消息字段号与既有字段直接冲突，protoc 无法编译，且 role 字段号标错

- **位置**: `tasks.md` Task 0.1（ContentPost 消息字段定义）
- **问题**:
  - `新增 section_code(3)` → 与既有 `title=3` 冲突（重复字段号 3）；
  - `新增 status(10)` → 与既有 `created_at=10` 冲突；
  - `新增 attachment_count(11)` → 与既有 `updated_at=11` 冲突；
  - `role(6) 改 ContentPostRole` → 实际 role 在字段 **5**（字段 6 是 publisher），号码错误。
  - 若照字面实现，同一个消息内 title/section_code 同号 3、created_at/status 同号 10、updated_at/attachment_count 同号 11、publisher/role 同号 6 → protoc 报 "Field number already used"，**编译即失败**；即便勉强编译，字段号复用（不同语义）也是 wire 兼容事故。
- **修复**: 保留既有字段号（`content→text` 用 4、`role→ContentPostRole` 用 5，语义不变仅改名/改枚举类型，wire 兼容），新增字段**追加新号**（如 `section_code=13`、`status=14`、`attachment_count=15`）。若坚持全量重排，则须列出**全部**字段的新号并保证唯一（当前任务未列出 title/publisher/created_at/updated_at 等的新号，无法自洽）。

#### M2（interface-proto）entry_status / UpdateContentPost.status 的 REST 与 proto 枚举数值语义不一致 + 「透传」导致 submit 路径失效

- **位置**: `tasks.md` Task 0.1（proto enum `ContentPostEntryStatus {UNSPECIFIED=0; DRAFT=1; SUBMITTED=2}`）vs Task 1.21（REST `EntryStatus int32：0=draft 默认/1=submitted`、`Status int32：仅接受 1=submitted`）vs Task 1.22（`entry_status`/`status`「透传」）
- **问题**: REST 层 1=submitted，proto 层 1=DRAFT、2=SUBMITTED。照「透传」实现：
  - CreateContentPost：REST submitted(1) → proto DRAFT(1) → 被当作 draft 落库（应立即提交反而变草稿）；
  - UpdateContentPost：REST submitted(1) → proto DRAFT(1) → 只接受 SUBMITTED(2) 的 RPC 判 080005，submit 动作永远失败。
  - 两阶段状态机的「提交」核心动作在两个入口都被破坏，业务不可用。
- **修复**: 二选一并落进 Task：
  1. 统一数值编码——proto 枚举去掉 DRAFT=1 的歧义（`UNSPECIFIED=0` 语义即「draft 默认」+ `SUBMITTED=1`），REST 与 proto 同号；或
  2. 在 Task 1.22 API→RPC 显式映射（REST 0→DRAFT、1→SUBMITTED），禁止裸「透传」，并在 Task 1.21 标注与 proto 数值的对应关系。

### 🟡 SHOULD FIX

#### S1（data-model）content_post_scope 缺 updated_at/deleted_at，design 声明「符合编码规范 §3.1」但未登记偏离；且 withdrawn 帖 scope 行永久保留使表单调增长，增长估算未计残留

- **位置**: `design.md` §数据模型（content_post_scope 建表）；migration 003
- **问题**: 硬性约束 #3.1「时间字段统一 created_at/updated_at/deleted_at，全链路一致」。content_post_scope 仅 `created_at`；设计注释声称合规，但未登记为显式偏离。同时 REQ-CPB-10 要求撤回**保留** scope 行（不软删、不物理删），则表随「历史上所有发布过的帖」单调增长，design 的「行数 ≈ 帖数×目标小区数」估算未计入 withdrawn 残留。
- **建议**: 在 design 显式登记偏离（纯关联表、created_at-only，理由：行不可变、撤回由主表 deleted_at 表达、改写走 delete+reinsert 物理替换）；增长估算改为「累计发布帖数×目标小区数」，并补一句「撤回帖 scope 行永久保留，量级仍小、索引兜底」的显式结论。

#### S2（interface-proto）`GetMarqueeNotices` 新消息命名为 `NoticeMarqueeItem`，与 D4 一次到位通用化命名不一致

- **位置**: `tasks.md` Task 0.1（`NoticeMarqueeItem`）；design.md §GetMarqueeNotices
- **问题**: 本变更将所有 `Notice*` 更名为 `ContentPost*`（D4 一次到位），唯新消息仍带 `Notice` 前缀，契约命名不统一（虽是 wire 名、低风险）。
- **建议**: 命名为 `ContentPostMarqueeItem`（如介意前端 path 未改，则至少保持 RPC 名 `GetMarqueeNotices` 为既有约定并显式说明为何保留 Notice 词）。

### 🔵 INFO

#### I1（任务粒度）Task 1.9 / 1.10 的 RED 场景数超出「单任务测试用例 1~10 个」上限
- Task 1.9 RED 列了 ~14 个场景、Task 1.10 ~10 个。多为 table-driven 可合并的断言，属边界超标；建议在 task 内注明「同一 table-driven 测试表的多个 case」以豁免或拆任务。

#### I2（数据模型）Migration 003 `RENAME TABLE notices→content_posts` 不可幂等
- MySQL 无 `RENAME TABLE IF EXISTS`；版本化迁移按序执行一次可接受，但 Task 6.1 应注明「003 仅执行一次，勿重跑」，与「Migration 专项检查」的幂等要求对齐说明。

#### I3（数据模型）content_post_scope 注释「物理删除语义」与撤回保留行为措辞冲突
- 注释说「物理删除语义」，但 REQ-CPB-10 撤回**保留** scope 行，仅 draft 改写时 delete+reinsert。建议注释澄清「纯关联表无软删列；撤回保留由主表软删表达；改写物理替换」。

#### I4（interface-proto）`ResolveAdminDivision` 对「多 division grant」未定义守卫
- REQ-CPP-2 约定 community_admin 管辖唯一 division，故多 grant 出契约；但 Task 1.7 未对多 division 场景做 fail-closed 判定（0→取第一个？还是 080005？）。建议补一条守卫（多 division → 080005，与「唯一管辖」契约一致）。

---
VERDICT: REVISION
---

## 记忆遵守

- `[[proto-jstype]]`（int64 全 JS_STRING）、`[[restore-compensation-zero-time]]`（published_at sql.NullTime）、`[[grpc-only-comms]]`（禁直读 rel_user_role/uploaded_file）、`[[insert-ignore-swallows-errors]]`（显式 DELETE 撤销 421）、`[[permission-seed-api-path-must-match-routes]]`、`[[api-required-field-marked-optional]]`（community_id 必填勿 optional）等引用与代码现状核对一致，未发现误引用。
- 注意：M2（枚举数值跨层不一致）与本记忆库无直接对应，属新识别问题，建议沉淀为新 pitfall（「REST 枚举数值 ≠ proto enum 数值，禁止裸透传」）。
