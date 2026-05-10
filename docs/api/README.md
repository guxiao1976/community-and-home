# API Documentation

## Overview

This documentation provides comprehensive API reference for the Community & Home Management System. The system consists of two main microservices:

- **Identity Service**: User authentication, authorization, role management, and homeowner verification
- **Masterdata Service**: Administrative divisions, communities, configurations, and sensitive word management

## Architecture

```
┌─────────────────┐         ┌─────────────────┐
│  Identity API   │         │ Masterdata API  │
│   Port: 8888    │         │   Port: 8889    │
└────────┬────────┘         └────────┬────────┘
         │                           │
         ├───────────┬───────────────┤
         │           │               │
    ┌────▼────┐ ┌───▼────┐    ┌────▼─────┐
    │  MySQL  │ │ Redis  │    │   Etcd   │
    │  :3306  │ │ :6379  │    │  :2379   │
    └─────────┘ └────────┘    └──────────┘
```

## Services

### Identity Service
- **Base URL**: `http://localhost:8888/api/identity`
- **Database**: `identity_db`
- **Features**: Authentication, user management, roles, permissions, homeowner verification, property binding

### Masterdata Service
- **Base URL**: `http://localhost:8889/api/masterdata`
- **Database**: `masterdata_db`
- **Features**: Administrative divisions, community management, system configurations, sensitive word filtering

## Authentication

### JWT Token Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your_jwt_token>
```

### Token Lifecycle

- **Access Token**: Expires in 24 hours (86400 seconds)
- **Refresh Token**: Expires in 7 days (604800 seconds)
- **Refresh Endpoint**: `POST /api/identity/auth/token/refresh`

### Public Endpoints (No Authentication Required)

- `POST /api/identity/auth/login` - Login with password
- `POST /api/identity/auth/login/sms` - Login with SMS code
- `POST /api/identity/auth/register` - User registration
- `POST /api/identity/auth/sms/send` - Send SMS verification code
- `POST /api/identity/auth/token/refresh` - Refresh access token

## Quick Links

- [Quick Start Guide](./quick-start.md) - Get started with basic API calls
- [Identity Service API](./identity-service.md) - Complete Identity API reference (28 endpoints)
- [Masterdata Service API](./masterdata-service.md) - Complete Masterdata API reference (18 endpoints)
- [Constants & Enums](./constants.md) - All status codes, enums, and validation rules

## Default Credentials

For testing and initial setup:

- **Phone**: `13800000000`
- **Password**: `Admin@123456`
- **Role**: Super Administrator

## Common Response Format

### Success Response

```json
{
  "data": { ... },
  "code": 0,
  "message": "success"
}
```

### Error Response

```json
{
  "code": 400,
  "message": "error description"
}
```

### Common Error Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 400 | Invalid Parameter |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 500 | Internal Server Error |
| 501 | Database Error |
| 502 | Cache Error |
| 503 | RPC Error |

## Pagination

List endpoints support pagination with these query parameters:

- `page` (default: 1) - Page number (1-indexed)
- `page_size` (default: 20) - Items per page

Response includes:

```json
{
  "list": [...],
  "total": 100
}
```

## Important Notes

### ⚠️ Masterdata Service Authentication

**CRITICAL**: JWT authentication is NOT yet implemented in Masterdata Service. All endpoints are currently accessible without authentication. Frontend should still send JWT tokens in preparation for future implementation.

### ⚠️ Known Issues

1. Some Identity API authenticated endpoints (roles, permissions, users list) may timeout - investigation in progress
2. Masterdata Service has TODO comments indicating pending JWT implementation

## Development Setup

### Prerequisites

- Go 1.21+
- MySQL 8.0
- Redis 7.0
- Docker & Docker Compose

### Starting Services

```bash
# Start infrastructure
docker-compose up -d

# Start Identity API
cd services/identity/api
go run identity.go -f etc/identity-api.yaml

# Start Masterdata API
cd services/masterdata/api
go run masterdata.go -f etc/masterdata-api.yaml
```

## Support

For issues or questions:
- Check the [Quick Start Guide](./quick-start.md)
- Review service-specific documentation
- Verify your JWT token is valid and not expired
- Check service logs for detailed error messages

## Version

- **API Version**: v1.0
- **Last Updated**: 2024
- **Framework**: Go-Zero 1.6+
