# Design: 访问控制与数据权限改造（access-control）

> 阶段：架构设计（阶段3）。输入为 proposal.md + 5 个 spec（已通过 2/3 需求评审）+ `.change.yaml`。
> 设计依据（权威）：`docs/specs/access-control-design.md`。
> 前置变更：`access-data-permission` 已交付角色分层（需求1）与数据权限统一模型（需求3，`AssertPublishScope`/`GetDataScopes`/scope 三态）。

---

## 0. 阶段3 开放项定稿（6 项，逐一定稿）

### STAGE3-1 认证粒度：per-community（目标小区认证）✅ 定稿

「认证用户不受每年次数限制」的认证粒度定稿为 **per-community（目标小区认证）**：仅当用户在**目标小区**持有已认证（`verf_status=approved`，status=2）的 `owner`/`tenant` 角色时，才视为该目标小区的「认证用户」，豁免该小区每年的新加入次数限制。

理由（用户拍板）：
1. **认证是 per-community 的**——一个小区认证只代表「是这个小区的业主」，不能代表「是其他小区的业主」；所有信息的组织都以小区为单位。
2. 因此「认证用户不受每年限制」按**目标小区**判定：A 小区认证的业主，加入 B 小区时对 B 仍是「未认证用户」，受 B 的每年新加入次数限制。
3. 现状代码 `isVerifiedOwnerOrTenant` 为**全局**粒度（遍历任一 status==approved 的 owner/tenant），**需改为 per-community 判定**（校验目标小区的认证状态），由 Task 3.5 承接，不是补测试锁定。

### STAGE3-2 计数公式矛盾：design §7 公式正确，需澄清状态机 ✅ 定稿

澄清：`status`（业务态）与 `moderation_status`（审核态）**正交**。现状代码 `createlostfoundlogic.go` 在创建内容时即设 `status="active"`（业务上「在版」），待审期间 `moderation_status=0` 但 `status` 仍是 `active`。因此 design §7 计数谓词 `deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)` **成立且无矛盾**：待审（0）与展示（1）同占配额，驳回（2）/已解决（status=resolved）/已删除（deleted_at 非空）释放。

「待审内容占配额」之所以成立，**前提是创建即 `status='active'`**——这是实现不变量，须在 tasks 中以 TDD 锁定，并在 `docs/specs/access-control-design.md` §7 补记该状态机说明（任务 T0.6）。现网 `status` 仅 `active`/`resolved` 两态，未来「下架」复用「status 非 active」语义。

### STAGE3-3 错误码命名空间对齐 ✅ 定稿（最终分配表）

**规范化结论**：全仓运行时错误码一律为 **5 位 `XXYYY`**（XX=服务中心 2 位，YYY=具体码 3 位），`NewBaseRespWithError(code,msg)` 的实参均为 5 位。设计稿/CLAUDE 中的 `050006`、`080xxx` 是文档层把「08」写作 2 位中心码的**误写**，须归一到 5 位。服务中心码（运行时实测）：`10`=User、`50`=Auth、`60`=Permission、`80`=Community-hub、`40`=Moderation、`99`=Common（`工程结构.md` 表中 `06/07` 为 permission/file 的 2 位写法的残留，运行时实为 `60/70`）。

| 语义 | 服务 | 最终码 | 状态 | 说明 |
|------|------|:---:|:---:|------|
| 该账号为移动端用户，请使用移动端 APP（端限制拒绝） | auth | **50007** | 新增 | `50006` 已被「注销失败/Token 已拉黑」占用；设计稿「050006」为 6 位误写，弃用 |
| 超出发布配额（板块配额超限） | community-hub | **80007** | 新增 | `80001-80006` 已用，`80007` 空闲；CLAUDE「08xxxx」写法归一为 `80xxx` |
| 该房屋已满员（每户 ≤6） | user | **10014** | 新增 | `10013` 后首个空闲码 |
| 目标小区不在数据范围（当前小区切换越界） | user | **10015** | 新增 | 与 `10010`（CheckAccess 权限不足）语义区分 |
| 同时加入小区数超限（≤3） | user | 10006 | 复用（已存在） | 现状代码 `join_community_logic.go` 已实现；`user-service/docs/design.md` 中「最多加入 5 个」为陈旧值，实际默认 3（`user.max_community_join_count` 可配），一并修正 |
| 每年新加入小区数超限（≤3，仅非认证） | user | 10012 | 复用（已存在） | spec 未列码，代码已实现 |
| 终身加入小区总数超限（≤12） | user | 10013 | 复用（已存在） | spec 未列码，代码已实现 |

> 说明：member-constraint 的「同时/每年/终身」三项约束现状代码**已实现**（`10006/10012/10013`），本次不新增码，仅补齐 spec 引用与「终身限制须对全部用户生效」的对齐（见 T3.5）。

### STAGE3-4 fail-open 与 design §4.1 默认映射 ✅ 定稿

两者是**不同层面的「默认」**，不矛盾，定稿如下：
- **运行时回退（fail-open）**：auth-service 端判定时，某角色 `platforms` 为空/NULL → 该角色视为**允许所有端**，不拦截。这是 spec `REQ-PLAT-1` 的行为契约，也是「UX 引导而非安全边界」的体现（缺配置不得把用户锁在门外）。
- **种子初值（design §4.1）**：permission-service `init_permissions.sql` 为 8 个内置角色写入**显式** `platforms` 初值（见 §2 数据模型），属 `sys_admin` 可改的配置初值，不是运行时回退。

落地：内置角色由种子保证显式配置（故 fail-open 仅作用于 sys_admin 新建且未配 platforms 的自定义角色）；design §4.1 的「默认映射」定位为**种子初值**，并在 `access-control-design.md` §4.1 补注二者语义差异（任务 T0.6）。

### STAGE3-5 proto_changes #5「或」 ✅ 定稿

定稿为 **新增独立 `sys_section_quota` 表 + 新增 `GetSectionQuota` RPC**（去掉「并入 sys_config」）。理由：design §7 已明确 `sys_section_quota(section_type, max_count)` 为**独立配置表**（非 `md_configuration` 键值）；配额需按板块结构化查询；`md_configuration` 仅适合标量 `sys_config` 键。详见 §2/§3。

### STAGE3-6 CLAR-4「增字段/增列」一致性 ✅ 定稿

现状核实：`building/unit/room` **已存在**，CLAR-4 的「增字段/增列」表述失真——proto（`api/user/v1/user.proto` 的 `JoinCommunityRequest`/`CommunityMembership`）已于 commit `1841fd2` 落地，model（`user_community_membership.go`）已含 `Building/Unit/Room`，migration `003_add_address_fields.sql` 已 `ADD COLUMN`。**无 proto/schema 变更**。

剩余工作为**逻辑层**（落点定稿）：
1. **必填校验**：`JoinCommunity` 逻辑层强制 `building/unit/room` 三字段必填（现状 `if in.Building > 0 && in.Room > 0` 跳过校验），API 层 `types.go` 移除 `,optional`（引 `[[api-required-field-marked-optional]]`）。
2. **房屋计数 ≤6 替换地址唯一性**：现状 `FindByAddress`+`10011「该地址已有人加入」` 是一户一人；本次改为 `CountActiveByAddress`（`bind_status=active`，排除当前用户）≤ `user.max_house_members`（默认 6）→ `10014`。

`user_residence`（认证后创建，`BindResidence`）与 membership 的 `building/unit/room`（加入即采集、自报）是**两套正交数据**：同屋互见/房屋上限只依据 membership 的 join-time 采集值，不改 `user_residence`。

---

## 1. 服务归属决策

| 功能 | 归属服务 | 理由 |
|------|---------|------|
| `sys_role.platforms` 存储/透出 + 权限码注册 | permission-service | 角色/权限码数据所有权在 permission-service |
| 端准入判定（登录/刷新） | auth-service | 登录/签发 Token 是 auth-service 领域；读 permission `GetUserRoles` 取 platforms |
| `user_app_state` 当前小区持久化 + 切换校验 | user-service | 账号级应用状态数据所有权在 user-service；切换校验消费 permission `GetDataScopes` |
| 成员约束（同时/每年/终身/每户 ≤6） | user-service | 成员域独立规则，JoinCommunity 权威执行（`design §2` 附注） |
| 同屋互见判定（手机号解密/脱敏 + 楼栋房屋号） | user-service | 用户详情接口的户级可见性规则，membership 数据在 user-service |
| `sys_section_quota` 配置 + `GetSectionQuota` RPC | master-data-service | 配置数据所有权在 master-data |
| 板块发布配额校验（计数 ≤ 上限） | community-hub-service | 发布写路径在 community-hub，消费 master-data 配额配置 |
| 前端（引导/切换/表单/互见展示/platforms 维护） | web/pc + web/mobile | 见 §7 |

新增服务依赖：
```
auth-service ──gRPC──▶ permission-service   （新增：GetUserRoles 取 platforms 做端判定）
user-service ──gRPC──▶ permission-service   （已有：AssignRole/GetUserRoles/GetDataScopes）
community-hub ──gRPC──▶ master-data-service （新增：GetSectionQuota 读配额；ResolveScopeAncestors 已有）
```

---

## 2. 数据模型

### 2.1 变更：`sys_role` 增 `platforms`（permission-service）

```sql
ALTER TABLE sys_role
  ADD COLUMN platforms VARCHAR(32) NOT NULL DEFAULT '' COMMENT '允许登录的端，逗号分隔：pc,mobile；空=未声明（运行时 fail-open 允许所有端）';
```
- 值域：`''`（未声明，fail-open）、`pc`、`mobile`、`pc,mobile`。
- 种子（`init_permissions.sql`，`sys_admin` 可改）：`sys_admin=[pc]`、`community_admin=[pc,mobile]`、`property_admin=[pc]`、`owner/tenant/grid_worker/committee/merchant/registered_user=[mobile]`（`registered_user` 角色 id=9，已由前置变更创建）。

### 2.2 新增：`user_app_state`（user-service，账号级当前小区）

```sql
CREATE TABLE user_app_state (
    user_id BIGINT PRIMARY KEY COMMENT '账号（雪花 ID，同 user_base.id）',
    current_community_id BIGINT NOT NULL DEFAULT 0 COMMENT '当前小区；0=未设置',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) COMMENT '用户应用状态：当前小区（账号级、跨设备一致）';
```
- 索引：主键即 user_id，无额外索引。
- 取代 `user_base.preferences.default_community_id`：`preferences` 列**保留**（不破坏性迁移），但 JoinCommunity 不再向其中写 `default_community_id`，改为 upsert `user_app_state`（任务 T3.3/T3.4）。

### 2.3 新增：`sys_section_quota`（master-data，板块发布配额）

```sql
CREATE TABLE sys_section_quota (
    id BIGINT PRIMARY KEY COMMENT '雪花 ID',
    section_type VARCHAR(64) NOT NULL COMMENT '板块类型：lost_found/notice/contact/second_hand',
    max_count INT NOT NULL COMMENT '占配额上限',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_section_type (section_type)
) COMMENT '板块发布配额（配置了限制；未配置=不限）';
```
- 种子：`INSERT ... (section_type='lost_found', max_count=5)`。`notice`/`contact`/`second_hand` 不种子 → 未配置=不限（CLAR-3）。

---

## 3. 接口设计

### 3.1 permission-service（Proto `Role.platforms` 透出）
- **Role**（`permission.proto`）增 `repeated string platforms = 10`。
- 透出链路：`ListRoles`/`GetRole`/`GetRolesByIds`/`GetUserRoles`（`UserRoleInfo.role.platforms`）全部经由 `toProtoRole` 填充。
- 缓存：角色缓存失效已随既有 `UpdateRole` 流程；`platforms` 属角色字段，无额外缓存键。

### 3.2 auth-service（端准入判定）
- 新增内部 helper `checkPlatformAccess(ctx, svcCtx, userID int64, deviceType string) error`：
  1. 端归类：`web`/`admin`→`pc`；`ios`/`android`/`miniapp`→`mobile`；**空/未知 → 放行（fail-open）**。
  2. `PermissionClient.GetUserRoles(userId)` → 遍历 `UserRoleInfo`。
  3. 任一角色 `role.platforms` 为空 → 放行（fail-open）；任一角色 platforms 含当前端 → 放行。
  4. 全部不满足 → 返回 `50007`（该账号为移动端用户，请使用移动端 APP）。
- 挂载点：`Login`/`LoginSms`/`Register`/`RefreshToken` 的**签发 Token 前**。
- 新依赖：auth-service 增 `PermissionServiceRpc` gRPC 客户端（`config.go` + `servicecontext.go` + `etc/authservice.yaml`）。

### 3.3 user-service
- **GetAppState(GetAppStateRequest{user_id}) → GetAppStateResponse{current_community_id, updated_at}**：读 `user_app_state`，无记录返回 `current_community_id=0`。
- **SetCurrentCommunity(SetCurrentCommunityRequest{user_id, community_id}) → SetCurrentCommunityResponse{}**：
  1. 调 `PermissionClient.GetDataScopes(user_id, "community")`。
  2. `state==GLOBAL` → 放行；`state==EMPTY` → 拒绝 `10015`；`LIMITED` → `community_id ∈ scope_ids` 才放行，否则 `10015`。
  3. upsert `user_app_state`。
- **JoinCommunity 房屋约束**（改造）：三字段必填（缺任一 → `10040`）；`CountActiveByAddress(community_id,building,unit,room)` 排除当前用户、仅 active，`>= user.max_house_members`（默认 6）→ `10014`。移除 `FindByAddress` 唯一性校验（`10011` 弃用）。
- **JoinCommunity 终身限制对齐**：终身 `10013` 校验移出 `!isVerifiedOwnerOrTenant` 块（对全部用户生效）；每年 `10012` 保留仅非认证（STAGE3-1 per-community 粒度，认证按目标小区判定）。
- **GetUser 同屋互见**（改造）：
  - `GetUserRequest` 增 `optional int64 viewer_id = 2`；`GetUserResponse` 增 `SameHouseInfo same_house = 3`。
  - `SameHouseInfo{ bool same_house; int32 building; int32 unit; int32 room; }`。
  - 判定：`viewer_id` 为空 → 手机号脱敏（默认安全）；`viewer_id==target` → 明文+自身房屋号；否则查双方 active membership 是否同 `community+building+unit+room` → 同屋明文+房屋号，非同屋脱敏、不返回房屋号。
- **API 层**（`services/user-service/api`）：新增 `GET /api/users/me/app-state`、`PUT /api/users/me/current-community`；`types.go` 移除 `JoinCommunityReq` 楼/单元/房号的 `,optional`。

### 3.4 community-hub-service
- 新增 `sectionquota` helper：`CheckSectionQuota(ctx, userID, communityID, sectionType) error`：
  1. `MasterDataClient.GetSectionQuota(sectionType)` → `configured=false` → 跳过（不限）。
  2. `CountQuotaOccupied(userID, communityID, sectionType)`（谓词 `deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)`）。
  3. `count >= max_count` → `80007`。
- 挂载点：`CreateLostFound` 在 `AssertPublishScope` 之后、落库之前调用（notices/contacts 默认未配置，调用同 helper 自动跳过，结构上预留）。

### 3.5 master-data-service
- **GetSectionQuota(GetSectionQuotaReq{section_type}) → GetSectionQuotaResp{max_count, configured}**：查 `sys_section_quota`，未配置返回 `configured=false,max_count=0`。

---

## 4. 业务流程

### 4.1 端限制（登录准入）
```
Login/LoginSms/Register/RefreshToken
  → 校验凭据/RT 合法
  → checkPlatformAccess(user_id, device_type)   // 读 permission GetUserRoles.platforms
       ├─ 任一角色空 platforms / 含当前端 → 放行
       └─ 否则 → 50007（引导移动端）
  → 拉取角色（JWT）→ 签发 Token
```

### 4.2 当前小区
```
mobile 切换小区 → PUT /api/users/me/current-community
  → user-service SetCurrentCommunity
      → permission GetDataScopes(user_id, "community")
          ├─ GLOBAL → 放行
          ├─ EMPTY  → 10015
          └─ LIMITED → community_id ∈ scope_ids ? 放行 : 10015
      → upsert user_app_state
```

### 4.3 板块发布配额
```
发布（lost_found 等写接口）
  → PermMiddleware（功能权限）
  → AssertPublishScope（数据权限，permission）
  → CheckSectionQuota（master-data 读配额 → 计数 → 80007 超限）
  → 落库 → 审核
```

### 4.4 成员约束 + 同屋互见
```
JoinCommunity：必填校验(10040) → 同时≤3(10006) → 每年≤3(10012,仅非认证) → 终身≤12(10013,全部) → 每户≤6(10014) → 落库 → AssignRole
查看用户详情：GetUser(viewer_id) → 同屋判定 → 手机号明文/脱敏 + 房屋号
```

---

## 5. Proto 变更

| 文件 | 变更类型 | 说明 |
|------|:---:|------|
| `api/auth/v1/auth.proto` | 兼容新增 | `RefreshTokenRequest` 增 `string device_type = 2`（CLAR-1 a） |
| `api/permission/v1/permission.proto` | 兼容新增 | `Role` 增 `repeated string platforms = 10` |
| `api/user/v1/user.proto` | 兼容新增 | ① `GetUserRequest` 增 `optional int64 viewer_id = 2 [jstype=JS_STRING]`；② `GetUserResponse` 增 `SameHouseInfo same_house = 3`（新 message）；③ 新增 `GetAppState`/`SetCurrentCommunity` RPC + 消息 |
| `api/masterdata/v1/masterdata.proto` | 兼容新增 | 新增 `GetSectionQuota` RPC + `GetSectionQuotaReq/Resp` |

- 全部为**新增字段/消息/RPC**，无字段删除或类型变更 → `breaking-check` 应无破坏性变更。
- 已核实**无需变更**：`JoinCommunityRequest`/`CommunityMembership` 的 `building/unit/room` 已存在（CLAR-4 仅逻辑层，见 STAGE3-6）。
- 变更后按 `api-proto/CLAUDE.md` 通知清单：auth/user/permission/masterdata 四方 + 消费方（auth 消费 user+permission；user 消费 permission；community-hub 消费 masterdata）。

---

## 6. 安全考虑

- **端限制定位为 UX 引导，非安全边界**：真正安全由 RBAC + 数据权限兜底；auth 判定 fail-open（空 platforms / 未知 device_type / 零角色均放行），严禁 `is_system` 短路端判定（引 `[[is-system-no-permission-shortcut]]`）。
- **手机号 PII**：同屋互见返回真实手机号前必须 `crypto.AESDecrypt`；非同屋一律脱敏且不返回楼栋房屋号；严禁明文落日志（引 `[[phone-encryption]]`、`[[pii-plaintext-logging]]`）。
- **数据权限兜底**：`SetCurrentCommunity` 切换校验、`AssertPublishScope` 发布校验均以 JWT 身份为准，`community_id`/`publisher_id`/`device_type` 客户端传值不可信。
- **房屋上限/配额不得前端硬编码**：上限从后端配置读取（`user.max_house_members` / `sys_section_quota`），前端仅 UX 提示（引 `[[frontend-business-rule-hardcode]]`）。
- **自动授权防绕过**：加入即 grant `status=0` 已即时赋予 scope+level-0 能力，房屋/次数限制须防「反复退出重加入」绕过（引 `[[auto-grant-unverified-grant-confers-scope-level0]]`）。
- **错误码语义唯一**：新码 `50007/80007/10014/10015` 全仓 grep 确认未占用，一码一义（引 `[[error-code-collision-and-namespace-alignment]]`）。

---

## 7. 前端归属

| 功能 | 前端 | 落点 |
|------|------|------|
| 端限制登录引导（50007 → 「请使用移动端 APP」） | web/pc | `web/pc/src/views/auth` 登录页错误分支 |
| 角色 platforms 配置（sys_admin 维护） | web/pc | `web/pc/src/views/roles` 角色表单增「允许登录端」多选 |
| 加入小区楼/单元/房号必填表单 | web/mobile | `web/mobile/src/pages/join-community` |
| 当前小区切换 UI | web/mobile | 首页/「我的」顶部上下文切换入口 |
| 同屋手机号 + 楼栋房屋号展示 | web/mobile | 用户详情页消费 `GetUserResponse.same_house` |

---

## 8. 记忆引用（设计阶段预防性注入）

| 记忆 | 适用章节 | 设计决策 |
|------|---------|---------|
| `[[is-system-no-permission-shortcut]]` | 端限制 | `platforms` 为配置驱动属性，auth 判定不得用 `is_system` 短路 |
| `[[phone-encryption]]` | 同屋互见 | 返回真实手机号前 `AESDecrypt`，脱敏与加密存储互斥 |
| `[[api-required-field-marked-optional]]` | 成员约束 | JoinCommunity 楼/单元/房号必填，API types.go 移除 `,optional` |
| `[[frontend-business-rule-hardcode]]` | 前端 | 房屋上限 ≤6/配额上限不得前端硬编码，从后端配置读取 |
| `[[error-code-collision-and-namespace-alignment]]` | 错误码 | 新码 50007/80007/10014/10015 语义唯一、避让现状命名空间 |
| `[[auto-grant-unverified-grant-confers-scope-level0]]` | 成员约束 | 加入自动授权 status=0 即时生效，房屋/次数限制防退出重加入绕过 |
| `[[pii-plaintext-logging]]` | 同屋互见 | 加密手机号严禁明文打日志 |
| `[[migration-must-execute]]` | 数据模型 | 新增 migration 提交后必须在库执行验证 |
| `[[grpc-only-comms]]` | 接口设计 | auth→permission、community-hub→master-data 均走 gRPC，禁止直连 DB |
| `[[proto-jstype]]` | Proto 变更 | 新 int64 ID 字段（viewer_id 等）标注 `[jstype=JS_STRING]` |
