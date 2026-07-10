# Mock 设置指南 — 完善测试用例

本文档说明如何为已创建的测试模板完善 Mock 设置。

---

## 一、概述

**当前状态**:
- ✅ 测试框架完整
- ✅ gRPC Mock 已生成 (7个服务)
- ✅ 测试模板已创建 (3个服务, 27个用例)
- ⚠️ Mock EXPECT 设置待完善

**目标**: 为测试用例添加完整的 Mock EXPECT 设置，使测试可以独立运行。

---

## 二、Mock 设置模式

### 2.1 基本模式

```go
mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
    // 1. 创建 Mock 对象
    mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)
    
    // 2. 设置期望调用
    mockUserRpc.EXPECT().
        CreateUser(
            gomock.Any(),  // context.Context
            gomock.Any(),  // request
        ).
        Return(&userv1.CreateUserResponse{UserId: 1}, nil)
    
    // 3. 注入到 ServiceContext
    return &svc.ServiceContext{
        UserRpc: mockUserRpc,
    }
}
```

### 2.2 精确匹配模式

```go
mockUserRpc.EXPECT().
    CreateUser(
        gomock.Any(),
        &userv1.CreateUserRequest{
            Phone:    "13800138000",
            Nickname: "测试用户",
        },
    ).
    Return(&userv1.CreateUserResponse{UserId: 1}, nil)
```

### 2.3 自定义匹配器

```go
mockUserRpc.EXPECT().
    CreateUser(
        gomock.Any(),
        gomock.AssignableToTypeOf(&userv1.CreateUserRequest{}),
    ).
    DoAndReturn(func(ctx context.Context, req *userv1.CreateUserRequest) (*userv1.CreateUserResponse, error) {
        // 自定义逻辑
        if req.Phone == "" {
            return nil, errors.New("phone is required")
        }
        return &userv1.CreateUserResponse{UserId: 1}, nil
    })
```

---

## 三、示例：user-service

### 3.1 CreateUser Mock 设置

```go
// File: services/user-service/api/internal/logic/user/user_logic_test.go

func TestCreateUserLogic_CreateUser(t *testing.T) {
    tests := []struct {
        name      string
        req       *types.CreateUserReq
        mockSetup func(*gomock.Controller) *svc.ServiceContext
        want      *types.CreateUserResp
        wantErr   bool
    }{
        {
            name: "success - 正常创建用户",
            req: &types.CreateUserReq{
                Phone:    "13800138000",
                Nickname: "测试用户",
            },
            mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
                mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)
                
                mockUserRpc.EXPECT().
                    CreateUser(
                        gomock.Any(),
                        &userv1.CreateUserRequest{
                            Phone:    "13800138000",
                            Nickname: "测试用户",
                        },
                    ).
                    Return(&userv1.CreateUserResponse{
                        UserId: 1,
                    }, nil)
                
                return &svc.ServiceContext{
                    UserRpc: mockUserRpc,
                }
            },
            want: &types.CreateUserResp{
                UserId: 1,
            },
            wantErr: false,
        },
        {
            name: "error - 手机号为空",
            req: &types.CreateUserReq{
                Phone:    "",
                Nickname: "测试用户",
            },
            mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
                // 不设置 EXPECT，因为 Logic 层应该在调用 RPC 前验证
                return &svc.ServiceContext{}
            },
            wantErr: true,
        },
    }
    
    // ... 测试执行代码
}
```

### 3.2 GetUser Mock 设置

```go
{
    name: "success - 正常获取用户",
    req: &types.GetUserReq{
        Id: 1,
    },
    mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
        mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)
        
        mockUserRpc.EXPECT().
            GetUser(
                gomock.Any(),
                &userv1.GetUserRequest{
                    Id: 1,
                },
            ).
            Return(&userv1.GetUserResponse{
                User: &userv1.User{
                    Id:       1,
                    Phone:    "13800138000",
                    Nickname: "测试用户",
                },
            }, nil)
        
        return &svc.ServiceContext{
            UserRpc: mockUserRpc,
        }
    },
    want: &types.GetUserResp{
        User: types.UserInfo{
            Id:       1,
            Nickname: "测试用户",
            Phone:    "13800138000",
        },
    },
    wantErr: false,
},
```

---

## 四、示例：auth-service

### 4.1 Login Mock 设置

```go
{
    name: "success - 正常登录",
    req: &types.LoginReq{
        Phone:    "13800138000",
        Password: "password123",
    },
    mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
        mockAuthRpc := authmocks.NewMockAuthServiceClient(ctrl)
        
        mockAuthRpc.EXPECT().
            Login(
                gomock.Any(),
                &authv1.LoginRequest{
                    Phone:    "13800138000",
                    Password: "password123",
                },
            ).
            Return(&authv1.LoginResponse{
                AccessToken:  "mock_access_token",
                RefreshToken: "mock_refresh_token",
                ExpiresIn:    7200,
            }, nil)
        
        return &svc.ServiceContext{
            AuthRpc: mockAuthRpc,
        }
    },
    want: &types.LoginResp{
        AccessToken:  "mock_access_token",
        RefreshToken: "mock_refresh_token",
        ExpiresIn:    7200,
    },
    wantErr: false,
},
```

---

## 五、测试辅助函数

创建通用的 Mock 辅助函数简化测试编写：

```go
// File: services/user-service/api/internal/logic/user/test_helpers.go

package user

import (
    "github.com/golang/mock/gomock"
    
    userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
    usermocks "github.com/guxiao1976/api-proto/gen/go/user/v1/mocks"
    "github.com/guxiao1976/community-user/api/internal/svc"
)

// NewMockServiceContext 创建一个带有默认 Mock 的 ServiceContext
func NewMockServiceContext(ctrl *gomock.Controller) (*svc.ServiceContext, *usermocks.MockUserServiceClient) {
    mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)
    
    svcCtx := &svc.ServiceContext{
        UserRpc: mockUserRpc,
    }
    
    return svcCtx, mockUserRpc
}

// SetupCreateUserSuccess 设置 CreateUser 成功的 Mock
func SetupCreateUserSuccess(mockUserRpc *usermocks.MockUserServiceClient, userId int64) {
    mockUserRpc.EXPECT().
        CreateUser(gomock.Any(), gomock.Any()).
        Return(&userv1.CreateUserResponse{UserId: userId}, nil)
}

// SetupGetUserSuccess 设置 GetUser 成功的 Mock
func SetupGetUserSuccess(mockUserRpc *usermocks.MockUserServiceClient, user *userv1.User) {
    mockUserRpc.EXPECT().
        GetUser(gomock.Any(), gomock.Any()).
        Return(&userv1.GetUserResponse{User: user}, nil)
}
```

使用示例：

```go
mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
    svcCtx, mockUserRpc := NewMockServiceContext(ctrl)
    SetupCreateUserSuccess(mockUserRpc, 1)
    return svcCtx
}
```

---

## 六、常见问题

### Q1: Mock 调用次数不匹配

**问题**: `missing call(s) to *MockUserServiceClient.CreateUser`

**原因**: EXPECT 设置了但实际没调用，或调用次数不匹配

**解决**:
```go
// 允许任意次数调用
mockUserRpc.EXPECT().CreateUser(gomock.Any(), gomock.Any()).AnyTimes()

// 精确次数
mockUserRpc.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(2)

// 至少一次
mockUserRpc.EXPECT().CreateUser(gomock.Any(), gomock.Any()).MinTimes(1)
```

### Q2: 参数不匹配

**问题**: `Unexpected call to CreateUser`

**原因**: 实际参数与 EXPECT 设置的不匹配

**解决**:
```go
// 使用 Any() 匹配任意参数
mockUserRpc.EXPECT().CreateUser(gomock.Any(), gomock.Any())

// 或使用自定义匹配器
mockUserRpc.EXPECT().CreateUser(
    gomock.Any(),
    gomock.AssignableToTypeOf(&userv1.CreateUserRequest{}),
)
```

### Q3: Nil pointer dereference

**问题**: `panic: runtime error: invalid memory address`

**原因**: ServiceContext 中的依赖为 nil

**解决**:
```go
// 确保所有依赖都设置了 Mock
return &svc.ServiceContext{
    UserRpc:  mockUserRpc,  // 不要遗漏
    Config:   &config.Config{},  // 如果需要
    // ... 其他依赖
}
```

---

## 七、运行测试

### 7.1 运行单个测试

```bash
cd services/user-service
go test ./api/internal/logic/user -v -run TestCreateUserLogic_CreateUser
```

### 7.2 运行所有测试

```bash
cd services/user-service
go test ./api/internal/logic/... -v
```

### 7.3 生成覆盖率报告

```bash
# 生成覆盖率文件
go test ./api/internal/logic/... -coverprofile=coverage.out

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html
```

---

## 八、下一步

1. **完善 user-service 测试**
   - 为所有测试用例添加 Mock EXPECT
   - 运行测试验证
   - 生成覆盖率报告

2. **完善 auth-service 测试**
   - 添加 Mock EXPECT
   - 验证登录/登出/刷新流程

3. **完善 moderation-service 测试**
   - 添加 Mock EXPECT
   - 验证审核流程

4. **持续改进**
   - 根据实际 Logic 调整 Mock
   - 补充边界条件测试
   - 提高覆盖率

---

## 九、参考资料

- gomock 文档: https://github.com/golang/mock
- Table-driven tests: https://github.com/golang/go/wiki/TableDrivenTests
- go-zero 测试最佳实践: https://go-zero.dev/docs/tutorials/test

---

**文档版本**: 1.0  
**最后更新**: 2026-06-23  
**状态**: ✅ 可用
