# Proto 管理规范

> 最高优先级 · 硬性架构约束 · 所有服务必须遵守，无一例外
> 最后更新 2026-06-09

---

## 1. 统一存放

**所有微服务的 gRPC 接口定义必须放在 `api-proto/` 中**，不得在服务目录中私自定义 proto。

```
api-proto/api/
  auth/v1/auth.proto              # AuthService
  user/v1/user.proto              # UserService
  permission/v1/permission.proto  # PermissionService
  aimodel/v1/ai_model.proto       # AiModelService
  common/v1/common.proto          # 共享类型
  file/v1/file.proto              # FileService
  masterdata/v1/masterdata.proto  # MasterDataService
  moderation/v1/moderation.proto  # ModerationService
```

- ✅ 所有 Proto 定义 → `api-proto/`
- ❌ 服务本地 proto → 禁止（`rpc/pb/` 只能放跳转指针）
- ❌ 服务间直接用 HTTP 调用 → 必须走 gRPC

## 2. 变更流程

```
1. 修改 api-proto/api/<service>/v1/*.proto         ← 在此编辑
2. cd api-proto && make generate                    ← 生成 Go 代码
3. cd api-proto && make lint                        ← 规范检查
4. cd api-proto && make breaking-check              ← 破坏性变更检测
5. 记录到 api-proto/CHANGELOG.md                    ← 变更日志
6. 通知受影响的服务重新构建                            ← 影响评估
```

### 影响范围

| 修改 | 通知 |
|------|------|
| `auth/v1` | auth-service |
| `user/v1` | user-service, auth-service（auth 调用 user） |
| `permission/v1` | permission-service |
| `aimodel/v1` | ai-model-service, moderation-service |
| `common/v1` | 所有使用 api-proto 的服务 |
| `file/v1` | file-service |
| `masterdata/v1` | master-data-service |
| `moderation/v1` | moderation-service |

## 3. 权限边界

| 角色 | 可以做什么 | 不能做什么 |
|------|-----------|-----------|
| **全局 Claude**（本实例） | 修改 api-proto、评估影响、通知服务 | 不写服务具体代码 |
| **子 Claude**（服务实例） | 使用 api-proto 生成的代码 | 禁止修改 api-proto/；需要时告知用户切回全局 Claude |

## 4. 子 Claude 必须知道的公约

每个服务的 `CLAUDE.md` **必须包含**以下标准章节，确保子 Claude 感知全局规则：

```markdown
## 全局公约

本项目所有微服务遵守统一的架构规范，详见根 [CLAUDE.md](../../CLAUDE.md)。

与本服务相关的关键约束：
- **Proto 定义在 api-proto/**：本服务的 gRPC 接口定义在 `api-proto/api/<xxx>/v1/`，修改 proto 需告知用户切换到全局 Claude
- **服务间通信仅通过 gRPC**：调用其他服务必须走 gRPC（etcd 服务发现），禁止直连数据库
- **设计文档在 docs/design.md**：数据库、业务流程、接口设计见 [docs/design.md](docs/design.md)
- **变更记录在 CHANGELOG.md**：每次变更必须更新 [CHANGELOG.md](CHANGELOG.md)
- **提交前运行机械化检查**：代码变更提交前必须运行 `bash .harness/scripts/harness-checks.sh --service <服务目录名>`，有 FAIL 则不可提交
```

## 5. 相关记忆

- [[proto-jstype]] — Proto int64 字段必须加 `[jstype = JS_STRING]`
- [[grpc-only-comms]] — 服务间通信仅通过 gRPC

## 6. 常用命令

```bash
cd api-proto
make generate           # buf generate — 生成 Go 代码
make lint               # buf lint — Proto 规范检查
make format             # buf format -w — 格式化 proto 文件
make ci                 # lint + breaking-check + generate
```
