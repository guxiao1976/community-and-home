# Proposal: 通知模块多小区发布 + 通栏跑马灯 + 附件安全

> **优先级**: P1 · **改动规模**: 大（L） · **影响风险**: 高
> **核心风险点**: 跨 5 个服务协同（community-hub / file / permission / master-data / api-proto）+ 数据模型变更（notices 弃用 community_id 需迁移去 NOT NULL **且 published_at 需迁移去 NOT NULL（D30，与 D27 审核通过时设置语义一致）**）+ 权限种子变更（grid_worker 授 421、property_admin 收 421、owner·tenant 收 421、421 置 min_verf_level=2，均属行为回归）+ 附件安全两层校验（第二层 magic-bytes 判定）；通知模块是首页核心内容，回归面大
> **变更类型**: modify（改造既有「单小区通知」；原行为→新行为 diff 见 §影响范围）
> **设计文档**: `docs/superpowers/specs/2026-08-15-notice-module-design.md`（用户已拍板）+ 用户已拍板决策包（Q-ATTACH-TOTAL-CAP / Q-ATTACH-VERIFY-CARRIER / Q-CREATE-IDEMPOTENCY / Q-MAGIC-BYTE-WIDENING / Q-PROPERTY-ADMIN-MOBILE / Q-PUBLISHED-AT-ANCHOR / Q-WITHDRAW-ATTACHMENTS，映射 D23-D28）
> **REVISION 已修订**: 第 1-4 轮 4 视角评审 MUST FIX 已逐条修订（决策日志 D9-D30）；spec 已按磁盘最新内容收敛（见各 spec「REVISION 已解决」标注）。**第 4 轮 REV-1/REV-2/REV-3/REV-4 四项已解决**：① published_at 迁移去 NOT NULL（D30，方案 a：创建写 NULL、审核通过回调置 now）；② 080005/080006 错误码消歧（目标级解析失败统一 080006，080005 仅保留请求形状类）；③ 未迁移 published_at 时创建失败异常门禁；④ 跑马灯 15 天窗口统一以 published_at（审核通过日）为锚（删除「审核滞留 >15 天」歧义表述）

## 为什么做

当前通知模块是**单小区通知**：notices 表 `community_id` 单值（BIGINT NOT NULL），一次发布只对一个小区，网格员需逐个小区重复发布；首页只有简单公告列表，无跑马灯聚合；附件上传无类型/大小安全校验，存在上传可执行文件的风险。

本变更把通知升级为**首页核心内容 + 可复用模块**：一次发布可覆盖多个小区（多角色按数据范围发布，范围关联走独立 notice_scope 表），首页通栏跑马灯聚合最近 15 天通知（published_at 自审核通过日起算），附件走 file-service 白名单 + 大小 + **单通知总量**校验（两层：GetUploadUrl 快速拒绝 + ConfirmUpload magic-bytes 回读）。让网格员/社区管理员/业委按各自管辖范围高效发布（**property_admin 移动端发布角色本变更剔除**，收敛为 grid_worker/community_admin/committee），业主侧有统一的阅读入口，同时封堵附件上传的安全缺口。

## 做什么

1. **数据模型升级（community-hub-service）**：notices 弃用 `community_id`（**迁移去 NOT NULL，兼容期保留列不再写入**，查询改走 `notice_scope`）；**`published_at` 迁移去 NOT NULL（D30：`ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL`，创建时写 NULL、审核通过回调置 now，与 D27 语义一致；不迁移则创建 INSERT 违反 NOT NULL，发布主链路不可用）**；新增 `notice_scope` 关联表（notice_id + community_id 均 NOT NULL，`uk_notice_community` 唯一约束 + **`community_id` 读索引**）；`notice_attachments` 加 `file_type` 字段
2. **多小区发布**：CreateNoticeRequest 新增 `repeated community_ids`（新字段号 8，`[jstype=JS_STRING]`，旧 `community_id` deprecated 保留）；发布落库前走 `AssertPublishScope(user, targets)` 逐目标校验（任一越权整体拒绝）；**目标级解析失败（不存在/不可解析的 community_id）统一 080006，与 scope 未覆盖同处理（D31，fail-closed）**；**写路径角色状态门槛 = 种子给 421 置 min_verf_level=2**（功能层强制已认证，与 can_publish 判定一致）；社区管理员按 `division_id`（值域 `md_administrative_division.id`）提交，由后端经 masterdata `GetResidentialAreasByDivision(community_div_id=division_id, status=1 仅审核通过)` 展开为具体小区快照；`division_id` 仅 community_admin 可用；单次发布目标数上限 ≤100（超限 080003）；`publisher_id`/`role` 以 JWT 实际身份/实际角色为准（不信任请求体）；**撤回复用 `DeleteNotice`（仅发布者本人，全局生效，notice_scope 物理删除；notice_attachments 行与 MinIO 对象全部保留，D28）**；**CreateNotice 后端声明不幂等（D25：不引入幂等键，重复提交产生重复通知，防重由前端提交中禁用承担）**
3. **读路径按 scope 过滤**：ListNotices 保持 `community_id` 入参（现有 `role` 筛选字段保留，语义不变），经 notice_scope 关联过滤当前小区可见通知；响应保持单 `community_id`（取请求小区，由 notice_scope 匹配行派生，不读弃用列）；GetNotice 新增 `community_id` 请求上下文（**必填**，字段号 2，缺失 → 080005），scope 外/不存在/审核未过 → 080001（不泄露）；**跑马灯数据新增专用 RPC `GetMarqueeNotices(community_id)`**（返回 ≤10 条 title 摘要，置顶优先 + 最近 15 天倒序 + 仅审核通过；community_id 缺失/空 → 080005）；**`published_at` 语义锚定审核通过时（D27/D30：经 UpdateNoticeModerationStatus 通过回调时设置、创建时写 NULL，窗口从可见日起算；原实现创建时即设，属行为登记）**
4. **GetPublishPermission 新增 RPC**：后端单一判定（level-2：status=2 且 verified_at NOT NULL 且未过期，`expires_at==0 OR > now` 基于 RPC 输出），返回可发布角色 + `can_publish`，驱动【我的】页入口显隐（前端不判断权限）；**可发布角色收敛为 grid_worker/community_admin/committee（D26）**；**property_admin 不在移动端可发布角色集**（can_publish=false + 种子回收 421，行为回归：物业原可发布通知 → 只读）；sys_admin 不在移动端可发布角色集（D16，管理台是 sys_admin 的发布面，写路径不额外拦截）
5. **权限种子变更（permission-service）**：给 grid_worker 授 `community:notice:create-api`(421)（否则网格员发布在功能权限层被 080002 拦截）；**把 421 的 `min_verf_level` 由 0 改为 2**（写路径角色状态门槛，与 level-2 判定一致）；**回收 property_admin（role 2）的 421（D26）**；回收 owner/tenant 的 421 绑定（业主/租户只读）；**删除权收窄登记**：DeleteNotice 由「现 CheckPublishScope 数据范围判定（辖区内可删）」收窄为「仅发布者本人」——community_admin 等失去辖区内删除他人通知的能力，属行为回归，与 421 回收回归一并登记
6. **附件安全（file-service）**：白名单 `png/jpg/jpeg/gif/pdf/doc/docx`，禁止 `exe/bat/sh/cmd/com/msi/apk/js/vbs/ps1/py/pl/php` + 所有 `zip/rar`，单文件 ≤10MB（硬上限）**且单通知附件 ≤10 个 且 总大小 ≤50MB（D23，总量上限在 CreateNotice 绑定校验**）；两层校验：GetUploadUrl 快速拒绝 + ConfirmUpload 回读 MinIO 实际对象**按 magic-bytes 内容嗅探**判定真实类型（非扩展名）；docx 按 ZIP+OOXML `word/document.xml` 内容签名放行、**doc 按 OLE2/CFB（D0CF11E0）+ `WordDocument` 流识别放行（防 .msi/.xls/.ppt 改 doc 绕过）**；**附件引用校验载体 = 扩展 FileInfo（新增 `file_type` + `confirmed` 字段）+ 复用 GetFileUrl 链路（D24）**，community-hub 经 GetFileUrl(file_id) 校验存在 + confirmed + user_id 归属，file_type 从该契约回读（非客户端回传）；**新增登记错误码 070004（类型不支持，int32=70004）/070005（大小超限，int32=70005）**，既有 70001-70003 不重编号、file.proto 头注释错误码块对齐实际常量；白名单全局基线 + 按 entity_type 可扩展
7. **移动端（web/mobile）**：首页通栏跑马灯 NoticeMarquee（消费 GetMarqueeNotices，最近 15 天标题滚动，更多→浏览页）；浏览页 notice-browse 按发布时间倒序；详情页 notice-detail 标题/正文/时间/附件；【我的】页 can_publish 才显示发布入口；发布表单（标题/正文/附件上传/范围选择，**提交中禁用防双击，D25**）；组件化 NoticeMarquee / NoticePublisher / NoticeList / NoticeDetail + 附件安全通用校验（含单通知总量预校验）
8. **审核流**：复用现有 moderation（发布→审核→通过才可见），不新增审核链路；**通过回调时设置 published_at（D27，创建时写 NULL，D30）**；`is_pinned` 继续走既有 UpdateNotice（不新增置顶 UI）

## 决策日志（Step 2，Step 8 追溯唯一依据）

| ID | 决策内容（结论） | 依据 |
|----|-----------------|------|
| D1 | notices 弃用 community_id（兼容期保留列不写入），范围关联单源走 notice_scope | 设计文档 §2.1/§2.2（用户确认） |
| D2 | 读/响应保持单 community_id（最小改动） | Q2 用户拍板 |
| D3 | 新增 GetPublishPermission RPC（后端单一判定，前端不判权限） | Q3 用户拍板 |
| D4 | 附件白名单全局基线 + 按 entity_type 可扩展 | Q4 用户拍板 |
| D5 | 两层校验：GetUploadUrl 快速拒绝 + ConfirmUpload 回读实际对象校验 | Q5 用户拍板 |
| D6 | 仅移动端（web/mobile），PC 后续单独排期 | Q6 用户拍板 |
| D7 | 跑马灯：置顶优先 + 最近 15 天倒序 + 单条点击进详情 + 封顶 10 条 | Q7 用户拍板 |
| D8 | 不处理存量（旧通知不可见，后续可补迁移回填） | Q1 用户拍板 |
| D9 | **写路径角色状态门槛 = 种子给 421 置 min_verf_level=2（方案 A）**，功能层强制已认证，与 can_publish（level-2）一致 | REVISION-1（coverage MF-1）修复定案 |
| D10 | **division_id = int64，值域 `md_administrative_division.id`（community_div_id 同源）**；community-hub 经 GetResidentialAreasByDivision(community_div_id=division_id, **status=1**) 展开 | REVISION-2（clarity MF-1）+ REVISION-3（clarity SHOULD-1）修复定案 |
| D11 | **新错误码 070004（类型不支持）/070005（大小超限），int32=70004/70005**；70001-70003 不重编号（ErrCodeFileOperationFailed 保持 70003）；file.proto 头注释错误码块对齐实际常量 | REVISION-3（clarity MF-3）修复定案 |
| D12 | 跑马灯数据新增专用 RPC GetMarqueeNotices（返回 ≤10 条 title 摘要，置顶+15天+倒序） | REVISION-4（coverage S-1）修复定案 |
| D13 | 单次发布目标数上限 ≤100（community_ids 长度或 division 展开数），超限 → 080003 超限 | coverage S-2 修复定案 |
| D14 | division_id 仅 community_admin 可用；其他角色传 division_id → 080005 | coverage S-4 修复定案 |
| D15 | GetNoticeRequest.community_id 必填（请求上下文，字段号 2）；缺失/为空 → 080005 参数无效 | clarity SF-4 / INFO I-1 修复定案 |
| D16 | sys_admin 不在移动端可发布角色集（can_publish=false；管理台为 sys_admin 发布面，写路径不额外拦截） | coverage INFO I-5 修复定案 |
| D17 | division→community 授权集解析契约为 design gate（REV-17）；permission-service 判据逻辑是否变更由设计评审定夺，spec 不预先断言「不改」 | clarity SF-1 修复定案 |
| D18 | doc 按 OLE2/CFB（D0CF11E0）+ `WordDocument` 流识别映射 doc，docx 按 ZIP+OOXML `word/document.xml` 识别；其他 OLE2/OOXML 子类型与通用 zip/rar 拒绝 | REVISION-3（clarity MF-2）+ REVISION-5（validity CRITICAL-1）修复定案 |
| D19 | 撤回复用 DeleteNotice（仅发布者本人，全局生效，notice_scope 物理删除） | 设计文档 §4.1/§八（用户确认） |
| D20 | 附件绑定 attachment_ids 仅引用已确认且属发布者的文件 | 设计文档 §2.3/§4.1（用户确认） |
| D21 | 审核复用现有 moderation 流（发布→审核→通过才可见，通知级整体审核） | 设计文档 §七（用户确认） |
| D22 | 组件化 NoticeMarquee / NoticePublisher / NoticeList / NoticeDetail + 附件安全通用校验 | 设计文档 §六（用户确认） |
| D23 | **单通知附件 ≤10 个 且 总大小 ≤50MB（总量上限，CreateNotice 绑定校验时强制）** | Q-ATTACH-TOTAL-CAP 用户拍板 |
| D24 | **附件引用校验载体 = 扩展 FileInfo（新增 file_type + confirmed 字段）+ 复用 GetFileUrl 链路**（community-hub 经 GetFileUrl(file_id) 校验存在/confirmed/归属，file_type 从契约回读） | Q-ATTACH-VERIFY-CARRIER 用户拍板 |
| D25 | **CreateNotice 前端防重（提交中禁用/防双击）+ 后端声明不幂等（不引入幂等键，重复提交产生重复通知）** | Q-CREATE-IDEMPOTENCY 用户拍板 |
| D26 | **剔除 property_admin 移动端发布角色（收敛为 grid_worker/community_admin/committee）**：can_publish=false + 种子回收 421（行为回归：物业原可发布通知 → 只读） | Q-PROPERTY-ADMIN-MOBILE 用户拍板 |
| D27 | **审核通过时设 published_at（经 UpdateNoticeModerationStatus 通过回调），跑马灯/列表窗口从可见日起算**（原实现创建时即设，行为登记） | Q-PUBLISHED-AT-ANCHOR 用户拍板 |
| D28 | **撤回全部保留附件（notice_attachments 行 + MinIO 对象），仅 notices 软删 + notice_scope 物理删** | Q-WITHDRAW-ATTACHMENTS 用户拍板 |
| D29 | **community/v1 头注释错误码块对齐实际语义**（080003=单次发布目标数超限、080005=参数无效含小区ID无效，剔除过时的「080003 寻失发布次数已达上限」注释——该语义实际码为 080007，见 section_quota.go:13） | REVISION-6（clarity MF-1）修复定案 |
| D30 | **published_at 迁移去 NOT NULL**：`ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL`；创建时写 NULL（待审不可见）、审核通过回调（UpdateNoticeModerationStatus pass）置 now（D27）；未迁移则创建 INSERT 违反 NOT NULL → 发布主链路不可用（REQ-NP-1/REQ-NP-MOD-4 异常门禁）；pending 行 published_at=NULL 不参与 idx_published 排序/跑马灯窗口（仅审核通过可见行进入排序） | REVISION-4（REV-1 clarity + REV-3 validity）修复定案 |
| D31 | **080005/080006 错误码消歧（目标级解析失败统一 080006）**：不存在/不可解析的 community_id → 080006（fail-closed，与 scope 未覆盖同处理、与 REQ-NP-4 展开后越权 080006 一致）；080005 仅保留请求形状类错误（空/双载、非 admin 传 division、非法 division_id 值、附件引用/超限） | REVISION-4（REV-2 clarity）修复定案 |
| D32 | **跑马灯 15 天窗口统一以 published_at（审核通过日）为锚**：删除「审核滞留 >15 天」以创建-通过间隔为锚的歧义表述；统一为「published_at 距今（now-published_at）>15 天不入跑马灯但入浏览列表」；正文/标题/验收标准全部以 published_at 为唯一锚（D27） | REVISION-4（REV-4 validity）修复定案 |

> 转换追溯表（决策 → spec Requirement）见 Step 8 交付（StructuredOutput.traceability）。

## 影响范围

| 服务 | 变更类型 | 说明（原行为 → 新行为） |
|------|:---:|------|
| api-proto | 修改 | community/v1：`CreateNoticeRequest` 新增 `community_ids`(8，`[jstype=JS_STRING]`)+`division_id`(9，`[jstype=JS_STRING]`)、`community_id`(1) deprecated；`GetNoticeRequest` 新增 `community_id`(2)；新增 `GetPublishPermission` RPC、`GetMarqueeNotices` RPC；`NoticeAttachment` 加 `file_type`(5)。**file/v1：`FileInfo` 新增 `file_type`(11)+`confirmed`(12)（D24）**；file/v1 头注释错误码块对齐实际常量并登记 070004/070005；**community/v1 头注释错误码块对齐实际语义**（080003=单次发布目标数超限、080005=参数无效含小区ID无效，剔除过时的「080003 寻失发布次数已达上限」注释——该语义实际码为 080007，见 section_quota.go:13，D29）；CHANGELOG 登记全部变更（含 file.proto 头注释语义迁移） |
| community-hub-service | 修改 | **migration：notices.community_id 去 NOT NULL（DEFAULT NULL）+ notices.published_at 去 NOT NULL（DEFAULT NULL，D30）**；新增 notice_scope 表（含 community_id 读索引、双列 NOT NULL）；notice_attachments 加 file_type；写入走 notice_scope、查询走 notice_scope JOIN（原：单列查询 → 新：关联查询）；发布校验 AssertPublishScope（越权 080006）+ **写路径角色状态门槛经 421 min_verf_level=2 由功能层强制**；GetPublishPermission 实现（经 permission GetUserRoles RPC，可发布角色收敛为 grid_worker/community_admin/committee）；GetMarqueeNotices 实现；GetNotice scope 校验 + community_id 必填；division 展开调 masterdata GetResidentialAreasByDivision(community_div_id, status=1)；**CreateNotice 附件绑定校验经 file GetFileUrl 读扩展 FileInfo（confirmed + user_id 归属 + 单通知 ≤10 个/总 ≤50MB 总量上限，D23/D24）**；**published_at 审核通过时设置（经 UpdateNoticeModerationStatus 通过回调置 now，创建时写 NULL，D27/D30）**；**撤回保留附件、DeleteNotice 收窄为仅发布者本人（行为回归登记）** |
| permission-service | 修改 | **种子变更**：grid_worker 授 421（原：无 421，网格员发布被 080002 拦截）；**property_admin 回收 421（D26，原：§3 物业持 421）**；**421 置 min_verf_level=2**（原 §4.2 置 0）；owner/tenant 回收 421（原：种子 §4.8 已授 421）；行为回归登记。AssertPublishScope / GetUserRoles / GetDataScopes 逻辑只读复用；**division→community 授权集解析是否需改判据逻辑由 design gate（REV-17）定夺**（不预先断言「不改判据」） |
| file-service | 修改 | **新增错误码 70004/70005 常量登记**（原 errcode 仅 70001-70003；070004/070005 为新增，70003 保持 ErrCodeFileOperationFailed 不重编号）；**FileInfo 新增 file_type + confirmed 字段（D24）**；GetUploadUrl/ConfirmUpload 增加白名单 + 10MB 两层校验（原：不校验 → 新：声明校验 + magic-bytes 回读校验）；doc OLE2/CFB+WordDocument 流 + docx OOXML word/document.xml 容器识别；通用校验器沉淀 |
| master-data-service | 只读复用 | GetResidentialAreasByDivision（division→小区展开，community_div_id>0 分支 + status=1 仅审核通过）供社区管理员发布展开；ResolveScopeAncestors（祖先链）供 AssertPublishScope 解析；**design gate：division 授权如何落入 community 授权集须经设计评审验证（REV-17）** |
| moderation-service | 只读复用 | 现有发布→审核→通过才可见流不变；UpdateNoticeModerationStatus 回调不变（通过回调新增 published_at 设置语义，见 D27） |
| web/mobile | 新增/修改 | 新增 NoticeMarquee/NoticePublisher 组件；改造 notice/notice-browse/notice-detail/mine 页；范围选择经 permission GetDataScopes（division 选项经 masterdata division 树）；跑马灯消费 GetMarqueeNotices；发布表单提交中禁用（防双击，D25）；附件预校验含单通知总量（≤10 个/总 ≤50MB）；property_admin 无移动端发布入口（D26） |
| web/pc | 不做 | Q6 已拍板：PC 后续单独排期 |

## 风险评估

- **高：数据模型弃用 community_id + published_at 可空化的迁移风险** — 现状 `community_id BIGINT NOT NULL`（migration/001_initial.sql:12）**且 `published_at DATETIME NOT NULL`（001_initial.sql:19，model/notice.go:52-56 插入语句含 published_at、createnoticelogic.go:62 创建时写入）**；若不做去 NOT NULL 迁移，新写入（community_id 置空 + published_at 置空）必违反约束导致发布主链路不可用。缓解：本变更纳入 migration（`ALTER TABLE notices MODIFY community_id BIGINT DEFAULT NULL` + `ALTER TABLE notices MODIFY published_at DATETIME DEFAULT NULL`，D30），**迁移必须先于功能上线**（REQ-NP-1 异常场景门禁）；兼容期保留列不写入不删除；存量数据 Q1 拍板不迁移（旧通知暂不可见，后续补回填挂 BACKLOG）；notices 上 idx_community/idx_published 兼容期保留、标记 deprecated，新读路径走 notice_scope 索引；**pending 行 published_at=NULL 不参与 idx_published 排序/跑马灯窗口——仅审核通过可见行进入排序（REQ-NP-MOD-1 可见性门禁先行）**
- **高：权限种子变更属行为回归** — 给 grid_worker 授 421、**把 421 置 min_verf_level=2、回收 property_admin/owner/tenant 的 421**，直接改变现网功能权限（property_admin/owner/tenant 原可发布通知；未认证用户原可在 level-0 发布；community_admin 原可删辖区内他人通知 → 收窄为仅发布者本人）。缓解：纳入 permission-service 变更范围 + 种子变更 Requirement（REQ-PP-4）+ 验收场景；与 rbac-design.md §6.5 验收矩阵同步更新；写路径角色状态门槛与 can_publish（level-2）判定一致
- **高：AssertPublishScope 权限链路跨 3 服务 + division 展开机制未验证** — 现 resolveUserScope 仅收集 scope_type=community 的 grant，代码库无 community_div scope_type；社区管理员「选社区」展开后 targets 为具体小区（community），其授权能否覆盖取决于 division grant 的落位方式。缓解：**design gate——division→community 授权集解析契约须在设计评审阶段以权限服务单测/集成验收验证，未验证不得进入编码**（D17/REQ-NP-4 验收场景 + scope.go CodePublishScopeDenied 复用 080006）；发布时展开为具体小区 id 快照写入，targets 与现有 AssertPublishScope 输入一致；spec 服务职责边界如实表述「判据逻辑是否变更待设计评审定夺」
- **中：附件校验绕过风险（客户端直传模式）** — 缓解：Q5 拍板两层结合 — GetUploadUrl 按声明类型/大小快速拒绝 + ConfirmUpload 回读 MinIO 实际对象按 **magic-bytes 内容嗅探**判定真实类型（非扩展名/元数据），扩展名与魔数一致才放行；doc 按 OLE2/CFB + **`WordDocument` 流**、docx 按 ZIP+OOXML `word/document.xml` 内容签名显式放行，其他 OLE2/OOXML 子类型（msi/xls/ppt/xlsx/pptx）一律 070004 拒绝（D18，REVISION-5 封堵 .msi 改 .doc 绕过）
- **中：错误码 070004/070005 与既有 70001-70003 的冲突处置** — 070004/070005（int32 70004/70005）为全新整数位，与既有 70001-70003 无冲突；file.proto 头注释原「070002 上传失败/070003 文件类型不支持」与实际常量（70002 访问被拒绝/70003 操作失败）漂移，本变更将对齐为实际常量并登记 070004/070005（D11）；**community/v1 头注释同类漂移一并处置（D29：080003=目标数超限、080005=参数无效，剔除陈旧「080003 寻失发布次数已达上限」→ 实际码 080007）**。缓解：spec 显式说明 70003 不重编号
- **中：跨服务字段一致性回归** — 缓解：Proto 生成 + TS 类型同步；沿用 `[[snake-camel-field-mismatch]]` 记忆，QA 自动检查 snake_case/camelCase；新 RPC/字段走兼容路径（新增字段号）；community_ids 显式 `[jstype=JS_STRING]`（防 TS 精度丢失，硬性约束 #3）
- **中：与进行中变更服务重叠** — 冲突预检（check-change-conflict.sh）检出 6 处 C1/C2 重叠（permission-service/web/mobile/api-proto/master-data-service 与 rel-user-role-migration-publish-fix / spec-pipeline-e2e-l / test-pipeline-work-records；其中与 rel-user-role-migration-publish-fix 在 init_permissions.sql 存在 C2 文件级重叠）。缓解：本变更聚焦 notice 域，proto 均走新增字段/RPC 兼容路径，种子变更仅动 421 关联 + min_verf_level + property_admin 421 回收；合并前各变更走 `make ci` breaking-check 交叉验证 + 与 rel-user-role-migration-publish-fix 的种子变更对齐

## 不做清单（Won't have — 本轮明确不实现）

- **不处理存量数据**（Q1）：旧通知在迁移回填前不可见；不做 community_id→notice_scope 回填脚本（回填任务挂 BACKLOG，建议优先级 P1）
- **不做 PC 端**（Q6）：web/pc 后续单独排期（property_admin 的发布面随 PC 一并顺延，本变更移动端 property_admin 不可发布，D26）
- **不做小区/村层级统筹**（task-016）：community_type 不影响范围表达，范围单位统一 md_residential_area.id
- **不做附件病毒扫描 / 图片处理 / CDN**：file-service 既有边界（1.2 不做什么）；本变更的 magic-bytes 嗅探仅用于类型白名单判定，不替代病毒扫描
- **不做下载侧 file_url 的 scope 绑定**：上传白名单不约束已披露 URL 的再传播；下载沿用 file-service 既有预签名有效期
- **不做批量操作**：通知撤回一次一条（DeleteNotice），不做批量撤回/批量审核
- **不做缓存层**：跑马灯/浏览直查 MySQL（沿用现有缓存策略「当前不使用 Redis」）；读路径靠 notice_scope 索引兜底
- **不改 ListNotices 响应结构**（Q2）：响应保持单 community_id，不做 repeated community_ids 响应
- **不新增置顶 UI / 不改置顶写入路径**：is_pinned 继续走既有 UpdateNotice（编辑后 moderation_status 重置为 0 的既有语义保持），本变更不覆盖置顶入口
- **不改 UpdateNotice 的 scope 编辑能力**：UpdateNotice 保持通知级（title/content/is_pinned），scope 编辑不在本变更范围
- **不回收 owner/tenant 的 435/436、不回收 property_admin 的其他权限**：435（寻失发布）/436（联络维护）属其他模块，非本变更范围（本变更仅回收 421 通知发布）
- **不引入新审核链路**：复用现有 moderation 流；不引入幂等键（后端不幂等，防重由前端承担，D25）
- **不在本变更内实现 docx 之外的容器型文件白名单扩展**：白名单仅 png/jpg/jpeg/gif/pdf/doc/docx
- **sys_admin 不在移动端做发布入口**：sys_admin 的发布面为管理台（PC），移动端 can_publish=false（D16）；写路径不额外拦截 sys_admin

## 验收标准

- [ ] migration 后：notices.community_id 可空 **且 notices.published_at 可空（DEFAULT NULL，D30）**；网格员可一次发布到多个小区，notices 写入 community_id=NULL + published_at=NULL（待审），notice_scope 写入 N 条关联（含 community_id 读索引、双列 NOT NULL），旧 idx_community/idx_published 兼容期保留
- [ ] 未迁移时（community_id 或 published_at 仍 NOT NULL）：多小区发布 INSERT 被 NOT NULL 拒绝，发布失败（迁移先于上线门禁，REQ-NP-1 异常门禁）
- [ ] 社区管理员选 division（值域 md_administrative_division.id）→ 后端经 GetResidentialAreasByDivision(community_div_id, status=1) 展开为具体小区 → AssertPublishScope 全部通过 → 快照写入；展开后无小区 → 080005；展开结果仅含审核通过小区
- [ ] division_id 仅 community_admin 可用，其他角色传 division_id → 080005
- [ ] 单次发布目标数 >100（community_ids 长度或 division 展开数）→ 080003 超限
- [ ] grid_worker 持 421 功能权限放行（min_verf_level=2），发布越权 → **080006**（非 080002）；无发布角色 → 080002；**未认证/已过期 grid_worker 持数据范围直接 CreateNotice → 080002（写路径角色状态门槛，与 can_publish=false 一致）**
- [ ] **目标级解析失败（community_ids=[99999] 目标小区不存在/不可解析）→ 080006（与 scope 未覆盖同处理，D31）；080005 仅保留请求形状类（空/双载、非 admin 传 division、非法 division_id 值、附件引用/超限）**
- [ ] **property_admin 回收 421 后：can_publish=false、移动端无发布入口、直接 CreateNotice → 080002（D26）**；owner/tenant 回收 421 后：can_publish=false、入口隐藏、直接 CreateNotice → 080002
- [ ] GetPublishPermission 返回 can_publish + 可发布角色（level-2 判定，可发布角色集 = grid_worker/community_admin/committee）；【我的】页据此显隐发布入口；sys_admin can_publish=false
- [ ] GetNotice 带 community_id（必填，字段号 2）上下文：缺失 → 080005；scope 外/不存在/审核未过 → 080001；响应 community_id = 请求小区
- [ ] **GetMarqueeNotices(community_id) 缺失/空 → 080005；返回最近 15 天通知标题（published_at >= now-15*24h 含边界，置顶优先 + 倒序 + 封顶 10 条，仅审核通过），更多→浏览页**
- [ ] **published_at 在审核通过时设置（D27/D30，创建时写 NULL）：跑马灯 15 天窗口/浏览列表倒序/详情发布时间均以 published_at（审核通过日）为唯一锚——published_at 距今（now-published_at）>15 天的通知不入跑马灯但入浏览列表；审核滞留（创建-通过间隔）不作为窗口锚（D32）**；浏览页按发布时间倒序；详情页展示标题/正文/时间/附件（含 file_type）
- [ ] 附件白名单外拒绝（070004）、单文件 >10MB 拒绝（070005）、zip/rar 拒绝、exe/sh 拒绝；改名 png 上传 exe 在 ConfirmUpload magic-bytes 回读时拦截（070004）；docx 按 OOXML word/document.xml 签名放行、doc 按 OLE2/CFB+WordDocument 流签名放行；**msi/xls/ppt 改 .doc、xlsx/pptx 改 .docx 均 070004 拒绝**
- [ ] **单通知附件 ≤10 个 且 总大小 ≤50MB：CreateNotice 绑定校验超限 → 080005（D23）**；附件引用校验经 file GetFileUrl 读扩展 FileInfo（confirmed=true + user_id 归属）→ 未确认/他人文件 080005（D24）
- [ ] **撤回全局生效（多小区一次撤，仅发布者本人，notice_scope 物理删除；notice_attachments 行与 MinIO 对象全部保留；非发布者 → 080002）（D28）**
- [ ] **发布表单提交中禁用（防双击），后端不幂等：双击/重试产生重复通知（D25）**
- [ ] 通知经 moderation 审核通过才可见（多小区一致）
