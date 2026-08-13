# CHANGELOG — community-hub-service

## 2026-08-13 — 审核可见性门禁：读路径仅返回审核通过内容

### 做了什么
- `model/notice.go` + `model/lost_found_item.go`：
  - `FindList` 查询加 `moderation_status = 1` 过滤（列表只返回审核通过内容）
  - 新增 `FindOnePublished`（读路径专用，仅返回 moderation_status=1；`FindOne` 保留给写接口 / 审核回调内部使用）
- `getnoticelogic.go` + `getlostfoundlogic.go`：读路径 `FindOne` → `FindOnePublished`
- 新增 TDD 测试：`TestGetNotice_FilterByModerationStatus` / `TestGetLostFound_FilterByModerationStatus`

### 为什么
数据权限 Wave 阶段④遗留（summary.md WARN）：读列表/详情只按数据范围（scope）过滤，未按审核状态过滤，
普通用户能看到「待审核/拒绝」的内容。本变更落地「最小实现」：全员读路径仅见审核通过内容。

### 影响
- 读路径语义：moderation_status=0（待审核）/2（拒绝）内容对普通用户不可见（表现为「不存在」80001/80004）
- 写接口 / moderation 回调内部仍用 FindOne（不受影响）
- 门禁：harness-checks 16 PASS / 0 FAIL / 0 WARN

## 2026-08-12 — TDD 证据补齐（8 个包装函数无测试，QA TDD FAIL → 修复）

### 背景
终局 LLM QA 的 TDD 证据表发现：上一轮（Get-by-ID 数据范围 + RPC 身份伪造修复）核心函数
（GetNotice/GetLostFound scope 过滤、API CallCtx+ToError、回环绑定、scope 包）全部有测试，
但**同批被连带修改的 8 个包装函数无对应测试**，判定 QA FAIL（TDD 证据不足）：
- RPC `ListLostFoundLogic.ListLostFound` / `ListContactsLogic.ListContacts`（新增 FilterAllowed scope 过滤）
- API `ListNotices/ListLostFound/ListContactsLogic`（新增 CallCtx 身份注入）
- API `UpdateNotice/DeleteNotice/ResolveLostFoundLogic`（新增 CallCtx + responsex.ToError 透出 080006）

根因：Generator 只给变更的 headline 函数写测试（本轮=Get-by-ID），把同构的列表/更新/删除/解决
包装函数当作"同逻辑已被代表函数覆盖"，但 Go 测试隔离使这些函数从未被任何测试引用；机械门禁
（harness-checks.go_test / tdd-evidence-validator）只查包级测试文件与 RED 列文本格式，无按函数维度门禁。

### 修复内容（仅补测试 + fake，无生产代码改动）
- 新建 `rpc/internal/logic/lostfound/listlostfound_filter_test.go`：`TestListLostFound_FilterByScope`
  四态（EMPTY 不查库 / LIMITED 未命中不查库 / LIMITED 命中查询 / GLOBAL 查询），照抄
  `rpc/internal/logic/notice/listnotices_filter_test.go` 模板；复用已存在的 fakeReadPerm/readPerm/ctxWithUserID。
- 新建 `rpc/internal/logic/contact/listcontacts_filter_test.go`：`TestListContacts_FilterByScope` 同四态；
  本文件新增 contact 包的 `fakeReadPerm`/`listPerm`（GetDataScopes fake）。
- 新建 `api/internal/logic/notice/listnoticeslogic_test.go`、`api/internal/logic/lostfound/listlostfoundlogic_test.go`、
  `api/internal/logic/contact/listcontactslogic_test.go`：`TestList*_InjectsIdentity`，照抄
  getnoticelogic_test.go 的 `TestGet*_InjectsIdentity` 模板（出站 metadata 断言 user_id=42）。
- 新建 `api/internal/logic/notice/updatenoticelogic_test.go`、`deletenoticelogic_test.go`、
  `api/internal/logic/lostfound/resolvelostfoundlogic_test.go`：`Test*_InjectJWTAndSurfaceBaseError`
  双路径（成功 Base → nil；080006 Base → errx.FromError Code==80006 + 身份注入），照抄
  `api/internal/logic/contact/upsertcontactslogic_test.go` 的 InjectJWT + SurfaceBaseError 模板。
- fake 扩展：`rpc/internal/logic/lostfound/publishscope_test.go` 的 `fakeLostFoundModel` 补
  `FindList`（findListCalled 标志）；`rpc/internal/logic/contact/upsertcontactslogic_test.go` 的
  `fakeContactModel` 补 `FindByCommunityId`（findByCalled 标志）。

### TDD 证据（RED 摘录，均通过 `git stash` 回退生产文件到 HEAD 复现真实 FAIL 后恢复）
- RPC 过滤（HEAD 无 FilterAllowed，EMPTY/LIMITED-未命中仍查库）：
  - `listlostfound_filter_test.go:64: Error: Not equal: expected: false`（findListCalled 应为 false）
  - `listcontacts_filter_test.go:83: Error: Not equal: expected: false`
- API 列表（HEAD 无 CallCtx，出站 ctx 无 metadata）：
  - `listnoticeslogic_test.go:48 / listlostfoundlogic_test.go:48 / listcontactslogic_test.go:47:
    Error: Should be true / 出站 metadata 必须存在`
- API 写路径（HEAD 无 CallCtx + 无 ToError，080006 Base 被静默吞掉）：
  - `updatenoticelogic_test.go:59 / deletenoticelogic_test.go:59 / resolvelostfoundlogic_test.go:59:
    Error: An error is expected but got nil.`（越权 080006 未透出）
  - 同测试注入子用例：`Error: Should be true / 出站 metadata 必须存在`

### 影响
- 生产代码零改动（8 个逻辑文件维持工作树原状）；仅新增 8 个测试文件 + 扩展 2 个 fake
- 门禁：`go build ./...` + `go vet ./...` 全绿；`go test ./... -count=1` 全绿（8 个新测试函数全 PASS）
- 修复后 TDD 证据表：26 个新增/修改函数全部有测试命中，无 ❌

---

## 2026-08-12 — 多视角评审修复（Get-by-ID 数据范围 + RPC 身份伪造，TDD）

### 背景
阶段④交付后多视角审查 1/3 PASS：安全架构 FAIL(2 CRITICAL)、设计业务 FAIL(1 CRITICAL)。
本条目修复全部 3 项 CRITICAL。

### 修复内容

**1. GetNotice/GetLostFound 单条读取补数据范围过滤（security-arch + design-biz 双重 CRITICAL）**
- 根因：T4.6/REQ-1.6 读过滤只挂载于 List（ListNotices/ListLostFound/ListContacts），
  Get-by-ID 直接 `FindOne(id)` 返回完整内容（notice 全文/附件、lost_found 的
  description/contact_phone），无 FilterAllowed/AssertCommunityScope。LIMITED/EMPTY 用户
  被列表过滤后仍可按 ID（Snowflake 时间有序可枚举 / 分享链接）越权读取，违背
  『注册用户读不到小区内部内容』。
- 修复（RPC 层）：
  - `rpc/internal/logic/notice/getnoticelogic.go` / `lostfound/getlostfoundlogic.go`：
    `FindOne` 后 reverse-lookup 内容 community_id → `scope.FilterAllowed(userID, communityID)`；
    不通过 → `scope.DenyBase()`(080006) + 空内容，且不查询附件/不返回越权内容；
    传输错误原样传播（fail-closed）。
- 修复（API 层）：
  - `api/internal/logic/notice/getnoticelogic.go` / `lostfound/getlostfoundlogic.go`：
    补 `CallCtx` 注入（与 List 一致，使 JWT 身份经 metadata 到达 RPC 读过滤）；
    `responsex.ToError(resp.GetBase())` 透出 080006/80001 业务错误（此前静默返回 200 空）。
- 回归测试（先 RED 后 GREEN，断言越权 Get → 080006 + 空 + 不查附件）：
  - `rpc/internal/logic/notice/getnoticelogic_test.go`（5 用例：LIMITED 命中/未命中、EMPTY、无身份 fail-closed、GLOBAL）
  - `rpc/internal/logic/lostfound/getlostfoundlogic_test.go`（同矩阵 5 用例）
  - `api/internal/logic/notice/getnoticelogic_test.go`、`api/internal/logic/lostfound/getlostfoundlogic_test.go`
    （身份注入 + 080006 透出）

**2. RPC 身份伪造风险（security-arch CRITICAL，仓库级模式）**
- 根因：数据权限身份经未认证的 gRPC metadata 传输，RPC 绑定 `0.0.0.0:8088` 且无服务间鉴权，
  `UserIDFromCtx` 盲信入站 user_id → 局域网 / Docker 桥接网络对端可注入任意身份，击穿数据权限。
- 修复（本服务可落地部分）：
  - `rpc/etc/communityhub.yaml`：`ListenOn: 0.0.0.0:8088 → 127.0.0.1:8088`（回环绑定）。
    go-zero `figureOutListenOn` 对非 0.0.0.0 host 原样注册 etcd；单机拓扑（scripts/start.sh
    全部 go run 于宿主）下 API 网关 + moderation 回调可正常发现。阻断局域网 / Docker 桥接对端。
  - `rpc/internal/logic/scope/userctx.go`：信任边界文档化（metadata 盲信的安全前提 = 网络隔离），
    并标注仓库级根治方向（服务凭据/mTLS + unary 拦截器，涉及 common/ 与全部 9 服务，Owner 协调）。
  - `rpc/internal/config/config_test.go`：配置不变式测试 `TestRpcConfig_BindsLoopback`，
    ListenOn 回退 0.0.0.0 即 FAIL（RED 已复现，`Should be true / host="0.0.0.0"`）。
  - 沉淀记忆 `global/rpc-identity-spoofing-loopback-isolation.md`（仓库级 9 服务 0.0.0.0 无鉴权模式），
    已建索引 + 登记 MEMORY.md should-follow。

### TDD 证据（RED 摘录）
- `expected: 80006, actual: 0` + `Expected nil, but got: &communityv1.Notice{...CommunityId:200...}`
  —— 越权 Get 原本返回内容（GetNotice 4 个拒绝用例全 FAIL）。
- `Should be true / 出站 metadata 必须存在`、`An error is expected but got nil` —— API 未注入身份/未透出 080006。
- `Should be true / RPC 必须绑定回环（127.0.0.1/localhost），当前 host="0.0.0.0"` —— 配置不变式。

### 影响
- 配置：rpc/etc/communityhub.yaml ListenOn 0.0.0.0 → 127.0.0.1（脚本端口探测 netstat `:8088` 仍匹配，无破坏）
- 兼容：RPC/API 请求响应契约未变（身份仍经 gRPC metadata）；Get-by-ID 越权时由 200 空 → 080006 错误，属预期修复
- 依赖：无新增；复用 scope.FilterAllowed / scope.DenyBase
- 门禁：harness-checks 16 PASS / 0 FAIL / 0 WARN；`go build ./...` + `go test ./...`（含 -count=1 全量）全绿

---

## 2026-08-12 — 数据权限消费方（access-data-permission 阶段④ T4.0-T4.8，TDD）

### 做了什么
- **T4.0 身份注入通道**：新增 `api/internal/util/userctx.go`（`JWTUserID(ctx) (int64,error)` 提取 JWT、`WithUserID` 注入出站 gRPC metadata）；`ServiceContext.CallCtx` 统一封装「提取 JWT → 注入 metadata」，供所有 API→RPC 调用使用
- **T4.1 publisher_id 规范化**：CreateLostFound/CreateNotice API 逻辑用 JWT user_id 覆盖 gRPC 请求 publisher_id，忽略客户端 body 值（防伪造）
- **T4.2-4.4 AssertPublishScope 挂载（写接口）**：新增 `rpc/internal/logic/scope` 包（`AssertCommunityScope` / `CheckPublishScope` / `CheckSystemPublishScope` / `FilterAllowed`）；RPC 层落库前对目标 community_id 校验，`allowed=false` → 080006，映射 permission 060007 → 080006；覆盖 CreateLostFound/ResolveLostFound/CreateNotice/UpdateNotice/DeleteNotice/UpsertContacts
- **T4.5 moderation 回调身份校验（S4）**：UpdateNotice/UpdateLostFoundModerationStatus 先 reverse-lookup 内容 community_id（查不到拒绝），以系统身份（system_user_id=0 常量，global scope）调 AssertPublishScope（服务身份回调放行，不按作者 scope）
- **T4.6 读列表按 GetDataScopes 过滤**：新增 `rpc/internal/logic/scope/filter.go`；ListNotices/ListLostFound/ListContacts 按 GLOBAL 不过滤 / LIMITED IN(ids) / EMPTY 空列表过滤（空列表在逻辑层返回，不拼空 IN）；API 列表逻辑同步注入 metadata
- **T4.7 错误码 080006 注册**：`api/internal/types/types.go` 登记 08xxxx 常量（080002 功能权限 / 080006 数据权限分层语义）；`rpc/internal/logic/scope.CodePublishScopeDenied=80006`
- RPC 配置新增 `PermissionRpc`（communityhub.yaml + config.go + servicecontext.go 挂 permission 客户端）

### 关键设计决策
- **身份链路**：JWT（rest.WithJwt 注入 ctx）→ `JWTUserID` 提取 → `WithUserID` 注入出站 gRPC metadata → RPC 层 `UserIDFromCtx` 读取，用于 AssertPublishScope/GetDataScopes。不信任客户端 body 的 publisher_id/userId
- **校验顺序**：功能权限（PermMiddleware，中间件链先于 handler）→ 数据权限（RPC 落库前）→ 落库
- **系统身份**：system_user_id=0 是合法 global 身份（moderation 回调），走同一 grant 判定路径，无代码级短路；与 CheckPublishScope 的「userID==0 拒绝」分支区分（经 CheckSystemPublishScope 直接 AssertCommunityScope）
- **fail-closed**：读过滤 userID==0 恒拒绝（0 是系统身份，禁止用户读路径借用）；写接口无身份 → 080006
- API 写逻辑（Resolve/Update/Delete/Upsert）同步补 `responsex.ToError(resp.GetBase())`，使 RPC 业务错误（080006）透出客户端（此前忽略 Base.Code 的既有缺陷在本变更触及处修复）

### TDD 证据
- RED 摘录示例：`expected: 80006, actual: 0` / `expected: 1001, actual: 999999`（伪造 publisher_id 未被覆盖）/ `undefined: JWTUserID`（新函数编译期 RED）
- 新增测试文件（11 个）：util/userctx_test、svc/servicecontext_test、api logic（lostfound/notice create、contact upsert）、rpc logic（lostfound publishscope、notice publishscope、notice updatemoderationstatus、contact upsert、scope scope/filter、notice listnotices_filter）

### 影响
- 配置：rpc/etc/communityhub.yaml 新增 `PermissionRpc`（permission.rpc）
- 依赖：无新外部模块；复用 api-proto permission/v1 生成代码（AssertPublishScope/GetDataScopes，契约已提交 031f4e4+c245c09）
- 兼容：不触碰 api-proto/ 与 common/；RPC 写/读接口请求参数与响应契约未变（身份经 gRPC metadata 传递，非 proto 字段）
- 门禁：harness-checks 16 PASS / 0 FAIL / 0 WARN；`go build ./...` + `go test ./...` 全绿

---

## 2026-06-06 — 服务初始化

### 做了什么
- 创建 community-hub-service 微服务，实现社区枢纽功能
- 实现 RPC + REST API 双层架构（端口 8087/8887）
- 实现 NoticeService（通知公告 CRUD + 软删除）
- 实现 ContactService（便民联络列表 + 批量更新）
- 实现 LostFoundService（寻失互助 CRUD + 标记解决）
- 创建 4 张数据库表（notices, notice_attachments, community_contacts, lost_found_items）
- 使用 go-zero sqlx 风格数据模型（4 个 Model）
- 所有 int64 ID 使用 json:",string" 标签
- 使用 configx.MustLoad 加载配置（支持 ${ENV_VAR}）
- 使用 Snowflake 生成分布式唯一 ID

### 为什么
社区平台需要小区信息汇聚中心，提供通知公告、便民联络、寻失互助等社区内容场景

### 影响
- Proto: api-proto/api/community/v1/community.proto（已定义）
- 调用方: 无（新服务，暂无外部 gRPC 调用方）
- 数据库: 新增 community_hub_db 库，4 张表
- 关联: go.work 已添加 services/community-hub-service
