# Plan Review — notice-multicommunity-publish（清晰可执行视角 · v3）

**审查维度**: 粒度 / 歧义 / 一致性（SHALL/MUST 唯一解释、Scenario 具体到实现者得出相同行为、术语一致）
**审查版本**: P1.3（fallback:r2:rc1）— 与 r1:rc1 哈希不同，按磁盘最新内容独立重审
**审查对象**: proposal.md + specs/{notice-publish, notice-read, publish-permission, attachment-security, notice-moderation, notice-mobile}/spec.md
**对照基准**: 现状代码契约（community.proto 错误码头注释/`CreateNoticeRequest`/`GetNoticeRequest`/`NoticeAttachment`、file.proto 错误码块 + errcode.go 70001-70003、permission-service `grantSatisfiedLevel`/rel.go scope_type/GetDataScopes、init_permissions.sql §4.2/§4.8、masterdata `GetResidentialAreasByDivision`(三坐标分支 + status 参数)、community-hub CLAUDE.md 错误码登记、access-data-permission design.md）

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 4 / 🔵 INFO: 3

## 上轮（r1:rc1）问题修复验证

| v2 # | 问题 | 状态 |
|:---:|------|:---:|
| MF-1 | division_id 值域未定义（部分修复） | **已修复** — REQ-NP-3 明确 `division_id` 类型 int64、值域 `md_administrative_division.id`（community_div_id 同源），展开经 `GetResidentialAreasByDivision(community_div_id=division_id)`（与 RPC switch 分支 `CommunityDivId>0` 吻合，已核实现码 getresidentialareasbydivisionlogic.go） |
| MF-2 | doc OOXML 事实错误 | **已修复** — REQ-AS-3 明确 doc=OLE2/CFB（D0 CF 11 E0…）、docx=ZIP+OOXML 内容签名（D18） |
| MF-3 | 070003 整数冲突 | **已修复** — 新码 070004/070005（int32 70004/70005），70003 保持 ErrCodeFileOperationFailed 不重编号（D11）；已核实 errcode.go 70001/70002/70003 实存、70004/70005 空位，无冲突 |
| SHOULD 2 | REQ-NP-2 NOT NULL 未入正文 | **已修复** — 正文明确两列 SHALL 为 NOT NULL |
| SHOULD 3 | REQ-NP-6 错误码「或等价」 | **已修复** — 钉死 080005 参数无效（附件引用无效） |
| SHOULD 4 | GetNotice community_id 必填未定义 | **已修复** — REQ-NR-2 明确必填、缺失→080005 |
| INFO 1 | 070003/070004 主次码 | **已修复** — 固定 070004 主码（类型为第一安全边界）、070005 次码 |
| INFO 2 | 展开仅含审核通过小区 | **已修复** — REQ-NP-4 增加场景；残留 mechanism 缺口见本轮 SHOULD 1 |
| INFO 3 | 空关联场景未拆 | **已修复** — notice_id 缺失 / community_id 缺失分场景 |

上轮 MUST FIX 全部闭环，本轮为独立重审新增发现。

## 发现

### 🔴 MUST FIX

| # | 文件:章节 | 问题 | 修复建议 |
|---|---------|------|---------|
| 1 | notice-publish REQ-NP-3 / proposal §影响范围(api-proto) | **080003/080005 与 community.proto 头注释漂移未处置，spec 仅对齐了 file.proto（D11）漏了 community.proto**。spec 将 080003 定义为「单次发布目标数超限」（REQ-NP-3 + D13 + 验收），080005 定义为「参数无效」（空范围/双载/非 admin 传 division/附件引用无效/GetNotice 缺 community_id）。但 community.proto 头注释（已核实行 31/33）仍登记「080003 — 寻失发布次数已达上限」「080005 — 小区ID无效」——且 080003 实为**陈旧**：寻失配额现码是 `CodeSectionQuotaExceeded=80007`（section_quota.go:13，API 面 080007），服务侧权威登记（community-hub CLAUDE.md「080003(超限)/080005(参数无效)」、access-data-permission design.md「080003=超限」）均与 spec 一致。即：spec 语义与「服务侧文档」一致，但与「proto 头注释」同一整数两语义，实现者读 proto 头注释会得到与 spec 相悖的解释。spec 已在 D11 将 file.proto 头注释对齐纳入变更范围，却未对 community.proto 做同等登记——同类漂移、同类 MUST（D11 即 REVISION-3 clarity MF-3 定案）。 | 在 proposal §影响范围 api-proto 行 + REQ-NP-3 显式登记：**community/v1 头注释错误码块对齐实际语义**（080003=目标数超限、080005=参数无效（含小区ID无效），剔除过时的「080003 寻失发布次数已达上限」注释——该语义实际码为 080007），并写入 community CHANGELOG；杜绝同整数双语义，与 D11 的 file.proto 处理同标准。 |

### 🟡 SHOULD FIX

| # | 文件:章节 | 问题 | 建议 |
|---|---------|------|------|
| 1 | notice-publish REQ-NP-3/REQ-NP-4 | **division 展开未指定 status 参数取值**。`GetResidentialAreasByDivisionReq`（masterdata.proto:171）有 `status`（0=all, 1=approved only），spec 要求「展开仅含审核通过小区」但只写「invoking … with `community_div_id = division_id`」，未显式传 status=1（否则实现者要么传 0 后自行过滤，要么漏过滤）。 | 在 REQ-NP-3/REQ-NP-4 明确「调用 GetResidentialAreasByDivision(community_div_id=division_id, **status=1** 仅审核通过)」，使机制确定化。 |
| 2 | notice-publish REQ-NP-6 | **附件「存在/已确认/属发布者」校验的机制（哪个 file-service RPC）未指定**。当前 file.proto 仅 GetUploadUrl/ConfirmUpload/GetFileUrl/DeleteFile/ListFiles，无「按 id 校验归属+确认状态」专用 RPC；现有 createnoticelogic.go 亦未消费 attachment_ids。实现者无法确定数据源。 | 钉死校验数据源：复用 ListFiles（按 user+entity_type 过滤）或 GetFileUrl，或新增 file-service 校验 RPC（明确新增则纳入 api-proto 变更范围）。 |
| 3 | notice-mobile REQ-NM-5 / publish-permission | **社区管理员 division 选择器的数据源与现状 GetDataScopes 契约冲突**。REQ-NM-5 称「范围选项数据源为 permission GetDataScopes」且渲染「可选的 division」；但 GetDataScopes 现 scope_type 仅 community/building/unit/grid（permission.proto GetDataScopesRequest 注释 + rel.go 无 community_div），不会返回 division。spec 未说明 GetDataScopes 是否扩展返回 division（api-proto 变更清单未列）或前端改从 masterdata 取 division 树。 | 明确 division 选项数据源（GetDataScopes 扩展 community_div scope_type 并纳入变更范围，或前端调 masterdata GetDivisionTree），并说明与 REV-17 design gate 的关系。 |
| 4 | publish-permission REQ-PP-1 / notice-publish REQ-NP-3 | **sys_admin `can_publish=false` 仅约束入口显隐，写路径（421 + min_verf_level=2）并不拦截 sys_admin 直接 CreateNotice，spec 未声明该行为是否有意**。sys_admin（role_id=8）经种子持全部权限含 421，level-2 满足则写路径放行——移动端「入口隐藏」不等于「不能发」。实现者可能误加隐藏写路径拦截（与 REQ-PP-3 冲突）或漏防。 | 显式声明：sys_admin 写路径**不额外拦截**（管理台为发布面，移动端仅入口隐藏），或补充移动端写路径拦截契约，二选一写入，消除实现歧义。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | REQ-PP-1 措辞「sys_admin holds 421 via the all-permissions seed (is_system role)」与权限服务 CLAUDE.md「is_system 不自动授全权限」微矛盾——421 实来自 init_permissions.sql role_id=8 的显式 `SELECT 8, id FROM sys_permission` 绑定；建议改为「经种子对 role 8 显式全量绑定」。 |
| 2 | REQ-NR-2 未在 spec 内标注 GetNoticeRequest.community_id 的字段号（proposal §影响范围给 2）；建议 spec 内一并写出字段号，保证「spec 自足」。 |
| 3 | GetMarqueeNotices 返回「notice id + title 摘要」但未定义响应消息结构（新消息如 NoticeMarqueeItem 或复用 Notice）；建议设计阶段明确，避免实现者各自造轮子。 |

## 问题跟踪表

| # | 状态 |
|---|------|
| 1（community.proto 头注释漂移未登记） | 待修复 |

---
VERDICT: REVISION
---
