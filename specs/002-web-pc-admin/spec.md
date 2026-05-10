# Feature Specification: Web PC Admin Frontend Development

**Feature Branch**: `002-web-pc-admin`  
**Created**: 2026-05-03  
**Status**: Draft  
**Input**: User description: "请基于项目用户故事、及后端已经开发完的功能，接口文档见docs/api,以及项目宪法文件内容，规划项目前端开发，目前的主数据管理、用户权限管理都是web pc端。"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Authentication & Session Management (Priority: P1)

Headquarters administrators and provincial/municipal staff need to securely access the management platform using their registered phone numbers and passwords or SMS verification codes. The system must maintain their login session for 24 hours and automatically refresh tokens to prevent interruptions during work.

**Why this priority**: Without authentication, no other features can be accessed. This is the foundation for all administrative operations and must be implemented first.

**Independent Test**: Can be fully tested by registering a new admin account, logging in with phone/password, verifying token storage, and confirming automatic token refresh works before the 24-hour expiration. Delivers immediate value by securing platform access.

**Acceptance Scenarios**:

1. **Given** an unregistered phone number, **When** admin enters phone, password, SMS code and clicks register, **Then** account is created and admin is logged in with valid JWT tokens
2. **Given** a registered admin account, **When** admin enters correct phone and password and clicks login, **Then** system returns access token (24h) and refresh token (7d) and redirects to dashboard
3. **Given** a registered admin account, **When** admin requests SMS code and enters correct code, **Then** system authenticates via SMS and returns valid tokens
4. **Given** an active session with token expiring in 5 minutes, **When** admin performs any action, **Then** system automatically refreshes the access token without interrupting the workflow
5. **Given** an authenticated admin, **When** admin clicks logout, **Then** tokens are invalidated and admin is redirected to login page
6. **Given** an expired or invalid token, **When** admin attempts any action, **Then** system redirects to login page with appropriate error message

---

### User Story 2 - Administrative Division Management (Priority: P1)

Headquarters administrators need to create and manage the five-tier administrative division hierarchy (province → city → district → street → community) that forms the foundation of the platform's governance structure. They must be able to view divisions as a tree, add new divisions at any level, edit division details, and prevent deletion of divisions that have child divisions or associated communities.

**Why this priority**: Administrative divisions are master data required by all other features (user scope restrictions, community assignments, property locations). Must be implemented immediately after authentication.

**Independent Test**: Can be fully tested by logging in as headquarters admin, creating a complete division hierarchy from province to community level, editing division names, attempting to delete divisions with/without children, and verifying the tree structure displays correctly. Delivers value by establishing the governance foundation.

**Acceptance Scenarios**:

1. **Given** headquarters admin is logged in, **When** admin navigates to division management, **Then** system displays existing divisions in a collapsible tree view with levels clearly indicated
2. **Given** admin is viewing the division tree, **When** admin clicks "Add Division" under a parent node, **Then** system opens a form to create a child division at the next level (e.g., city under province)
3. **Given** admin is creating a new division, **When** admin enters name, code, and sort order and clicks save, **Then** division is created and appears in the tree under the correct parent
4. **Given** admin is viewing a division, **When** admin clicks edit and updates the name or code, **Then** changes are saved and reflected immediately in the tree
5. **Given** a division has child divisions or associated communities, **When** admin attempts to delete it, **Then** system prevents deletion and shows error message listing dependencies
6. **Given** a leaf division with no dependencies, **When** admin clicks delete and confirms, **Then** division is soft-deleted and removed from the tree view

---

### User Story 3 - Community Management & Review Workflow (Priority: P1)

Provincial and municipal administrators need to create community records under their assigned administrative divisions, submit them for headquarters review, and track submission status. Headquarters administrators need to review submitted communities, approve or reject them with notes, and ensure only approved communities are active in the system.

**Why this priority**: Communities are the core organizational unit for the platform. Without community management, property units, homeowners, and all community-level features cannot function. This implements the two-tier governance model.

**Independent Test**: Can be fully tested by logging in as provincial admin, creating a community under their division, submitting it for review, then logging in as headquarters admin, reviewing the submission, approving/rejecting with notes, and verifying status updates. Delivers value by enabling community onboarding.

**Acceptance Scenarios**:

1. **Given** provincial admin is logged in, **When** admin navigates to community management, **Then** system displays communities within their administrative scope with filters for division and status
2. **Given** admin is creating a new community, **When** admin selects division, enters name, address, area, population, community type and clicks save, **Then** community is created in Draft status
3. **Given** a community in Draft or Rejected status, **When** admin clicks "Submit for Review", **Then** community status changes to Submitted and appears in headquarters review queue
4. **Given** headquarters admin is viewing submitted communities, **When** admin clicks review on a community, **Then** system displays full community details with approve/reject options
5. **Given** headquarters admin is reviewing a community, **When** admin clicks approve with optional notes, **Then** community status changes to Approved and submitter is notified
6. **Given** headquarters admin is reviewing a community, **When** admin clicks reject and enters rejection notes, **Then** community status changes to Rejected, notes are saved, and submitter can view the feedback
7. **Given** a community in Approved status, **When** provincial admin attempts to edit or delete it, **Then** system prevents modification and shows message that only headquarters can modify approved communities

---

### User Story 4 - User Management (Priority: P2)

Administrators need to view all registered users (staff and homeowners), create new staff accounts, edit user profiles, disable/enable user accounts, and filter users by type, status, and verification status. They must be able to see each user's assigned roles and administrative scope.

**Why this priority**: User management is essential for platform administration but can be implemented after the foundational master data (divisions, communities) is in place. Staff accounts can be manually created initially.

**Independent Test**: Can be fully tested by logging in as admin, viewing the user list with pagination, creating a new staff account, editing user details, disabling/enabling accounts, filtering by user type and status, and viewing user role assignments. Delivers value by enabling user administration.

**Acceptance Scenarios**:

1. **Given** admin is logged in, **When** admin navigates to user management, **Then** system displays paginated user list with phone, nickname, user type, status, verification status, and last login
2. **Given** admin is viewing user list, **When** admin applies filters for user type, status, or verification status, **Then** list updates to show only matching users
3. **Given** admin clicks "Create User", **When** admin enters phone, password, nickname, user type, and scope and clicks save, **Then** new user account is created and appears in the list
4. **Given** admin is viewing a user, **When** admin clicks edit and updates nickname, avatar, or scope, **Then** changes are saved and reflected in the user profile
5. **Given** an active user account, **When** admin clicks disable, **Then** user status changes to disabled and user cannot log in
6. **Given** a disabled user account, **When** admin clicks enable, **Then** user status changes to active and user can log in again
7. **Given** admin is viewing a user, **When** admin clicks "View Permissions", **Then** system displays all roles assigned to the user and their effective permissions

---

### User Story 5 - Role & Permission Management (Priority: P2)

Administrators need to create custom roles, assign menu and button-level permissions to roles, assign roles to users, and manage the permission hierarchy. The system must protect the built-in Super Administrator role from deletion and ensure permission changes take effect immediately.

**Why this priority**: Role-based access control is critical for security but can be implemented after basic user management. Initial deployments can use the Super Administrator role until custom roles are needed.

**Independent Test**: Can be fully tested by logging in as Super Administrator, creating a new role (e.g., "Community Manager"), assigning specific menu and button permissions, assigning the role to a test user, logging in as that user, and verifying they only see permitted menus and buttons. Delivers value by enabling fine-grained access control.

**Acceptance Scenarios**:

1. **Given** admin is logged in, **When** admin navigates to role management, **Then** system displays all roles with name, code, description, system flag, and status
2. **Given** admin clicks "Create Role", **When** admin enters name, code, description and clicks save, **Then** new role is created with no permissions assigned
3. **Given** admin is viewing a role, **When** admin clicks "Assign Permissions", **Then** system displays permission tree with checkboxes for menu and button permissions
4. **Given** admin is assigning permissions, **When** admin checks/unchecks permissions and clicks save, **Then** role permissions are updated and users with this role immediately gain/lose access
5. **Given** admin is viewing a custom role, **When** admin clicks edit and updates name or description, **Then** changes are saved and reflected in the role list
6. **Given** admin is viewing a custom role with no assigned users, **When** admin clicks delete and confirms, **Then** role is soft-deleted and removed from the list
7. **Given** admin is viewing the Super Administrator role, **When** admin attempts to delete it, **Then** system prevents deletion and shows error message that system roles cannot be deleted
8. **Given** admin is viewing a user, **When** admin assigns/removes roles and clicks save, **Then** user's role assignments are updated and their permissions change immediately

---

### User Story 6 - Homeowner Verification Review (Priority: P2)

Administrators need to review homeowner verification requests submitted by residents, view uploaded property documents (up to 9 images), approve or reject verifications with notes, and track verification status. Approved homeowners can then bind to property units.

**Why this priority**: Homeowner verification is essential for property binding but can be implemented after the core administrative features. Initial testing can use manually verified accounts.

**Independent Test**: Can be fully tested by having a test homeowner submit verification with property documents, logging in as admin, viewing the verification request with document previews, approving/rejecting with notes, and verifying the homeowner's verification status updates. Delivers value by enabling homeowner onboarding.

**Acceptance Scenarios**:

1. **Given** admin is logged in, **When** admin navigates to verification management, **Then** system displays all verification requests with user info, submission time, and status (pending/approved/rejected)
2. **Given** admin is viewing verification list, **When** admin filters by status or date range, **Then** list updates to show only matching verifications
3. **Given** admin clicks on a pending verification, **When** verification details load, **Then** system displays homeowner's real name, ID card number (partially masked), property unit, and uploaded document images (up to 9)
4. **Given** admin is viewing verification documents, **When** admin clicks on a document thumbnail, **Then** system opens full-size image in a modal for detailed review
5. **Given** admin is reviewing a verification, **When** admin clicks approve with optional notes and confirms, **Then** verification status changes to Approved, homeowner's verification_status becomes 1, and homeowner can now bind properties
6. **Given** admin is reviewing a verification, **When** admin clicks reject, enters rejection notes, and confirms, **Then** verification status changes to Rejected, notes are saved, and homeowner can view the feedback and resubmit
7. **Given** admin is viewing an approved or rejected verification, **When** admin views the details, **Then** system displays reviewer name, review time, and review notes

---

### User Story 7 - System Configuration Management (Priority: P3)

Administrators need to manage system-wide configuration parameters organized by module, including string, number, boolean, and JSON values. They must be able to create, edit, and delete configurations, mark configurations as public or private, and track configuration approval status.

**Why this priority**: Configuration management is important for system flexibility but not critical for initial launch. Default configurations can be hardcoded initially and moved to the configuration system later.

**Independent Test**: Can be fully tested by logging in as admin, creating a new configuration (e.g., "max_upload_size" = 10MB), editing its value, marking it as public, and verifying the configuration is available to the system. Delivers value by enabling runtime configuration changes without code deployment.

**Acceptance Scenarios**:

1. **Given** admin is logged in, **When** admin navigates to configuration management, **Then** system displays all configurations grouped by module with key, value, type, and status
2. **Given** admin clicks "Create Configuration", **When** admin enters module, key, value, type (string/number/boolean/json), description, and public flag and clicks save, **Then** new configuration is created and appears in the list
3. **Given** admin is viewing a configuration, **When** admin clicks edit and updates the value, **Then** changes are saved and take effect immediately (or after cache refresh)
4. **Given** admin is viewing a configuration, **When** admin toggles the "public" flag, **Then** configuration visibility changes (public configs may be exposed to frontend)
5. **Given** admin is viewing a configuration with no dependencies, **When** admin clicks delete and confirms, **Then** configuration is soft-deleted and removed from the list
6. **Given** admin is viewing configurations, **When** admin filters by module or searches by key, **Then** list updates to show only matching configurations

---

### User Story 8 - Sensitive Word Management (Priority: P3)

Administrators need to manage a sensitive word list for content filtering, including adding new words, categorizing them, setting severity levels (1-3), defining actions (warn/block/review), and enabling/disabling words. This supports content moderation across the platform.

**Why this priority**: Sensitive word filtering is important for content safety but not critical for initial launch. Manual moderation can be used initially until the word list is built up.

**Independent Test**: Can be fully tested by logging in as admin, adding sensitive words with different categories and severity levels, editing word settings, disabling words, and verifying the word list is available for content filtering. Delivers value by enabling automated content moderation.

**Acceptance Scenarios**:

1. **Given** admin is logged in, **When** admin navigates to sensitive word management, **Then** system displays all sensitive words with word, category, severity, action, and status
2. **Given** admin clicks "Add Sensitive Word", **When** admin enters word, category, severity (1-3), action (warn/block/review) and clicks save, **Then** new word is added to the list and becomes active for filtering
3. **Given** admin is viewing a sensitive word, **When** admin clicks edit and updates category, severity, or action, **Then** changes are saved and filtering behavior updates immediately
4. **Given** an active sensitive word, **When** admin clicks disable, **Then** word status changes to disabled and is no longer used for filtering
5. **Given** a disabled sensitive word, **When** admin clicks enable, **Then** word status changes to active and is used for filtering again
6. **Given** admin is viewing sensitive words, **When** admin filters by category or severity, **Then** list updates to show only matching words
7. **Given** admin is viewing a sensitive word with no usage history, **When** admin clicks delete and confirms, **Then** word is soft-deleted and removed from the list

---

### Edge Cases

- What happens when an admin's token expires during a long-running operation (e.g., uploading multiple verification documents)?
- How does the system handle concurrent edits to the same division or community by multiple administrators?
- What happens when an admin tries to create a division with a duplicate code at the same level?
- How does the system handle verification document uploads that fail due to network issues or file size limits?
- What happens when an admin assigns a role to a user but the role is deleted before the user logs in again?
- How does the system handle pagination when the underlying data changes (e.g., new communities added while admin is viewing page 2)?
- What happens when an admin tries to approve a community that was deleted by another admin?
- How does the system handle special characters or very long names in division/community names?
- What happens when an admin's administrative scope is changed while they are actively using the system?
- How does the system handle image formats that are not supported for verification documents?

## Requirements *(mandatory)*

### Functional Requirements

**Authentication & Authorization**
- **FR-001**: System MUST provide login interface accepting phone number and password with validation for Chinese mobile format (1[3-9]\d{9})
- **FR-002**: System MUST provide SMS login interface with SMS code request and verification
- **FR-003**: System MUST provide registration interface with phone, password, SMS verification, and nickname
- **FR-004**: System MUST store JWT access token (24h) and refresh token (7d) securely in browser storage
- **FR-005**: System MUST automatically refresh access token before expiration without user interruption
- **FR-006**: System MUST redirect to login page when token is invalid or expired (401 response)
- **FR-007**: System MUST provide logout functionality that invalidates tokens and clears local storage
- **FR-008**: System MUST verify user permissions before rendering menu items and action buttons
- **FR-009**: System MUST restrict access to pages and actions based on user's assigned roles and permissions

**Administrative Division Management**
- **FR-010**: System MUST display administrative divisions in a collapsible tree view with five levels (province, city, district, street, community)
- **FR-011**: System MUST allow headquarters admin to create new divisions at any level with name, code, and sort order
- **FR-012**: System MUST validate that division level matches parent level + 1 (e.g., city under province)
- **FR-013**: System MUST allow editing division name, code, and sort order
- **FR-014**: System MUST prevent deletion of divisions that have child divisions or associated communities
- **FR-015**: System MUST allow soft deletion of leaf divisions with no dependencies
- **FR-016**: System MUST display division path (e.g., "Guangdong > Shenzhen > Nanshan") in breadcrumb format

**Community Management**
- **FR-017**: System MUST display communities in a paginated table with filters for division, submission status, and community type
- **FR-018**: System MUST allow provincial/municipal admin to create communities within their administrative scope
- **FR-019**: System MUST require division, name, address, area, population, and community type when creating communities
- **FR-020**: System MUST allow editing community details when status is Draft or Rejected
- **FR-021**: System MUST allow submitting communities for review when status is Draft or Rejected
- **FR-022**: System MUST display submitted communities in headquarters admin's review queue
- **FR-023**: System MUST allow headquarters admin to approve communities with optional notes
- **FR-024**: System MUST allow headquarters admin to reject communities with required rejection notes
- **FR-025**: System MUST prevent provincial/municipal admin from editing or deleting Approved communities
- **FR-026**: System MUST display submission status, submitter, reviewer, and review notes for each community

**User Management**
- **FR-027**: System MUST display users in a paginated table with phone, nickname, user type, status, verification status, and last login
- **FR-028**: System MUST provide filters for user type (staff/homeowner), status (active/disabled), and verification status
- **FR-029**: System MUST allow admin to create new staff accounts with phone, password, nickname, user type, and administrative scope
- **FR-030**: System MUST allow editing user nickname, avatar URL, and administrative scope
- **FR-031**: System MUST allow disabling and enabling user accounts
- **FR-032**: System MUST prevent deletion of user accounts (soft delete only)
- **FR-033**: System MUST display user's assigned roles and effective permissions in a detail view

**Role & Permission Management**
- **FR-034**: System MUST display roles in a table with name, code, description, system flag, and status
- **FR-035**: System MUST allow creating custom roles with name, code, and description
- **FR-036**: System MUST display permissions in a tree view with parent-child relationships
- **FR-037**: System MUST allow assigning menu and button permissions to roles via checkboxes
- **FR-038**: System MUST prevent deletion of system roles (is_system = true)
- **FR-039**: System MUST allow soft deletion of custom roles with no assigned users
- **FR-040**: System MUST allow assigning multiple roles to a user
- **FR-041**: System MUST immediately apply permission changes when roles are updated

**Homeowner Verification Review**
- **FR-042**: System MUST display verification requests in a paginated table with user info, submission time, and status
- **FR-043**: System MUST provide filters for verification status (pending/approved/rejected) and date range
- **FR-044**: System MUST display verification details including real name, partially masked ID card, property unit, and document images
- **FR-045**: System MUST display up to 9 document images as thumbnails with click-to-enlarge functionality
- **FR-046**: System MUST allow admin to approve verifications with optional notes
- **FR-047**: System MUST allow admin to reject verifications with required rejection notes
- **FR-048**: System MUST display reviewer name, review time, and review notes for processed verifications
- **FR-049**: System MUST update homeowner's verification_status to 1 upon approval

**Configuration Management**
- **FR-050**: System MUST display configurations grouped by module with key, value, type, and status
- **FR-051**: System MUST allow creating configurations with module, key, value, type (string/number/boolean/json), description, and public flag
- **FR-052**: System MUST validate value format based on selected type
- **FR-053**: System MUST allow editing configuration values
- **FR-054**: System MUST allow toggling public/private flag for configurations
- **FR-055**: System MUST allow soft deletion of configurations with no dependencies
- **FR-056**: System MUST provide search and filter by module and key

**Sensitive Word Management**
- **FR-057**: System MUST display sensitive words in a paginated table with word, category, severity, action, and status
- **FR-058**: System MUST allow adding sensitive words with category, severity (1-3), and action (warn/block/review)
- **FR-059**: System MUST allow editing word category, severity, and action
- **FR-060**: System MUST allow enabling and disabling sensitive words
- **FR-061**: System MUST allow soft deletion of sensitive words with no usage history
- **FR-062**: System MUST provide filters for category and severity

**UI/UX Requirements**
- **FR-063**: System MUST use Element Plus component library for all UI components
- **FR-064**: System MUST display loading states during API calls
- **FR-065**: System MUST display success/error messages using Element Plus notifications
- **FR-066**: System MUST validate all form inputs before submission
- **FR-067**: System MUST display validation errors inline on form fields
- **FR-068**: System MUST provide breadcrumb navigation for hierarchical data (divisions, permissions)
- **FR-069**: System MUST support pagination with configurable page size (default 20)
- **FR-070**: System MUST provide search and filter capabilities on all list views
- **FR-071**: System MUST desensitize sensitive data (phone numbers, ID cards) in list views
- **FR-072**: System MUST provide confirmation dialogs for destructive actions (delete, disable, reject)

### Key Entities

- **User**: Represents platform users (staff and homeowners) with phone, password, nickname, avatar, user type, status, verification status, administrative scope, and last login tracking
- **Role**: Represents permission groups with name, code, description, system flag, and status; many-to-many relationship with users and permissions
- **Permission**: Represents access control rules in a hierarchical tree structure with name, code, type (menu/button), parent-child relationships, path, and sort order
- **AdministrativeDivision**: Represents five-tier geographic hierarchy (province → city → district → street → community) with parent-child relationships, level, name, code, path, and sort order
- **Community**: Represents residential communities with division association, name, address, area, population, community type, submission workflow status, submitter/reviewer tracking, and review notes
- **HomeownerVerification**: Represents verification requests with user association, property unit, real name, ID card, document URLs (up to 9), verification status, reviewer tracking, and review notes
- **Configuration**: Represents system-wide settings with module grouping, key-value pairs, value type, description, public flag, and approval status
- **SensitiveWord**: Represents content filtering rules with word, category, severity (1-3), action (warn/block/review), and status

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Administrators can complete login and access the dashboard in under 10 seconds
- **SC-002**: Administrators can create a complete five-tier division hierarchy (province to community) in under 5 minutes
- **SC-003**: Provincial administrators can create and submit a community for review in under 3 minutes
- **SC-004**: Headquarters administrators can review and approve/reject a community submission in under 2 minutes
- **SC-005**: Administrators can create a new role and assign permissions in under 3 minutes
- **SC-006**: Administrators can review a homeowner verification with document images in under 2 minutes
- **SC-007**: System maintains responsive performance (page load under 2 seconds) with up to 10,000 divisions and 100,000 communities
- **SC-008**: Token refresh occurs automatically without user awareness or workflow interruption
- **SC-009**: 95% of form submissions succeed on first attempt with clear validation feedback
- **SC-010**: Administrators can find specific records using search and filters in under 30 seconds
- **SC-011**: All list views support pagination and load pages in under 1 second
- **SC-012**: Permission changes take effect immediately (within 5 seconds) after role updates

## Assumptions

- Users (administrators) have stable internet connectivity and use modern browsers (Chrome, Firefox, Edge, Safari latest versions)
- Backend APIs documented in docs/api are fully functional and return data in the specified format
- Backend handles all business logic validation; frontend performs basic input validation only
- Administrative scope restrictions are enforced by backend; frontend displays scope-appropriate data
- File uploads for verification documents are handled by backend MinIO integration; frontend only sends file data
- Masterdata Service JWT authentication will be implemented by backend before production deployment
- Default test credentials (phone: 13800000000, password: Admin@123456) are available for development and testing
- Mobile responsive design is out of scope; this is a desktop-only PC admin interface
- Internationalization (i18n) is out of scope for initial version; all UI text is in Chinese
- Dark mode and theme customization are out of scope for initial version
- Real-time notifications (WebSocket) are out of scope; users must refresh to see updates from other administrators
- Audit logging is handled by backend; frontend does not need to implement audit trail UI
- Print functionality for reports is out of scope for initial version
- Export to Excel/CSV is out of scope for initial version
- Advanced search with complex query builders is out of scope; basic filters are sufficient
- Offline mode and PWA features are out of scope
