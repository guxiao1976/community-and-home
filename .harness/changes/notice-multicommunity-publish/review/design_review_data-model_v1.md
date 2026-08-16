# Design Review — notice-multicommunity-publish（data-model + interface-proto 视角）

**审查维度**: 数据模型（字段/关系/Snowflake/时间/软删除）+ 接口契约（gRPC/Proto 自洽、破坏性标注、鉴权/错误码语义）
**审查对象**: `.harness/changes/notice-multicommunity-publish/design.md` + `tasks.md`
**对照**: 6 个 spec（notice-publish/notice-read/notice-moderation/publish-permission/attachment-security/notice-mobile）+ 实际代码证据（migration 001/002、community.proto、file.proto、permission AssertPublishScope、moderation task_handler、model/notice.go、init_permissions.sql）

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 2

## 已核实的正确设计（证据确认，非问题）

- 迁移断言与实际 DDL 一致：`notices.community_id BIGINT NOT NULL`、`published_at DATETIME NOT NULL`、`notice_attachments` 无 file_type、`uploaded_file` 无 file_type/confirmed；community-hub 现有 001/002、file 现有 001，故 003/002 编号正确。
- Proto 字段号无冲突：CreateNoticeRequest 现有至 7，新增 community_ids=8 / division_id=9；GetNoticeRequest 现有 id=1，新增 community_id=2；NoticeAttachment 现有至 4，新增 file_type=5；FileInfo 现有至 10，新增 file_type=11/confirmed=12。全部新字段号 + 全 int64 带 `[jstype=JS_STRING]`，兼容新增无破坏。
- 头注释错误码漂移声明与实际 proto 一致（community 080003/080005 漂移、file 070002/070003/070004/070005 漂移；70001-70003 不重编号）。
- Design Gate（REV-17）证据准确：`AssertPublishScopeLogic` 调 `resolveUserScope(scopeType='community')`，其实现 `g.ScopeType != scopeType → continue`，community_div 授权确实不被收集——判据必须变更的结论成立。
- GetUserRoles RPC 输出 `UserRoleInfo` 含 status/verified_at/expires_at（jstype JS_STRING），level-2 判定可基于 RPC 输出实现，无需直读 rel_user_role。
- masterdata `GetResidentialAreasByDivision` 存在（Req 含 community_div_id + status），division 展开契约可落地。
- 种子证据：property_admin(2)/community_admin(3)/committee(6)/owner(1)/tenant(5) 持 421，grid_worker(4) 未持——设计「授 421 给 4 + 回收 2/1/5」的回收/授权集正确；423/424 为全新码位。
- notice_scope 复合 PK = 唯一约束 + `idx_scope_community(community_id 左)` 满足 REQ-NP-2；物理删除 + 软删除策略与 REQ-NP-5 一致。

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|---------|------|---------|
| M1 | design.md §可见性门禁/§ListNotices/§GetMarqueeNotices/§GetNotice + tasks.md Task 1.4/1.8/1.10 vs Task 1.13 | **读门禁 pass 集合与回调 pass 集合自相矛盾**：Task 1.13/§UpdateNoticeModerationStatus 定义 pass={1=machine_pass,3=human_pass}（两者都置 published_at=NOW）；但所有读查询（Task 1.4 FindListByCommunity、Task 1.8 FindOnePublished、Task 1.10 FindMarquee）与 design 正文恒写 `moderation_status=1`。DB 存的是回调原始值（现有 updatemoderationstatuslogic 直接存 in.ModerationStatus），status=3 会真实落库。后果：一条 human_pass(3) 通知会被回调置上 published_at，却永远不出现在列表/详情/跑马灯——违反 REQ-NP-MOD-1「moderation status indicates pass」及本设计自身的 pass 集合。当前 moderation-service 回调只发 1/2（task_handler.go 只映射 machine_pass/machine_fail，3 尚未被发出），故为潜在缺陷；但本设计重写全部读查询 + 定义 pass={1,3}，必须携带一致的 pass 谓词，不能把矛盾带进代码。 | 全读路径统一 `moderation_status IN (1,3)`：抽一个共享谓词（如 `isPassed(status)` helper）供 FindListByCommunity/FindOnePublished/FindMarquee 与回调判断共用，杜绝两处漂移；design 正文「恒过滤 moderation_status=1」改为「恒过滤通过态(1/3)」。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|---------|------|------|
| S1 | design.md §CreateNotice（输入列表）+ tasks.md Task 1.7 | `notices.publisher` 是 NOT NULL 列，但 §CreateNotice 输入列表只列 title/content/community_ids/division_id/attachment_ids，仅声明 role/publisher_id 从 JWT 派生；publisher（展示名）的取值来源未明确。若执行方按输入列表字面实现（移除 publisher），INSERT 会因 NOT NULL 失败，发布主路径不可用。 | 在 §CreateNotice 显式声明 publisher 取值：保持请求体展示字符串（非安全字段，仅展示）或从 JWT 用户展示名派生（后者需补 user-service 调用，任务中缺失）；Task 1.7 落库清单明确包含 publisher 列。 |
| S2 | design.md §GetNotice / Proto 变更表 + tasks.md Task 0.1 | `GetNoticeRequest.community_id` 标必填（缺失→080005）在 wire 上兼容，但对只传 id 的既有调用方是**行为回归**——尤以本变更明确不做（Q6）的 web/pc 为甚：PC 详情页将全部 080005。design 已登记存量数据不可见（Q1），但未登记此 API 语义回归。 | 在 design/CHANGELOG 显式登记「GetNotice 必填 community_id 对未升级消费方（PC）的回归影响」；确认 PC 读路径同步处理（一并升级或版本隔离），避免上线即坏。 |
| S3 | tasks.md Task 1.6 | `AssertCommunitiesScope` 设计为「逐目标 AssertCommunityScope」（每目标一次 gRPC），但 permission `AssertPublishScopeRequest` 本身是批量接口（`repeated ScopeRef targets`，本变更 ≤100 目标）。逐目标是 N+1 次 RPC，且批量调用的 all-or-nothing 语义（permission 层遇任一失败即拒绝）与本设计目标一致。 | 改为单次批量调用：一次 `AssertPublishScope` 携带全部 target ScopeRef，060007→080006 映射一次完成；仅当需要区分逐目标失败时保留循环（但语义仍整体拒绝）。 |
| S4 | tasks.md Task 3.3 说明 | 新端点 `GET /api/perm/data-scopes` 的授权模型用「或」表述（「绑定 422 同款读角色集**或**经既有 read 权限放行」），未钉死。该端点返回调用者自身 scope_ids，属权限面决策，应在编码前定案。 | 在 Task 3.2/3.3 钉死：新增专用权限码绑定移动端角色，或显式复用 422 绑定集；删除「或」的双轨表述。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| I1 | `uploaded_file` 行仅在 ConfirmUpload 创建（GetUploadUrl 不落库），故 `confirmed` 对任何存在的行恒为 true；spec REQ-AS-7「未确认文件→confirmed=false」场景在现有流程下不可达（GetFileUrl 返回 not-found 而非 confirmed=false）。CreateNotice 的 confirmed 检查实际退化为「文件存在 + 归属本人」。功能安全成立，建议在 design 记录该语义，避免执行方误以为 confirmed 是状态机标志。 |
| I2 | `notice_scope` 仅含 created_at（无 updated_at/deleted_at），与纯关联表 + 物理删除语义一致，符合硬约束 #3.1（时间字段命名统一，非强制每表全列）；建议迁移注释注明以避审读误判。 |

## 报告自检
1. MUST FIX 定位到具体章节/任务（M1: design §可见性门禁 + Task 1.4/1.8/1.10 vs 1.13）✅
2. 每条附可落地修复建议 ✅
3. data-model + interface-proto 两维度均检查：数据模型（迁移/索引/软删除/时间字段/confirmed/published_at）✅；接口（Proto 字段号/jstype/破坏性/错误码/鉴权/幂等/必填回归/N+1）✅

---
VERDICT: REVISION（M1 为 MUST FIX，需架构设计师修订后复审）
---
