## ADDED Requirements

### Requirement: Submit management tab with status filter
The system SHALL provide a "提交管理" tab in the grassroots management page. The tab SHALL display a status filter bar at the top with options: 全部, 待提交, 已提交, 已批准, 已拒绝, 待删除. The system SHALL default to showing "待提交" status when the tab is first activated.

#### Scenario: Default view shows pending submissions
- **WHEN** user switches to the "提交管理" tab for the first time
- **THEN** the system SHALL load and display all grassroots divisions (level 4 and level 5) with submission_status = 0

#### Scenario: Filter by submission status
- **WHEN** user selects a status option (e.g., "已提交")
- **THEN** the system SHALL query grassroots divisions with the selected submission_status and display the results in a flat table

### Requirement: Flat table with full path column
The submit management tab SHALL display data in a flat (non-tree) table. The table SHALL include a "完整路径" column showing the human-readable location path (e.g., "湖南省 > 长沙市 > 岳麓区 > 麓谷街道"). Other columns SHALL include: 区划代码, 区划名称, 级别, 提交状态, 操作.

#### Scenario: Path displays resolved names
- **WHEN** the submit management table renders a row
- **THEN** the "完整路径" column SHALL show division names separated by " > " (e.g., "湖南省 > 长沙市 > 岳麓区 > 麓谷街道") instead of raw IDs

#### Scenario: Path resolution uses cached name map
- **WHEN** the submit management tab loads data
- **THEN** the system SHALL fetch level 1-3 divisions once to build an ID-to-name cache, and use this cache to resolve path IDs to names

### Requirement: Single and batch submit
The submit management tab SHALL support submitting a single division and batch submitting multiple selected divisions. Rows with submission_status 0 (待提交) or 3 (已拒绝) SHALL be submittable. Selection checkboxes SHALL only appear on submittable rows.

#### Scenario: Submit a single division
- **WHEN** user clicks "提交" on a row with submission_status 0 or 3
- **THEN** the system SHALL call `submitDivision(id)`, refresh the table, and show a success message

#### Scenario: Batch submit selected divisions
- **WHEN** user selects multiple rows and clicks "批量提交"
- **THEN** the system SHALL call `batchSubmitDivisions({ ids })`, refresh the table, and show the success count

#### Scenario: Non-submittable rows cannot be selected
- **WHEN** a row has submission_status 1 (已提交) or 2 (已批准)
- **THEN** the row SHALL NOT display a selection checkbox
