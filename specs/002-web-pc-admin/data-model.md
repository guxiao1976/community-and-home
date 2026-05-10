# Data Model: Web PC Admin Frontend

**Feature**: 002-web-pc-admin  
**Date**: 2026-05-03  
**Status**: Complete

## Overview

This document defines the frontend data model for the PC admin application. All entities are synchronized with backend API responses documented in `docs/api/`. The frontend uses TypeScript interfaces that mirror backend Go structs and protobuf definitions.

## Core Entities

### 1. User

Represents platform users (staff and homeowners) with authentication and profile information.

**TypeScript Interface**:
```typescript
interface User {
  id: number;
  phone: string;
  nickname: string;
  avatar: string;
  userType: UserType;
  status: UserStatus;
  verificationStatus: VerificationStatus;
  scope: string;
  lastLoginAt: string;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
}

enum UserType {
  Staff = 1,
  Homeowner = 2
}

enum UserStatus {
  Active = 1,
  Disabled = 2,
  Locked = 3
}

enum VerificationStatus {
  Unverified = 0,
  Verified = 1,
  Rejected = 2
}
```

**Field Descriptions**:
- `id`: Unique user identifier (auto-generated)
- `phone`: Chinese mobile phone number (format: 1[3-9]\d{9}), unique, used for login
- `nickname`: Display name (2-20 characters)
- `avatar`: Avatar image URL (MinIO storage)
- `userType`: 1=Staff (admin), 2=Homeowner (resident)
- `status`: 1=Active, 2=Disabled, 3=Locked
- `verificationStatus`: 0=Unverified, 1=Verified, 2=Rejected (required for property binding)
- `scope`: Administrative scope (JSON string, e.g., `{"divisionIds": [1,2,3]}`)
- `lastLoginAt`: Last login timestamp (ISO 8601)
- `createdAt`: Creation timestamp
- `updatedAt`: Last update timestamp
- `deleteTime`: Soft delete timestamp (0 = not deleted)

**Validation Rules**:
- Phone: Required, unique, format `/^1[3-9]\d{9}$/`
- Nickname: Required, 2-20 characters
- UserType: Required, must be 1 or 2
- Status: Required, must be 1, 2, or 3

**State Transitions**:
- Active → Disabled (admin disables user)
- Disabled → Active (admin enables user)
- Active → Locked (system locks after failed login attempts)
- Unverified → Verified (admin approves verification)
- Unverified → Rejected (admin rejects verification)
- Rejected → Unverified (user resubmits verification)

---

### 2. Role

Represents permission groups with many-to-many relationships to users and permissions.

**TypeScript Interface**:
```typescript
interface Role {
  id: number;
  name: string;
  code: string;
  description: string;
  isSystem: boolean;
  status: RoleStatus;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
}

enum RoleStatus {
  Active = 1,
  Disabled = 2
}
```

**Field Descriptions**:
- `id`: Unique role identifier
- `name`: Display name (e.g., "Community Manager")
- `code`: Unique code (e.g., "community_manager")
- `description`: Role description
- `isSystem`: true=System role (cannot delete), false=Custom role
- `status`: 1=Active, 2=Disabled
- `createdAt`: Creation timestamp
- `updatedAt`: Last update timestamp
- `deleteTime`: Soft delete timestamp

**Validation Rules**:
- Name: Required, max 50 characters
- Code: Required, unique, alphanumeric + underscore
- IsSystem: System roles cannot be deleted

**Business Rules**:
- Super Administrator role (isSystem=true) is protected
- Cannot delete roles with assigned users
- Disabling a role immediately revokes permissions for assigned users

---

### 3. Permission

Represents access control rules in a hierarchical tree structure (menu and button permissions).

**TypeScript Interface**:
```typescript
interface Permission {
  id: number;
  parentId: number;
  name: string;
  code: string;
  type: PermissionType;
  path: string;
  icon: string;
  sortOrder: number;
  status: PermissionStatus;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
  children?: Permission[];
}

enum PermissionType {
  Menu = 1,
  Button = 2
}

enum PermissionStatus {
  Active = 1,
  Disabled = 2
}
```

**Field Descriptions**:
- `id`: Unique permission identifier
- `parentId`: Parent permission ID (0 for root)
- `name`: Display name (e.g., "User Management")
- `code`: Unique code (e.g., "identity:user:view")
- `type`: 1=Menu (navigation), 2=Button (action)
- `path`: Route path for menu permissions (e.g., "/users")
- `icon`: Icon name for menu permissions
- `sortOrder`: Display order
- `status`: 1=Active, 2=Disabled
- `children`: Child permissions (tree structure)

**Validation Rules**:
- Code: Required, unique, format `service:resource:action`
- Type: Required, must be 1 or 2
- ParentId: Must reference existing permission or be 0

**Hierarchy Rules**:
- Menu permissions can have menu or button children
- Button permissions are leaf nodes (no children)
- Root permissions have parentId=0

---

### 4. AdministrativeDivision

Represents five-tier geographic hierarchy (province → city → district → street → community).

**TypeScript Interface**:
```typescript
interface AdministrativeDivision {
  id: number;
  parentId: number;
  level: DivisionLevel;
  name: string;
  code: string;
  path: string;
  sortOrder: number;
  status: DivisionStatus;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
  children?: AdministrativeDivision[];
}

enum DivisionLevel {
  Province = 1,
  City = 2,
  District = 3,
  Street = 4,
  Community = 5
}

enum DivisionStatus {
  Active = 1,
  Inactive = 2
}
```

**Field Descriptions**:
- `id`: Unique division identifier
- `parentId`: Parent division ID (0 for provinces)
- `level`: 1=Province, 2=City, 3=District, 4=Street, 5=Community
- `name`: Division name (e.g., "Guangdong Province")
- `code`: Unique hierarchical code (e.g., "440000")
- `path`: Full path (e.g., "/1/2/3")
- `sortOrder`: Display order within same level
- `status`: 1=Active, 2=Inactive
- `children`: Child divisions (tree structure)

**Validation Rules**:
- Name: Required, max 100 characters
- Code: Required, unique, max 20 characters
- Level: Required, must be 1-5
- ParentId: Required for levels 2-5, must be 0 for level 1

**Hierarchy Rules**:
- Level 1 (Province): parentId=0
- Level 2 (City): parent must be level 1
- Level 3 (District): parent must be level 2
- Level 4 (Street): parent must be level 3
- Level 5 (Community): parent must be level 4
- Cannot delete divisions with children or associated communities

---

### 5. Community

Represents residential communities with submission workflow and review tracking.

**TypeScript Interface**:
```typescript
interface Community {
  id: number;
  divisionId: number;
  name: string;
  address: string;
  area: number;
  population: number;
  communityType: CommunityType;
  submissionStatus: SubmissionStatus;
  submitterId: number;
  submittedAt: string;
  reviewerId: number;
  reviewedAt: string;
  reviewNotes: string;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
  division?: AdministrativeDivision;
  submitter?: User;
  reviewer?: User;
}

enum CommunityType {
  Residential = 1,
  Village = 2,
  Mixed = 3
}

enum SubmissionStatus {
  Draft = 0,
  Submitted = 1,
  Approved = 2,
  Rejected = 3
}
```

**Field Descriptions**:
- `id`: Unique community identifier
- `divisionId`: Associated administrative division (level 5)
- `name`: Community name (max 100 characters)
- `address`: Full address (max 255 characters)
- `area`: Area in square meters
- `population`: Estimated population
- `communityType`: 1=Residential, 2=Village, 3=Mixed
- `submissionStatus`: 0=Draft, 1=Submitted, 2=Approved, 3=Rejected
- `submitterId`: User who submitted for review
- `submittedAt`: Submission timestamp
- `reviewerId`: Admin who reviewed
- `reviewedAt`: Review timestamp
- `reviewNotes`: Review feedback (required for rejection)

**Validation Rules**:
- Name: Required, max 100 characters
- Address: Required, max 255 characters
- Area: Required, positive number
- Population: Required, positive integer
- CommunityType: Required, must be 1, 2, or 3
- DivisionId: Required, must reference level 5 division

**State Transitions**:
- Draft → Submitted (provincial/municipal admin submits)
- Submitted → Approved (headquarters admin approves)
- Submitted → Rejected (headquarters admin rejects with notes)
- Rejected → Submitted (provincial/municipal admin resubmits)
- Cannot edit or delete Approved communities (headquarters only)

---

### 6. HomeownerVerification

Represents homeowner identity verification requests with document uploads.

**TypeScript Interface**:
```typescript
interface HomeownerVerification {
  id: number;
  userId: number;
  propertyUnit: string;
  realName: string;
  idCard: string;
  documentUrls: string[];
  verificationStatus: VerificationStatus;
  reviewerId: number;
  reviewedAt: string;
  reviewNotes: string;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
  user?: User;
  reviewer?: User;
}

enum VerificationStatus {
  Pending = 0,
  Approved = 1,
  Rejected = 2
}
```

**Field Descriptions**:
- `id`: Unique verification identifier
- `userId`: Homeowner user ID
- `propertyUnit`: Property unit identifier (e.g., "Building 1, Unit 101")
- `realName`: Homeowner's real name (max 50 characters)
- `idCard`: ID card number (18 characters, partially masked in UI)
- `documentUrls`: Array of document image URLs (1-9 images)
- `verificationStatus`: 0=Pending, 1=Approved, 2=Rejected
- `reviewerId`: Admin who reviewed
- `reviewedAt`: Review timestamp
- `reviewNotes`: Review feedback

**Validation Rules**:
- RealName: Required, max 50 characters
- IdCard: Required, format `/^\d{17}[\dXx]$/`
- DocumentUrls: Required, 1-9 URLs
- PropertyUnit: Required

**State Transitions**:
- Pending → Approved (admin approves, user.verificationStatus becomes 1)
- Pending → Rejected (admin rejects with notes)
- Rejected → Pending (user resubmits)

**Security**:
- ID card displayed as masked in list views (e.g., "110***********1234")
- Full ID card only visible in detail view to authorized admins

---

### 7. Configuration

Represents system-wide configuration parameters organized by module.

**TypeScript Interface**:
```typescript
interface Configuration {
  id: number;
  module: string;
  key: string;
  value: string;
  valueType: ConfigValueType;
  description: string;
  isPublic: boolean;
  approvalStatus: ApprovalStatus;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
}

enum ConfigValueType {
  String = 'string',
  Number = 'number',
  Boolean = 'boolean',
  Json = 'json'
}

enum ApprovalStatus {
  Draft = 0,
  PendingApproval = 1,
  Approved = 2
}
```

**Field Descriptions**:
- `id`: Unique configuration identifier
- `module`: Module name (e.g., "auth", "upload", "notification")
- `key`: Configuration key (e.g., "max_upload_size")
- `value`: Configuration value (stored as string, parsed by valueType)
- `valueType`: string, number, boolean, or json
- `description`: Configuration description
- `isPublic`: true=Visible to all users, false=Internal only
- `approvalStatus`: 0=Draft, 1=Pending, 2=Approved

**Validation Rules**:
- Module: Required
- Key: Required, unique within module
- Value: Required, must match valueType format
- ValueType: Required

**Value Type Parsing**:
- `string`: Use as-is
- `number`: Parse with `parseFloat(value)`
- `boolean`: Parse with `value === 'true'`
- `json`: Parse with `JSON.parse(value)`

---

### 8. SensitiveWord

Represents content filtering rules with severity and action configuration.

**TypeScript Interface**:
```typescript
interface SensitiveWord {
  id: number;
  word: string;
  category: string;
  severity: Severity;
  action: SensitiveWordAction;
  status: SensitiveWordStatus;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
}

enum Severity {
  Low = 1,
  Medium = 2,
  High = 3
}

enum SensitiveWordAction {
  Warn = 1,
  Block = 2,
  Review = 3
}

enum SensitiveWordStatus {
  Active = 1,
  Inactive = 2
}
```

**Field Descriptions**:
- `id`: Unique word identifier
- `word`: Sensitive word (max 100 characters, unique)
- `category`: Category (e.g., "profanity", "political", "spam")
- `severity`: 1=Low, 2=Medium, 3=High
- `action`: 1=Warn user, 2=Block submission, 3=Send to review
- `status`: 1=Active (filtering enabled), 2=Inactive (filtering disabled)

**Validation Rules**:
- Word: Required, unique, max 100 characters
- Severity: Required, must be 1, 2, or 3
- Action: Required, must be 1, 2, or 3

**Business Rules**:
- Active words are used for content filtering
- Inactive words are retained for historical reference
- Cannot delete words with usage history

---

## Supporting Types

### Authentication Types

```typescript
interface LoginRequest {
  phone: string;
  password: string;
}

interface LoginSmsRequest {
  phone: string;
  smsCode: string;
}

interface RegisterRequest {
  phone: string;
  password?: string;
  smsCode: string;
  nickname: string;
}

interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: User;
}

interface RefreshTokenRequest {
  refreshToken: string;
}

interface RefreshTokenResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}
```

### Pagination Types

```typescript
interface PageRequest {
  page: number;
  pageSize: number;
}

interface PageResponse<T> {
  list: T[];
  total: number;
}

interface PaginatedResponse<T> extends PageResponse<T> {
  page: number;
  pageSize: number;
}
```

### API Response Types

```typescript
interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

interface ApiError {
  code: number;
  message: string;
}
```

### Filter Types

```typescript
interface UserFilter extends PageRequest {
  userType?: UserType;
  status?: UserStatus;
  verificationStatus?: VerificationStatus;
  keyword?: string;
}

interface CommunityFilter extends PageRequest {
  divisionId?: number;
  submissionStatus?: SubmissionStatus;
  communityType?: CommunityType;
  keyword?: string;
}

interface VerificationFilter extends PageRequest {
  verificationStatus?: VerificationStatus;
  startDate?: string;
  endDate?: string;
}

interface ConfigFilter extends PageRequest {
  module?: string;
  keyword?: string;
}

interface SensitiveWordFilter extends PageRequest {
  category?: string;
  severity?: Severity;
  status?: SensitiveWordStatus;
}
```

## Relationships

### User ↔ Role (Many-to-Many)
- Users can have multiple roles
- Roles can be assigned to multiple users
- Join table: `user_roles` (not exposed to frontend)

### Role ↔ Permission (Many-to-Many)
- Roles can have multiple permissions
- Permissions can be assigned to multiple roles
- Join table: `role_permissions` (not exposed to frontend)

### AdministrativeDivision (Self-Referencing Tree)
- Each division has one parent (except provinces)
- Each division can have multiple children
- Tree depth: 5 levels (province → city → district → street → community)

### Community → AdministrativeDivision (Many-to-One)
- Each community belongs to one division (level 5)
- Each division can have multiple communities

### Community → User (Review Tracking)
- `submitterId` references User (who submitted)
- `reviewerId` references User (who reviewed)

### HomeownerVerification → User (Many-to-One)
- Each verification belongs to one user
- Each user can have multiple verification attempts
- `reviewerId` references User (who reviewed)

## Data Desensitization Rules

For security and privacy, certain fields must be desensitized in list views:

### Phone Numbers
- **List view**: `138****0000` (mask middle 4 digits)
- **Detail view**: `13800000000` (full number)
- **Implementation**: `phone.replace(/(\d{3})\d{4}(\d{4})/, '$1****$2')`

### ID Card Numbers
- **List view**: `110***********1234` (mask middle 11 digits)
- **Detail view**: `110101199001011234` (full number, admin only)
- **Implementation**: `idCard.replace(/(\d{3})\d{11}(\d{4})/, '$1***********$2')`

### Real Names
- **List view**: `张*` (show first character + asterisk)
- **Detail view**: `张三` (full name, admin only)
- **Implementation**: `name.charAt(0) + '*'.repeat(name.length - 1)`

## Summary

This data model defines 8 core entities and supporting types for the PC admin frontend. All entities are synchronized with backend APIs and include:

- **Type safety**: TypeScript interfaces with strict enums
- **Validation rules**: Field constraints and format requirements
- **State transitions**: Workflow states and allowed transitions
- **Relationships**: Entity associations and tree structures
- **Security**: Data desensitization rules for sensitive fields

The model supports all 8 user stories defined in the feature spec and aligns with Constitution Principle VI (Frontend Development Specifications).
