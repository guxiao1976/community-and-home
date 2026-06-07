# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 角色定位

这是 **全局架构协调层**。本 Claude 实例的职责是：

- **api-proto 管理**：所有服务接口的 Proto 定义、代码生成、破坏性变更检测
- **跨服务功能协调**：涉及多个微服务的大型功能开发、服务间接口对齐
- **全局规范定义**：错误码规范、响应格式、公共库版本策略
- **架构决策**：服务拆分/合并、技术选型、依赖方向

**本实例不负责单个服务的具体实现**。当需要开发具体功能时，切换到对应服务的子 Claude 实例（见下方索引）。

开发流程和工具选择标准见 [docs/specs/](docs/specs/)：
- [AI 团队工作规则](docs/specs/ai-dev-team-design.md) — 五层架构、角色定义、完整流水线
- [工具选择标准](docs/specs/tool-selection-standard.md) — Bug→Edit / 小功能→Agent / 大功能→OpenSpec+Workflow / 批量→Ralph
- [执行日志](docs/specs/execution-log.md) — 每次任务的流程遵守情况和复盘

## 子 Claude 实例索引

| 中文名 | 目录 | CLAUDE.md | 职责范围 |
|--------|------|-----------|---------|
| 用户服务 | `services/user-service/` | [CLAUDE.md](services/user-service/CLAUDE.md) | 用户服务（API+RPC 双层） |
| 认证服务 | `services/auth-service/` | [CLAUDE.md](services/auth-service/CLAUDE.md) | 认证服务（API+RPC 双层，AT+RT 双 Token） |
| 权限服务 | `services/permission-service/` | [CLAUDE.md](services/permission-service/CLAUDE.md) | 权限服务（API+RPC 双层，RBAC） |
| 主数据服务 | `services/master-data-service/` | [CLAUDE.md](services/master-data-service/CLAUDE.md) | 主数据服务（API+RPC+Cron） |
| 审核服务 | `services/moderation-service/` | [CLAUDE.md](services/moderation-service/CLAUDE.md) | 内容审核服务（API+RPC 双层） |
| AI模型服务 | `services/ai-model-service/` | [CLAUDE.md](services/ai-model-service/CLAUDE.md) | AI 模型服务（Go+Python 混合） |
| 文件服务 | `services/file-service/` | [CLAUDE.md](services/file-service/CLAUDE.md) | 文件服务（MinIO 上传/下载） |
| `api-proto/` | `api-proto/` | [CLAUDE.md](api-proto/CLAUDE.md) | API 契约定义、Proto 代码生成 |
| `common/` | `common/` | [CLAUDE.md](common/CLAUDE.md) | Go 共享库（v2），10 个工具包 |
| 前端 | `web/` | [web/CLAUDE.md](web/CLAUDE.md) | 前端层入口 |
| 管理后台 | `web/pc/` | [web/pc/CLAUDE.md](web/pc/CLAUDE.md) | Vue 3 管理后台（详细） |
| 移动端 | `web/mobile/` | [web/mobile/CLAUDE.md](web/mobile/CLAUDE.md) | Uni-app (Vue 3) 移动端 |
| 社区枢纽服务 | `services/community-hub-service/` | [CLAUDE.md](services/community-hub-service/CLAUDE.md) | 社区枢纽服务（API+RPC 双层，通知/联络/寻失） |

## 文档约定

### 每个服务必须标配的文件

子 Claude 是无状态的——每次启动只看到服务目录下的文件。这些文件决定了它能写出多正确的代码：

| 文件 | 优先级 | 作用 | 谁写 | 谁维护 |
|------|:---:|------|------|------|
| `CLAUDE.md` | 必须 | 启动上下文：角色定位、关键规则、全局公约引用、常用命令 | 全局 Claude 初始化 | 子 Claude 维护规则，全局 Claude 维护公约 |
| `docs/design.md` | 必须 | 设计决策：为什么这样设计、数据模型（DDL+ER）、业务流程、接口契约、缓存策略、安全机制 | 全局 Claude 写 | 子 Claude 实现后更新"实际与设计差异" |
| `CHANGELOG.md` | 必须 | 变更历史：每个 PR 记录做了什么、为什么、影响范围、关联 commit | 子 Claude | 子 Claude |
| `.claude/memory/` | 推荐 | 持久记忆：用户偏好、踩过的坑、技术决策原因、`[[链接]]` 到其他 decisions | 子 Claude 写 | 子 Claude |

### 文件位置

- **服务级文档** → `services/<name>/`（随服务代码存放，子 Claude 本地可访问）
- **跨服务规范** → `docs/specs/`（服务间交互协议、联调测试计划等全局性文档）
- 根 `docs/specs/` 中原有的服务级设计文档已迁移，保留跳转指针避免旧链接失效

### CHANGELOG.md 模板

```markdown
# CHANGELOG — <service-name>

## YYYY-MM-DD — <简短标题>

### 做了什么
- <变更1>
- <变更2>

### 为什么
<一句话原因>

### 影响
- Proto: <有无 api-proto 变更>
- 调用方: <影响哪些服务>
- 数据库: <有无 migration>
- 关联: <PR/issue 链接>
```

## Proto 管理规范（最高优先级）

这是本项目的**硬性架构约束**，所有服务必须遵守，无一例外。

### 1. 统一存放

**所有微服务的 gRPC 接口定义必须放在 `api-proto/` 中**，不得在服务目录中私自定义 proto。

```
api-proto/api/
  auth/v1/auth.proto              # AuthService
  user/v1/user.proto              # UserService
  permission/v1/permission.proto  # PermissionService
  aimodel/v1/ai_model.proto       # AiModelService
  common/v1/common.proto          # 共享类型
  file/v1/file.proto              # FileService
```

- ✅ 所有 Proto 定义 → `api-proto/`
- ❌ 服务本地 proto → 禁止（`rpc/pb/` 只能放跳转指针）
- ❌ 服务间直接用 HTTP 调用 → 必须走 gRPC

### 2. 变更流程

```
1. 修改 api-proto/api/<service>/v1/*.proto         ← 在此编辑
2. cd api-proto && make generate                    ← 生成 Go 代码
3. cd api-proto && make lint                        ← 规范检查
4. cd api-proto && make breaking-check              ← 破坏性变更检测
5. 记录到 api-proto/CHANGELOG.md                    ← 变更日志
6. 通知受影响的服务重新构建                            ← 影响评估
```

### 3. 权限边界

| 角色 | 可以做什么 | 不能做什么 |
|------|-----------|-----------|
| **全局 Claude**（本实例） | 修改 api-proto、评估影响、通知服务 | 不写服务具体代码 |
| **子 Claude**（服务实例） | 使用 api-proto 生成的代码 | 禁止修改 api-proto/；需要时告知用户切回全局 Claude |

### 4. 子 Claude 必须知道的公约

每个服务的 `CLAUDE.md` **必须包含**以下标准章节，确保子 Claude 感知全局规则：

```markdown
## 全局公约

本项目所有微服务遵守统一的架构规范，详见根 [CLAUDE.md](../../CLAUDE.md)。

与本服务相关的关键约束：
- **Proto 定义在 api-proto/**：本服务的 gRPC 接口定义在 `api-proto/api/<xxx>/v1/`，修改 proto 需告知用户切换到全局 Claude
- **服务间通信仅通过 gRPC**：调用其他服务必须走 gRPC（etcd 服务发现），禁止直连数据库
- **设计文档在 docs/design.md**：数据库、业务流程、接口设计见 [docs/design.md](docs/design.md)
- **变更记录在 CHANGELOG.md**：每次变更必须更新 [CHANGELOG.md](CHANGELOG.md)
- **提交前运行机械化检查**：代码变更提交前必须运行 `bash scripts/harness-checks.sh --service <服务目录名>`，有 FAIL 则不可提交
```

### 子 Claude 必需文件清单

| # | 文件 | 子 Claude 开发时 |
|----|------|---------|
| 1 | `CLAUDE.md` | 启动即读：知道我是谁、规则、怎么构建 |
| 2 | `docs/design.md` | 写代码前读：理解数据模型和业务流程 |
| 3 | `CHANGELOG.md` | 改完后写：留痕，下次启动知道历史 |
| 4 | `.claude/memory/` | 持续积累：偏好、坑、决策原因 |

### 5. 待迁移清单

以下服务尚未将 proto 迁移到 api-proto（需后续完成）：

| 服务 | 当前状态 | Proto 位置 |
|------|:---:|------|
| master-data-service | ✅ 已迁移 | `api-proto/api/masterdata/v1/` |
| moderation-service | ✅ Proto 已建立（gRPC 层待实现） | `api-proto/api/moderation/v1/` |
| ai-model-service | ✅ 已迁移 | `api-proto/api/aimodel/v1/` |
| user-service | ✅ | `api-proto/api/user/v1/` |
| auth-service | ✅ | `api-proto/api/auth/v1/` |
| permission-service | ✅ | `api-proto/api/permission/v1/` |
| file-service | ✅ | `api-proto/api/file/v1/` |

## 全局硬规则

这些规则适用于所有子 Claude 实例，也适用于本实例：

1. **所有服务间通信必须通过 gRPC 接口**，接口定义在 `api-proto/` 中。禁止直接访问其他服务的数据库。
2. **Proto 变更必须在本全局实例中执行**。子 Claude 实例禁止修改 `api-proto/`。如需修改 proto，子实例应告知用户切换到全局 Claude。
3. **修改 `common/` 会影响所有依赖服务**。涉及 common 的变更需要本全局实例评估影响范围。
4. **go.work 联调规则**：本地开发通过 `go.work` 解析模块依赖，修改 common 或 api-proto 后各服务立即可见新代码，无需远程发布。
5. **Snowflake ID 序列化规范**（JavaScript 安全整数兼容）：
   - 所有 Proto `int64` ID 字段必须添加 `[jstype = JS_STRING]`，确保 protojson 序列化时以字符串输出
   - 所有 REST API 类型中 `int64` ID 字段必须使用 `json:"...,string"` 标签
   - 前端所有 ID 字段 TypeScript 类型必须为 `string`，axios 使用 `lossless-json` 解析器
   - 原因：Snowflake 19 位 ID 超过 JS `Number.MAX_SAFE_INTEGER`（~16 位），JSON 数字解析时精度丢失
6. **代码变更提交前必须通过机械化检查**：
   - 任何 `services/<name>/` 下的 Go 代码变更，提交前必须运行 `bash scripts/harness-checks.sh --service <name>`
   - 有 FAIL → 不能提交，必须先修复
   - 此规则适用于全局 Claude 和所有子 Claude 实例，无一例外

## Proto 变更工作流

当需要新增或修改 gRPC 接口时，按以下流程操作：

```bash
# 1. 编辑 proto 文件（api/auth/v1/*.proto 等）
# 2. 生成所有语言的代码
cd api-proto && make generate

# 3. Lint 检查
cd api-proto && make lint

# 4. 如涉及破坏性变更，运行 breaking check
cd api-proto && make breaking-check

# 5. 记录变更到 api-proto/CHANGELOG.md
# 6. 告知用户通知所有依赖该 proto 的服务更新代码
```

受影响的服务（根据 proto 使用情况）：
- 修改 `auth/v1` → 通知 `auth-service`
- 修改 `user/v1` → 通知 `user-service`, `auth-service`（auth 调用 user）
- 修改 `permission/v1` → 通知 `permission-service`
- 修改 `aimodel/v1` → 通知 `ai-model-service`，以及调用方 `moderation-service`
- 修改 `common/v1` → 通知所有使用 api-proto 的服务
- 修改 `file/v1` → 通知 `file-service`

## 常用命令

### 基础设施
```bash
docker compose up -d    # 启动中间件（MySQL, etcd, Redis, APISIX, MinIO）
docker compose down     # 停止中间件
```

### Proto 管理（api-proto/）
```bash
cd api-proto
make generate           # buf generate — 生成 Go 代码
make lint               # buf lint — Proto 规范检查
make format             # buf format -w — 格式化 proto 文件
make ci                 # lint + breaking-check + generate
```

### Go 构建/测试（全局）
```bash
# 构建任意模块（go.work 自动解析本地依赖）
cd services/<name> && go build ./...
cd services/<name> && go test ./...

# 公共库测试
cd common && go test ./... -cover
```

### 前端
```bash
cd web/pc
npm run dev             # Vite dev server
npm run build           # Type-check + production build
npm run lint            # vue-tsc type check
npm run test:unit       # Vitest
```

## 密钥与配置

1. **密钥在 `.env`（gitignored）中统一管理**，`.env.example`（入库）提供模板。
2. **服务配置中 `${VAR}` 由 `configx.MustLoad`（`common/pkg/configx/`）展开**。新服务入口必须用 `configx.MustLoad` 而非 go-zero 原生的 `conf.MustLoad`。
3. **简单密钥**（AES_KEY、JWT_*）走 `.env` → `${VAR}`。**RSA 密钥**（PEM 多行）暂保留在 YAML 中。
4. 启动/停止/状态查看：
   ```bash
   bash scripts/start.sh
   bash scripts/stop.sh
   bash scripts/status.sh
   ```

## 架构概览

### Go Workspace（`go.work`）

根 `go.work` 包含 9 个 Go 模块，本地修改互见：

| Module | Directory | go.work |
|--------|-----------|:------:|
| `github.com/guxiao1976/api-proto` | `api-proto/` | ✓ |
| `github.com/guxiao1976/community-common/v2` | `common/` | ✓ |
| `github.com/guxiao1976/community-user` | `services/user-service/` | ✓ |
| `github.com/guxiao1976/community-auth` | `services/auth-service/` | ✓ |
| `github.com/guxiao1976/community-permission` | `services/permission-service/` | ✓ |
| `github.com/guxiao1976/community-file` | `services/file-service/` | ✓ |
| `github.com/guxiao1976/community-master-data-service` | `services/master-data-service/` | ✓ |
| `github.com/guxiao1976/community-moderation-service` | `services/moderation-service/` | ✓ |
| `github.com/guxiao/community-and-home/services/ai-model/rpc` | `services/ai-model-service/rpc/` | ✓ |
| `github.com/guxiao/community-and-home/services/ai-model/api` | `services/ai-model-service/api/` | ✓ |

注意：`ai-model-service` 使用不同的 GitHub org（`guxiao` vs `guxiao1976`），已加入 go.work 仅用于本地解析 api-proto 依赖。

### 服务分层模式

**服务命名规范**：所有 `services/` 下的目录统一使用 `-service` 后缀（如 `auth-service`、`user-service`）。

Go 服务遵循 go-zero 的双层模式（`goctl` 生成）：

- **`rpc/`** — gRPC 入口（`zrpc.MustNewServer`），注册到 etcd。`internal/logic/` 业务逻辑，`internal/svc/` 依赖注入。
- **`api/`** — REST 网关（`rest.MustNewServer`），代理到 gRPC。`internal/handler/` + `internal/logic/`。
- **`model/`** — GORM 数据库模型。

例外：`auth-service`/`permission-service` RPC-only，`moderation-service` REST-only，`ai-model-service` Go+Python 混合。

### Proto 组织（`api-proto/`）

使用 Buf v2 管理，4 个 proto 包：

| Proto | gRPC Service | 消费方 |
|-------|-------------|--------|
| `api/auth/v1/auth.proto` | `AuthService` | auth-service |
| `api/user/v1/user.proto` | `UserService` | user-service, auth-service |
| `api/permission/v1/permission.proto` | `PermissionService` | permission-service |
| `api/aimodel/v1/ai_model.proto` | `AiModelService` | ai-model-service, moderation-service |
| `api/common/v1/common.proto` | 共享类型（无 Service） | 所有使用 api-proto 的服务 |

master-data-service、ai-model-service 使用各自 `rpc/pb/` 下的本地 proto。

### 服务间 gRPC 依赖

```
auth-service ──gRPC──> user-service (GetUserByPhone, CreateUser, UpdateUser)
permission-service ——> 独立（无出站 gRPC 调用）
master-data-service ——> 独立（无出站 gRPC 调用）
moderation-service ——> 直接读 masterdata_db 数据库（⚠️ 架构债务，非 gRPC）
ai-model-service ——> 独立（不在 go.work，独立 GitHub org）
```

### 错误码规范

5 位 `XXYYY` 格式：`XX`=服务中心（99=Common, 10=User, 50=Auth, 06=Permission, 07=File, 30=AI, 40=Moderation），`YYY`=具体错误（001-999），`0`=成功。

HTTP 响应：成功 `{code:0, msg:"success", data:<payload>}`，失败 `{code:<6-digit>, msg:"<desc>", data:null}`。gRPC 响应嵌入 `BaseResp` 为首字段。

### 前端代理映射（Vite dev server，端口 3003）

完整配置见 `web/pc/vite.config.ts`。开发时 Vite 直连 Go 服务，生产走 APISIX :9080。

### 中间件（docker-compose，固定 IP 172.18.0.0/24）

| 服务 | IP | 端口 |
|------|-----|------|
| MySQL 8.0 | 172.18.0.2 | 3306 |
| etcd v3.5 | 172.18.0.3 | 2379 |
| Redis 7 | 172.18.0.4 | 6379 |
| APISIX | 172.18.0.5 | 9080/9090 |
| MinIO | 172.18.0.6 | 9000/9001 |
