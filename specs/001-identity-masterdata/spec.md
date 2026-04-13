# Feature Specification: Identity and Masterdata Microservices

**Feature Branch**: `001-identity-masterdata`  
**Created**: 2026-04-13  
**Status**: Draft  
**Input**: User description: "身份认证与主数据微服务 (identity-masterdata)"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Headquarters Administrator Manages Master Data (Priority: P1)

As a headquarters administrator, I need to manage the national five-tier administrative division data (province → city → district → street → community) and review community/village submissions from provincial/municipal administrators, so that the platform has accurate and consistent master data across all regions.

**Why this priority**: This is the foundational data layer that all other features depend on. Without accurate administrative divisions and approved communities, no other business operations can function properly.

**Independent Test**: Can be fully tested by creating administrative divisions, having provincial admins submit community data, and headquarters approving/rejecting submissions. Delivers the complete master data management capability.

**Acceptance Scenarios**:

1. **Given** I am logged in as a headquarters administrator, **When** I create a new province entry, **Then** the province appears in the administrative division hierarchy and is available for city-level entries
2. **Given** a provincial administrator has submitted a new community for approval, **When** I review and approve it, **Then** the community becomes active and visible to all users in that region
3. **Given** a provincial administrator has submitted a new community for approval, **When** I review and reject it with a reason, **Then** the community remains inactive and the provincial admin receives the rejection reason
4. **Given** I need to update an existing administrative division, **When** I modify the data, **Then** all dependent records reflect the change and an audit log is created
5. **Given** I need to delete a community, **When** I perform a soft delete, **Then** the community is marked as deleted but data is retained for audit purposes

---

### User Story 2 - Backend User Authentication and Authorization (Priority: P1)

As a backend user (headquarters/provincial/municipal/community/property staff), I need to register, log in with my phone number and password/verification code, and have my permissions automatically enforced based on my role and administrative scope, so that I can access only the data and functions appropriate to my position.

**Why this priority**: Authentication and authorization are critical security requirements that must be in place before any business operations can be performed. This is a blocking dependency for all other features.

**Independent Test**: Can be fully tested by registering users with different roles, logging in, and verifying that each user can only access data within their administrative scope. Delivers complete access control capability.

**Acceptance Scenarios**:

1. **Given** I am a new backend user, **When** I register with my phone number and receive a verification code, **Then** my account is created and I can log in
2. **Given** I am a registered user, **When** I log in with my phone number and password, **Then** I receive a JWT token and can access the system
3. **Given** I am a provincial administrator, **When** I attempt to view community data, **Then** I can only see communities within my province
4. **Given** I am a headquarters administrator, **When** I access any data, **Then** I can view and manage data across all regions
5. **Given** my role permissions have been updated, **When** I make my next API request, **Then** the new permissions are immediately enforced
6. **Given** I am logged in, **When** my JWT token expires, **Then** I am prompted to re-authenticate

---

### User Story 3 - Homeowner Identity Verification and Property Binding (Priority: P2)

As a homeowner, I need to upload my property ownership certificate for verification and bind my account to my property unit number, so that I can access community services and participate in community governance as a verified resident.

**Why this priority**: This enables homeowner participation in the platform, which is essential for community engagement features. However, the platform can function for administrative users without this feature.

**Independent Test**: Can be fully tested by a homeowner uploading property documents, headquarters reviewing and approving the verification, and the homeowner accessing their property-specific features. Delivers complete homeowner onboarding capability.

**Acceptance Scenarios**:

1. **Given** I am a registered homeowner, **When** I upload my property ownership certificate (JPG/PNG, max 5MB), **Then** the image is stored in MinIO and my verification request is submitted for review
2. **Given** my property verification is under review, **When** a headquarters administrator approves it, **Then** my account is marked as verified and I can bind to my property unit
3. **Given** I am a verified homeowner, **When** I bind my account to my property unit number, **Then** I can access services specific to my property and community
4. **Given** I am bound to a property unit, **When** another user attempts to bind to the same unit, **Then** they must request access and I receive a notification
5. **Given** I am the primary user of a property unit, **When** I need to remove another user from my unit, **Then** I can revoke their access and they lose property-specific permissions

---

### User Story 4 - Family Center Management (Priority: P3)

As a verified homeowner, I need to create a family profile, add family members, and manage family information, so that all household members can access appropriate community services and maintain accurate household records.

**Why this priority**: This enhances the homeowner experience but is not critical for core platform operations. It can be added after basic homeowner verification is working.

**Independent Test**: Can be fully tested by a verified homeowner creating a family, adding members, updating information, and verifying that family members have appropriate access. Delivers complete family management capability.

**Acceptance Scenarios**:

1. **Given** I am a verified homeowner, **When** I create a family profile, **Then** a family record is created and I am designated as the family head
2. **Given** I have a family profile, **When** I add a family member with their details, **Then** they are added to my family and can be invited to join the platform
3. **Given** I have family members in my profile, **When** I update their information, **Then** the changes are saved and reflected in all relevant services
4. **Given** I am a family member (not the head), **When** I access family information, **Then** I can view but not modify family records unless granted permission

---

### User Story 5 - Role and Permission Management (Priority: P2)

As a headquarters administrator, I need to create custom roles, assign menu-level and button-level permissions, and assign roles to backend users, so that I can implement fine-grained access control aligned with organizational structure.

**Why this priority**: While basic authentication (P1) provides initial access control, custom roles enable the platform to scale across different organizational structures and security requirements.

**Independent Test**: Can be fully tested by creating custom roles with specific permissions, assigning them to users, and verifying that users can only access permitted menus and actions. Delivers complete RBAC capability.

**Acceptance Scenarios**:

1. **Given** I am a headquarters administrator, **When** I create a new role with specific menu and button permissions, **Then** the role is available for assignment to users
2. **Given** I have created a custom role, **When** I assign it to a user, **Then** the user immediately gains the permissions defined in that role
3. **Given** a user has a role, **When** I modify the role's permissions, **Then** all users with that role immediately have the updated permissions
4. **Given** the system has a "Super Administrator" role, **When** I attempt to delete or disable it, **Then** the system prevents the action and displays an error message
5. **Given** I am managing roles, **When** I view the permission hierarchy, **Then** I can see all available menus and buttons organized by module

---

### User Story 6 - Platform Configuration Management (Priority: P3)

As a headquarters administrator, I need to manage platform-wide configurable parameters (organized by module), AI review rules, sensitive word lists, evaluation rules, and blacklist rules, so that the platform behavior can be customized without code changes.

**Why this priority**: This provides operational flexibility but is not required for core functionality. Initial deployments can use default configurations.

**Independent Test**: Can be fully tested by creating and modifying configuration parameters, applying them to different modules, and verifying that the changes take effect. Delivers complete configuration management capability.

**Acceptance Scenarios**:

1. **Given** I am a headquarters administrator, **When** I create a new configuration parameter for a specific module, **Then** the parameter is available for that module and can be updated
2. **Given** I need to update AI review rules, **When** I modify the rule configuration, **Then** the changes are submitted for approval before taking effect
3. **Given** I manage the sensitive word list, **When** I add new words, **Then** they are immediately applied to content moderation across the platform
4. **Given** I configure evaluation rules, **When** I set rating thresholds and criteria, **Then** the evaluation system uses these rules for all assessments
5. **Given** I manage the blacklist, **When** I add a user or entity to the blacklist, **Then** they are restricted according to the blacklist rules

---

### Edge Cases

- What happens when a user's administrative scope changes (e.g., provincial admin reassigned to different province)?
- How does the system handle concurrent approval requests for the same community submission?
- What happens when a homeowner's property ownership is disputed or revoked?
- How does the system handle JWT token refresh when the user's permissions change mid-session?
- What happens when a family head transfers ownership to another family member?
- How does the system handle image upload failures or corrupted files?
- What happens when a user attempts to bind to a property unit that is already at maximum capacity?
- How does the system handle deletion of administrative divisions that have dependent data?
- What happens when configuration changes conflict with existing data or business rules?

## Requirements *(mandatory)*

### Functional Requirements

**Authentication & Authorization**

- **FR-001**: System MUST support user registration via phone number with SMS verification code
- **FR-002**: System MUST support user login via phone number with password or verification code
- **FR-003**: System MUST generate and validate JWT tokens for authenticated sessions
- **FR-004**: System MUST enforce role-based access control (RBAC) at menu and button levels
- **FR-005**: System MUST restrict provincial/municipal administrators to view and manage only data within their administrative scope
- **FR-006**: System MUST provide a "Super Administrator" role with all permissions that cannot be deleted or disabled
- **FR-007**: System MUST apply permission changes immediately without requiring user re-login

**Master Data Management**

- **FR-008**: System MUST support five-tier administrative division hierarchy (province → city → district → street → community)
- **FR-009**: System MUST allow only headquarters administrators to create, update, and delete administrative division data
- **FR-010**: System MUST allow provincial/municipal administrators to submit community/village data for headquarters approval
- **FR-011**: System MUST require headquarters approval before community/village submissions become active
- **FR-012**: System MUST support soft delete for all core master data with delete_time field
- **FR-013**: System MUST record audit logs for all master data changes including user, timestamp, and change details
- **FR-014**: System MUST support county/district economic and population data maintenance
- **FR-015**: System MUST provide platform-wide configurable parameters organized by module
- **FR-016**: System MUST support AI review rules and sensitive word list management
- **FR-017**: System MUST support evaluation and blacklist rule configuration
- **FR-018**: System MUST require approval workflow for all configuration changes

**Homeowner Verification & Property Management**

- **FR-019**: System MUST allow homeowners to upload property ownership certificates (JPG/PNG format, max 5MB per image, max 9 images per upload)
- **FR-020**: System MUST store uploaded images in MinIO object storage
- **FR-021**: System MUST provide headquarters administrators with interface to review and approve/reject homeowner verification requests
- **FR-022**: System MUST allow verified homeowners to bind their accounts to property unit numbers
- **FR-023**: System MUST support multiple users per property unit with primary user designation
- **FR-024**: System MUST allow primary property users to remove other users from their unit
- **FR-025**: System MUST notify existing users when new users request access to their property unit

**Family Management**

- **FR-026**: System MUST allow verified homeowners to create family profiles
- **FR-027**: System MUST support adding, updating, and removing family members
- **FR-028**: System MUST designate one family member as the family head with administrative privileges
- **FR-029**: System MUST maintain family information records accessible to all family members

**API Standards**

- **FR-030**: System MUST provide RESTful APIs with consistent response format
- **FR-031**: System MUST support pagination with page (page number) and pageSize (items per page) parameters
- **FR-032**: System MUST use RFC3339 format for all timestamp fields (2006-01-02T15:04:05Z07:00)
- **FR-033**: System MUST validate all user input and sanitize data before processing
- **FR-034**: System MUST use bcrypt for password encryption

### Key Entities

- **User**: Represents all platform users (backend staff and homeowners) with phone number, password hash, role assignments, administrative scope, and verification status
- **Role**: Defines a set of permissions with name, description, menu permissions, button permissions, and system role flag (for Super Administrator)
- **Permission**: Represents individual access rights organized by menu and button level
- **AdministrativeDivision**: Five-tier hierarchy (province, city, district, street, community) with parent-child relationships, geographic codes, and status
- **Community**: Represents residential communities or villages with basic information, administrative division reference, submission status, and approval workflow
- **PropertyUnit**: Represents individual property units within communities with unit number, building, floor, and community reference
- **PropertyBinding**: Links users to property units with primary user flag and binding status
- **HomeownerVerification**: Tracks homeowner identity verification requests with uploaded document references, review status, and reviewer notes
- **Family**: Represents household units with family head designation and member list
- **FamilyMember**: Individual family member records with relationship to family head and personal information
- **Configuration**: Platform-wide configurable parameters organized by module with key-value pairs and approval status
- **AuditLog**: Records all significant data changes with user, timestamp, entity type, entity ID, action, and change details
- **UploadedFile**: Tracks files stored in MinIO with file path, size, format, upload timestamp, and associated entity

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Backend users can complete registration and login within 1 minute
- **SC-002**: Provincial administrators can only access data within their assigned province with 100% enforcement
- **SC-003**: Permission changes take effect within 5 seconds without requiring user re-login
- **SC-004**: Homeowners can complete property verification submission within 3 minutes
- **SC-005**: Headquarters administrators can review and approve/reject verification requests within 2 minutes per request
- **SC-006**: System supports 10,000 concurrent users without performance degradation
- **SC-007**: All API responses return within 200ms at P99 latency
- **SC-008**: Image uploads to MinIO complete within 10 seconds for files up to 5MB
- **SC-009**: 95% of users successfully complete their primary task on first attempt
- **SC-010**: All master data changes are recorded in audit logs with 100% accuracy
- **SC-011**: System maintains 99.9% uptime for authentication services
- **SC-012**: Zero unauthorized access incidents to data outside user's administrative scope

## Assumptions

- Users have stable internet connectivity for image uploads
- SMS verification code delivery is handled by an external SMS gateway service (integration details to be defined in planning phase)
- Property ownership certificates are issued by recognized authorities and follow standard formats
- The platform will initially support Chinese language only (internationalization out of scope for v1)
- Mobile app support is out of scope; this specification covers backend services and web-based admin interface
- Email notifications are out of scope for v1; SMS is the primary notification channel
- Payment processing for community fees is handled by a separate finance microservice (out of scope for this feature)
- The Vue 3 frontend will be developed separately based on the API contracts defined in the planning phase
- MinIO is already deployed and accessible (connection details to be provided during implementation)
- Etcd cluster is already configured for service discovery
- Redis cluster is already configured for caching and session management
- MySQL database is already provisioned with appropriate capacity for 100k communities and 1M users
