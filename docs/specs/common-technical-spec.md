# Community Common 库技术规范

## 1. 错误码规范

### 1.1 错误码格式

采用 **5 位数字编码**：`XXYYY`

- **XX**: 服务中心编码（2位）
- **YYYY**: 具体错误编码（4位）

### 1.2 服务中心编码

```go
const (
    ServiceCommon     = 99 // 通用服务
    ServiceUser       = 10 // 用户中心
    ServiceContent    = 20 // 内容中心
    ServiceAI         = 30 // AI中心
    ServiceModeration = 40 // 审核中心
)
```

### 1.3 通用错误码

```go
const (
    CodeSuccess       = 0      // 成功
    CodeInvalidParam  = 99400 // 参数错误
    CodeUnauthorized  = 99401 // 未授权
    CodeForbidden     = 99403 // 禁止访问
    CodeNotFound      = 99404 // 资源不存在
    CodeInternalError = 99500 // 内部错误
    CodeDatabaseError = 99501 // 数据库错误
    CodeCacheError    = 99502 // 缓存错误
    CodeRPCError      = 99503 // RPC调用错误
)
```

### 1.4 使用示例

```go
// 创建错误
err := errx.NewCodeError(99400, "用户ID不能为空")
err := errx.NewInvalidParamError("参数错误")
err := errx.NewUnauthorizedError("未授权")

// 从错误中提取信息
if ce := errx.FromError(err); ce != nil {
    fmt.Printf("错误码: %d, 错误信息: %s\n", ce.Code, ce.Msg)
}
```

## 2. HTTP 响应规范

### 2.1 响应结构

所有通过 API 网关暴露给前端的 HTTP 接口，必须严格遵循以下结构：

```go
type Body struct {
    Code int         `json:"code"` // 业务状态码：0 代表成功，非 0 代表业务异常
    Msg  string      `json:"msg"`  // 提示信息：成功为 success，失败为具体的错误提示
    Data interface{} `json:"data"` // 业务数据：成功时为具体业务对象/列表，失败时必须为 null
}
```

### 2.2 响应规则

1. **成功响应**：
   - `code` = 0
   - `msg` = "success"
   - `data` = 业务数据（不能为 null，至少为空对象 `{}`）

2. **失败响应**：
   - `code` = 错误码（5位数字）
   - `msg` = 错误描述（中文，可直接弹窗展示）
   - `data` = null（严禁夹带脏数据）

### 2.3 使用示例

```go
// 方式1：直接使用 ResponseWriter
func handler(w http.ResponseWriter, r *http.Request) {
    user, err := getUserById(id)
    if err != nil {
        responsex.Error(w, err)
        return
    }
    responsex.Success(w, user)
}

// 方式2：使用统一响应处理
func handler(w http.ResponseWriter, r *http.Request) {
    user, err := getUserById(id)
    responsex.Response(w, user, err)
}

// 方式3：使用 Context 方法（推荐）
func handler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    user, err := getUserById(ctx, id)
    if err != nil {
        responsex.CtxError(ctx, err)
        return
    }
    responsex.CtxSuccess(ctx, user)
}

// 应用响应中间件
handler := responsex.ResponseMiddleware()(mux)
```

## 3. gRPC 响应规范

### 3.1 BaseResp 定义

所有 gRPC 服务的 Response 消息必须包含 `BaseResp` 字段：

```protobuf
// api/common/v1/common.proto
message BaseResp {
  int32 code = 1;  // 业务状态码：0 成功，非 0 失败
  string msg = 2;  // 错误描述
}

// 示例：用户服务响应
message GetUserResponse {
  BaseResp base = 1;  // 必须是第一个字段
  User user = 2;
}
```

### 3.2 使用示例

```go
import (
    commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
    "github.com/guxiao1976/community-common/v2/pkg/responsex"
)

// 创建成功响应
func (s *UserService) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.GetUserResponse, error) {
    user, err := s.userRepo.GetById(ctx, req.Id)
    if err != nil {
        return &pb.GetUserResponse{
            Base: responsex.NewBaseRespFromError(err),
        }, nil
    }
    
    return &pb.GetUserResponse{
        Base: responsex.NewBaseResp(),
        User: user,
    }, nil
}

// 检查响应是否成功
if responsex.IsSuccess(resp.Base) {
    // 处理成功逻辑
}

// 将 BaseResp 转换为 error
if err := responsex.ToError(resp.Base); err != nil {
    return err
}
```

### 3.3 辅助方法

```go
// 创建成功响应
resp := responsex.NewBaseResp()
// 返回: &BaseResp{Code: 0, Msg: "success"}

// 创建错误响应
resp := responsex.NewBaseRespWithError(99400, "参数错误")

// 从 error 创建响应
resp := responsex.NewBaseRespFromError(err)

// 检查是否成功
if responsex.IsSuccess(resp) {
    // 成功逻辑
}

// 转换为 error
err := responsex.ToError(resp)
```

## 4. Proto 文件管理

### 4.1 仓库结构

Proto 文件统一管理在独立的 `api-proto` 仓库：

```
api-proto/
├── api/
│   ├── common/v1/          # 公共消息定义
│   │   └── common.proto
│   ├── user/v1/            # 用户服务
│   │   └── user.proto
│   └── ...
├── gen/go/                 # 生成的 Go 代码
├── buf.yaml                # Buf 配置
├── buf.gen.yaml            # 代码生成配置
└── Makefile
```

### 4.2 代码生成

```bash
# 生成 Go 代码
make gen

# 检查 proto 文件
make lint

# 格式化 proto 文件
make format
```

### 4.3 在微服务中使用

```go
import (
    commonv1 "github.com/guxiao1976/api-proto/gen/go/common/v1"
    userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
)
```

## 5. 本地联调

### 5.1 Go Workspace 配置

在项目根目录创建 `go.work`：

```go
go 1.25

use (
    ./code/common
    ./code/api-proto
    ./code/services/master-data
    ./code/services/identity
    // 其他微服务...
)
```

### 5.2 优势

- 无需发布到远程仓库即可本地联调
- 修改 common 库或 api-proto 立即生效
- 支持跨仓库代码跳转和重构

## 6. 版本管理

### 6.1 语义化版本

- **Major**: 破坏性变更（如 v1 → v2）
- **Minor**: 新增功能，向后兼容
- **Patch**: Bug 修复

### 6.2 模块路径

```go
// common 库 v2
module github.com/guxiao1976/community-common/v2

// api-proto
module github.com/guxiao1976/api-proto
```

### 6.3 依赖引用

```go
require (
    github.com/guxiao1976/community-common/v2 v2.0.0
    github.com/guxiao1976/api-proto v0.1.0
)
```

## 7. 测试规范

### 7.1 单元测试

- 所有公共函数必须有单元测试
- 测试覆盖率目标：>80%
- 使用表驱动测试（table-driven tests）

### 7.2 运行测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./pkg/errx -v

# 查看覆盖率
go test ./... -cover
```

## 8. 代码规范

### 8.1 包组织

```
pkg/                    # 可被外部引用的包
├── errx/              # 错误处理
├── responsex/         # 响应处理
├── jwt/               # JWT 认证
├── crypto/            # 加密组件
├── snowflake/         # ID 生成器
├── pagination/        # 分页工具
├── mask/              # 数据脱敏
└── ...

model/                 # 数据模型
├── base.go           # 基础模型
├── encrypt_type.go   # 加密类型
└── outbox.go         # Outbox 模式

examples/             # 使用示例
docs/                 # 文档
```

### 8.2 命名规范

- 包名：小写，单数形式
- 文件名：小写，下划线分隔
- 函数名：驼峰命名，导出函数首字母大写
- 常量：驼峰命名或全大写下划线分隔

## 9. 更新日志

### v2.0.0 (2024-05-29)

**破坏性变更**：
- 错误码从 HTTP 状态码风格（400, 401, 500）升级为 5 位数字格式（99400, 99401, 99500）
- responsex 的 gRPC 支持改为使用 api-proto 中的 BaseResp 定义
- IsSuccess() 和 ToError() 从方法改为函数

**新增功能**：
- 添加服务中心编码常量
- 添加 gRPC 响应辅助方法
- 完善测试覆盖

**修复**：
- 修复 NewBaseRespFromError 中硬编码的错误码
- 更新所有测试以匹配新的错误码格式
