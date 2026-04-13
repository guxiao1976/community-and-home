# Identity Service RPC Contract

**Service**: identity-rpc  
**Protocol**: gRPC  
**Port**: 8081  
**Package**: identity

## Service Definition

```protobuf
syntax = "proto3";

package identity;

option go_package = "./pb";

// Identity service for internal microservice calls
service Identity {
  // User operations
  rpc GetUser(GetUserReq) returns (GetUserResp);
  rpc GetUsersByIds(GetUsersByIdsReq) returns (GetUsersByIdsResp);
  rpc ValidateToken(ValidateTokenReq) returns (ValidateTokenResp);
  rpc CheckPermission(CheckPermissionReq) returns (CheckPermissionResp);
  
  // Role operations
  rpc GetRolesByIds(GetRolesByIdsReq) returns (GetRolesByIdsResp);
  rpc GetUserRoles(GetUserRolesReq) returns (GetUserRolesResp);
  
  // Permission operations
  rpc GetUserPermissions(GetUserPermissionsReq) returns (GetUserPermissionsResp);
  
  // Property operations
  rpc GetPropertyUnit(GetPropertyUnitReq) returns (GetPropertyUnitResp);
  rpc GetUserProperties(GetUserPropertiesReq) returns (GetUserPropertiesResp);
  
  // Verification operations
  rpc GetVerificationStatus(GetVerificationStatusReq) returns (GetVerificationStatusResp);
}
```

---

## Message Definitions

### User Operations

#### GetUserReq / GetUserResp

Get user details by ID.

```protobuf
message GetUserReq {
  int64 user_id = 1;
}

message GetUserResp {
  int64 user_id = 1;
  string phone = 2;
  string nickname = 3;
  string avatar_url = 4;
  int32 user_type = 5;        // 1=Backend, 2=Homeowner
  int32 status = 6;            // 1=Active, 2=Disabled, 3=Locked
  int32 verification_status = 7; // 0=Unverified, 1=Verified, 2=Rejected
  int64 scope_id = 8;          // Administrative scope (0=headquarters)
  string last_login_time = 9;  // RFC3339 format
  string created_time = 10;
  string updated_time = 11;
}
```

**Usage**: Other services need user information for display or validation.

---

#### GetUsersByIdsReq / GetUsersByIdsResp

Batch get users by IDs.

```protobuf
message GetUsersByIdsReq {
  repeated int64 user_ids = 1;
}

message GetUsersByIdsResp {
  repeated GetUserResp users = 1;
}
```

**Usage**: Efficiently fetch multiple users in one RPC call.

---

#### ValidateTokenReq / ValidateTokenResp

Validate JWT token and extract claims.

```protobuf
message ValidateTokenReq {
  string token = 1;
}

message ValidateTokenResp {
  bool valid = 1;
  int64 user_id = 2;
  repeated int64 role_ids = 3;
  int64 scope_id = 4;
  string expires_at = 5;  // RFC3339 format
}
```

**Usage**: API gateway validates tokens before routing requests.

---

#### CheckPermissionReq / CheckPermissionResp

Check if user has specific permission.

```protobuf
message CheckPermissionReq {
  int64 user_id = 1;
  string permission_code = 2;  // e.g., "user:create"
  int64 resource_scope_id = 3; // Optional: check scope-based permission
}

message CheckPermissionResp {
  bool allowed = 1;
  string reason = 2;  // Reason if not allowed
}
```

**Usage**: Fine-grained permission checks in business logic.

---

### Role Operations

#### GetRolesByIdsReq / GetRolesByIdsResp

Batch get roles by IDs.

```protobuf
message GetRolesByIdsReq {
  repeated int64 role_ids = 1;
}

message Role {
  int64 role_id = 1;
  string role_name = 2;
  string role_code = 3;
  string description = 4;
  bool is_system = 5;
  int32 status = 6;
}

message GetRolesByIdsResp {
  repeated Role roles = 1;
}
```

**Usage**: Display role information in other services.

---

#### GetUserRolesReq / GetUserRolesResp

Get all roles assigned to a user.

```protobuf
message GetUserRolesReq {
  int64 user_id = 1;
}

message GetUserRolesResp {
  repeated Role roles = 1;
}
```

**Usage**: Determine user's role-based capabilities.

---

### Permission Operations

#### GetUserPermissionsReq / GetUserPermissionsResp

Get all permissions for a user (aggregated from roles).

```protobuf
message GetUserPermissionsReq {
  int64 user_id = 1;
}

message Permission {
  int64 permission_id = 1;
  string permission_name = 2;
  string permission_code = 3;
  int32 type = 4;  // 1=Menu, 2=Button
  string path = 5;
  string icon = 6;
}

message GetUserPermissionsResp {
  repeated Permission permissions = 1;
}
```

**Usage**: Build user's menu tree and button permissions in frontend.

---

### Property Operations

#### GetPropertyUnitReq / GetPropertyUnitResp

Get property unit details.

```protobuf
message GetPropertyUnitReq {
  int64 property_unit_id = 1;
}

message PropertyUnit {
  int64 property_unit_id = 1;
  int64 community_id = 2;
  string building = 3;
  string unit = 4;
  string floor = 5;
  double area = 6;
  int32 property_type = 7;  // 1=Residential, 2=Commercial, 3=Mixed
  int32 status = 8;
}

message GetPropertyUnitResp {
  PropertyUnit property_unit = 1;
}
```

**Usage**: Other services need property information for business logic.

---

#### GetUserPropertiesReq / GetUserPropertiesResp

Get all properties bound to a user.

```protobuf
message GetUserPropertiesReq {
  int64 user_id = 1;
}

message PropertyBinding {
  int64 binding_id = 1;
  PropertyUnit property_unit = 2;
  bool is_primary = 3;
  int32 bind_status = 4;  // 1=Active, 2=Pending, 3=Revoked
  string bind_time = 5;
}

message GetUserPropertiesResp {
  repeated PropertyBinding bindings = 1;
}
```

**Usage**: Determine which properties a user can access.

---

### Verification Operations

#### GetVerificationStatusReq / GetVerificationStatusResp

Get homeowner verification status.

```protobuf
message GetVerificationStatusReq {
  int64 user_id = 1;
}

message GetVerificationStatusResp {
  int32 verification_status = 1;  // 0=Unverified, 1=Verified, 2=Rejected
  string real_name = 2;
  string id_card_number = 3;
  string submit_time = 4;
  string review_time = 5;
  string review_notes = 6;
}
```

**Usage**: Other services check if user is verified homeowner.

---

## Error Handling

gRPC uses standard status codes:

| Code | Status | Description |
|------|--------|-------------|
| 0 | OK | Success |
| 3 | INVALID_ARGUMENT | Invalid request parameters |
| 5 | NOT_FOUND | Resource not found |
| 7 | PERMISSION_DENIED | Insufficient permissions |
| 13 | INTERNAL | Internal server error |
| 14 | UNAVAILABLE | Service unavailable |

Error details included in status message.

---

## Service Discovery

Service registered in Etcd with key:
```
/services/identity-rpc/192.168.1.100:8081
```

Clients use go-zero's built-in service discovery to find instances.

---

## Performance Considerations

- **Caching**: Frequently accessed data (users, roles, permissions) cached in Redis
- **Batch Operations**: Use batch RPCs (GetUsersByIds, GetRolesByIds) to reduce round trips
- **Connection Pooling**: go-zero manages gRPC connection pools automatically
- **Timeout**: Default RPC timeout 3 seconds, configurable per call
- **Circuit Breaker**: Automatic circuit breaking on repeated failures

---

## Usage Example (Go)

```go
// Client initialization
identityRpc := zrpc.MustNewClient(c.IdentityRpc)
client := identity.NewIdentityClient(identityRpc.Conn())

// Validate token
resp, err := client.ValidateToken(ctx, &identity.ValidateTokenReq{
    Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
})
if err != nil {
    return err
}
if !resp.Valid {
    return errors.New("invalid token")
}

// Check permission
permResp, err := client.CheckPermission(ctx, &identity.CheckPermissionReq{
    UserId:         resp.UserId,
    PermissionCode: "user:create",
    ResourceScopeId: 23,
})
if err != nil {
    return err
}
if !permResp.Allowed {
    return errors.New(permResp.Reason)
}
```

---

## Security

- **mTLS**: Enable mutual TLS for production deployments
- **Authentication**: RPC calls authenticated via service credentials
- **Authorization**: Service-level authorization via Etcd ACL
- **Rate Limiting**: Per-service rate limiting to prevent abuse
- **Audit Logging**: All RPC calls logged with caller service and timestamp
