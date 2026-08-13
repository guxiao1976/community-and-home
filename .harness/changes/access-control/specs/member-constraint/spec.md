# Member Constraint（成员约束·反滥用）Specification

## Purpose

定义加入小区与房屋注册的反滥用约束：限制用户同时加入的小区数、每年新加入数、终身加入总数，以及单户房屋的注册人数上限（≤6）。目的在防冒领、防反复退出/重加入刷「未认证」身份发布内容。所有上限均可配置。

## 术语

- **active membership ⟺ bind_status = 1（Active/有效）**：表示「该用户仍是该小区/该房屋的有效成员」。
- **成员生命周期（bind_status）与认证状态（verf_status / 能力分层）是两个正交维度**：`bind_status` 回答「还在不在这个小区/房屋」，`verf_status` 回答「认证到哪一层」。本节全部约束**仅依据 bind_status = Active 计数**，与认证状态无关。

## Requirements

### Requirement: 同时加入小区数上限
系统 SHALL 限制用户同时持有的有效小区成员关系（active membership）数量不超过配置上限（默认 3）；超限时拒绝加入并返回语义唯一的「加入小区数超限」错误（参考错误码 10006，最终码阶段3对齐）。

#### Scenario: 未达上限加入
- **GIVEN** 用户当前同时加入 2 个小区，上限 3
- **WHEN** 用户加入第 3 个小区
- **THEN** 系统放行，成员关系建立

#### Scenario: 达上限加入
- **GIVEN** 用户当前同时加入 3 个小区，上限 3
- **WHEN** 用户加入第 4 个小区
- **THEN** 系统拒绝并返回「加入小区数超限」错误

### Requirement: 每年新加入数上限
系统 SHALL 限制用户每年（自然年）新加入小区数量不超过配置上限（默认 3，仅对非认证用户受限）；超限时拒绝加入。

#### Scenario: 非认证用户年内达上限
- **GIVEN** 非认证用户本年度已新加入 3 个小区，上限 3
- **WHEN** 该用户本年度再次加入新小区
- **THEN** 系统拒绝加入

#### Scenario: 认证用户不受每年次数限制
- **GIVEN** 用户为已认证业主，本年度新加入数已达上限
- **WHEN** 该用户加入新小区
- **THEN** 系统不按每年次数限制拦截（仅受同时/终身次数限制）

> ✅ 已定（STAGE3-1）：认证粒度 per-community——仅「目标小区认证」的 owner/tenant 视为该小区的认证用户豁免每年限制；A 小区认证不代表 B 小区认证。

### Requirement: 终身加入总数上限
系统 SHALL 限制用户终身累计加入小区总数不超过配置上限（默认 12）；超限时拒绝加入。

#### Scenario: 终身未达上限
- **GIVEN** 用户终身累计已加入 5 个小区（含已退出的），上限 12
- **WHEN** 用户加入新小区
- **THEN** 系统放行，成员关系建立

#### Scenario: 终身达上限
- **GIVEN** 用户终身累计已加入 12 个小区（含已退出的），上限 12
- **WHEN** 用户再加入新小区
- **THEN** 系统拒绝加入

#### Scenario: 反复退出重加入刷身份
- **GIVEN** 用户通过反复退出/重加入企图重置「未认证」身份
- **WHEN** 系统累计其终身加入次数
- **THEN** 终身次数上限仍拦截，无法绕过（计数含历史退出记录）

### Requirement: 每户房屋注册人数上限
系统 SHALL 限制同一小区内同一房屋（楼号+单元号+房号）的活跃注册人数（active membership）不超过配置上限（默认 6，业主+租户合计）；超限时拒绝并返回语义唯一的「该房屋已满员」错误（参考错误码 10014，最终码阶段3对齐）。计数 MUST 排除当前操作用户自身，且仅统计活跃成员。

#### Scenario: 房屋未满员注册
- **GIVEN** 某房屋现有 5 名活跃成员，上限 6
- **WHEN** 新用户注册加入该房屋
- **THEN** 系统放行，成为第 6 名成员

#### Scenario: 房屋满员注册
- **GIVEN** 某房屋已有 6 名活跃成员，上限 6
- **WHEN** 第 7 人注册加入该房屋
- **THEN** 系统拒绝并返回「该房屋已满员」错误

#### Scenario: 退出后重新加入同一房屋
- **GIVEN** 用户此前退出某房屋（非活跃）
- **WHEN** 用户重新注册加入该房屋
- **THEN** 系统仅按当前活跃成员计数（不含已退出者），未满员则放行

#### Scenario: 计数排除当前用户
- **GIVEN** 用户已是某房屋成员并更新自身信息
- **WHEN** 系统校验该房屋人数
- **THEN** 计数不把当前用户自身计入上限判定

> ✅ 已定（CLAR-4，选项 a+c）：JoinCommunity 时即采集楼/单元/房号，membership 增 building/unit/room 三列。
>
> ✅ 已定：错误码阶段3 统一分配（现状 10006 语义冲突，需新增码）。

## 职责边界

- **user-service**：JoinCommunity 时执行全部成员约束校验（同时/每年/终身/每户人数），权威执行。
- **master-data-service**：提供约束上限的 `sys_config` 键（`user.max_community_join_count` / `user.max_new_communities_per_year` / `user.max_total_communities_lifetime` / `user.max_house_members`）。
- **permission-service**：不参与成员约束判定（约束是成员域独立规则，见设计 §2 附注）。
