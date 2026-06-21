# QA Report — pc (web/pc/)

**验证时间**: 2026-06-19 10:15
**验证范围**: main 分支当前状态
**QA Agent**: 前端服务验证

---

## 机械化检查结果 (harness-checks-frontend.sh — FRESH run)

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | TypeScript type check | ✅ | `vue-tsc --noEmit` 通过 |
| 2 | Unit tests | ✅ | 52 tests passed (5 test files) |
| 3 | Production build | ❌ | **147 TypeScript errors** |
| 4 | Hardcoded secrets | ✅ | no secrets detected in source |
| 5 | Debug artifacts | ✅ | no debug artifacts in production code |
| 6 | TypeScript type safety | ⚠️ | 67 'as any' usages (aspirational target ≤10) |

**JSON 输出**:
```json
{
  "timestamp": "2026-06-19T02:14:21Z",
  "service": "pc",
  "results": [
    {"check":"type_check","status":"PASS","detail":"pc: type check passed"},
    {"check":"unit_test","status":"PASS","detail":"pc: 52 tests passed"},
    {"check":"build","status":"FAIL","detail":"pc: 147 TS errors"},
    {"check":"hardcoded_secrets","status":"PASS","detail":"no secrets detected in source"},
    {"check":"debug_artifacts","status":"PASS","detail":"no debug artifacts in production code"},
    {"check":"type_safety","status":"WARN","detail":"67 'as any' usages"}
  ],
  "summary": {"pass": 4, "fail": 1, "warn": 1},
  "exit_code": 1
}
```

---

## 构建失败详情（npm run build — FRESH run，exit code 1）

### 错误分类

**总计 147 个 TypeScript 错误**，主要分为以下类别：

#### 1. erasableSyntaxOnly 语法错误（7 个）
`tsconfig.app.json` 开启了 `"erasableSyntaxOnly": true`，禁止使用 `const enum`，但 `web/common/types/identity.ts` 使用了 `enum` 导出：

```
../common/types/identity.ts(18,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
../common/types/identity.ts(23,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
../common/types/identity.ts(29,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
../common/types/identity.ts(49,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
../common/types/identity.ts(70,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
../common/types/identity.ts(75,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
../common/types/identity.ts(98,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
```

**受影响的枚举**:
- `UserType` (L18)
- `UserStatus` (L23)
- `VerificationStatus` (L29)
- `RoleStatus` (L49)
- `PermissionType` (L70)
- `PermissionStatus` (L75)
- `HomeownerVerificationStatus` (L98)

#### 2. API 响应类型不匹配（~50 个）
多个组件中，`DefaultRow` 类型无法赋值给业务类型，说明 Element Plus Table 的 row 数据需要显式类型断言：

**aimodel/**:
- `src/api/aimodel.ts(320,324)`: `model_name` 字段重复定义且类型冲突
- `src/views/aimodel/ModelList.vue`: `DefaultRow` → `ModelConfig` 不兼容（多处）
- `src/views/aimodel/Statistics.vue`: 缺少 `models`/`statistics` 字段，`model_name` 不存在
- `src/views/aimodel/TemplateList.vue`: `DefaultRow` → `PromptTemplate` 不兼容

**moderation/**:
- `src/components/moderation/ReviewTable.vue`: `DefaultRow` → `ReviewListItem` 不兼容
- `src/components/moderation/LayerConfigPanel.vue`: 事件处理器类型不匹配（`boolean` vs `string | number | boolean`）
- `src/components/moderation/PipelineResultPanel.vue`: `string` 不能赋值给 Element Plus `type` 属性

**其他模块**:
- `src/views/division/Index.vue`: `DefaultRow` → `AdministrativeDivision` 不兼容
- `src/views/approval-center/Index.vue`: `DefaultRow` → `ApprovalPendingItem` 不兼容
- `src/views/deleted-recovery/Index.vue`: `DefaultRow` → `DeletedItem` 不兼容
- `src/stores/division.ts(128)`: `PaginatedResponse<T>` 被当作 `T[]` 使用

#### 3. 类型定义错误（~10 个）
- `src/components/business/PermissionTree.vue(63,73)`: `string[]` vs `number[]` 不兼容
- `src/stores/division.ts(106)`: `hasChildren` 不存在于 `AdministrativeDivision`
- `src/views/aimodel/ModelForm.vue(304)`: `config_key` 不存在于类型中
- `src/views/amap-sync/Index.vue`: `number` 不能赋值给 `string`（多处）

#### 4. 死代码（2 个）
- `src/views/aimodel/Statistics.vue(118)`: `reactive` 声明但未使用
- `src/views/aimodel/TemplateList.vue(167)`: `ModelConfig` 声明但未使用

---

## 单元测试（npm run test:unit — FRESH run，exit 0）

```
✅ Test Files  5 passed (5)
✅ Tests      52 passed (52)
   Duration   3.19s
```

**测试文件清单**:
1. `tests/e2e/auth.spec.ts`
2. `tests/unit/directives/permission.spec.ts`
3. `tests/unit/stores/auth.spec.ts`
4. `tests/unit/stores/permission.spec.ts`
5. `tests/unit/utils/request.spec.ts`
6. `tests/unit/views/aimodel/ModelForm.spec.ts`

---

## TDD 证据检查

**现状**: 无法从当前测试运行中提取 TDD 证据（RED→GREEN 过程），因为测试套件已处于稳定状态。

**测试覆盖评估**:
- ✅ 核心功能有测试：auth store、permission store、request utils
- ⚠️ 大量业务组件缺少测试：aimodel/Statistics.vue、division/Index.vue、moderation/* 等
- ⚠️ API 层缺少测试：`src/api/*.ts` 无对应测试文件

---

## 类型安全警告（WARN，非阻塞）

检测到 **67 处 `as any` 使用**（期望目标 ≤10）。高频文件：
- `web/pc/src/stores/division.ts:135`
- `web/pc/src/utils/permission.ts:44`
- `web/pc/src/utils/request.ts:93,108,114,211`
- `web/pc/src/utils/crypto.ts:74`
- `web/pc/src/views/division/Index.vue:156,164,312,355,368,382`
- 其他组件：多处

**影响**: 绕过 TypeScript 类型检查，可能隐藏运行时错误。

---

## 根本原因分析

### 问题 1: erasableSyntaxOnly vs enum（P0）

**配置冲突**:
- `tsconfig.app.json` 设置 `"erasableSyntaxOnly": true`（禁止运行时代码，仅允许类型标注）
- `web/common/types/identity.ts` 导出 7 个 `enum`（会编译为运行时代码）

**解决方案**:
1. **选项 A（推荐）**: 将 enum 改为 `const enum`（编译时内联，无运行时代码）
2. **选项 B**: 移除 `erasableSyntaxOnly` 配置（允许 enum 运行时代码）
3. **选项 C**: 将 enum 改为 `type` + `as const` 对象（纯类型定义）

### 问题 2: PaginatedResponse 解包错误（P0）

**示例** (`src/stores/division.ts:128`):
```typescript
// 错误：将 PaginatedResponse<T> 直接当作 T[] 使用
buildTree(response as AdministrativeDivision[])

// 正确：应该先解包
buildTree(response.data || [])
```

**影响范围**: 所有使用分页 API 的组件。

### 问题 3: Element Plus Table row 类型（P1）

**现象**: `@row-click="handler"` 收到 `DefaultRow` 类型，但业务代码期望具体类型（如 `ModelConfig`）。

**原因**: Element Plus Table 的泛型定义未正确传递。

**解决方案**:
```vue
<!-- 当前（错误） -->
<el-table @row-click="handleEdit(row)">

<!-- 正确 -->
<el-table @row-click="(row) => handleEdit(row as ModelConfig)">
```

---

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| 🔴 CRITICAL | 构建失败（147 个 TS 错误），无法生成生产构建 | 必须修复才能部署 |
| 🔴 P0 | `erasableSyntaxOnly` + `enum` 配置冲突（7 个错误） | 将 enum 改为 `const enum` 或移除配置项 |
| 🔴 P0 | `PaginatedResponse<T>` 未正确解包（多处） | 统一修改为 `response.data` 访问 |
| 🟡 P1 | Element Plus Table `DefaultRow` 类型断言缺失（~50 个） | 在事件处理器中添加类型断言 |
| 🟡 P1 | `as any` 过多（67 处 vs 目标 ≤10） | 逐步消除类型逃逸 |
| ⚪ P2 | 测试覆盖不足（大量业务组件无测试） | 为关键业务组件补充单元测试 |

---

## 矛盾现象说明

**为何 `type-check` 通过但 `build` 失败？**

```bash
# type-check: 仅检查类型，不生成代码（通过）
npm run type-check  # → vue-tsc --noEmit

# build: 类型检查 + 生成代码（失败）
npm run build       # → vue-tsc -b && vite build
```

**原因**: `vue-tsc -b` 启用了项目引用模式（`-b` = build），会额外检查 `tsconfig.json` 中的 `references`，包括 `../common/types/identity.ts`，而该文件的 `enum` 语法违反了 `erasableSyntaxOnly` 规则。

`vue-tsc --noEmit` 仅检查 `src/` 下的文件（`tsconfig.app.json` 的 `include` 范围），不检查 `../common/`，因此未发现错误。

---

## VERDICT: ❌ FAIL

**阻塞原因**:
1. **生产构建失败** — 147 个 TypeScript 错误导致 `npm run build` 失败（exit code 1）
2. **配置冲突** — `erasableSyntaxOnly` 与 `enum` 使用不兼容（7 个错误）
3. **类型错误** — API 响应解包错误、类型断言缺失（~140 个错误）

**非阻塞警告**:
- 67 处 `as any` 使用（超出期望目标）
- 测试覆盖不足（仅 5 个测试文件）

---

## 修复建议优先级

### Phase 1（立即修复，阻塞部署）
1. 修复 `web/common/types/identity.ts` 的 enum 问题（改为 `const enum`）
2. 修复 `PaginatedResponse<T>` 解包错误（统一使用 `response.data`）
3. 修复 `src/api/aimodel.ts` 的 `model_name` 重复定义

### Phase 2（短期，1-2 天）
4. 为 Element Plus Table 事件处理器添加类型断言
5. 修复 `stores/division.ts` 的类型错误（`hasChildren`、`PaginatedResponse` 使用）
6. 移除死代码（`reactive`、`ModelConfig` 未使用声明）

### Phase 3（中期，1 周）
7. 逐步消除 `as any` 使用（目标降至 ≤10）
8. 为关键业务组件补充单元测试

---

**验证方式**:
```bash
cd web/pc
npm run build   # 必须 exit 0
npm run test:unit   # 必须全部通过
bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service pc --json   # 必须 PASS
```

---
**QA Agent**: 前端服务验证（只读权限）
**生成时间**: 2026-06-19 10:15
