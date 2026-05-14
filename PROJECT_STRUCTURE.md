# 社区与家园管理系统 - 项目结构

## 项目概览
基于 Go + Vue3 的微服务架构社区管理系统

## 根目录结构

```
community-and-home/
├── common/              # 公共库
├── services/            # 微服务
│   ├── identity/       # 身份认证服务
│   ├── masterdata/     # 主数据服务
│   ├── ai-model/       # AI模型统一管理服务
│   └── moderation/     # 内容审核服务
├── web/                 # 前端应用
├── gateway/             # API网关
├── openspec/            # OpenSpec提案和任务管理
├── docs/                # 文档
├── specs/               # 需求规格说明
├── scripts/             # 脚本工具
├── deploy/              # 部署配置
├── mysql/               # MySQL数据目录
├── redis/               # Redis数据目录
├── minio/               # MinIO对象存储数据
├── etcd/                # etcd配置中心数据
├── Sensitive-lexicon/   # 敏感词库资源
├── model/               # 数据模型（废弃，已迁移到各服务内）
├── .claude/             # Claude Code配置和缓存
└── .specify/            # Specify工具配置
```

---

## 详细目录说明

### 1. common/ - 公共库

**功能模块:**
- `errorx/` - 错误处理封装
- `jwtx/` - JWT认证工具
- `miniox/` - MinIO对象存储客户端
- `responsex/` - 统一响应格式

---

### 2. services/ - 微服务

#### 2.1 services/identity/ - 身份认证服务

**服务端口:**
- API: 8888
- RPC: 8080

**主要功能:**
- 用户认证（登录/注册）
- 用户管理（CRUD）
- 权限验证
- 物业绑定验证

**目录结构:**
```
identity/
├── api/                    # HTTP API服务
│   ├── etc/               # 配置文件
│   │   └── identity-api.yaml
│   ├── internal/
│   │   ├── config/       # 配置结构
│   │   ├── handler/      # HTTP处理器
│   │   │   ├── auth/    # 认证相关（登录/注册）
│   │   │   └── user/    # 用户管理（CRUD）
│   │   ├── logic/        # 业务逻辑
│   │   ├── svc/          # 服务上下文
│   │   └── types/        # 类型定义
│   ├── identity.go       # 主程序入口
│   └── identity.api      # API定义文件
├── model/                  # 数据模型
│   ├── idUserModel.go
│   ├── idRoleModel.go
│   ├── idPermissionModel.go
│   └── idPropertyBindingModel.go
└── rpc/                    # gRPC服务
    ├── etc/
    │   └── identity.yaml
    ├── internal/
    │   ├── config/
    │   ├── logic/
    │   └── svc/
    ├── pb/                # Protobuf生成代码
    ├── identity.go        # RPC主程序
    └── identity.proto     # Protobuf定义
```

#### 2.2 services/masterdata/ - 主数据服务

**服务端口:**
- API: 8889
- RPC: 8081

**主要功能:**
- 行政区划管理
- 社区管理
- 敏感词管理
- 系统配置管理

**目录结构:**
```
masterdata/
├── api/
│   ├── etc/
│   │   └── masterdata-api.yaml
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── division/      # 行政区划
│   │   │   ├── community/     # 社区管理
│   │   │   ├── sensitiveword/ # 敏感词
│   │   │   └── configuration/ # 系统配置
│   │   ├── logic/
│   │   ├── svc/
│   │   └── types/
│   ├── masterdata.go
│   └── masterdata.api
├── model/
│   ├── mdAdministrativeDivisionModel.go
│   ├── mdResidentialAreaModel.go
│   ├── mdSensitiveWordModel.go
│   └── mdConfigurationModel.go
└── rpc/
    ├── etc/
    ├── internal/
    └── masterdata.proto
```

#### 2.3 services/ai-model/ - AI模型统一管理服务

**服务端口:**
- API: 8891
- RPC: 8084

**主要功能:**
- 多模型统一调用（Claude、GPT、Ollama等）
- 模型配置管理
- API密钥管理
- 提示词模板管理
- 调用日志和成本统计
- 健康检查和监控

**目录结构:**
```
ai-model/
├── api/                    # HTTP API服务
│   ├── etc/
│   │   └── aimodelapi.yaml
│   ├── internal/
│   │   ├── handler/
│   │   ├── logic/
│   │   └── svc/
│   └── aimodelapi.go
├── rpc/                    # gRPC服务
│   ├── etc/
│   │   └── aiModel.yaml
│   ├── internal/
│   │   ├── adapter/       # 模型适配器（Claude、OpenAI、Ollama）
│   │   ├── logic/
│   │   ├── manager/       # 业务管理器
│   │   └── svc/
│   ├── model/             # 数据模型
│   ├── pb/                # Protobuf生成代码
│   └── aiModel.go
├── python-engine/         # Python AI引擎（可选）
├── sql/                   # 数据库脚本
└── docker-compose.yml
```

**数据库:** ai_model_db

#### 2.4 services/moderation/ - 内容审核服务

**服务端口:**
- API: 8892

**主要功能:**
- 文本内容审核
- 图片内容审核
- 敏感词检测
- AI辅助审核

**目录结构:**
```
moderation/
├── api/                   # HTTP API服务
│   ├── etc/
│   ├── internal/
│   └── moderation.go
├── model/                 # 数据模型
├── migrations/            # 数据库迁移
└── internal/              # 内部逻辑
```

---

### 3. web/ - 前端应用

#### 3.1 web/pc/ - PC管理后台

**技术栈:**
- Vue 3.4+
- TypeScript 5.0+
- Element Plus
- Vite 6.0+
- Pinia (状态管理)
- Vue Router

**端口:** 3003

**目录结构:**
```
pc/
├── src/
│   ├── api/              # API接口定义
│   │   ├── auth.ts      # 认证接口
│   │   ├── user.ts      # 用户接口
│   │   └── masterdata.ts # 主数据接口
│   ├── assets/           # 静态资源
│   ├── components/       # 公共组件
│   │   ├── business/    # 业务组件
│   │   ├── common/      # 通用组件
│   │   └── layout/      # 布局组件
│   ├── router/           # 路由配置
│   │   └── index.ts
│   ├── stores/           # Pinia状态管理
│   │   ├── auth.ts      # 认证状态
│   │   └── user.ts      # 用户状态
│   ├── utils/            # 工具函数
│   │   └── request.ts   # Axios封装
│   ├── views/            # 页面视图
│   │   ├── auth/        # 登录/注册
│   │   ├── dashboard/   # 仪表盘
│   │   ├── users/       # 用户管理
│   │   ├── division/    # 行政区划
│   │   ├── communities/ # 社区管理
│   │   ├── sensitive-words/ # 敏感词
│   │   └── config/      # 系统配置
│   ├── App.vue
│   └── main.ts
├── .env.development      # 开发环境配置
├── .env.production       # 生产环境配置
├── vite.config.ts
├── tsconfig.json
└── package.json
```

#### 3.2 web/mobile/ - 移动端应用
（待开发）

#### 3.3 web/common/ - 前端公共库
```
common/
├── api/          # 共享API定义
├── constants/    # 常量定义
├── types/        # TypeScript类型
└── utils/        # 工具函数
```

---

### 4. gateway/ - API网关

**端口:** 8080

**功能:**
- 统一API入口
- 路由转发
- 负载均衡（未来）

**路由规则:**
```
/api/identity/*   → localhost:8888  (Identity服务)
/api/masterdata/* → localhost:8889  (Masterdata服务)
/api/ai-model/*   → localhost:8891  (AI-Model服务)
/api/moderation/* → localhost:8892  (Moderation服务)
/health           → 健康检查
```

**文件:**
```
gateway/
└── main.go       # 网关主程序（Go HTTP反向代理）
```

---

### 5. 数据存储

#### 5.1 mysql/ - MySQL数据库
```
mysql/
├── conf/         # MySQL配置
│   └── my.cnf
└── data/         # 数据文件
    ├── identity_db/      # 身份认证数据库
    └── masterdata_db/    # 主数据数据库
```

**数据库表:**
- **identity_db**: 
  - id_user (用户表)
  - id_role (角色表)
  - id_permission (权限表)
  - id_property_binding (物业绑定表)

- **masterdata_db**:
  - md_administrative_division (行政区划表)
  - md_residential_area (住宅小区表)
  - md_sensitive_word (敏感词表)
  - md_configuration (系统配置表)

- **ai_model_db**:
  - am_model_config (模型配置表)
  - am_api_key (API密钥表)
  - am_prompt_template (提示词模板表)
  - am_call_log (调用日志表)
  - am_usage_statistics (使用统计表)
  - am_health_check (健康检查表)

#### 5.2 redis/ - Redis缓存
```
redis/
└── data/         # Redis数据文件
```

**用途:**
- 数据模型缓存（go-zero cache）
- Session存储
- 分布式锁

#### 5.3 minio/ - 对象存储
```
minio/
└── data/         # MinIO数据文件
```

**用途:**
- 用户头像
- 社区图片
- 文件上传

#### 5.4 etcd/ - 配置中心
```
etcd/
└── data/         # etcd数据文件
```

**用途:**
- 服务发现
- 配置管理

---

### 6. 其他目录

#### 6.1 openspec/ - OpenSpec提案管理
```
openspec/
├── config.yaml          # OpenSpec配置文件
├── changes/             # 变更提案目录
│   ├── ai-model-service/
│   │   ├── proposal.md
│   │   ├── tasks.md
│   │   ├── design.md
│   │   └── specs/
│   ├── ai-model-api-testing-fixes/
│   └── ...
└── specs/               # 规格说明
```

**用途:**
- 管理功能提案和变更
- 跟踪开发任务
- 记录设计决策

#### 6.2 Sensitive-lexicon/ - 敏感词库资源
```
Sensitive-lexicon/
└── data/                # 敏感词数据文件
```

**用途:**
- 存储敏感词库资源
- 供敏感词管理功能使用

#### 6.3 .claude/ - Claude Code配置
```
.claude/
├── projects/            # 项目特定配置
├── worktrees/          # 工作树缓存
└── scheduled_tasks.json # 定时任务
```

**用途:**
- Claude Code AI助手的配置和缓存
- 不应提交到版本控制

#### 6.4 specs/ - 需求规格
```
specs/
├── 001-identity-masterdata/    # Phase 1规格
│   ├── checklists/
│   └── contracts/
└── 002-web-pc-admin/           # Phase 2规格
    ├── checklists/
    └── contracts/
```

#### 6.5 scripts/ - 脚本工具
```
scripts/
└── sql/          # SQL脚本
    ├── identity_db.sql
    └── masterdata_db.sql
```

#### 6.6 docs/ - 文档
```
docs/
└── api/          # API文档
```

#### 6.7 deploy/ - 部署配置
```
deploy/
├── docker-compose.yml
└── kubernetes/
```

---

## 核心文件说明

### 配置文件
- `services/identity/api/etc/identity-api.yaml` - Identity API配置
- `services/masterdata/api/etc/masterdata-api.yaml` - Masterdata API配置
- `services/ai-model/api/etc/aimodelapi.yaml` - AI-Model API配置
- `services/ai-model/rpc/etc/aiModel.yaml` - AI-Model RPC配置
- `web/pc/.env.development` - 前端开发环境配置
- `openspec/config.yaml` - OpenSpec配置
- `CLAUDE.md` - Claude AI工作指南
- `DEPLOYMENT.md` - 部署文档
- `PROJECT_STRUCTURE.md` - 本文档

### API定义
- `services/identity/api/identity.api` - Identity HTTP API定义
- `services/identity/rpc/identity.proto` - Identity gRPC定义
- `services/masterdata/api/masterdata.api` - Masterdata HTTP API定义
- `services/ai-model/rpc/pb/ai_model.proto` - AI-Model gRPC定义

### 数据模型
- `services/identity/model/*Model.go` - Identity数据模型
- `services/masterdata/model/*Model.go` - Masterdata数据模型

---

## 请求流程示例

### 用户登录流程
```
1. 浏览器访问: http://172.31.39.71:3003
2. 前端发起登录: POST http://172.31.39.71:8080/api/identity/auth/login
3. API网关转发: POST http://localhost:8888/api/identity/auth/login
4. Identity API处理:
   - Handler: services/identity/api/internal/handler/auth/login_handler.go
   - Logic: services/identity/api/internal/logic/auth/login_logic.go
   - Model: services/identity/model/idUserModel.go
5. 查询数据库: identity_db.id_user
6. 返回JWT Token
```

### 行政区划查询流程
```
1. 前端请求: GET http://172.31.39.71:8080/api/masterdata/divisions
2. API网关转发: GET http://localhost:8889/api/masterdata/divisions
3. Masterdata API处理:
   - Handler: services/masterdata/api/internal/handler/division/getDivisionsHandler.go
   - Logic: services/masterdata/api/internal/logic/division/getDivisionsLogic.go
   - Model: services/masterdata/model/mdAdministrativeDivisionModel.go
4. 查询缓存: Redis
5. 缓存未命中则查询: masterdata_db.md_administrative_division
6. 返回JSON数据
```

---

## 服务启动顺序

### 1. 基础设施
```bash
# MySQL, Redis, MinIO, etcd
docker-compose up -d
```

### 2. 后端服务
```bash
# Identity RPC (端口8080)
cd services/identity/rpc && go run identity.go

# Identity API (端口8888)
cd services/identity/api && go run identity.go

# Masterdata API (端口8889)
cd services/masterdata/api && go run masterdata.go
```

### 3. API网关
```bash
# API Gateway (端口8080)
cd gateway && go run main.go
```

### 4. 前端
```bash
# PC管理后台 (端口3003)
cd web/pc && npm run dev
```

---

## 访问地址

- **前端管理后台**: http://172.31.39.71:3003
- **API网关**: http://172.31.39.71:8080
- **Identity API**: http://localhost:8888
- **Masterdata API**: http://localhost:8889
- **AI-Model API**: http://localhost:8891
- **Moderation API**: http://localhost:8892
- **Identity RPC**: http://localhost:8080
- **AI-Model RPC**: http://localhost:8084

---

## 技术栈总结

### 后端
- **语言**: Go 1.21+
- **框架**: go-zero 1.6+
- **数据库**: MySQL 8.0
- **缓存**: Redis 7.0
- **对象存储**: MinIO
- **配置中心**: etcd
- **通信协议**: HTTP/REST, gRPC

### 前端
- **语言**: TypeScript 5.0+
- **框架**: Vue 3.4+
- **UI库**: Element Plus
- **构建工具**: Vite 6.0+
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP客户端**: Axios

### 架构模式
- 微服务架构
- DDD领域驱动设计
- RESTful API
- JWT认证
- 缓存优先策略
- API网关模式

---

## 开发规范

### 后端代码规范
- 使用go-zero框架的标准目录结构
- Handler层只负责参数验证和调用Logic
- Logic层实现业务逻辑
- Model层负责数据访问
- 统一使用responsex包返回响应
- 统一使用errorx包处理错误

### 前端代码规范
- 使用TypeScript严格模式
- 组件使用Composition API
- API调用统一通过api目录
- 状态管理使用Pinia
- 路由配置集中管理
- 样式使用SCSS

### 数据库规范
- 表名使用下划线命名法
- 主键统一使用id
- 时间字段: created_time, updated_time, delete_time
- 软删除: delete_time IS NULL
- 状态字段: status (1=启用, 0=禁用)

---

## 常用命令

### 后端开发
```bash
# 生成API代码
goctl api go -api identity.api -dir .

# 生成Model代码
goctl model mysql datasource -url="user:pass@tcp(127.0.0.1:3306)/db" -table="table_name" -dir="./model"

# 生成RPC代码
goctl rpc protoc identity.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

### 前端开发
```bash
# 安装依赖
npm install

# 开发模式
npm run dev

# 构建生产版本
npm run build

# 类型检查
npm run type-check

# 代码格式化
npm run format
```

### 数据库操作
```bash
# 导入SQL
mysql -u root -p identity_db < scripts/sql/identity_db.sql

# 备份数据库
mysqldump -u root -p identity_db > backup.sql
```

---

## 故障排查

### 后端服务无法启动
1. 检查端口是否被占用: `lsof -i:8888`
2. 检查MySQL/Redis是否运行: `docker ps`
3. 检查配置文件路径是否正确
4. 查看服务日志

### 前端无法访问API
1. 检查API网关是否运行: `curl http://172.31.39.71:8080/health`
2. 检查.env.development配置
3. 检查浏览器控制台Network标签
4. 检查CORS配置

### 数据库连接失败
1. 检查MySQL容器状态
2. 检查数据库用户权限
3. 检查配置文件中的数据库连接信息
4. 测试连接: `mysql -h localhost -u root -p`

---

## 相关文档

- [CLAUDE.md](./CLAUDE.md) - Claude AI工作指南
- [DEPLOYMENT.md](./DEPLOYMENT.md) - 部署文档
- [specs/](./specs/) - 需求规格说明
