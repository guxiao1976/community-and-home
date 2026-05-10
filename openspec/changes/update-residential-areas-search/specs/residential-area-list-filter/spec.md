## MODIFIED Requirements

### Requirement: Five-level administrative division cascading filter
The residential area list page filter SHALL display five independent dropdown selectors: Province (level 1), City (level 2), District/County (level 3), Street/Town (level 4), and Community (level 5). Each level loads its children when the parent is selected. Changing an upper-level selection SHALL clear all lower-level selections.

#### Scenario: Selecting a district loads streets
- **WHEN** the user selects a district (level 3)
- **THEN** the street dropdown is populated with level 4 divisions under that district
- **AND** the community dropdown is cleared and remains disabled

#### Scenario: Selecting a street loads communities
- **WHEN** the user selects a street (level 4)
- **THEN** the community dropdown is populated with level 5 divisions under that street

#### Scenario: Changing district clears street and community
- **WHEN** the user changes the district selection
- **THEN** the street and community selections are cleared
- **AND** the street dropdown is repopulated with divisions under the new district

### Requirement: Search uses most specific division selected
When executing a search, the system SHALL use the most specific division level selected. If community is selected, use `community_div_id`; otherwise if district is selected, use `county_id`.

#### Scenario: Search with community selected
- **WHEN** the user selects a community and clicks search
- **THEN** the API is called with `community_div_id` parameter set to the selected community ID

#### Scenario: Search with district but no street/community
- **WHEN** the user selects a district but no street or community and clicks search
- **THEN** the API is called with `county_id` parameter set to the selected district ID
