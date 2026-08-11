# P1 债务修复指南

**创建时间**: 2026-06-23 12:00  
**状态**: 📋 **待执行指南**

---

## 债务清单

### ❌ 未处理的 P1 债务

| # | 债务 | 工作量 | 风险 | 优先级 |
|---|------|--------|------|:------:|
| 3 | Outbox 非事务性写入 | 1-2h | 高 | P1 |
| 4 | Publisher 模块无测试 | 2-3h | 中 | P1 |
| 5 | 前端类型错误 (140) | 6-8h | 低 | P1 |

**总工作量**: 9-13 小时

---

## 债务 3: Outbox 非事务性写入修复指南

### 问题描述

**文件**: `services/master-data-service/api/internal/logic/approval/reviewItemLogic.go`  
**位置**: 第 96-113 行

**当前实现**（错误）:
```go
// 第 96 行：Update 实体
if err := l.svcCtx.MdResidentialAreaModel.Update(l.ctx, area); err != nil {
    return nil, errx.NewDefaultError("审核操作失败: " + err.Error())
}

// 第 113 行：写入 Outbox（事务外，返回值被忽略）
_ = publisher.WriteOutboxMessage(l.ctx, l.svcCtx.MdOutboxMessageModel, 
    "residential_area", area.Id, eventType, payload)
```

**问题**:
- `Update` 和 `WriteOutboxMessage` 不在同一事务中
- 如果 `Update` 成功但 `WriteOutboxMessage` 失败 → 数据不一致
- 实体状态已变更，但事件未发布

**风险**: 
- 审批状态已变更，但下游系统未收到通知
- 数据一致性问题

---

### 修复方案

#### 方案 A: 使用 sqlx.Trans（推荐）

**步骤 1**: 检查 ServiceContext 是否有 DB 连接

```go
// 查看 api/internal/svc/serviceContext.go
// 找到 sqlx.SqlConn 或原始 *sql.DB
```

**步骤 2**: 重构为事务

```go
// 修复后的实现
func (l *ReviewItemLogic) ReviewItem(req *types.ReviewItemReq) (resp *types.ReviewItemResp, err error) {
    // ... 前面的代码保持不变 ...
    
    // 开始事务
    err = l.svcCtx.DB.TransactCtx(l.ctx, func(ctx context.Context, session sqlx.Session) error {
        // 1. 更新实体（在事务内）
        if err := l.svcCtx.MdResidentialAreaModel.Update(ctx, session, area); err != nil {
            return err
        }
        
        // 2. 写入 Outbox（在同一事务内）
        eventType := "rejected"
        if action == "approve" {
            eventType = "approved"
        }
        payload := map[string]interface{}{
            "id":              area.Id,
            "name":            area.Name,
            "action":          action,
            "reviewer_id":     reviewerId,
            "review_time":     now.Format(time.RFC3339),
            "submission_type": area.SubmissionType.Int64,
        }
        
        if err := publisher.WriteOutboxMessage(ctx, session, l.svcCtx.MdOutboxMessageModel, 
            "residential_area", area.Id, eventType, payload); err != nil {
            return err
        }
        
        return nil
    })
    
    if err != nil {
        return nil, errx.NewDefaultError("审核操作失败: " + err.Error())
    }
    
    return &types.ReviewItemResp{Success: true}, nil
}
```

**步骤 3**: 修改 publisher.WriteOutboxMessage 签名

```go
// 修改前
func WriteOutboxMessage(ctx context.Context, model MdOutboxMessageModel, ...) error

// 修改后（支持事务）
func WriteOutboxMessage(ctx context.Context, session sqlx.Session, model MdOutboxMessageModel, ...) error
```

**步骤 4**: 修改 Model 方法支持 Session

```go
// 在 model/mdResidentialAreaModel_gen.go 中添加
func (m *defaultMdResidentialAreaModel) UpdateWithSession(ctx context.Context, session sqlx.Session, data *MdResidentialArea) error {
    // 使用 session 而不是 m.conn
    _, err := session.ExecCtx(ctx, mdResidentialAreaUpdateQuery, ...)
    return err
}
```

---

#### 方案 B: 简化方案（临时）

如果无法快速实现事务，可以先捕获错误：

```go
// 临时修复
if err := publisher.WriteOutboxMessage(l.ctx, l.svcCtx.MdOutboxMessageModel, 
    "residential_area", area.Id, eventType, payload); err != nil {
    // 记录错误，但不回滚 Update（因为已经提交）
    logx.Errorf("Failed to write outbox message after approval: %v", err)
    // TODO: 需要补偿机制
}
```

**注意**: 这不是正确的解决方案，只是减少静默失败。

---

### 测试要求

**测试用例**:
1. ✅ 正常审批流程（Update + Outbox 都成功）
2. ❌ Update 失败 → 事务回滚，Outbox 未写入
3. ❌ Outbox 写入失败 → 事务回滚，Update 未提交
4. ✅ 并发审批（多个审批员同时操作）

**验证点**:
- 数据库中实体状态
- outbox_messages 表中的记录
- 事务回滚后状态一致

---

### 预估工作量

| 任务 | 工作量 |
|------|--------|
| 理解 go-zero 事务机制 | 30 min |
| 重构 reviewItemLogic.go | 60 min |
| 修改 Model 支持 Session | 30 min |
| 编写测试用例 | 30 min |
| 验证所有场景 | 30 min |
| **总计** | **3 小时** |

---

## 债务 4: Publisher 模块测试补充指南

### 问题描述

**文件**: 
- `api/internal/publisher/outbox_publisher.go` (143 LOC)
- `api/internal/publisher/redis_publisher.go` (45 LOC)

**问题**: 无任何测试文件

**影响**: 
- 关键业务逻辑（MQ 发布、重试、状态管理）
- 测试覆盖率 0%

---

### 修复方案

**创建测试文件**:

#### 1. outbox_publisher_test.go

```go
package publisher

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

// Mock MQ Client
type MockMQClient struct {
    mock.Mock
}

func (m *MockMQClient) Publish(ctx context.Context, topic string, message []byte) error {
    args := m.Called(ctx, topic, message)
    return args.Error(0)
}

// Test Case 1: 成功发布
func TestOutboxPublisher_PublishSuccess(t *testing.T) {
    mockMQ := new(MockMQClient)
    mockMQ.On("Publish", mock.Anything, "test-topic", mock.Anything).Return(nil)
    
    publisher := NewOutboxPublisher(mockMQ)
    err := publisher.Publish(context.Background(), "test-topic", map[string]interface{}{
        "id": "123",
        "type": "test",
    })
    
    assert.NoError(t, err)
    mockMQ.AssertExpectations(t)
}

// Test Case 2: 发布失败
func TestOutboxPublisher_PublishFailure(t *testing.T) {
    mockMQ := new(MockMQClient)
    mockMQ.On("Publish", mock.Anything, "test-topic", mock.Anything).Return(errors.New("MQ error"))
    
    publisher := NewOutboxPublisher(mockMQ)
    err := publisher.Publish(context.Background(), "test-topic", map[string]interface{}{})
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "MQ error")
}

// Test Case 3: 重试机制
func TestOutboxPublisher_RetryMechanism(t *testing.T) {
    mockMQ := new(MockMQClient)
    // 第 1 次失败，第 2 次成功
    mockMQ.On("Publish", mock.Anything, "test-topic", mock.Anything).Return(errors.New("temporary error")).Once()
    mockMQ.On("Publish", mock.Anything, "test-topic", mock.Anything).Return(nil).Once()
    
    publisher := NewOutboxPublisher(mockMQ, WithRetry(2))
    err := publisher.Publish(context.Background(), "test-topic", map[string]interface{}{})
    
    assert.NoError(t, err)
    mockMQ.AssertNumberOfCalls(t, "Publish", 2)
}

// Test Case 4: 最大重试后仍失败
func TestOutboxPublisher_MaxRetryExceeded(t *testing.T) {
    mockMQ := new(MockMQClient)
    mockMQ.On("Publish", mock.Anything, "test-topic", mock.Anything).Return(errors.New("persistent error"))
    
    publisher := NewOutboxPublisher(mockMQ, WithRetry(3))
    err := publisher.Publish(context.Background(), "test-topic", map[string]interface{}{})
    
    assert.Error(t, err)
    mockMQ.AssertNumberOfCalls(t, "Publish", 3)
}

// Test Case 5: Outbox 写入
func TestOutboxPublisher_WriteOutbox(t *testing.T) {
    // Mock OutboxModel
    mockModel := new(MockOutboxModel)
    mockModel.On("Insert", mock.Anything, mock.Anything).Return(nil)
    
    err := WriteOutboxMessage(context.Background(), mockModel, "entity_type", "entity_id", "event_type", map[string]interface{}{})
    
    assert.NoError(t, err)
    mockModel.AssertExpectations(t)
}
```

#### 2. redis_publisher_test.go

```go
package publisher

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestRedisPublisher_Publish(t *testing.T) {
    // Mock Redis client
    // ...测试逻辑
}
```

---

### 测试覆盖目标

| 模块 | 当前覆盖率 | 目标覆盖率 |
|------|-----------|----------|
| outbox_publisher.go | 0% | ≥ 80% |
| redis_publisher.go | 0% | ≥ 80% |

---

### 预估工作量

| 任务 | 工作量 |
|------|--------|
| 创建 Mock 对象 | 30 min |
| 编写 5 个测试用例 | 90 min |
| 运行并调试 | 30 min |
| 达到 80% 覆盖率 | 30 min |
| **总计** | **3 小时** |

---

## 债务 5: 前端类型错误修复指南

### 问题描述

**错误数量**: 140 个  
**错误分布**:
- TS2345 (66个): 参数类型不匹配
- TS2322 (34个): 赋值类型不匹配
- TS2339 (15个): 属性不存在
- 其他 (25个)

**临时方案**（已实施）:
- 放宽 `tsconfig.app.json` 严格度
- `strict: false`, `skipLibCheck: true`

---

### 修复方案

#### 分批修复策略

**第 1 批**: Top 10 文件（4 小时）

| 文件 | 错误数 | 优先级 |
|------|--------|:------:|
| residential-areas/List.vue | 16 | 高 |
| division/Index.vue | 15 | 高 |
| grassroots/Index.vue | 14 | 高 |
| identity/admin-user/AdminUserList.vue | 10 | 高 |
| users/UserRoles.vue | 8 | 中 |
| aimodel/ModelList.vue | 8 | 中 |
| moderation/ModerationConfigTest.vue | 7 | 中 |
| masterdata-query/ResidentialQuery.vue | 6 | 中 |
| aimodel/Statistics.vue | 6 | 中 |
| residential-areas/Review.vue | 5 | 中 |

**第 2 批**: 剩余文件（4 小时）

---

#### 修复模式

**模式 1: DefaultRow 类型断言**

```typescript
// 错误
const handleEdit = (row: DefaultRow) => {
  editResidentialArea(row)  // TS2345: DefaultRow → ResidentialArea
}

// 修复
const handleEdit = (row: DefaultRow) => {
  editResidentialArea(row as ResidentialArea)
}
```

**模式 2: 类型守卫**

```typescript
// 更好的修复
function isResidentialArea(row: DefaultRow): row is ResidentialArea {
  return 'id' in row && 'name' in row && 'county_id' in row
}

const handleEdit = (row: DefaultRow) => {
  if (isResidentialArea(row)) {
    editResidentialArea(row)
  }
}
```

**模式 3: 泛型约束**

```typescript
// 最佳修复
interface TableProps<T> {
  data: T[]
  onEdit: (row: T) => void
}

// 使用
<Table<ResidentialArea>
  :data="areas"
  @edit="handleEdit"
/>
```

---

### 预估工作量

| 批次 | 文件数 | 工作量 |
|------|--------|--------|
| 第 1 批 (Top 10) | 10 | 4 小时 |
| 第 2 批 (剩余) | 19 | 3 小时 |
| 恢复 strict 模式 | - | 30 min |
| 回归测试 | - | 30 min |
| **总计** | 29 | **8 小时** |

---

## 总计

| 债务 | 工作量 | 风险 | 建议时间 |
|------|--------|------|---------|
| Outbox 事务性 | 3h | 高 | 本周 |
| Publisher 测试 | 3h | 中 | 本周 |
| 前端类型错误 | 8h | 低 | 分 2 周 |
| **总计** | **14h** | - | 2-3 周 |

---

## 执行建议

### 优先级排序

1. **第 1 周**: Outbox 事务性（P1，高风险）
2. **第 2 周**: Publisher 测试 + 前端第 1 批
3. **第 3 周**: 前端第 2 批 + 恢复 strict 模式

### 验证清单

- [ ] Outbox 事务性：所有测试通过，数据一致
- [ ] Publisher 测试：覆盖率 ≥ 80%
- [ ] 前端类型：构建成功，0 错误
- [ ] 回归测试：所有功能正常

---

**创建时间**: 2026-06-23 12:00  
**状态**: 📋 **详细指南已就绪**  
**建议**: 作为独立任务执行，每个债务 1-3 小时

**文档位置**: `.harness/changes/p1-debt-fix-guide.md`
