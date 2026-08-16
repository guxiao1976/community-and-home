# Contact List Page & Data Foundation Capability Specification

## Purpose

新建「便民联络」列表页，承载原首页内嵌的联络拨号网格（按类别图标 + 名称 + 电话，点击拨号），数据复用 ListContacts API；同时补齐 community-hub-service 运行库缺失的 `community_contacts` 表（当前运行库缺表导致 ListContacts 报 Table doesn't exist），以幂等 migration 004 建表且不预置种子数据（空态，运营后续维护）。目标是让便民联络入口的落地页可用、读链路不再因缺表失败。涉及 web/mobile（新页面 + pages.json 注册）与 community-hub-service（migration 004 补表）。

> **REVISION 修订记录**：本轮修订解决评审（clarity SHOULD #1 / coverage INFO #7 / structure S1）——`community_contacts` 表 DDL 已存在于 `migration/001_initial.sql`（列结构与 CommunityContactModel 一致），004 是对「001 已应用后运行库仍缺表」的幂等补救，004 DDL 须与 001/CommunityContactModel 完全对齐，并声明「运行库表存在但结构漂移时 IF NOT EXISTS 不自动修复、需人工订正」的边界。

## Requirements

### Requirement: REQ-CLP-1 — 联络列表页经 ListContacts API 渲染拨号网格

The contact list page SHALL load contact data via the ListContacts API (`GET /api/community/contacts`, `community_id` = current community) and SHALL render a dial grid presenting each contact's category icon, category name, and phone number, consistent with the previous homepage contact grid presentation. Tapping a contact SHALL invoke the phone dial for that contact's number. When no contact data exists, the page SHALL render the empty state "暂无联络信息"; when the API fails, the page SHALL surface an explicit error (禁止静默吞错).

#### Scenario: 联络数据渲染为拨号网格
- **GIVEN** a community whose `community_contacts` table holds contacts across categories (water/electricity/...)
- **WHEN** a user opens the contact list page
- **THEN** the page renders each contact as a grid cell with category icon, category name, and phone number

#### Scenario: 点击联络卡片拨号
- **GIVEN** a rendered contact grid cell with phone number P
- **WHEN** the user taps the cell
- **THEN** the system invokes the phone dial for P

#### Scenario: 联络数据为空显示空态
- **GIVEN** a community with an empty `community_contacts` table (no seed data, D4)
- **WHEN** a user opens the contact list page
- **THEN** the page renders the empty state "暂无联络信息" and does not crash

#### Scenario: 列表加载失败明确提示
- **GIVEN** the ListContacts API fails (backend unavailable / table missing)
- **WHEN** a user opens the contact list page
- **THEN** the page shows an explicit load-failure message and does not silently render an empty-looking page

### Requirement: REQ-CLP-2 — community_contacts 表幂等补齐（migration 004，不预置种子，DDL 与 001/模型对齐）

The community-hub-service SHALL provide migration 004 that idempotently creates the `community_contacts` table so that the ListContacts read no longer fails with a missing-table error in the runtime database. The 004 DDL SHALL match the existing `community_contacts` definition in `migration/001_initial.sql` and the CommunityContactModel exactly (columns: id / community_id / category / name / phone / sort_order / created_at / updated_at, plus an index on community_id), so that 001 remains the single schema authority and 004 is a duplicate-definition-free remediation for runtime DB drift. The migration SHALL NOT insert seed data (empty state, maintained later by operations, D4). Re-running the migration SHALL be harmless (idempotent). The root cause of the runtime missing table is a migration-application gap (001 was applied before its DDL was complete, or the runtime DB drifted from the migration set); 004 compensates idempotently and SHALL NOT be treated as an excuse to remove or diverge from the 001 definition.

#### Scenario: 迁移执行后 ListContacts 可用
- **GIVEN** a runtime database where `community_contacts` is missing (ListContacts currently errors "table doesn't exist")
- **WHEN** the migration 004 is applied
- **THEN** the table is created and ListContacts returns successfully (empty list for a fresh table), no missing-table error

#### Scenario: 迁移幂等可重跑
- **GIVEN** migration 004 has already been applied once
- **WHEN** the migration is executed again (e.g. on redeploy)
- **THEN** the migration completes without error and existing table data is untouched (CREATE TABLE IF NOT EXISTS)

#### Scenario: 不预置种子数据
- **GIVEN** migration 004 applied on a fresh database
- **WHEN** the migration completes
- **THEN** the `community_contacts` table exists but contains zero rows (empty state, no seed data inserted)

#### Scenario: 表结构含既有模型字段与索引
- **GIVEN** migration 004 applied
- **WHEN** the schema is inspected
- **THEN** the table exposes id/community_id/category/name/phone/sort_order/created_at/updated_at and an index on community_id, matching the existing CommunityContactModel contract (REQ-CLP-1 read path) and the `migration/001_initial.sql` definition

#### Scenario: 运行库表存在但结构漂移时不自动修复（边界）
- **GIVEN** a runtime database where a `community_contacts` table exists but its columns drift from CommunityContactModel (e.g. missing sort_order or altered column type)
- **WHEN** migration 004 is applied
- **THEN** `CREATE TABLE IF NOT EXISTS` does not repair the drifted structure (MySQL IF NOT EXISTS skips an existing table); the drift is NOT silently masked and requires manual correction per operations, and the ListContacts read on the drifted table SHALL NOT be assumed healthy

## 服务职责边界

- **web/mobile**: 新增 `pages/contact-list/contact-list.vue`（拨号网格：类别图标/名称/电话，点击拨号）并在 `pages.json` 注册；经 ListContacts API 读当前小区联络数据；空态与加载失败明确处理
- **community-hub-service**: migration 004 幂等建 `community_contacts` 表（不预置种子）；ListContacts 读路径复用既有 CommunityContactModel（无新增接口，补表后即可用）
