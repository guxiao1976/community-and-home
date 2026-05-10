## ADDED Requirements

### Requirement: Grassroots management page with district cascade selector
The system SHALL provide a "基层组织" (Grassroots Organizations) page under the "主数据管理" menu. The page SHALL display a province-city-district three-level cascade selector (`el-cascader`, lazy mode) at the top, a "搜索" button, and a "新增" button. Selecting a district and clicking "搜索" SHALL load the street/town divisions under that district. Clicking "新增" SHALL open a dialog to create a new street/town under the selected district.

#### Scenario: Load streets by district selection
- **WHEN** user selects a district from the cascade selector and clicks "搜索"
- **THEN** the system SHALL call `getAdministrativeDivisions({ parent_id: <districtId> })` and display the returned street/town records in a tree table

#### Scenario: Add new street under district
- **WHEN** user selects a district and clicks "新增"
- **THEN** the system SHALL open a create dialog with the parent division pre-filled to the selected district and level fixed to 4 (街道/乡镇)

#### Scenario: Cascade selector lazy loading
- **WHEN** user selects a province in the cascade selector
- **THEN** the system SHALL load cities under that province; upon selecting a city, the system SHALL load districts under that city

### Requirement: Grassroots tree table display
The system SHALL display street/town and community/village divisions in a tree table (`el-table` with lazy loading). Streets SHALL be the root nodes, and communities SHALL be loaded lazily when a street node is expanded. Each row SHALL display: code, name, level, path, submission status, and action buttons.

#### Scenario: Expand street to load communities
- **WHEN** user clicks the expand arrow on a street row
- **THEN** the system SHALL call `loadChildren` to fetch community/village records under that street and display them as child rows

#### Scenario: Tree table columns
- **WHEN** the grassroots table is rendered
- **THEN** it SHALL display columns: 区划代码, 区划名称, 级别, 路径, 提交状态, 操作

### Requirement: Grassroots CRUD operations
The system SHALL support editing, submitting for approval, adding subordinate (community under street), and deleting street/town and community/village records. These operations SHALL reuse the existing `useDivisionStore` actions.

#### Scenario: Edit a street or community
- **WHEN** user clicks "编辑" on a street or community row
- **THEN** the system SHALL open an edit dialog pre-filled with the record's current data (name, sort_order)

#### Scenario: Add community under street
- **WHEN** user clicks "添加下级" on a street row
- **THEN** the system SHALL open a create dialog with parent pre-filled to that street and level fixed to 5 (社区/村)

#### Scenario: Submit grassroots record for approval
- **WHEN** user clicks "提交" on a street or community row with submission_status 0 or 3
- **THEN** the system SHALL call `submitDivision(id)` and refresh the table

#### Scenario: Delete grassroots record
- **WHEN** user clicks "删除" on a street or community row with submission_status 0 or 3 and confirms the deletion
- **THEN** the system SHALL call `deleteDivision(id)` and remove the row from the table

### Requirement: Grassroots routing and menu
The system SHALL register a route at `/masterdata/grassroots` with title "基层组织" and add a corresponding menu item under "主数据管理" in the sidebar, positioned after "行政区划".

#### Scenario: Navigate to grassroots page
- **WHEN** user clicks "基层组织" in the sidebar
- **THEN** the router SHALL navigate to `/masterdata/grassroots` and display the grassroots management page
