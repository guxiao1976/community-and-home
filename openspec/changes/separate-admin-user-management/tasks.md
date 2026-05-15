## 1. Backend API Enhancement

- [x] 1.1 确认identity服务中管理员角色的标识方式（role_id或role字段的值）
- [x] 1.2 在用户查询API的请求结构中添加role_type过滤参数（可选字段）
- [x] 1.3 更新.api文件，添加role_type参数到UserListReq
- [x] 1.4 在用户查询逻辑中实现按role_type过滤的SQL条件
- [x] 1.5 使用goctl重新生成API代码
- [x] 1.6 测试API过滤功能（分别传递admin和regular参数）

## 2. Frontend - Admin User Management Page

- [x] 2.1 在web/pc/src/views/identity/下创建admin-user目录
- [ ] 2.2 创建AdminUserList.vue组件（列表页面）
- [ ] 2.3 实现管理员用户列表查询（调用API时传递role_type=admin）
- [ ] 2.4 实现分页功能
- [ ] 2.5 实现搜索功能（用户名、手机号）
- [ ] 2.6 创建AdminUserEdit.vue组件或对话框（编辑功能）
- [ ] 2.7 实现管理员用户信息修改功能
- [ ] 2.8 添加表单验证（邮箱格式、必填项等）
- [ ] 2.9 实现权限控制（基于identity:admin-user:list和identity:admin-user:update）

## 3. Frontend - Regular User Management Page

- [ ] 3.1 在web/pc/src/views/identity/下创建regular-user目录
- [ ] 3.2 创建RegularUserList.vue组件（列表页面）
- [ ] 3.3 实现普通用户列表查询（调用API时传递role_type=regular）
- [ ] 3.4 实现分页功能
- [ ] 3.5 实现搜索功能（用户名、手机号）
- [ ] 3.6 创建RegularUserEdit.vue组件或对话框（编辑功能）
- [ ] 3.7 实现普通用户信息修改功能
- [ ] 3.8 添加表单验证（邮箱格式、必填项等）
- [ ] 3.9 实现权限控制（基于identity:regular-user:list和identity:regular-user:update）

## 4. Frontend - Routing Configuration

- [ ] 4.1 在router配置中添加管理员管理页面路由
- [ ] 4.2 在router配置中添加普通用户管理页面路由
- [ ] 4.3 配置路由权限守卫（meta.permissions）

## 5. Menu and Permission Configuration

- [ ] 5.1 在数据库菜单表中添加"管理员管理"菜单项
- [ ] 5.2 配置"管理员管理"菜单的权限码（identity:admin-user:list, identity:admin-user:update）
- [ ] 5.3 在数据库菜单表中添加"普通用户管理"菜单项
- [ ] 5.4 配置"普通用户管理"菜单的权限码（identity:regular-user:list, identity:regular-user:update）
- [ ] 5.5 将新菜单项关联到合适的父菜单（用户管理模块）

## 6. Testing and Verification

- [ ] 6.1 测试管理员管理页面：列表加载、搜索、分页
- [ ] 6.2 测试管理员管理页面：编辑功能、表单验证
- [ ] 6.3 测试普通用户管理页面：列表加载、搜索、分页
- [ ] 6.4 测试普通用户管理页面：编辑功能、表单验证
- [ ] 6.5 测试权限控制：无权限用户访问被拒绝
- [ ] 6.6 测试权限控制：无编辑权限时编辑按钮被隐藏
- [ ] 6.7 验证API向后兼容性（不传role_type参数时返回所有用户）
- [ ] 6.8 检查现有用户管理页面是否需要保留或重定向
