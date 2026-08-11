# Plan Review — test-pipeline-work-records（Clarity 视角）

**审查维度**: 清晰可执行

**审查时间**: 2026-06-22

**审查者**: Clarity Reviewer

## 摘要
- 🔴 MUST FIX: 0
- 🟡 SHOULD FIX: 3
- 🔵 INFO: 5

## 发现

### 🔴 MUST FIX

无 MUST FIX 问题。所有需求描述具备唯一解释，无阻塞性歧义。

---

### 🟡 SHOULD FIX

| # | 文件:行号/章节 | 问题 | 建议 |
|---|-------------|------|------|
| 1 | backend/spec.md:50 | 外键约束检查描述为"应用层校验"，但未明确校验时机（创建时？更新时？）和具体实现位置（Logic 层？Handler 层？） | 补充说明：校验应在 Logic 层的 CreateWorkRecord 和 UpdateWorkRecord 函数中执行，查询 work_items 表确认 work_item_id 存在且 status=1 且 delete_time IS NULL |
| 2 | backend/spec.md:250 | "默认返回最近 30 天"的逻辑未在 Scenario 中明确：是后端自动补充 start_date/end_date？还是返回全部然后前端筛选？ | 明确说明：后端在未收到 start_date/end_date 参数时，自动设置 start_date = today - 30 days, end_date = today，写入查询逻辑 |
| 3 | frontend/spec.md:265 | "默认查询最近 30 天"与后端 Scenario 250 呼应，但前端未说明页面加载时是否显式传递日期参数，还是依赖后端默认行为 | 建议前端明确传递参数：在 `onMounted` 时设置 `query.start_date = dayjs().subtract(30, 'day').format('YYYY-MM-DD')` 和 `query.end_date = dayjs().format('YYYY-MM-DD')`，保持前后端逻辑对齐 |

---

### 🔵 INFO

| # | 建议 |
|---|------|
| 1 | backend/spec.md §数据库表结构：建议在 migration 脚本中添加注释说明 `pipeline_test_` 前缀的用途和清理策略（如："测试模块数据，可随时删除 DROP TABLE IF EXISTS"） |
| 2 | backend/spec.md:196 | 返回响应中的 `work_item_name` 冗余字段设计合理，建议在 Data Model Summary 章节补充说明该字段通过 LEFT JOIN 查询获得，避免实施时误理解为数据库冗余存储 |
| 3 | frontend/spec.md:251 | "今天、昨天、本周、上周、本月、上月、最近 7 天、最近 30 天"共 8 个快捷选项，建议明确"本周"定义（周一至周日？周日至周六？），推荐使用 ISO 周（周一起始） |
| 4 | frontend/spec.md:338 | 工作内容描述截断逻辑（100 字符 + tooltip）合理，建议补充：如果描述包含换行符，列表中是否显示换行？建议替换 `\n` 为空格后再截断 |
| 5 | proposal.md:99 | 验收标准第 4 条引用 [[edit-form-data-integrity]] 记忆，frontend/spec.md §数据完整性验证已详细展开（8 层检查），追溯完整，建议在 backend/spec.md 也添加类似章节，明确后端响应字段完整性检查（确保 GET /id 返回所有表单需要的字段） |

---

## 详细分析

### 1. 需求描述清晰度

**✅ 优点**：
- 所有 Requirement 使用 SHALL/MUST 明确强制性，避免 SHOULD/MAY 引发的模糊性
- Scenario 采用 GIVEN-WHEN-THEN 结构，步骤清晰，期望输出明确
- JSON 示例完整，包含字段类型（如 `"id": "1234567890123456789"` 明确 Snowflake ID 序列化为字符串）
- 错误场景覆盖全面（空值、重复、超范围、关联冲突、不存在）

**🟡 待改进**：
- 外键约束检查（backend/spec.md:50）描述为"应用层校验"，但未明确实现细节，可能导致实施时遗漏 status=1 条件或忘记检查 delete_time
- 默认日期范围逻辑（backend:250, frontend:265）前后端描述不一致，可能导致实施时一方等待另一方补充逻辑

### 2. 实现步骤可行性

**✅ 优点**：
- 数据库表结构完整（索引、约束、软删除字段齐全）
- API Endpoint Summary 提供 10 个端点的清单，路径、方法、描述明确
- Data Model Summary 提供 Go struct 示例，包含 JSON 标签（`json:",string"` 用于 Snowflake ID）
- 前端组件结构清晰（建议文件树 + 路由配置示例）

**🟡 待改进**：
- 后端未明确哪些操作需要事务（spec 第 351-365 行提到事务，但 Scenario 未标注哪些 API 必须使用事务）
- 前端未明确日期快捷选项的具体实现（"本周"是 ISO 周还是自然周？）

### 3. 验收标准可验证性

**✅ 优点**：
- Proposal 验收标准分为功能验收、技术验收、流水线评估验收三类，覆盖全面
- 功能验收每条都可通过手动测试验证（如"工作事项可新增、编辑、删除、查询"）
- 技术验收引用记忆（[[api-response-single-wrap]], [[proto-jstype]], [[pre-commit-checks]]），标准明确
- Frontend spec §数据完整性验证提供 8 层检查清单，可操作性强

**🔵 建议增强**：
- 验收标准可补充自动化测试场景（如 API 单元测试用例数、前端 E2E 测试覆盖的核心路径）

### 4. 占位符和待定项检查

**✅ 无占位符**：
- 全文无 `[NEEDS CLARIFICATION]`、`TBD`、`TODO`、`待定` 等标记
- 所有字段类型、长度、范围、错误码均已明确
- 所有 API 路径、HTTP 方法、请求/响应结构均已定义

### 5. 粒度合理性

**✅ 粒度适中**：
- 每个 Requirement 聚焦单一职责（数据库、工作事项 API、工作记录 API、响应格式、ID 序列化、错误处理、事务、日志）
- 每个 Scenario 描述单一测试场景（正常流程、边界条件、异常情况），便于编写测试用例
- 前端组件拆分合理（WorkItems.vue / WorkRecords.vue 为页面容器，WorkItemForm.vue / WorkRecordForm.vue 为复用表单组件）

### 6. 前后端一致性检查

**✅ 一致性良好**：
- API 路径前后端对齐（`/api/masterdata/pipeline-test/work-items`）
- 数据模型字段名一致（backend Go struct 使用 snake_case，JSON 序列化后与 frontend TypeScript interface 的 snake_case 一致）
- Snowflake ID 序列化规则前后端一致（backend `json:",string"`，frontend TypeScript `id: string`）
- 错误码处理前后端一致（backend 定义 code/msg，frontend 拦截器提取 data 字段）

**🟡 待对齐**：
- 默认日期范围逻辑（见 SHOULD FIX #2, #3）

### 7. 依赖和前置条件

**✅ 依赖明确**：
- Backend spec 明确依赖 community-common/v2 errx 包（错误处理）
- Frontend spec 明确使用 Element Plus 组件库并列出具体组件
- 数据隔离策略明确（表前缀 `pipeline_test_`，API 路径 `/pipeline-test/`）
- 无跨服务依赖（独立模块，仅 master-data-service 内部实现）

### 8. 边界条件覆盖

**✅ 边界覆盖全面**：
- 字段长度边界（name 最大 100，description 最大 2000）
- 数值范围边界（sort_order 0-9999，duration_hours 0.1-24，最多 2 位小数）
- 空值边界（必填字段为空）
- 关联完整性边界（work_item_id 不存在、已禁用、已软删除）
- 并发边界（事务保证删除检查和软删除的原子性）

---

## Proposal ↔ Specs 追溯检查

| Proposal 需求 | Backend Spec | Frontend Spec | 一致性 |
|--------------|--------------|---------------|:------:|
| 2 张数据库表（pipeline_test_ 前缀） | REQ-数据库表结构设计 | - | ✅ |
| 工作事项 4 个 API（CRUD） | REQ-工作事项 CRUD API | API 函数实现 | ✅ |
| 工作记录 4 个 API（CRUD） | REQ-工作记录 CRUD API | API 函数实现 | ✅ |
| 日期范围筛选 | Scenario:242-255 | Scenario:251-261 | 🟡 默认范围逻辑需对齐 |
| 工作事项下拉列表（仅启用） | Scenario:119-122 | Scenario:267-271 | ✅ |
| 软删除 | 表结构 delete_time + Scenario:153-169 | - | ✅ |
| Snowflake ID → string | REQ-Snowflake ID 序列化 | TypeScript 类型 id: string | ✅ |
| API 响应单层包装 | REQ-API 响应格式 | axios 拦截器提取 data | ✅ |
| 编辑表单数据完整性 [[edit-form-data-integrity]] | - | REQ-数据完整性验证（8 层检查） | ✅ |

---

## 可执行性评估

### 后端实施路径清晰度：⭐⭐⭐⭐☆（4/5）

**清晰点**：
- 数据库 migration 脚本可直接编写（表结构、索引、约束已明确）
- Model 层 struct 定义可直接编写（字段、类型、JSON 标签已明确）
- Handler 层可直接编写（10 个端点的请求/响应格式已明确）
- 错误处理可直接编写（错误码、HTTP 状态码、msg 格式已明确）

**待补充**（扣 1 星）：
- 事务使用场景需在实施前明确（哪些 API 必须用事务？是 sqlx.Tx 还是 gorm.Transaction？）
- 外键约束检查的具体 SQL 查询语句未提供（需实施时自行构造 `SELECT id FROM work_items WHERE id=? AND status=1 AND delete_time IS NULL`）

### 前端实施路径清晰度：⭐⭐⭐⭐⭐（5/5）

**清晰点**：
- 路由配置可直接编写（路径、meta、children 结构已明确）
- TypeScript 类型定义可直接编写（interface 字段、类型已明确）
- API 函数可直接编写（axios 请求路径、参数、返回类型已明确）
- 组件结构可直接实施（文件树、表单字段、验证规则已明确）
- 交互逻辑清晰（loading 状态、成功/错误提示、确认对话框均有 Scenario）

---

## 歧义风险评估

| 条目 | 歧义风险 | 影响 | 缓解建议 |
|------|:-------:|:----:|---------|
| 默认日期范围逻辑 | 🟡 中 | 中 | 前后端统一：后端无参数时自动设置 30 天，前端页面加载时显式传递 30 天参数（双保险） |
| 外键约束检查时机 | 🟡 中 | 中 | Backend spec 补充：Logic 层 CreateWorkRecord/UpdateWorkRecord 开头执行，失败返回 errx.New(99001, "工作事项不存在或已禁用") |
| "本周"快捷选项定义 | 🟢 低 | 低 | Frontend spec 补充：使用 dayjs ISO 周（周一至周日），或明确说明采用自然周 |
| 事务使用场景 | 🟢 低 | 低 | Backend spec 补充：CreateWorkRecord、DeleteWorkItem 使用事务，其他单表操作无需事务 |

---

## 实施建议

1. **优先修复 SHOULD FIX #2 和 #3**：前后端同步明确默认日期范围逻辑，避免实施时出现"前端以为后端会补充"或"后端以为前端会传参"的推诿
2. **后端补充事务清单**：在 Backend Spec §数据库事务章节补充表格，列出哪些 API 使用事务、哪些不用，理由是什么
3. **前端补充日期快捷选项实现**：在 Frontend Spec 补充 dayjs 代码示例，明确"本周"、"上周"等选项的计算逻辑
4. **验收标准可操作化**：Proposal 验收标准可补充测试脚本路径（如"执行 `bash test-e2e.sh` 所有场景通过"）

---

## VERDICT

**APPROVED**

**理由**：
- ✅ 所有需求描述无歧义，SHALL/MUST 语义明确，无占位符或待定项
- ✅ 实现步骤清晰，前后端文件结构、API 路径、数据模型、验证规则均已定义
- ✅ 验收标准可验证，功能验收、技术验收、流水线评估验收三类齐全
- ✅ Proposal ↔ Specs 追溯完整，需求覆盖无遗漏
- 🟡 存在 3 个 SHOULD FIX 问题（默认日期范围逻辑、外键约束检查时机、快捷选项定义），但均为实施细节补充，不影响整体可执行性
- 🔵 5 个 INFO 建议为锦上添花，不阻塞进入下一阶段

**执行建议**：
- 进入阶段 3（任务拆解），同时在任务拆解时补充 SHOULD FIX 问题的实施细节
- 或回阶段 1 用 15 分钟快速修订 Backend/Frontend Spec，补充 SHOULD FIX #1-3，然后进入阶段 3

---

**审查完成时间**: 2026-06-22
**下一步**: 等待 Coverage 和 Structure 视角评审结果，Owner Agent 根据 3 个视角投票决定是否进入阶段 3
