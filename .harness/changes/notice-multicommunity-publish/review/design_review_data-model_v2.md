# Design Review — notice-multicommunity-publish（data-model + interface-proto 视角，本轮复审）

**审查模式**: 模式一.5 设计评审 · **视角**: data-model（数据模型/服务归属）+ interface-proto（接口契约/Proto 破坏性/鉴权-幂等-错误码/依赖顺序）
**审查对象**: `.harness/changes/notice-multicommunity-publish/design.md` + `tasks.md`（07:32/07:33 修订版，已纳入 data-model v1-v4 与 interface-proto v1-v4 全部修订）
**对照基准**: v1-v4 评审问题清单 + 6 个 spec + 磁盘真相（community-hub migration 001/002、model/notice.go、types.go、web/mobile community.ts、community.proto/file.proto/permission.proto、init_permissions.sql、scope.go resolveUserScope、AssertPublishScope/GetDataScopes/GetUserRoles 契约）
**审查时间**: 2026-08-16

## 摘要
- 🔴 MUST FIX: 1 / 🟡 SHOULD FIX: 1 / 🔵 INFO: 2
- **VERDICT: REVISION**（1 个 MUST FIX：移动端发布主路径未接线 `publisher`（展示名）发送，设计自定的非空校验 080005 会拒绝全部移动端发布）

---

## 一、上一轮（data-model v4 + interface-proto v4）MUST FIX / SHOULD 修复核验（磁盘证据确认）

| 项 | 问题 | 落地证据 | 状态 |
|---|------|---------|:---:|
| data-model v1 M1 | 读门禁与回调 pass 集合漂移 | §可见性门禁 `IsModerationPassed`（IN 1,3）单一谓词 + Task 1.4/1.5/1.13 共用 + Task 5.2 status=3 冒烟；community.proto:316 语义 1/2/3/4 一致 | ✅ |
| data-model v3 M1 | S4 附件重生缺 file_id 载体 | migration 003 增 `notice_attachments.file_id BIGINT DEFAULT 0`；Task 1.3 `FileId` 模型+Insert；Task 1.7 落库写 file_id；Task 1.8 GetNotice 按 file_id 重生 + file_id=0 回退 stored file_url；Task 0.4 CHANGELOG 登记 | ✅ |
| data-model v4 S1-S4 | toProtoNotice 适配 / 撤回单事务 / JOIN 投影 / file_url 加宽 | Task 1.3 补 toProtoNotice（`*int64`+`sql.NullTime`）；Task 1.12 钉死 `conn.Transact` 共享 session + 失败注入测试；Task 1.4 显式投影 `notices.*`+`notice_scope.community_id`；Task 1.1 `file_url VARCHAR(1024)` + Task 1.7 新行写占位空串 | ✅ |
| interface v4 MUST 1 | 读路径权限矩阵（详情 426 缺失 + 422 排除发布者） | design §权限种子新增 **426 GET:/api/community/notices/:id** + **422 扩展全部移动端角色** {1,3,4,5,6,7,8,9}；Task 3.2 种子 + Task 5.1/5.2 断言 | ✅ |
| interface v4 MUST 2 | 写路径权限矩阵（DELETE/PUT 无码） | design §权限种子新增 **427 DELETE** + **428 PUT** 绑定全部移动端角色；真正越权判定交 080002 作者校验 | ✅ |
| interface v4 SHOULD 3 | parent_id 未指定 | 423/424/426→410、427/428→420、425→310（磁盘确认 310 permission:read / 410 community:read / 420 community:notice 均存在） | ✅ |
| interface v4 SHOULD 4 / INFO 5/6 | division_id 空串 / 幻影 435 / rbac-design §6.5 | Task 4.1/4.6 省略字段不传空串；Task 3.2 幻影措辞；Task 3.2 rbac-design §6.5 补登 | ✅ |

**磁盘真相复核（与设计一致，非问题）**：
- Proto 字段号全部空闲且兼容新增：CreateNoticeRequest 现状 1-7 → community_ids=8/division_id=9；GetNoticeRequest id=1 → community_id=2；NoticeAttachment 1-4 → file_type=5；FileInfo 1-10 → file_type=11/confirmed=12。全 int64 带 `[jstype=JS_STRING]`，`buf breaking-check` 可过。
- 迁移前提准确：001 的 `community_id BIGINT NOT NULL` / `published_at DATETIME NOT NULL` / `notice_attachments.file_url VARCHAR(500) NOT NULL` 均核实；003 去 NOT NULL + 加列 + 加宽方向正确。
- 权限种子现状：421 绑 (2,3,6,1,5)、grid_worker(4) 未持、min_verf_level 现 0（line 201-202 UPDATE）；422 绑 (9,1,5)；423/424/425/426/427/428 无行 —— 设计「授 4 / 回收 2/1/5 / 置 2 / 422 扩展 / 新增 423-428」与现状一致。
- Design Gate 证据：`resolveUserScope` `g.ScopeType != scopeType → continue`（scope.go:46）只收集 community，community_div 授权不落入授权集 —— 判据须变更结论成立；blast radius（lostfound/contacts 共享 RPC）已显式声明 + Task 3.1 回归门禁。
- GetDataScopes（scope_ids+state）、AssertPublishScope（repeated targets 批量）、GetUserRoles（status/verified_at/expires_at）契约均存在，level-2 判定可基于 RPC 输出。
- 依赖顺序 Proto(0)→迁移/模型(1.1-1.5)→scope 基础设施(1.6)→写/读逻辑(1.7-1.13)→注册/API(1.14-1.17)→前端(4)→运维验证(5)，合规。

---

## 二、本轮新发现

### 🔴 MUST FIX（1）

| # | 文件:章节 | 问题 | 修复建议 |
|---|---------|------|---------|
| 1 | tasks.md Task 4.1 / Task 4.6 + design §CreateNotice（校验顺序 step 1） | **移动端发布主路径未接线 `publisher`（展示名）发送 → 全部移动端发布被设计自定的 080005 拒绝**。design §CreateNotice 明确：`notices.publisher` 为 NOT NULL 列，来源 = **请求体展示字符串**，且校验 step 1「`publisher`（展示名）非空 → 否则 080005」；`publisher`(5) proto 字段不 deprecated（评审 S1 定案）。但移动端任务链未定义发送：Task 4.1 `createNotice` client 参数仅 `community_ids`/`division_id`/`attachment_ids`（无 publisher）；Task 4.6 NoticePublisher 表单仅收集「标题/正文/附件/范围选择」+「空标题/正文前端校验」（无 publisher）。磁盘确认：types.go `CreateNoticeReq.Publisher string json:"publisher"` 为既有必填展示列；web/mobile `community.ts` 无 createNotice 函数（本变更新增，接口-proto v2 INFO 3 已核实），Notice 类型 publisher 仅读。**按 tasks.md 字面实现，移动端每次发布 publisher 恒空 → 校验 step 1 080005 → REQ-NM-5/REQ-NP-3 发布主交付不可用**（服务端 Task 1.7 的「publisher 空 080005」测试虽绿，但客户端从不发送）。 | 补前端接线：Task 4.1 `createNotice` client 增加 `publisher: string` 参数（来源 = 当前用户展示名，如 userStore/profile 显示名）；Task 4.6 发布表单提交时携带该展示名（展示名非安全字段，但必须显式发送，禁止悬空）。同步补断言：Task 1.15 api 层测试增加「`publisher` 缺省 → 080005」；Task 4.1/4.6 组件/单测断言 createNotice 请求体含非空 `publisher`。若意图改为服务端派生（经 user-service 查展示名），则须在 design §CreateNotice 显式声明来源 RPC 并纳入任务。 |

### 🟡 SHOULD FIX（1）

| # | 文件:章节 | 问题 | 建议 |
|---|---------|------|------|
| 1 | design §CreateNotice「role 派生」+ spec REQ-PP-1 sys_admin 场景 + tasks.md Task 1.7 | **sys_admin 直连 CreateNotice 的角色派生缺口 → `notices.role` NOT NULL INSERT 失败（500）**。磁盘确认 sys_admin(8) 持 421（seed line 178-180 `SELECT 8, id FROM sys_permission WHERE status=1` 全权限绑定），且 421 经本变更置 min_verf_level=2。spec REQ-PP-1 sys_admin 场景显式声明「sys_admin 写路径不额外拦截（满足 421+level-2 的直连 CreateNotice 通过）」。但 CreateNotice role 派生（Task 1.7）映射仅 grid_worker/community_admin/committee → NoticeRole，sys_admin 无映射；design 自身注记「`notices.role` 为 NOT NULL 列，落空将 INSERT 失败」，并断言「无发布角色由功能层覆盖（RPC 层 mock 不设）」——该断言对持 421 的 sys_admin 不成立。结果：level-2 已认证且持小区数据范围的 sys_admin 直连移动端 CreateNotice 会走到角色派生 → 无映射 → role 落空 → 500。 | 定案 sys_admin 写路径行为并落地：方案 A（与 D26「收敛发布角色」一致）——在 CreateNotice role 派生显式拒绝非发布角色（sys_admin/owner/tenant 等）→ 080002，并同步修订 spec REQ-PP-1「写路径不额外拦截」措辞；方案 B——为 sys_admin 映射一个展示角色（如 NOTICE_ROLE_COMMUNITY）。二者选一，并给 Task 1.7 补 sys_admin 测试用例（RED 断言 080002 或落库成功），消除 500 路径。 |

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | permission.proto `ScopeRef` 头注释「本变更仅 community」在 Design Gate 落地后易误导：grant 侧 resolvePublishScope 收集 `community` ∪ `community_div` 双 scope_type（target 侧仍仅 community）。建议更新注释或 design 加一句说明，避免执行方误以为 scope 校验只认 community grant。 |
| 2 | GetNotice 附件重生为逐附件 N 次 GetFileUrl RPC（≤10）；若某附件文件被删（GetFileUrl not-found），重生失败且新行 file_url=''（占位空串）→ 附件展示空 URL。建议 GetNotice 重生失败时记 Infof 降级日志，并确认前端对空 file_url 附件渲染降级不崩溃。 |
| 3 | Task 3.1 需保证 `resolveUserScope` 读路径（GetDataScopes）保持 `community` 单 scope_type 不被误改（design 已声明，执行时注意回归）。 |

---

## 三、核验清单（data-model + interface-proto 两维度）

- [x] 数据模型：notice_scope（复合 PK=唯一约束 + idx_scope_community 左 community_id、双列 NOT NULL、纯关联表物理删除）✅；community_id/published_at 去 NOT NULL 与磁盘一致 ✅；published_at 审核锚定 + sql.NullTime ✅；file_type/file_id/file_url 加宽 ✅；时间字段/软删除符合硬约束 #3.1 ✅
- [x] 接口契约：Proto 字段号/新 RPC 全兼容、JS_STRING ✅；GetFileUrl+FileInfo 附件载体一次 RPC 双满足 ✅；批量 AssertPublishScope 非 N+1 ✅；level-2 判定经 GetUserRoles ✅；错误码 D11/D29/D31 消歧 ✅；必填 community_id 语义破坏已登记 ✅；幂等（不幂等 + 前端防重 / 回调幂等）✅
- [x] 鉴权矩阵：读 422/423/424/425/426 + 写 421/427/428 全部绑定 + parent_id 正确，fail-closed 语义完整 ✅
- [x] 任务粒度：单任务不跨服务、不混合三类、Proto/Migration 独立成任务、测试 1~10 ✅
- [x] 依赖顺序：基础设施→核心→辅助→前端，合规 ✅

## 报告自检
1. MUST FIX 定位到具体任务（Task 4.1/4.6）+ design §CreateNotice step 1 + 磁盘证据（types.go/community.ts）✅
2. 每条附可落地修复建议（client 参数 + 表单携带 + 断言 / sys_admin 双方案）✅
3. data-model + interface-proto 两维度逐项覆盖 ✅

## 问题跟踪表
| 状态 | 说明 |
|------|------|
| 已修复 | data-model v1 M1 / v3 M1 / v4 S1-S4、interface v4 MUST 1/2 + SHOULD 3/4 + INFO 5/6（本轮磁盘复核 ✅） |
| 待修复（本轮 MUST 1） | 移动端 `publisher` 未接线 → 发布主路径 080005，阻塞级 |
| 待修复（本轮 SHOULD 1） | sys_admin role 派生缺口 → 潜在 500（不阻塞，可随 MUST 一并修订） |

---
VERDICT: REVISION
---
