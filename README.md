# Community-Home 微服务项目

这是一个基于 Go-Zero 的微服务架构项目，包含用户、认证、权限、审核、文件、社区等多个服务。

## 🚀 快速开始

### 1. 环境要求

- Go 1.21+
- Docker & Docker Compose
- Git
- Node.js 18+ (前端开发)

### 2. 首次设置

```bash
# 克隆仓库
git clone <repository-url>
cd community-and-home

# 安装 Git Hooks（必须！）
bash scripts/install-git-hooks.sh

# 启动基础设施（MySQL, Redis, etcd, MinIO, Neo4j, APISIX 等）
docker compose up -d

# 初始化数据库（幂等建库：只创建缺失的库，对已有数据零影响）
bash scripts/init-databases.sh            # 或 --check-only 仅检查缺失

# 启动所有后端服务（RPC + API，按依赖顺序）
bash scripts/start.sh

# 启动前端（新终端）
bash scripts/start-frontend.sh            # 全部前端
#   --pc-only     仅 PC 管理后台 (web/pc, 默认 http://localhost:5173)
#   --mobile-only 仅移动端 H5 (web/mobile)
#   --stop        停止前端进程
```

> **数据安全说明**：`init-databases.sh` 只执行 `CREATE DATABASE IF NOT EXISTS`
> （幂等，库已存在即跳过），不包含任何 `DROP / TRUNCATE / DELETE`，可放心重复执行。
> 业务数据请通过备份机制保护（mysqldump / docker volume）。

### 2.1 表结构变更（Migration）— 人工执行 ⚠️

**建库 ≠ 建表**。`init-databases.sh` 只创建数据库外壳，**表结构由各服务的
migration SQL 人工按序执行**（不做自动化执行，schema 变更必须人工确认）。

执行流程：

1. **确认库已存在**：`bash scripts/init-databases.sh --check-only`
2. **找到目标 migration**：表结构定义在各服务 `services/<svc>/migration*/` 下
   （如 `user-service/migration/000_initial_schema.sql`），按文件名编号顺序执行
3. **确认目标库**（三重标注，执行前必须核对一致）：
   - migration 文件头注释：`-- Database: user`
   - 文件内 `USE xxx;`（部分文件自带）
   - 服务 DSN：`services/<svc>/api|rpc/etc/*.yaml` 的 `DataSource: .../xxx_db`
4. **人工执行**：
   ```bash
   docker exec mysql mysql -u root -p<密码> <库名> < services/<svc>/migration/<file>.sql
   ```
5. **执行后验证**：
   ```bash
   docker exec mysql mysql -u root -p<密码> <库名> -e "DESCRIBE <表名>;"
   ```

**铁律（must-follow）**：写 SQL → 提交 → **执行** → 验证，四步缺一不可。
只提交不执行 = 上线必踩 `Error 1054: unknown column`（历史教训，见
`.harness/knowledge/memory/migration-must-execute.md`）。

### 3. Git Hooks 说明

**为什么必须安装 Git Hooks？**

Git Hooks 会在你每次 `git commit` 时自动运行质量检查，在问题进入代码库之前就拦截：

- ✅ **检查 Logic 文件必须有测试** — 防止提交未测试的业务逻辑
- ✅ **运行变更包的单元测试** — 确保代码改动不破坏现有功能
- ✅ **代码格式检查（gofmt）** — 保持代码风格一致
- ✅ **静态分析（golangci-lint）** — 发现潜在问题

**收益**：
- 🚀 本地 5-15 秒立即反馈，而不是等待 CI 运行 2-5 分钟
- 💰 减少 CI 失败率从 ~30% → ~5%，节省时间和资源
- 🎯 只推送高质量代码，提升团队效率

**跳过检查**（不推荐）：
```bash
git commit --no-verify
```

## 📁 项目结构

```
community-and-home/
├── api-proto/              # Proto 定义和代码生成
├── common/                 # Go 共享库（v2）
├── services/               # 微服务
│   ├── user-service/       # 用户服务
│   ├── auth-service/       # 认证服务
│   ├── permission-service/ # 权限服务
│   ├── master-data-service/# 主数据服务
│   ├── moderation-service/ # 审核服务
│   ├── ai-model-service/   # AI 模型服务
│   ├── file-service/       # 文件服务
│   ├── community-hub-service/ # 社区枢纽服务
│   └── monitoring-service/ # 监控服务
├── web/                    # 前端
│   ├── pc/                 # 管理后台
│   └── mobile/             # 移动端
├── scripts/                # 工具脚本
└── .harness/               # AI 驱动的开发流水线
```

## 🛠️ 常用命令

### 服务管理

```bash
# 启动所有服务
bash scripts/start.sh

# 停止所有服务
bash scripts/stop.sh

# 查看服务状态
bash scripts/status.sh

# 重启服务
bash scripts/restart.sh
```

### 开发与测试

```bash
# 运行单个服务的测试
cd services/<service-name>
go test ./... -v

# 运行机械化质量检查（15 项）
bash .harness/skills/qa/scripts/harness-checks.sh --service <service-name>

# 运行所有检查并生成 JSON 报告
bash .harness/skills/qa/scripts/harness-checks.sh --service <service-name> --json
```

### Proto 管理

```bash
cd api-proto

# 生成代码
make generate

# 规范检查
make lint

# 完整 CI 检查（lint + breaking-check + generate）
make ci
```

### 前端开发

```bash
# 管理后台
cd web/pc
npm install
npm run dev        # 开发服务器
npm run build      # 构建生产版本
npm run test:unit  # 运行单元测试

# 移动端
cd web/mobile
npm install
npm run dev
```

## 🔍 开发流程

### 新功能开发

1. **创建分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **TDD 开发**（推荐）
   ```bash
   # 先写测试（让它失败）
   vim services/<service>/api/internal/logic/xxx_logic_test.go
   
   # 实现功能（让测试通过）
   vim services/<service>/api/internal/logic/xxx_logic.go
   
   # 重构（保持测试通过）
   ```

3. **本地验证**
   ```bash
   # Git Hook 会自动运行，但你也可以手动检查
   bash .harness/skills/qa/scripts/harness-checks.sh --service <service-name>
   ```

4. **提交代码**
   ```bash
   git add .
   git commit -m "feat(service): add xxx feature"
   # Git Hook 自动运行检查
   ```

5. **推送并创建 PR**
   ```bash
   git push origin feature/your-feature-name
   # 在 GitHub 上创建 Pull Request
   ```

### Bug 修复

1. **复现问题** — 先写一个失败的测试来复现 bug
2. **修复代码** — 让测试通过
3. **回归测试** — 确保没有引入新问题
4. **提交** — Git Hook 自动验证

## 📚 文档索引

| 文档 | 用途 |
|------|------|
| [CLAUDE.md](CLAUDE.md) | Claude Code 协作指南 |
| [.harness/rules/工程结构.md](.harness/rules/工程结构.md) | 工程结构和分层规范 |
| [.harness/rules/Proto管理规范.md](.harness/rules/Proto管理规范.md) | Proto 变更流程 |
| [.harness/rules/项目编码规范.md](.harness/rules/项目编码规范.md) | 编码规范和约束 |
| [.harness/knowledge/INDEX.md](.harness/knowledge/INDEX.md) | 项目知识库 |
| [.harness/tasks/BACKLOG.md](.harness/tasks/BACKLOG.md) | 任务队列 |

## 🤖 AI 驱动的开发流水线

本项目使用自定义的 `.harness/` 流水线，包含：

- **机械化检查**：21 项自动化检查（Go 15 项 + 前端 6 项）
- **TDD 工作流**：强制 RED → GREEN → REFACTOR
- **多视角代码审查**：3 视角 × 9 维度并行审查
- **记忆驱动编码**：自动注入历史经验和最佳实践

详见 [docs/specs/ai-dev-team-design.md](docs/specs/ai-dev-team-design.md)

## 🔧 故障排查

### Git Hook 问题

```bash
# 重新安装 Git Hooks
bash scripts/install-git-hooks.sh

# 检查是否安装成功
ls -la .git/hooks/pre-commit

# 应该看到符号链接指向：
# .git/hooks/pre-commit -> ../../.harness/scripts/git-hooks/pre-commit
```

### 测试失败

```bash
# 查看详细错误
go test ./services/<service>/... -v

# 只运行特定测试
go test ./services/<service>/... -v -run TestXxx
```

### 服务启动失败

```bash
# 检查日志
docker compose logs <service-name>

# 检查端口占用
netstat -tlnp | grep <port>

# 重启中间件
docker compose restart
```

## 🌟 核心规范

1. **Proto 变更仅在主仓库执行** — 子服务禁止修改 `api-proto/`
2. **Logic 文件必须有测试** — `*_logic.go` 必须有对应的 `*_logic_test.go`
3. **提交前必须通过检查** — Git Hook 会自动验证
4. **Snowflake ID 使用字符串** — Proto `[jstype=JS_STRING]` + Go `json:",string"`
5. **密钥使用 `.env`** — 不可硬编码在代码中

完整规范见 [.harness/rules/项目编码规范.md](.harness/rules/项目编码规范.md)

## 🤝 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 运行 `bash scripts/install-git-hooks.sh` 安装 Git Hooks
4. 提交改动 (`git commit -m 'feat: add amazing feature'`)
5. 推送到分支 (`git push origin feature/amazing-feature`)
6. 创建 Pull Request

**注意**：PR 必须通过 CI 检查才能合并。

## 📞 获取帮助

- 项目文档：`.harness/knowledge/INDEX.md`
- 历史问题：`.harness/knowledge/memory/MEMORY.md`
- 任务追踪：`.harness/tasks/BACKLOG.md`

---

**License**: MIT

**团队**: Community-Home Dev Team
