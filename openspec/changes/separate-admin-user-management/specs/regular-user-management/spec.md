## ADDED Requirements

### Requirement: Display regular user list
The system SHALL provide a dedicated page that displays only regular (non-administrator) users in a paginated list.

#### Scenario: Regular user list loads successfully
- **WHEN** user navigates to the regular user management page
- **THEN** system displays a list of users with regular role only

#### Scenario: Pagination works correctly
- **WHEN** user clicks on page navigation controls
- **THEN** system loads the corresponding page of regular users

#### Scenario: Empty list handling
- **WHEN** there are no regular users in the system
- **THEN** system displays an appropriate empty state message

### Requirement: Search and filter regular users
The system SHALL allow users to search and filter the regular user list by common criteria.

#### Scenario: Search by username
- **WHEN** user enters a username in the search field
- **THEN** system filters the list to show only matching regular users

#### Scenario: Search by phone number
- **WHEN** user enters a phone number in the search field
- **THEN** system filters the list to show only matching regular users

### Requirement: View regular user details
The system SHALL display detailed information for each regular user in the list.

#### Scenario: Display user information
- **WHEN** regular user list is loaded
- **THEN** system displays username, phone, email, status, and creation time for each regular user

### Requirement: Edit regular user information
The system SHALL allow authorized users to modify regular user information.

#### Scenario: Open edit dialog
- **WHEN** user clicks the edit button for a regular user
- **THEN** system opens a dialog with editable fields pre-filled with current values

#### Scenario: Update regular user successfully
- **WHEN** user modifies fields and submits the form
- **THEN** system updates the regular user information and refreshes the list

#### Scenario: Validation on edit
- **WHEN** user submits invalid data (e.g., invalid email format)
- **THEN** system displays validation errors and prevents submission

### Requirement: Access control for regular user management
The system SHALL restrict access to the regular user management page based on permissions.

#### Scenario: Authorized access
- **WHEN** user with `identity:regular-user:list` permission navigates to the page
- **THEN** system displays the regular user management page

#### Scenario: Unauthorized access
- **WHEN** user without required permissions attempts to access the page
- **THEN** system denies access and displays an appropriate error message

#### Scenario: Edit permission check
- **WHEN** user without `identity:regular-user:update` permission views the page
- **THEN** system hides or disables the edit button
