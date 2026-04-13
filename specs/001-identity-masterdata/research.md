# Research: Identity and Masterdata Microservices

**Feature**: 001-identity-masterdata  
**Date**: 2026-04-13  
**Purpose**: Technical research and decision documentation for implementation planning

## Research Areas

### 1. go-zero Framework Best Practices

**Decision**: Use go-zero 1.6+ with standard three-layer architecture (API → Logic → Model)

**Rationale**:
- go-zero provides built-in support for microservices patterns (service discovery, load balancing, circuit breaker)
- goctl code generation ensures consistency and reduces boilerplate
- Built-in Redis caching integration with automatic cache invalidation
- Native gRPC support for inter-service communication
- Proven at scale (used by companies handling millions of requests)

**Alternatives Considered**:
- **go-micro**: More flexible but requires more manual configuration; go-zero's opinionated structure better fits constitutional requirements
- **go-kit**: Lower-level toolkit requiring more boilerplate; go-zero provides higher productivity
- **Kratos (Bilibili)**: Similar capabilities but go-zero has better Chinese documentation and community support

**Implementation Notes**:
- Use `goctl api new` for API services
- Use `goctl rpc new` for RPC services  
- Use `goctl model mysql datasource -cache=true` for Model generation
- Enable Redis cache for all Model operations to meet P99 ≤ 200ms requirement

---

### 2. JWT Token Management Strategy

**Decision**: Use golang-jwt/jwt v5 with Redis-based token blacklist for revocation

**Rationale**:
- JWT provides stateless authentication suitable for microservices
- Redis blacklist enables immediate permission revocation (FR-007 requirement)
- Token expiration + refresh token pattern balances security and UX
- go-zero middleware integration for automatic token validation

**Alternatives Considered**:
- **Session-based auth**: Requires sticky sessions, not suitable for distributed microservices
- **OAuth2**: Overkill for internal backend users; adds unnecessary complexity
- **Paseto**: Better security but less ecosystem support; JWT sufficient for internal use

**Implementation Notes**:
- Access token TTL: 2 hours
- Refresh token TTL: 7 days
- Store token blacklist in Redis with TTL matching token expiration
- Include user ID, role IDs, and administrative scope in JWT claims
- Middleware checks Redis blacklist on every request for revoked tokens

---

### 3. Role-Based Access Control (RBAC) Implementation

**Decision**: Casbin for policy enforcement with go-zero middleware integration

**Rationale**:
- Casbin provides flexible RBAC/ABAC model supporting menu and button-level permissions
- Policy stored in MySQL, cached in Redis for performance
- Supports hierarchical roles and dynamic permission updates
- Native Go implementation with good performance

**Alternatives Considered**:
- **Custom RBAC**: Reinventing the wheel; Casbin is battle-tested
- **OPA (Open Policy Agent)**: More powerful but heavier; Casbin sufficient for our needs
- **go-zero built-in auth**: Too basic; doesn't support fine-grained permissions

**Implementation Notes**:
- Use RBAC with domains model (domain = administrative scope)
- Policy format: `p, role, resource, action, domain`
- Cache policies in Redis with 5-minute TTL
- Reload policies on permission changes via Redis pub/sub
- Middleware enforces policies before reaching Logic layer

---

### 4. Administrative Scope Isolation

**Decision**: Database-level row filtering using scope_id in WHERE clauses

**Rationale**:
- Ensures 100% enforcement of scope-based data isolation (SC-002)
- Prevents accidental data leakage through SQL injection or logic errors
- Leverages MySQL indexes for performance
- Simple to implement and audit

**Alternatives Considered**:
- **Database views per scope**: Complex to manage with 100k communities
- **Application-level filtering**: Risk of developer error bypassing checks
- **Multi-tenancy database**: Overkill; single database with row-level filtering sufficient

**Implementation Notes**:
- Add `scope_id` column to all scope-sensitive tables
- Index on `scope_id` for query performance
- Middleware extracts scope from JWT and injects into context
- Model layer automatically appends `WHERE scope_id IN (...)` to queries
- Headquarters users have scope_id = NULL (access all scopes)

---

### 5. Soft Delete and Audit Logging

**Decision**: Use `delete_time` column (nullable timestamp) + separate audit_log table

**Rationale**:
- Soft delete preserves data for audit and recovery (FR-012)
- Separate audit table prevents performance impact on main tables
- Timestamp-based approach simpler than boolean flags
- Supports compliance requirements

**Alternatives Considered**:
- **Boolean deleted flag**: Doesn't capture when deletion occurred
- **Audit columns in main tables**: Bloats main tables, impacts query performance
- **Event sourcing**: Overkill for this use case; adds significant complexity

**Implementation Notes**:
- Add `delete_time TIMESTAMP NULL DEFAULT NULL` to all core tables
- Queries filter `WHERE delete_time IS NULL` by default
- Audit log captures: user_id, entity_type, entity_id, action, old_value, new_value, timestamp
- Use database triggers for automatic audit logging (ensures no bypass)
- Partition audit_log table by month for performance

---

### 6. MinIO Integration for Image Storage

**Decision**: Use MinIO Go SDK with presigned URLs for upload/download

**Rationale**:
- MinIO already deployed via Docker (infrastructure ready)
- S3-compatible API with excellent Go SDK
- Presigned URLs offload upload/download traffic from API servers
- Supports 5MB file size limit and format validation

**Alternatives Considered**:
- **Direct upload to API**: Increases API server load and memory usage
- **Local filesystem**: Not scalable, no redundancy
- **Cloud storage (Aliyun OSS)**: MinIO sufficient for on-premise deployment

**Implementation Notes**:
- Generate presigned upload URL (15-minute expiration)
- Client uploads directly to MinIO using presigned URL
- API validates file metadata (size, format) before generating URL
- Store file path in `uploaded_file` table with reference to entity
- Use bucket naming: `identity-verification`, `property-documents`
- Enable MinIO versioning for accidental deletion recovery

---

### 7. Database Schema Design for 5-Tier Hierarchy

**Decision**: Adjacency list model with materialized path for efficient queries

**Rationale**:
- Adjacency list (parent_id) simple to maintain
- Materialized path (e.g., `/1/23/456/`) enables efficient subtree queries
- Supports 100k communities without performance issues
- MySQL indexes on both parent_id and path for different query patterns

**Alternatives Considered**:
- **Nested sets**: Complex to maintain with frequent updates
- **Closure table**: Requires additional table, more storage overhead
- **Pure adjacency list**: Recursive queries slow for deep hierarchies

**Implementation Notes**:
- Table: `md_administrative_division`
- Columns: `id`, `parent_id`, `level` (1-5), `name`, `code`, `path`
- Index on `parent_id` for children queries
- Index on `path` for subtree queries (e.g., `WHERE path LIKE '/1/23/%'`)
- Trigger updates path on parent_id change
- Cache full hierarchy tree in Redis (refresh on change)

---

### 8. SMS Verification Code Integration

**Decision**: Abstract SMS gateway behind interface, use Aliyun SMS as default provider

**Rationale**:
- Interface allows swapping providers without code changes
- Aliyun SMS widely used in China, reliable and cost-effective
- Rate limiting prevents abuse (max 5 codes per phone per hour)
- Verification codes stored in Redis with 5-minute TTL

**Alternatives Considered**:
- **Twilio**: Better international coverage but more expensive in China
- **Tencent Cloud SMS**: Similar to Aliyun; chose Aliyun for existing infrastructure
- **Mock SMS in dev**: Still need real provider for production

**Implementation Notes**:
- Interface: `type SMSGateway interface { SendCode(phone, code string) error }`
- Store code in Redis: `sms:verify:{phone}` with 5-minute TTL
- Rate limit: `sms:limit:{phone}` counter with 1-hour TTL
- 6-digit numeric code, cryptographically random
- Log all SMS sends for audit and cost tracking

---

### 9. API Response Format Standardization

**Decision**: Unified JSON response wrapper with code, message, data fields

**Rationale**:
- Consistent error handling across all APIs
- Frontend can handle responses uniformly
- Supports i18n for error messages
- Compatible with go-zero's error handling

**Format**:
```json
{
  "code": 0,           // 0 = success, >0 = error code
  "message": "success", // Human-readable message
  "data": {}           // Response payload (null on error)
}
```

**Error Code Ranges**:
- 0: Success
- 1000-1999: Client errors (validation, auth, permission)
- 2000-2999: Server errors (database, external service)
- 3000-3999: Business logic errors (duplicate, not found, conflict)

**Implementation Notes**:
- Create `common/responsex` package with helper functions
- Middleware catches panics and returns 500 error
- Log all errors with request ID for tracing
- Include request_id in response for debugging

---

### 10. Performance Optimization Strategy

**Decision**: Multi-layer caching (Redis + local cache) with cache-aside pattern

**Rationale**:
- Meets P99 ≤ 200ms API latency requirement
- Redis cache for shared data (roles, permissions, admin divisions)
- Local cache for read-heavy, rarely-changing data (config, static data)
- go-zero's built-in cache integration simplifies implementation

**Caching Strategy**:
- **Hot data** (roles, permissions): Redis with 5-minute TTL
- **Warm data** (admin divisions): Redis with 1-hour TTL
- **Cold data** (audit logs): No cache, query on demand
- **User sessions**: Redis with token TTL

**Implementation Notes**:
- Use go-zero's `sqlc` and `sqlx` with automatic cache
- Cache invalidation via Redis pub/sub on data changes
- Monitor cache hit rate (target >90% for hot data)
- Use Redis cluster for high availability

---

## Technology Stack Summary

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Language | Go | 1.21+ | Primary development language |
| Framework | go-zero | 1.6+ | Microservices framework |
| API Protocol | HTTP/REST | - | External API exposure |
| RPC Protocol | gRPC | - | Inter-service communication |
| Database | MySQL | 8.0 | Relational data storage |
| Cache | Redis | 7.0 | Session, cache, pub/sub |
| Object Storage | MinIO | Latest | Image file storage |
| Service Discovery | Etcd | Latest | Service registry |
| Authentication | JWT | golang-jwt/jwt v5 | Token-based auth |
| Authorization | Casbin | v2 | RBAC policy enforcement |
| Password Hashing | bcrypt | golang.org/x/crypto | Secure password storage |
| Testing | testify | v1.8+ | Unit and integration tests |
| SMS Gateway | Aliyun SMS | Latest SDK | Verification code delivery |

---

## Open Questions Resolved

All technical unknowns from the specification have been resolved through this research phase. No blocking questions remain for Phase 1 design.
