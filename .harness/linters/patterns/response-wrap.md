# 正确模式：REST API 单层包装

## 核心原则

**Logic 返回纯业务数据，Handler 用 `response.Success()` 统一包装一层。禁止双层嵌套。**

## 为什么

### 双层嵌套问题

```json
// ❌ 双层嵌套（错误）
{
  "code": 0,
  "msg": "success",
  "data": {
    "code": 0,          // ← 第二层 code/msg/data
    "msg": "success",
    "data": {
      "user_id": "123",
      "username": "alice"
    }
  }
}

// ✅ 单层包装（正确）
{
  "code": 0,
  "msg": "success",
  "data": {
    "user_id": "123",    // ← 直接是业务数据
    "username": "alice"
  }
}
```

**问题根源**：
1. goctl 生成的 Response 类型嵌入了 `BaseResponse`
2. Logic 返回了这个 Response 类型
3. Handler 又用 `response.Success()` 包了一层
4. 结果：`{code, msg, data: {code, msg, data: {...}}}`

## ❌ 错误模式

### goctl 生成的 types.go

```go
// services/user-service/api/internal/types/types.go

type BaseResponse struct {
    Code int32  `json:"code"`
    Msg  string `json:"msg"`
}

type GetUserInfoResponse struct {
    BaseResponse                          // ❌ 嵌入 BaseResponse
    Data *User `json:"data,omitempty"`   // ❌ 已经有 data 字段
}
```

### 错误的 Logic 实现

```go
// services/user-service/api/internal/logic/get_user_info_logic.go

func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoRequest) (*types.GetUserInfoResponse, error) {
    user, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &pb.GetUserInfoReq{
        UserId: req.UserId,
    })
    if err != nil {
        return nil, err
    }
    
    // ❌ 返回 goctl 生成的 Response 类型
    return &types.GetUserInfoResponse{
        BaseResponse: types.BaseResponse{Code: 0, Msg: "success"},  // ❌ 第一层包装
        Data: &types.User{
            UserId:   user.UserId,
            Username: user.Username,
        },
    }, nil
}
```

### 错误的 Handler 实现

```go
// services/user-service/api/internal/handler/get_user_info_handler.go

func GetUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.GetUserInfoRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }
        
        l := logic.NewGetUserInfoLogic(r.Context(), svcCtx)
        resp, err := l.GetUserInfo(&req)
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
        } else {
            response.Success(w, resp)  // ❌ 第二层包装（resp 已经有 BaseResponse）
        }
    }
}
```

### 输出结果（双层嵌套）

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "code": 0,        // ← 多余的第二层
    "msg": "success",
    "data": {
      "user_id": "123",
      "username": "alice"
    }
  }
}
```

### 前端的混乱

```typescript
// 前端代码不知道该取 res.data 还是 res.data.data
const res = await getUserInfo({user_id: "123"});

// ❌ 混乱写法
const user = res.data?.data;  // 双层 data
const user = (res as any)?.data?.data;  // 类型断言
```

## ✅ 正确模式

### 方案 1：Logic 返回纯业务数据（推荐）

#### 步骤 1：Logic 返回业务实体

```go
// services/user-service/api/internal/logic/get_user_info_logic.go

// ✅ 返回 *User，不返回 *GetUserInfoResponse
func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoRequest) (*types.User, error) {
    userResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &pb.GetUserInfoReq{
        UserId: req.UserId,
    })
    if err != nil {
        return nil, errx.Wrap(err, "获取用户信息失败")
    }
    
    // ✅ 直接返回业务数据
    return &types.User{
        UserId:   userResp.User.UserId,
        Username: userResp.User.Username,
        Mobile:   userResp.User.Mobile,
        RoleId:   userResp.User.RoleId,
        Status:   userResp.User.Status,
    }, nil
}
```

#### 步骤 2：Handler 统一包装

```go
// services/user-service/api/internal/handler/get_user_info_handler.go

func GetUserInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.GetUserInfoRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }
        
        l := logic.NewGetUserInfoLogic(r.Context(), svcCtx)
        data, err := l.GetUserInfo(&req)  // ← data 是 *User
        if err != nil {
            response.Error(w, err)
        } else {
            response.Success(w, data)  // ✅ 只包一层
        }
    }
}
```

#### 输出结果（单层包装）

```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "user_id": "123",
    "username": "alice",
    "mobile": "13800138000",
    "role_id": "456",
    "status": 1
  }
}
```

### 方案 2：Logic 返回 map（灵活场景）

适用于返回多个不同类型的字段。

```go
func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoRequest) (interface{}, error) {
    // 业务逻辑...
    
    // ✅ 返回 map 或自定义结构体
    return map[string]interface{}{
        "user": user,
        "role": role,
        "permissions": permissions,
    }, nil
}
```

### 方案 3：修改 types.go（不推荐，但可行）

如果必须使用 goctl 生成的 Response 类型：

```go
// types.go — 移除 BaseResponse 嵌入
type GetUserInfoResponse struct {
    User *User `json:"user"`  // ✅ 不嵌入 BaseResponse
}

// Logic 返回
return &types.GetUserInfoResponse{
    User: &types.User{...},
}, nil

// Handler 包装
response.Success(w, resp)  // ✅ 只包一层
```

## 完整示例

### 示例 1：AI Model Service - CreateModel

**文件**: `services/ai-model-service/api/internal/logic/create_model_logic.go:25-50`

```go
// ✅ 返回 *model.AmModelConfig（数据库实体），不返回 Response 类型
func (l *CreateModelLogic) CreateModel(req *types.CreateModelRequest) (*model.AmModelConfig, error) {
    // 参数校验
    if req.ModelName == "" {
        return nil, errx.NewCodeError(40000, "模型名称不能为空")
    }
    
    // 调用 RPC 创建
    createResp, err := l.svcCtx.AiModelRpc.CreateModelConfig(l.ctx, &pb.CreateModelConfigReq{
        ModelName: req.ModelName,
        ModelType: req.ModelType,
        ApiKey:    req.ApiKey,
        // ...
    })
    if err != nil {
        return nil, errx.Wrap(err, "创建模型配置失败")
    }
    
    // ✅ 直接返回业务数据
    return &model.AmModelConfig{
        Id:        createResp.Id,
        ModelName: createResp.ModelName,
        ModelType: createResp.ModelType,
        // ...
    }, nil
}
```

```go
// Handler
func CreateModelHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req types.CreateModelRequest
        if err := httpx.Parse(r, &req); err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }
        
        l := logic.NewCreateModelLogic(r.Context(), svcCtx)
        data, err := l.CreateModel(&req)
        if err != nil {
            response.Error(w, err)
        } else {
            response.Success(w, data)  // ✅ 只包一层
        }
    }
}
```

### 示例 2：列表接口返回多个字段

```go
// ✅ 定义纯业务数据结构（不嵌入 BaseResponse）
type ListUsersData struct {
    Users []*User `json:"users"`
    Total int64   `json:"total,string"`
}

func (l *ListUsersLogic) ListUsers(req *types.ListUsersRequest) (*ListUsersData, error) {
    // 查询逻辑...
    
    return &ListUsersData{
        Users: users,
        Total: total,
    }, nil
}
```

```json
// 输出
{
  "code": 0,
  "msg": "success",
  "data": {
    "users": [...],
    "total": "150"
  }
}
```

## 前端配套

### axios 拦截器（单层提取）

```typescript
// web/pc/src/utils/request.ts

axios.interceptors.response.use(
  (response) => {
    const res = response.data;
    
    if (res.code !== 0) {
      // 错误处理...
      return Promise.reject(new Error(res.msg || 'Error'));
    }
    
    // ✅ 提取 data 字段返回（单层提取）
    return res.data;
  },
  (error) => {
    // 错误处理...
    return Promise.reject(error);
  }
);
```

### 前端调用（清晰简洁）

```typescript
// web/pc/src/api/user.ts

export const getUserInfo = (params: { user_id: string }) => {
  return request.get<User>('/api/v1/user/info', { params });
};

// 组件中使用
const user = await getUserInfo({ user_id: '123' });
console.log(user.username);  // ✅ 直接访问，无需 user.data.username
```

## 常见问题

### Q1: goctl 生成的 Response 类型能删掉吗？

**A**: 可以。goctl 生成的 Response 类型只是模板，你可以：
- 删除 Response 类型定义
- 修改 Logic 返回类型为纯业务数据
- 或者移除 Response 中的 BaseResponse 嵌入

### Q2: 什么时候 Logic 应该返回 Response 类型？

**A**: **永远不应该**。Logic 只负责业务逻辑，返回纯业务数据。Handler 负责统一包装为 HTTP 响应格式。

### Q3: 如果 Logic 需要返回多个字段怎么办？

**A**: 定义一个纯业务数据结构（不嵌入 BaseResponse）：

```go
type CreateUserData struct {
    UserId string `json:"user_id"`
    Token  string `json:"token"`
}

func (l *CreateUserLogic) CreateUser(...) (*CreateUserData, error) {
    // ...
    return &CreateUserData{UserId: userId, Token: token}, nil
}
```

## 相关文档

- [项目编码规范 §9 — REST API 响应格式](../../rules/项目编码规范.md#9-rest-api-响应格式)
- [Memory: response-single-wrap](../../knowledge/memory/response-single-wrap.md)

## 检查清单

修复双层嵌套违规时，确认以下步骤：

- [ ] Logic 返回类型**不是** `*types.XxxResponse`（goctl 生成的）
- [ ] Logic 返回纯业务数据（struct/pointer/map）
- [ ] Handler 中使用 `response.Success(w, data)` 统一包装
- [ ] 前端 axios 拦截器提取 `res.data` 返回
- [ ] 前端调用代码直接使用返回值，无需 `.data.data`
- [ ] 运行 `bash .harness/skills/qa/scripts/harness-checks.sh --service <name>` 验证
