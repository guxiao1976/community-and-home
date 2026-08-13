# Platform Restriction（端限制·登录准入）Specification

## Purpose

定义角色与登录端的映射约束：移动端角色（业主、租户、网格员、业委会、商家、注册用户等）不应在 PC 端获得无意义体验，登录/刷新时据此引导用户使用正确端。本能力为 **UX 引导而非安全边界**，真正的安全由后端 RBAC 与数据权限兜底。

## Requirements

### Requirement: 角色声明允许的端（platforms）
每个角色 SHALL 声明其允许登录的端集合（`platforms`），取值可为 `pc`、`mobile` 或 `pc,mobile`，由 `sys_admin` 维护；未声明平台的角色 MUST 视为允许所有端（默认不拦截）。

#### Scenario: 角色配置了端
- **GIVEN** `owner` 角色的 `platforms=[mobile]`，`sys_admin` 的 `platforms=[pc]`，`community_admin` 的 `platforms=[pc,mobile]`
- **WHEN** 系统读取这些角色的平台声明
- **THEN** `owner` 仅移动端、`sys_admin` 仅 PC、`community_admin` 双端均可

#### Scenario: 角色未声明平台
- **GIVEN** 某角色未配置 `platforms` 字段
- **WHEN** 系统判定该角色是否允许某端
- **THEN** 该角色被视为允许所有端，不触发端限制拦截

### Requirement: 登录准入判定（任一角色允许即放行）
系统 SHALL 在登录成功签发 Token 前，按当前请求的 `device_type` 归类为 PC 或移动端；若用户持有的任一角色 `platforms` 包含当前端，则放行；否则拒绝并返回明确引导（语义：该账号为移动端用户，请使用移动端 APP）。

#### Scenario: 移动端角色登录移动端
- **GIVEN** 用户仅持有 `owner`（`platforms=[mobile]`）角色
- **WHEN** 用户以 `device_type=android` 登录
- **THEN** 系统放行，正常签发 Token

#### Scenario: 移动端角色登录 PC
- **GIVEN** 用户仅持有 `owner`（`platforms=[mobile]`）角色
- **WHEN** 用户以 `device_type=web` 登录
- **THEN** 系统拒绝并返回「该账号为移动端用户，请使用移动端 APP」的引导错误（参考错误码 050006，最终码阶段3对齐）

#### Scenario: 多角色用户跨端
- **GIVEN** 用户同时持有 `owner`（`platforms=[mobile]`）与 `community_admin`（`platforms=[pc,mobile]`）
- **WHEN** 用户以 `device_type=web` 登录
- **THEN** 因 `community_admin` 允许 PC，系统放行

#### Scenario: 端归类映射
- **GIVEN** 端归类规则为 `web`/`admin`→PC，`ios`/`android`/`miniapp`→移动端
- **WHEN** 系统收到 `device_type` 为上述任一值
- **THEN** 系统按规则将其归类到对应端进行判定

### Requirement: Token 刷新沿用端限制
系统 SHALL 在刷新 Token 时应用与登录一致的端限制判定；RefreshTokenRequest SHALL 携带 `device_type`，刷新时按该 `device_type` 归类并执行端准入。

#### Scenario: 移动端角色刷新 PC 会话
- **GIVEN** 用户仅持有 `owner`（`platforms=[mobile]`），持有一个 PC 端会话的 RT
- **WHEN** 用户以 `device_type=web` 的 RefreshTokenRequest 刷新
- **THEN** 系统拒绝刷新，返回与登录一致的端引导错误

#### Scenario: 移动端角色刷新移动端会话
- **GIVEN** 用户仅持有 `owner`（`platforms=[mobile]`）
- **WHEN** 用户以 `device_type=android` 的 RefreshTokenRequest 刷新
- **THEN** 系统放行，签发新 Token 对

> ✅ 已定（CLAR-1，选项 a）：RefreshTokenRequest 增 device_type，刷新时按与登录一致的端归类判定。

## 职责边界

- **auth-service**：登录/刷新时执行端准入判定（读取角色 `platforms` + 端归类）。
- **permission-service**：`sys_role.platforms` 的存储与透出（角色查询、登录链路 GetUserRoles）。
- **web/mobile 前端**：接收引导错误后展示「请使用移动端 APP」提示。
