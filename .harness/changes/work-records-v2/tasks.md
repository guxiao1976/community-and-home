# Tasks — test-pipeline-work-records

## 任务概览

| 服务 | 任务数 | 预估时间 |
|------|:-----:|:-------:|
| master-data-service (后端) | 17 | ~90 分钟 |
| web/pc (前端) | 8 | ~60 分钟 |
| **总计** | **25** | **~150 分钟** |

---

## 后端任务（master-data-service）

### Phase 1: 数据库层（2 tasks）

#### Task 1.1: 创建 Migration DDL

**描述**: 创建数据库迁移文件，包含 2 张表的 DDL

**文件**: `services/master-data-service/model/migration/pipeline_test_tables.sql`

**内容**:
```sql
CREATE TABLE `pipeline_test_work_items` (...);
CREATE TABLE `pipeline_test_work_records` (...);
```

**验证**: 
- SQL 语法正确
- 包含所有索引和约束
- 表注释完整

**预估**: 10 分钟

---

#### Task 1.2: 创建 GORM Model

**描述**: 创建 2 个 GORM 结构体

**文件**: `services/master-data-service/model/pipelinetestmodel/pipeline_test_work_item.go`  
**文件**: `services/master-data-service/model/pipelinetestmodel/pipeline_test_work_record.go`

**TDD 步骤**:
1. **RED**: 编写表名测试 `TestTableName()`
2. **GREEN**: 实现 `TableName()` 方法
3. **REFACTOR**: 添加字段标签和注释

**验证**:
- 所有 int64 ID 使用 `json:",string"` 标签
- `TableName()` 返回正确表名
- 测试覆盖率 100%

**预估**: 15 分钟

---

### Phase 2: Logic 层（10 tasks）

#### Task 2.1: 创建工作事项 Logic

**描述**: 实现 `CreateWorkItem(ctx, req) (*WorkItem, error)`

**文件**: `services/master-data-service/logic/pipelinetestlogic/create_work_item_logic.go`

**TDD 步骤**:
1. **RED**: 测试用例
   - 正常创建
   - 名称为空 → 返回错误
   - 名称重复 → 返回 "工作事项名称已存在"
   - 并发创建同名 → 捕获 MySQL 1062 → 返回友好提示
2. **GREEN**: 实现逻辑
   - 参数校验
   - 重复检查（SELECT WHERE name=? AND delete_time IS NULL）
   - 生成 Snowflake ID
   - INSERT，捕获唯一索引冲突
3. **REFACTOR**: 提取重复代码

**注入 Review 问题**: C-01（并发同名处理）

**验证**:
- 4 个测试用例全部 PASS
- 覆盖率 ≥ 80%

**预估**: 12 分钟

---

#### Task 2.2: 查询工作事项列表 Logic

**描述**: 实现 `GetWorkItemList(ctx, req) (*ListResp, error)`

**文件**: `services/master-data-service/logic/pipelinetestlogic/get_work_item_list_logic.go`

**TDD 步骤**:
1. **RED**: 测试用例
   - 查询全部（status 未指定）
   - 查询启用（status=1）
   - 查询禁用（status=0）
   - 分页（page=2, page_size=10）
2. **GREEN**: 实现逻辑
   - 构造查询条件（WHERE delete_time IS NULL AND status=?）
   - 排序（ORDER BY sort_order ASC）
   - 分页（LIMIT OFFSET）
3. **REFACTOR**: 优化查询

**验证**:
- 4 个测试用例全部 PASS
- 软删除记录不在列表中

**预估**: 10 分钟

---

#### Task 2.3: 查询单个工作事项 Logic

**描述**: 实现 `GetWorkItem(ctx, id) (*WorkItem, error)`

**TDD 步骤**:
1. **RED**: 测试用例
   - 正常查询
   - ID 不存在 → 返回 "工作事项不存在"
   - ID 已软删除 → 返回 "工作事项不存在"
2. **GREEN**: 实现逻辑（WHERE id=? AND delete_time IS NULL）
3. **REFACTOR**: 复用查询代码

**预估**: 8 分钟

---

#### Task 2.4: 更新工作事项 Logic

**描述**: 实现 `UpdateWorkItem(ctx, id, req) error`

**TDD 步骤**:
1. **RED**: 测试用例
   - 正常更新
   - ID 不存在 → 返回错误
   - 更新为已存在的名称 → 返回 "工作事项名称已存在"
2. **GREEN**: 实现逻辑
   - 检查存在性
   - 检查名称重复（排除自己）
   - UPDATE
3. **REFACTOR**: 优化重复检查

**预估**: 10 分钟

---

#### Task 2.5: 删除工作事项 Logic

**描述**: 实现 `DeleteWorkItem(ctx, id) error`

**TDD 步骤**:
1. **RED**: 测试用例
   - 正常删除
   - ID 不存在 → 返回错误
   - 存在关联记录 → 返回 "该工作事项存在关联记录，无法删除"
2. **GREEN**: 实现逻辑
   - 事务开始
   - 检查关联记录（COUNT work_records WHERE work_item_id=? AND delete_time IS NULL）
   - 软删除（UPDATE SET delete_time=NOW()）
   - 事务提交
3. **REFACTOR**: 提取事务模板

**验证**:
- 3 个测试用例全部 PASS
- 事务回滚测试

**预估**: 12 分钟

---

#### Task 2.6: 创建工作记录 Logic

**描述**: 实现 `CreateWorkRecord(ctx, req) (*WorkRecord, error)`

**TDD 步骤**:
1. **RED**: 测试用例
   - 正常创建
   - work_item_id 不存在 → 返回 "工作事项不存在或已禁用"
   - work_item_id 已禁用 → 返回 "工作事项不存在或已禁用"
   - work_item_id 已删除 → 返回 "工作事项不存在或已禁用"
   - 工作时长 ≤ 0 → 返回错误
   - 工作时长 > 24 → 返回错误
2. **GREEN**: 实现逻辑
   - 参数校验
   - 外键校验（SELECT id FROM work_items WHERE id=? AND status=1 AND delete_time IS NULL）
   - 生成 Snowflake ID
   - INSERT
3. **REFACTOR**: 提取外键校验函数

**注入 Review 问题**: C-02（空列表前端处理，后端需确保校验逻辑）

**验证**:
- 6 个测试用例全部 PASS

**预估**: 15 分钟

---

#### Task 2.7: 查询工作记录列表 Logic

**描述**: 实现 `GetWorkRecordList(ctx, req) (*ListResp, error)`

**TDD 步骤**:
1. **RED**: 测试用例
   - 查询全部（无日期筛选）
   - 日期范围查询（start_date ~ end_date）
   - 按 work_item_id 筛选
   - 分页
   - start_date > end_date → 返回空列表（不报错）
2. **GREEN**: 实现逻辑
   - JOIN 查询获取 work_item_name
   - 构造 WHERE 条件
   - 排序（ORDER BY work_date DESC, created_time DESC）
   - 分页
3. **REFACTOR**: 优化 JOIN 查询

**注入 Review 问题**: C-03（日期范围边界）

**验证**:
- 5 个测试用例全部 PASS
- 返回结果包含 work_item_name

**预估**: 12 分钟

---

#### Task 2.8: 查询单个工作记录 Logic

**描述**: 实现 `GetWorkRecord(ctx, id) (*WorkRecord, error)`

**TDD 步骤**:
1. **RED**: 测试用例
   - 正常查询（包含 work_item_name）
   - ID 不存在 → 返回错误
2. **GREEN**: 实现逻辑（JOIN 查询）
3. **REFACTOR**: 复用 JOIN 逻辑

**预估**: 8 分钟

---

#### Task 2.9: 更新工作记录 Logic

**描述**: 实现 `UpdateWorkRecord(ctx, id, req) error`

**TDD 步骤**:
1. **RED**: 测试用例
   - 正常更新
   - ID 不存在 → 返回错误
   - 更新 work_item_id 为不存在 → 返回 "工作事项不存在或已禁用"
2. **GREEN**: 实现逻辑
   - 检查存在性
   - 外键校验
   - UPDATE
3. **REFACTOR**: 复用外键校验

**预估**: 10 分钟

---

#### Task 2.10: 删除工作记录 Logic

**描述**: 实现 `DeleteWorkRecord(ctx, id) error`

**TDD 步骤**:
1. **RED**: 测试用例
   - 正常删除
   - ID 不存在 → 返回错误
2. **GREEN**: 实现逻辑（软删除）
3. **REFACTOR**: 提取软删除模板

**预估**: 8 分钟

---

### Phase 3: Handler 层（2 tasks）

#### Task 3.1: 工作事项 Handler 组

**描述**: 实现 5 个 HTTP 端点

**文件**: `services/master-data-service/api/internal/handler/pipelinetest/work_item_handler.go`

**端点**:
1. POST `/api/masterdata/pipeline-test/work-items` → CreateWorkItemHandler
2. GET `/api/masterdata/pipeline-test/work-items` → GetWorkItemListHandler
3. GET `/api/masterdata/pipeline-test/work-items/:id` → GetWorkItemHandler
4. PUT `/api/masterdata/pipeline-test/work-items/:id` → UpdateWorkItemHandler
5. DELETE `/api/masterdata/pipeline-test/work-items/:id` → DeleteWorkItemHandler

**实现步骤**:
1. 创建 Request/Response 结构体
2. 实现 Handler 函数（参数绑定 → 调用 Logic → 返回响应）
3. 注册路由到 `api/internal/handler/routes.go`

**验证**:
- go build 编译通过
- 路由注册正确

**预估**: 15 分钟

---

#### Task 3.2: 工作记录 Handler 组

**描述**: 实现 5 个 HTTP 端点

**文件**: `services/master-data-service/api/internal/handler/pipelinetest/work_record_handler.go`

**端点**:
1. POST `/api/masterdata/pipeline-test/work-records` → CreateWorkRecordHandler
2. GET `/api/masterdata/pipeline-test/work-records` → GetWorkRecordListHandler
3. GET `/api/masterdata/pipeline-test/work-records/:id` → GetWorkRecordHandler
4. PUT `/api/masterdata/pipeline-test/work-records/:id` → UpdateWorkRecordHandler
5. DELETE `/api/masterdata/pipeline-test/work-records/:id` → DeleteWorkRecordHandler

**实现步骤**:
1. 创建 Request/Response 结构体
2. 实现 Handler 函数
3. 注册路由

**验证**:
- go build 编译通过
- 路由注册正确

**预估**: 15 分钟

---

### Phase 4: 后端测试（3 tasks）

#### Task 4.1: Logic 层单元测试补充

**描述**: 补充边界测试用例（Review 发现的场景）

**测试场景**:
- C-01: 并发创建同名工作事项（多协程测试）
- C-03: 日期范围倒序（start_date > end_date）
- C-04: 工作时长精度超限（超过 2 位小数）
- C-07: 分页越界（page > 总页数）

**验证**:
- 所有边界测试 PASS
- Logic 层覆盖率 ≥ 80%

**预估**: 12 分钟

---

#### Task 4.2: API 集成测试

**描述**: 端到端测试 10 个 API 端点

**文件**: `services/master-data-service/api/test/pipeline_test_integration_test.go`

**测试流程**:
1. 启动测试服务器
2. 创建工作事项 → 查询列表 → 更新 → 删除
3. 创建工作记录（JOIN 验证） → 查询列表 → 更新 → 删除
4. 软删除验证（删除后列表不显示）

**验证**:
- 所有 API 返回正确状态码和响应格式
- 单层包装（data 字段）

**预估**: 15 分钟

---

#### Task 4.3: API 冒烟测试

**描述**: 使用 curl/Postman 手动测试

**测试点**:
- JWT 认证（无 token → 401）
- 参数校验（空值 → 400）
- 业务规则（名称重复 → 400）
- 关联记录删除保护

**验证**:
- 所有端点可访问
- 错误码符合规范

**预估**: 10 分钟

---

## 前端任务（web/pc）

### Phase 5: API 层（1 task）

#### Task 5.1: API 函数 + TypeScript 类型

**描述**: 创建 API 函数和类型定义

**文件**: `web/pc/src/api/pipeline-test.ts`

**内容**:
1. 类型定义（WorkItem, WorkRecord, Query 接口）
2. API 函数（10 个，对应后端端点）
3. 响应解包（data 字段提取）

**验证**:
- TypeScript 编译通过
- 所有 ID 字段类型为 string

**预估**: 15 分钟

---

### Phase 6: 路由配置（1 task）

#### Task 6.1: 路由 + 菜单注册

**描述**: 配置一级菜单和 2 个子路由

**文件**: `web/pc/src/router/index.ts`

**路由结构**:
```
/pipeline-test
  ├─ /pipeline-test/work-items (工作事项维护)
  └─ /pipeline-test/work-records (工作内容记录)
```

**验证**:
- 菜单显示在侧边栏
- 路由跳转正常
- 面包屑正确

**预估**: 10 分钟

---

### Phase 7: 页面开发（4 tasks）

#### Task 7.1: 工作事项维护页面

**描述**: 实现工作事项列表页面

**文件**: `web/pc/src/views/pipeline-test/WorkItems.vue`

**功能**:
1. 列表显示（名称、排序号、状态、操作）
2. 新增按钮 → 打开对话框
3. 编辑按钮 → 打开对话框（回显数据）
4. 删除按钮 → 确认对话框
5. 状态筛选（全部/启用/禁用）
6. 分页

**验证**:
- CRUD 流程完整
- 状态切换实时生效
- 删除有关联记录时显示错误提示

**预估**: 20 分钟

---

#### Task 7.2: 工作事项表单对话框

**描述**: 实现新增/编辑对话框

**文件**: `web/pc/src/views/pipeline-test/components/WorkItemForm.vue`

**功能**:
1. 表单字段（名称、排序号、状态）
2. 校验规则（名称必填最大 100、排序号整数 0-9999）
3. 提交（新增/编辑模式切换）
4. 取消（重置表单）

**验证**:
- 编辑模式回显所有字段（8 层检查）
- 名称重复显示错误提示

**预估**: 15 分钟

---

#### Task 7.3: 工作内容记录页面

**描述**: 实现工作记录列表页面

**文件**: `web/pc/src/views/pipeline-test/WorkRecords.vue`

**功能**:
1. 列表显示（工作事项、日期、描述、时长、操作）
2. 日期范围查询（快捷选项：今天/近7天/近30天）
3. 新增按钮 → 打开对话框
4. 编辑按钮 → 打开对话框（回显数据）
5. 删除按钮 → 确认对话框
6. 工作时长统计（列表底部合计行）
7. 分页

**验证**:
- 日期查询实时生效
- 统计行显示正确
- 默认查询最近 30 天

**预估**: 25 分钟

---

#### Task 7.4: 工作记录表单对话框

**描述**: 实现新增/编辑对话框

**文件**: `web/pc/src/views/pipeline-test/components/WorkRecordForm.vue`

**功能**:
1. 表单字段（工作事项下拉、日期、描述、时长）
2. 工作事项下拉列表加载（status=1）
3. 空列表检查（C-02）→ 显示提示 + 禁用提交
4. 校验规则（必填、时长 0.1-24、描述最大 2000）
5. 提交（新增/编辑模式）
6. 取消（重置表单）

**注入 Review 问题**: C-02（空列表异常处理）

**验证**:
- 编辑模式回显所有字段
- 工作事项下拉列表正确加载
- 空列表提示显示

**预估**: 20 分钟

---

### Phase 8: 前端测试（2 tasks）

#### Task 8.1: 组件单元测试

**描述**: 测试表单组件

**文件**: `web/pc/src/views/pipeline-test/components/__tests__/WorkItemForm.spec.ts`  
**文件**: `web/pc/src/views/pipeline-test/components/__tests__/WorkRecordForm.spec.ts`

**测试场景**:
- 表单初始化
- 编辑模式回显
- 校验规则触发
- 提交事件触发

**验证**:
- 测试覆盖率 ≥ 70%

**预估**: 15 分钟

---

#### Task 8.2: E2E 测试

**描述**: 端到端流程测试

**文件**: `web/pc/cypress/e2e/pipeline-test.cy.ts`

**测试流程**:
1. 访问工作事项页面 → 新增 → 列表显示
2. 编辑工作事项 → 状态切换
3. 访问工作记录页面 → 新增 → 日期查询
4. 删除工作记录 → 列表刷新
5. 删除有关联的工作事项 → 显示错误

**验证**:
- E2E 测试全部 PASS

**预估**: 20 分钟

---

## 验证任务（2 tasks）

#### Task 9.1: harness-checks 检查

**描述**: 运行 15 项机械化检查

**命令**:
```bash
bash .harness/skills/qa/scripts/harness-checks.sh --service master-data-service
```

**检查项**:
1. go build 编译
2. go test 测试
3. go vet 静态分析
4. gofmt 格式
5. Snowflake ID 序列化
6. API 响应格式
7. 数据库 Migration
8. ...（共 15 项）

**验证**:
- 所有检查 PASS（15/15）

**预估**: 10 分钟

---

#### Task 9.2: 编辑回显完整性检查

**描述**: 验证编辑表单数据完整性（8 层检查）

**检查点**:
1. API 返回字段完整
2. 前端类型定义匹配
3. 表单回显所有字段
4. ID 字段字符串序列化
5. 日期格式正确
6. 状态值正确
7. 可选字段处理
8. 提交时数据未丢失

**验证**:
- 手动测试 + 截图验证

**预估**: 10 分钟

---

## 任务依赖关系

```
Phase 1 (数据库)
  ↓
Phase 2 (Logic) — 并行执行 10 个任务
  ↓
Phase 3 (Handler) — 依赖 Logic
  ↓
Phase 4 (后端测试) — 依赖 Handler
  ‖
Phase 5 (前端 API) — 可与后端并行
  ↓
Phase 6 (前端路由)
  ↓
Phase 7 (前端页面) — 并行执行 4 个任务
  ↓
Phase 8 (前端测试)
  ↓
Phase 9 (验证) — 全部完成后执行
```

---

## 流水线记录要求

每个任务完成后记录：
- 实际耗时
- 遇到的问题（编译错误/测试失败/规范违反）
- 修复方法
- 工具使用情况（harness-checks 发现的问题）

记录位置：`.harness/changes/test-pipeline-work-records/implementation-log.md`

---

## 关联文档

| 文档 | 路径 |
|------|------|
| Design | `.harness/changes/test-pipeline-work-records/design.md` |
| Backend Spec | `.harness/changes/test-pipeline-work-records/specs/backend/spec.md` |
| Frontend Spec | `.harness/changes/test-pipeline-work-records/specs/frontend/spec.md` |
| 编码规范 | `.harness/rules/项目编码规范.md` |
| 工程结构 | `.harness/rules/工程结构.md` |
