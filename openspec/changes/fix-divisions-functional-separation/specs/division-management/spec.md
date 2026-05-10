## MODIFIED Requirements

### Requirement: Division page scope limited to province-city-district
The administrative division management page (`/masterdata/divisions`) SHALL only manage and display divisions at level 1 (province), level 2 (city), and level 3 (district). Divisions at level 4 (street) and level 5 (community) SHALL NOT appear in this page's tree table.

#### Scenario: Initial load shows only provinces
- **WHEN** the division page loads
- **THEN** the system SHALL fetch divisions with `level=1` and display them as root nodes

#### Scenario: Lazy load children limited to level 3
- **WHEN** user expands a city node in the division tree
- **THEN** the system SHALL only load and display districts (level 3) as children; streets and communities SHALL be filtered out

#### Scenario: No street or community data visible
- **WHEN** the division tree is fully expanded
- **THEN** no rows with level 4 or level 5 SHALL appear in the table

### Requirement: Division add/edit restricted to level 1-3
The division create/edit dialog SHALL only offer level options for province (1), city (2), and district (3). The "添加下级" button SHALL NOT appear on district rows (since their children are streets, which are managed in the grassroots page).

#### Scenario: Create new division level options
- **WHEN** user opens the create division dialog
- **THEN** the level selector SHALL only show options: 省级 (1), 市级 (2), 区县级 (3)

#### Scenario: No add-child button on districts
- **WHEN** the division tree renders a district (level 3) row
- **THEN** the "添加下级" button SHALL NOT appear in the action column

#### Scenario: Add-child button on provinces and cities
- **WHEN** the division tree renders a province (level 1) or city (level 2) row
- **THEN** the "添加下级" button SHALL appear and allow creating a child at the next level (2 or 3 respectively)
