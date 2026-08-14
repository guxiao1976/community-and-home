# 正确模式：Proto int64 字段的 jstype 注解

## 核心原则

**所有 int64 类型的 ID 字段必须添加 `[(gogoproto.jstype) = JS_STRING]` 注解。**

## 为什么

### JavaScript 精度问题

Snowflake ID 生成的是 **19 位整数**（例如 `1234567890123456789`），但 JavaScript 的 `Number` 类型只能安全表示 **约 16 位整数**：

```javascript
// JavaScript
Number.MAX_SAFE_INTEGER  // 9007199254740991 (16 位)

// Snowflake ID
1234567890123456789  // 19 位 — 超出安全范围

// 精度丢失
const id = 1234567890123456789;
console.log(id);  // 输出: 1234567890123456800 ← 末尾变成 00
```

### JSON 传输链路

```
Go Backend (int64)
    ↓ JSON 序列化
{"user_id": 1234567890123456789}  ← 数字格式
    ↓ 前端 axios 接收
JSON.parse('{"user_id": 1234567890123456789}')
    ↓ JavaScript 解析
{user_id: 1234567890123456800}  ← 精度丢失！
    ↓ 提交回后端
Backend 收到错误的 ID，查询失败
```

### jstype 的作用

添加 `jstype = JS_STRING` 后，protojson 序列化时会将 int64 输出为字符串：

```json
{"user_id": "1234567890123456789"}
```

前端 TypeScript 声明为 `string` 类型，完全避免精度问题。

## ❌ 错误模式

```protobuf
// api-proto/api/user/v1/user.proto

syntax = "proto3";

package user.v1;

import "google/protobuf/empty.proto";
import "gogoproto/gogo.proto";

option go_package = "github.com/guxiao/community-and-home/api-proto/gen/go/user/v1;userv1";

message User {
  int64 user_id = 1;           // ❌ 缺少 jstype 注解
  string username = 2;
  string mobile = 3;
  int64 created_at = 4;        // ❌ 时间戳也会精度丢失
}

message GetUserInfoReq {
  int64 user_id = 1;           // ❌ 请求中的 ID 也需要注解
}

message CreateUserReq {
  string username = 1;
  string mobile = 2;
  int64 role_id = 3;           // ❌ 关联 ID 也需要注解
}
```

**后果**：
- Go 后端生成的 JSON：`{"user_id": 1234567890123456789}`（数字）
- 前端解析后：`{user_id: 1234567890123456800}`（精度丢失）
- 用户看到错误的 ID，提交回后端时查询失败

## ✅ 正确模式

```protobuf
// api-proto/api/user/v1/user.proto

syntax = "proto3";

package user.v1;

import "google/protobuf/empty.proto";
import "gogoproto/gogo.proto";  // ✅ 导入 gogoproto

option go_package = "github.com/guxiao/community-and-home/api-proto/gen/go/user/v1;userv1";

message User {
  int64 user_id = 1 [(gogoproto.jstype) = JS_STRING];     // ✅ 添加 jstype 注解
  string username = 2;
  string mobile = 3;
  int64 created_at = 4 [(gogoproto.jstype) = JS_STRING];  // ✅ 时间戳也需要
}

message GetUserInfoReq {
  int64 user_id = 1 [(gogoproto.jstype) = JS_STRING];     // ✅ 请求参数也需要
}

message CreateUserReq {
  string username = 1;
  string mobile = 2;
  int64 role_id = 3 [(gogoproto.jstype) = JS_STRING];     // ✅ 关联 ID 也需要
}

message ListUsersResp {
  repeated User users = 1;
  int64 total = 2;  // ❌ 注意：total 是数量，不是 ID，不需要 jstype
}
```

### 生成的 Go 代码

```go
// api-proto/gen/go/user/v1/user.pb.go

type User struct {
    UserId    int64  `protobuf:"varint,1,opt,name=user_id,json=userId,proto3" json:"user_id,omitempty" jstype:"string"`
    Username  string `protobuf:"bytes,2,opt,name=username,proto3" json:"username,omitempty"`
    Mobile    string `protobuf:"bytes,3,opt,name=mobile,proto3" json:"mobile,omitempty"`
    CreatedAt int64  `protobuf:"varint,4,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty" jstype:"string"`
}
```

### 序列化效果

```go
// Go 后端
user := &userv1.User{
    UserId:   1234567890123456789,
    Username: "alice",
}

// protojson.Marshal(user)
{
  "user_id": "1234567890123456789",  // ✅ 字符串格式
  "username": "alice"
}
```

## repeated 字段也需要 jstype

```protobuf
message BatchDeleteReq {
  // ❌ 错误：repeated 字段忘记加 jstype
  repeated int64 user_ids = 1;
}

message BatchDeleteReq {
  // ✅ 正确：repeated 字段也需要 jstype
  repeated int64 user_ids = 1 [(gogoproto.jstype) = JS_STRING];
}
```

### 生成的 JSON

```json
// ❌ 错误（数字数组）
{"user_ids": [1234567890123456789, 9876543210987654321]}

// ✅ 正确（字符串数组）
{"user_ids": ["1234567890123456789", "9876543210987654321"]}
```

## 哪些字段需要 jstype？

| 字段类型 | 是否需要 jstype | 示例 |
|---------|:-------------:|------|
| ID 字段（user_id, role_id 等） | ✅ 是 | `int64 user_id = 1 [(gogoproto.jstype) = JS_STRING];` |
| 时间戳（created_at, updated_at） | ✅ 是 | `int64 created_at = 2 [(gogoproto.jstype) = JS_STRING];` |
| 数量/计数（total, count） | ❌ 否 | `int64 total = 3;` |
| 金额/价格（amount, price） | ⚠️ 取决 | 如果用分（cents）存储且超过 16 位，需要 jstype |
| 版本号（version） | ❌ 否 | `int64 version = 4;` |
| repeated ID | ✅ 是 | `repeated int64 user_ids = 5 [(gogoproto.jstype) = JS_STRING];` |

### 判断标准

**如果字段值可能超过 `9007199254740991`（16 位），必须加 jstype。**

Snowflake ID 的范围：
```
最小值: 1000000000000000000 (19 位)
最大值: 9223372036854775807 (19 位)
```

显然远超 JavaScript 安全整数范围。

## 完整示例

### 示例 1：User Service Proto

**文件**: `api-proto/api/user/v1/user.proto:15-45`

```protobuf
message User {
  int64 user_id = 1 [(gogoproto.jstype) = JS_STRING];
  string username = 2;
  string mobile = 3;
  string avatar = 4;
  int64 role_id = 5 [(gogoproto.jstype) = JS_STRING];
  int32 status = 6;  // 状态码：小整数，不需要 jstype
  int64 created_at = 7 [(gogoproto.jstype) = JS_STRING];
  int64 updated_at = 8 [(gogoproto.jstype) = JS_STRING];
}

message GetUserInfoReq {
  int64 user_id = 1 [(gogoproto.jstype) = JS_STRING];
}

message ListUsersResp {
  repeated User users = 1;
  int64 total = 2;  // 数量：虽然是 int64，但不会超过 16 位，可以不加 jstype
                    // 如果严格执行规范，建议所有 int64 都加 jstype
}

message BatchDeleteUsersReq {
  repeated int64 user_ids = 1 [(gogoproto.jstype) = JS_STRING];
}
```

### 示例 2：Auth Service Proto

**文件**: `api-proto/api/auth/v1/auth.proto:22-40`

```protobuf
message LoginResp {
  string access_token = 1;
  string refresh_token = 2;
  int64 user_id = 3 [(gogoproto.jstype) = JS_STRING];
  int64 expires_at = 4 [(gogoproto.jstype) = JS_STRING];
}

message TokenClaims {
  int64 user_id = 1 [(gogoproto.jstype) = JS_STRING];
  int64 role_id = 2 [(gogoproto.jstype) = JS_STRING];
  int64 issued_at = 3 [(gogoproto.jstype) = JS_STRING];
  int64 expires_at = 4 [(gogoproto.jstype) = JS_STRING];
}
```

## 前端配套

### TypeScript 类型声明

```typescript
// web/pc/src/api/user.ts

export interface User {
  user_id: string;      // ✅ 声明为 string
  username: string;
  mobile: string;
  role_id: string;      // ✅ 关联 ID 也是 string
  status: number;       // 状态码是小整数，用 number
  created_at: string;   // ✅ 时间戳是 string
  updated_at: string;
}

export interface GetUserInfoReq {
  user_id: string;      // ✅ 请求参数也是 string
}
```

### axios 配置（已配置 lossless-json）

```typescript
// web/pc/src/utils/request.ts

import axios from 'axios';
import LosslessJSON from 'lossless-json';  // ✅ 使用无损 JSON 解析器

const request = axios.create({
  transformResponse: [(data) => {
    if (typeof data === 'string') {
      return LosslessJSON.parse(data, (key, value) => {
        // 如果是 LosslessNumber 且字段名包含 'id' 或以 '_at' 结尾，转为字符串
        if (value && value.isLosslessNumber) {
          if (key.includes('id') || key.endsWith('_at')) {
            return value.toString();
          }
        }
        return value;
      });
    }
    return data;
  }],
});
```

## 常见问题

### Q1: 为什么不在 Go 代码中统一加 `json:",string"` 标签？

**A**: Proto 是跨语言的接口契约，jstype 确保所有语言（Go/TypeScript/Java 等）都以字符串序列化。Go 的 `json:",string"` 只影响 Go 的 JSON 序列化，不影响 protojson。

### Q2: total/count 这种字段需要加 jstype 吗？

**A**: 
- **严格模式**（推荐）：所有 int64 字段都加 jstype，避免判断失误
- **宽松模式**：只有 ID 和时间戳加 jstype，数量字段不加（假设不会超过 16 位）

本项目采用**严格模式** — 所有 int64 都加 jstype，除非明确是小整数（如 status/type 枚举）。

### Q3: 如果忘记加 jstype 会怎样？

**A**: 
- 开发环境：QA 检查第 4 项会失败，阻止提交
- 生产环境：用户会看到错误的 ID，查询/更新失败

## 检查清单

修复 Proto jstype 违规时，确认以下步骤：

- [ ] 在 proto 文件顶部导入 `import "gogoproto/gogo.proto";`
- [ ] 所有 ID 字段（xxx_id）添加 `[(gogoproto.jstype) = JS_STRING]`
- [ ] 所有时间戳字段（xxx_at）添加 `[(gogoproto.jstype) = JS_STRING]`
- [ ] repeated ID 字段也添加 jstype
- [ ] 运行 `cd api-proto && make generate` 生成代码
- [ ] 运行 `cd api-proto && make lint` 验证规范
- [ ] 前端 TypeScript 类型声明对应字段为 `string`
- [ ] 运行 `bash .harness/skills/qa/scripts/harness-checks.sh` 验证

## 相关文档

- [项目编码规范 §5 — Snowflake ID 序列化规范](../../rules/项目编码规范.md#5-snowflake-id-序列化规范)
- [Proto 管理规范](../../rules/Proto管理规范.md)
- [Memory: proto-jstype](../../knowledge/memory/proto-jstype.md)
- [json-string.md](./json-string.md) — Go REST API 配套规范
