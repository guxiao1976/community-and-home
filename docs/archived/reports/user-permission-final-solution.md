# 用户权限问题最终报告

## 问题
用户 18562955889 登录后看不到角色、权限管理菜单

---

## 🔍 根本原因

### 问题1: user_base 表没有 user_type 字段

**错误信息**: `Unknown column 'user_type' in 'field list'`

**原因**: 
- 当前数据库表结构中没有 `user_type` 字段
- 系统可能采用纯 RBAC 模型（基于角色的访问控制）
- 不通过用户类型区分管理员，而是通过角色和权限

### 问题2: permission-service 未启动

**影响**:
- 无法获取用户权限列表
- 前端菜单权限检查失败
- 所有需要权限的菜单都不显示

---

## 💡 解决方案

### 方案1: 启动 permission-service + 分配权限（推荐）

这是标准的 RBAC 权限模型解决方案：

**步骤1: 启动 permission-service**

```bash
# 启动 RPC 服务
cd services/permission-service/rpc
export MYSQL_USER=root
export MYSQL_PASSWORD=root123456
export REDIS_PASSWORD=123456
nohup go run permissionservice.go -f etc/permissionservice.yaml > /tmp/microservices-logs/permission-service.log 2>&1 &

# 启动 API 服务（注意端口冲突）
cd services/permission-service/api
export MYSQL_USER=root
export MYSQL_PASSWORD=root123456
export REDIS_PASSWORD=123456
export JWT_ACCESS_SECRET=your-secret-key
nohup go run perm.go -f etc/perm-api.yaml > /tmp/microservices-logs/permission-service-api.log 2>&1 &
```

**步骤2: 初始化权限数据**

创建基础角色和权限：

```sql
-- 创建超级管理员角色
INSERT INTO permission.perm_role (role_code, role_name, description, is_system, status) 
VALUES ('super_admin', '超级管理员', '拥有所有权限', 1, 1);

-- 创建角色管理权限
INSERT INTO permission.perm_permission (permission_code, permission_name, resource_type) 
VALUES 
('role:view', '查看角色', 'menu'),
('role:create', '创建角色', 'action'),
('role:update', '编辑角色', 'action'),
('role:delete', '删除角色', 'action'),
('role:permission', '配置角色权限', 'action');

-- 分配权限给角色
-- (需要先查询 role_id 和 permission_id)

-- 将角色分配给用户
-- (需要先查询用户 user_id)
```

**步骤3: 给用户分配超级管理员角色**

通过前端界面或API：
1. 登录有权限的管理员账号
2. 进入用户管理
3. 编辑用户 18562955889
4. 分配"超级管理员"角色

---

### 方案2: 添加 user_type 字段（快速方案）

如果想简单快速解决，可以添加用户类型字段：

```sql
-- 添加 user_type 字段
ALTER TABLE user.user_base 
ADD COLUMN user_type TINYINT NOT NULL DEFAULT 1 
COMMENT '用户类型：1-普通用户 2-管理员' 
AFTER id;

-- 将该用户设为管理员
UPDATE user.user_base 
SET user_type = 2 
WHERE phone = '18562955889';
```

**前端代码需要修改**:
```typescript
// 在登录后检查 user_type
if (userInfo.user_type === 2) {
  // 管理员，显示所有菜单
  showAllMenus();
} else {
  // 普通用户，按权限过滤
  filterMenusByPermissions();
}
```

---

### 方案3: 临时禁用权限检查（开发环境）

**仅用于开发测试**，不要在生产环境使用：

修改 `config/modules/identity.config.ts`：

```typescript
{
  path: '/roles',
  title: '角色管理',
  icon: UserFilled,
  // permission: 'role:view'  // 临时注释掉权限检查
}
```

---

## 🎯 推荐操作步骤

### 完整方案（推荐）

**步骤1: 检查 permission 数据库表结构**
```bash
docker exec mysql mysql -uroot -proot123456 -e "USE permission; SHOW TABLES;"
```

**步骤2: 如果有完整的权限表，启动 permission-service**

**步骤3: 初始化基础数据**
- 创建超级管理员角色
- 创建基础权限
- 将角色分配给用户

**步骤4: 用户重新登录**
- 退出登录
- 重新登录
- 刷新页面

---

### 快速方案（临时）

**添加 user_type 字段并设为管理员**：

```bash
docker exec mysql mysql -uroot -proot123456 << 'SQL'
USE user;

-- 添加字段
ALTER TABLE user_base 
ADD COLUMN user_type TINYINT NOT NULL DEFAULT 1 
COMMENT '用户类型：1-普通用户 2-管理员';

-- 设为管理员
UPDATE user_base 
SET user_type = 2 
WHERE phone = '18562955889';

-- 验证
SELECT phone, nickname, user_type FROM user_base WHERE phone = '18562955889';
SQL
```

---

## 📊 当前系统状态

### permission 数据库

- ✅ 数据库存在
- ❓ 表结构未知（需要查看）
- ❌ permission-service 未启动

### user 数据库

- ✅ user_base 表存在
- ❌ 没有 user_type 字段
- ✅ 用户 18562955889 存在

### 前端菜单

- ✅ 菜单配置完整
- ⚠️ 需要权限: `role:view`
- ❌ 用户没有该权限

---

## 🔧 为什么看不到菜单？

### 权限检查流程

```
用户登录
  ↓
获取用户信息
  ↓
调用 permission-service 获取权限列表
  ↓ (失败：服务未启动)
权限列表为空
  ↓
前端过滤菜单
  ↓
菜单项需要 role:view 权限
  ↓
用户没有该权限
  ↓
菜单不显示 ✗
```

---

## ✅ 最快解决方案

### 选择方案2：添加 user_type 字段

**原因**:
1. 最快速
2. 不依赖 permission-service
3. 适合开发测试阶段

**执行**:
```bash
docker exec mysql mysql -uroot -proot123456 -e "
USE user;
ALTER TABLE user_base 
ADD COLUMN user_type TINYINT NOT NULL DEFAULT 1 
COMMENT '用户类型：1-普通用户 2-管理员';

UPDATE user_base 
SET user_type = 2 
WHERE phone = '18562955889';
"
```

**验证**:
```bash
docker exec mysql mysql -uroot -proot123456 -e "
SELECT phone, nickname, user_type 
FROM user.user_base 
WHERE phone = '18562955889';
"
```

**前端重新登录后应该能看到菜单**

---

## 📝 后续完善建议

1. **启动 permission-service**
2. **建立完整的 RBAC 权限体系**
3. **创建基础角色**（超级管理员、管理员、普通用户）
4. **创建基础权限**（菜单权限、操作权限）
5. **通过前端界面管理角色和权限**

---

**报告时间**: 2026-07-12 21:00  
**建议**: 立即执行添加 user_type 字段的 SQL
