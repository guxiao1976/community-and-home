## MODIFIED Requirements

### Requirement: Query edit tab with district locator
The grassroots management page SHALL use `el-tabs` with a "查询编辑" tab as the first tab. This tab SHALL contain the province-city-district cascade selector, a search button, an optional status filter, and the tree table for streets and communities. The submit button SHALL be removed from this tab.

#### Scenario: Tab layout on page load
- **WHEN** the grassroots management page loads
- **THEN** the system SHALL display two tabs: "查询编辑" (active by default) and "提交管理"

#### Scenario: No submit button in query edit tab
- **WHEN** the query edit tab renders a table row
- **THEN** the action column SHALL NOT include a "提交" button

#### Scenario: Status filter as optional辅助 in query edit tab
- **WHEN** user selects a district and optionally selects a status filter
- **THEN** the system SHALL load streets under that district, filtered by the selected submission_status if provided

#### Scenario: Switching tabs preserves state
- **WHEN** user switches from "查询编辑" tab to "提交管理" tab and back
- **THEN** the previous tab's filter selections and table data SHALL be preserved

### Requirement: Query edit tab CRUD operations
The query edit tab SHALL support editing, adding subordinate communities under streets, and deleting streets and communities. These operations SHALL reuse the existing `useDivisionStore` and `submitDivision` API.

#### Scenario: Edit a street or community in query edit tab
- **WHEN** user clicks "编辑" on a street or community row
- **THEN** the system SHALL open an edit dialog with the record's current data

#### Scenario: Add community under street in query edit tab
- **WHEN** user clicks "添加下级" on a street row
- **THEN** the system SHALL open a create dialog with parent pre-filled and level fixed to 5

#### Scenario: Delete in query edit tab
- **WHEN** user clicks "删除" on a row with submission_status 0 or 3 and confirms
- **THEN** the system SHALL delete the record and refresh the table
