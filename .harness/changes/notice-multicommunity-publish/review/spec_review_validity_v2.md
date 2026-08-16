# Plan Review — notice-multicommunity-publish（业务有效性视角）

**审查维度**: 业务自洽 / 非功能（安全·性能·兼容） / 合规
**审查版本**: P1.3（fallback:r1:rc3）— 磁盘最新内容独立审查（spec/proposal 已于 8-16 00:27-00:28 更新，本轮不沿用旧轮结论）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 0 / 🟡 SHOULD FIX: 6 / 🔵 INFO: 3
- VERDICT: **APPROVED**

## 上轮（v1，fallback:r0:rc3）MUST FIX / SHOULD FIX 复验（磁盘最新内容）

| # | 上轮问题 | 当前磁盘状态 | 结论 |
|---|---------|-------------|------|
| 1 | **MUST FIX-1：published_at 语义迁移（D27）与 `published_at DATETIME NOT NULL` 无迁移/占位契约，创建主链路业务不可用** | **已修复**：REQ-NP-1 现显式声明第二迁移 `ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL`（D30），并补「published_at 未迁移则创建被拒（迁移先于上线门禁）」独立场景；REQ-NP-MOD-4 场景 1/5 同步（创建写 NULL、通过回调置 now）；proposal 验收标准 108 行登记异常门禁 | ✅ |
| 2 | **MUST FIX-2：跑马灯 15 天窗口验收标准与场景正文矛盾（审核滞留 vs 通过后）** | **已修复**：D32 统一窗口锚 = published_at（审核通过日），删除「审核滞留 >15 天」歧义表述；REQ-NR-3 场景「created 20 天前、今天通过 → 入跑马灯」「published_at 距今 >15 天 → 不入跑马灯但入浏览列表」；proposal 验收标准 118 行同步为「now-published_at >15 天」 | ✅ |
| 3 | v1 SHOULD-1：附件绑定边界（重复引用/已删/总量） | REQ-NP-6 已定义「存在/confirmed/归属校验 + ≤10 个/≤50MB 总量（FileInfo.file_size）→ 080005」；「已被删除（GetFileUrl 404）」与「attachment_ids 内重复 file_id」仍未显式归 080005 | 🟡 部分解决（见 SHOULD-6） |
| 4 | v1 SHOULD-2：第二层 magic-bytes 回读 fail-closed 未定义 | REQ-AS-3「仅当声明扩展名与嗅探类型一致才接受」已隐含 fail-closed，但 IO/解析失败仍未显式登记 | 🟡 沿用（见 SHOULD-2） |
| 5 | v1 SHOULD-3：can_publish 未叠加 421/数据范围非空 | REQ-PP-1 仍仅按 level-2 角色状态判定；spec 明确 can_publish 为入口提示、不替代写路径校验（设计上正确），但空数据范围死入口未处置 | 🟡 沿用（见 SHOULD-1） |
| 6 | v1 SHOULD-4：division 展开错误码优先级 | D31 已给出 080005/080006 划分原则（请求形状类 vs 目标级解析/越权），与现有 scope.go（unknown node→060007→080006，CodePublishScopeDenied=80006）一致 | ✅ 已解决 |

## 代码事实核对（本视角独立验证）

| 事实 | 磁盘证据 | 与 spec 一致性 |
|------|---------|--------------|
| notices.community_id / published_at 现为 NOT NULL | migration/001_initial.sql:12,19 | REQ-NP-1 D30 迁移契约吻合 |
| 现创建逻辑创建时即写 published_at | createnoticelogic.go:62 `PublishedAt: time.Now()` | D27 行为登记（创建时写 NULL→通过回调置 now）准确 |
| 越权/未知节点 → 080006 | scope.go:24 CodePublishScopeDenied=80006；unknown node→060007→080006 | REQ-NP-3 的 D31 080006 映射与现行为一致 |
| DeleteNotice 现按数据范围判定 | deletenoticelogic.go:38 CheckPublishScope | REQ-NP-5 收窄为「仅发布者本人」属实行为回归，已登记 |
| min_verf_level 机制真实存在 | checkpermissionlogic.go + helpers.go:26 grantSatisfiedLevel（level-2=status==2 AND verified_at NOT NULL AND 未过期） | REQ-PP-3 写路径门槛（421 置 min_verf_level=2）可实现 |
| 种子角色映射 | init_permissions.sql：role2=property_admin、3=community_admin、4=grid_worker、5=tenant、6=committee、8=sys_admin；421 现授予 role 1/2/3/5/6（+8 全量）；grid_worker(4) 现无 421；§4.2 置 min_verf_level=0 | REQ-PP-4 授 4/回收 1,2,5/翻 2 与现网事实完全吻合 |
| FileInfo 含 user_id | file.proto FileInfo.user_id=2 | REQ-NP-6 归属校验（GetFileUrl→FileInfo.user_id==认证用户）可实现 |
| division 展开 status=1 仅审核通过 | getresidentialareasbydivisionlogic.go:52 `status==1 && SubmissionStatus!=2 skip` | REQ-NP-3/REQ-NP-4 展开契约吻合 |
| notice_attachments 按 notice_id 关联 | 001_initial.sql:28-36 | 无 community_id 隐藏冲突，多小区模型自洽 |

## 发现

### 🔴 MUST FIX
无。

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|-----------|------|------|
| 1 | specs/publish-permission/spec.md REQ-PP-1 | can_publish 仅基于 level-2 角色状态，未叠加「数据范围非空」。level-2 已认证但社区数据范围 EMPTY 的 grid_worker/committee 会拿到 can_publish=true 并显示发布入口，但每次发布必失败（080006/无范围选项）——空转死入口。 | GetPublishPermission 叠加「该角色存在非空社区数据范围」判定；或明确前端以 can_publish + GetDataScopes 非空共同决定入口显示。 |
| 2 | specs/attachment-security/spec.md REQ-AS-3 | 第二层 ConfirmUpload magic-bytes 回读的失败语义未显式定义：MinIO 回读不可达 / 对象读取 IO 异常 / 容器结构解析失败（非「嗅探到非白名单类型」而是「无法解析」）时的处置未写。当前「仅当声明与嗅探一致才接受」隐含 fail-closed，但未显式登记，存在实现侧对校验器异常乐观放行的空间。 | 显式声明：回读/解析失败一律拒绝（070004 或依赖错误码），绝不因校验器异常而放行入库；补「MinIO 回读不可用 → 拒绝且不注册元数据」场景。 |
| 3 | specs/notice-publish/spec.md REQ-NP-4 + specs/notice-mobile/spec.md REQ-NM-5 | division→community 授权机制整体悬置：permission 侧无 community_div scope_type、GetDataScopes 不返回 division 范围，spec 已挂 design gate（REV-17）。但 REQ-NM-5 新增的「division 选项按用户 county 区域过滤」未定义 county 区域来源（既无 scope 也无接口），移动端社区管理员选项列表无可落地数据源。community_admin 多小区发布是本变更核心特性，若设计评审未通过将整体不可用。 | 强化 design gate 为硬门禁（未验证不得进入编码）；REQ-NM-5 补 county 区域来源（如 user-service 房产/住址归属或 masterdata 登录区域），或标注为 design gate 决议项。 |
| 4 | specs/notice-moderation/spec.md REQ-NP-MOD-4 | 通过回调时 moderation_status 与 published_at 的设置未要求原子性。若分两条语句/事务更新，崩溃窗口会留下「已过审但 published_at=NULL」的行——MySQL DESC 排序 NULL 置顶（该行排在列表第一位），详情 published_at 显示 0。 | 要求回调在单条 UPDATE（`SET moderation_status=?, published_at=?`）或同一事务内原子设置；补「回调更新失败保持 fail-closed（状态不置为通过）」说明。 |
| 5 | specs/publish-permission/spec.md REQ-PP-4 | 权限种子变更（授 grid_worker 421 + 翻转 min_verf_level=2 + 回收 property_admin/owner/tenant 421）未登记缓存失效。permission-service 规则「修改角色/权限必须批量刷新 Redis 缓存」；perm:def（min_verf_level）缓存 TTL 30min、用户角色集若有缓存则更久——不失效则存量会话在 TTL 内持陈旧权限（grid_worker 仍被拦、property_admin 仍可发）。 | REQ-PP-4 补「种子变更后必须批量失效权限/角色缓存（或部署序列含缓存预热）」验收项，与 permission-service 规则 1 对齐。 |
| 6 | specs/notice-publish/spec.md REQ-NP-6 | attachment_ids 重复引用未定义：同一 file_id 出现两次时，≤10 计数与 ≤50MB 求和会重复计；已删除/不存在的 file_id（GetFileUrl 404）也未显式归 080005。 | 明确：attachment_ids 内重复 file_id 去重后校验（或直接 080005）；GetFileUrl 404 → 080005 参数无效。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | GetNotice 新增必填 community_id（字段号 2）对存量客户端（PC 端，Q6 顺延）是行为破坏（缺字段→080005）。建议在影响范围表登记 PC 兼容期，避免排期 PC 时遗漏。 |
| 2 | GetResidentialAreasByDivision 现硬编码 page=1/pageSize=1000（getresidentialareasbydivisionlogic.go:34）；division 展开 >1000 会静默截断（截断结果仍 >100 → 080003，不绕过 100 上限）。建议设计阶段确认展开上限与 100-cap 的关系。 |
| 3 | 用户同时持有多个可发布角色（如 grid_worker + committee）时，通知 role 记录（REQ-NP-4 映射）未定义优先级。建议在 design 阶段定夺。 |

## 问题跟踪表

| 编号 | 问题 | 状态 |
|------|------|------|
| 1-2 | 上轮 MUST FIX（published_at 迁移 / 跑马灯窗口锚） | 已验证修复 |
| 3 | can_publish 未叠加数据范围非空 | 待修复（SHOULD） |
| 4 | magic-bytes 回读 fail-closed 未显式化 | 待修复（SHOULD） |
| 5 | division→community 授权 + division 选项来源悬置（design gate 强化） | 待修复（SHOULD，设计评审门禁） |
| 6 | published_at 回调原子性 | 待修复（SHOULD） |
| 7 | 权限种子变更缓存失效 | 待修复（SHOULD） |
| 8 | attachment_ids 重复/已删引用 | 待修复（SHOULD） |
| 9-11 | 见 INFO | 建议采纳 |

---
VERDICT: **APPROVED**（0 MUST FIX —— 上轮 published_at 迁移契约缺失与跑马灯窗口锚矛盾均已按 D30/D32 修复；本视角无 CRITICAL 级一票否决项）
---
