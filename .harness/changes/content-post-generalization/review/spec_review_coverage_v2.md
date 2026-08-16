# Plan Review — content-post-generalization（覆盖完整性 视角）

**审查维度**: 需求覆盖 / 场景完整性 / 边界识别 / NEEDS CLARIFICATION 遗漏
**审查版本**: fallback:r1:rc1（P1.3，与历史 r0:rc1 哈希不同 → 已按磁盘最新内容独立重新审查）
**审查对象**: proposal.md + specs/content-post-publish / read / moderation / permission / attachment-security 共 5 个 spec.md
**核对基线**: .change.yaml、docs/superpowers/specs/2026-08-16-content-post-design.md、migration/001_initial.sql、002_add_moderation_status.sql、model/notice.go、rpc/internal/logic/notice/helper.go（roleToString）、api-proto community.proto / permission.proto / masterdata.proto

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 6 / 🔵 INFO: 3

## 上轮问题修复验证（r0:rc1 → r1:rc1）
- MF-1（状态机存储映射）✅ 已修复 — REQ-CPB-1 status 枚举含 draft=0/submitted=1/approved=2/rejected=3/withdrawn=4；REQ-CPB-4 定义转换；draft→submitted 由 UpdateContentPost.status=submitted 触发（无独立 Submit RPC）
- MF-2（published_at 恒 NULL 致跑马灯恒空）✅ 已修复 — submit 即隐式通过置 status=approved + published_at=NOW()（REQ-CPB-1/4、REQ-CPM-5、REQ-CPR-1/3 一致）
- SF-1（保留字段清单）✅ 已修复 — REQ-CPB-1 显式保留 is_pinned/role/publisher
- SF-2（attachment_count 重算）✅ 已修复 — REQ-CPB-9 同事务重算 + 提交时冻结
- SF-3（边界场景）✅ 已修复 — REQ-CPB-5 补齐全部三边界；REQ-CPP-2 补 committee 场景
- SF-4（删除语义/RPC 契约）✅ 已修复 — REQ-CPB-10（软删+withdrawn+scope/附件保留）+ REQ-CPB-9（Update 契约）
- I-1/I-2/I-3 ✅ 已处理 — topic retention（REQ-CPM-1）、存量不可见登记（proposal 风险节）、UpdateNoticeModerationStatus 移除（D21/REQ-CPM-4）

## 决策覆盖核对（D1-D22 + REUSE:notice-D\*）
D1-D22 全部映射到对应 Requirement 且有正/异常场景（REQ-CPB-1 覆盖 D1/2/11/12，REQ-CPB-2 覆盖 D13，REQ-CPB-3 覆盖 D14，REQ-CPB-4/9 覆盖 D9/19，REQ-CPB-5 覆盖 REUSE:notice-D13/14/25/29/31，REQ-CPB-6 覆盖 REUSE:notice-D23/24，REQ-CPB-7 覆盖 D3/7/20，REQ-CPB-8 覆盖 D15，REQ-CPB-10 覆盖 REUSE:notice-D19，REQ-CPM-1 覆盖 D8/17/20，REQ-CPM-2 覆盖 D7，REQ-CPM-3/4 覆盖 D3/4/21，REQ-CPM-5 覆盖 D16/18，REQ-CPP-1/2/3 覆盖 D5/6/22，REQ-CPR-3 覆盖 D5，REQ-CAS-1/2/3 覆盖 REUSE:notice-D4/5/11）。决策点覆盖整体完整。

每个 Requirement 均满足「≥1 正向 + ≥1 异常场景」，边界场景覆盖良好。下述为本轮新发现的覆盖缺口。

## 发现

### 🔴 MUST FIX
| # | 文件:章节 | 问题 | 修复建议 |
|---|----------|------|---------|
| MF-1 | content-post-publish/spec.md REQ-CPB-1 | **community_id 去 NOT NULL 迁移未显式声明 + 无门禁场景（与 published_at 门禁不对称）**。REQ-CPB-1 场景「创建草稿 published_at 置空」THEN 明确要求新行 `community_id left unset (NULL)`，且正文「deprecate community_id (retain the column but SHALL NOT write new values)」。但现 schema `community_id BIGINT NOT NULL`（migration/001_initial.sql:12），spec 只对 `published_at` 显式声明 `ALTER TABLE ... MODIFY ... DEFAULT NULL` 并给了「未迁移则创建被拒」门禁场景，对 community_id 只写「弃用保留列」——未声明其可空化迁移，也无「community_id 未迁移则创建被拒」门禁场景（design.md §2.1 迁移节同样只写「弃用 community_id + published_at 去 NOT NULL」，未列 community_id 去 NOT NULL）。若实现按字面执行：新发布 INSERT 的 community_id 字段写 NULL 将违反 NOT NULL 约束 → 发布主链路不可用（与 published_at 门禁同类，但 community_id 侧缺失）。归档 notice 设计曾显式含 `MODIFY community_id BIGINT DEFAULT NULL`，本变更迁移声明未继承。 | REQ-CPB-1 显式补充 community_id 可空化迁移语句（`ALTER TABLE content_posts MODIFY community_id BIGINT DEFAULT NULL`）并新增门禁场景「community_id 未迁移则创建被拒」（与 published_at 场景对称）；同步更新 proposal migration 行、design.md §2.1、验收标准（「community_id 弃用保留列」→「community_id 弃用可空保留列」）。 |

### 🟡 SHOULD FIX
| # | 文件:章节 | 问题 | 建议 |
|---|----------|------|------|
| SF-1 | content-post-publish/spec.md REQ-CPB-1 + content-post-read/spec.md REQ-CPR-1 + content-post-permission/spec.md REQ-CPP-1 | **role 列取值空间未定义（RBAC 发布角色 → role 列字符串映射缺失）**。REQ-CPP-1 定义发布角色集 {grid_worker, community_admin, property_admin, committee}；REQ-CPR-1 role 过滤场景用 `role=grid_officer` 匹配 grid_worker 发布者；REQ-CPB-1 说 role 经「RBAC→publish-role 映射」派生。但 grid_worker→"grid_officer"、property_admin→"property"、community_admin→"community"、committee→"committee" 的映射关系未在任一 Requirement 定义（仅 helper.go roleToString 存 legacy 映射，spec 未引用）。实现者无法确定 role 列存什么值，role 过滤与 is_pinned 操作者的角色判定会漂移。 | 在 REQ-CPB-1（或 REQ-CPP-1）显式定义 4 个发布角色的 role 列取值映射（如 grid_worker→"grid_officer"、property_admin→"property"、community_admin→"community"、committee→"committee"，注明是否沿用 NoticeRole 字符串值），并同步 REQ-CPR-1 role 过滤场景措辞。 |
| SF-2 | content-post-publish/spec.md REQ-CPB-1 | **注册 section_code 板集合未枚举（哪些 code 本期合法未定义）**。REQ-CPB-1 说 section_code 白名单，异常场景用「unknown board code」演示，但「registered board set」具体包含哪些值（notice？repair 本期是否可建？）未列出；proposal out_of_scope 说 repair 本期仅 section_code 预留不实现板块逻辑，与「repair 为合法板集合成员」的边界模糊。 | 显式枚举本期合法 section_code 集合（如 {notice}），并明确 repair 等是否本期可写入（建议与「预留不实现」一致：本期仅 notice 合法）。 |
| SF-3 | content-post-publish/spec.md REQ-CPB-7 + content-post-moderation/spec.md REQ-CPM-1 | **pending-push 重推循环的 quarantine 边界未定义**。REQ-CPB-7/REQ-CPM-1 均说重推「until acknowledged or quarantined」，但 quarantine 的触发条件（重推次数上限？超时时长？人工介入？）未定义——存在无限重推、或永不从 pending-push 出队的风险边界。 | 定义 quarantine 判据（如重推 N 次/超时 T 后置 quarantine 标记 + 告警 + 人工介入），并补对应场景。 |
| SF-4 | content-post-publish/spec.md REQ-CPB-7 + REQ-CPB-10 | **撤回（Delete）与 pending-push 并存场景未覆盖**。submitted 时若 Kafka 不可用进入 pending-push，随后发布者撤回（DeleteContentPost → status=withdrawn），定时重推扫描仍会对已撤回帖重复投递 content-review → 后期消费者审核一条已撤回帖（状态回写会覆盖 withdrawn？）。无场景覆盖此交互。 | 补场景：withdrawn/pending-push 帖在重推扫描中的处理（跳过 / 标记不再投递 / 消费者侧过滤），明确撤回是否清除 pending-push 标记。 |
| SF-5 | content-post-publish/spec.md REQ-CPB-1 + REQ-CPB-5 | **publisher 展示名解析机制未定义（从哪个服务/RPC 取真实档案）**。REQ-CPB-1/5 要求 publisher 展示名取「authenticated user's real profile / 用户真实档案」，但未指明经哪个服务哪个 RPC 获取（user-service? auth? 现 community-hub 是否已有该调用通道?）。实现者无法落地该安全要求。 | 明确展示名解析的来源服务/RPC（或 JWT 中现有字段），并确认该 RPC 现网存在（或纳入本变更 api-proto 范围）。 |
| SF-6 | content-post-publish/spec.md REQ-CPB-1 | **新行 moderation_status/moderation_time 取值未定义**。002 迁移使 `moderation_status TINYINT NOT NULL DEFAULT 0`，REQ-CPB-1 说兼容期保留并「progressively transitioned」，但新 content_posts 行（status=approved）上 moderation_status 写什么值未定义（默认 0=pending 与 status=approved 并存是否合法？）。 | 明确新行 moderation_status/moderation_time 的写入策略（沿用默认 0 / 置与 status 对齐 / 读路径完全忽略），避免双状态字段语义漂移。 |

### 🔵 INFO
| # | 建议 |
|---|------|
| I-1 | REQ-CPR-1：ListContentPosts 传未知 section_code 的行为未定义（拒绝 080005 vs 空结果）。建议补一句明确。 |
| I-2 | REQ-CPB-6：附件绑定调 GetFileUrl 时 file-service 不可用/返回错误路径未覆盖（超时/降级处理）。建议补异常场景。 |
| I-3 | 废弃草稿（draft 永不提交）的生命周期/清理策略未定义（无过期清理）。建议登记为后续技术债。 |

## 问题跟踪表
| 问题 | 状态 |
|------|------|
| MF-1 community_id 可空化迁移 | 待修复（本轮新增） |
| SF-1 role 列取值映射 | 待修复（本轮新增） |
| SF-2 section_code 板集合枚举 | 待修复（本轮新增） |
| SF-3 quarantine 边界 | 待修复（本轮新增） |
| SF-4 撤回×pending-push | 待修复（本轮新增） |
| SF-5 publisher 展示名来源 RPC | 待修复（本轮新增） |
| SF-6 moderation_status 新行取值 | 待修复（本轮新增） |
| 上轮 MF-1/MF-2/SF-1..4/I-1..3 | 已修复（本轮验证通过） |

---
VERDICT: REVISION
---
