## ADDED Requirements

### Requirement: Category dropdown in query form
The query form SHALL provide a dropdown list for category selection that corresponds to the `category` field in the `md_sensitive_word` database table.

#### Scenario: Display category dropdown
- **WHEN** user opens the sensitive word management page
- **THEN** the query form displays a category dropdown with all available categories

#### Scenario: Select category for filtering
- **WHEN** user selects a category from the dropdown
- **THEN** the system filters sensitive words by the selected category

#### Scenario: Clear category filter
- **WHEN** user clears the category selection
- **THEN** the system shows all sensitive words regardless of category

### Requirement: Category dropdown in add form
The add sensitive word form SHALL provide a dropdown list for category selection instead of a text input field.

#### Scenario: Display category dropdown in add form
- **WHEN** user opens the add sensitive word form
- **THEN** the form displays a category dropdown with all available categories

#### Scenario: Select category when adding
- **WHEN** user selects a category from the dropdown and submits the form
- **THEN** the system creates a new sensitive word with the selected category

#### Scenario: Category is required
- **WHEN** user attempts to submit the add form without selecting a category
- **THEN** the system displays a validation error requiring category selection

### Requirement: Category options source
The category dropdown options SHALL be populated from the database schema or a predefined list matching the `md_sensitive_word.category` field values.

#### Scenario: Load category options
- **WHEN** the form is initialized
- **THEN** the system loads all valid category values for the dropdown

#### Scenario: Consistent categories across forms
- **WHEN** category options are displayed in query and add forms
- **THEN** both forms show the same set of category options
