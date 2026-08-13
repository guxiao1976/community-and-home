# Section Publish Quota（板块发布配额）Specification

## Purpose

定义板块级发布配额：限制单一用户在单一小区、单一板块内同时占用的发布数量，防止刷版与审核队列无限堆积。配额上限按板块可配置，计数遵循明确的「占配额」口径。

## Requirements

### Requirement: 板块配额可配置
系统 SHALL 按板块（`section_type`）配置发布数量上限（`max_count`），上限值由配置决定、不硬编码；未配置上限的板块 MUST 视为不限量。

#### Scenario: 读取板块配额
- **GIVEN** 配置中 `lost_found` 上限为 5
- **WHEN** 系统校验用户在 `lost_found` 板块的发布
- **THEN** 以 5 为该板块上限进行计数判定

#### Scenario: 板块未配置上限
- **GIVEN** 某板块在配置中无对应上限记录
- **WHEN** 系统校验该板块的发布
- **THEN** 不触发配额拦截（视为不限量）

### Requirement: 发布配额校验
系统 SHALL 在内容发布（数据权限校验 `AssertPublishScope` 之后、落库之前）执行配额校验：若该用户在目标小区、该板块内的「占配额内容」数量已达到上限，则拒绝发布并返回语义唯一的「超出配额」错误（参考错误码 80007，最终码阶段3对齐）。

#### Scenario: 未达上限发布
- **GIVEN** 用户在小区 A 的 `lost_found` 板块已有 4 条占配额内容，上限 5
- **WHEN** 用户发布第 5 条
- **THEN** 系统放行，落库成功

#### Scenario: 达上限发布
- **GIVEN** 用户在小区 A 的 `lost_found` 板块已有 5 条占配额内容，上限 5
- **WHEN** 用户发布第 6 条
- **THEN** 系统拒绝并返回「超出配额」错误

#### Scenario: 按目标小区计数（非当前小区）
- **GIVEN** 用户当前小区为 A，在小区 B 的板块已占满配额（用户同时为 B 成员）
- **WHEN** 用户向小区 B 的该板块发布
- **THEN** 系统按目标小区 B 的计数拒绝；反之按 A 计数独立判定，互不影响

### Requirement: 占配额判定（唯一计数谓词）
系统 SHALL 按唯一谓词判定内容是否占配额：**未删除（`deleted_at IS NULL`）且 `moderation_status` 为待审(0)或通过(1) 且 `status` 为 `active`** 的内容占配额；其余状态（已驳回、已解决、下架/移除、已删除）均释放配额。

#### Scenario: 待审内容占配额
- **GIVEN** 用户发布一条内容，处于待审（`moderation_status=0`）
- **WHEN** 系统统计该用户该板块的占配额数量
- **THEN** 该内容计入配额（防「发→删→重发」无限堆积审核队列）

#### Scenario: 展示中内容占配额
- **GIVEN** 用户内容审核通过且处于展示中（`moderation_status=1` 且 `status=active`）
- **WHEN** 系统统计占配额数量
- **THEN** 该内容计入配额（正在占用版面）

#### Scenario: 已驳回释放配额
- **GIVEN** 用户某内容被驳回（`moderation_status=2`）
- **WHEN** 系统统计占配额数量
- **THEN** 该条不计入配额，用户可再发布 1 条

#### Scenario: 已解决/下架/移除释放配额
- **GIVEN** 用户某内容 `status` 由 `active` 转为非 `active`（已解决 `resolved`；下架/移除复用同一「`status` 非 `active`」语义）
- **WHEN** 系统统计占配额数量
- **THEN** 该条不计入配额，释放对应额度

#### Scenario: 已删除释放配额
- **GIVEN** 用户某内容被删除（`deleted_at` 非空）
- **WHEN** 系统统计占配额数量
- **THEN** 该条不计入配额，释放对应额度

> [STAGE3] 唯一计数谓词 `deleted_at IS NULL AND moderation_status IN (0,1) AND status='active'` 与 design §7 原文公式存在张力：design §7 写「计数条件：deleted_at IS NULL AND status='active' AND moderation_status IN (0,1)」，但「待审内容占配额」要求待审（moderation_status=0）内容即便尚未展示也必须计入——须保证待审内容创建时 `status` 即为 `active`，或明确「待审」在计数谓词中的落入方式。阶段3 需统一内容状态机，并在 tasks.md 记录「阶段3 修正 design §7 计数公式」。本 spec 以本节「唯一计数谓词」为行为契约。
>
> [STAGE3] 现网 lost_found 的 `status` 仅 `active`/`resolved` 两态，无独立「下架」状态；若未来引入下架，须复用「`status` 非 `active`」语义纳入释放口径。

### Requirement: 配额计数口径为个人 × 小区 × 板块
系统 SHALL 按「用户 × 目标小区 × 板块」三维口径独立计数，适用所有发布者。

#### Scenario: 不同板块独立计数
- **GIVEN** 用户在小区 A 的 `lost_found` 已满额，但 `second_hand` 未满额
- **WHEN** 用户向 `second_hand` 发布
- **THEN** 系统按 `second_hand` 板块独立计数，放行（若未满额）

#### Scenario: 不同小区独立计数
- **GIVEN** 用户在小区 A 板块已满额，小区 B（同为成员）板块未满额
- **WHEN** 用户向小区 B 同板块发布
- **THEN** 系统按小区 B 独立计数，放行（若未满额）

#### Scenario: 不同用户配额互不影响（个人维度）
- **GIVEN** 用户 X 在小区 A 的 `lost_found` 板块已占满配额（5/5），用户 Y 在同一小区同一板块为 0 条
- **WHEN** 用户 Y 向小区 A 的 `lost_found` 发布
- **THEN** 系统按 Y 个人的配额独立计数，Y 放行（X 的满额不影响 Y）

> ✅ 已定（CLAR-3，选项 c）：以 sys_section_quota 配置为权威，配置了限制、未配置不限；管理员/官方发布默认不配置。

## 职责边界

- **master-data-service**：`sys_section_quota(section_type, max_count)` 配置存储与读取。
- **community-hub-service**：在 `AssertPublishScope` 之后执行配额计数与超限拒绝（消费 master-data 的配额配置）。
- **permission-service**：先于配额执行的数据权限校验（`AssertPublishScope`，已由 access-data-permission 交付）。
