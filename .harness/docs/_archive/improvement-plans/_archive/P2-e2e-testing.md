# P2-1: E2E 测试集成

## 背景

**现状问题**：
- **只有单元测试**：`go test` 覆盖单个函数逻辑
- **缺少集成测试**：服务间 gRPC 调用、数据库操作、Redis 缓存的集成场景未覆盖
- **缺少端到端测试**：前后端联调、完整业务流程的验证依赖人工测试

**暴露的问题**：
- 单元测试全过 ≠ 系统正常工作（接口对接错误、数据格式不一致）
- 生产环境才发现集成问题（如 gRPC 超时、数据库死锁）
- 前端 API 调用 404（路由定义与前端不一致）

**影响**：
- 质量风险高（集成问题进入生产）
- 人工测试成本高（每次变更需要手动回归测试）
- 修复周期长（问题在生产环境发现 → 回溯根因 → 修复 → 发布）

## 目标

建立 **三层测试金字塔**：
```
       E2E Tests (5-10 个关键场景)
      /                          \
 Integration Tests (50+ 场景)     
/                                  \
Unit Tests (500+ 测试用例)           
```

重点补全 **E2E 测试层**，覆盖 5-10 个核心业务流程。

## 技术方案

### 1. 测试分层策略

#### Layer 1: Unit Tests（已有）
- **覆盖**：单个函数、单个组件
- **工具**：Go `testing` / Vitest（前端）
- **运行时机**：每次代码提交（QA Agent 自动运行）
- **数量**：500+（目标）

#### Layer 2: Integration Tests（新增）
- **覆盖**：
  - 服务间 gRPC 调用
  - 数据库 CRUD + 事务
  - Redis 缓存读写
  - 消息队列（如果有）
- **工具**：Go `testing` + 真实依赖（TestContainers）
- **运行时机**：
  - 本地：开发者手动运行
  - CI：每个 PR 自动运行
- **数量**：50+（目标）

#### Layer 3: E2E Tests（新增）
- **覆盖**：完整业务流程（前端 → API → gRPC → 数据库 → 响应）
- **工具**：Playwright（前端）+ Postman/Newman（API）
- **运行时机**：
  - 本地：开发者手动运行
  - CI：每日定时运行 + 发布前运行
  - 阶段 6（集成验证）：Owner Agent 自动运行关键场景
- **数量**：5-10（关键场景）

### 2. E2E 测试框架选型

#### 前端 E2E: Playwright

**优势**：
- 跨浏览器（Chrome / Firefox / Safari）
- 录制和生成测试代码
- 截图和视频录制（失败时自动保存）
- 稳定性高（自动等待、重试）

**示例测试**（用户登录 → 创建活动）：
```typescript
// web/pc/tests/e2e/activity-creation.spec.ts
import { test, expect } from '@playwright/test'

test('创建社区活动端到端流程', async ({ page }) => {
  // 1. 登录
  await page.goto('http://localhost:5173/login')
  await page.fill('input[name="username"]', 'admin')
  await page.fill('input[name="password"]', 'password123')
  await page.click('button[type="submit"]')
  await expect(page).toHaveURL(/.*\/dashboard/)
  
  // 2. 进入活动管理页面
  await page.click('text=活动管理')
  await expect(page).toHaveURL(/.*\/activities/)
  
  // 3. 创建新活动
  await page.click('button:has-text("创建活动")')
  await page.fill('input[name="title"]', 'E2E 测试活动')
  await page.fill('textarea[name="description"]', '这是一个测试活动')
  await page.selectOption('select[name="category"]', '文体活动')
  await page.click('button:has-text("提交")')
  
  // 4. 验证活动创建成功
  await expect(page.locator('.success-message')).toBeVisible()
  await expect(page.locator('text=E2E 测试活动')).toBeVisible()
  
  // 5. 验证数据库中存在该活动（通过 API 查询）
  const response = await page.request.get('http://localhost:8080/api/v1/activities')
  const activities = await response.json()
  const created = activities.data.find(a => a.title === 'E2E 测试活动')
  expect(created).toBeDefined()
  expect(created.category).toBe('文体活动')
})
```

#### API E2E: Newman (Postman CLI)

**优势**：
- 复用 Postman Collection（如果已有）
- 支持环境变量、预处理脚本、断言
- CI 友好（命令行运行）

**示例 Collection**（`tests/e2e/api/activity-flow.postman_collection.json`）：
```json
{
  "info": {"name": "Activity Creation Flow"},
  "item": [
    {
      "name": "Login",
      "request": {
        "method": "POST",
        "url": "{{base_url}}/api/v1/auth/login",
        "body": {"username": "admin", "password": "password123"}
      },
      "event": [{
        "listen": "test",
        "script": {
          "exec": [
            "pm.test('Login successful', function() {",
            "  pm.response.to.have.status(200);",
            "  var json = pm.response.json();",
            "  pm.environment.set('access_token', json.data.access_token);",
            "});"
          ]
        }
      }]
    },
    {
      "name": "Create Activity",
      "request": {
        "method": "POST",
        "url": "{{base_url}}/api/v1/activities",
        "header": [{"key": "Authorization", "value": "Bearer {{access_token}}"}],
        "body": {
          "title": "E2E 测试活动",
          "description": "这是一个测试活动",
          "category": "文体活动"
        }
      },
      "event": [{
        "listen": "test",
        "script": {
          "exec": [
            "pm.test('Activity created', function() {",
            "  pm.response.to.have.status(200);",
            "  var json = pm.response.json();",
            "  pm.expect(json.data.title).to.eql('E2E 测试活动');",
            "  pm.environment.set('activity_id', json.data.id);",
            "});"
          ]
        }
      }]
    }
  ]
}
```

**运行**：
```bash
newman run tests/e2e/api/activity-flow.postman_collection.json \
  --environment tests/e2e/api/test.postman_environment.json \
  --reporters cli,json \
  --reporter-json-export e2e-api-results.json
```

### 3. 测试环境准备

#### Docker Compose 测试环境

**文件**：`docker-compose.test.yml`

```yaml
version: '3.8'

services:
  mysql-test:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: test123
      MYSQL_DATABASE: community_test
    ports:
      - "3307:3306"  # 避免与本地 MySQL 冲突
    volumes:
      - ./tests/e2e/fixtures/init.sql:/docker-entrypoint-initdb.d/init.sql
  
  redis-test:
    image: redis:7-alpine
    ports:
      - "6380:6379"
  
  etcd-test:
    image: bitnami/etcd:latest
    environment:
      ALLOW_NONE_AUTHENTICATION: yes
    ports:
      - "2380:2379"
```

**启动测试环境**：
```bash
bash .harness/scripts/e2e-env-up.sh
# 1. docker-compose -f docker-compose.test.yml up -d
# 2. 等待服务就绪（健康检查）
# 3. 运行初始化脚本（创建测试数据）
```

#### 测试数据初始化

**文件**：`tests/e2e/fixtures/init.sql`

```sql
-- 创建测试用户
INSERT INTO users (id, username, password, role, created_at)
VALUES (1, 'admin', '$2a$10$...', 'admin', NOW()),
       (2, 'user1', '$2a$10$...', 'user', NOW());

-- 创建测试小区
INSERT INTO communities (id, name, address, created_at)
VALUES (1, '测试小区A', '北京市朝阳区xxx', NOW());

-- 其他测试数据...
```

**清理策略**：
- 每次测试前重置数据库（`TRUNCATE` + 重新初始化）
- 使用事务隔离（如果可能）
- 测试后自动清理

### 4. 关键场景定义

**选择标准**：
- 高业务价值（核心功能）
- 高风险（历史上出过问题）
- 跨服务依赖（集成复杂度高）

**推荐场景**（5 个）：

| # | 场景 | 覆盖服务 | 关键断言 |
|---|------|---------|---------|
| 1 | 用户登录 → 查看个人信息 | auth-service, user-service, web | JWT 有效，用户数据正确 |
| 2 | 发布活动 → 审核 → 报名 | community-hub, moderation, user | 活动状态流转正确 |
| 3 | 上传文件 → 预览 | file-service, web | 文件存储成功，URL 可访问 |
| 4 | 发布内容 → AI 审核 → 通过/拒绝 | moderation, ai-model, community-hub | 审核结果符合预期 |
| 5 | 紧急联络人设置 → 查询 | community-hub, user | 数据正确存储和返回 |

### 5. Harness 集成

**阶段 6 集成验证增强**：

```markdown
## 阶段 6: 集成归档

### 6.1 全链路编译验证（已有）
\`\`\`bash
cd $PROJECT_ROOT && go build ./...
go vet ./...
\`\`\`

### 6.2 运行时冒烟测试（已有，非阻塞）
\`\`\`bash
bash .harness/scripts/harness-smoke.sh
\`\`\`

### 6.3 E2E 关键场景验证（新增，有条件执行）
**触发条件**：
- 跨服务变更（≥2 个服务）
- 涉及 API 层或 RPC 接口变更
- 用户显式要求（`--run-e2e`）

**执行**：
\`\`\`bash
# 启动测试环境
bash .harness/scripts/e2e-env-up.sh

# 启动所有服务（测试模式）
bash scripts/start.sh --env=test

# 运行 E2E 测试（只运行与变更相关的场景）
bash .harness/scripts/run-e2e-tests.sh --filter-by-changes

# 结果
- PASS → 记录到 summary.md
- FAIL → 非阻塞，记录到「例外 & 未解决问题」，需人工确认
\`\`\`

**输出示例**：
\`\`\`
E2E 测试结果 (3/5 场景执行):
  ✅ 场景1: 用户登录 → 查看个人信息 (2.3s)
  ✅ 场景2: 发布活动 → 审核 → 报名 (5.1s)
  ❌ 场景4: 发布内容 → AI 审核 (失败: 审核服务超时)
  
总耗时: 8.7s
建议: 场景4 失败，可能是测试环境 AI 服务未启动，需人工验证
\`\`\`
```

## 实施步骤

### Phase 1: 基础设施（2 天）

**Task 1.1**: 搭建测试环境
- `docker-compose.test.yml` 定义
- 初始化脚本（init.sql + seed data）
- 启动/停止脚本（`e2e-env-up.sh`, `e2e-env-down.sh`）

**Task 1.2**: 集成 Playwright
- 安装：`cd web/pc && npm install -D @playwright/test`
- 配置：`playwright.config.ts`（base URL, timeout, 截图设置）
- 示例测试：登录流程

**Task 1.3**: 集成 Newman
- 安装：`npm install -g newman`
- 创建 Postman Collection 模板
- 示例测试：API 健康检查

### Phase 2: 核心场景实现（4 天）

**Task 2.1**: 实现场景 1-5（每个场景 ~0.5-1 天）
- 编写 Playwright 测试脚本
- 编写 Newman/Postman Collection
- 本地验证通过

**Task 2.2**: 测试数据管理
- Fixture 文件（初始数据）
- 清理脚本（测试后重置）
- 数据隔离策略

**Task 2.3**: 错误诊断增强
- 失败时自动截图 + 视频录制
- 失败时导出服务日志（`docker-compose logs`）
- 生成可读的失败报告

### Phase 3: CI/CD 集成（1 天）

**Task 3.1**: 添加 GitHub Actions workflow
- `.github/workflows/e2e-tests.yml`
- 触发条件：每日定时 + PR 标签 `run-e2e`
- 运行测试 + 上传结果

**Task 3.2**: Harness Pipeline 集成
- 阶段 6 增加 E2E 执行逻辑
- 过滤器：只运行相关场景
- 非阻塞模式：FAIL 记录但不回退

**Task 3.3**: 结果可视化
- HTML 报告生成（Playwright Reporter）
- 上传到 artifacts（CI）
- 链接到 summary.md

### Phase 4: 测试和文档（1 天）

**Task 4.1**: 回归测试
- 运行所有 E2E 场景 10 次
- 统计稳定性（通过率 ≥95%）

**Task 4.2**: 文档
- E2E 测试编写指南
- 故障排查手册
- 测试环境维护指南

**Task 4.3**: 培训
- 演示如何编写和调试 E2E 测试
- 演示如何本地运行测试

## 验收标准

### 功能验收

- [ ] 5 个核心场景全部实现
- [ ] 测试环境一键启动/停止
- [ ] 本地运行成功率 ≥95%
- [ ] CI 运行成功率 ≥90%（允许偶发网络问题）
- [ ] 失败时自动保存截图/日志

### 性能验收

| 指标 | 目标 | 实际 |
|------|------|------|
| 单个场景平均耗时 | <10 秒 | - |
| 5 个场景总耗时 | <1 分钟 | - |
| 测试环境启动时间 | <30 秒 | - |

### 质量验收

- [ ] 测试能捕获已知的集成问题（回归验证）
- [ ] 测试不依赖外部服务（自包含）
- [ ] 测试数据不污染开发环境

## 风险和依赖

### 风险

**R1: 测试脆弱**
- **描述**：测试频繁失败（flaky tests）
- **缓解**：
  - 使用稳定的定位器（避免 XPath，优先 `data-testid`）
  - 增加显式等待（`waitForSelector`）
  - 重试机制（Playwright 内置）

**R2: 测试环境维护成本**
- **描述**：测试数据陈旧、Docker 镜像过时
- **缓解**：
  - 自动化测试环境更新（定期重建镜像）
  - 版本化测试数据（Fixture 文件纳入 Git）

**R3: 执行时间长**
- **描述**：E2E 测试慢，影响开发体验
- **缓解**：
  - 并行执行（Playwright `--workers=3`）
  - 按需运行（只运行相关场景）
  - 分层运行（快速烟雾测试 + 完整回归）

### 依赖

**D1: 测试环境稳定性**
- 需要可靠的 Docker 环境
- 行动：在 CI 中使用 Docker-in-Docker 或 Testcontainers

**D2: 前后端联调环境**
- 前端需要后端 API 可访问
- 行动：确保测试环境网络配置正确

## 效果预估

### 质量提升

| 指标 | 改进前 | 改进后 | 提升 |
|------|-------|--------|------|
| 集成问题检出率 | ~30%（人工） | ~80%（自动） | +167% |
| 生产环境故障率 | 基线 | -30%（预期） | - |
| 修复周期 | 2 天（生产发现） | 0.5 天（测试发现） | ↓ 75% |

### 成本分析

| 项目 | 成本 | 说明 |
|------|------|------|
| 初期开发 | 8 人日 | 一次性投入 |
| 持续维护 | 1 人日/月 | 更新测试用例、修复 flaky tests |
| CI 运行时间 | +5 分钟/次 | E2E 测试执行时间 |

**ROI**：
- 节省生产故障损失：5 人日/次 × 3.6 次/年（30% 减少）= 18 人日/年
- 投资：8 人日 + 12 人日/年（维护）= 20 人日/年
- **首年持平，第二年起纯收益 18 人日/年**

## 后续优化

1. **视觉回归测试**：Playwright 截图对比，检测 UI 意外变化
2. **性能测试**：集成 Lighthouse，监控页面加载时间
3. **移动端测试**：Uni-app 项目的 E2E 测试（Appium）
4. **测试数据工厂**：动态生成测试数据，避免硬编码
5. **BDD 风格**：使用 Cucumber 编写可读的场景描述
