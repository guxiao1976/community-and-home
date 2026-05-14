## Why

The current sensitive word management UI lacks proper form controls and has inconsistent severity level definitions. Users need dropdown selections for categories (matching database schema) and a simplified two-tier severity system (违规/可疑) for better usability and data consistency.

## What Changes

- Replace category text input with dropdown list in query conditions (sourced from `md_sensitive_word.category` field)
- Simplify severity levels to two options: 违规 (violation) and 可疑 (suspicious) in both query and add forms
- Remove "处理动作" (handling action) field from the add sensitive word form
- Update add sensitive word form to use category dropdown instead of text input
- Ensure form validation aligns with the new two-tier severity system

## Capabilities

### New Capabilities
- `sensitive-word-category-dropdown`: Category selection using dropdown populated from database schema
- `sensitive-word-severity-levels`: Two-tier severity classification system (违规, 可疑)

### Modified Capabilities
<!-- No existing capabilities are being modified - this is a new UI feature -->

## Impact

- **Frontend**: Sensitive word management page (`src/views/masterdata/sensitive-word/` or similar)
- **Forms**: Query form and add/edit forms need dropdown components
- **API**: May need endpoint to fetch category options if not hardcoded
- **Database**: Uses existing `md_sensitive_word` table structure (no schema changes)
- **User Experience**: Improved data consistency and easier form completion
