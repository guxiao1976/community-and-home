# 社区与家园管理系统 - 部署文档

## 系统架构

```
Frontend (Vue3)          API Gateway           Backend Services
http://172.31.39.71:3000 ──> :8080 ──┬──> Identity API :8888
                                     └──> Masterdata API :8889
```

## 服务列表

| 服务名称 | 端口 | 状态 | 说明 |
|---------|------|------|------|
| Frontend (Vue3) | 3000 | ✓ Running | 前端界面 |
| API Gateway | 8080 | ✓ Running | 统一API入口 |
| Identity API | 8888 | ✓ Running | 用户认证服务 |
| Identity RPC | 8080 | ✓ Running | 用户RPC服务 |
| Masterdata API | 8889 | ✓ Running | 主数据服务 |

## 访问地址

- **前端界面**: http://172.31.39.71:3000
- **API网关**: http://172.31.39.71:8080
- **API文档**: 
  - Identity: http://172.31.39.71:8888/swagger
  - Masterdata: http://172.31.39.71:8889/swagger

## API网关配置

API网关负责将前端请求路由到不同的后端服务：

- `/api/identity/*` → Identity API (8888)
- `/api/masterdata/*` → Masterdata API (8889)

网关自动处理：
- CORS跨域请求
- 请求转发
- 统一入口管理

## 测试结果

所有核心功能已测试通过：

### 1. 用户认证模块 ✓
- [x] 用户登录
- [x] Token认证
- [x] 用户列表查询

### 2. 主数据管理模块 ✓
- [x] 行政区划管理（创建、查询、分页）
- [x] 社区管理（创建、查询、过滤）
- [x] 系统配置管理
- [x] 敏感词管理

## 启动服务

### 1. 启动后端服务

```bash
# Identity API
cd services/identity/api
./identity-api &

# Identity RPC
cd services/identity/rpc
./identity &

# Masterdata API
cd services/masterdata/api
./masterdata-api &
```

### 2. 启动API网关

```bash
cd gateway
go run main.go &
```

### 3. 启动前端

```bash
cd web/pc
npm run dev
```

## 环境配置

### 前端配置 (web/pc/.env.development)
```
VITE_API_BASE_URL=http://172.31.39.71:8080
```

### Identity API配置 (services/identity/api/etc/identity-api.yaml)
- 端口: 8888
- 数据库: MySQL
- Redis缓存

### Masterdata API配置 (services/masterdata/api/etc/masterdata-api.yaml)
- 端口: 8889
- 数据库: MySQL
- Redis缓存

## 数据库

### 数据库名称
- `community_identity` - 用户认证数据
- `community_masterdata` - 主数据

### 测试数据
系统已包含测试数据：
- 管理员账号: phone=13800000000, password=Admin@123456
- 测试用户: phone=13900000001, password=Test@123456
- 行政区划: 5条测试数据
- 社区: 4条测试数据
- 系统配置: 8条配置项
- 敏感词: 3条测试数据

## 故障排查

### 前端无法访问API
1. 检查API网关是否运行: `curl http://localhost:8080/health`
2. 检查后端服务是否运行: `ps aux | grep -E "identity-api|masterdata-api"`
3. 检查前端配置: `cat web/pc/.env.development`

### API返回404
1. 确认请求路径是否正确（必须包含 `/api/identity/` 或 `/api/masterdata/`）
2. 检查网关日志
3. 直接访问后端服务测试

### 数据库连接失败
1. 检查MySQL服务状态
2. 验证数据库配置文件中的连接信息
3. 确认数据库用户权限

## 已修复的问题

1. ✓ 行政区划创建失败 - 修复datetime字段处理
2. ✓ 社区管理404错误 - 实现业务逻辑
3. ✓ 敏感词管理404错误 - 已正常工作
4. ✓ 系统配置404错误 - 已正常工作
5. ✓ 用户列表400错误 - 已正常工作
6. ✓ 前端无法同时访问两个服务 - 添加API网关

## 下一步工作

1. 实现其他待完成的业务逻辑（更新、删除等）
2. 添加更多的单元测试和集成测试
3. 完善错误处理和日志记录
4. 优化性能和缓存策略
5. 添加API文档（Swagger）
6. 配置生产环境部署

## 联系方式

如有问题，请查看：
- 项目文档: /home/jiaoxh/my-code/community-and-home/README.md
- 规范文档: /home/jiaoxh/my-code/community-and-home/.specify/
