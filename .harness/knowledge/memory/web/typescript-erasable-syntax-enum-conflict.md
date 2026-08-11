---
triggers: ["TypeScript", "erasableSyntaxOnly", "enum", "TS1294", "npm run build", "vue-tsc", "tsconfig"]
type: pitfall
severity: must-follow
service: web
status: active
created: 2026-06-17
updated: 2026-08-09
apply_count: 0
---

# TypeScript erasableSyntaxOnly 与 enum 冲突

**类型**: `pitfall`  
**严重程度**: `must-follow`  
**适用范围**: web/pc, web/mobile（所有前端服务）  
**触发关键词**: `TypeScript`, `erasableSyntaxOnly`, `enum`, `TS1294`, `npm run build`, `vue-tsc`

---

## 问题现象

### 症状

```bash
$ npm run build
> vue-tsc -b && vite build

../common/types/identity.ts(18,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
../common/types/identity.ts(23,13): error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled.
```

**关键特征**:
1. `npm run type-check` 通过（exit 0）
2. `npm run build` 失败（exit 1）
3. 错误信息：`error TS1294: This syntax is not allowed when 'erasableSyntaxOnly' is enabled`
4. 所有 `export enum` 声明被拒绝

### 受影响的操作

- ✅ `npm run type-check` — 仍然通过
- ✅ `npm run test:unit` — 仍然通过
- ❌ `npm run build` — 失败（阻塞部署）
- ❌ 生产构建流程 — 无法生成可部署的静态资源

---

## 根本原因

### 技术背景

`erasableSyntaxOnly` 是 TypeScript 5.5 引入的编译选项，用于强制隔离类型系统和运行时代码。

**定义**（来自 TypeScript 官方文档）:
> When enabled, only allows syntax that can be fully erased during type checking and before JavaScript emit. This means enums, namespaces, parameter properties, and other runtime-affecting TypeScript features are disallowed.

**项目配置**:
```json
// web/pc/tsconfig.app.json
{
  "compilerOptions": {
    "erasableSyntaxOnly": true,  // ← 阻塞所有 enum 声明
    ...
  }
}
```

### 为什么 type-check 通过但 build 失败？

| 命令 | 使用的工具 | 行为 | erasableSyntaxOnly 检查 |
|------|-----------|------|------------------------|
| `npm run type-check` | `vue-tsc --noEmit` | 只做类型检查，不生成代码 | ❌ 不触发（因为不生成 JS） |
| `npm run build` | `vue-tsc -b` | 增量构建模式，生成类型声明文件 | ✅ 触发（因为需要 emit） |

**关键差异**: `--noEmit` 跳过了代码生成阶段，因此不会执行 `erasableSyntaxOnly` 检查。

### 冲突点

项目中大量使用了 `export enum`：
```typescript
// web/common/types/identity.ts
export enum UserType {        // ← 违反 erasableSyntaxOnly
  Staff = 1,
  Homeowner = 2
}

export enum UserStatus {      // ← 违反 erasableSyntaxOnly
  Active = 1,
  Disabled = 2,
  Locked = 3
}
```

TypeScript 的 `enum` 会生成运行时代码：
```javascript
// 编译后的 JavaScript
var UserType;
(function (UserType) {
    UserType[UserType["Staff"] = 1] = "Staff";
    UserType[UserType["Homeowner"] = 2] = "Homeowner";
})(UserType || (UserType = {}));
```

这与 `erasableSyntaxOnly` 的"类型完全可擦除"理念冲突。

---

## 复现路径

### 前置条件

1. `tsconfig.app.json` 中启用了 `"erasableSyntaxOnly": true`
2. 项目中存在 `export enum` 声明（非 `const enum`）
3. 使用 `vue-tsc -b`（增量构建模式）

### 复现步骤

```bash
cd web/pc
npm run build  # FAIL: error TS1294
```

### 影响范围

- 所有前端服务（web/pc, web/mobile）
- 所有使用 `export enum` 的共享类型（web/common/types/）

---

## 解决方案

### 方案 1: 移除 erasableSyntaxOnly（推荐）

**适用场景**: 项目已有大量 enum 使用，且没有强制类型隔离的需求。

**操作**:
```diff
// web/pc/tsconfig.app.json
{
  "compilerOptions": {
    "strict": true,
-   "erasableSyntaxOnly": true,
    ...
  }
}
```

**优点**:
- 立即修复，零代码改动
- 保持现有 enum 使用模式
- 无需重构类型引用

**缺点**:
- 失去类型隔离保护（类型可能影响运行时）

**工作量**: 1 分钟

---

### 方案 2: 改用 const enum

**适用场景**: 希望保留 `erasableSyntaxOnly` 约束，且 enum 值可以内联。

**操作**:
```diff
// web/common/types/identity.ts
- export enum UserType {
+ export const enum UserType {
    Staff = 1,
    Homeowner = 2
  }
```

**优点**:
- 保留 erasable 约束
- 运行时零开销（编译时内联）
- 类型安全不变

**缺点**:
- 需全局替换所有 enum
- `const enum` 不能在运行时反射（`Object.keys(UserType)` 不可用）
- 可能影响类型推断（在某些边界情况下）

**工作量**: 2 小时（需修改所有 enum 声明 + 验证编译）

---

### 方案 3: 改用 as const 对象（现代最佳实践）

**适用场景**: 新项目或愿意投入时间重构的项目。

**操作**:
```typescript
// web/common/types/identity.ts
export const UserType = {
  Staff: 1,
  Homeowner: 2
} as const

export type UserType = typeof UserType[keyof typeof UserType]  // 1 | 2
```

**优点**:
- 完全可擦除（纯类型）
- 可以运行时反射（`Object.keys(UserType)`）
- TypeScript 现代最佳实践

**缺点**:
- 需重构所有 enum 引用
- 运行时有对象开销（相比 const enum）
- 类型声明更冗长

**工作量**: 4 小时（需修改所有 enum 声明 + 所有引用点 + 验证编译）

---

## 方案对比

| 方案 | 修复速度 | 运行时开销 | 类型隔离 | 反射能力 | 工作量 |
|------|---------|-----------|---------|---------|--------|
| **移除 erasableSyntaxOnly** | 立即 | enum 对象 | ❌ | ✅ | 1 分钟 |
| **const enum** | 快 | 零（内联） | ✅ | ❌ | 2 小时 |
| **as const 对象** | 慢 | 对象 | ✅ | ✅ | 4 小时 |

---

## 决策建议

### 当前项目选择：方案 1（移除 erasableSyntaxOnly）

**理由**:
1. 项目已有大量 enum 使用（identity.ts, aimodel.ts, masterdata.ts 等）
2. 没有强制类型隔离的业务需求
3. 成本/收益比：1 分钟 vs 2-4 小时（其他方案）
4. 风险最低（零代码改动）

**实施**:
```bash
# 编辑 tsconfig.app.json，删除 "erasableSyntaxOnly": true
vim web/pc/tsconfig.app.json

# 验证
cd web/pc
npm run build  # 应通过
```

---

## 预防措施

### 新项目启动清单

如果从 Vue 3 官方模板创建新项目（默认启用 `erasableSyntaxOnly`），需要：

1. **决策阶段**: 确定是否需要严格类型隔离
   - 需要 → 使用 `const enum` 或 `as const`
   - 不需要 → 移除 `erasableSyntaxOnly`

2. **编码规范**: 在 CLAUDE.md 中明确 enum 使用规范
   - 禁止 `export enum`（如果启用 erasableSyntaxOnly）
   - 强制 `export const enum` 或 `as const`

3. **CI 检查**: 确保 `npm run build` 在提交前执行（而非仅 `type-check`）

---

## 相关记录

- **发现时间**: 2026-06-22
- **影响服务**: web/pc
- **QA 报告**: `web/pc/_qa.md` (2026-06-22 22:07)
- **错误码**: TS1294
- **TypeScript 版本**: 5.5+

---

## 参考资料

- [TypeScript 5.5 Release Notes](https://www.typescriptlang.org/docs/handbook/release-notes/typescript-5-5.html#the-erasablesyntaxonly-flag)
- [Vue 3 TypeScript 最佳实践](https://vuejs.org/guide/typescript/overview.html)
- [Enum vs Const Enum vs As Const](https://www.typescriptlang.org/docs/handbook/enums.html)
