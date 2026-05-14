## Context

The sensitive word management UI currently uses text inputs for category selection and a three-tier severity system (低/中/高). The form also includes a "处理动作" (handling action) field. This design leads to inconsistent data entry and unnecessary complexity. The database schema (`md_sensitive_word` table) has a `category` field that should be constrained to specific values, and the severity system needs simplification to two levels: 违规 (violation) and 可疑 (suspicious).

Current implementation is in `web/pc/src/views/sensitive-words/List.vue` using Vue 3 + Element Plus.

## Goals / Non-Goals

**Goals:**
- Replace category text input with dropdown in both query and add forms
- Simplify severity levels from three tiers (低/中/高) to two tiers (违规/可疑)
- Remove "处理动作" field from add/edit form
- Maintain existing submission workflow and permissions logic
- Ensure form validation aligns with new constraints

**Non-Goals:**
- Backend API changes (assuming API accepts the new severity values)
- Database schema modifications
- Changes to table display or other UI sections beyond query/add forms
- Migration of existing severity data

## Decisions

### Decision 1: Category dropdown source
**Choice:** Use a predefined constant array for category options in the frontend.

**Rationale:** 
- Categories are relatively stable and don't change frequently
- Avoids additional API call overhead
- Simpler implementation for initial version
- Can be refactored to API-driven if categories become dynamic

**Alternatives considered:**
- Fetch from API endpoint: More flexible but adds complexity and latency
- Read from database enum: Requires backend changes

### Decision 2: Severity level mapping
**Choice:** Map new two-tier system (违规/可疑) to numeric values for API compatibility.

**Mapping:**
- 违规 (violation) → value: 3 (high severity)
- 可疑 (suspicious) → value: 1 (low severity)

**Rationale:**
- Maintains backward compatibility with existing API
- 违规 represents serious violations (high severity)
- 可疑 represents suspicious content (low severity)
- Eliminates the ambiguous middle tier

**Alternatives considered:**
- New string values: Would require backend changes
- Keep numeric 1-3: Doesn't simplify the UI as required

### Decision 3: Remove handling action field
**Choice:** Remove "处理动作" field entirely from the form, do not send to API.

**Rationale:**
- User requirement explicitly states to remove this field
- Simplifies the form and reduces decision fatigue
- Handling actions may be determined by severity level on the backend

**Alternatives considered:**
- Set default value: Still adds unnecessary complexity
- Make it optional: Doesn't meet the requirement to remove it

### Decision 4: Form component structure
**Choice:** Modify existing `List.vue` component inline rather than creating separate components.

**Rationale:**
- Changes are localized to form fields
- Existing component structure is simple and maintainable
- No need for additional component abstraction at this stage

**Alternatives considered:**
- Extract form to separate component: Premature abstraction for this scope

## Risks / Trade-offs

**Risk:** Backend API may not accept the new severity values (违规/可疑 mapped to 3/1)
→ **Mitigation:** Test API integration thoroughly; if needed, coordinate backend changes to accept new values

**Risk:** Existing data with severity=2 (中) will not map to new two-tier system
→ **Mitigation:** This is a data migration concern outside this change's scope; document the mapping for future reference

**Risk:** Removing "处理动作" field may break existing workflows if backend requires it
→ **Mitigation:** Verify API contract; if required, set a default value on backend or make field optional

**Trade-off:** Hardcoded category options reduce flexibility
→ **Accepted:** Categories are stable; can be refactored to API-driven later if needed

## Migration Plan

1. **Deploy frontend changes** to staging environment
2. **Test scenarios:**
   - Query form with category dropdown and new severity filters
   - Add form with category dropdown, new severity options, no handling action
   - Verify API requests contain correct values
   - Test form validation
3. **Verify backward compatibility:** Existing records display correctly in table
4. **Rollback strategy:** Revert frontend deployment if API integration fails

## Open Questions

1. What are the valid category values for the dropdown? (e.g., 政治、色情、暴力、广告)
   - **Resolution needed:** Confirm with product/backend team or inspect existing data
2. Does the backend API require "处理动作" field, or can it be omitted?
   - **Resolution needed:** Review API documentation or test API calls
3. Should existing severity=2 records be displayed with a special indicator?
   - **Resolution needed:** Clarify with product team
