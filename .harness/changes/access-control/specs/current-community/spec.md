# Current Community（用户应用状态·当前小区）Specification

## Purpose

定义「当前小区」这一用户应用状态的权威持久化：用户在多小区身份下，系统需要记住其默认上下文（首页展示、发布默认目标、跨设备一致），并允许用户切换，切换时校验目标小区属于该用户的数据范围。

## Requirements

### Requirement: 当前小区服务端持久化
系统 SHALL 在服务端按账号持久化「当前小区」（账号级、跨设备一致），每个用户至多一个当前小区值；当用户尚未加入任何小区时，当前小区 SHALL 为空。

#### Scenario: 记录当前小区
- **GIVEN** 用户已加入小区 A、B
- **WHEN** 系统读取该用户的当前小区
- **THEN** 返回一个明确的当前小区 id（或空，若尚未设置），且跨设备读取结果一致

#### Scenario: 未加入任何小区
- **GIVEN** 用户未加入任何小区（无数据范围）
- **WHEN** 系统读取该用户的当前小区
- **THEN** 返回空，不报错

### Requirement: 切换当前小区（范围校验）
系统 SHALL 提供切换当前小区的接口；切换时 MUST 校验目标小区属于该用户的数据范围，范围外一律拒绝。

#### Scenario: 切到已加入小区
- **GIVEN** 用户已加入小区 A、B，当前小区为 A
- **WHEN** 用户请求切换到小区 B
- **THEN** 系统更新当前小区为 B，并跨设备立即生效

#### Scenario: 切到未加入小区
- **GIVEN** 用户仅加入小区 A，未加入小区 C
- **WHEN** 用户请求切换到小区 C
- **THEN** 系统拒绝切换，当前小区保持不变

#### Scenario: 注册用户（空数据范围）切换
- **GIVEN** 用户为 `registered_user` 基角色，无任何数据范围
- **WHEN** 用户请求切换到任意小区
- **THEN** 系统拒绝（目标小区不在其数据范围内）

#### Scenario: global 角色切换
- **GIVEN** 用户持 global 数据范围（如审核员）
- **WHEN** 用户请求切换到任意小区
- **THEN** 系统放行（global 覆盖所有小区）

### Requirement: 当前小区读取接口
系统 SHALL 提供读取当前小区应用状态的接口，返回当前小区 id（可为空）及其更新时间，供首页/发布默认上下文使用。

#### Scenario: 读取应用状态
- **GIVEN** 用户已设置当前小区为 A
- **WHEN** 用户请求读取应用状态
- **THEN** 系统返回 `{ current_community_id: A, updated_at: <时间> }`

#### Scenario: 读取未设置状态
- **GIVEN** 用户尚未设置当前小区
- **WHEN** 用户请求读取应用状态
- **THEN** 系统返回 `current_community_id` 为空

### Requirement: 新接口权限注册
系统 SHALL 为 `app-state` 与 `current-community` 两个新接口注册权限码，并纳入自动发现。

#### Scenario: 新接口可被鉴权
- **GIVEN** `app-state` / `current-community` 接口已注册权限码
- **WHEN** 用户请求访问这两个接口
- **THEN** 系统按注册的权限码执行功能权限校验，未授权者拒绝

> ✅ 已定（CLAR-2，选项 a）：`user_app_state.current_community_id` 取代 `user_base.preferences.default_community_id`。开发阶段无存量数据，无需迁移；`preferences` 字段去留由阶段3定。

## 职责边界

- **user-service**：`user_app_state` 的存储、读取、切换校验（切换校验目标小区 ∈ 数据范围，经 permission-service GetDataScopes）。
- **permission-service**：提供数据范围（GetDataScopes）供切换校验消费。
- **新接口权限码**：`app-state` / `current-community` 注册权限码并纳入自动发现。
- **web/mobile 前端**：当前小区展示与切换入口，首页默认上下文。

## 范围外（本次不实现）

- **community-hub 消费侧（读当前小区作发布默认上下文）**：本次不实现，留待后续独立跟进。本次仅交付 user-service 的持久化 + 接口 + 切换校验。
