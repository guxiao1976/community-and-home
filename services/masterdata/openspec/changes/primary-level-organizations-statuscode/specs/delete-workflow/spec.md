## ADDED Requirements

### Requirement: Request delete API
The system SHALL provide `POST /api/masterdata/divisions/:id/request-delete` endpoint to initiate a delete request for an approved division record.

#### Scenario: Successful request delete
- **WHEN** client calls `POST /divisions/:id/request-delete` on a record with `submission_status=2`
- **THEN** the system SHALL set `submission_status=0, submission_type=3` and return `{ success: true }`

#### Scenario: Request delete on non-approved record
- **WHEN** client calls `POST /divisions/:id/request-delete` on a record with `submission_status != 2`
- **THEN** the system SHALL return error 400 with message "仅已批准状态的数据可以发起删除"

### Requirement: Cancel delete API
The system SHALL provide `POST /api/masterdata/divisions/:id/cancel-delete` endpoint to cancel a pending delete request.

#### Scenario: Successful cancel delete
- **WHEN** client calls `POST /divisions/:id/cancel-delete` on a record with `submission_status=0, submission_type=3`
- **THEN** the system SHALL set `submission_status=2` (preserve `submission_type=3` for audit) and return `{ success: true }`

#### Scenario: Cancel delete on non-delete-pending record
- **WHEN** client calls `POST /divisions/:id/cancel-delete` on a record that is NOT `submission_status=0, submission_type=3`
- **THEN** the system SHALL return error 400 with message "仅删除待提交状态的数据可以取消删除"

### Requirement: Withdraw submission API
The system SHALL provide `POST /api/masterdata/divisions/:id/withdraw` endpoint to withdraw a submitted record back to pending state.

#### Scenario: Successful withdraw
- **WHEN** client calls `POST /divisions/:id/withdraw` on a record with `submission_status=1`
- **THEN** the system SHALL set `submission_status=0` (preserve `submission_type`) and return `{ success: true }`

#### Scenario: Withdraw on non-submitted record
- **WHEN** client calls `POST /divisions/:id/withdraw` on a record with `submission_status != 1`
- **THEN** the system SHALL return error 400 with message "仅已提交状态的数据可以撤回"

### Requirement: Physical delete with status guard
The existing `DELETE /api/masterdata/divisions/:id` endpoint SHALL add a pre-check: only allow execution when `submission_status=0 AND submission_type=1`.

#### Scenario: Physical delete on new pending allowed
- **WHEN** client calls `DELETE /divisions/:id` on a record with `submission_status=0, submission_type=1`
- **THEN** the system SHALL physically delete the record (soft delete via `delete_time`)

#### Scenario: Physical delete on other states rejected
- **WHEN** client calls `DELETE /divisions/:id` on a record that is NOT `submission_status=0, submission_type=1`
- **THEN** the system SHALL return error 403 with message "仅新建待提交的数据可以物理删除，已批准数据请使用发起删除功能"

### Requirement: Approve delete triggers soft delete
When the approval center approves a delete request (`submission_type=3`), the system SHALL execute soft delete on the record.

#### Scenario: Approve delete submission
- **WHEN** approver approves a division record with `submission_status=1, submission_type=3`
- **THEN** the system SHALL set `submission_status=2`, set `delete_time=now()`, and the record SHALL no longer appear in any query results

#### Scenario: Reject delete submission restores approved
- **WHEN** approver rejects a division record with `submission_status=1, submission_type=3`
- **THEN** the system SHALL set `submission_status=2` directly (rejecting a delete means keeping the data)

### Requirement: Frontend API functions
The frontend SHALL provide three new API call functions in `web/pc/src/api/masterdata.ts`:

- `requestDeleteDivision(id: number)` — calls POST `/divisions/:id/request-delete`
- `cancelDeleteDivision(id: number)` — calls POST `/divisions/:id/cancel-delete`
- `withdrawDivision(id: number)` — calls POST `/divisions/:id/withdraw`

#### Scenario: API function signatures
- **WHEN** the frontend module imports from `@/api/masterdata`
- **THEN** the three new functions SHALL be available for import

### Requirement: Delete confirmation dialogs
The frontend SHALL show different confirmation messages based on the delete operation type.

#### Scenario: Physical delete confirmation
- **WHEN** user clicks delete on a new pending record (0,1)
- **THEN** the system SHALL show confirmation dialog: "确定要删除吗？删除后将无法恢复。"

#### Scenario: Request delete confirmation
- **WHEN** user clicks delete on an approved record (2)
- **THEN** the system SHALL show confirmation dialog: "确定要申请删除该组织机构吗？删除需审批通过后生效。"

#### Scenario: Cancel delete confirmation
- **WHEN** user clicks cancel delete on a delete-pending record (0,3)
- **THEN** the system SHALL show confirmation dialog: "确定要取消删除申请吗？取消后数据将恢复为已批准状态。"
