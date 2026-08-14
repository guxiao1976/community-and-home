# 正确模式：Go REST API json:",string" 标签

## 核心原则

**Go REST API 中所有 int64 类型的 ID 字段必须使用 `json:",string"` 标签。**

## 为什么

这是 Proto jstype 规范的 **Go REST API 层配套**。

### 完整链路

```
Proto Definition (int64 + jstype=JS_STRING)
    ↓ buf generate
Go gRPC Code (int64 in Go, string in JSON)
    ↓ REST API Handler
Go REST API types.go (需要手动加 json:",string")
    ↓ JSON 序列化
{"user_id": "1234567890123456789"}
    ↓ 前端接收
TypeScript (string 类型)
```

- **Proto 层**：`[(gogoproto.jstype) = JS_STRING]` — protojson 序列化为字符串
- **Go REST API 层**：`json:"user_id,string"` — Go 标准库 JSON 序列化为字符串
- **前端层**：`user_id: string` — TypeScript 类型声明

### 为什么需要两层都加？

1. **gRPC 接口**（RPC Internal）：使用 protojson 序列化，jstype 生效
2. **REST API 接口**（HTTP External）：使用 Go 标准库 `encoding/json` 序列化，需要 `json:",string"` 标签

## ❌ 错误模式

```go
// services/user-service/api/internal/types/types.go

type User struct {
    UserId    int64  `json:"user_id"`           // ❌ 缺少 string 选项
    Username  string `json:"username"`
    Mobile    string `json:"mobile"`
    RoleId    int64  `json:"role_id"`           // ❌ 关联 ID 也缺少
    Status    int32  `json:"status"`
    CreatedAt int64  `json:"created_at"`        // ❌ 时间戳也缺少
    UpdatedAt int64  `json:"updated_at"`        // ❌ 时间戳也缺少
}

type GetUserInfoRequest struct {
    UserId int64 `json:"user_id"`              // ❌ 请求参数也缺少
}

type CreateUserRequest struct {
    Username string `json:"username"`
    Mobile   string `json:"mobile"`
    RoleId   int64  `json:"role_id"`           // ❌ 关联 ID 也缺少
}
```

**后果**：
```json
// Go 序列化输出（数字格式）
{"user_id": 1234567890123456789}

// JavaScript 解析后（精度丢失）
{user_id: 1234567890123456800}
```

## ✅ 正确模式

```go
// services/user-service/api/internal/types/types.go

type User struct {
    UserId    int64  `json:"user_id,string"`           // ✅ 添加 string 选项
    Username  string `json:"username"`
    Mobile    string `json:"mobile"`
    RoleId    int64  `json:"role_id,string"`           // ✅ 关联 ID 也需要
    Status    int32  `json:"status"`                   // 状态码是小整数，不需要
    CreatedAt int64  `json:"created_at,string"`        // ✅ 时间戳也需要
    UpdatedAt int64  `json:"updated_at,string"`        // ✅ 时间戳也需要
}

type GetUserInfoRequest struct {
    UserId int64 `json:"user_id,string"`              // ✅ 请求参数也需要
}

type CreateUserRequest struct {
    Username string `json:"username"`
    Mobile   string `json:"mobile"`
    RoleId   int64  `json:"role_id,string"`           // ✅ 关联 ID 也需要
}
```

### 序列化效果

```go
// Go 代码
user := &User{
    UserId:   1234567890123456789,
    Username: "alice",
}

// json.Marshal(user)
{
  "user_id": "1234567890123456789",  // ✅ 字符串格式
  "username": "alice"
}
```

## 与 omitempty 组合使用

```go
type User struct {
    UserId    int64  `json:"user_id,omitempty,string"`    // ✅ string 在最后
    RoleId    int64  `json:"role_id,string,omitempty"`    // ✅ 位置可以互换
    CreatedAt int64  `json:"created_at,string"`           // ✅ 不需要 omitempty 时
}
```

**规则**：
- `string` 和 `omitempty` 可以任意顺序
- 推荐写法：`json:"field_name,omitempty,string"` 或 `json:"field_name,string"`

## path/form/header 标签不需要 string

```go
type GetUserInfoRequest struct {
    UserId int64 `path:"user_id"`       // ✅ path 参数不需要 string（已是字符串）
}

type LoginRequest struct {
    Mobile   string `form:"mobile"`
    Password string `form:"password"`
    DeviceId int64  `form:"device_id"`  // ✅ form 参数不需要 string（已是字符串）
}

type AuthHeader struct {
    UserId int64 `header:"X-User-Id"`   // ✅ header 不需要 string（已是字符串）
}
```

**原因**：HTTP 路径、表单、Header 都是字符串传输，go-zero 框架会自动转换为 int64。

**只有 JSON body 序列化需要 `json:",string"` 标签。**

## db 标签不需要 string

```go
// services/user-service/rpc/model/user_model.go

type User struct {
    UserId    int64  `db:"user_id"`     // ✅ 数据库字段不需要 string
    Username  string `db:"username"`
    CreatedAt int64  `db:"created_at"`  // ✅ 数据库存储 BIGINT，不需要 string
}
```

**原因**：数据库层使用 int64 存储，不涉及 JSON 序列化。

## 完整示例

### 示例 1：User Service Types

**文件**: `services/user-service/api/internal/types/types.go:18-45`

```go
type User struct {
    UserId    int64  `json:"user_id,string"`
    Username  string `json:"username"`
    Nickname  string `json:"nickname,omitempty"`
    Mobile    string `json:"mobile"`
    Avatar    string `json:"avatar,omitempty"`
    RoleId    int64  `json:"role_id,string"`
    Status    int32  `json:"status"`
    CreatedAt int64  `json:"created_at,string"`
    UpdatedAt int64  `json:"updated_at,string"`
}

type GetUserInfoRequest struct {
    UserId int64 `path:"user_id"`  // path 不需要 string
}

type GetUserInfoResponse struct {
    BaseResponse
    User *User `json:"user"`
}

type ListUsersRequest struct {
    Page     int32  `form:"page,default=1"`
    PageSize int32  `form:"page_size,default=20"`
    RoleId   int64  `form:"role_id,optional"`  // form 不需要 string
}

type ListUsersResponse struct {
    BaseResponse
    Users []*User `json:"users"`
    Total int64   `json:"total,string"`  // ✅ 如果严格规范，total 也加 string
}

type CreateUserRequest struct {
    Username string `json:"username"`
    Mobile   string `json:"mobile"`
    Password string `json:"password"`
    RoleId   int64  `json:"role_id,string"`  // ✅ JSON body 需要 string
}

type CreateUserResponse struct {
    BaseResponse
    UserId int64 `json:"user_id,string"`  // ✅ 返回的 ID 需要 string
}

type BatchDeleteUsersRequest struct {
    UserIds []int64 `json:"user_ids"`  // ❌ 错误：slice 也需要特殊处理
}
```

### 示例 2：Auth Service Types

**文件**: `services/auth-service/api/internal/types/types.go:25-50`

```go
type LoginRequest struct {
    Mobile   string `json:"mobile"`
    Password string `json:"password"`
}

type LoginResponse struct {
    BaseResponse
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    UserId       int64  `json:"user_id,string"`    // ✅ 用户 ID
    ExpiresAt    int64  `json:"expires_at,string"` // ✅ 过期时间戳
}

type TokenClaims struct {
    UserId    int64 `json:"user_id,string"`
    RoleId    int64 `json:"role_id,string"`
    IssuedAt  int64 `json:"issued_at,string"`
    ExpiresAt int64 `json:"expires_at,string"`
}
```

## []int64 切片的特殊处理

**问题**：Go 标准库不支持 `[]int64` 的 `json:",string"` 标签。

```go
type BatchDeleteUsersRequest struct {
    UserIds []int64 `json:"user_ids,string"`  // ❌ 无效！Go 不支持
}
```

**解决方案 1**：使用自定义类型

```go
type Int64String int64

func (i Int64String) MarshalJSON() ([]byte, error) {
    return []byte(fmt.Sprintf("\"%d\"", i)), nil
}

func (i *Int64String) UnmarshalJSON(data []byte) error {
    s := strings.Trim(string(data), "\"")
    val, err := strconv.ParseInt(s, 10, 64)
    if err != nil {
        return err
    }
    *i = Int64String(val)
    return nil
}

type BatchDeleteUsersRequest struct {
    UserIds []Int64String `json:"user_ids"`  // ✅ 使用自定义类型
}
```

**解决方案 2**：前端发送字符串数组

```typescript
// 前端
await batchDeleteUsers({
  user_ids: ["1234567890123456789", "9876543210987654321"]  // 字符串数组
});
```

```go
// 后端接收
type BatchDeleteUsersRequest struct {
    UserIds []string `json:"user_ids"`  // ✅ 接收字符串数组
}

func (l *BatchDeleteUsersLogic) BatchDeleteUsers(req *types.BatchDeleteUsersRequest) error {
    // 转换为 int64
    ids := make([]int64, len(req.UserIds))
    for i, s := range req.UserIds {
        id, err := strconv.ParseInt(s, 10, 64)
        if err != nil {
            return errx.NewCodeError(40000, "无效的用户ID")
        }
        ids[i] = id
    }
    
    // 使用 ids...
}
```

## 常见问题

### Q1: goctl 生成的 types.go 会自动加 json:",string" 吗？

**A**: 不会。goctl 只根据 .api 文件生成基本的 json 标签。你需要手动添加 `,string` 选项。

### Q2: 如果忘记加 json:",string" 会怎样？

**A**: 
- 开发环境：QA 检查第 5 项会失败，阻止提交
- 生产环境：前端收到数字格式的 ID，精度丢失

### Q3: total/count 这种字段需要加 json:",string" 吗？

**A**: 
- **严格模式**（推荐）：所有 int64 都加 `,string`
- **宽松模式**：只有 ID 和时间戳加 `,string`

本项目采用**严格模式**。

## 相关文档

- [项目编码规范 §5 — Snowflake ID 序列化规范](../../rules/项目编码规范.md#5-snowflake-id-序列化规范)
- [proto-jstype.md](./proto-jstype.md) — Proto 层配套规范
- [Memory: proto-jstype](../../knowledge/memory/proto-jstype.md)

## 检查清单

修复 json:",string" 违规时，确认以下步骤：

- [ ] 所有 int64 ID 字段添加 `json:"field_name,string"`
- [ ] 所有 int64 时间戳字段添加 `json:"field_name,string"`
- [ ] 确认 path/form/header 标签**不需要** string
- [ ] 确认 db 标签**不需要** string
- [ ] 如果有 []int64 字段，考虑改为 []string 或自定义类型
- [ ] 运行 `bash .harness/skills/qa/scripts/harness-checks.sh --service <name>` 验证
