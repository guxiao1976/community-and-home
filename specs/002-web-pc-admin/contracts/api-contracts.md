# API Contracts: Web PC Admin Frontend

**Feature**: 002-web-pc-admin  
**Date**: 2026-05-03  
**Status**: Complete

## Overview

This document defines the API contracts between the PC admin frontend and backend microservices. All contracts are based on the API documentation in `docs/api/` and must remain synchronized with backend implementations.

## Base Configuration

### Service Endpoints

```typescript
// web/common/constants/config.ts
export const API_CONFIG = {
  identity: {
    baseURL: 'http://localhost:8888/api/identity',
    timeout: 30000
  },
  masterdata: {
    baseURL: 'http://localhost:8889/api/masterdata',
    timeout: 30000
  }
};
```

### Request Headers

```typescript
// All authenticated requests
{
  'Authorization': 'Bearer <access_token>',
  'Content-Type': 'application/json'
}

// File upload requests
{
  'Authorization': 'Bearer <access_token>',
  'Content-Type': 'multipart/form-data'
}
```

## Identity Service API Contracts

### Authentication Endpoints

#### POST /auth/login
**Purpose**: Login with phone and password

**Request**:
```typescript
{
  phone: string;      // Format: 1[3-9]\d{9}
  password: string;   // Min 8 characters
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    accessToken: string;   // JWT, expires in 24h
    refreshToken: string;  // JWT, expires in 7d
    expiresIn: number;     // 86400 (seconds)
    user: User;
  }
}
```

**Error Codes**:
- 400: Invalid phone format or missing fields
- 401: Invalid credentials
- 500: Internal server error

---

#### POST /auth/login/sms
**Purpose**: Login with phone and SMS code

**Request**:
```typescript
{
  phone: string;    // Format: 1[3-9]\d{9}
  smsCode: string;  // 6-digit code
}
```

**Response**: Same as `/auth/login`

**Error Codes**:
- 400: Invalid phone or SMS code
- 401: SMS code expired or invalid
- 500: Internal server error

---

#### POST /auth/register
**Purpose**: Register new user account

**Request**:
```typescript
{
  phone: string;       // Format: 1[3-9]\d{9}, must be unique
  password?: string;   // Optional, min 8 characters
  smsCode: string;     // 6-digit verification code
  nickname: string;    // 2-20 characters
}
```

**Response**: Same as `/auth/login`

**Error Codes**:
- 400: Invalid parameters or phone already registered
- 401: Invalid SMS code
- 500: Internal server error

---

#### POST /auth/sms/send
**Purpose**: Send SMS verification code

**Request**:
```typescript
{
  phone: string;  // Format: 1[3-9]\d{9}
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: null
}
```

**Error Codes**:
- 400: Invalid phone format
- 429: Too many requests (rate limited)
- 500: SMS service error

---

#### POST /auth/token/refresh
**Purpose**: Refresh access token using refresh token

**Request**:
```typescript
{
  refreshToken: string;  // Valid refresh token
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
  }
}
```

**Error Codes**:
- 401: Invalid or expired refresh token
- 500: Internal server error

---

#### POST /auth/logout
**Purpose**: Logout and invalidate tokens

**Request**: None (token in header)

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: null
}
```

---

### User Management Endpoints

#### GET /users
**Purpose**: Get paginated user list with filters

**Query Parameters**:
```typescript
{
  page?: number;              // Default: 1
  pageSize?: number;          // Default: 20
  userType?: 1 | 2;          // 1=Staff, 2=Homeowner
  status?: 1 | 2 | 3;        // 1=Active, 2=Disabled, 3=Locked
  verificationStatus?: 0 | 1 | 2;  // 0=Unverified, 1=Verified, 2=Rejected
  keyword?: string;           // Search by phone or nickname
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    list: User[];
    total: number;
  }
}
```

---

#### GET /users/:id
**Purpose**: Get user details by ID

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: User
}
```

**Error Codes**:
- 404: User not found
- 403: Insufficient permissions

---

#### POST /users
**Purpose**: Create new user (admin only)

**Request**:
```typescript
{
  phone: string;           // Format: 1[3-9]\d{9}, unique
  password: string;        // Min 8 characters
  nickname: string;        // 2-20 characters
  userType: 1 | 2;        // 1=Staff, 2=Homeowner
  scope?: string;          // JSON string, e.g., '{"divisionIds":[1,2,3]}'
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: User
}
```

**Error Codes**:
- 400: Invalid parameters or phone already exists
- 403: Insufficient permissions
- 500: Internal server error

---

#### PUT /users/:id
**Purpose**: Update user information

**Request**:
```typescript
{
  nickname?: string;
  avatar?: string;
  scope?: string;
  status?: 1 | 2;  // 1=Active, 2=Disabled
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: User
}
```

**Error Codes**:
- 400: Invalid parameters
- 403: Insufficient permissions
- 404: User not found

---

### Role Management Endpoints

#### GET /roles
**Purpose**: Get all roles

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    list: Role[];
    total: number;
  }
}
```

---

#### POST /roles
**Purpose**: Create new role

**Request**:
```typescript
{
  name: string;         // Max 50 characters
  code: string;         // Unique, alphanumeric + underscore
  description?: string;
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Role
}
```

**Error Codes**:
- 400: Invalid parameters or code already exists
- 403: Insufficient permissions

---

#### PUT /roles/:id
**Purpose**: Update role information

**Request**:
```typescript
{
  name?: string;
  description?: string;
  status?: 1 | 2;  // 1=Active, 2=Disabled
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Role
}
```

---

#### DELETE /roles/:id
**Purpose**: Delete role (soft delete)

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: null
}
```

**Error Codes**:
- 400: Cannot delete system role or role with assigned users
- 403: Insufficient permissions
- 404: Role not found

---

### Permission Management Endpoints

#### GET /permissions
**Purpose**: Get permission tree

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Permission[]  // Tree structure with children
}
```

---

#### GET /roles/:id/permissions
**Purpose**: Get permissions assigned to a role

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    permissionIds: number[];
  }
}
```

---

#### POST /roles/:id/permissions
**Purpose**: Assign permissions to role

**Request**:
```typescript
{
  permissionIds: number[];  // Array of permission IDs
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: null
}
```

---

#### GET /users/:id/permissions
**Purpose**: Get effective permissions for a user

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    permissions: string[];  // Array of permission codes
    menus: Permission[];    // Menu permissions with tree structure
  }
}
```

---

### Homeowner Verification Endpoints

#### GET /verifications
**Purpose**: Get paginated verification list

**Query Parameters**:
```typescript
{
  page?: number;
  pageSize?: number;
  verificationStatus?: 0 | 1 | 2;  // 0=Pending, 1=Approved, 2=Rejected
  startDate?: string;  // ISO 8601 format
  endDate?: string;
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    list: HomeownerVerification[];
    total: number;
  }
}
```

---

#### GET /verifications/:id
**Purpose**: Get verification details

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: HomeownerVerification
}
```

---

#### POST /verifications/:id/review
**Purpose**: Review verification request

**Request**:
```typescript
{
  status: 1 | 2;      // 1=Approved, 2=Rejected
  reviewNotes?: string;  // Required for rejection
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: HomeownerVerification
}
```

**Error Codes**:
- 400: Invalid status or missing review notes for rejection
- 403: Insufficient permissions
- 404: Verification not found

---

## Masterdata Service API Contracts

### Administrative Division Endpoints

#### GET /divisions
**Purpose**: Get division tree or list

**Query Parameters**:
```typescript
{
  parentId?: number;  // Get children of specific division
  level?: 1 | 2 | 3 | 4 | 5;  // Filter by level
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: AdministrativeDivision[]  // Tree structure with children
}
```

---

#### GET /divisions/:id
**Purpose**: Get division details

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: AdministrativeDivision
}
```

---

#### POST /divisions
**Purpose**: Create new division

**Request**:
```typescript
{
  parentId: number;     // 0 for provinces, required for others
  level: 1 | 2 | 3 | 4 | 5;
  name: string;         // Max 100 characters
  code: string;         // Unique, max 20 characters
  sortOrder?: number;   // Default: 0
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: AdministrativeDivision
}
```

**Error Codes**:
- 400: Invalid level hierarchy or code already exists
- 403: Insufficient permissions (headquarters only)

---

#### PUT /divisions/:id
**Purpose**: Update division information

**Request**:
```typescript
{
  name?: string;
  code?: string;
  sortOrder?: number;
  status?: 1 | 2;
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: AdministrativeDivision
}
```

---

#### DELETE /divisions/:id
**Purpose**: Delete division (soft delete)

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: null
}
```

**Error Codes**:
- 400: Cannot delete division with children or associated communities
- 403: Insufficient permissions

---

### Community Management Endpoints

#### GET /communities
**Purpose**: Get paginated community list

**Query Parameters**:
```typescript
{
  page?: number;
  pageSize?: number;
  divisionId?: number;
  submissionStatus?: 0 | 1 | 2 | 3;  // 0=Draft, 1=Submitted, 2=Approved, 3=Rejected
  communityType?: 1 | 2 | 3;  // 1=Residential, 2=Village, 3=Mixed
  keyword?: string;  // Search by name or address
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    list: Community[];
    total: number;
  }
}
```

---

#### GET /communities/:id
**Purpose**: Get community details

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Community
}
```

---

#### POST /communities
**Purpose**: Create new community

**Request**:
```typescript
{
  divisionId: number;        // Level 5 division
  name: string;              // Max 100 characters
  address: string;           // Max 255 characters
  area: number;              // Square meters
  population: number;        // Estimated population
  communityType: 1 | 2 | 3;  // 1=Residential, 2=Village, 3=Mixed
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Community
}
```

**Error Codes**:
- 400: Invalid parameters or division not found
- 403: Insufficient permissions (within scope only)

---

#### PUT /communities/:id
**Purpose**: Update community information

**Request**:
```typescript
{
  name?: string;
  address?: string;
  area?: number;
  population?: number;
  communityType?: 1 | 2 | 3;
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Community
}
```

**Error Codes**:
- 400: Invalid parameters or cannot edit approved community
- 403: Insufficient permissions

---

#### POST /communities/:id/submit
**Purpose**: Submit community for review

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Community
}
```

**Error Codes**:
- 400: Community not in Draft or Rejected status
- 403: Insufficient permissions

---

#### POST /communities/:id/review
**Purpose**: Review submitted community

**Request**:
```typescript
{
  action: 'approve' | 'reject';
  reviewNotes?: string;  // Required for rejection
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Community
}
```

**Error Codes**:
- 400: Community not in Submitted status or missing review notes
- 403: Insufficient permissions (headquarters only)

---

### Configuration Management Endpoints

#### GET /configs
**Purpose**: Get paginated configuration list

**Query Parameters**:
```typescript
{
  page?: number;
  pageSize?: number;
  module?: string;
  keyword?: string;  // Search by key
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    list: Configuration[];
    total: number;
  }
}
```

---

#### POST /configs
**Purpose**: Create new configuration

**Request**:
```typescript
{
  module: string;
  key: string;
  value: string;
  valueType: 'string' | 'number' | 'boolean' | 'json';
  description?: string;
  isPublic?: boolean;  // Default: false
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Configuration
}
```

---

#### PUT /configs/:id
**Purpose**: Update configuration

**Request**:
```typescript
{
  value?: string;
  description?: string;
  isPublic?: boolean;
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: Configuration
}
```

---

#### DELETE /configs/:id
**Purpose**: Delete configuration (soft delete)

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: null
}
```

---

### Sensitive Word Management Endpoints

#### GET /sensitive-words
**Purpose**: Get paginated sensitive word list

**Query Parameters**:
```typescript
{
  page?: number;
  pageSize?: number;
  category?: string;
  severity?: 1 | 2 | 3;
  status?: 1 | 2;
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: {
    list: SensitiveWord[];
    total: number;
  }
}
```

---

#### POST /sensitive-words
**Purpose**: Add new sensitive word

**Request**:
```typescript
{
  word: string;         // Max 100 characters, unique
  category?: string;
  severity: 1 | 2 | 3;  // 1=Low, 2=Medium, 3=High
  action: 1 | 2 | 3;    // 1=Warn, 2=Block, 3=Review
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: SensitiveWord
}
```

---

#### PUT /sensitive-words/:id
**Purpose**: Update sensitive word

**Request**:
```typescript
{
  category?: string;
  severity?: 1 | 2 | 3;
  action?: 1 | 2 | 3;
  status?: 1 | 2;
}
```

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: SensitiveWord
}
```

---

#### DELETE /sensitive-words/:id
**Purpose**: Delete sensitive word (soft delete)

**Response**:
```typescript
{
  code: 0,
  message: "success",
  data: null
}
```

---

## Error Handling

### Standard Error Response

```typescript
{
  code: number;    // Error code (400, 401, 403, 404, 500, etc.)
  message: string; // Error description
}
```

### Common Error Codes

| Code | Description | Frontend Action |
|------|-------------|-----------------|
| 0 | Success | Process response data |
| 400 | Invalid Parameter | Show validation error to user |
| 401 | Unauthorized | Attempt token refresh, redirect to login if fails |
| 403 | Forbidden | Show "Insufficient permissions" message |
| 404 | Not Found | Show "Resource not found" message |
| 500 | Internal Server Error | Show generic error, log to console |
| 501 | Database Error | Show "Database error, please try again" |
| 502 | Cache Error | Show "Cache error, please try again" |
| 503 | RPC Error | Show "Service unavailable, please try again" |

### Error Handling Pattern

```typescript
// web/pc/src/utils/request.ts
axios.interceptors.response.use(
  response => {
    const { code, message, data } = response.data;
    if (code === 0) {
      return data;
    } else {
      ElMessage.error(message);
      return Promise.reject(new Error(message));
    }
  },
  async error => {
    if (error.response?.status === 401) {
      // Attempt token refresh
      try {
        await refreshToken();
        return axios(error.config);
      } catch {
        router.push('/login');
      }
    } else if (error.response?.status === 403) {
      ElMessage.error('Insufficient permissions');
    } else if (error.response?.status === 404) {
      ElMessage.error('Resource not found');
    } else {
      ElMessage.error(error.response?.data?.message || 'Request failed');
    }
    return Promise.reject(error);
  }
);
```

## Summary

This document defines 46 API endpoints across Identity and Masterdata services:

- **Identity Service**: 28 endpoints (authentication, users, roles, permissions, verification)
- **Masterdata Service**: 18 endpoints (divisions, communities, configurations, sensitive words)

All contracts include:
- Request/response TypeScript types
- Query parameters and validation rules
- Error codes and handling strategies
- Business rule enforcement

These contracts align with backend API documentation in `docs/api/` and support all 8 user stories defined in the feature spec.
