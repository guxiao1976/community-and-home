# Quick Start Guide

This guide will help you get started with the Community & Home Management System API.

## Prerequisites

- Services running on localhost (Identity: 8888, Masterdata: 8889)
- curl or any HTTP client
- Default admin account credentials

## Step 1: Login

Get an access token by logging in with the default admin account.

```bash
curl -X POST http://localhost:8888/api/identity/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800000000",
    "password": "Admin@123456"
  }'
```

**Response:**

```json
{
  "user_id": 1,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expire": 1704067200
}
```

Save the `token` value - you'll need it for authenticated requests.

## Step 2: Make Authenticated Requests

Use the token in the Authorization header for protected endpoints.

### Example: Get User List

```bash
curl -X GET "http://localhost:8888/api/identity/users?page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### Example: Get Community List

```bash
curl -X GET "http://localhost:8889/api/masterdata/communities?page=1&page_size=10" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

**Note**: Masterdata endpoints currently don't require authentication, but include the token for future compatibility.

## Step 3: Refresh Token

When your access token expires (after 24 hours), use the refresh token to get a new one.

```bash
curl -X POST http://localhost:8888/api/identity/auth/token/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN_HERE"
  }'
```

## Common Workflows

### User Registration Flow

1. **Send SMS Code**

```bash
curl -X POST http://localhost:8888/api/identity/auth/sms/send \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13900000001"
  }'
```

2. **Register with SMS Code**

```bash
curl -X POST http://localhost:8888/api/identity/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13900000001",
    "code": "123456",
    "nickname": "New User",
    "password": "Password123!"
  }'
```

### Homeowner Verification Flow

1. **Submit Verification** (requires authentication)

```bash
curl -X POST http://localhost:8888/api/identity/verification \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "property_unit_id": 1,
    "real_name": "张三",
    "id_card_number": "110101199001011234",
    "document_urls": [
      "https://minio.example.com/docs/id-card-front.jpg",
      "https://minio.example.com/docs/id-card-back.jpg"
    ]
  }'
```

2. **Check Verification Status**

```bash
curl -X GET http://localhost:8888/api/identity/verification/my \
  -H "Authorization: Bearer YOUR_TOKEN"
```

3. **Admin Reviews Verification**

```bash
curl -X PUT http://localhost:8888/api/identity/verification/1/review \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": 1,
    "review_notes": "Approved"
  }'
```

### Community Management Flow

1. **Create Community**

```bash
curl -X POST http://localhost:8889/api/masterdata/communities \
  -H "Content-Type: application/json" \
  -d '{
    "division_id": 1,
    "name": "阳光花园",
    "address": "北京市朝阳区某某街道123号",
    "area": 50000.5,
    "population": 5000,
    "community_type": 1
  }'
```

2. **Submit for Review**

```bash
curl -X POST http://localhost:8889/api/masterdata/communities/1/submit \
  -H "Content-Type: application/json" \
  -d '{}'
```

3. **Admin Reviews Community**

```bash
curl -X POST http://localhost:8889/api/masterdata/communities/1/review \
  -H "Content-Type: application/json" \
  -d '{
    "action": "approve",
    "review_notes": "Community information verified"
  }'
```

## Testing with Different User Types

### Create Backend Staff User

```bash
curl -X POST http://localhost:8888/api/identity/users \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13900000002",
    "password": "Staff123!",
    "nickname": "Staff Member",
    "user_type": 1
  }'
```

### Create Homeowner User

```bash
curl -X POST http://localhost:8888/api/identity/users \
  -H "Authorization: Bearer ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13900000003",
    "password": "Owner123!",
    "nickname": "Homeowner",
    "user_type": 2
  }'
```

## Pagination Example

Get paginated results:

```bash
curl -X GET "http://localhost:8889/api/masterdata/communities?page=2&page_size=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

Response includes total count:

```json
{
  "list": [...],
  "total": 150
}
```

## Filtering Examples

### Filter Users by Type

```bash
curl -X GET "http://localhost:8888/api/identity/users?user_type=2&status=1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Filter Communities by Division

```bash
curl -X GET "http://localhost:8889/api/masterdata/communities?division_id=1&submission_status=2" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Get Division Tree

```bash
curl -X GET "http://localhost:8889/api/masterdata/divisions?mode=tree" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## Error Handling

### Invalid Credentials

```bash
curl -X POST http://localhost:8888/api/identity/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800000000",
    "password": "WrongPassword"
  }'
```

Response:

```json
{
  "code": 401,
  "message": "登录失败"
}
```

### Expired Token

When your token expires, you'll get:

```json
{
  "code": 401,
  "message": "unauthorized"
}
```

Use the refresh token endpoint to get a new access token.

### Missing Required Fields

```bash
curl -X POST http://localhost:8889/api/masterdata/communities \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Community"
  }'
```

Response:

```json
{
  "code": 400,
  "message": "invalid parameter: division_id is required"
}
```

## Troubleshooting

### Service Not Responding

Check if services are running:

```bash
# Check Identity API
curl http://localhost:8888/api/identity/auth/login

# Check Masterdata API
curl http://localhost:8889/api/masterdata/communities
```

### Token Issues

1. Verify token format: `Bearer <token>`
2. Check token expiration (24 hours for access token)
3. Use refresh token to get new access token
4. Ensure no extra spaces in Authorization header

### Database Connection Issues

Verify infrastructure is running:

```bash
docker-compose ps
```

All services (MySQL, Redis, Etcd, MinIO) should be "Up".

### Redis Authentication Error

If you see "NOAUTH Authentication required":

1. Check `etc/identity-api.yaml` has `Pass: "123456"` in Cache config
2. Restart the Identity API service

## Next Steps

- Review [Identity Service API](./identity-service.md) for complete endpoint reference
- Review [Masterdata Service API](./masterdata-service.md) for complete endpoint reference
- Check [Constants & Enums](./constants.md) for valid values and validation rules

## Tips for Frontend Development

1. **Store tokens securely**: Use httpOnly cookies or secure storage
2. **Implement token refresh**: Automatically refresh before expiration
3. **Handle 401 errors**: Redirect to login when unauthorized
4. **Validate inputs**: Check constants.md for valid enum values
5. **Show loading states**: Some operations may take time
6. **Implement retry logic**: For network failures
7. **Cache static data**: Divisions, roles, permissions change infrequently
