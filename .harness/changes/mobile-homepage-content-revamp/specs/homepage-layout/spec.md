# Homepage Layout Consolidation Capability Specification

## Purpose

整合移动端首页的信息架构：新增「邻里互助」占位区块（本期无后端数据源、无列表页/详情页、无用户发布入口，D8/D9）、「寻失互助」区块保持当前展示风格不变、3 个广告位从分散位置统一移到页面下部垂直堆叠集中展示（内容仍前端硬编码、点击预留不跳转，D6/D7）、首页跑马灯保留（滚动内容同源 30 天数据，行为见 REQ-NTW-3）。涉及 web/mobile（首页区块重排与新增占位）。

## Requirements

### Requirement: REQ-HL-1 — 首页新增邻里互助占位区块（无后端、无页面）

The homepage SHALL render a 邻里互助 section as a placeholder: since this iteration has no backend data source and no list/detail pages and no user publish entry (D8/D9), the section SHALL display placeholder/empty content indicating the feature is not yet available, SHALL NOT navigate to any page when tapped, and SHALL NOT fabricate help-need data. The "show 7-day requests, 3 on homepage" behavior depends on a backend that is out of scope this iteration and SHALL be deferred (declared in the Won't-have list).

#### Scenario: 占位区块展示且不跳转
- **GIVEN** a mobile homepage with a community selected
- **WHEN** the homepage renders below the function entries
- **THEN** a 邻里互助 section appears with placeholder content (e.g. 「互助功能开发中」/ empty state) and does not render fabricated requests

#### Scenario: 点击占位区块不产生导航
- **GIVEN** the 邻里互助 placeholder section is rendered
- **WHEN** the user taps the section
- **THEN** the system stays on the homepage (no navigation to a list/detail page that does not exist)

### Requirement: REQ-HL-2 — 寻失互助区块保持当前展示风格

The homepage 寻失互助 section SHALL keep its current presentation unchanged: horizontal-scroll cards with first image (or placeholder), lost/found type tag, title, and time, plus the "全部 →" footer; its data source (getLostFoundList) and click behavior (detail toast) SHALL remain as today. No layout, style, or data change is introduced by this change.

#### Scenario: 寻失互助样式与数据不变
- **GIVEN** a community with lost/found items
- **WHEN** the homepage renders the 寻失互助 section
- **THEN** the section renders identically to today (horizontal cards, type tag, title, time, "全部 →"), with the same data

#### Scenario: 寻失互助空态保持不变
- **GIVEN** a community with no lost/found items
- **WHEN** the homepage renders the 寻失互助 section
- **THEN** the section shows the existing empty state (「暂无寻失信息」) as today

### Requirement: REQ-HL-3 — 3 个广告位集中到页面底部垂直堆叠

The homepage SHALL remove the 3 ad banners from their current scattered positions (currently 2 below the contact grid and 1 below the lost-found section) and SHALL render them consolidated at the bottom of the page, stacked vertically one above another (D6). Each banner SHALL preserve its original hardcoded visual content. Tapping an ad SHALL remain a reserved no-op with no navigation (D7).

#### Scenario: 广告位移除原分散位置并集中到底部堆叠
- **GIVEN** a mobile homepage whose 3 ad banners were previously scattered (2 mid-page, 1 below lost-found)
- **WHEN** the homepage renders after the change
- **THEN** no ad banner appears in the former scattered positions; exactly 3 ad banners appear at the page bottom, vertically stacked in order

#### Scenario: 广告内容保持前端硬编码
- **GIVEN** the 3 ad banners consolidated at the page bottom
- **WHEN** the homepage renders
- **THEN** each banner shows its original hardcoded content (label/title/desc/button) unchanged; no backend data is involved

#### Scenario: 广告点击预留不跳转
- **GIVEN** a rendered ad banner at the page bottom
- **WHEN** the user taps the ad
- **THEN** the system performs no navigation (reserved for future ad integration, D7)

### Requirement: REQ-HL-4 — 首页区块垂直全序固定

The homepage SHALL render its content sections in the following fixed vertical order: (1) notice section (marquee bar + up-to-3 notice cards, per REQ-NTW-1/3/5), (2) the 4 function icon entries (REQ-FE-1), (3) the 邻里互助 placeholder section (REQ-HL-1), (4) the 寻失互助 section (REQ-HL-2, unchanged), (5) the consolidated bottom ad banners (REQ-HL-3). No section SHALL be reordered, removed, or interleaved by this change. (REVISION: 评审 coverage SHOULD #6——补齐首页垂直区块链全序，作为前端布局唯一依据；此前顺序仅隐式存在于 proposal 与各能力 Purpose)

#### Scenario: 首页按固定顺序渲染全部区块
- **GIVEN** a mobile homepage with a community selected and non-empty sections
- **WHEN** the homepage renders
- **THEN** the vertical order is 通知（跑马灯+卡片）→ 4 功能入口 → 邻里互助占位 → 寻失互助 → 底部广告位, with no reordering

#### Scenario: 部分区块空态时顺序仍保持
- **GIVEN** a mobile homepage where the notice window is empty and 寻失互助 is empty
- **WHEN** the homepage renders
- **THEN** each section still renders in the fixed order with its own empty state (通知空态「暂无通知公告」/ 寻失空态保持现状), and the overall order is not disturbed

## 服务职责边界

- **web/mobile**: 首页区块重排——按 REQ-HL-4 固定垂直全序；新增邻里互助占位区块（无后端/无页面，不跳转）；寻失互助区块保持现状；3 个广告位移除原位置、页面底部垂直堆叠集中展示（内容前端硬编码、点击预留）；首页跑马灯保留（由首页通知列表派生，见 REQ-NTW-3）
- 其他服务：本能力无后端变更（邻里互助/广告均不涉及后端数据源）
