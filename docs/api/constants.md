# Constants & Enums Reference

This document provides all constants, enums, status codes, and validation rules used across the Community & Home Management System APIs.

## Error Codes

### HTTP Status Codes

| Code | Description | Usage |
|------|-------------|-------|
| 0 | Success | Operation completed successfully |
| 400 | Invalid Parameter | Request validation failed |
| 401 | Unauthorized | Missing or invalid authentication |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 500 | Internal Server Error | Unexpected server error |
| 501 | Database Error | Database operation failed |
| 502 | Cache Error | Redis/cache operation failed |
| 503 | RPC Error | RPC service call failed |

## Identity Service Constants

### User Types (`user_type`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Backend Staff | System administrators, property managers |
| 2 | Homeowner | Residential property owners |

### User Status (`status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Active | User can login and use system |
| 2 | Disabled | User account disabled by admin |
| 3 | Locked | User account locked (e.g., too many failed logins) |

### User Verification Status (`verification_status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 0 | Unverified | User has not submitted verification |
| 1 | Verified | User identity verified by admin |
| 2 | Rejected | Verification rejected by admin |

**Business Rule**: Users must have `verification_status = 1` before they can bind properties.

### Homeowner Verification Status (`verification_status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 0 | Pending | Awaiting admin review |
| 1 | Approved | Verification approved |
| 2 | Rejected | Verification rejected |

**Validation**: Maximum 9 document URLs per verification request.

### Property Binding Status (`bind_status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Active | Property binding is active |
| 2 | Pending | Property binding pending approval |
| 3 | Revoked | Property binding has been revoked |

### Property Primary Flag (`is_primary`)

| Value | Description | Notes |
|-------|-------------|-------|
| 0 | Secondary | Not the user's primary residence |
| 1 | Primary | User's primary residence |

### Property Type (`property_type`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Residential | Residential property |
| 2 | Commercial | Commercial property |
| 3 | Mixed | Mixed-use property |

### Property Unit Status (`status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Active | Unit is active |
| 2 | Inactive | Unit is inactive |

### Role Status (`status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Active | Role is active and can be assigned |
| 2 | Disabled | Role is disabled |

### System Role Flag (`is_system`)

| Value | Description | Notes |
|-------|-------------|-------|
| 0 | Custom Role | Can be modified or deleted |
| 1 | System Role | Cannot be deleted, protected |

### Permission Type (`type`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Menu | Menu/navigation permission |
| 2 | Button | Button/action permission |

### Permission Status (`status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Active | Permission is active |
| 2 | Disabled | Permission is disabled |

### Family Member Gender (`gender`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Male | Male |
| 2 | Female | Female |

## Masterdata Service Constants

### Administrative Division Level (`level`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Province | Provincial level |
| 2 | City | City level |
| 3 | District | District/county level |
| 4 | Street | Street/township level |
| 5 | Community | Community/village level |

**Business Rule**: Levels must follow hierarchy (1 → 2 → 3 → 4 → 5).

### Division Status (`status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Active | Division is active |
| 2 | Inactive | Division is inactive |

### Community Type (`community_type`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Residential | Residential community |
| 2 | Village | Rural village |
| 3 | Mixed | Mixed-use community |

### Community Submission Status (`submission_status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 0 | Draft | Community not yet submitted |
| 1 | Submitted | Awaiting admin review |
| 2 | Approved | Community approved |
| 3 | Rejected | Community rejected |

**Business Rules**:
- Can only submit from Draft (0) or Rejected (3) status
- Can only review Submitted (1) status communities

### Configuration Value Type (`value_type`)

| Value | Description | Notes |
|-------|-------------|-------|
| string | Text value | Default type |
| number | Numeric value | Integer or float |
| boolean | Boolean value | true/false |
| json | JSON object | Complex structured data |

### Configuration Public Flag (`is_public`)

| Value | Description | Notes |
|-------|-------------|-------|
| 0 | Internal | Not visible to all users |
| 1 | Public | Visible to all users |

### Configuration Approval Status (`approval_status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 0 | Draft | Configuration not yet approved |
| 1 | Pending Approval | Awaiting approval |
| 2 | Approved | Configuration approved |

### Sensitive Word Severity (`severity`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Low | Minor concern |
| 2 | Medium | Moderate concern |
| 3 | High | Serious concern |

### Sensitive Word Action (`action`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Warn | Show warning to user |
| 2 | Block | Block content submission |
| 3 | Review | Send to manual review |

### Sensitive Word Status (`status`)

| Value | Description | Notes |
|-------|-------------|-------|
| 1 | Active | Word is actively filtered |
| 2 | Inactive | Word is not filtered |

## Authentication & Token Configuration

### Token Expiration

| Token Type | Duration | Seconds | Notes |
|------------|----------|---------|-------|
| Access Token | 24 hours | 86400 | Used for API authentication |
| Refresh Token | 7 days | 604800 | Used to get new access token |

### Token Claims

| Claim | Type | Description |
|-------|------|-------------|
| userId | int64 | User ID |
| iat | int64 | Issued at timestamp |
| exp | int64 | Expiration timestamp |

## Validation Rules & Constraints

### Phone Number
- **Required**: Yes (for login and registration)
- **Format**: Chinese mobile phone format
- **Unique**: Must be unique in system
- **Usage**: Primary login identifier

### Password
- **Hashing**: bcrypt with DefaultCost
- **Required**: For password-based login
- **Optional**: For SMS-only registration
- **Min Length**: Recommended 8+ characters
- **Complexity**: Recommended mixed case + numbers + symbols

### Document URLs (Verification)
- **Minimum**: 1 document
- **Maximum**: 9 documents
- **Storage**: JSON array in database
- **Format**: Full URL to MinIO or external storage

### ID Card Number
- **Max Length**: 18 characters
- **Required**: For homeowner verification
- **Format**: Chinese ID card format

### Real Name
- **Max Length**: 50 characters
- **Required**: For homeowner verification

### Community Name
- **Max Length**: 100 characters
- **Required**: Yes

### Community Address
- **Max Length**: 255 characters
- **Required**: Yes

### Division Code
- **Max Length**: 20 characters
- **Unique**: Must be unique
- **Required**: Yes
- **Format**: Hierarchical code (e.g., "110000" for Beijing)

### Division Name
- **Max Length**: 100 characters
- **Required**: Yes

### Sensitive Word
- **Max Length**: 100 characters
- **Unique**: Must be unique
- **Required**: Yes

## Pagination Defaults

| Parameter | Default Value | Notes |
|-----------|---------------|-------|
| page | 1 | 1-indexed page number |
| page_size | 20 | Items per page |

## Audit Log Actions

| Action | Description | Usage |
|--------|-------------|-------|
| CREATE | Record creation | New entity created |
| UPDATE | Record update | Entity modified |
| DELETE | Record deletion | Entity deleted (soft delete) |
| REVIEW_approve | Community approval | Community approved by admin |
| REVIEW_reject | Community rejection | Community rejected by admin |

## Service Configuration

### Identity API
- **Host**: 0.0.0.0
- **Port**: 8888
- **Timeout**: 30000ms (30 seconds)

### Masterdata API
- **Host**: 0.0.0.0
- **Port**: 8889
- **Timeout**: 30000ms (30 seconds)

### MinIO Configuration
- **Bucket Name**: community-home
- **Endpoint**: localhost:9000
- **Access Key**: admin
- **Secret Key**: 12345678
- **Use SSL**: false

### Supported Entity Types (File Upload)
- `homeowner_verification` - Verification documents
- `user_avatar` - User avatar images
- `community_image` - Community photos

## Key Validation Points for Frontend

### User Creation
- Must specify `user_type` (1 or 2)
- Phone number must be unique
- Password required for password-based auth

### Community Submission
- Can only submit from Draft (0) or Rejected (3) status
- All required fields must be filled
- Division must exist and be active

### Community Review
- Can only review Submitted (1) status communities
- Action must be "approve" or "reject"
- Review notes are optional

### Property Binding
- User must have `verification_status = 1` before binding
- Property cannot already have active binding
- Property unit must exist

### Verification Review
- Can only review Pending (0) status verifications
- Status must be 1 (approved) or 2 (rejected)
- When approved, user's `verification_status` is automatically set to 1

### Division Hierarchy
- Levels must be 1-5 in order
- Parent division must exist for levels 2-5
- Cannot delete division with children

### Sensitive Words
- Severity and Action are required fields
- Word must be unique
- Category is optional but recommended

## Frontend Implementation Tips

### Status Display

```javascript
// User Status
const USER_STATUS = {
  1: { label: 'Active', color: 'green' },
  2: { label: 'Disabled', color: 'red' },
  3: { label: 'Locked', color: 'orange' }
};

// Verification Status
const VERIFICATION_STATUS = {
  0: { label: 'Pending', color: 'blue' },
  1: { label: 'Verified', color: 'green' },
  2: { label: 'Rejected', color: 'red' }
};

// Community Submission Status
const SUBMISSION_STATUS = {
  0: { label: 'Draft', color: 'gray' },
  1: { label: 'Submitted', color: 'blue' },
  2: { label: 'Approved', color: 'green' },
  3: { label: 'Rejected', color: 'red' }
};
```

### Form Validation

```javascript
// Phone validation
const isValidPhone = (phone) => /^1[3-9]\d{9}$/.test(phone);

// ID Card validation
const isValidIdCard = (id) => /^\d{17}[\dXx]$/.test(id);

// Password strength
const isStrongPassword = (pwd) => 
  pwd.length >= 8 && 
  /[A-Z]/.test(pwd) && 
  /[a-z]/.test(pwd) && 
  /\d/.test(pwd) &&
  /[!@#$%^&*]/.test(pwd);
```

### Enum Helpers

```javascript
// Check if user can bind property
const canBindProperty = (user) => user.verification_status === 1;

// Check if community can be submitted
const canSubmitCommunity = (community) => 
  [0, 3].includes(community.submission_status);

// Check if verification can be reviewed
const canReviewVerification = (verification) => 
  verification.verification_status === 0;
```
