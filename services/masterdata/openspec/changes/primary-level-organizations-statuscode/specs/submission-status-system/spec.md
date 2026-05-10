## ADDED Requirements

### Requirement: Dual-field status system
The system SHALL use two fields `submission_status` (workflow state) and `submission_type` (operation type) to track the lifecycle of administrative division records at level 4-5.

`submission_status` values: 0=待提交, 1=已提交, 2=已批准, 3=已拒绝.
`submission_type` values: 1=新增, 2=修改, 3=删除.

#### Scenario: Newly created record
- **WHEN** a data maintainer creates a new division record
- **THEN** the system SHALL set `submission_status=0` and `submission_type=1`

#### Scenario: Edit approved record
- **WHEN** a data maintainer edits a record with `submission_status=2`
- **THEN** the backend SHALL set `submission_status=0` and `submission_type=2` upon save

#### Scenario: Edit pending record
- **WHEN** a data maintainer edits a record with `submission_status=0`
- **THEN** the system SHALL update the record content WITHOUT changing `submission_status` or `submission_type`

#### Scenario: Edit rejected record
- **WHEN** a data maintainer edits a record with `submission_status=3`
- **THEN** the system SHALL set `submission_status=0` and preserve the existing `submission_type` value

#### Scenario: Submit for approval
- **WHEN** a data maintainer submits a record with `submission_status=0`
- **THEN** the system SHALL set `submission_status=1` and preserve `submission_type`

#### Scenario: Approve submission
- **WHEN** an approver approves a record with `submission_status=1`
- **THEN** the system SHALL set `submission_status=2` and preserve `submission_type` as audit record

#### Scenario: Reject submission
- **WHEN** an approver rejects a record with `submission_status=1`
- **THEN** the system SHALL set `submission_status=3` and preserve `submission_type`

### Requirement: Division response includes submission_type
The `GET /divisions` and `GET /divisions/:id` responses SHALL include `submission_type` field in each Division object.

#### Scenario: List divisions returns submission_type
- **WHEN** client calls `GET /divisions` with any filter
- **THEN** each item in the response list SHALL contain `submission_type` field (integer or null)

#### Scenario: Single division returns submission_type
- **WHEN** client calls `GET /divisions/:id`
- **THEN** the response Division object SHALL contain `submission_type` field

### Requirement: Legacy data compatibility
Records with `submission_type=NULL` SHALL be treated as approved data (`submission_status=2`) for display and query purposes.

#### Scenario: Query with legacy data
- **WHEN** the database contains records with `submission_type=NULL`
- **THEN** those records SHALL appear in query results normally, and frontend SHALL treat them as approved records

### Requirement: Operation permission matrix
The system SHALL enforce the following operation permissions based on `submission_status` and `submission_type` combination:

| Operation | (0,1) 新建待提交 | (0,2) 修改待提交 | (0,3) 删除待提交 | (1) 已提交 | (2) 已批准 | (3) 已拒绝 |
|---|---|---|---|---|---|---|
| 编辑 | YES | YES | NO | NO | NO | YES |
| 删除 | 物理删除 | NO | NO | NO | 发起删除申请 | NO |
| 提交审批 | YES | YES | YES | NO | NO | 编辑后提交 |
| 撤回提交 | NO | NO | NO | YES→(0) | NO | NO |
| 取消删除 | NO | NO | YES→(2) | NO | NO | NO |
| 添加下级 | NO | NO | NO | NO | YES | NO |

#### Scenario: Physical delete on new pending
- **WHEN** user requests delete on a record with `submission_status=0, submission_type=1`
- **THEN** the system SHALL physically delete the record from database

#### Scenario: Physical delete rejected on non-new-pending
- **WHEN** user requests delete on a record that is NOT `submission_status=0, submission_type=1`
- **THEN** the system SHALL return error 403 with message indicating physical delete is only allowed for newly created pending records

#### Scenario: Request delete on approved
- **WHEN** user requests delete on a record with `submission_status=2`
- **THEN** the system SHALL set `submission_status=0, submission_type=3`

#### Scenario: Cancel delete on delete-pending
- **WHEN** user requests cancel delete on a record with `submission_status=0, submission_type=3`
- **THEN** the system SHALL set `submission_status=2` and preserve `submission_type=3`

#### Scenario: Withdraw submission
- **WHEN** user requests withdraw on a record with `submission_status=1`
- **THEN** the system SHALL set `submission_status=0` and preserve `submission_type`

#### Scenario: Add child only on approved
- **WHEN** user tries to add child division under a parent that is NOT `submission_status=2`
- **THEN** the "添加下级" button SHALL NOT appear in the UI

### Requirement: Status filter bar alignment
Both 查询编辑 Tab and 提交管理 Tab SHALL use the same status filter options: 全部 / 待提交 / 已提交 / 已批准 / 已拒绝. The "待删除" option SHALL be removed.

#### Scenario: Filter options in query-edit tab
- **WHEN** user views the 查询编辑 Tab
- **THEN** the status filter radio group SHALL show: 全部, 待提交, 已提交, 已批准, 已拒绝

#### Scenario: Filter options in submit-management tab
- **WHEN** user views the 提交管理 Tab
- **THEN** the status filter radio group SHALL show: 全部, 待提交, 已提交, 已批准, 已拒绝
