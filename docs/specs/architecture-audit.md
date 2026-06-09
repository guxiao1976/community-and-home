# 架构审查报告

**日期**: 2026-06-04
**审查范围**: api-proto、common、7 个微服务（auth-service, user-service, permission-service, file-service, master-data-service, moderation-service, ai-model-service）、web 前端层
**审查维度**: Proto 一致性、Go 模块结构、配置规范、文档完整、依赖一致性、架构合规
**审查方法**: 自动化扫描 + 人工核实，覆盖 6 个维度

---

## 总体评级: **有风险 (At Risk)**

**评级依据**:
- 6/6 审查维度均未通过
- 10 项 CRITICAL 问题：涉及端口冲突、认证绕过、跨服务通信违规、直接数据库访问、Snowflake ID 规范遗漏
- 25 项 WARNING 问题：配置不规范、模块路径不一致、文档过时、错误码体系混乱
- 多个 CRITICAL 问题存在关联关系，表明架构层面缺乏系统性治理

**核心风险**: 3 项 CRITICAL 问题可导致生产环境故障 — (1) RPC 端口冲突导致服务无法启动，(2) SMS 验证码认证绕过，(3) JWT 密钥硬编码导致鉴权失效。

---

## 一、CRITICAL（架构不一致 — 必须修复）

### 共性问题 1: 跨服务通信架构违规（2 项关联）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| C1 | **moderation-service 通过 HTTP 直连 ai-model-service**（`services/moderation-service/internal/llm/remote.go:59`），推送 `POST /api/moderate/text` 到 Python FastAPI 端点。ai-model-service 已暴露 `AiModelService.ModerateText` gRPC RPC，但 moderation-service 未使用。违反 CLAUDE.md 全局规则"所有服务间通信必须通过 gRPC"。 | CRITICAL | P0 | moderation-service |
| C2 | **moderation-service 直接访问 master-data-service 数据库**（`md_sensitive_word` 表）。`WordStore`（`internal/wordstore/store.go`）通过直接 SQL 查询加载敏感词。master-data-service 的任何 schema 变更都会静默破坏 moderation-service。与 C1 同属跨服务通信违规，且 proto 注释（`moderation.proto:15`）已标记此为架构债务。 | CRITICAL | P0 | moderation-service, master-data-service |

### 共性问题 2: Snowflake ID / jstype 规范遗漏（3 项关联）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| C3 | **file.proto 完全缺失 jstype 注解**（`api-proto/api/file/v1/file.proto`）：11 个 int64 ID 字段（id, user_id, entity_id, file_id 等）均无 `[jstype = JS_STRING]`。该 proto 于 2026-05-31 新增，但 2026-06-03 的全局 Snowflake 变更（CHANGELOG 记录）仅覆盖了 user/auth/permission，遗漏了 file.proto。违反 CLAUDE.md 全局规则。 | CRITICAL | P0 | api-proto（全局实例） |
| C4 | **moderation.proto 仅 2 个字段有 jstype**（`audit_log_id`, `reviewer_id`），约 6 个 int64 字段缺少注解。虽然比 file.proto 稍好，但仍不符合全局规范。 | CRITICAL | P1 | api-proto（全局实例） |
| C5 | **master-data 本地 proto 缺失 jstype**（`services/master-data-service/rpc/masterdata.proto`）：所有 int64 ID 字段均无 jstype 注解。api-proto 版本（`api-proto/api/masterdata/v1/masterdata.proto`）已全部添加，但本地旧文件未被删除或同步。 | CRITICAL | P0 | master-data-service + api-proto（全局实例） |

### 配置层面严重隐患（3 项）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| C6 | **RPC gRPC 端口冲突**: master-data-service rpc（`:8083`）与 auth-service rpc（`:8083`）冲突；ai-model-service rpc 与 permission-service rpc（`:8084`）冲突。ai-model-service 自身有 4 个 YAML 文件配置不同端口（8080/8084），默认入口指向 8080 但 API 网关指向 8084。 | CRITICAL | P0 | master-data-service, ai-model-service |
| C7 | **JWT Secret 硬编码为占位符**: `perm-api.yaml` 使用 `"your-jwt-secret-key-change-in-production"`，`file-api.yaml` 使用 `"changeme"`，而非 `${JWT_ACCESS_SECRET}`。两个服务将静默接受任意 JWT 或鉴权失败。 | CRITICAL | P0 | permission-service, file-service |
| C8 | **SMS 验证码认证绕过**（`auth-service/rpc/internal/logic/auth/registerlogic.go:47-65`, `loginsmslogic.go:43-65`）：验证码写入 Redis 后从未与用户输入比对，任意 6 位数字码通过注册/登录验证。这是关键认证绕过漏洞。 | CRITICAL | P0 | auth-service |

### 模块依赖缺口（1 项）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| C9 | **master-data-service 缺失 community-common/v2 replace 指令**（`services/master-data-service/go.mod`）：声明依赖 `github.com/guxiao1976/community-common/v2 v2.0.0` 且 50+ 源文件引用，但无 `replace` 指令。在 go.work 外运行 `go mod tidy` 或 `go build` 将失败。所有其他服务均有此 replace。 | CRITICAL | P0 | master-data-service |

---

## 二、WARNING（潜在风险 — 应尽快修复）

### 共性问题 3: 错误码体系碎片化（4 项关联）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| W1 | **错误码位数不一致**：Proto 文档定义 6 位 XXYYYY 格式，但实际代码使用 5 位码（50001, 60001, 70001, 100001）。`errx.go` 定义 6 位码（990400, 400400），但服务使用 5 位码，导致同语义码数值差 10 倍。 | WARNING | P1 | 全局实例 + 各服务 |
| W2 | **未文档化的错误码**：auth-service 使用 509503/509504（RPC 调用失败）、50006（登出失败）；permission-service 使用 60006（重复角色编码），均不在 proto 头注释中。auth-service 还将 user-service 的错误码（100002）通过 BaseResp 直接透传给客户端，泄露其他服务的错误命名空间。 | WARNING | P1 | auth-service, permission-service |
| W3 | **ai-model-service 无错误码系统**：所有错误返回 `fmt.Errorf(...)` 纯字符串，调用方无法程序化区分错误类型（超时 vs 无效输入 vs 鉴权失败）。其他服务依赖 ai-model-service，这破坏了统一错误处理模式。 | WARNING | P1 | ai-model-service |
| W4 | **master-data-service 无专属错误码前缀**：独占使用 common/errx 通用 helper（`NewDefaultError`, `NewNotFoundError`, `NewInvalidParamError`），全部映射到 99xxxx（Common 范围）。调用方无法区分"master-data 未找到"vs"其他服务未找到"。 | WARNING | P1 | master-data-service |

### 共性问题 4: Proto 元数据不一致（3 项关联）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| W5 | **RPC 消息命名混乱（4 种并存）**: Request/Response（auth/user/permission/file/moderation）、Req/Resp（masterdata 全部 11 个 RPC）、混用（aimodel 同时有 `ModelCallRequest` 和 `ModelHealthCheckReq`）、孤立后缀（`ModelConfigResp`）。影响代码生成和客户端可维护性。 | WARNING | P1 | api-proto（全局实例） |
| W6 | **Proto BaseResp 采用不完整**: 3/7 proto 包不使用 `common.v1.BaseResp`（aimodel/v1, masterdata/v1, moderation/v1），各自定义 ad-hoc 响应结构。导致 gRPC 客户端无法统一提取错误码，跨服务错误传播不一致。 | WARNING | P1 | api-proto（全局实例） + 各服务 |
| W7 | **master-data 本地 proto 未清理**（`services/master-data-service/rpc/masterdata.proto`）：已迁移至 api-proto 但旧文件未删除或替换为指针注释（应参考 ai-model-service 的做法）。package 命名（`masterdata` vs `masterdata.v1`）、go_package（`./pb` vs api-proto 路径）、Service 名称（`Masterdata` vs `MasterdataService`）均不一致。 | WARNING | P0 | master-data-service |

### 共性问题 5: 配置标准化不足（4 项关联）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| W8 | **configx.MustLoad 采用率仅 43%**（6/14 入口）：master-data-service、permission-service、file-service、ai-model-service 的 api+rpc 入口仍使用 `conf.MustLoad`，无法展开 `${ENV_VAR}` 占位符，不可使用 `.env` 变量。 | WARNING | P1 | 4 个服务 |
| W9 | **硬编码数据库密码**：12 个 YAML 文件均硬编码 DB 密码，且存在两种不一致模式（`root:root123456` vs `root:123456`）。master-data api 和 moderation api/rpc 还硬编码 Redis 密码。 | WARNING | P1 | 所有服务 |
| W10 | **硬编码第三方 API 密钥**：masterdata-api.yaml 暴露高德地图 API Key、file-service rpc 硬编码 MinIO `admin:admin123`、ai-model rpc 硬编码加密密钥。均应改为 `${ENV_VAR}`。 | WARNING | P1 | master-data-service, file-service, ai-model-service |
| W11 | **ai-model-service 存在 4 个同名 YAML 文件**（rpc/etc/）：`ai-model.yaml`, `aiModel.yaml`, `aimodel.yaml`, `ai_model.yaml`，端口和 etcd key 各不相同。默认入口使用 `aimodel.yaml`（端口 8080），但 API 网关指向 `:8084`，导致连接失败。 | WARNING | P0 | ai-model-service |

### 模块与依赖规范（4 项）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| W12 | **模块路径命名不一致**：4 个服务使用短名 `community-xxx`（auth, file, permission, user），2 个使用长名 `community-xxx-service`（master-data, moderation）。目录名统一用 `-service` 后缀但模块路径不匹配。无迁移计划。 | WARNING | P2 | 全局实例 |
| W13 | **ai-model-service 使用不同 GitHub org**（`guxiao` vs `guxiao1976`）和不同 repo 前缀（`community-and-home` vs `community-*`）。这是有意设计（CLAUDE.md 记录），但阻止了 ai-model-service 与其他服务共享 go.work 模块解析。 | WARNING | P2 | ai-model-service |
| W14 | **go-zero 版本文档过时**：8 个服务的 CLAUDE.md 文档声明 go-zero v1.10.1，实际 go.mod 均为 v1.10.2。 | WARNING | P2 | 8 个服务 |
| W15 | **api-proto 版本引用不一致**：master-data-service 和 ai-model-service 引用 `v0.0.0`，其他服务引用 `v0.1.0`。 | WARNING | P2 | master-data-service, ai-model-service |

### 文档与规范合规（5 项）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| W16 | **残留审查文件**：`services/ai-model-service/_review.md`（139 行）和 `services/moderation-service/_review.md`（166 行）应在审查完成后删除。 | WARNING | P2 | ai-model-service, moderation-service |
| W17 | **docs/specs/ 迁移不完整**：`user.md`、`auth.md`、`permission.md` 是原始设计内容，应转换为指向各服务 `docs/design.md` 的指针（其兄弟文件已完成转换）。 | WARNING | P2 | 全局实例 |
| W18 | **moderation-service CLAUDE.md 过时**：描述为"REST-only"且"gRPC 层待实现"，但 go.mod 已声明 grpc 依赖，rpc/ 目录已包含完整的 gRPC 服务器实现。 | WARNING | P2 | moderation-service |
| W19 | **master-data-service CLAUDE.md 模块路径不准确**：描述自身为 `community-master-data`，实际 go.mod 为 `community-master-data-service`。 | WARNING | P2 | master-data-service |
| W20 | **docs/specs/ 中存在过时文件**：`task-1.md`（任务描述）、`migration.sql`（应移至服务 migrations/ 目录）。 | WARNING | P2 | 全局实例 |

### 其他规范不一致（5 项）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| W21 | **configx.MustLoad 不满足 CLAUDE.md 要求的 8 个入口**：master-data-service (api+rpc), permission-service (api+rpc), file-service (api+rpc), ai-model-service (api+rpc) 使用 `conf.MustLoad` 而非 `configx.MustLoad`。 | WARNING | P1 | 4 个服务 |
| W22 | **etcd Key 命名不一致**：6 个服务使用 `<name>.rpc` 格式，ai-model-service 混用 `ai-model.rpc`（含连字符）和 `aimodel.rpc`。连字符可能导致 etcd 客户端和 DNS 服务发现问题。 | WARNING | P1 | ai-model-service |
| W23 | **JWT 配置段命名不一致**：auth-service 使用 `JwtAuth`，user-service/moderation 使用 `Auth`，字段结构各不相同。代码解析不同 struct tag，虽非 bug 但极易混淆。 | WARNING | P2 | user-service, moderation-service |
| W24 | **Cache Redis 配置格式不统一**：三种格式并存 — YAML列表含Pass、YAML列表不含Pass、平铺字符串。同一基础设施配置方式过于分散。 | WARNING | P2 | 所有服务 |
| W25 | **replace 指令格式不一致**：moderation-service 拆分为两个独立 `replace` 块；master-data-service 使用单行 replace 而非括号分组。虽语法有效但风格不一致。 | WARNING | P3 | moderation-service, master-data-service |

---

## 三、NOTE（改进建议 — 可择机处理）

| ID | 问题 | 严重程度 | 优先级 | 负责方 |
|----|------|----------|--------|--------|
| N1 | `.env.example` 覆盖不足：缺失 `DB_PASSWORD`, `REDIS_PASS`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `ENCRYPTION_KEY`（ai-model）, `AMAP_KEY`（master-data）, `LARGE_MODEL_API_KEY`（moderation）。 | NOTE | P2 | 全局实例 |
| N2 | Go 工具链版本微小不一致：`go.work` 和 `ai-model-service/api` 使用 go 1.25.10，其他 7 个 go.mod 使用 go 1.25.0。 | NOTE | P3 | 各服务 |
| N3 | 数据库主机格式不一致：多数服务使用 `localhost:3306`，file-service rpc 使用 `127.0.0.1:3306` 并附加 `loc=Asia%2FShanghai`。mix localhost/127.0.0.1 可能导致 socket vs TCP 连接差异。 | NOTE | P3 | file-service |
| N4 | RSA 密钥在 auth-service api+rpc 中重复：虽为有意设计（CLAUDE.md 记录），但密钥轮换需编辑两处，增加操作风险。 | NOTE | P3 | auth-service |
| N5 | REST API 端口 8885-8888 未分配：存在间隙但无实际影响，可在新增服务时按序分配。 | NOTE | P3 | 全局实例 |
| N6 | go.work 中 ai-model-service 以子模块（api, rpc）独立注册：与实现一致但与其他服务差异明显，文档中应予说明。 | NOTE | P3 | 全局实例 |
| N7 | ai-model-service/api 对 self-rpc 使用 replace 依赖：有效但脆弱，外部服务无法直接依赖其 rpc client。 | NOTE | P3 | ai-model-service |
| N8 | `docs/specs/` 中存在 4 个尚未实现服务的原始规格（approve.md, risk.md, notify.md, audit.md）：当前无法迁移。跟踪为技术债务。 | NOTE | P3 | 全局实例 |
| N9 | TODO/FIXME 标记（9 处）：3 处涉及 master-data-service 硬编码零值（CreatedBy, UserId, EntityId），1 处 JWT middleware 为存根，1 处 SMS 供应商未集成。 | NOTE | P2 | 各服务 |
| N10 | moderation-service SubmitReview handler 存在但未注册路由：`content_review/submit_review_handler.go` 未在 `routes.go` 中注册，types.go 缺失 `SubmitReviewReq/SubmitReviewResp`，代码无法编译。 | NOTE | P2 | moderation-service |
| N11 | moderation-service AuthMiddleware 为存根：自定义中间件体仅透传，实际鉴权依赖框架级 `rest.WithJwt()`。若框架级检查被移除或静默失败，鉴权被绕过。 | NOTE | P2 | moderation-service |
| N12 | `api-proto/CHANGELOG.md` 仅覆盖 2026-05-31 至今：auth/v1, user/v1, permission/v1, common/v1 的初始创建无记录。近期迁移已完整记录。 | NOTE | P3 | api-proto（全局实例） |
| N13 | `docs/specs/` 中 `migration.sql` 位置不当：应移至对应服务的 `migrations/` 目录。 | NOTE | P3 | 全局实例 |
| N14 | `docs/specs/task-1.md` 为任务描述而非设计规格，完成后应删除。 | NOTE | P3 | 全局实例 |

---

## 四、跨维度共性问题总结

### 问题 A: 跨服务通信治理缺失（C1, C2）
`moderation-service` 同时以两种方式违反全局规则：HTTP 直连 ai-model-service 的 Python 端点，以及直接 SQL 查询 master-data-service 的数据库表。两个问题根因相同 — 服务间缺乏强制 gRPC-only 治理机制。ai-model-service 已有 gRPC 接口但未被使用，master-data-service 缺少 moderation 所需的敏感词查询 gRPC 端点。

### 问题 B: Snowflake ID 规范执行不完整（C3, C4, C5）
2026-06-03 的全局变更仅覆盖 4 个 proto 包（user, auth, permission, common），遗漏 file（新增）和 masterdata（本地残留）。说明全局规范的执行依赖人工记忆而非自动化检查，缺少 CI 中的 breaking/buf lint 规则来强制执行 jstype。

### 问题 C: master-data-service 迁移不彻底（C5, C9, W7）
同一个服务在 Proto 审查、模块审查、架构审查中均被发现同一类问题：proto 未切换至 api-proto 版本、旧文件未按规范清理、replace 指令缺失。这是"服务从独立 proto 迁移到 api-proto 统一管理"过程中留下的技术债务，需要一次性系统性解决。

### 问题 D: 配置管理缺乏统一标准（C6, C7, W8-W11）
跨 4 个服务发现配置问题：`conf.MustLoad` vs `configx.MustLoad` 混用、硬编码密钥、端口冲突、YAML 文件歧义。根本原因是缺乏配置模板和 CI 检查，新增服务时未被要求遵循统一配置规范。

### 问题 E: 错误处理体系碎片化（W1-W4）
Proto 定义的 6 位错误码与实际代码 5 位码不一致、部分服务无错误码系统、部分使用通用码无专属前缀、错误码透传跨服务。根因是 `common/errx` 未成为强制依赖，各服务自行演进出不同实践。

---

## 五、修复优先级矩阵

| 优先级 | 数量 | 典型问题 | 建议修复窗口 |
|--------|------|----------|-------------|
| P0（立即） | 8 | C1-C9（除C4）| 本周内 |
| P1（本迭代） | 9 | C4, W1-W8, W11 | 2 周内 |
| P2（下迭代） | 13 | W9-W25, N1-N5, N7-N11 | 1 个月内 |
| P3（技术债务） | 8 | N2-N4, N6, N8, N12-N14 | 跟踪，适时处理 |

---

## 六、各服务问题分布

| 服务 | CRITICAL | WARNING | NOTE | 关注度 |
|------|----------|---------|------|--------|
| master-data-service | C5, C6, C9 | W4, W7, W8, W9, W19 | N1 | **高** |
| moderation-service | C1, C2 | W18, W8, W9 | N10, N11 | **高** |
| ai-model-service | C6 | W3, W8, W9, W10, W11, W13 | N2, N7 | **高** |
| auth-service | C8 | W1, W2, W23 | N4 | **中** |
| permission-service | C7 | W1, W2, W8, W23 | — | **中** |
| file-service | C7 | W8, W9, W10 | N3 | **中** |
| api-proto（全局） | C3, C4 | W5, W6 | N12 | **中** |
| user-service | — | W1, W23 | — | **低** |
| common/ | — | — | — | **低** |
| web/ | — | — | — | **低** |

---

## 七、建议修复路线图

### 阶段 1: 止血（P0，本周）
1. 修复 SMS 验证码认证绕过（C8）— 补全 `registerlogic.go` 和 `loginsmslogic.go` 中的验证码比对逻辑
2. 修复 RPC 端口冲突（C6）— 为 master-data-service 分配独立端口（建议 :8085），统一 ai-model-service 为单一 YAML 文件和端口
3. 修复 JWT Secret 占位符（C7）— 将 perm-api.yaml 和 file-api.yaml 的 AccessSecret 改为 `${JWT_ACCESS_SECRET}`
4. 为 file.proto 和 masterdata.proto 添加 jstype 注解（C3, C5）
5. master-data-service 添加 community-common/v2 replace 指令（C9）
6. moderation-service 切换至 ai-model-service gRPC 调用（C1）
7. moderation-service 通过 master-data-service gRPC 查询敏感词（C2）
8. 删除 master-data-service 本地旧 proto 文件或替换为指针注释（W7）

### 阶段 2: 规范化（P1，2 周内）
1. 统一错误码格式和分析（W1-W4）— 决定 5 位还是 6 位，全量更新
2. 迁移 4 个服务入口到 configx.MustLoad（W8/W21）
3. 清理硬编码第三方密钥（W10）
4. 统一 RPC 消息命名约定（W5）
5. 统一 ai-model-service YAML 配置（W11）
6. 为 moderation.proto 补充缺失的 jstype 注解（C4）
7. BaseResp 标准化（W6）
8. 将敏感配置迁移至 `.env`（W9）
9. 统一 etcd Key 命名（W22）

### 阶段 3: 清理（P2-P3，1 个月内）
1. 统一模块路径命名（W12）或形成文档化的差异说明
2. 更新过时 CLAUDE.md 文档（W14, W18, W19）
3. 清理残留审查文件（W16）
4. 完成 docs/specs/ 迁移（W17, W20）
5. 补齐 `.env.example`（N1）
6. 处理 TODO/FIXME 标记（N9）
7. 注册 moderation-service SubmitReview handler（N10）

---

## 八、附录：审查方法说明

本次审查覆盖 6 个维度：
- **Proto 一致性**：验证 8 个 api-proto proto 文件（含 buffer 管理的本地 proto）的 6 项规范（package 命名、go_package、jstype、Service 命名、版本同步、buf 管理）
- **Go 模块结构**：审查 7 个微服务的 go.mod、go.work、CLAUDE.md 的对齐性
- **配置规范检查**：扫描 14 个入口点 + 16 个 YAML 文件的 8 个检查点（configx 合规、端口分配、硬编码密钥、etcd 命名、超时、.env 覆盖）
- **文档完整性**：审查 7 个服务 + api-proto + common + web 的 docs/specs 迁移状态和设计文档一致性
- **依赖一致性审计**：分析 10 个 go.mod 模块的依赖版本、replace 指令、间接依赖的正确性
- **架构合规性检查**：验证 7 个维度的运行时架构规则（跨服务通信、数据库访问、认证安全、错误码、命名约定、BaseResp、TODO 标记）

---

*报告生成时间: 2026-06-04 | 审查工具: Claude Code 架构审查 Agent 链（6 个维度并行审查 + 汇总 Agent）*
