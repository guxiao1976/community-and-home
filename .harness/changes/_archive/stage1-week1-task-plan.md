# Week 1 任务执行计划 — API 层测试框架搭建

**目标**: 为 3 个核心服务建立 API 测试模板，覆盖率达到 30%

**执行时间**: 2026-06-23
**预计完成**: 1 周内

---

## 一、任务拆解

### 任务 1.1: 搭建测试基础设施

**目标**: 建立 gomock + httptest 测试框架

**步骤**:
1. 安装 gomock 工具
2. 为 Model 层生成 mock 接口
3. 创建测试工具函数库

**产出**:
```
tools/
├── install-gomock.sh           # gomock 安装脚本
├── generate-mocks.sh           # 批量生成 mock
└── testing/
    ├── mock_helpers.go         # Mock 辅助函数
    └── assert_helpers.go       # 断言辅助函数
```

---

### 任务 1.2: user-service API 层测试

**目标**: 用户服务 API logic 层 30% 覆盖率

**优先接口** (按业务重要性):
1. ✅ user_logic.go — 用户信息查询（已有 RPC 层测试）
2. 创建用户（如存在）
3. 更新用户（如存在）
4. 社区角色管理

**测试模板**:
```go
// services/user-service/api/internal/logic/user/user_logic_test.go

package user

import (
    "context"
    "testing"
    
    "github.com/golang/mock/gomock"
    "github.com/stretchr/testify/assert"
    
    "github.com/guxiao1976/community-user/api/internal/svc"
    "github.com/guxiao1976/community-user/api/internal/types"
    "github.com/guxiao1976/community-user/model"
    "github.com/guxiao1976/community-user/model/mocks"
)

func TestUserLogic_GetUserInfo(t *testing.T) {
    tests := []struct {
        name      string
        userId    int64
        mockSetup func(*mocks.MockUserModel)
        want      *types.UserInfoResp
        wantErr   bool
    }{
        {
            name:   "success - 正常获取用户信息",
            userId: 1,
            mockSetup: func(m *mocks.MockUserModel) {
                m.EXPECT().
                    FindOne(gomock.Any(), int64(1)).
                    Return(&model.User{
                        Id:       1,
                        Nickname: "测试用户",
                        Phone:    "13800138000",
                    }, nil)
            },
            want: &types.UserInfoResp{
                Id:       1,
                Nickname: "测试用户",
                Phone:    "13800138000",
            },
            wantErr: false,
        },
        {
            name:   "error - 用户不存在",
            userId: 999,
            mockSetup: func(m *mocks.MockUserModel) {
                m.EXPECT().
                    FindOne(gomock.Any(), int64(999)).
                    Return(nil, model.ErrNotFound)
            },
            wantErr: true,
        },
        {
            name:   "error - 数据库错误",
            userId: 1,
            mockSetup: func(m *mocks.MockUserModel) {
                m.EXPECT().
                    FindOne(gomock.Any(), int64(1)).
                    Return(nil, errors.New("db error"))
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockUserModel := mocks.NewMockUserModel(ctrl)
            tt.mockSetup(mockUserModel)
            
            // Create logic with mock
            l := NewUserLogic(context.Background(), &svc.ServiceContext{
                UserModel: mockUserModel,
            })
            
            // Execute
            got, err := l.GetUserInfo(&types.UserInfoReq{UserId: tt.userId})
            
            // Assert
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            
            assert.NoError(t, err)
            assert.Equal(t, tt.want.Id, got.Id)
            assert.Equal(t, tt.want.Nickname, got.Nickname)
        })
    }
}
```

---

### 任务 1.3: auth-service API 层测试

**目标**: 认证服务 API logic 层 30% 覆盖率

**优先接口**:
1. 登录 (LoginLogic)
2. 刷新 Token (RefreshTokenLogic)
3. 登出 (LogoutLogic)

**关键测试场景**:
- 正常登录
- 手机号不存在
- 密码错误
- Token 生成
- Redis 缓存测试

---

### 任务 1.4: moderation-service API 层测试

**目标**: 审核服务 API logic 层 30% 覆盖率

**优先接口**:
1. 文本审核 (ReviewTextLogic) 
2. 图片审核 (ReviewImageLogic)
3. 审核结果查询

**关键测试场景**:
- 正常审核流程
- 敏感词检测
- 管线配置加载
- 审核结果持久化

---

## 二、技术方案

### 2.1 Mock 生成策略

**为什么使用 gomock?**
- ✅ 官方推荐，社区成熟
- ✅ 类型安全，编译时检查
- ✅ 支持链式调用和复杂匹配

**生成方式**:
```bash
# 方式 1: 手动生成（推荐）
cd services/user-service
mockgen -source=model/usermodel_gen.go -destination=model/mocks/usermodel_mock.go -package=mocks

# 方式 2: 接口模式
mockgen -destination=model/mocks/usermodel_mock.go -package=mocks \
  github.com/guxiao1976/community-user/model UserModel

# 方式 3: 批量生成脚本
bash tools/generate-mocks.sh user-service
```

---

### 2.2 测试隔离原则

**隔离外部依赖**:
```
✅ Mock: Model 层（数据库）
✅ Mock: gRPC Client（跨服务调用）
✅ Mock: Redis Client（缓存）
❌ 不 Mock: 纯函数、业务逻辑

原则：只 Mock I/O 边界，不 Mock 业务逻辑
```

**ServiceContext 注入**:
```go
// 生产代码
svcCtx := &svc.ServiceContext{
    Config:    c,
    UserModel: model.NewUserModel(sqlConn),
    AuthRpc:   authx.MustNewClient(c.AuthRpc),
}

// 测试代码
svcCtx := &svc.ServiceContext{
    UserModel: mockUserModel,  // Mock
    AuthRpc:   mockAuthRpc,    // Mock
}
```

---

### 2.3 Table-Driven Tests 模式

**标准结构**:
```go
tests := []struct {
    name      string          // 测试场景名称
    input     *RequestType    // 输入参数
    mockSetup func(*MockType) // Mock 设置
    want      *ResponseType   // 期望输出
    wantErr   bool            // 是否期望错误
}{
    // 正常路径
    {name: "success - 正常场景", ...},
    
    // 边界条件
    {name: "edge - 空值", input: &Req{Name: ""}, wantErr: true},
    {name: "edge - 零值", input: &Req{Id: 0}, wantErr: true},
    
    // 错误路径
    {name: "error - 数据库错误", ...},
    {name: "error - 业务规则违反", ...},
}
```

---

## 三、执行步骤

### Step 1: 安装工具（10分钟）

```bash
# 1. 安装 gomock
go install github.com/golang/mock/mockgen@latest

# 2. 安装 testify（断言库）
cd services/user-service
go get github.com/stretchr/testify/assert

# 3. 验证安装
which mockgen
```

---

### Step 2: 生成 Mock（每服务 5分钟）

```bash
# user-service
cd services/user-service
mkdir -p model/mocks
mockgen -source=model/usermodel_gen.go \
  -destination=model/mocks/usermodel_mock.go \
  -package=mocks

# auth-service
cd services/auth-service
mkdir -p model/mocks
mockgen -source=model/authtokenmodel_gen.go \
  -destination=model/mocks/authtokenmodel_mock.go \
  -package=mocks

# moderation-service
cd services/moderation-service
mkdir -p model/mocks
mockgen -source=model/mod_audit_log_gen.go \
  -destination=model/mocks/mod_audit_log_mock.go \
  -package=mocks
```

---

### Step 3: 编写测试（每个 logic 30分钟）

**优先级队列**:
1. user-service/api/internal/logic/user/ (2-3 个 logic)
2. auth-service/api/internal/logic/auth/ (2-3 个 logic)
3. moderation-service/api/internal/logic/text_review/ (2-3 个 logic)

**每个测试包含**:
- ✅ 至少 3 个测试用例（正常/边界/错误）
- ✅ Mock 设置完整
- ✅ 断言覆盖关键字段

---

### Step 4: 运行测试并验证覆盖率

```bash
# 运行测试
cd services/user-service
go test ./api/internal/logic/... -v

# 检查覆盖率
go test ./api/internal/logic/... -cover

# 生成覆盖率报告
go test ./api/internal/logic/... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

**目标**:
- ✅ API logic 层覆盖率 ≥ 30%
- ✅ 所有测试通过
- ✅ 无 data race

---

## 四、验收标准

### 定量指标

| 服务 | 测试文件数 | 测试函数数 | 覆盖率 | 状态 |
|------|:---:|:---:|:---:|:---:|
| user-service | ≥3 | ≥9 | ≥30% | 🔲 |
| auth-service | ≥3 | ≥9 | ≥30% | 🔲 |
| moderation-service | ≥3 | ≥9 | ≥30% | 🔲 |

### 定性标准

- ✅ 每个测试都是 table-driven
- ✅ Mock 设置清晰，可读性强
- ✅ 覆盖正常/边界/错误路径
- ✅ 测试命名遵循 `Test<Logic>_<Scenario>_<Expected>` 规范
- ✅ 所有测试独立，可并行运行

---

## 五、风险与应对

### 风险 1: Mock 生成失败

**原因**: Model 接口不清晰或依赖循环

**应对**:
- 手动定义 Mock 接口
- 重构 Model 层，明确接口边界

### 风险 2: 测试编写成本高

**原因**: ServiceContext 依赖复杂

**应对**:
- 创建测试工具函数 `NewMockServiceContext()`
- 只 Mock 必要的依赖，其他传 nil

### 风险 3: 覆盖率难以达标

**原因**: Handler 层代码过多

**应对**:
- 专注 Logic 层（业务逻辑）
- Handler 层通过集成测试覆盖

---

## 六、输出产物

### 代码产物

```
services/user-service/
├── model/mocks/
│   └── usermodel_mock.go
└── api/internal/logic/user/
    ├── user_logic_test.go
    ├── create_user_logic_test.go
    └── update_user_logic_test.go

services/auth-service/
├── model/mocks/
│   └── authtokenmodel_mock.go
└── api/internal/logic/auth/
    ├── login_logic_test.go
    ├── refresh_token_logic_test.go
    └── logout_logic_test.go

services/moderation-service/
├── model/mocks/
│   └── mod_audit_log_mock.go
└── api/internal/logic/text_review/
    ├── review_text_logic_test.go
    ├── review_image_logic_test.go
    └── query_result_logic_test.go
```

### 文档产物

```
.harness/changes/stage1-week1/
├── task-plan.md                # 本文件
├── execution-log.md            # 执行日志
├── test-coverage-report.md     # 覆盖率报告
└── lessons-learned.md          # 经验总结
```

---

## 七、下一步

Week 1 完成后进入 Week 2:

**Week 2: TDD 纪律强制执行**
- 修改 harness-pipeline.js 集成 TDD 证据验证
- 无 RED 证据 → QA FAIL
- 更新 QA 报告模板

---

**状态**: 📋 计划已制定，待执行  
**预计开始**: 立即  
**预计完成**: 7 天内
