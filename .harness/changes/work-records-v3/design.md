# Design Document — test-pipeline-work-records

## 概述

为测试流水线验证提供工作记录管理功能。在 PC 管理后台新增一级菜单"AI-Coding自我提升"，包含工作事项维护和工作内容记录两个页面。后端归属 master-data-service，使用 `pipeline_test_` 表前缀隔离。

## 服务归属

**master-data-service**

**理由**：
1. 主数据服务定位是"基础数据管理中心"
2. 工作事项和工作记录属于元数据性质
3. 该服务已有成熟的 CRUD 模式和审批流程可参考
4. 避免为测试模块创建独立微服务
5. 表前缀 `pipeline_test_` 隔离，易于清理

## 架构决策

### 1. 不新增 RPC 接口
- **决策**: 仅提供 REST API，不新增 gRPC 接口
- **理由**: 仅用于管理后台展示，无其他服务消费需求
- **影响**: 无 Proto 变更，不涉及全局组依赖

### 2. 不使用审批流程
- **决策**: 不接入 master-data-service 现有的审批工作流
- **理由**: 测试模块数据，无需严格审批
- **影响**: 简化开发流程，聚焦流水线验证

### 3. 工作时长使用小数（小时）
- **决策**: `duration_hours DECIMAL(5,2)`，范围 0.01~24.00
- **理由**: 
  - 便于统计（直接求和）
  - 支持灵活粒度（0.5h = 30 分钟）
  - 避免分钟转换复杂度

### 4. 并发冲突处理（注入 Review C-01）
- **决策**: 数据库 UNIQUE INDEX `uk_name` (name, delete_time)
- **理由**: 防止并发创建同名工作事项
- **实现**: 
  - 软删除后 `delete_time != NULL`，允许重名（已删除）
  - 未删除时 `delete_time IS NULL`，触发唯一约束冲突
  - 应用层捕获 MySQL 1062 错误，返回 "工作事项名称已存在"

### 5. 空列表异常处理（注入 Review C-02）
- **决策**: 前端新增/编辑工作记录时检查工作事项下拉列表
- **理由**: 所有工作事项禁用/删除时，用户无法选择
- **实现**:
  - 前端打开对话框时调用 `getWorkItems({status: 1})`
  - 若返回空列表，显示提示"请先创建并启用工作事项"
  - 禁用表单提交按钮

---

## 数据模型

### 表 1: `pipeline_test_work_items` — 工作事项

```sql
CREATE TABLE `pipeline_test_work_items` (
  `id` BIGINT PRIMARY KEY COMMENT 'Snowflake ID',
  `name` VARCHAR(100) NOT NULL COMMENT '工作事项名称',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序号（升序）',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态 0=禁用 1=启用',
  `created_by` BIGINT NOT NULL COMMENT '创建人 ID（从 JWT 提取）',
  `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `delete_time` DATETIME NULL DEFAULT NULL COMMENT '软删除时间',
  
  INDEX `idx_status` (`status`, `delete_time`),
  INDEX `idx_sort_order` (`sort_order`),
  UNIQUE INDEX `uk_name` (`name`, `delete_time`) COMMENT '名称唯一（软删除后可重用）'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='测试流水线-工作事项';
```

**字段说明**：
- `id`: Snowflake ID，JSON 序列化为字符串（`json:",string"`）
- `name`: 工作事项名称，前端最大 100 字符，后端校验非空
- `sort_order`: 排序号，查询列表时 ORDER BY sort_order ASC
- `status`: 0=禁用（不在下拉列表显示），1=启用
- `delete_time`: NULL=未删除，NOT NULL=已软删除

**索引设计**：
- `uk_name`: 唯一约束，防止并发创建同名（C-01）
  - `delete_time IS NULL` 时触发约束
  - `delete_time IS NOT NULL` 时允许重名（已删除）
- `idx_status`: 支持按状态筛选 + 排除软删除
- `idx_sort_order`: 支持排序查询

---

### 表 2: `pipeline_test_work_records` — 工作记录

```sql
CREATE TABLE `pipeline_test_work_records` (
  `id` BIGINT PRIMARY KEY COMMENT 'Snowflake ID',
  `work_item_id` BIGINT NOT NULL COMMENT '关联工作事项 ID',
  `work_date` DATE NOT NULL COMMENT '工作日期',
  `description` TEXT NOT NULL COMMENT '工作内容描述（最大 2000 字符）',
  `duration_hours` DECIMAL(5,2) NOT NULL COMMENT '工作时长（小时，0.01-24.00）',
  `created_by` BIGINT NOT NULL COMMENT '创建人 ID',
  `created_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `delete_time` DATETIME NULL DEFAULT NULL,
  
  INDEX `idx_work_date` (`work_date`, `delete_time`),
  INDEX `idx_work_item` (`work_item_id`, `delete_time`),
  CONSTRAINT `chk_duration` CHECK (`duration_hours` > 0 AND `duration_hours` <= 24)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='测试流水线-工作记录';
```

**字段说明**：
- `work_item_id`: 关联工作事项，应用层校验（不使用外键）
- `work_date`: 工作日期，DATE 类型，前端传 YYYY-MM-DD
- `description`: 工作内容，TEXT 类型，前端限制 2000 字符
- `duration_hours`: DECIMAL(5,2)，范围 0.01-24.00，前端步进 0.5

**索引设计**：
- `idx_work_date`: 支持日期范围查询（主要查询条件）
- `idx_work_item`: 支持按工作事项筛选

**约束**：
- CHECK 约束确保工作时长合法（> 0 且 <= 24）

---

## 业务流程

### 流程 1: 创建工作事项（含并发冲突处理）

```
用户填写表单 → 前端校验（名称非空、长度≤100） → POST /api/masterdata/pipeline-test/work-items
                                                          ↓
                                        Handler 提取 JWT user_id → Logic 层
                                                          ↓
                              检查名称重复（SELECT name WHERE name=? AND delete_time IS NULL）
                                                          ↓
                                                    [已存在?]
                                                 ├── 是 → 返回 400 "工作事项名称已存在"
                                                 └── 否 → 生成 Snowflake ID → INSERT
                                                          ↓
                                                    [唯一索引冲突?]（并发场景）
                                                 ├── 是（MySQL 1062）→ 返回 400 "工作事项名称已存在"
                                                 └── 否 → 返回 200 + 新记录
```

**关键点**：
- 应用层先查询，避免大部分重复
- 数据库 UNIQUE INDEX 保证并发安全（C-01）
- Logic 层捕获 MySQL 1062 错误码，转换为友好提示

---

### 流程 2: 创建工作记录（含外键校验 + 空列表检查）

```
用户打开新增对话框 → 加载工作事项下拉列表（GET /work-items?status=1）
                                                          ↓
                                                    [列表为空?]（C-02）
                                                 ├── 是 → 显示提示"请先创建并启用工作事项"，禁用提交
                                                 └── 否 → 渲染下拉列表
                                                          ↓
用户填写表单 → 前端校验 → POST /api/masterdata/pipeline-test/work-records
                                                          ↓
                              Logic 层校验 work_item_id 存在性（SELECT id FROM work_items 
                                                          WHERE id=? AND status=1 AND delete_time IS NULL）
                                                          ↓
                                                    [不存在?]
                                                 ├── 是 → 返回 400 "工作事项不存在或已禁用"
                                                 └── 否 → INSERT → 返回 200
```

**关键点**：
- 前端预检查避免无效表单提交（C-02）
- 后端应用层校验确保引用完整性
- 不使用数据库外键（避免软删除冲突）

---

### 流程 3: 删除工作事项（关联记录检查）

```
用户点击删除 → 前端确认对话框 → DELETE /api/masterdata/pipeline-test/work-items/:id
                                                          ↓
                              Logic 层事务开始 → 检查关联记录（SELECT COUNT(*) FROM work_records
                                                          WHERE work_item_id=? AND delete_time IS NULL）
                                                          ↓
                                                    [count > 0?]
                                                 ├── 是 → ROLLBACK → 返回 400 "该工作事项存在关联记录，无法删除"
                                                 └── 否 → 软删除（UPDATE SET delete_time=NOW()）→ COMMIT → 返回 200
```

**关键点**：
- 事务保证检查和删除的原子性
- 软删除保留数据用于历史追溯

---

### 流程 4: 查询工作记录列表（日期范围）

```
用户选择日期范围 → 点击查询 → GET /api/masterdata/pipeline-test/work-records?start_date=X&end_date=Y
                                                          ↓
                              Logic 层构造查询（SELECT r.*, i.name as work_item_name
                                                          FROM work_records r
                                                          LEFT JOIN work_items i ON r.work_item_id = i.id
                                                          WHERE r.work_date BETWEEN ? AND ?
                                                          AND r.delete_time IS NULL
                                                          ORDER BY r.work_date DESC, r.created_time DESC）
                                                          ↓
                                                    返回 {list: [...], total: N}
```

**关键点**：
- JOIN 查询获取 `work_item_name` 冗余字段（避免前端二次请求）
- 日期范围索引优化查询性能
- 前端默认查询最近 30 天（`onMounted` 时设置参数）

---

## 接口设计

### REST API

所有端点前缀 `/api/masterdata/pipeline-test/`

#### 工作事项（5 个端点）

| Method | Path | 说明 |
|--------|------|------|
| POST | `/work-items` | 创建工作事项 |
| GET | `/work-items` | 查询工作事项列表（支持 status/include_disabled 筛选） |
| GET | `/work-items/:id` | 查询单个工作事项 |
| PUT | `/work-items/:id` | 更新工作事项 |
| DELETE | `/work-items/:id` | 删除工作事项（软删除） |

#### 工作记录（5 个端点）

| Method | Path | 说明 |
|--------|------|------|
| POST | `/work-records` | 创建工作记录 |
| GET | `/work-records` | 查询工作记录列表（支持 start_date/end_date/work_item_id 筛选） |
| GET | `/work-records/:id` | 查询单个工作记录 |
| PUT | `/work-records/:id` | 更新工作记录 |
| DELETE | `/work-records/:id` | 删除工作记录（软删除） |

### Go 数据结构

```go
// Model 层
type PipelineTestWorkItem struct {
    Id          int64      `gorm:"primaryKey;column:id" json:"id,string"`
    Name        string     `gorm:"column:name" json:"name"`
    SortOrder   int        `gorm:"column:sort_order" json:"sort_order"`
    Status      int        `gorm:"column:status" json:"status"`
    CreatedBy   int64      `gorm:"column:created_by" json:"created_by,string"`
    CreatedTime time.Time  `gorm:"column:created_time" json:"created_time"`
    UpdatedTime time.Time  `gorm:"column:updated_time" json:"updated_time"`
    DeleteTime  *time.Time `gorm:"column:delete_time" json:"delete_time,omitempty"`
}

func (PipelineTestWorkItem) TableName() string {
    return "pipeline_test_work_items"
}

type PipelineTestWorkRecord struct {
    Id            int64      `gorm:"primaryKey;column:id" json:"id,string"`
    WorkItemId    int64      `gorm:"column:work_item_id" json:"work_item_id,string"`
    WorkItemName  string     `gorm:"-" json:"work_item_name"`  // JOIN 查询填充
    WorkDate      string     `gorm:"column:work_date" json:"work_date"`  // DATE → string
    Description   string     `gorm:"column:description" json:"description"`
    DurationHours float64    `gorm:"column:duration_hours" json:"duration_hours"`
    CreatedBy     int64      `gorm:"column:created_by" json:"created_by,string"`
    CreatedTime   time.Time  `gorm:"column:created_time" json:"created_time"`
    UpdatedTime   time.Time  `gorm:"column:updated_time" json:"updated_time"`
    DeleteTime    *time.Time `gorm:"column:delete_time" json:"delete_time,omitempty"`
}

func (PipelineTestWorkRecord) TableName() string {
    return "pipeline_test_work_records"
}
```

**关键点**：
- 所有 int64 ID 使用 `json:",string"` 标签（[[proto-jstype]]）
- `WorkItemName` 字段不映射数据库列（`gorm:"-"`），由 JOIN 查询填充
- `WorkDate` 使用 string 类型简化前后端对接（YYYY-MM-DD）

---

## 前端设计

### 路由配置

```typescript
// @/config/route.config.ts
{
  path: '/pipeline-test',
  name: 'PipelineTest',
  meta: {
    title: 'AI-Coding自我提升',
    icon: 'el-icon-document',
    requiresAuth: true
  },
  children: [
    {
      path: 'work-items',
      name: 'WorkItems',
      component: () => import('@/views/pipeline-test/WorkItems.vue'),
      meta: { title: '工作事项维护' }
    },
    {
      path: 'work-records',
      name: 'WorkRecords',
      component: () => import('@/views/pipeline-test/WorkRecords.vue'),
      meta: { title: '工作内容记录' }
    }
  ]
}
```

### TypeScript 类型

```typescript
// @/api/pipeline-test.ts
export interface WorkItem {
  id?: string;
  name: string;
  sort_order: number;
  status: number;
  created_by?: string;
  created_time?: string;
  updated_time?: string;
  delete_time?: string;
}

export interface WorkRecord {
  id?: string;
  work_item_id: string;
  work_item_name?: string;
  work_date: string;
  description: string;
  duration_hours: number;
  created_by?: string;
  created_time?: string;
  updated_time?: string;
  delete_time?: string;
}

export interface WorkItemQuery {
  page?: number;
  page_size?: number;
  status?: number;
  include_disabled?: boolean;
}

export interface WorkRecordQuery {
  page?: number;
  page_size?: number;
  start_date?: string;
  end_date?: string;
  work_item_id?: string;
}
```

### 组件文件结构

```
web/pc/src/
├── api/
│   └── pipeline-test.ts          // API 函数 + TypeScript 类型
├── views/
│   └── pipeline-test/
│       ├── WorkItems.vue         // 工作事项维护页面
│       ├── WorkRecords.vue       // 工作内容记录页面
│       ├── components/
│       │   ├── WorkItemForm.vue  // 工作事项表单对话框
│       │   └── WorkRecordForm.vue // 工作记录表单对话框
```

---

## 安全考虑

### 1. 认证与授权
- **认证**: 所有 API 经过 JWT 中间件验证
- **授权**: 测试模块不做细粒度权限控制，管理员可访问
- **用户追踪**: `created_by` 从 JWT 提取，记录操作人

### 2. 输入验证

**前端验证**：
- 工作事项名称：必填、最大 100 字符
- 排序号：必填、整数、范围 0-9999
- 工作时长：必填、数字、范围 0.1-24、最多 2 位小数
- 工作内容：必填、最大 2000 字符

**后端验证**：
- 参数绑定失败 → 400 错误
- 业务规则校验（名称重复、外键不存在）→ 400 错误
- 数据库约束冲突 → 捕获并转换为友好提示

### 3. SQL 注入防护
- 使用 GORM 参数化查询，不拼接 SQL

### 4. XSS 防护
- 后端返回纯数据，前端 Vue 自动转义
- 工作内容描述显示时使用 `{{ }}` 插值（自动转义）

---

## 性能考虑

### 1. 数据库优化
- 索引覆盖主要查询条件（status、work_date、work_item_id）
- 分页查询避免全表扫描
- JOIN 查询使用索引关联（work_item_id）

### 2. 缓存策略
- **不使用 Redis 缓存**（测试数据，低访问量）
- 依赖数据库索引性能

### 3. 前端优化
- 工作事项下拉列表缓存（对话框打开时加载一次）
- 表格虚拟滚动（若数据量 > 1000，使用 el-table 虚拟滚动）
- 防抖搜索输入（300ms）

---

## 测试策略

### 1. 单元测试
- **后端 Logic 层**:
  - 创建工作事项（正常、名称重复、并发冲突）
  - 删除工作事项（正常、存在关联记录）
  - 外键校验（work_item_id 不存在、已禁用）
  - 日期范围查询（正常、start_date > end_date）
- **前端 API 函数**: Mock axios，验证请求参数和响应处理

### 2. 集成测试
- **API 端到端**:
  - CRUD 完整流程（创建 → 查询 → 更新 → 删除）
  - 软删除验证（删除后列表不显示、可恢复）
  - 并发创建同名工作事项（多线程测试）

### 3. 前端 E2E 测试
- **工作事项页面**:
  - 新增 → 列表显示 → 编辑回显 → 删除确认
  - 状态切换 → 列表筛选
- **工作记录页面**:
  - 新增（下拉列表加载）→ 日期范围查询 → 工作时长统计
  - 空列表提示（禁用所有工作事项后测试，C-02）

### 4. 边界测试（注入 Review 发现）
- 并发创建同名工作事项（C-01）
- 工作事项下拉列表为空（C-02）
- 日期范围倒序（C-03）
- 工作时长精度超限（C-04）
- 分页越界（C-07）

---

## 交付物

### 1. 后端
- [x] Migration 文件（2 张表 DDL）
- [x] Model 层（2 个 GORM struct）
- [x] Logic 层（工作事项 5 个方法 + 工作记录 5 个方法）
- [x] Handler 层（10 个 HTTP 端点）
- [x] API 文件配置（routes 注册）
- [x] 单元测试（Logic 层核心场景）

### 2. 前端
- [x] API 函数（10 个 + TypeScript 类型）
- [x] 路由配置（1 个一级菜单 + 2 个子路由）
- [x] 工作事项维护页面（列表 + 表单对话框）
- [x] 工作内容记录页面（列表 + 表单对话框 + 日期快捷选项 + 统计行）
- [x] E2E 测试（核心流程覆盖）

### 3. 验证
- [x] `harness-checks.sh --service master-data-service` 全部 PASS
- [x] `npm run build` TypeScript 编译通过
- [x] 编辑回显数据完整性检查（[[edit-form-data-integrity]] 8 层验证）

---

## 流水线评估记录点

开发过程中重点记录以下数据：

1. **时间分布**:
   - 需求分析时长
   - 架构设计时长
   - 后端开发时长（分 Migration/Model/Logic/Handler）
   - 前端开发时长（分 API/路由/页面）
   - 测试时长
   - 修复 Bug 时长

2. **问题类型**:
   - Review 发现的 MUST FIX 数量和修复成本
   - 编译/构建错误次数
   - 运行时错误类型（参数校验/业务逻辑/数据库）
   - 前后端联调问题（类型不匹配/字段缺失）

3. **工具使用**:
   - `harness-checks.sh` 发现的问题数量
   - Review Agent 覆盖率和准确性
   - TDD 流程执行情况

4. **待改进点**:
   - 流程阻塞点（等待时间长/反复修改）
   - 工具不足（缺少自动化检查项）
   - 文档不足（规范不清晰/示例缺失）

---

## 关联资源

| 资源 | 路径 |
|------|------|
| Proposal | `.harness/changes/test-pipeline-work-records/proposal.md` |
| Backend Spec | `.harness/changes/test-pipeline-work-records/specs/backend/spec.md` |
| Frontend Spec | `.harness/changes/test-pipeline-work-records/specs/frontend/spec.md` |
| Coverage Review | `.harness/changes/test-pipeline-work-records/review/spec_review_coverage_v1.md` |
| Clarity Review | `.harness/changes/test-pipeline-work-records/review/spec_review_clarity_v1.md` |
| Structure Review | `.harness/changes/test-pipeline-work-records/review/spec_review_structure_v1.md` |
| 编码规范 | `.harness/rules/项目编码规范.md` |
| 工程结构 | `.harness/rules/工程结构.md` |
| 记忆：API 响应包装 | `.harness/knowledge/memory/api-response-single-wrap.md` |
| 记忆：编辑回显完整性 | `.harness/knowledge/memory/edit-form-data-integrity.md` |
| 记忆：提交前检查 | `.harness/knowledge/memory/pre-commit-checks.md` |
