# Masterdata Service API Contract

**Service**: masterdata-api  
**Protocol**: HTTP/REST  
**Base URL**: `/api/v1`  
**Authentication**: JWT Bearer Token  
**Response Format**: JSON

## Common Response Structure

All API responses follow this format:

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

---

## Administrative Division Endpoints

### GET /administrative-divisions

Get administrative division tree or list.

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `parent_id`: Get children of specific division (omit for root level)
- `level`: Filter by level (1-5)
- `keyword`: Search by name or code
- `flat`: Return flat list instead of tree (default: false)

**Response (Tree)**:
```json
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "division_id": 1,
      "parent_id": null,
      "level": 1,
      "name": "北京市",
      "code": "110000",
      "path": "/1/",
      "status": 1,
      "children": [
        {
          "division_id": 2,
          "parent_id": 1,
          "level": 2,
          "name": "市辖区",
          "code": "110100",
          "path": "/1/2/",
          "status": 1,
          "children": []
        }
      ]
    }
  ]
}
```

**Response (Flat)**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 500,
    "items": [
      {
        "division_id": 1,
        "parent_id": null,
        "level": 1,
        "name": "北京市",
        "code": "110000",
        "path": "/1/",
        "status": 1
      }
    ]
  }
}
```

---

### GET /administrative-divisions/:id

Get division details.

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "division_id": 1,
    "parent_id": null,
    "level": 1,
    "name": "北京市",
    "code": "110000",
    "path": "/1/",
    "sort_order": 0,
    "status": 1,
    "created_by": 1,
    "created_time": "2026-01-01T08:00:00+08:00",
    "updated_time": "2026-01-01T08:00:00+08:00"
  }
}
```

---

### POST /administrative-divisions

Create administrative division (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "parent_id": 1,
  "level": 2,
  "name": "朝阳区",
  "code": "110105",
  "sort_order": 5
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "division_id": 23,
    "path": "/1/23/"
  }
}
```

**Validation**:
- parent_id: Must exist and be level-1 for level 2
- code: Must be unique
- level: Must be 1-5

---

### PUT /administrative-divisions/:id

Update administrative division (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "name": "朝阳区（更新）",
  "sort_order": 6,
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

**Validation**: Cannot change parent_id or level after creation

---

### DELETE /administrative-divisions/:id

Soft delete administrative division (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": null
}
```

**Validation**: Cannot delete if has active children or associated communities

---

## Community Management Endpoints

### GET /communities

List communities with filters.

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `page`: Page number
- `page_size`: Items per page
- `division_id`: Filter by administrative division
- `submission_status`: Filter by status (0/1/2/3)
- `community_type`: Filter by type (1/2/3)
- `keyword`: Search by name or address

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 100,
    "page": 1,
    "page_size": 20,
    "items": [
      {
        "community_id": 1001,
        "division_id": 23,
        "division_name": "朝阳区",
        "name": "XX小区",
        "address": "北京市朝阳区XX路XX号",
        "area": 0.5,
        "population": 5000,
        "community_type": 1,
        "submission_status": 2,
        "submitter_id": 100,
        "submitter_name": "张三",
        "submit_time": "2026-03-01T10:00:00+08:00",
        "reviewer_id": 1,
        "review_time": "2026-03-02T14:00:00+08:00",
        "created_time": "2026-03-01T09:00:00+08:00"
      }
    ]
  }
}
```

**Scope Enforcement**: Provincial/municipal admins only see communities within their scope

---

### GET /communities/:id

Get community details.

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "community_id": 1001,
    "division_id": 23,
    "division_path": "北京市/市辖区/朝阳区",
    "name": "XX小区",
    "address": "北京市朝阳区XX路XX号",
    "area": 0.5,
    "population": 5000,
    "community_type": 1,
    "submission_status": 2,
    "submitter": {
      "user_id": 100,
      "nickname": "张三"
    },
    "submit_time": "2026-03-01T10:00:00+08:00",
    "reviewer": {
      "user_id": 1,
      "nickname": "超级管理员"
    },
    "review_time": "2026-03-02T14:00:00+08:00",
    "review_notes": "审核通过",
    "created_time": "2026-03-01T09:00:00+08:00",
    "updated_time": "2026-03-02T14:00:00+08:00"
  }
}
```

---

### POST /communities

Create community (provincial/municipal admin).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "division_id": 23,
  "name": "YY小区",
  "address": "北京市朝阳区YY路YY号",
  "area": 0.3,
  "population": 3000,
  "community_type": 1
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "community_id": 1002,
    "submission_status": 0
  }
}
```

**Validation**:
- division_id: Must be level 5 (community) and within user's scope
- Initial status: 0 (Draft)

---

### PUT /communities/:id

Update community (creator only, before submission).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "name": "YY小区（更新）",
  "address": "北京市朝阳区YY路YY号",
  "area": 0.35,
  "population": 3200
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

**Validation**: Can only update if submission_status = 0 (Draft)

---

### POST /communities/:id/submit

Submit community for approval.

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "submission_status": 1,
    "submit_time": "2026-04-13T15:00:00+08:00"
  }
}
```

**Validation**: Can only submit if status = 0 (Draft) or 3 (Rejected)

---

### POST /communities/:id/review

Review community submission (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "submission_status": 2,  // 2=Approved, 3=Rejected
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

**Validation**: Can only review if status = 1 (Submitted)

---

### DELETE /communities/:id

Soft delete community (headquarters only).

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

## District Economic Data Endpoints

### GET /district-economic-data

List economic data with filters.

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `division_id`: Filter by district
- `year`: Filter by year
- `year_start`: Filter by year range
- `year_end`: Filter by year range

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 50,
    "items": [
      {
        "data_id": 5001,
        "division_id": 23,
        "division_name": "朝阳区",
        "year": 2025,
        "population": 3500000,
        "gdp": 750000.00,
        "per_capita_income": 85000.00,
        "unemployment_rate": 3.2,
        "data_source": "统计局",
        "created_time": "2026-02-01T10:00:00+08:00"
      }
    ]
  }
}
```

---

### POST /district-economic-data

Create economic data record (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "division_id": 23,
  "year": 2025,
  "population": 3500000,
  "gdp": 750000.00,
  "per_capita_income": 85000.00,
  "unemployment_rate": 3.2,
  "data_source": "统计局"
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "data_id": 5001
  }
}
```

**Validation**:
- division_id: Must be level 3 (district)
- year: Between 2000 and current year + 1
- Unique constraint on (division_id, year)

---

### PUT /district-economic-data/:id

Update economic data (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "population": 3550000,
  "gdp": 760000.00,
  "per_capita_income": 86000.00
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

---

## Configuration Management Endpoints

### GET /configurations

List configurations by module.

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `module`: Filter by module (required)
- `is_public`: Filter by public flag (0/1)
- `approval_status`: Filter by approval status (0/1/2)

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 20,
    "items": [
      {
        "config_id": 3001,
        "module": "auth",
        "config_key": "sms.rate_limit",
        "config_value": "5",
        "value_type": "number",
        "description": "SMS验证码每小时限制次数",
        "is_public": 0,
        "approval_status": 2,
        "created_time": "2026-01-01T08:00:00+08:00"
      }
    ]
  }
}
```

---

### GET /configurations/:id

Get configuration details.

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "config_id": 3001,
    "module": "auth",
    "config_key": "sms.rate_limit",
    "config_value": "5",
    "value_type": "number",
    "description": "SMS验证码每小时限制次数",
    "is_public": 0,
    "approval_status": 2,
    "created_by": 1,
    "created_time": "2026-01-01T08:00:00+08:00",
    "updated_time": "2026-01-01T08:00:00+08:00"
  }
}
```

---

### POST /configurations

Create configuration (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "module": "auth",
  "config_key": "jwt.access_token_ttl",
  "config_value": "7200",
  "value_type": "number",
  "description": "JWT访问令牌有效期（秒）",
  "is_public": 0
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "config_id": 3002,
    "approval_status": 0
  }
}
```

**Validation**:
- config_key: Alphanumeric with dots
- value_type: Must be string/number/boolean/json
- Initial approval_status: 0 (Draft)

---

### PUT /configurations/:id

Update configuration (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "config_value": "7200",
  "description": "JWT访问令牌有效期（秒）- 更新"
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

**Side Effect**: approval_status reset to 0 (Draft) if changed

---

### POST /configurations/:id/approve

Approve configuration change (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "approval_status": 2
  }
}
```

**Validation**: Can only approve if status = 1 (Pending)

---

## Sensitive Word Management Endpoints

### GET /sensitive-words

List sensitive words.

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `page`: Page number
- `page_size`: Items per page
- `category`: Filter by category
- `severity`: Filter by severity (1/2/3)
- `status`: Filter by status (1/2)

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 500,
    "items": [
      {
        "word_id": 4001,
        "word": "敏感词",
        "category": "political",
        "severity": 3,
        "action": 2,
        "status": 1,
        "created_time": "2026-01-01T08:00:00+08:00"
      }
    ]
  }
}
```

---

### POST /sensitive-words

Add sensitive word (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "word": "新敏感词",
  "category": "violence",
  "severity": 2,
  "action": 2
}
```

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "word_id": 4002
  }
}
```

**Validation**:
- word: Trimmed, lowercase
- severity: 1-3
- action: 1=Warn, 2=Block, 3=Review

---

### PUT /sensitive-words/:id

Update sensitive word (headquarters only).

**Headers**: `Authorization: Bearer {access_token}`

**Request Body**:
```json
{
  "category": "political",
  "severity": 3,
  "action": 2,
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

---

### DELETE /sensitive-words/:id

Delete sensitive word (headquarters only).

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

## Audit Log Endpoints

### GET /audit-logs

Query audit logs.

**Headers**: `Authorization: Bearer {access_token}`

**Query Parameters**:
- `page`: Page number
- `page_size`: Items per page
- `user_id`: Filter by user
- `entity_type`: Filter by entity type
- `entity_id`: Filter by entity ID
- `action`: Filter by action (CREATE/UPDATE/DELETE)
- `start_time`: Filter by time range
- `end_time`: Filter by time range

**Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total": 1000,
    "items": [
      {
        "log_id": 10001,
        "user_id": 100,
        "user_name": "张三",
        "entity_type": "md_community",
        "entity_id": 1001,
        "action": "UPDATE",
        "old_value": "{\"name\":\"XX小区\"}",
        "new_value": "{\"name\":\"XX小区（更新）\"}",
        "ip_address": "192.168.1.100",
        "created_time": "2026-04-13T15:30:00+08:00"
      }
    ]
  }
}
```

**Permissions**: Headquarters can view all logs, others can view logs within their scope

---

## Pagination & Timestamps

Same as Identity API:
- Pagination: `page`, `page_size` parameters
- Timestamps: RFC3339 format with timezone

---

## Rate Limiting

- Configuration changes: 100 per hour per user
- Sensitive word operations: 200 per hour per user
- General API requests: 1000 per minute per user
