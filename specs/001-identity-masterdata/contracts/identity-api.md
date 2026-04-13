# Identity Service API Contract

**Service**: identity-api  
**Protocol**: HTTP/REST  
**Base URL**: `/api/v1`  
**Authentication**: JWT Bearer Token (except login/register endpoints)  
**Response Format**: JSON

## Common Response Structure

All API responses follow this format:

```json
{
  "code": 0,           // 0=success, >0=error code
  "message": "success", // Human-readable message
  "data": {}           // Response payload (null on error)
}
```

## Common Error Codes

| Code | Message | Description |
|------|---------|-------------|
| 0 | success | Request successful |
| 1001 | invalid_params | Request parameter validation failed |
| 1002 | unauthorized | Not authenticated or token invalid |
| 1003 | forbidden | Insufficient permissions |
| 1004 | not_found | Resource not found |
| 1005 | duplicate | Resource already exists |
| 2001 | internal_error | Internal server error |
| 2002 | database_error | Database operation failed |
| 2003 | external_service_error | External service call failed |

---

## Authentication Endpoints

### POST /auth/register

Register a new user account.

**Request Body**:
```json
{
  "phone": "13800138000",
  "password": "Password123!",
  "verification_code": "123456",
  "user_type": 1,  // 1=Backend, 2=Homeowner
  "nickname": "张三"
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1001,
    "phone": "13800138000",
    "nickname": "张三",
    "user_type": 1
  }
}
```

**Validation**:
- phone: 11-digit Chinese mobile number
- password: Min 8 chars, must contain uppercase, lowercase, and number
- verification_code: 6-digit code from SMS
- user_type: 1 or 2

---

### POST /auth/login

User login with phone and password.

**Request Body**:
```json
{
  "phone": "13800138000",
  "password": "Password123!"
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 7200,  // seconds
    "user": {
      "user_id": 1001,
      "phone": "13800138000",
      "nickname": "张三",
      "user_type": 1,
      "verification_status": 1,
      "roles": ["ADMIN"],
      "permissions": ["user:create", "user:update"]
    }
  }
}
```

---

### POST /auth/login-sms

User login with phone and SMS verification code.

**Request Body**:
```json
{
  "phone": "13800138000",
  "verification_code": "123456"
}
```

**Response**: Same as `/auth/login`

---

### POST /auth/refresh-token

Refresh access token using refresh token.

**Request Body**:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 7200
  }
}
```

---

### POST /auth/logout

User logout (invalidate tokens).

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

### POST /auth/send-sms-code

Send SMS verification code.

**Request Body**:
```json
{
  "phone": "13800138000",
  "scene": "register"  // register/login/reset_password
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "expires_in": 300  // seconds
  }
}
```

**Rate Limit**: 5 codes per phone per hour

---

## User Management Endpoints

### GET /users

List users with pagination and filters.

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 20, max: 100)
- `user_type`: Filter by user type (1=Backend, 2=Homeowner)
- `status`: Filter by status (1=Active, 2=Disabled, 3=Locked)
- `verification_status`: Filter by verification status (0/1/2)
- `keyword`: Search by phone or nickname
- `scope_id`: Filter by administrative scope (headquarters only)

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 150,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "user_id": 1001,
        "phone": "13800138000",
        "nickname": "张三",
        "user_type": 1,
        "status": 1,
        "verification_status": 1,
        "scope_id": 23,
        "scope_name": "北京市",
        "last_login_time": "2026-04-13T10:30:00+08:00",
        "created_time": "2026-01-01T08:00:00+08:00"
      }
    ]
  }
}
```

**Scope Enforcement**: Provincial/municipal admins only see users within their scope

---

### GET /users/:id

Get user details by ID.

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1001,
    "phone": "13800138000",
    "nickname": "张三",
    "avatar_url": "https://minio.example.com/avatars/1001.jpg",
    "user_type": 1,
    "status": 1,
    "verification_status": 1,
    "scope_id": 23,
    "scope_name": "北京市",
    "roles": [
      {
        "role_id": 1,
        "role_name": "省级管理员",
        "role_code": "PROVINCE_ADMIN"
      }
    ],
    "last_login_time": "2026-04-13T10:30:00+08:00",
    "last_login_ip": "192.168.1.100",
    "created_time": "2026-01-01T08:00:00+08:00",
    "updated_time": "2026-04-13T10:30:00+08:00"
  }
}
```

---

### POST /users

Create a new backend user (headquarters/provincial admin only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "phone": "13800138001",
  "password": "Password123!",
  "nickname": "李四",
  "user_type": 1,
  "scope_id": 23,
  "role_ids": [2, 3]
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "user_id": 1002,
    "phone": "13800138001",
    "nickname": "李四"
  }
}
```

---

### PUT /users/:id

Update user information.

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "nickname": "李四（更新）",
  "status": 1,
  "scope_id": 24,
  "role_ids": [2, 3, 4]
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**Permissions**: 
- Users can update their own nickname/avatar
- Admins can update status/scope/roles within their scope

---

### DELETE /users/:id

Soft delete a user (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

---

## Role Management Endpoints

### GET /roles

List all roles.

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `page`: Page number
- `page_size`: Items per page
- `status`: Filter by status (1=Active, 2=Disabled)

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 10,
    "items": [
      {
        "role_id": 1,
        "role_name": "超级管理员",
        "role_code": "SUPER_ADMIN",
        "description": "系统超级管理员",
        "is_system": 1,
        "status": 1,
        "permission_count": 50,
        "created_time": "2026-01-01T08:00:00+08:00"
      }
    ]
  }
}
```

---

### GET /roles/:id

Get role details with permissions.

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "role_id": 2,
    "role_name": "省级管理员",
    "role_code": "PROVINCE_ADMIN",
    "description": "省级管理员角色",
    "is_system": 0,
    "status": 1,
    "permissions": [
      {
        "permission_id": 10,
        "permission_name": "用户管理",
        "permission_code": "user:view",
        "type": 1
      }
    ],
    "created_time": "2026-01-01T08:00:00+08:00"
  }
}
```

---

### POST /roles

Create a new role (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "role_name": "市级管理员",
  "role_code": "CITY_ADMIN",
  "description": "市级管理员角色",
  "permission_ids": [10, 11, 12]
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "role_id": 5,
    "role_name": "市级管理员",
    "role_code": "CITY_ADMIN"
  }
}
```

---

### PUT /roles/:id

Update role information and permissions.

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "role_name": "市级管理员（更新）",
  "description": "更新后的描述",
  "permission_ids": [10, 11, 12, 13],
  "status": 1
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**Validation**: Cannot update system roles (is_system=1)

---

### DELETE /roles/:id

Delete a role (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**Validation**: Cannot delete system roles or roles with assigned users

---

## Permission Management Endpoints

### GET /permissions

Get permission tree (hierarchical structure).

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "permission_id": 1,
      "permission_name": "用户管理",
      "permission_code": "user",
      "type": 1,
      "path": "/users",
      "icon": "user",
      "children": [
        {
          "permission_id": 10,
          "permission_name": "查看用户",
          "permission_code": "user:view",
          "type": 2
        },
        {
          "permission_id": 11,
          "permission_name": "创建用户",
          "permission_code": "user:create",
          "type": 2
        }
      ]
    }
  ]
}
```

---

## Homeowner Verification Endpoints

### POST /homeowner-verifications

Submit homeowner verification request.

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "property_unit_id": 5001,
  "document_urls": [
    "https://minio.example.com/verification/doc1.jpg",
    "https://minio.example.com/verification/doc2.jpg"
  ],
  "real_name": "王五",
  "id_card_number": "110101199001011234"
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "verification_id": 3001,
    "verification_status": 0,
    "submit_time": "2026-04-13T14:00:00+08:00"
  }
}
```

**Validation**:
- document_urls: Max 9 URLs, must be from MinIO
- id_card_number: Valid 18-digit Chinese ID card format
- Only one pending verification per user

---

### GET /homeowner-verifications

List verification requests (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `page`: Page number
- `page_size`: Items per page
- `verification_status`: Filter by status (0/1/2)
- `submit_time_start`: Filter by submit time range
- `submit_time_end`: Filter by submit time range

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "items": [
      {
        "verification_id": 3001,
        "user_id": 2001,
        "user_phone": "13800138002",
        "property_unit_id": 5001,
        "property_address": "北京市朝阳区XX小区1号楼101",
        "real_name": "王五",
        "id_card_number": "110101199001011234",
        "document_urls": ["url1", "url2"],
        "verification_status": 0,
        "submit_time": "2026-04-13T14:00:00+08:00"
      }
    ]
  }
}
```

---

### PUT /homeowner-verifications/:id/review

Review homeowner verification (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "verification_status": 1,  // 1=Approved, 2=Rejected
  "review_notes": "审核通过"
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**Side Effects**:
- If approved: user.verification_status = 1
- If rejected: user can resubmit

---

## Property Management Endpoints

### GET /property-units

List property units.

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `page`: Page number
- `page_size`: Items per page
- `community_id`: Filter by community
- `building`: Filter by building
- `keyword`: Search by unit number

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 200,
    "items": [
      {
        "property_unit_id": 5001,
        "community_id": 1001,
        "community_name": "XX小区",
        "building": "1号楼",
        "unit": "101",
        "floor": "1",
        "area": 120.50,
        "property_type": 1,
        "bound_user_count": 2,
        "primary_user": {
          "user_id": 2001,
          "nickname": "王五"
        }
      }
    ]
  }
}
```

---

### POST /property-bindings

Bind user to property unit.

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "property_unit_id": 5001,
  "is_primary": 0
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "binding_id": 6001,
    "bind_status": 2  // 2=Pending (requires primary user approval)
  }
}
```

**Validation**:
- User must be verified homeowner
- If property has primary user, binding status = Pending
- If no primary user, binding status = Active and is_primary = 1

---

### DELETE /property-bindings/:id

Remove user from property unit (primary user only).

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**Permissions**: Only primary user can remove other users

---

## Family Management Endpoints

### POST /families

Create family profile.

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "property_unit_id": 5001,
  "family_name": "王家"
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "family_id": 7001,
    "family_head_id": 2001
  }
}
```

**Validation**: User must be verified and bound to property unit

---

### POST /families/:id/members

Add family member.

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "name": "王小明",
  "relationship": "儿子",
  "phone": "13800138003",
  "id_card_number": "110101201001011234",
  "birth_date": "2010-01-01",
  "gender": 1
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "member_id": 8001
  }
}
```

---

### GET /families/:id/members

List family members.

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "family_id": 7001,
    "family_name": "王家",
    "family_head": {
      "user_id": 2001,
      "nickname": "王五"
    },
    "members": [
      {
        "member_id": 8001,
        "name": "王小明",
        "relationship": "儿子",
        "phone": "13800138003",
        "birth_date": "2010-01-01",
        "gender": 1
      }
    ]
  }
}
```

---

## File Upload Endpoints

### POST /files/upload-url

Generate presigned upload URL for MinIO.

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "file_name": "property_cert.jpg",
  "file_size": 2048576,
  "file_type": "image/jpeg",
  "entity_type": "homeowner_verification"
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "upload_url": "https://minio.example.com/bucket/path?signature=...",
    "file_path": "verification/2026/04/13/uuid.jpg",
    "expires_in": 900  // seconds
  }
}
```

**Validation**:
- file_size: Max 5MB (5242880 bytes)
- file_type: Only image/jpeg, image/png
- Max 9 files per upload batch

---

## Pagination

All list endpoints support pagination with these parameters:
- `page`: Page number (1-indexed, default: 1)
- `page_size`: Items per page (default: 20, max: 100)

Response includes:
- `total`: Total item count
- `page`: Current page
- `page_size`: Items per page
- `items`: Array of items

---

## Timestamp Format

All timestamps use RFC3339 format with timezone:
```
2026-04-13T15:04:05+08:00
```

---

## Rate Limiting

- SMS verification codes: 5 per phone per hour
- Login attempts: 10 per IP per hour
- API requests: 1000 per user per minute (general)
- File uploads: 20 per user per hour
