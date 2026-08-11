# CLAUDE.md — file-service

## 角色定位

这是 **文件服务**（`github.com/guxiao1976/community-file`），提供 MinIO 对象存储的统一文件上传（预签名 URL）、下载、删除和列表。**RPC + API 双层**，gRPC 供其他服务调用，REST API 供前端直调。

## 启动上下文

按顺序读取：

1. `docs/graph-context.md` — Neo4j 自动生成的服务上下文
2. `docs/design.md` — 数据模型、上传流程
3. `../../.harness/rules/项目编码规范.md` — 编码硬性约束
4. `../../.harness/knowledge/memory/MEMORY.md` — 全局经验，按触发词匹配
5. `.claude/memory/` — 本服务经验记忆

## 关键规则

以下仅列出本服务特有规则，通用约束见 `.harness/rules/`：

1. **客户端直传模式** — 上传用预签名 URL，客户端直传 MinIO，文件流不经过本服务
2. **MinIO 失败不阻塞 DB** — 删除时优先保证 DB 元数据一致性，MinIO 失败仅日志
3. **复用 common/pkg/minio** — 下载/删除用 common 封装，预签名上传用原始 minio-go 客户端

## 全局公约

全局约束见根 [`CLAUDE.md`](../../CLAUDE.md) §7条硬性约束。提交前 `bash ../../.harness/skills/qa/scripts/harness-checks.sh --service file-service`。
## 常用命令

```bash
go build ./...        # 构建
go test ./...         # 测试
cd rpc && go run fileservice.go   # 运行 RPC (8085)
cd api && go run file.go          # 运行 API (8884)
```

## 架构

```
api/                    # REST API 层（预签名 URL + 确认上传）
rpc/                    # gRPC 服务（Upload/Download/Delete/List）
  internal/
    logic/file/         # 业务逻辑（MinIO 操作 + DB 元数据）
model/                  # GORM 数据模型
```
