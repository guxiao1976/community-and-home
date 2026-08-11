# Requesting Code Review 技能使用模板

## 技能名称
`requesting-code-review`

## 功能描述
生成标准化、结构化的代码审查请求，包括变更摘要、审查清单、测试覆盖和部署计划，提升代码审查的效率和质量。

## 使用场景
- 完成代码开发，准备提交 PR
- 生成 GitHub/GitLab PR 描述
- 标准化审查流程
- 确保审查覆盖所有关键点
- 提升审查者效率

## 输入要求

### 必需信息
- 变更内容概述
- 影响的文件和模块
- QA 测试结果
- Review 结果（如有）

### 可选信息
- 截图或演示视频
- 性能测试数据
- 安全性分析
- 相关 Issue/文档链接

## 输出内容
1. PR 标题（简洁，< 70 字符）
2. 变更说明
3. 审查清单
4. 测试覆盖说明
5. 部署计划
6. 回滚方案

## 调用示例

### 示例 1: 新功能 PR
```
为用户权限管理功能生成代码审查请求：

## 变更概述
实现基于 RBAC 的用户权限管理系统

## 变更文件
后端 (permission-service):
- api/internal/types/types.go (+150, -0)
- api/internal/logic/role/*.go (+400, -0)
- api/internal/logic/permission/*.go (+350, -0)
- rpc/internal/logic/*.go (+500, -0)
- model/role.go (+120, -0)
- model/permission.go (+100, -0)

前端 (web/pc):
- src/views/permission/RoleList.vue (+200, -0)
- src/views/permission/PermissionConfig.vue (+250, -0)
- src/api/permission.ts (+80, -0)

数据库:
- migrations/002_create_permission_tables.sql (+50, -0)

## QA 结果
✅ 构建通过
✅ 单元测试通过 (覆盖率 85%)
✅ E2E 测试通过 (8/8 用例)
✅ 代码规范检查通过

## Review 结果
✅ 安全架构审查: PASS
✅ 规范工程审查: PASS (2 个 WARNING)
✅ 设计业务审查: PASS

## 技术决策
1. 使用 Casbin 作为权限引擎
2. 权限数据缓存到 Redis (TTL 5分钟)
3. 前端使用树形控件展示权限层级

## 测试覆盖
- 单元测试: 28 个测试用例
- E2E 测试: 8 个用户流程
- 手工测试: 权限边界条件验证

## 部署计划
1. 数据库迁移
2. 部署 permission-service RPC
3. 部署 permission-service API
4. 部署前端
5. 验证基本功能

## 已知问题
- 2 个代码规范 WARNING (非阻塞)
  - role_test.go: 测试覆盖率 78% (目标 80%)
  - permission_config.vue: 组件复杂度 15 (建议 < 10)
```

**期望输出**:
```markdown
# feat(permission): 实现 RBAC 权限管理系统

## 📋 变更说明

本 PR 实现了基于 RBAC (Role-Based Access Control) 的用户权限管理系统，包括后端 gRPC 服务和前端管理界面。

### 核心功能
- ✅ 角色管理（创建、编辑、删除、列表）
- ✅ 权限配置（为角色分配权限，树形结构）
- ✅ 用户角色关联（为用户分配多个角色）
- ✅ 权限检查中间件（API 层和 gRPC 层）

### 技术亮点
- 使用 Casbin 作为权限引擎，灵活且性能高
- 权限数据缓存到 Redis，减少数据库压力
- 前端树形控件支持全选/反选/搜索

## 📂 变更文件

### 后端 (permission-service)
- `api/internal/logic/role/` - 角色管理 API 逻辑 (+400 行)
- `api/internal/logic/permission/` - 权限配置 API 逻辑 (+350 行)
- `rpc/internal/logic/` - gRPC 服务实现 (+500 行)
- `model/role.go`, `model/permission.go` - 数据模型 (+220 行)

### 前端 (web/pc)
- `src/views/permission/RoleList.vue` - 角色列表页 (+200 行)
- `src/views/permission/PermissionConfig.vue` - 权限配置页 (+250 行)
- `src/api/permission.ts` - API 调用封装 (+80 行)

### 数据库
- `migrations/002_create_permission_tables.sql` - 数据表创建 (+50 行)

**总计**: +1,730 行, -0 行

## ✅ 审查清单

请审查者重点关注以下方面：

### 功能正确性
- [ ] 角色 CRUD 操作逻辑正确
- [ ] 权限配置保存和加载正确
- [ ] 用户角色关联逻辑正确
- [ ] 权限检查中间件正确拦截无权限请求

### 代码质量
- [ ] 遵循项目编码规范
- [ ] 错误处理完善
- [ ] 日志记录合理
- [ ] 注释清晰

### 性能
- [ ] Redis 缓存策略合理
- [ ] 数据库查询优化（索引、N+1 问题）
- [ ] 前端组件渲染性能（大量权限时）

### 安全性
- [ ] 权限检查无遗漏
- [ ] SQL 注入防护
- [ ] XSS 防护（前端输入验证）
- [ ] 敏感操作有二次确认

### 可维护性
- [ ] 代码结构清晰，模块职责单一
- [ ] 测试覆盖充分
- [ ] API 文档完整

## 🧪 测试覆盖

### 单元测试
- **后端**: 28 个测试用例，覆盖率 85%
  - `role_test.go`: 8 个用例 (CRUD + 边界条件)
  - `permission_test.go`: 10 个用例 (配置保存/加载/验证)
  - `middleware_test.go`: 10 个用例 (权限检查各种场景)

- **前端**: 12 个组件测试
  - `RoleList.spec.ts`: 6 个用例
  - `PermissionConfig.spec.ts`: 6 个用例

### E2E 测试
- ✅ 创建角色并分配权限
- ✅ 为用户分配角色
- ✅ 无权限用户访问受保护接口 (403)
- ✅ 有权限用户访问受保护接口 (200)
- ✅ 编辑角色权限，验证缓存更新
- ✅ 删除角色，验证用户失去相应权限
- ✅ 批量分配角色
- ✅ 权限树搜索和筛选

### 手工测试清单
- [ ] 超级管理员可以访问所有功能
- [ ] 普通管理员只能访问被授权的功能
- [ ] 禁用角色后，用户立即失去权限（缓存失效）
- [ ] 并发修改权限配置，数据一致性保持

## 🚀 部署计划

### 前置条件
- MySQL 数据库可用
- Redis 可用
- etcd 可用

### 部署步骤
```bash
# 1. 数据库迁移
cd services/permission-service
mysql -u root -p community < migrations/002_create_permission_tables.sql

# 2. 部署 RPC 服务
cd rpc
go build -o permission-rpc
./permission-rpc -f etc/permission.yaml

# 3. 部署 API 服务
cd ../api
go build -o permission-api
./permission-api -f etc/permission-api.yaml

# 4. 部署前端
cd ../../web/pc
npm run build
# 将 dist/ 部署到 Web 服务器
```

### 验证步骤
1. 访问 `/permission/roles` 页面，验证角色列表加载正常
2. 创建测试角色，分配权限，保存成功
3. 为测试用户分配角色
4. 使用测试用户登录，验证权限生效

### 回滚方案
```bash
# 如果出现问题，回滚步骤：
# 1. 停止新版本服务
systemctl stop permission-api permission-rpc

# 2. 启动旧版本服务
systemctl start permission-api-old permission-rpc-old

# 3. 回滚数据库（如果需要）
mysql -u root -p community < migrations/002_rollback.sql

# 4. 回滚前端（恢复旧版本）
cd /var/www/html/pc
rm -rf assets
cp -r assets.backup/* .
```

### 监控指标
- 权限检查接口响应时间 < 50ms
- Redis 缓存命中率 > 95%
- 权限配置保存成功率 > 99.9%

## 📸 截图

### 角色列表页
![角色列表](./screenshots/role-list.png)

### 权限配置页
![权限配置](./screenshots/permission-config.png)

## ⚠️ 已知问题

### Non-blocking Warnings (不阻塞合并)
1. **role_test.go 测试覆盖率 78%**
   - 目标: 80%
   - 原因: 部分错误分支未覆盖
   - 计划: 后续补充

2. **PermissionConfig.vue 组件复杂度 15**
   - 建议: < 10
   - 原因: 权限树交互逻辑复杂
   - 计划: 后续重构为多个子组件

## 🔗 相关链接

- Issue: #123 - 实现用户权限管理
- 设计文档: `docs/specs/rbac-design.md`
- API 文档: `services/permission-service/docs/api.md`
- Casbin 文档: https://casbin.org/

## 📝 审查注意事项

1. **安全性优先**: 权限检查是安全关键功能，务必仔细审查
2. **性能考虑**: 关注缓存策略和数据库查询性能
3. **边界条件**: 关注空角色、空权限、并发修改等边界情况

---

**Co-Authored-By**: Claude Opus 4.8 (1M context) <noreply@anthropic.com>

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

### 示例 2: Bug 修复 PR
```
为竞态条件修复生成代码审查请求：

## 问题描述
activate_pipeline 接口在高并发下出现竞态条件，导致部分 pipeline 未正确激活。

## 根因
使用逐条循环更新，每次更新提交一个事务，导致中间状态不一致。

## 修复方案
使用数据库事务 + 批量更新替代逐条循环。

## 变更文件
- `rpc/internal/logic/activate_pipeline_logic.go` (+15, -25)
- `rpc/internal/logic/activate_pipeline_logic_test.go` (+50, -0)

## QA 结果
✅ 构建通过
✅ 单元测试通过 (新增回归测试)
✅ 并发压测通过 (100 并发，0 错误)

## Review 结果
✅ 所有审查通过
```

**期望输出**: Bug 修复 PR 描述，包含问题描述、根因分析、修复方案、测试验证...

### 示例 3: 性能优化 PR
```
为 API 性能优化生成代码审查请求：

## 优化目标
用户列表接口响应时间从 3-5s 优化到 < 500ms

## 优化措施
1. 添加数据库索引 (phone, created_at)
2. 解决 N+1 查询问题（使用 JOIN）
3. 添加 Redis 缓存（热点数据）
4. 减少返回字段（移除不必要的 JOIN）

## 性能测试结果
- 优化前: 平均 3.2s, P95 5.1s
- 优化后: 平均 280ms, P95 450ms
- 提升: 91% (10x+)

## 变更文件
- `api/internal/logic/list_users_logic.go` (+30, -50)
- `model/user.go` (+20, -5)
- `migrations/003_add_indexes.sql` (+10, -0)
```

**期望输出**: 性能优化 PR 描述，包含优化前后对比、性能测试数据、风险评估...

## 最佳实践

### 1. 标题规范
遵循 Conventional Commits 规范：

```
feat(scope): 简短描述 (< 70 字符)
fix(scope): 修复问题描述
perf(scope): 性能优化描述
refactor(scope): 重构描述
docs(scope): 文档更新
test(scope): 测试相关
chore(scope): 构建/工具链
```

### 2. 变更说明清晰
- 说明"做了什么"和"为什么"
- 突出技术亮点和关键决策
- 提及影响范围

### 3. 审查清单完整
覆盖：
- 功能正确性
- 代码质量
- 性能影响
- 安全性
- 可维护性

### 4. 测试覆盖充分
- 单元测试覆盖率
- E2E 测试场景
- 手工测试清单
- 性能测试数据（如有）

### 5. 部署计划详细
- 部署步骤
- 验证步骤
- 回滚方案
- 监控指标

## PR 描述模板

```markdown
# <type>(<scope>): <subject>

## 变更说明
<详细说明本次变更的内容和原因>

## 变更文件
<列出主要变更文件及行数>

## 审查清单
- [ ] 功能正确性
- [ ] 代码质量
- [ ] 性能影响
- [ ] 安全性
- [ ] 可维护性

## 测试覆盖
<单元测试、E2E 测试、手工测试清单>

## 部署计划
<部署步骤、验证步骤、回滚方案>

## 截图/演示
<如适用>

## 相关链接
- Issue: #xxx
- 设计文档: docs/xxx.md
```

## 与其他技能配合

### requesting-code-review + harness-pipeline
```
1. harness-pipeline 完成开发和测试
2. requesting-code-review 生成 PR 描述
3. 提交 PR，等待审查
```

### requesting-code-review + code-review (plugin)
```
1. requesting-code-review 生成审查请求
2. code-review plugin 自动化审查
3. 人工 Review 最终确认
```

## 注意事项

1. **标题简洁**
   - < 70 字符
   - 遵循规范
   - 描述清晰

2. **变更说明完整**
   - 说明"是什么"和"为什么"
   - 突出关键点
   - 提及风险

3. **审查清单针对性**
   - 根据变更类型调整
   - 突出重点审查项
   - 避免泛泛而谈

4. **测试证据充分**
   - 提供覆盖率数据
   - 列出关键测试场景
   - 附上测试结果

5. **部署计划可执行**
   - 步骤清晰
   - 可重复执行
   - 有回滚方案

## 常见问题

### Q: PR 描述应该多详细？
A: 足够让审查者理解变更内容和审查重点。关键变更要详细，简单变更可简略。

### Q: 如何处理大型 PR？
A: 建议拆分为多个小 PR。如果必须大 PR，提供清晰的模块划分和审查顺序。

### Q: 审查清单是否需要全部勾选？
A: 审查清单是提醒事项，不是强制要求。根据实际情况调整。

### Q: 如何写好部署计划？
A: 像写给运维工程师的操作手册，步骤清晰，命令可复制粘贴。

## 相关文档
- [Conventional Commits](https://www.conventionalcommits.org/)
- [How to Write a Git Commit Message](https://chris.beams.io/posts/git-commit/)
- [Google's Code Review Guidelines](https://google.github.io/eng-practices/review/)
