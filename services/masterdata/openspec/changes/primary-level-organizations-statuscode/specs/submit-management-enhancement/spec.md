## ADDED Requirements

### Requirement: Operation type column in submit management
The 提交管理 Tab table SHALL include an "操作类型" column that displays `submission_type` as a colored Tag.

Tag colors: 新增(type=1)=蓝色, 修改(type=2)=橙色, 删除(type=3)=红色.

#### Scenario: Display operation type tags
- **WHEN** submit management tab loads data with `submission_status=0` (待提交)
- **THEN** the table SHALL show "操作类型" column with tags: 新增(blue) for type=1, 修改(orange) for type=2, 删除(red) for type=3

#### Scenario: Hide operation type when no pending data
- **WHEN** the filter is set to 已批准 or 已拒绝
- **THEN** the "操作类型" column SHALL still be visible but show appropriate tags based on `submission_type` value

### Requirement: Withdraw button in submit management
The 提交管理 Tab SHALL show a "撤回" button for records with `submission_status=1` (已提交).

#### Scenario: Withdraw button visible on submitted records
- **WHEN** submit management tab shows records with `submission_status=1`
- **THEN** each row SHALL display a "撤回" button

#### Scenario: Withdraw button hidden on other statuses
- **WHEN** submit management tab shows records with `submission_status != 1`
- **THEN** the "撤回" button SHALL NOT appear

#### Scenario: Successful withdraw from submit management
- **WHEN** user clicks "撤回" and confirms
- **THEN** the system SHALL call withdraw API, show success message, and refresh the table

### Requirement: Physical delete button for new pending in submit management
The 提交管理 Tab SHALL show a "删除" button (physical delete) only for records with `submission_status=0, submission_type=1`.

#### Scenario: Delete button on new pending
- **WHEN** submit management tab shows a record with `submission_status=0, submission_type=1`
- **THEN** the row SHALL display both "提交" and "删除" buttons

#### Scenario: No delete button on update/delete pending
- **WHEN** submit management tab shows a record with `submission_status=0` but `submission_type != 1`
- **THEN** the row SHALL display only "提交" button, no "删除" button

### Requirement: Operation type column in query-edit tab
The 查询编辑 Tab tree table SHALL show an "操作类型" column when the filter is set to 待提交 or when any pending records exist.

#### Scenario: Operation type visible for pending records
- **WHEN** query-edit tab displays records with `submission_status=0`
- **THEN** the table SHALL show "操作类型" column with corresponding tags

#### Scenario: Operation type tag for approved records
- **WHEN** query-edit tab displays records with `submission_status=2`
- **THEN** the "操作类型" column SHALL NOT display a tag for approved records (no type applicable)

### Requirement: Unified operation button helpers
The frontend SHALL use a set of helper functions to determine which operation buttons to show for each row, based on `submission_status`, `submission_type`, and `level`.

Helper functions: `canEdit()`, `canDelete()`, `canSubmit()`, `canWithdraw()`, `canCancelDelete()`, `canAddChild()`.

#### Scenario: Button visibility follows permission matrix
- **WHEN** a row has a specific `(submission_status, submission_type, level)` combination
- **THEN** the visible buttons SHALL exactly match the operation permission matrix defined in submission-status-system spec

### Requirement: Batch submit includes all types
The batch submit functionality SHALL allow selecting and submitting all records with `submission_status=0` regardless of `submission_type`.

#### Scenario: Batch select pending records
- **WHEN** submit management tab has records with `submission_status=0` and varying `submission_type` values (1, 2, 3)
- **THEN** all those records SHALL be selectable for batch submit

#### Scenario: Batch submit mixed types
- **WHEN** user selects records with `submission_type=1, 2, 3` and clicks batch submit
- **THEN** all selected records SHALL have `submission_status` changed from 0 to 1
