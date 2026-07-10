# 问题 5 & 6 实施路线图

## 📋 执行计划

由于问题 5 和 6 需要较长时间（3-5 天），建议分阶段实施。

---

## 🎯 问题 5：添加确定性验证层

### 阶段 1：基础确定性检查（4-6 小时）

**目标**：在 AI 判断前先运行确定性检查

**实施步骤**：

1. **创建确定性验证脚本**
   ```bash
   .harness/scripts/deterministic-checks.sh
   ```
   
   **检查项**：
   - ✅ 编译成功（go build）
   - ✅ 测试通过（go test）
   - ✅ 测试覆盖率 ≥ 阈值
   - ✅ 静态分析无错误（go vet）
   - ✅ 代码格式正确（gofmt）
   - ✅ 依赖检查（go mod verify）

2. **修改 Harness Pipeline**
   ```javascript
   // 在 Generator 之前运行
   const deterministicResult = await deterministicChecks(serviceDir)
   
   if (!deterministicResult.passed) {
     log(`❌ 确定性检查失败，跳过 AI 生成`)
     return { status: 'FAIL', reason: 'deterministic-checks-failed' }
   }
   
   // 确定性检查通过，继续 AI 判断
   const aiResult = await runGenerator(...)
   ```

3. **添加验证规则**
   ```yaml
   # .harness/config/deterministic-rules.yml
   compile:
     required: true
     blocker: true
   
   tests:
     required: true
     blocker: true
     min_coverage: 80
   
   static_analysis:
     required: true
     blocker: false  # 警告但不阻塞
   ```

**预期效果**：
- AI 误判 FAIL 但确定性检查 PASS → 提示人工审查
- AI 误判 PASS 但确定性检查 FAIL → 自动 FAIL

---

### 阶段 2：AI 判断结果验证（4-6 小时）

**目标**：验证 AI 判断的合理性

**实施步骤**：

1. **创建 AI 判断验证器**
   ```javascript
   // .harness/validators/ai-judgment-validator.js
   
   function validateAIJudgment(aiResult, deterministicResult) {
     const conflicts = []
     
     // 冲突 1：AI 说 PASS，但编译失败
     if (aiResult.status === 'PASS' && !deterministicResult.compile) {
       conflicts.push({
         type: 'ai-pass-compile-fail',
         severity: 'critical',
         action: 'override-to-fail'
       })
     }
     
     // 冲突 2：AI 说 FAIL，但所有确定性检查通过
     if (aiResult.status === 'FAIL' && deterministicResult.allPassed) {
       conflicts.push({
         type: 'possible-false-negative',
         severity: 'warning',
         action: 'human-review'
       })
     }
     
     return conflicts
   }
   ```

2. **集成到 Pipeline**
   ```javascript
   const conflicts = validateAIJudgment(aiResult, deterministicResult)
   
   if (conflicts.some(c => c.action === 'override-to-fail')) {
     log(`⚠️  AI 判断被确定性检查推翻`)
     return { status: 'FAIL', overridden: true }
   }
   
   if (conflicts.some(c => c.action === 'human-review')) {
     log(`⚠️  检测到可能的误判，建议人工审查`)
     // 发送通知或标记 PR
   }
   ```

**预期效果**：
- 减少 AI 误判的影响
- 提供人工审查提示

---

### 阶段 3：判断结果追踪（2-3 小时）

**目标**：收集数据，优化判断准确率

**实施步骤**：

1. **记录判断结果**
   ```javascript
   // .harness/logs/judgments/
   {
     "timestamp": "2026-07-10T20:30:00Z",
     "service": "auth-service",
     "deterministic": {
       "compile": true,
       "tests": true,
       "coverage": 85,
       "static_analysis": true
     },
     "ai_judgment": "PASS",
     "conflicts": [],
     "final_result": "PASS",
     "human_override": null
   }
   ```

2. **生成分析报告**
   ```bash
   bash .harness/scripts/analyze-judgments.sh
   
   # 输出：
   # - AI 准确率：95%
   # - 误判类型分布
   # - 推翻次数统计
   ```

---

## 🚀 问题 6：部署和回滚

### 阶段 1：集成测试（6-8 小时）

**目标**：添加端到端测试阶段

**实施步骤**：

1. **创建集成测试框架**
   ```bash
   .harness/tests/integration/
   ├── setup.sh           # 启动测试环境
   ├── test-auth-flow.sh  # 认证流程测试
   ├── test-user-crud.sh  # 用户 CRUD 测试
   └── teardown.sh        # 清理环境
   ```

2. **集成测试脚本**
   ```bash
   #!/bin/bash
   # test-auth-flow.sh
   
   # 1. 启动服务（Docker Compose）
   docker-compose -f test-compose.yml up -d
   
   # 2. 等待服务就绪
   wait_for_service "auth-service:8080/health"
   
   # 3. 执行测试
   curl -X POST http://localhost:8080/api/auth/register \
     -d '{"username":"test","password":"test123"}'
   
   # 4. 验证结果
   assert_http_status 200
   assert_json_field "token" exists
   ```

3. **添加到 Pipeline**
   ```javascript
   // After QA + Review pass
   log('Running integration tests...')
   
   const integrationResult = await agent(
     'Run integration tests for auth-service',
     { schema: INTEGRATION_TEST_SCHEMA }
   )
   
   if (!integrationResult.passed) {
     log('❌ Integration tests failed')
     return { status: 'FAIL', stage: 'integration' }
   }
   ```

---

### 阶段 2：Docker 镜像构建（4-6 小时）

**目标**：构建可部署的 Docker 镜像

**实施步骤**：

1. **创建多阶段 Dockerfile**
   ```dockerfile
   # services/auth-service/Dockerfile
   
   # Stage 1: Build
   FROM golang:1.21 AS builder
   WORKDIR /app
   COPY go.mod go.sum ./
   RUN go mod download
   COPY . .
   RUN CGO_ENABLED=0 go build -o auth-service ./api
   
   # Stage 2: Runtime
   FROM alpine:3.18
   RUN apk --no-cache add ca-certificates
   COPY --from=builder /app/auth-service /app/
   COPY --from=builder /app/etc /app/etc
   EXPOSE 8080
   CMD ["/app/auth-service", "-f", "/app/etc/config.yaml"]
   ```

2. **构建脚本**
   ```bash
   # .harness/scripts/build-docker-image.sh
   
   SERVICE=$1
   VERSION=$(git describe --tags --always)
   
   docker build -t "registry.example.com/${SERVICE}:${VERSION}" \
     -f "services/${SERVICE}/Dockerfile" \
     "services/${SERVICE}"
   
   docker push "registry.example.com/${SERVICE}:${VERSION}"
   ```

3. **集成到 Pipeline**
   ```javascript
   log('Building Docker image...')
   
   await bash(`.harness/scripts/build-docker-image.sh ${serviceName}`)
   
   const imageTag = await getGitTag()
   log(`✅ Image built: ${serviceName}:${imageTag}`)
   ```

---

### 阶段 3：部署到测试环境（6-8 小时）

**目标**：自动部署到测试环境并验证

**实施步骤**：

1. **创建部署脚本**
   ```bash
   # .harness/scripts/deploy-to-test.sh
   
   SERVICE=$1
   VERSION=$2
   
   # 1. 更新 Kubernetes manifest
   kubectl set image deployment/${SERVICE} \
     ${SERVICE}=registry.example.com/${SERVICE}:${VERSION} \
     -n test
   
   # 2. 等待 rollout 完成
   kubectl rollout status deployment/${SERVICE} -n test
   
   # 3. 健康检查
   for i in {1..30}; do
     if curl -f http://test.example.com/${SERVICE}/health; then
       echo "✅ Service healthy"
       exit 0
     fi
     sleep 10
   done
   
   echo "❌ Health check failed"
   exit 1
   ```

2. **集成到 Pipeline**
   ```javascript
   log('Deploying to test environment...')
   
   const deployResult = await bash(
     `.harness/scripts/deploy-to-test.sh ${serviceName} ${version}`
   )
   
   if (deployResult.exitCode !== 0) {
     log('❌ Deployment failed')
     await rollback(serviceName)
     return { status: 'FAIL', stage: 'deploy' }
   }
   ```

---

### 阶段 4：自动回滚（4-6 小时）

**目标**：部署失败时自动回滚

**实施步骤**：

1. **记录部署历史**
   ```json
   // .harness/deployments/history.json
   {
     "auth-service": [
       {
         "version": "v1.2.3",
         "timestamp": "2026-07-10T20:00:00Z",
         "status": "success",
         "health_check": true
       },
       {
         "version": "v1.2.4",
         "timestamp": "2026-07-10T21:00:00Z",
         "status": "failed",
         "health_check": false,
         "rolled_back_to": "v1.2.3"
       }
     ]
   }
   ```

2. **回滚脚本**
   ```bash
   # .harness/scripts/rollback.sh
   
   SERVICE=$1
   TARGET_VERSION=${2:-previous}
   
   if [ "$TARGET_VERSION" = "previous" ]; then
     # 获取上一个成功版本
     TARGET_VERSION=$(jq -r \
       ".\"${SERVICE}\" | .[] | select(.status==\"success\") | .version" \
       .harness/deployments/history.json | head -1)
   fi
   
   echo "Rolling back ${SERVICE} to ${TARGET_VERSION}..."
   
   kubectl set image deployment/${SERVICE} \
     ${SERVICE}=registry.example.com/${SERVICE}:${TARGET_VERSION} \
     -n test
   
   kubectl rollout status deployment/${SERVICE} -n test
   ```

3. **自动回滚逻辑**
   ```javascript
   async function deployWithRollback(service, version) {
     // 记录当前版本
     const currentVersion = await getCurrentVersion(service)
     
     try {
       // 尝试部署
       await deploy(service, version)
       
       // 健康检查
       const healthy = await healthCheck(service, { timeout: 300 })
       
       if (!healthy) {
         throw new Error('Health check failed')
       }
       
       // 记录成功部署
       await recordDeployment(service, version, 'success')
       
       return { status: 'SUCCESS' }
       
     } catch (error) {
       log(`❌ Deployment failed: ${error.message}`)
       log(`🔄 Rolling back to ${currentVersion}...`)
       
       // 自动回滚
       await rollback(service, currentVersion)
       
       // 记录失败
       await recordDeployment(service, version, 'failed', {
         rolled_back_to: currentVersion,
         error: error.message
       })
       
       return { status: 'FAIL', rolled_back: true }
     }
   }
   ```

---

## 📊 实施优先级

### 推荐顺序：

1. **问题 5 - 阶段 1**：确定性检查（4-6 小时）✅ 高价值，快速完成
2. **问题 5 - 阶段 2**：AI 判断验证（4-6 小时）✅ 减少误判
3. **问题 6 - 阶段 1**：集成测试（6-8 小时）✅ 质量保证
4. **问题 6 - 阶段 2**：Docker 构建（4-6 小时）✅ 部署基础
5. **问题 6 - 阶段 3**：部署测试环境（6-8 小时）✅ 自动化部署
6. **问题 5 - 阶段 3**：判断追踪（2-3 小时）📊 数据分析
7. **问题 6 - 阶段 4**：自动回滚（4-6 小时）🔄 风险控制

**总计时间**：30-43 小时 ≈ **4-5 天**

---

## 🎯 快速启动指南

### 第 1 天：确定性验证

```bash
# 创建脚本
vim .harness/scripts/deterministic-checks.sh

# 修改 Pipeline
vim .harness/workflows/harness-pipeline-core.js

# 测试
bash .harness/scripts/deterministic-checks.sh services/auth-service
```

### 第 2 天：AI 判断验证

```bash
# 创建验证器
vim .harness/validators/ai-judgment-validator.js

# 集成到 Pipeline
# 测试冲突检测
```

### 第 3 天：集成测试

```bash
# 创建测试框架
mkdir -p .harness/tests/integration
vim .harness/tests/integration/test-auth-flow.sh

# 运行测试
bash .harness/tests/integration/test-auth-flow.sh
```

### 第 4 天：Docker + 部署

```bash
# 创建 Dockerfile
vim services/auth-service/Dockerfile

# 构建镜像
bash .harness/scripts/build-docker-image.sh auth-service

# 部署测试
bash .harness/scripts/deploy-to-test.sh auth-service v1.2.3
```

### 第 5 天：回滚 + 测试

```bash
# 创建回滚脚本
vim .harness/scripts/rollback.sh

# 测试回滚
bash .harness/scripts/rollback.sh auth-service previous

# 端到端测试
```

---

## 📋 检查清单

### 问题 5
- [ ] 确定性检查脚本
- [ ] 验证规则配置
- [ ] AI 判断验证器
- [ ] Pipeline 集成
- [ ] 冲突检测测试
- [ ] 判断结果记录
- [ ] 分析报告生成

### 问题 6
- [ ] 集成测试框架
- [ ] 测试用例（3-5 个）
- [ ] Dockerfile（每个服务）
- [ ] 镜像构建脚本
- [ ] 部署脚本
- [ ] 健康检查
- [ ] 回滚脚本
- [ ] 部署历史记录
- [ ] 端到端测试

---

**文档生成时间**：2026-07-10 20:35 UTC  
**预计完成时间**：4-5 天  
**建议开始时间**：下次会话
