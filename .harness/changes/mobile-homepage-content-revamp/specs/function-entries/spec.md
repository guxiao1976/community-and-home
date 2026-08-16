# Homepage Function Entries Capability Specification

## Purpose

在首页通知区块下方提供 4 个功能图标入口按钮——便民联络、物业报修、二手闲置、租房卖房。便民联络为「做实」入口：跳转新建的联络列表页承载拨号网格（复用 ListContacts API，见 contact-list-page 能力）；其余 3 个为纯入口占位：点击提示「功能开发中」且不跳转。目标是让首页成为「通知 + 常用功能入口」的清晰聚合页，同时明确占位入口的行为边界（目标落地页本期未定义）。涉及 web/mobile（首页入口渲染 + 交互）。

## Requirements

### Requirement: REQ-FE-1 — 首页通知下方渲染 4 个功能图标入口

The homepage SHALL render, immediately below the notice section, exactly 4 function icon entry buttons in the fixed order: 便民联络, 物业报修, 二手闲置, 租房卖房. Each entry SHALL display an icon and its label.

#### Scenario: 4 个入口按固定顺序渲染
- **GIVEN** a mobile homepage with a community selected
- **WHEN** the homepage renders the content below the notice section
- **THEN** 4 function entry buttons appear in the order 便民联络 / 物业报修 / 二手闲置 / 租房卖房, each with an icon and label

#### Scenario: 未加入小区时入口区隐藏
- **GIVEN** a user who has not joined any community (hasCommunities == false)
- **WHEN** the homepage renders
- **THEN** the 4 function entries are not rendered (the page shows the join-community hint as today, per existing no-community behavior)

### Requirement: REQ-FE-2 — 便民联络入口做实，跳转联络列表页

When a user taps the 便民联络 entry, the system SHALL navigate to the contact list page (`pages/contact-list/contact-list`), which renders the dial grid from the ListContacts API (REQ-CLP-1). The homepage SHALL NOT render the in-page contact dial grid anymore (the contact data presentation moves to the contact list page, D3).

#### Scenario: 便民联络入口跳转联络列表页
- **GIVEN** a homepage rendering the 4 function entries
- **WHEN** the user taps 便民联络
- **THEN** the system navigates to the contact list page; the homepage does not render the previous in-page contact grid

#### Scenario: 联络列表页数据为空时展示空态
- **GIVEN** a community with no contact data (community_contacts empty)
- **WHEN** the user taps 便民联络 and the contact list page loads
- **THEN** the contact list page shows its empty state (「暂无联络信息」) and no crash or blank failure occurs (REQ-CLP-1)

### Requirement: REQ-FE-3 — 物业报修 / 二手闲置 / 租房卖房为入口占位，点击提示「功能开发中」

When a user taps 物业报修, 二手闲置, or 租房卖房, the system SHALL show a "功能开发中" toast and SHALL NOT navigate to any page (the target landing pages are not defined this iteration, D3).

#### Scenario: 占位入口点击提示且不跳转
- **GIVEN** a homepage rendering the 4 function entries
- **WHEN** the user taps 物业报修 (or 二手闲置 / 租房卖房)
- **THEN** the system shows the toast "功能开发中" and stays on the homepage (no navigation)

#### Scenario: 重复点击占位入口仍只提示不跳转
- **GIVEN** the user repeatedly taps 二手闲置 within a short interval
- **WHEN** the taps occur
- **THEN** each tap shows the "功能开发中" toast and never triggers navigation (no accidental page open)

## 服务职责边界

- **web/mobile**: 首页渲染 4 个功能图标入口（固定顺序）；便民联络 → 跳转联络列表页；物业报修/二手闲置/租房卖房 → 占位 toast「功能开发中」、不跳转；原首页内嵌联络拨号网格移除
- **community-hub-service**: ListContacts API（既有，复用）为联络列表页提供数据；无本能力新增后端变更
