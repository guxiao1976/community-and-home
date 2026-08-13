# Same-House Visibility（同屋互见·户级数据可见性）Specification

## Purpose

定义户级数据可见性规则：同一小区同一房屋（楼号+单元号+房号）的活跃成员之间互相可见手机号与楼栋房屋号，用于同屋互相监督、防止冒领房屋与冒充身份；非同屋用户查看时手机号一律脱敏。

## 术语

- **active membership ⟺ bind_status = 1（Active/有效）**：表示「该用户仍是该小区/该房屋的有效成员」。
- **成员生命周期（bind_status）与认证状态（verf_status / 能力分层）是两个正交维度**：同屋互见**仅依据 bind_status = Active 判定**，与认证状态无关（未认证业主与租户同样参与互见）。

## Requirements

### Requirement: 同屋互见判定
系统 SHALL 在用户查看另一用户详情时，判定双方是否存在「同一小区 + 同一楼栋 + 同一单元 + 同一房号」的 active membership；同屋则对可见方展示手机号（解密后的真实值）与楼栋房屋号，否则手机号脱敏。

#### Scenario: 同屋用户互见手机号
- **GIVEN** 用户 A、B 均在某小区某房屋（community+building+unit+room 相同）持有 active membership
- **WHEN** A 查看 B 的用户详情
- **THEN** 系统返回 B 的手机号（解密后的真实值）与楼栋房屋号

#### Scenario: 非同屋用户查看脱敏
- **GIVEN** 用户 A、C 不在同一房屋（无同屋 active membership）
- **WHEN** A 查看 C 的用户详情
- **THEN** 系统返回 C 的脱敏手机号，且不返回楼栋房屋号（或同样脱敏）

#### Scenario: 租户参与同屋互见
- **GIVEN** 用户 A 为业主、B 为租户，同屋 active membership 成立
- **WHEN** A 查看 B（或反之）
- **THEN** 同屋互见规则同样适用，双方互相可见手机号与楼栋房屋号

### Requirement: 退出房屋即撤销互见
系统 SHALL 在用户退出房屋（active membership 失效，bind_status 置非 Active）后立即撤销其与该屋其他成员的互见关系。

#### Scenario: 退出后互见消失
- **GIVEN** 用户 A、B 原为同屋互见
- **WHEN** A 退出该房屋（bind_status 置非 Active）
- **THEN** B 再查看 A 详情时，A 的手机号脱敏（互见关系随 active membership 消失）

#### Scenario: 非活跃成员不参与互见
- **GIVEN** A 与 B 历史上同屋但 A 已退出（bind_status 非 Active）
- **WHEN** 系统判定互见关系
- **THEN** 该关系不成立，手机号脱敏

### Requirement: 隐私边界
系统 SHALL 仅在同屋互见场景返回真实手机号；其余场景手机号 MUST 脱敏。系统不采集身份证号、真实姓名，不存在此类隐私面。

#### Scenario: 手机号解密
- **GIVEN** 手机号以加密形态存储
- **WHEN** 系统需要返回同屋可见的真实手机号
- **THEN** 系统先解密再返回，且非同屋场景不返回真实值

#### Scenario: 非同屋不泄露楼栋房号
- **GIVEN** 用户 A 与 B 非同屋
- **WHEN** A 查看 B 详情
- **THEN** 系统不向 A 暴露 B 的楼栋房屋号

> ✅ 已定（CLAR-4）：楼/单元/房号在 JoinCommunity 采集，同屋判定基于 membership 的 building/unit/room。

## 职责边界

- **user-service**：用户详情接口执行同屋互见判定（查询 active membership 的 community+building+unit+room），手机号解密与脱敏。
- **permission-service**：不参与（此为用户详情接口的数据可见性规则，独立于功能权限）。
