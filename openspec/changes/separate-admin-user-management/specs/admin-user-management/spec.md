## ADDED Requirements

### Requirement: Display admin user list
The system SHALL provide a dedicated page that displays only administrator users in a paginated list.

#### Scenario: Admin user list loads successfully
- **WHEN** user navigates to the admin user management page
- **THEN** system displays a list of users with administrator role only

#### Scenario: Pagination works correctly
- **WHEN** user clicks on page navigation controls
- **THEN** system loads the corresponding page of admin users

#### Scenario: Empty list handling
- **WHEN** there are no admin users in the system
- **THEN** system displays an appropriate empty state message

### Requirement: Search and filter admin users
The system SHALL allow users to search and filter the admin user list by common criteria.

#### Scenario: Search by username
- **WHEN** user enters a username in the search field
- **THEN** system filters the list to show only matching admin users

#### Scenario: Search by phone number
- **WHEN** user enters a phone number in the search field
- **THEN** system filters the list to show only matching admin users

### Requirement: View admin user details
The system SHALL display detailed information for each admin user in the list.

#### Scenario: Display user information
- **WHEN** admin user list is loaded
- **THEN** system displays username, phone, email, status, and creation time for each admin user

### Requirement: Edit admin user information
The system SHALL allow authorized users to modify admin user information.

#### Scenario: Open edit dialog
- **WHEN** user clicks the edit button for an admin user
- **THEN** system opens a dialog with editable fields pre-filled with current values

#### Scenario: Update admin user successfully
- **WHEN** user modifies fields and submits the form
- **THEN** system updates the admin user information and refreshes the list

#### Scenario: Validation on edit
- **WHEN** user submits invalid data (e.g., invalid email format)
- **THEN** system displays validation errors and prevents submission

### Requirement: Access control for admin management
The system SHALL restrict access to the admin user management page based on permissions.

#### Scenario: Authorized access
- **WHEN** user with `identity:admin-user:list` permission navigates to the page
- **THEN** system displays the admin user management page

#### Scenario: Unauthorized access
- **WHEN** user without required permissions attempts to access the page
- **THEN** system denies access and displays an appropriate error message

#### Scenario: Edit permission check
- **WHEN** user without `identity:admin-user:update` permission views the page
- **THEN** system hides or disables the edit button
