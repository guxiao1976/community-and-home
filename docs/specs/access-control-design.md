# 访问控制与数据权限设计方案

> 最后更新：2026-08-12
> 定位：`rbac-design.md` 的姊妹篇，覆盖**角色分层与认证**、**端限制（登录准入）**、**数据权限（发布范围）**、**用户应用状态（当前小区）**、**发布配额与成员约束**。
> 状态：设计稿（评审中）。已确认项无 ⚠️ 标记；剩余待确认项以 ⚠️ 标注。

---

## 1. 概述

### 1.1 背景

- 系统分 PC 端与移动端：PC 端给系统管理员、主数据审核员、内容审核员等管理人员；业主、租户、网格员等绝大多数角色用移动端。
- 系统面向**全国小区**，业主/租户是最大群体，**管理操作不能依赖人工审核**（否则管理员忙不过来），认证等审核未来走 AI 自动化，人工仅兜底。
- 业主可加入多个小区，"当前小区"唯一；发布只能发到自己所在的小区。
- 未来：商户通过 App 注册 → 认证为商户 → 选范围购买 → 支付完成后形成"范围 + 期限"广告订单。**商户广告本次不实现，权限模型需与数据权限统一。**

### 1.2 核心原则

1. **后端权威**：凡是"你是谁 / 你在哪 / 你能做什么"，后端必须是权威，客户端传值只能当提示、不能当事实（`device_type`、`community_id`、`publisher_id` 均不可信）。
2. **四层正交**：功能权限（能否调接口）→ 数据权限（能发到哪）→ 配额（能发多少）→ 应用状态（默认上下文），独立叠加。
3. **配置驱动**：角色→端、房屋上限、板块配额等映射由配置决定，不硬编码。
4. **认证自动化**：业主/租户最大的群体，加入、发布不依赖人工审批；认证（房产证）面向 AI 自动化，人工兜底。

### 1.3 范围

| 本次实现 | 未来实现（模型已兼容，不实现） |
|----------|------------------------------|
| 角色分层：注册用户 / 未认证业主·租户 / 认证业主 | 商户广告投放（注册→认证→选范围购买→支付→订单） |
| 加入小区 = 自动授权；退出 = 撤销 | 广告订单授权（范围+期限，订单=一次投放） |
| 端限制：角色→端 配置驱动登录准入 | 街道/区县/市级层级授权 + 小区广告上限 10（可配置） |
| 数据权限：发布范围校验（统一祖先链模型，community 级） | 订单过期提醒 + 自动下架 |
| 用户应用状态：当前小区服务端持久化 | 认证审核 AI 自动化（OCR+比对，人工兜底） |
| 板块发布配额 + 房屋注册上限 + 同屋互见 | |

### 1.4 角色分层状态机（本次核心）

```
注册 → 注册用户(registered_user, 基角色)
          ├─ 可浏览部分内容；不可发布；无数据范围（scope=空）
          │
          │  加入小区 + 注册房屋（自有→owner / 租住→tenant）
          ▼
        未认证业主/租户
          ├─ 自动获得：owner/tenant 角色 + 该小区 scope（未认证状态）
          ├─ 可发布（配额限制内）；不可参与业委会选举
          │
          │  上传房产证 → 认证审核（未来 AI 自动化）
          ▼
        认证业主
          ├─ 解锁：业委会选举等高信任能力
          │
          │  退出小区
          ▼
        撤销该小区角色 + scope（回到注册用户层级）
```

**关键规则：加入小区即自动数据授权，退出即撤销；认证状态是"能力分层"阶梯，不是"能不能用"开关。**

---

## 2. 四层权限模型 + 成员约束

| 层 | 管什么 | 机制 | 归属 |
|----|--------|------|------|
| 功能权限 | 能不能调这个接口 | `PermMiddleware` + `CheckPermission` | 已有 |
| 数据权限 | 能发到**哪**（范围） | 祖先链命中规则（§5） | permission-service 算，消费方校验 |
| 发布配额 | 能发**多少**（数量） | 计数 ≤ 上限（§7） | master-data 配，community-hub 执行 |
| 应用状态 | **默认上下文**（当前小区） | 服务端持久化 + 切换校验（§6） | user-service |

**成员约束**（不属于四层，是成员域独立规则，见 §3.5）：

| 约束 | 规则 | 归属 |
|------|------|------|
| 加入小区次数 | 同时 ≤3、每年新加 ≤3、终身 ≤12（可配置） | user-service 执行 |
| 房屋注册人数 | 每户 ≤6（可配置） | user-service 执行 |

---

## 3. 角色与认证

### 3.1 注册用户（基角色）

- `sys_role` 新增 `registered_user`（**正式角色**），权限为 browse 类，配置在 `rel_role_permission`。
- 用户注册时（user-service CreateUser 流程）自动分配，`status=2`（永久有效）。
- **数据范围 = 空**（无任何小区 scope），因此不可发布、不可见小区内部数据。
- 权限叠加：注册用户权限 ∪ 社区角色权限。

### 3.2 加入小区 = 自动授权

- `JoinCommunity`（user-service）成功 → 同步调 permission-service `AssignRole`：
  - 角色：`owner`（自有）/ `tenant`（租住）——由房屋注册时选择的"自有/租住"决定
  - scope：`community` + 该小区 id
  - status：`0 未认证`（未认证业主/租户，即可发布）
  - **幂等**：已存在同角色同 scope 则不动
- `LeaveCommunity` 成功 → 同步调 permission-service 撤销该小区角色/scope。
- 自动授权使 membership 与 scope 天然一致，无需人工审批。

### 3.3 认证（房产证）

- 未认证业主/租户上传房产证 → 认证审核 → `status=2 已认证`。
- 审核未来走 **AI 自动化**（OCR + 不动产权比对，接 ai-model-service / moderation-service），人工仅兜底/异常。
- 审核期间（待审）保持未认证能力（可发布），仅高信任能力（选举）不开放。

### 3.4 能力分层（修订 rbac-design.md §2.5）

> 现状 rbac-design §2.5 规定 status 0/1 不能鉴权、status 2 才可——**本规范修订为"能力分层"**：

- `sys_permission` 增加 `min_verf_level` 属性：
  - `0`：持有角色 + 数据范围即可（如发布类：`lostfound:create`）
  - `2`：需已认证（如选举类：`committee:election:vote`）
- `CheckPermission`：未认证用户可执行 `min_verf_level=0` 的权限，`=2` 的拒绝；已认证全部放行。

| 能力 | 注册用户 | 未认证业主/租户 | 认证业主 |
|------|:---:|:---:|:---:|
| 浏览 | ✅ | ✅ | ✅ |
| 发布（配额内） | ❌ | ✅ | ✅ |
| 参与业委会选举 | ❌ | ❌ | ✅ |

### 3.5 反滥用与房屋约束（user-service 执行）

| 约束 | 规则 | 配置键（master-data sys_config） |
|------|------|------|
| 同时加入小区数 | ≤3 | `user.max_community_join_count` |
| 每年新加入 | ≤3（仅非认证用户受限） | `user.max_new_communities_per_year` |
| 终身加入 | ≤12 | `user.max_total_communities_lifetime` |
| **每户注册人数** | **≤6（业主+租户合计）** | `user.max_house_members` |

- 房屋人数校验（JoinCommunity）：`count(active membership WHERE community_id+building+unit+room) >= 6` → 拒绝（错误码 10014 "该房屋已满员"），**计数排除当前用户、只算 active**。
- 目的：配合 §5.7 同屋互见，防止反复退出/重加入刷"未认证"身份发布内容。

---

## 4. 端限制（登录准入）

> 目的：移动端角色不应在 PC 端获得无意义体验；**不作为安全边界**（真正安全由后端 RBAC 兜底）。

### 4.1 配置模型

- `sys_role` 增加 `platforms` 属性（`[pc]` / `[mobile]` / `[pc,mobile]`），`sys_admin` 维护。
- 默认：PC=`sys_admin`/`community_admin`/`property_admin`/审核员类；移动=`owner`/`tenant`/`grid_worker`/`committee`/`merchant`/`registered_user`。
- 语义区分（STAGE3-4 定稿）：上述「默认」是**种子初值**（`init_permissions.sql` 显式写入 9 个内置角色）；运行时回退为 **fail-open**——`platforms` 为空的角色视为允许所有端。

### 4.2 判定

```
登录/刷新：存在任一角色 platforms 含当前端 → 放行
          否则 → 50007 "该账号为移动端用户，请使用移动端 APP"
```
- 多角色用户：任一角色允许即可；`RefreshToken` 同规则校验。
- 端归类：`web`/`admin`→PC；`ios`/`android`/`miniapp`→移动端。

---

## 5. 数据权限（统一模型）

### 5.1 scope 三态

| scope 状态 | 含义 | 适用 |
|-----------|------|------|
| `global` | 放行全部数据范围 | 超管、审核员 |
| `community` 等限定 | 限定到具体范围节点 | 业主/租户（加入自动授权）、未来广告订单 |
| **空** | 无数据范围 | `registered_user` 基角色 |

> **禁止**把"空"当成 `global`——注册用户一旦被赋 global，数据权限会放行全国数据，灾难。

### 5.2 统一规则：祖先链命中

覆盖树 = 行政区划树（省→市→区县→街道→社区 division 节点）+ 小区（`residential_area`）叶子。

```
目标 t 被授权集合 S 覆盖 ⟺ A(t) ∩ S ≠ ∅   （A(t) = t 及其祖先，≤6 节点）
一次发布放行 ⟺ 目标集全部被覆盖
```
实现：目标祖先链（短、固定）逐个哈希判断，O(6)，与授权规模/城市大小解耦。

### 5.3 覆盖树与祖先链解析

- 行政区划已有 `parent_id`/`path`；小区经 `community_div_id` 入树。
- **master-data 提供"任意 scope 节点 → 祖先链"能力**（RPC/缓存）。行政区划低频变更，整树缓存。

### 5.4 授权来源（可插拔）

| 场景 | 授权来源 | 粒度 | 状态 |
|------|---------|------|------|
| 业主/租户发布 | 成员资格（加入自动授权，§3.2） | 小区 | 本次实现 |
| 商户广告 | 广告订单（范围+期限） | 小区/街道/区县/市 | 未来 |
| 超管/审核员 | global | 全部 | 本次实现 |

### 5.5 写操作校验 `AssertPublishScope`

```go
AssertPublishScope(ctx, userId, targets []ScopeRef) error
// ① global 放行
// ② 逐 target: 祖先链 ∩ 授权节点集合 ≠ ∅ ? 继续 : 返回 80006
// ③ 全部通过 → 放行
```
- `publisher_id` 一律取自 JWT；API→RPC 全程透传认证身份。
- 挂载：community-hub 所有写接口（lostfound、notices、contact 等）。

### 5.6 读操作过滤

- 读列表按用户可见小区过滤：`WHERE community_id IN (GetDataScopes(user_id, "community"))`。
- ⚠️ 隔离强度默认：写严格校验；读按小区清单过滤。

### 5.7 同屋互见（户级数据可见性）

- **规则**：用户 A 查看用户 B 详情时，若 A、B 存在**同屋**（同一小区 community+building+unit+room 的 active membership），则 B 对 A 可见：**手机号 + 楼栋房屋号**；否则手机号脱敏。
- 系统**不采集**身份证号、真实姓名，无此类隐私面；租户也参与互见（同屋 6 人互相可见）。
- **归属**：user-service 用户详情接口判定；退出房屋 → 互见关系随之消失（基于 active membership）。
- 目的：同屋人互相监督，防冒领房屋/冒充。

### 5.8 global 例外

`scope_type=global` 或持 `sys_admin` → 数据权限放行（跨小区查看/审核）。

### 5.9 缓存与失效

- 授权节点集合：复用 `perm:scopes:{userId}:{scopeType}` 模式；角色/scope 变更 DEL（加入/退出/认证时）。
- 覆盖树（division）：master-data 缓存。

---

## 6. 用户应用状态（当前小区）

- `user_app_state(user_id PK, current_community_id, updated_at)`，归属 **user-service**（账号级，跨设备一致）。
- API：`GET /api/users/me/app-state`；`PUT /api/users/me/current-community`（**切换校验目标小区 ∈ 数据范围**）。
- 缓存 Redis `user:appstate:{userId}`；用途：首页/发布默认上下文、跨设备一致。

---

## 7. 板块发布配额

- 配置：`sys_section_quota(section_type, max_count)`（master-data），如 `lost_found=5`、`second_hand=5`，上限按板块可配置。
- 校验（community-hub，在 `AssertPublishScope` 之后）：`count(占配额内容 WHERE user_id+community_id+section_type) >= 上限 → 80007`。
- **占配额状态定义**（内容有两条状态线：业务 `status` + 审核 `moderation_status`）：

  | 内容状态 | 占配额 | 理由 |
  |---------|:---:|------|
  | 待审（`moderation_status=0`） | ✅ | 防刷：否则"发→删→重发"无限堆积审核队列 |
  | 展示中（通过，`status=active`） | ✅ | 正在占用版面 |
  | 已驳回（`moderation_status=2`） | ❌ | 系统判无效、不可见即释放 |
  | 已解决（`status=resolved`） | ❌ | 使命完成，可再发新的 |
  | 已删除（`deleted_at IS NOT NULL`） | ❌ | 内容不存在 |
  | 下架/移除 | ❌ | 不可见即释放；惩罚走独立机制不混入配额 |

  计数条件：`deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)`。
  > 状态机说明（STAGE3-2 定稿）：`status`（业务态）与 `moderation_status`（审核态）正交；内容创建即 `status='active'`（业务在版），待审期间 `moderation_status=0` 但 `status` 仍为 active，故待审与展示同占配额；下架/移除复用「`status` 非 active」语义。
- 口径：**个人**（用户×小区×板块），按**目标小区**计（非"当前小区"）；适用所有发布者。
- 索引：`user_id+community_id+section_type+status`。

---

## 8. 服务归属与依赖

| 能力 | 归属服务 |
|------|---------|
| 角色、`platforms`、权限码、`registered_user` | permission-service |
| `rel_user_role` 自动授权/撤销 | user-service 编排 → permission-service 执行 |
| 认证状态能力分层 | permission-service（CheckPermission/GetDataScopes + `min_verf_level`） |
| 端限制判定 | auth-service |
| 房屋注册上限、同屋互见、当前小区 | user-service |
| 数据权限授权集 | permission-service |
| 覆盖树/祖先链解析 | master-data-service |
| 发布校验 + 配额执行 | community-hub-service |
| 所有可配置上限（sys_config） | master-data-service |
| 认证审核（AI 未来） | user-service（材料）+ ai-model/moderation（AI）+ permission-service（状态） |

**新增依赖**：
```
mobile ──加入/退出小区──▶ user-service ──Assign/RevokeRole──▶ permission-service
mobile ──切换小区──▶ user-service（校验 scope）
mobile ──发布──▶ community-hub（功能权限→数据权限→配额→落库→审核）
community-hub ──读当前小区──▶ user-service
community-hub ──读配额──▶ master-data
community-hub ──AssertPublishScope──▶ permission-service ──ResolveScopeAncestors──▶ master-data
user-service ──校验 scope──▶ permission-service（已有）
auth-service ──角色+platforms──▶ permission-service（已有链路扩展）
```

---

## 9. 接口权限确认

- community-hub 写接口确认已挂 `PermMiddleware`+`WithJwt`。
- 新接口（`app-state`、`current-community`）注册权限码并纳入自动发现。
- `sys_permission` 新增 `min_verf_level` 字段，随权限码维护。

---

## 10. 未来扩展

- **商户广告**：订单（范围+期限，订单=一次投放）作为新授权来源接入 §5.4；订单内粒度统一（不混合街道/县区）；小区广告上限 10 = §7 配额复用；订单过期提醒/自动下架 ≈ `expires_at` 懒校验 + 定时扫描。⚠️ 注意：现网 merchant 角色 scope 为 global，做商户广告时需改为非全局或授权判定只认订单。
- **AI 认证**：房产证 OCR + 比对，接入 ai-model-service/moderation-service，人工兜底。

---

## 11. 验收矩阵

### 11.1 角色分层

| 场景 | 期望 |
|------|------|
| 注册用户（无小区）发布 | ❌ 拒绝（无数据范围） |
| 未认证业主发布（配额内） | ✅ 通过 |
| 未认证业主参与业委会选举 | ❌ 拒绝（`min_verf_level=2`） |
| 认证业主参与选举 | ✅ 通过 |
| 认证期间（待审）发布 | ✅ 保持未认证能力 |

### 11.2 加入/退出自动授权 + 成员约束

| 场景 | 期望 |
|------|------|
| 加入小区 → rel_user_role 自动出现 owner/tenant + scope | ✅ |
| 退出小区 → 该角色/scope 撤销 | ✅ |
| 房屋第 7 人注册 | ❌ 10014 "该房屋已满员" |
| 第 4 个小区加入 | ❌ 10006 |
| 反复退出重加入刷未认证身份 | ❌ 每年/终身次数限制 |

### 11.3 端限制

| 场景 | 期望 |
|------|------|
| owner 在 PC 登录 | ❌ 50007 引导移动端 |
| owner+community_admin 在 PC | ✅（任一角色允许） |
| RT 刷新 PC 会话 | ❌ 同规则拦截 |

### 11.4 数据权限 + 配额

| 场景 | 期望 |
|------|------|
| owner@A 在 B（∉ scope）发布 | ❌ 80006 |
| 抓包改 `publisher_id` 伪装他人 | ❌ 无效（取自 JWT） |
| 审核员（global）跨小区查看 | ✅ |
| 板块已达 5 条再发 | ❌ 80007；驳回/删除后释放 |
| 同屋用户查看对方手机号 | ✅ 可见 |
| 非同屋用户查看手机号 | ❌ 脱敏 |
| 退出 B 后立刻在 B 发布 | ❌（scope 缓存 DEL 生效） |

### 11.5 当前小区

| 场景 | 期望 |
|------|------|
| 切到未加入小区 | ❌ 拒绝 |
| 切到已加入小区 | ✅ 更新，跨设备生效 |

---

## 12. 决策记录（已全部确认）

| # | 决策 | 状态 |
|---|------|------|
| 1 | 读隔离强度 | ✅ 写严格；读按小区清单过滤 |
| 2 | 配额有效状态口径 | ✅ 见 §7 明细（待审/展示占，驳回/解决/删除/下架释放） |
| 3 | 存量 scope 数据 | ✅ 不考虑（开发阶段，未上线） |
| 4 | 同屋互见字段 | ✅ 手机号 + 楼栋房屋号 |

---

## 13. 相关文件索引

| 类别 | 文件 |
|------|------|
| RBAC 基础 | `docs/specs/rbac-design.md`（§2.5 鉴权规则被本规范 §3.4 修订） |
| 数据范围 RPC | `api-proto/api/permission/v1/permission.proto`（GetDataScopes） |
| 登录/端 | `api-proto/api/auth/v1/auth.proto`（device_type） |
| 行政区划/小区 | `api-proto/api/masterdata/v1/masterdata.proto` |
| 房屋/成员 | `api-proto/api/user/v1/user.proto`（CommunityMembership/JoinCommunity）、`services/user-service/rpc/internal/logic/user/join_community_logic.go` |
| 内容发布 | `services/community-hub-service/rpc/internal/logic/lostfound/createlostfoundlogic.go` |
| 系统配置 | `services/master-data-service/migration/003_system_config_refactor.sql` |
