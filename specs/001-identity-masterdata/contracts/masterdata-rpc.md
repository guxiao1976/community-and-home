# Masterdata Service RPC Contract

**Service**: masterdata-rpc  
**Protocol**: gRPC  
**Port**: 8082  
**Package**: masterdata

## Service Definition

```protobuf
syntax = "proto3";

package masterdata;

option go_package = "./pb";

// Masterdata service for internal microservice calls
service Masterdata {
  // Administrative division operations
  rpc GetDivision(GetDivisionReq) returns (GetDivisionResp);
  rpc GetDivisionsByIds(GetDivisionsByIdsReq) returns (GetDivisionsByIdsResp);
  rpc GetDivisionTree(GetDivisionTreeReq) returns (GetDivisionTreeResp);
  rpc GetDivisionPath(GetDivisionPathReq) returns (GetDivisionPathResp);
  rpc ValidateScope(ValidateScopeReq) returns (ValidateScopeResp);
  
  // Community operations
  rpc GetCommunity(GetCommunityReq) returns (GetCommunityResp);
  rpc GetCommunitiesByIds(GetCommunitiesByIdsReq) returns (GetCommunitiesByIdsResp);
  rpc GetCommunitiesByDivision(GetCommunitiesByDivisionReq) returns (GetCommunitiesByDivisionResp);
  
  // Configuration operations
  rpc GetConfig(GetConfigReq) returns (GetConfigResp);
  rpc GetConfigsByModule(GetConfigsByModuleReq) returns (GetConfigsByModuleResp);
  
  // Sensitive word operations
  rpc CheckSensitiveWords(CheckSensitiveWordsReq) returns (CheckSensitiveWordsResp);
}
```

---

## Message Definitions

### Administrative Division Operations

#### GetDivisionReq / GetDivisionResp

Get division details by ID.

```protobuf
message GetDivisionReq {
  int64 division_id = 1;
}

message GetDivisionResp {
  int64 division_id = 1;
  int64 parent_id = 2;
  int32 level = 3;          // 1-5
  string name = 4;
  string code = 5;
  string path = 6;
  int32 sort_order = 7;
  int32 status = 8;
  string created_time = 9;
  string updated_time = 10;
}
```

**Usage**: Get division information for display or validation.

---

#### GetDivisionsByIdsReq / GetDivisionsByIdsResp

Batch get divisions by IDs.

```protobuf
message GetDivisionsByIdsReq {
  repeated int64 division_ids = 1;
}

message GetDivisionsByIdsResp {
  repeated GetDivisionResp divisions = 1;
}
```

**Usage**: Efficiently fetch multiple divisions in one RPC call.

---

#### GetDivisionTreeReq / GetDivisionTreeResp

Get division tree structure.

```protobuf
message GetDivisionTreeReq {
  int64 root_id = 1;        // Root division ID (0 for all)
  int32 max_level = 2;      // Max depth (0 for unlimited)
  bool include_inactive = 3; // Include inactive divisions
}

message DivisionNode {
  int64 division_id = 1;
  int64 parent_id = 2;
  int32 level = 3;
  string name = 4;
  string code = 5;
  int32 status = 6;
  repeated DivisionNode children = 7;
}

message GetDivisionTreeResp {
  repeated DivisionNode nodes = 1;
}
```

**Usage**: Build hierarchical division selector in frontend.

---

#### GetDivisionPathReq / GetDivisionPathResp

Get full path from root to division.

```protobuf
message GetDivisionPathReq {
  int64 division_id = 1;
  string separator = 2;  // Default: "/"
}

message GetDivisionPathResp {
  string path = 1;              // e.g., "北京市/市辖区/朝阳区"
  repeated int64 division_ids = 2; // IDs in path order
  repeated string names = 3;    // Names in path order
}
```

**Usage**: Display full administrative path for addresses.

---

#### ValidateScopeReq / ValidateScopeResp

Validate if resource is within user's administrative scope.

```protobuf
message ValidateScopeReq {
  int64 user_scope_id = 1;      // User's scope (0=headquarters)
  int64 resource_scope_id = 2;  // Resource's scope
}

message ValidateScopeResp {
  bool allowed = 1;
  string reason = 2;  // Reason if not allowed
}
```

**Usage**: Enforce scope-based access control in business logic.

**Logic**:
- If user_scope_id = 0 (headquarters): allowed = true
- If user_scope_id = resource_scope_id: allowed = true
- If resource_scope_id in user's subtree: allowed = true
- Otherwise: allowed = false

---

### Community Operations

#### GetCommunityReq / GetCommunityResp

Get community details by ID.

```protobuf
message GetCommunityReq {
  int64 community_id = 1;
}

message GetCommunityResp {
  int64 community_id = 1;
  int64 division_id = 2;
  string name = 3;
  string address = 4;
  double area = 5;
  int64 population = 6;
  int32 community_type = 7;  // 1=Residential, 2=Village, 3=Mixed
  int32 submission_status = 8; // 0=Draft, 1=Submitted, 2=Approved, 3=Rejected
  string created_time = 9;
  string updated_time = 10;
}
```

**Usage**: Get community information for property management.

---

#### GetCommunitiesByIdsReq / GetCommunitiesByIdsResp

Batch get communities by IDs.

```protobuf
message GetCommunitiesByIdsReq {
  repeated int64 community_ids = 1;
}

message GetCommunitiesByIdsResp {
  repeated GetCommunityResp communities = 1;
}
```

**Usage**: Efficiently fetch multiple communities.

---

#### GetCommunitiesByDivisionReq / GetCommunitiesByDivisionResp

Get all communities under a division.

```protobuf
message GetCommunitiesByDivisionReq {
  int64 division_id = 1;
  bool include_subtree = 2;  // Include communities in child divisions
  int32 status_filter = 3;   // Filter by submission_status (-1=all)
}

message GetCommunitiesByDivisionResp {
  repeated GetCommunityResp communities = 1;
  int64 total = 2;
}
```

**Usage**: List communities for a specific region.

---

### Configuration Operations

#### GetConfigReq / GetConfigResp

Get configuration value by module and key.

```protobuf
message GetConfigReq {
  string module = 1;
  string config_key = 2;
}

message GetConfigResp {
  int64 config_id = 1;
  string module = 2;
  string config_key = 3;
  string config_value = 4;
  string value_type = 5;  // string/number/boolean/json
  string description = 6;
  bool is_public = 7;
  int32 approval_status = 8;
}
```

**Usage**: Retrieve configuration values for business logic.

---

#### GetConfigsByModuleReq / GetConfigsByModuleResp

Get all configurations for a module.

```protobuf
message GetConfigsByModuleReq {
  string module = 1;
  bool approved_only = 2;  // Only return approved configs
}

message GetConfigsByModuleResp {
  repeated GetConfigResp configs = 1;
}
```

**Usage**: Load all module configurations at service startup.

---

### Sensitive Word Operations

#### CheckSensitiveWordsReq / CheckSensitiveWordsResp

Check text for sensitive words.

```protobuf
message CheckSensitiveWordsReq {
  string text = 1;
  repeated string categories = 2;  // Filter by categories (empty=all)
}

message SensitiveWordMatch {
  string word = 1;
  string category = 2;
  int32 severity = 3;    // 1=Low, 2=Medium, 3=High
  int32 action = 4;      // 1=Warn, 2=Block, 3=Review
  int32 position = 5;    // Position in text
}

message CheckSensitiveWordsResp {
  bool has_sensitive_words = 1;
  repeated SensitiveWordMatch matches = 2;
  int32 max_severity = 3;        // Highest severity found
  int32 recommended_action = 4;  // Recommended action based on max severity
}
```

**Usage**: Content moderation before publishing user-generated content.

**Logic**:
- Scan text for all active sensitive words
- Return all matches with positions
- Recommended action = highest action among matches

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

---

## Service Discovery

Service registered in Etcd with key:
```
/services/masterdata-rpc/192.168.1.101:8082
```

Clients use go-zero's built-in service discovery.

---

## Performance Considerations

- **Caching**: Division tree, configurations, and sensitive words cached in Redis
- **Batch Operations**: Use batch RPCs to reduce round trips
- **Tree Queries**: Division tree cached for 1 hour, invalidated on changes
- **Sensitive Word Matching**: Use Aho-Corasick algorithm for efficient multi-pattern matching
- **Timeout**: Default RPC timeout 3 seconds

---

## Usage Example (Go)

```go
// Client initialization
masterdataRpc := zrpc.MustNewClient(c.MasterdataRpc)
client := masterdata.NewMasterdataClient(masterdataRpc.Conn())

// Validate scope
scopeResp, err := client.ValidateScope(ctx, &masterdata.ValidateScopeReq{
    UserScopeId:     23,  // Beijing
    ResourceScopeId: 456, // Chaoyang District
})
if err != nil {
    return err
}
if !scopeResp.Allowed {
    return errors.New(scopeResp.Reason)
}

// Get division path
pathResp, err := client.GetDivisionPath(ctx, &masterdata.GetDivisionPathReq{
    DivisionId: 456,
    Separator:  "/",
})
if err != nil {
    return err
}
fmt.Println(pathResp.Path) // "北京市/市辖区/朝阳区"

// Check sensitive words
checkResp, err := client.CheckSensitiveWords(ctx, &masterdata.CheckSensitiveWordsReq{
    Text: "用户输入的内容",
})
if err != nil {
    return err
}
if checkResp.HasSensitiveWords {
    if checkResp.RecommendedAction == 2 {
        return errors.New("content contains blocked words")
    }
}
```

---

## Caching Strategy

### Division Tree
- **Cache Key**: `masterdata:division:tree:{root_id}`
- **TTL**: 1 hour
- **Invalidation**: On any division create/update/delete
- **Preload**: Full tree preloaded at service startup

### Configurations
- **Cache Key**: `masterdata:config:{module}:{key}`
- **TTL**: 5 minutes
- **Invalidation**: On configuration approval
- **Preload**: All approved configs preloaded at startup

### Sensitive Words
- **Cache Key**: `masterdata:sensitive:words`
- **TTL**: 10 minutes
- **Invalidation**: On word create/update/delete
- **Structure**: Aho-Corasick automaton for O(n) matching

### Communities
- **Cache Key**: `masterdata:community:{id}`
- **TTL**: 30 minutes
- **Invalidation**: On community update/approval

---

## Security

- **mTLS**: Enable mutual TLS for production
- **Service Authentication**: RPC calls authenticated via service credentials
- **Rate Limiting**: Per-service rate limiting
- **Audit Logging**: All RPC calls logged with caller service

---

## Monitoring

Key metrics to monitor:
- RPC call latency (P50, P95, P99)
- Cache hit rate (target >90%)
- Division tree query performance
- Sensitive word matching performance
- Error rate by RPC method
