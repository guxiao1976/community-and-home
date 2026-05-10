## ADDED Requirements

### Requirement: New button validates leaf node selection
The "新建小区" button on the list page SHALL validate that the user has selected an administrative division down to a leaf node before allowing navigation to the create form. A leaf node is defined as a community division (level 5) or a township/street (level 4) that has no child communities.

#### Scenario: User selected down to community (leaf node)
- **WHEN** user has selected province, city, district, street, and community and clicks "新建小区"
- **THEN** the system SHALL navigate to the create form with division info in query params

#### Scenario: User selected township with no children (leaf node)
- **WHEN** user has selected province, city, district, and a street/township that has zero child communities (`communityOptions.length === 0`) and clicks "新建小区"
- **THEN** the system SHALL navigate to the create form with division info in query params (no `community_div_id`)

#### Scenario: User selected street that has child communities (not leaf)
- **WHEN** user has selected province, city, district, and a street that has child communities (`communityOptions.length > 0`) but has NOT selected a community and clicks "新建小区"
- **THEN** the system SHALL show a warning message "该街道下有社区，请选择社区后再新建小区" and SHALL NOT navigate to the create form

#### Scenario: User only selected province and city (not deep enough)
- **WHEN** user has selected province and city but NOT district/street/community and clicks "新建小区"
- **THEN** the system SHALL show a warning message "请选择到区县后再新建小区" and SHALL NOT navigate

#### Scenario: User only selected province, city, and district (not deep enough)
- **WHEN** user has selected province, city, and district but NOT street and clicks "新建小区"
- **THEN** the system SHALL show a warning message "请选择街道/乡镇后再新建小区" and SHALL NOT navigate

#### Scenario: Community options failed to load (uncertain state)
- **WHEN** user has selected a street but the community options request failed (`communityOptions` is null or in error state)
- **THEN** the system SHALL show a warning message "无法确认是否为乡镇，请重试或选择社区后再新建" and SHALL NOT navigate

### Requirement: Create form displays read-only division path
The create form SHALL display the division hierarchy path (e.g., "陕西省 > 咸阳市 > 秦都区 > 人民路街道 > XX社区") as read-only descriptive text. The user SHALL NOT be able to modify these values.

#### Scenario: Display with community
- **WHEN** the create form is opened with `county_id`, `street_id`, and `community_div_id` in query
- **THEN** the form SHALL show the full 5-level path as read-only text

#### Scenario: Display with township only
- **WHEN** the create form is opened with `county_id` and `street_id` but no `community_div_id`
- **THEN** the form SHALL show the 4-level path ending at the township as read-only text
