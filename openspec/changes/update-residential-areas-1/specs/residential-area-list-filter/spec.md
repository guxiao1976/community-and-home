## MODIFIED Requirements

### Requirement: Province-City-District cascading filter
The residential area list page filter area SHALL display three independent dropdown selectors: Province (level 1), City (level 2), and District/County (level 3). Selecting a province SHALL load its child cities; selecting a city SHALL load its child districts. Changing an upper-level selection SHALL clear all lower-level selections.

#### Scenario: Page load shows province dropdown
- **WHEN** the residential area list page is loaded
- **THEN** the province dropdown is populated with all level 1 administrative divisions
- **AND** the city and district dropdowns are empty and disabled

#### Scenario: Selecting a province loads cities
- **WHEN** the user selects a province (e.g., "北京市")
- **THEN** the city dropdown is populated with level 2 divisions under that province
- **AND** the district dropdown is cleared and remains disabled

#### Scenario: Selecting a city loads districts
- **WHEN** the user selects a city (e.g., "北京市")
- **THEN** the district dropdown is populated with level 3 divisions under that city

#### Scenario: Changing province clears city and district
- **WHEN** the user changes the province selection
- **THEN** the city and district selections are cleared
- **AND** the city dropdown is repopulated with divisions under the new province

### Requirement: City selection required before search
The search button SHALL be disabled when no city is selected. The user MUST select at least a city before executing a search.

#### Scenario: Search button disabled without city
- **WHEN** no city is selected
- **THEN** the search button is disabled (cannot be clicked)

#### Scenario: Search enabled after city selection
- **WHEN** the user selects a city (with or without selecting a district)
- **THEN** the search button becomes enabled

#### Scenario: Search with city but no district
- **WHEN** the user selects a city but does not select a district and clicks search
- **THEN** the system queries all residential areas in that city (county_id not sent, but city-level filtering is handled by loading all districts under the city and searching with the city's district IDs)

#### Scenario: Search with city and district
- **WHEN** the user selects both city and district and clicks search
- **THEN** the system queries residential areas filtered by the selected district's county_id
