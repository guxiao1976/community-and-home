## ADDED Requirements

### Requirement: Backend auto-generates residential area code
The system SHALL auto-generate a unique code for each newly created residential area. The code format SHALL be the county's 6-digit administrative code followed by a 4-digit sequential number (e.g., `1101050001`, `1101050002`). The system SHALL ignore any `code` value sent by the frontend.

#### Scenario: First residential area in a county
- **WHEN** a new residential area is created with `county_id` for a county whose code is `110105` and no existing residential areas exist for that county
- **THEN** the system SHALL generate code `1101050001`

#### Scenario: Subsequent residential area in a county
- **WHEN** a new residential area is created with `county_id` for a county whose code is `110105` and the maximum existing code for that county is `1101050015`
- **THEN** the system SHALL generate code `1101050016`

#### Scenario: Code uniqueness conflict
- **WHEN** two concurrent requests attempt to create a residential area in the same county and generate the same code
- **THEN** the second request SHALL detect the duplicate via database unique constraint and retry with the next sequential number once

### Requirement: CreateResidentialAreaReq makes code optional
The `code` field in `CreateResidentialAreaReq` SHALL be optional. The backend SHALL NOT use any frontend-provided code value.

#### Scenario: Frontend omits code field
- **WHEN** a create request is sent without a `code` field
- **THEN** the system SHALL proceed to auto-generate the code and create the residential area successfully

#### Scenario: Frontend sends code field (backward compatibility)
- **WHEN** a create request is sent with a `code` field
- **THEN** the system SHALL ignore the provided code and auto-generate a new one

### Requirement: Address field is optional on create
The `address` field in `CreateResidentialAreaReq` SHALL be optional.

#### Scenario: Create without address
- **WHEN** a create request is sent without an `address` field
- **THEN** the system SHALL create the residential area with address as NULL

### Requirement: Community type auto-default based on division level
The frontend SHALL auto-select the community type based on whether `community_div_id` is provided: with `community_div_id` → 住宅小区 (1), without `community_div_id` (township only) → 村庄 (2). The user SHALL still be able to manually change the type.

#### Scenario: Create with community selected
- **WHEN** the create form is opened with `community_div_id` in the route query
- **THEN** `community_type` SHALL default to 1 (住宅小区)

#### Scenario: Create with township only (no community)
- **WHEN** the create form is opened without `community_div_id` in the route query
- **THEN** `community_type` SHALL default to 2 (村庄)

### Requirement: Create form carries division info from list page
The create form SHALL receive `county_id`, `street_id`, `community_div_id`, and division names via route query parameters and display them as read-only information. The user SHALL NOT be able to modify these fields on the create form.

#### Scenario: Navigate to create with full division path
- **WHEN** user clicks "新建小区" with county=朝阳区, street=某街道, community=某社区
- **THEN** the create form SHALL display "朝阳区 > 某街道 > 某社区" as read-only text and SHALL NOT show a division selector

### Requirement: Create form removes area and population fields
The create form SHALL NOT display fields for 小区面积 (area) and 人口数量 (population).

#### Scenario: Form field visibility
- **WHEN** the create form is rendered
- **THEN** no input fields for area or population SHALL be visible
