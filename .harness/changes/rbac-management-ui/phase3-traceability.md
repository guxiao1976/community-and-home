# Phase 3: Brainstorming 决策 → Spec 覆盖表

## 决策点追溯

| # | Brainstorming 决策点 | 决策结果 | 覆盖的 Spec | 覆盖章节 | 状态 |
|---|---------------------|---------|------------|---------|:----:|
| 1 | 角色设计范围 | 混合模式（4个系统角色 + 自定义角色） | `role-management/spec.md` | § 2.1-2.4, § 5.1 | ✅ |
| 2 | 权限粒度设计 | 菜单+功能点两层结构 | `permission-management/spec.md` | § 2.1, § 5.1 | ✅ |
| 2 | 权限粒度设计 | 菜单+功能点两层结构 | `menu-permission-control/spec.md` | § 2.1-2.3, § 5.1 | ✅ |
| 3 | 数据权限作用域 | 分配角色时指定作用域 | `user-role-assignment/spec.md` | § 2.2, § 5.2 | ✅ |
| 4 | 接口级权限控制 | 外部请求统一拦截 + 内部 REST 二次校验 | `api-permission-control/spec.md` | § 2.1, § 2.4 | ✅ |
| 4 | 接口级权限控制 | 内部 gRPC 不校验 | `api-permission-control/spec.md` | § 2.4 | ✅ |

---

## 用户需求点 → Spec 覆盖表

| 用户需求（原话） | 拆解后的需求点 | 覆盖的 Spec | 覆盖章节 | 状态 |
|----------------|--------------|------------|---------|:----:|
| "目前还没有完整的角色管理、权限管理等页面" | 角色 CRUD 界面 | `role-management/spec.md` | § 2.1-2.4 | ✅ |
| "目前还没有完整的角色管理、权限管理等页面" | 权限树配置界面 | `permission-management/spec.md` | § 2.1-2.2 | ✅ |
| "根据系统中已有的用户进行角色设计、权限设计" | 4个系统角色（owner/property_admin/community_admin/grid_worker） | `role-management/spec.md` | § 5.1, proposal § 2.1 | ✅ |
| "不同角色的人能够看到不同的菜单" | 前端路由动态过滤 | `menu-permission-control/spec.md` | § 2.1-2.2 | ✅ |
| "前端能够看到不同的内容，比如，发布通知，只有授权的人员才能看到相关图标" | 按钮级权限控制（`v-permission` 指令） | `menu-permission-control/spec.md` | § 2.3 | ✅ |
| "权限管理要实现接口级的管理，不能仅仅是前端是否展示相关按钮或者内容" | 后端 REST API 权限中间件 | `api-permission-control/spec.md` | § 2.1 | ✅ |
| "这样知道api时，容易绕过去" | 后端权限校验（防前端绕过） | `api-permission-control/spec.md` | § 2.1, § 9.4 | ✅ |
| "数据权限一并考虑，只能看到授权了的数据" | 用户角色分配含数据作用域 | `user-role-assignment/spec.md` | § 2.2, § 5.2 | ✅ |
| "数据权限一并考虑，只能看到授权了的数据" | 数据隔离实现（GetDataScopes） | `user-role-assignment/spec.md` | § 2.2（说明段落） | ✅ |

---

## Capability 完整性检查

| Capability | Spec 文件 | 必备章节完整性 | 状态 |
|-----------|----------|--------------|:----:|
| 角色管理 | `role-management/spec.md` | ✅ 功能需求、数据模型、接口清单、业务规则、界面交互、错误处理、测试场景、依赖约束、追溯 | ✅ |
| 权限管理 | `permission-management/spec.md` | ✅ 功能需求、数据模型、接口清单、业务规则、界面交互、权限树构建、初始化脚本、错误处理、测试场景、依赖约束、追溯 | ✅ |
| 用户角色分配 | `user-role-assignment/spec.md` | ✅ 功能需求、数据模型、接口清单、业务规则、界面交互、作用域关联查询、错误处理、测试场景、依赖约束、追溯 | ✅ |
| 菜单权限控制 | `menu-permission-control/spec.md` | ✅ 功能需求、数据模型、接口清单、业务规则、界面交互、路由配置、错误处理、测试场景、依赖约束、追溯 | ✅ |
| 接口权限控制 | `api-permission-control/spec.md` | ✅ 功能需求、数据模型、接口清单、业务规则、中间件集成、性能优化、错误处理、测试场景、依赖约束、追溯 | ✅ |

---

## 跨 Spec 依赖关系

```
role-management (角色管理)
  ↓ 依赖：角色列表、角色详情接口
permission-management (权限管理)
  ↓ 依赖：权限树数据
user-role-assignment (用户角色分配)
  ↓ 依赖：角色列表、作用域数据
menu-permission-control (菜单权限控制)
  ↓ 依赖：用户权限码列表
api-permission-control (接口权限控制)
  ↓ 依赖：CheckPermission 接口、权限表

所有 Spec ← role-management（提供角色基础数据）
menu-permission-control + api-permission-control ← permission-management（提供权限树数据）
```

**依赖闭环检查**：✅ 无循环依赖，所有依赖关系清晰

---

## 遗漏检查

### ✅ 已覆盖

- 角色 CRUD（创建、编辑、删除、列表、详情）
- 系统角色保护（不可删除、天然全权限）
- 权限树展示（三层：菜单、按钮、API）
- 角色权限配置（勾选权限树、全量替换）
- 用户角色分配（含数据作用域、批量分配）
- 菜单权限控制（路由过滤、侧边栏过滤、按钮级指令）
- 接口权限控制（REST 中间件、白名单、缓存策略）
- 权限数据初始化脚本
- 错误处理和安全审计

### ⚠️ 明确排除（OUT OF SCOPE）

- APISIX Gateway 统一拦截（基础设施层，配置层面）
- 内部 gRPC 权限校验（信任域内调用，不重复校验）
- 历史用户权限迁移（视现有数据情况决定，非功能性需求）

### ℹ️ 隐式依赖（需在实施时确认）

- `master-data-service` 提供作用域数据（小区、楼栋、单元、网格）→ `user-role-assignment` 依赖
- 前端路由配置需增加 `meta.permissions` 字段 → `menu-permission-control` 依赖
- 各服务需集成权限中间件 → `api-permission-control` 依赖

---

## 总结

| 维度 | 结果 |
|------|------|
| **决策点覆盖** | 4/4 ✅（所有 Brainstorming 决策点已转换为 Spec） |
| **需求点覆盖** | 9/9 ✅（所有用户需求点已覆盖） |
| **Capability 完整性** | 5/5 ✅（所有 Capability 的 Spec 章节完整） |
| **跨 Spec 依赖** | ✅ 无循环依赖 |
| **遗漏检查** | ✅ 无功能遗漏，排除项明确 |

**结论**：需求分析 Phase 2 产出完整，可进入 Phase 4 Self-Review。
