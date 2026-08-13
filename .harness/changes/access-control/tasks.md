# Tasks: 访问控制与数据权限改造（access-control）

> **对执行 Agent 的指令**：每个 Task 独立可测，含逻辑代码的 Task 按 TDD 执行（先写测试→看失败→写实现→看通过→重构）。精确到文件路径，零占位符。
> **依赖顺序**：全局/Proto → permission-service（platforms 存储）→ auth-service（端判定，依赖 platforms 透出）→ master-data（配额配置）→ community-hub（配额校验，依赖 GetSectionQuota）→ user-service（成员约束/当前小区/同屋互见，依赖 permission GetDataScopes/GetUserRoles）→ 前端。
> **对 predecessor `access-data-permission` 交付物的依赖**：`AssertPublishScope`/`GetDataScopes`/scope 三态/`AssignRole` 均已交付，直接消费，不重复实现。

---

## 全局 / Proto（由全局 Claude / Owner 执行）

### Task 0.1: auth.proto RefreshTokenRequest 增 device_type
- **文件**: `api-proto/api/auth/v1/auth.proto`
- [ ] `message RefreshTokenRequest`（L184）增 `string device_type = 2;`（注释「设备类型：web/ios/android/miniapp/admin，端限制刷新判定用，CLAR-1 a」）
- [ ] 保持 `refresh_token = 1` 不变（兼容新增，非破坏性）

### Task 0.2: permission.proto Role 增 platforms
- **文件**: `api-proto/api/permission/v1/permission.proto`
- [ ] `message Role`（L277）增 `repeated string platforms = 10;`（注释「允许登录的端：pc/mobile；空=未声明（fail-open 允许所有端）」）
- [ ] 不新增 int64 ID 字段（无需 jstype）

### Task 0.3: user.proto 当前小区 + 同屋互见
- **文件**: `api-proto/api/user/v1/user.proto`
- [ ] `GetUserRequest`（L183）增 `optional int64 viewer_id = 2 [jstype = JS_STRING];`（注释「查看者 ID；0/缺省=无查看上下文，手机号脱敏」）
- [ ] 新增 `message SameHouseInfo { bool same_house = 1; int32 building = 2; int32 unit = 3; int32 room = 4; }`
- [ ] `GetUserResponse`（L187）增 `SameHouseInfo same_house = 3;`
- [ ] 新增 `GetAppState`/`SetCurrentCommunity` RPC 及消息：`GetAppStateRequest{int64 user_id=1 [jstype=JS_STRING]}`、`GetAppStateResponse{BaseResp base; int64 current_community_id=2 [jstype=JS_STRING]; int64 updated_at=3;}`、`SetCurrentCommunityRequest{int64 user_id=1 [jstype=JS_STRING]; int64 community_id=2 [jstype=JS_STRING];}`、`SetCurrentCommunityResponse{BaseResp base;}`
- [ ] 在 `service UserService` 注册两个 RPC

### Task 0.4: masterdata.proto 增 GetSectionQuota
- **文件**: `api-proto/api/masterdata/v1/masterdata.proto`
- [ ] 新增 `message GetSectionQuotaReq { string section_type = 1; }`
- [ ] 新增 `message GetSectionQuotaResp { common.v1.BaseResp base = 1; int64 max_count = 2; bool configured = 3; }`（注释「configured=false 表示未配置=不限」）
- [ ] `service MasterdataService` 注册 `rpc GetSectionQuota(GetSectionQuotaReq) returns (GetSectionQuotaResp);`

### Task 0.5: 生成代码 + CI
- **文件**: `api-proto/`（执行命令）
- [ ] `cd api-proto && make generate`
- [ ] `make lint` → 0 errors
- [ ] `make breaking-check` → 确认无破坏性变更（全部新增字段/消息/RPC）
- [ ] `make ci` 全绿；更新 `api-proto/CHANGELOG.md` 记录 4 个 proto 变更
- [ ] 按 `api-proto/CLAUDE.md` 通知清单通知消费方（auth/user/permission/masterdata/community-hub）

### Task 0.6: 修正设计稿 §7 计数公式 + §4.1 fail-open 语义（STAGE3-2 / STAGE3-4）
- **文件**: `docs/specs/access-control-design.md`
- [ ] §7 补记状态机：「`status`（业务态）与 `moderation_status`（审核态）正交；内容创建即 `status='active'`（业务在版），待审期间 `moderation_status=0` 但 status 仍为 active，故待审与展示同占配额」；计数谓词 `deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)` 保持不变，注明「下架/移除复用 status 非 active 语义」
- [ ] §4.1 补注：「默认映射」为**种子初值**（`sys_role.platforms` 由 `init_permissions.sql` 显式写入），运行时回退为 **fail-open**（空 platforms 允许所有端）；二者语义区分，避免实现二义
- [ ] §4.2 错误码 `050006` → `50007`；§7 错误码 `80007` 保持；§3.5/§11.2 `10014`/`10006` 对齐最终错误码表（见 design §0 STAGE3-3）

---

## permission-service

### Task 1.1: sys_role 增 platforms 列（Migration）
- **创建**: `services/permission-service/migration/002_add_role_platforms.sql`
- [ ] `ALTER TABLE sys_role ADD COLUMN platforms VARCHAR(32) NOT NULL DEFAULT '' COMMENT '允许登录的端，逗号分隔：pc,mobile；空=未声明（fail-open）';`
- [ ] 提交后在库执行并验证（`SHOW COLUMNS FROM sys_role LIKE 'platforms'`）
- [ ] SEE: `[[migration-must-execute]]`

### Task 1.2: SysRole model 增 Platforms 字段 + 查询补列
- **修改**: `services/permission-service/model/permission.go`
- [ ] `type SysRole` 增 `Platforms string \`db:"platforms"\``
- [ ] `Insert` SQL 增 `platforms` 列；`FindOne`/`FindByCode`/`FindByIds`/`FindList` 的 `select *` 已覆盖新列（确认无显式列名遗漏）
- [ ] `Update` SQL 增 `platforms = ?`（sys_admin 维护 platforms 需可写）
- [ ] **TDD**: 无独立逻辑，仅 model 字段映射；`go build ./...` 通过即可

### Task 1.3: Role proto 透出 platforms（toProtoRole）
- **修改**: `services/permission-service/rpc/internal/logic/permission/helpers.go`（或等价转换函数）
- [ ] `toProtoRole` 填充 `Platforms: splitComma(s.Platforms)`（`"" → []`）
- [ ] **RED**: 在对应 `_test.go` 写 table-driven：`platforms="" → []`、`"pc" → ["pc"]`、`"pc,mobile" → ["pc","mobile"]`
- [ ] **GREEN**: 实现 split/join 转换，测试 PASS
- [ ] **REFACTOR**: 提取 `splitPlatforms/joinPlatforms` 复用，保持绿

### Task 1.4: seed platforms + 新接口权限码
- **修改**: `services/permission-service/scripts/init_permissions.sql`
- [ ] 为 9 个内置角色 `UPDATE sys_role SET platforms=... WHERE role_code=...`：`sys_admin=pc`、`community_admin=pc,mobile`、`property_admin=pc`、`owner/tenant/grid_worker/committee/merchant/registered_user=mobile`（`registered_user` id=9 与其余 8 角色同在此文件种子，见 L18-23/L229）
- [ ] 新增权限码（`sys_permission` type=3）：`GET:/api/users/me/app-state`（`user:appstate:read-api`）、`PUT:/api/users/me/current-community`（`user:currentcommunity:write-api`），并挂到 `registered_user`（及 owner/tenant 等）的 `rel_role_permission`
- [ ] 确认 path 与 user-service API 实际路由一致（引 `[[permission-seed-api-path-must-match-routes]]`）
- [ ] SEE: `[[is-system-no-permission-shortcut]]`（platforms 为配置属性，不参与权限短路）

---

## auth-service

### Task 2.1: 增 PermissionServiceRpc 客户端
- **修改**: `services/auth-service/rpc/internal/config/config.go`
- **修改**: `services/auth-service/rpc/internal/svc/servicecontext.go`
- **修改**: `services/auth-service/rpc/etc/authservice.yaml`
- [ ] config 增 `PermissionServiceRpc zrpc.RpcClientConf`；yaml 增 `PermissionServiceRpc` target
- [ ] servicecontext 增 `PermissionClient permissionv1.PermissionServiceClient`
- [ ] `go build ./...` 通过（未接线不报错）

### Task 2.2: 端准入判定 helper（checkPlatformAccess）
- **创建**: `services/auth-service/rpc/internal/logic/auth/platform.go`
- **创建**: `services/auth-service/rpc/internal/logic/auth/platform_test.go`
- [ ] 定义 `classifyDeviceType(deviceType string) string`：`web/admin→pc`、`ios/android/miniapp→mobile`、空/未知→`""`
- [ ] 定义 `checkPlatformAccess(ctx, svcCtx, userID, deviceType) error`：归类为空→放行；`GetUserRoles` 失败→放行（UX 引导，失败不锁人）并 `Infof`；遍历 roles，任一 platforms 空→放行、任一含当前端→放行；否则返回 `responsex.NewBaseRespWithError(50007, "该账号为移动端用户，请使用移动端 APP")`
- [ ] **RED**: table-driven 覆盖：mobile 角色+web→50007、mobile 角色+android→nil、双端角色+web→nil、空 platforms 角色+web→nil、未知 device_type→nil、零角色→nil
- [ ] **GREEN**: 实现，测试 PASS
- [ ] **REFACTOR**: 提取 `roleAllows(platforms, deviceClass)`，保持绿
- [ ] SEE: `[[is-system-no-permission-shortcut]]`（不得用 is_system 短路）

### Task 2.3: Login/LoginSms/Register/RefreshToken 挂载端判定
- **修改**: `services/auth-service/rpc/internal/logic/auth/loginlogic.go`
- **修改**: `services/auth-service/rpc/internal/logic/auth/loginsmslogic.go`
- **修改**: `services/auth-service/rpc/internal/logic/auth/registerlogic.go`
- **修改**: `services/auth-service/rpc/internal/logic/auth/refreshtokenlogic.go`
- **修改**: 各对应 `*_test.go`
- [ ] 在 4 处「签发 Token 前」调用 `checkPlatformAccess(svcCtx, userID, in.DeviceType)`，非 nil 直接返回 `Base=50007`
- [ ] refresh 使用 `in.DeviceType`（proto T0.1 新增）
- [ ] **RED**: 各 logic 增 1 条「端拒绝→50007」用例（mock PermissionClient 返回 mobile-only 角色）
- [ ] **GREEN**: 实现挂载，测试 PASS
- [ ] **REFACTOR**: 4 处调用统一，保持绿

---

## master-data-service

### Task 5.1: sys_section_quota 表 + 种子（Migration）
- **创建**: `services/master-data-service/migration/005_add_section_quota.sql`
- [ ] 建表（见 design §2.3，含 `UNIQUE KEY uk_section_type`）
- [ ] `INSERT ... (section_type='lost_found', max_count=5)`
- [ ] 提交后在库执行验证
- [ ] SEE: `[[migration-must-execute]]`

### Task 5.2: GetSectionQuota logic
- **创建**: `services/master-data-service/model/mdSectionQuotaModel.go`
- **创建**: `services/master-data-service/rpc/internal/logic/configuration/getSectionQuotaLogic.go`
- **创建**: `services/master-data-service/rpc/internal/logic/configuration/getSectionQuotaLogic_test.go`
- [ ] model：`FindBySectionType(ctx, sectionType) (*SectionQuota, error)`（`select ... where section_type=? limit 1`，`sqlx.ErrNotFound → ErrNotFound`）
- [ ] logic：查无记录 → `configured=false,max_count=0`；命中 → `configured=true,max_count`；DB 瞬时错误向上抛（区分 ErrNotFound 与瞬时错误，引 `[[notfound-cache-sentinel-vs-transient-error]]`）
- [ ] **RED**: table-driven：未配置→configured=false、lost_found→configured=true&max_count=5、DB 错误→返回 error
- [ ] **GREEN**: 实现，测试 PASS
- [ ] **REFACTOR**: 常量 `sectionQuotaTable="sys_section_quota"`，保持绿
- [ ] 在 svc 注册 handler（`rpc` 入口）

---

## community-hub-service

### Task 4.1: CountQuotaOccupied model 方法
- **修改**: `services/community-hub-service/model/lost_found_item.go`
- **创建**: `services/community-hub-service/model/lost_found_item_quota_test.go`
- [ ] 增 `CountQuotaOccupied(ctx, publisherId, communityId int64, typ string) (int64, error)`：`SELECT COUNT(*) FROM lost_found_items WHERE publisher_id=? AND community_id=? AND type=? AND deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)`
- [ ] **RED**: table-driven（直接打 DB 或 mock sqlx）：待审(0)+active 计入、通过(1)+active 计入、驳回(2) 不计、resolved 不计、deleted_at 非空不计
- [ ] **GREEN**: 实现，测试 PASS
- [ ] **REFACTOR**: 谓词提取常量/注释，保持绿

### Task 4.2: sectionquota 校验 helper
- **创建**: `services/community-hub-service/rpc/internal/logic/scope/section_quota.go`
- **创建**: `services/community-hub-service/rpc/internal/logic/scope/section_quota_test.go`
- [ ] 定义 `CheckSectionQuota(ctx, svcCtx, userID, communityID int64, sectionType string) error`
- [ ] 逻辑：`MasterDataClient.GetSectionQuota(sectionType)`；`!configured`→nil；`CountQuotaOccupied(...) >= max_count`→`NewBaseRespWithError(80007, "超出发布配额")`；否则 nil
- [ ] **RED**: mock MasterDataClient + model：未配置→nil、4/5→nil、5/5→80007、GetSectionQuota 失败→透传 error
- [ ] **GREEN**: 实现，测试 PASS
- [ ] **REFACTOR**: 抽 `quotaAllowed(count, max) bool`，保持绿

### Task 4.3: CreateLostFound 挂载配额校验
- **修改**: `services/community-hub-service/rpc/internal/logic/lostfound/createlostfoundlogic.go`
- **修改**: `services/community-hub-service/rpc/internal/logic/lostfound/createlostfoundlogic_test.go`
- [ ] 在 `AssertPublishScope` 之后、`Insert` 之前调用 `scope.CheckSectionQuota(ctx, svcCtx, publisherId, communityId, itemType)`，非 nil 直接返回
- [ ] **RED**: 增「达上限→80007」用例（mock 配额返回 5 且已占 5）
- [ ] **GREEN**: 实现挂载，测试 PASS
- [ ] **REFACTOR**: 保持绿

### Task 4.4: 错误码 80007 登记
- **修改**: `services/community-hub-service/api/internal/types/types.go`
- [ ] 增常量 `CodeSectionQuotaExceeded = 80007 // 超出发布配额（板块配额）`；注释区登记「080007 语义」
- [ ] `go build ./...` 通过

---

## user-service

### Task 3.1: user_app_state 表（Migration）
- **创建**: `services/user-service/migration/005_add_user_app_state.sql`
- [ ] 建表（见 design §2.2）
- [ ] 提交后在库执行验证
- [ ] SEE: `[[migration-must-execute]]`

### Task 3.2: UserAppState model
- **创建**: `services/user-service/model/user_app_state.go`
- [ ] 定义 struct + `FindOne(userId)`、`Upsert(userId, communityId)`（`INSERT ... ON DUPLICATE KEY UPDATE current_community_id=?, updated_at=NOW()`）
- [ ] `FindOne` 无记录返回 `ErrNotFound`
- [ ] **TDD**: 无独立逻辑；`go build ./...` 通过即可

### Task 3.3: GetAppState / SetCurrentCommunity logic
- **创建**: `services/user-service/rpc/internal/logic/user/get_app_state_logic.go`
- **创建**: `services/user-service/rpc/internal/logic/user/set_current_community_logic.go`
- **创建**: `services/user-service/rpc/internal/logic/user/current_community_logic_test.go`
- [ ] `GetAppState`：读 model，无记录返回 `current_community_id=0`；有返回 id+updated_at
- [ ] `SetCurrentCommunity`：调 `PermissionClient.GetDataScopes(user_id,"community")`；`state==GLOBAL`→放行；`EMPTY`→`10015 "目标小区不在数据范围"`；`LIMITED`→`community_id ∈ scope_ids` 才放行否则 `10015`；放行后 `Upsert`
- [ ] **RED**: table-driven：GLOBAL 放行、EMPTY 拒绝 10015、LIMITED 命中放行、LIMITED 未命中拒绝 10015、GetDataScopes 失败→透传
- [ ] **GREEN**: 实现，测试 PASS
- [ ] **REFACTOR**: 抽 `inScope(state, scopeIds, communityID) bool`，保持绿

### Task 3.4: JoinCommunity 房屋必填 + 每户 ≤6
- **修改**: `services/user-service/rpc/internal/logic/user/join_community_logic.go`
- **修改**: `services/user-service/model/user_community_membership.go`
- **修改**: `services/user-service/rpc/internal/logic/user/join_community_logic_test.go`
- [ ] 顶部增必填校验：`in.Building <= 0 || in.Unit <= 0 || in.Room <= 0` → `10040 "楼/单元/房号必填"`（替换现状 `if in.Building > 0 && in.Room > 0` 的跳过式校验）
- [ ] model 增 `CountActiveByAddress(ctx, communityId, building, unit, room, excludeUserId) (int64, error)`：`bind_status=active AND user_id<>excludeUserId`（新建用户 excludeUserId 传 0）
- [ ] 用 `CountActiveByAddress >= maxHouseMembers`（`user.max_house_members`，默认 6，sys_config 可配）替换 `FindByAddress` 唯一性校验，超限 → `10014 "该房屋已满员"`；移除 `10011` 路径
- [ ] **RED**: 覆盖：缺 building→10040、房屋 5 人+新用户→放行、房屋 6 人+新用户→10014、退出者不计、计数排除当前用户（重新激活场景）
- [ ] **GREEN**: 实现，测试 PASS
- [ ] **REFACTOR**: 保持绿
- [ ] SEE: `[[api-required-field-marked-optional]]`、`[[auto-grant-unverified-grant-confers-scope-level0]]`

### Task 3.5: 终身限制对齐 spec（对全部用户生效）+ 认证粒度改 per-community
- **修改**: `services/user-service/rpc/internal/logic/user/join_community_logic.go`
- **修改**: `services/user-service/rpc/internal/logic/user/join_community_logic_test.go`
- [ ] 将「终身 `10013` 校验」从 `if !isVerifiedOwnerOrTenant(...)` 块**移出**，仅保留在 `isFirstJoin` 块内（终身对全部用户生效）
- [ ] 将 `isVerifiedOwnerOrTenant` 从「全局（任一小区认证）」改为「per-community（目标小区认证）」：签名增 `targetCommunityId`，仅校验目标小区 `community_id` 的 owner/tenant 认证状态（STAGE3-1 per-community）
- [ ] **RED**: 增「认证用户终身达 12 → 10013」用例；增「A 小区认证、加入 B 小区 → 受 B 每年限制」用例
- [ ] **GREEN**: 实现，测试 PASS
- [ ] **REFACTOR**: 保持绿

### Task 3.6: GetUser 同屋互见判定
- **修改**: `services/user-service/rpc/internal/logic/user/get_user_logic.go`
- **创建**: `services/user-service/rpc/internal/logic/user/same_house.go`
- **创建**: `services/user-service/rpc/internal/logic/user/same_house_test.go`
- [ ] 定义 `maskPhone(phone string) string`（`138****1234`）
- [ ] 定义 `isSameHouse(ctx, svcCtx, viewerID, targetID int64) (bool, building, unit, room int32, err error)`：查双方 active membership，`community+building+unit+room` 全同 → true
- [ ] `GetUser`：`in.ViewerId == 0` → 脱敏+无房屋号；`ViewerId == targetId` → 明文+自身房屋号；否则 `isSameHouse` 决定明文/脱敏+房屋号
- [ ] **RED**: table-driven：无 viewer→脱敏、self→明文、同屋→明文+房屋号、非同屋→脱敏+无房屋号、解密失败兜底返回原值
- [ ] **GREEN**: 实现（`crypto.AESDecrypt` 明文、脱敏、同屋判定），测试 PASS
- [ ] **REFACTOR**: 保持绿
- [ ] SEE: `[[phone-encryption]]`、`[[pii-plaintext-logging]]`（严禁明文日志）

### Task 3.7: API 层 types + 新接口 handler
- **修改**: `services/user-service/api/internal/types/types.go`
- **修改**: `services/user-service/api/internal/handler/routes.go`
- **创建**: `services/user-service/api/internal/handler/user/*_handler.go`（app-state/current-community）
- **创建**: `services/user-service/api/internal/logic/user/*_logic.go`（app-state/current-community）
- [ ] `types.go`：`JoinCommunityReq` 的 `Building/Unit/Room` 移除 `,optional`（必填，引 `[[api-required-field-marked-optional]]`）；新增 `GetAppStateReq/Resp`、`SetCurrentCommunityReq/Resp`
- [ ] `routes.go`：注册 `GET /api/users/me/app-state`、`PUT /api/users/me/current-community`（JWT 认证）
- [ ] handler/logic 转发到 RPC（`GetAppState`/`SetCurrentCommunity`），透出 `10015`
- [ ] **RED**: 各 logic 单测（转发 + 透错）
- [ ] **GREEN**: 实现，测试 PASS

---

## 前端 web/pc

### Task 6.1: 登录 50007 引导
- **修改**: `web/pc/src/views/auth`（登录页 + 登录 api）
- [ ] 登录/刷新失败捕获 `code==50007` 时展示「该账号为移动端用户，请使用移动端 APP」引导，不发通用错误 toast
- [ ] 复用 `web/common` 已有错误处理，不重复定义类型（引 `[[web-common-type-reuse-no-redefine]]`）
- [ ] `npm run build`（或 vue-tsc）通过

### Task 6.2: 角色管理 platforms 配置
- **修改**: `web/pc/src/views/roles`（角色表单/列表）
- [ ] 角色表单增「允许登录端」多选（pc/mobile），写入/回显 `platforms`（消费 permission `Role.platforms`）
- [ ] 列表/详情展示 platforms
- [ ] 从 `web/common` 复用 Role 类型（若有），不重定义（引 `[[web-common-type-reuse-no-redefine]]`）

---

## 前端 web/mobile

### Task 7.1: 加入小区楼/单元/房号必填
- **修改**: `web/mobile/src/pages/join-community`
- [ ] 表单楼/单元/房号改为必填（前端校验仅 UX 提示，权威在后端）；区间校验不硬编码，与后端对齐（引 `[[frontend-business-rule-hardcode]]`）
- [ ] 提交移除可选兜底，缺省即拦

### Task 7.2: 当前小区切换 UI
- **修改**: `web/mobile/src/pages`（首页/「我的」顶部上下文）+ `web/mobile/src/api`
- [ ] 新增 `app-state`/`current-community` api 调用封装
- [ ] 顶部展示当前小区，切换下拉调用 `PUT /api/users/me/current-community`，失败（10015）提示「目标小区不在你的数据范围」
- [ ] 切换后刷新首页上下文

### Task 7.3: 同屋手机号 + 楼栋房屋号展示
- **修改**: `web/mobile/src/pages`（用户详情页）+ `web/mobile/src/api`
- [ ] 消费 `GetUserResponse.same_house`：`same_house=true` 展示真实手机号+楼栋房屋号；否则展示脱敏手机号、不展示房屋号
- [ ] 复用 `web/common` 类型，不重定义（引 `[[web-common-type-reuse-no-redefine]]`）

---

## Self-Review 结论（Step 5）

- **占位符扫描**：无 `<任务描述>`/TBD/TODO；每个 Task 精确到文件路径。
- **TDD 覆盖**：含逻辑的 Task（T2.2/T2.3/T3.3/T3.4/T3.5/T3.6/T4.1/T4.2/T4.3/T5.2）均含 RED→GREEN→REFACTOR；纯 model/migration/前端无逻辑代码标 `go build`/构建即可。
- **依赖顺序**：Proto(0.x) → permission(1.x) → auth(2.x 依赖 1.3 platforms 透出) → master-data(5.x) → community-hub(4.x 依赖 5.2) → user-service(3.x 依赖 permission GetDataScopes/GetUserRoles) → 前端(6.x/7.x 依赖各服务 API)。
- **独立可测**：各 Task 可独立完成；跨服务仅依赖已交付 predecessor 能力。
- **记忆引用检查**：9 个相关记忆（is-system-shortcut / phone-encryption / api-required-optional / frontend-hardcode / error-code-alignment / auto-grant-level0 / pii-logging / migration-execute / grpc-only）均已注入至 design §8 与对应 Task 的 SEE 标注；补充 `[[permission-seed-api-path-must-match-routes]]`（T1.4）、`[[notfound-cache-sentinel-vs-transient-error]]`（T5.2）、`[[web-common-type-reuse-no-redefine]]`（T6.x/T7.x）。
