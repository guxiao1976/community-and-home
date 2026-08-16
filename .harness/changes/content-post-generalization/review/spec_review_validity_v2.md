# Plan Review — content-post-generalization（业务有效性视角）

**审查维度**: 业务自洽、非功能（安全/性能/兼容）、合规、架构冲突/技术债/依赖风险
**审查版本**: P1.3 (fallback:r1:rc1) — 与上一轮 r0 哈希不同，按磁盘最新内容独立重新审查
**审查对象**: proposal.md + specs/content-post-{publish,read,moderation,permission,attachment-security}/spec.md + .change.yaml + 现网 JWT claims / community-hub 依赖 / FileInfo / NoticeRole 枚举 / init_permissions.sql 交叉核对

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 3 / 🔵 INFO: 4

## 上一轮问题修复校验（r0 → r1）

| 上轮 # | 级别 | 问题 | 本轮状态 |
|--------|------|------|---------|
| 1 | MUST | published_at 本期恒 NULL → 跑马灯恒空 + 列表排序不确定 | ✅ 已修复。REQ-CPB-1/REQ-CPB-4/REQ-CPM-5 统一为「submit 即置 status=approved + published_at=NOW()」，REQ-CPR-1/REQ-CPR-3 增加 NULLS LAST 防御并显式声明本期可见帖必非空；proposal 风险节「两阶段发布」已同步。 |
| 2 | SHOULD | attachment_count 在 draft 附件编辑时重算不变量未定义 | ✅ 已修复。REQ-CPB-9(b) 定义同事务重算不变量 + REQ-CPB-8 提交时冻结。 |
| 3 | SHOULD | Kafka best-effort 可静默丢失 → 审核盲区 | ✅ 已修复。REQ-CPB-7/REQ-CPM-1/D20 改为 at-least-once（落库待推标记 + 定时重推 + pending-push 可观测 + 业务风险显式登记）。 |
| 4 | SHOULD | content_posts 模型缺 role/publisher 列 + 展示名来源未定义 | ⚠️ 部分修复。列已显式列出（REQ-CPB-1），role/publisher 禁请求体信任已声明（REQ-CPB-1/REQ-CPB-5）；但「真实档案」**数据来源未定义**且 user-service 不在本变更服务范围，见下方 SHOULD #1。 |

## 发现

### 🔴 MUST FIX

无。

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| 1 | content-post-publish/spec.md REQ-CPB-1 / REQ-CPB-5 | **`publisher` 展示名取「用户真实档案」无数据来源，且 user-service 不在本变更范围 → 安全姿态不可实现**。REQ-CPB-1/REQ-CPB-5 要求 publisher 取真实档案、禁请求体信任，但：现网 JWT access token claims 仅含 user_id/jti/roles/exp/iat，无姓名 claim（auth-service/registerlogic.go:160）；community-hub-service 现无任何 user-service gRPC 客户端（communityhub.yaml 仅 MasterData/Moderation/Permission 三个 Rpc）；本变更 .change.yaml services 列表不含 user-service、api-proto 变更清单无 user 相关 RPC；现网唯一可提供姓名的入口是 user-service `GetUser`，但 community-hub 无该依赖。实现者无既定载体 → 要么新增未规划的跨服务依赖（架构范围外溢），要么回归信任请求体（重开伪造向量，与 REQ-CPB-5 场景「伪造展示名被纠正」冲突）。 | 明确来源并纳入变更范围：方案 A（推荐）新增 community-hub→user-service `GetUser` 依赖 + api-proto 登记（user-service 列入 services 列表），发布时经 gRPC 取实名；方案 B 为 access token 扩展 name claim（auth-service 签发处，需评估存量 token 兼容）。REQ-CPB-1/REQ-CPB-5 写出具体 RPC 与字段；设计评审阶段验证可用性。 |
| 2 | content-post-read/spec.md REQ-CPR-1 | **ListContentPosts 未显式要求「调用方对请求 community_id 有读权限」校验，存在数据范围泄露向量**。REQ-CPR-1 只写「scope 包含该 community_id 即返回」，未像 REQ-CPR-2（detail：caller has read access）与 REQ-CPR-3（marquee：restricted to caller's allowed communities）那样显式声明读范围门禁；服务职责边界虽提「FilterAllowed 语义，复用 notice 读路径」，但 requirement 正文缺位 → 实现者若只按 REQ-CPR-1 实现，任意移动端用户可传任意 community_id 枚举任一小区内容。 | REQ-CPR-1 正文补显式条款：调用方对请求 community_id 必须通过读权限数据范围校验（permission-service FilterAllowed 语义），未授权 → 与 detail 一致不透出内容（080001 或空列表，语义与 detail 对齐），并补「请求未授权小区」异常场景。 |
| 3 | content-post-read/spec.md REQ-CPR-1（role 过滤）+ content-post-publish/spec.md REQ-CPB-1 | **多发布角色用户一次发布多小区时，单值 `role` 列取值规则未定义 → role 过滤不确定**。REQ-CPP-1 返回可发布角色列表（复数），用户可在不同小区持不同发布角色（如 C1=grid_worker、C2=committee）一次发布到 [C1,C2]；content_posts.role 为单值列，spec 未定义取哪个角色，REQ-CPR-1 的 role 过滤对多角色发布者结果不确定（同一帖只能命中一种角色过滤）。 | 定义确定性取角色规则，三选一并写异常场景：按发布目标集合约束（多目标必须同角色，否则 080005 拒绝）/ 优先级取角色 / 或 role 下沉为每 scope 行。同时给出 role 列字符串值 ↔ NoticeRole 枚举（COMMUNITY/COMMITTEE/PROPERTY/GRID_OFFICER）的显式映射。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 4 | content-review 消息无 post 状态/撤稿标志：提交后撤回（withdrawn）的帖子其已投递/待推消息，未来消费者（D18）仍会审核并回写 withdrawn 帖的 status（deleted_at 已置、读路径仍不返回，无用户可见危害但浪费审核 + status 可能被覆盖）。建议：重推扫描显式跳过软删/withdrawn 帖，或契约注明消费者回写前复检当前状态。 |
| 5 | 列表/跑马灯 published_at DESC 在 DATETIME 秒级精度下，同秒提交多帖排序不确定（分页稳定性边界）；跑马灯（section_code 过滤 + published_at 窗口 + is_pinned desc, published_at desc 排序）无配套 (section_code, published_at) 索引，靠 scope join 后内存排序，大社区规模需评估（已声明读路径靠索引兜底）。可加 tie-break（id DESC）与索引评估。 |
| 6 | 权限码 421 名称仍为 `community:notice:create-api` 未随 CreateContentPost 改名（上轮已提，延续技术债）；section_code 白名单本期未枚举（notice + repair 预留，repair 本期能否创建未明），建议明确；DeleteContentPost 可删状态集（draft/submitted/approved）不含 rejected，未来消费者阶段 rejected 帖发布者无法撤回（本期无 rejected 帖，仅提示）。 |
| 7 | 权限种子 min_verf_level 0→2 与撤销 (1,421)/(5,421) 为行为变更，已与现网 init_permissions.sql:201-202/252-253 核对一致并登记风险；file.proto FileInfo 新增 file_type(11)/confirmed(12) 非破坏，与 REQ-CAS-5 一致。无漂移。 |

## 问题跟踪表

| # | 状态 |
|---|------|
| r0-1 | 已修复（published_at 锚定） |
| r0-2 | 已修复（attachment_count 重算不变量） |
| r0-3 | 已修复（Kafka at-least-once） |
| r0-4 | 部分修复 → 本轮 SHOULD #1（publisher 真实档案来源缺载体） |
| 本轮-2 | 待修复（List 读范围门禁显式化） |
| 本轮-3 | 待修复（多角色 role 取值规则） |

---
VERDICT: **APPROVED**（无 MUST FIX。上轮 MUST（published_at 恒 NULL）已彻底修复；本轮 3 项 SHOULD 建议在架构设计评审阶段落实，其中 SHOULD #1（publisher 真实档案来源）建议设计 gate 前置确认数据源归属，避免实现期新增未规划跨服务依赖）
---
