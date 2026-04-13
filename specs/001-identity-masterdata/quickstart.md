# Quickstart Guide: Identity and Masterdata Microservices

**Feature**: 001-identity-masterdata  
**Date**: 2026-04-13  
**Purpose**: Step-by-step guide to set up, develop, and test the microservices

## Prerequisites

Ensure the following are installed and running:

- **Go**: 1.21+ (`go version`)
- **goctl**: go-zero code generator (`goctl --version`)
- **Docker**: For infrastructure services
- **MySQL**: 8.0 (running via Docker)
- **Redis**: 7.0 (running via Docker)
- **Etcd**: Latest (running via Docker)
- **MinIO**: Latest (running via Docker)

Verify infrastructure is running:
```bash
docker ps | grep -E "mysql|redis|etcd|minio"
```

---

## Step 1: Install goctl

Install go-zero code generator:

```bash
# Install goctl
go install github.com/zeromicro/go-zero/tools/goctl@latest

# Verify installation
goctl --version
```

---

## Step 2: Create Project Structure

Create the microservices directory structure:

```bash
# From repository root
mkdir -p services/identity/{api,rpc,model}
mkdir -p services/masterdata/{api,rpc,model}
mkdir -p common/{errorx,jwtx,responsex,miniox}
```

---

## Step 3: Create Database Schemas

### Create Databases

```bash
# Connect to MySQL
docker exec -it <mysql-container-id> mysql -uroot -p

# Create databases
CREATE DATABASE identity_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE DATABASE masterdata_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### Run DDL Scripts

Create DDL scripts based on `data-model.md` and execute:

```bash
# Identity database
mysql -h127.0.0.1 -P3306 -uroot -p identity_db < scripts/sql/identity_schema.sql

# Masterdata database
mysql -h127.0.0.1 -P3306 -uroot -p masterdata_db < scripts/sql/masterdata_schema.sql
```

### Seed Initial Data

```bash
# Insert super admin user, default roles, and permissions
mysql -h127.0.0.1 -P3306 -uroot -p identity_db < scripts/sql/identity_seed.sql

# Insert national administrative divisions
mysql -h127.0.0.1 -P3306 -uroot -p masterdata_db < scripts/sql/masterdata_seed.sql
```

---

## Step 4: Generate Model Code

Use goctl to generate Model layer from database:

```bash
# Identity service models
cd services/identity/model
goctl model mysql datasource \
  -url="root:password@tcp(127.0.0.1:3306)/identity_db" \
  -table="auth_*" \
  -dir="." \
  -cache=true

# Masterdata service models
cd ../../masterdata/model
goctl model mysql datasource \
  -url="root:password@tcp(127.0.0.1:3306)/masterdata_db" \
  -table="md_*" \
  -dir="." \
  -cache=true
```

**Result**: Generated files like `authusermodel.go`, `authrolemodel.go`, etc.

---

## Step 5: Define API Contracts

### Identity API Definition

Create `services/identity/api/identity.api`:

```go
syntax = "v1"

info(
    title: "Identity Service API"
    desc: "Authentication and authorization API"
    author: "Community Platform Team"
    version: "v1"
)

type (
    LoginReq {
        Phone    string `json:"phone"`
        Password string `json:"password"`
    }
    
    LoginResp {
        AccessToken  string `json:"access_token"`
        RefreshToken string `json:"refresh_token"`
        ExpiresIn    int64  `json:"expires_in"`
        User         UserInfo `json:"user"`
    }
    
    UserInfo {
        UserId             int64    `json:"user_id"`
        Phone              string   `json:"phone"`
        Nickname           string   `json:"nickname"`
        UserType           int32    `json:"user_type"`
        VerificationStatus int32    `json:"verification_status"`
        Roles              []string `json:"roles"`
        Permissions        []string `json:"permissions"`
    }
)

@server(
    prefix: /api/v1
)
service identity-api {
    @handler Login
    post /auth/login (LoginReq) returns (LoginResp)
}

// Add more endpoints based on contracts/identity-api.md
```

### Generate API Code

```bash
cd services/identity/api
goctl api go -api identity.api -dir .
```

**Result**: Generated `handler/`, `logic/`, `types/`, `svc/` directories

---

## Step 6: Define RPC Contracts

### Identity RPC Definition

Create `services/identity/rpc/identity.proto`:

```protobuf
syntax = "proto3";

package identity;

option go_package = "./pb";

message GetUserReq {
  int64 user_id = 1;
}

message GetUserResp {
  int64 user_id = 1;
  string phone = 2;
  string nickname = 3;
  int32 user_type = 4;
  int32 status = 5;
}

service Identity {
  rpc GetUser(GetUserReq) returns (GetUserResp);
  // Add more methods based on contracts/identity-rpc.md
}
```

### Generate RPC Code

```bash
cd services/identity/rpc
goctl rpc protoc identity.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

**Result**: Generated `pb/`, `internal/`, `identityclient/` directories

---

## Step 7: Implement Business Logic

### Example: Login Logic

Edit `services/identity/api/internal/logic/loginlogic.go`:

```go
package logic

import (
    "context"
    "errors"
    
    "golang.org/x/crypto/bcrypt"
    
    "your-project/services/identity/api/internal/svc"
    "your-project/services/identity/api/internal/types"
    "your-project/common/jwtx"
    
    "github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
    return &LoginLogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
    // 1. Find user by phone
    user, err := l.svcCtx.AuthUserModel.FindOneByPhone(l.ctx, req.Phone)
    if err != nil {
        return nil, errors.New("user not found")
    }
    
    // 2. Verify password
    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
    if err != nil {
        return nil, errors.New("invalid password")
    }
    
    // 3. Generate JWT tokens
    accessToken, err := jwtx.GenerateToken(user.Id, user.ScopeId, 7200)
    if err != nil {
        return nil, err
    }
    
    refreshToken, err := jwtx.GenerateToken(user.Id, user.ScopeId, 604800)
    if err != nil {
        return nil, err
    }
    
    // 4. Get user roles and permissions
    roles, _ := l.svcCtx.AuthUserRoleModel.FindByUserId(l.ctx, user.Id)
    permissions, _ := l.getUserPermissions(user.Id)
    
    return &types.LoginResp{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    7200,
        User: types.UserInfo{
            UserId:             user.Id,
            Phone:              user.Phone,
            Nickname:           user.Nickname.String,
            UserType:           user.UserType,
            VerificationStatus: user.VerificationStatus,
            Roles:              roles,
            Permissions:        permissions,
        },
    }, nil
}
```

**Note**: Implement all Logic layer methods based on functional requirements in `spec.md`.

---

## Step 8: Configure Services

### Identity API Configuration

Edit `services/identity/api/etc/identity-api.yaml`:

```yaml
Name: identity-api
Host: 0.0.0.0
Port: 8080

# MySQL
Mysql:
  DataSource: root:password@tcp(127.0.0.1:3306)/identity_db?charset=utf8mb4&parseTime=true

# Redis
Redis:
  Host: 127.0.0.1:6379
  Type: node
  Pass: ""

# JWT
Auth:
  AccessSecret: your-access-secret-key-change-in-production
  AccessExpire: 7200

# RPC clients
MasterdataRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: masterdata-rpc
```

### Identity RPC Configuration

Edit `services/identity/rpc/etc/identity-rpc.yaml`:

```yaml
Name: identity-rpc
ListenOn: 0.0.0.0:8081

# Etcd for service discovery
Etcd:
  Hosts:
    - 127.0.0.1:2379
  Key: identity-rpc

# MySQL
Mysql:
  DataSource: root:password@tcp(127.0.0.1:3306)/identity_db?charset=utf8mb4&parseTime=true

# Redis
Redis:
  Host: 127.0.0.1:6379
  Type: node
  Pass: ""
```

---

## Step 9: Run Services

### Start Identity RPC Service

```bash
cd services/identity/rpc
go run identity.go -f etc/identity-rpc.yaml
```

**Expected Output**:
```
Starting rpc server at 0.0.0.0:8081...
Service registered in etcd: identity-rpc
```

### Start Identity API Service

```bash
cd services/identity/api
go run identity.go -f etc/identity-api.yaml
```

**Expected Output**:
```
Starting server at 0.0.0.0:8080...
```

### Start Masterdata Services

Repeat the same process for masterdata API and RPC services.

---

## Step 10: Test APIs

### Test Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "password": "Admin123!"
  }'
```

**Expected Response**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 7200,
    "user": {
      "user_id": 1,
      "phone": "13800138000",
      "nickname": "超级管理员",
      "user_type": 1,
      "verification_status": 1,
      "roles": ["SUPER_ADMIN"],
      "permissions": ["*:*"]
    }
  }
}
```

### Test Authenticated Endpoint

```bash
TOKEN="<access_token_from_login>"

curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN"
```

---

## Step 11: Run Tests

### Unit Tests

```bash
# Test specific logic
cd services/identity/api/internal/logic
go test -v

# Test with coverage
go test -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Run integration tests
cd services/identity
go test -tags=integration ./...
```

---

## Step 12: Build and Deploy

### Build Binaries

```bash
# Build identity services
cd services/identity/api
go build -o identity-api identity.go

cd ../rpc
go build -o identity-rpc identity.go

# Build masterdata services
cd ../../masterdata/api
go build -o masterdata-api masterdata.go

cd ../rpc
go build -o masterdata-rpc masterdata.go
```

### Docker Build

Create `services/identity/Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o identity-api api/identity.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/identity-api .
COPY --from=builder /app/api/etc/identity-api.yaml etc/

EXPOSE 8080
CMD ["./identity-api", "-f", "etc/identity-api.yaml"]
```

Build and run:

```bash
docker build -t identity-api:latest -f services/identity/Dockerfile .
docker run -p 8080:8080 identity-api:latest
```

---

## Step 13: Verify Deployment

### Health Check

```bash
# Check service health
curl http://localhost:8080/health

# Check Etcd registration
etcdctl get /services/identity-rpc --prefix
```

### Monitor Logs

```bash
# View API logs
tail -f logs/identity-api.log

# View RPC logs
tail -f logs/identity-rpc.log
```

---

## Common Issues & Solutions

### Issue: goctl command not found
**Solution**: Ensure `$GOPATH/bin` is in your PATH:
```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

### Issue: MySQL connection refused
**Solution**: Check Docker container is running and port is exposed:
```bash
docker ps | grep mysql
docker logs <mysql-container-id>
```

### Issue: Etcd connection failed
**Solution**: Verify Etcd is accessible:
```bash
etcdctl endpoint health
```

### Issue: Redis connection timeout
**Solution**: Check Redis is running and accessible:
```bash
docker exec -it <redis-container-id> redis-cli ping
```

### Issue: Model generation fails
**Solution**: Ensure database tables exist and connection string is correct:
```bash
mysql -h127.0.0.1 -P3306 -uroot -p -e "SHOW TABLES FROM identity_db;"
```

---

## Next Steps

1. **Implement remaining endpoints**: Follow `contracts/*.md` to implement all API endpoints
2. **Add middleware**: JWT validation, permission checking, rate limiting
3. **Write tests**: Achieve 85%+ unit test coverage
4. **Performance testing**: Use tools like `wrk` or `ab` to verify P99 ≤ 200ms
5. **Deploy to staging**: Test in staging environment before production
6. **Set up monitoring**: Prometheus metrics, Grafana dashboards
7. **Configure CI/CD**: Automated testing and deployment pipeline

---

## Useful Commands

```bash
# Generate API code
goctl api go -api identity.api -dir .

# Generate RPC code
goctl rpc protoc identity.proto --go_out=. --go-grpc_out=. --zrpc_out=.

# Generate Model code
goctl model mysql datasource -url="..." -table="auth_*" -dir="." -cache=true

# Format code
gofmt -w .

# Run linter
golangci-lint run

# Run tests with coverage
go test -cover ./...

# Build for production
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .
```

---

## Reference Documentation

- **Spec**: `spec.md` - Business requirements
- **Data Model**: `data-model.md` - Database schema
- **API Contracts**: `contracts/identity-api.md`, `contracts/masterdata-api.md`
- **RPC Contracts**: `contracts/identity-rpc.md`, `contracts/masterdata-rpc.md`
- **Research**: `research.md` - Technical decisions
- **go-zero Docs**: https://go-zero.dev/
- **goctl Guide**: https://go-zero.dev/docs/goctl/goctl
