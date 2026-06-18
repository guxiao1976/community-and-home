# P2-2: 可观测性系统

## 背景

**现状问题**：
- **黑盒运行**：Harness 流水线执行过程不可见，只能看到最终结果
- **故障诊断困难**：失败时无法快速定位瓶颈（是 Generator 慢？还是 QA 失败？）
- **性能优化无数据支撑**：不知道哪个环节最耗时，无法针对性优化
- **无历史趋势**：无法回答"流水线性能是否在退化？"

**影响**：
- 故障排查时间长（需要逐个查看日志）
- 性能优化靠猜测（无数据驱动）
- 无法预警（问题发生后才知道）

## 目标

建立 **3 层可观测性体系**：
```
L1: Metrics（指标）   — 实时性能监控、告警
L2: Logging（日志）   — 详细执行记录、故障诊断
L3: Tracing（追踪）   — 端到端调用链分析
```

重点实现 **L1 Metrics + L2 Logging**，L3 Tracing 作为可选增强。

## 技术方案

### 1. 可观测性架构

```
Harness Pipeline
  ↓ emit metrics/logs
.harness/runtime/metrics.jsonl     ← 本地 metrics 文件
.harness/runtime/pipeline.log      ← 详细日志
  ↓ optional: export to
Prometheus + Grafana                ← 监控面板（可选）
```

**设计原则**：
- **本地优先**：metrics/logs 先写入本地文件，不依赖外部服务
- **渐进式**：从简单的文件输出开始，需要时再集成 Prometheus
- **低侵入**：不修改核心逻辑，通过 wrapper 注入监控

### 2. Metrics 设计（RED 方法）

#### RED 方法核心指标

**R - Rate（请求率）**：
- `harness_pipeline_starts_total` — Pipeline 启动次数
- `harness_agent_calls_total` — Agent 调用次数（按 label 分类：Generator/QA/Review）

**E - Errors（错误率）**：
- `harness_pipeline_failures_total` — Pipeline 失败次数（按原因分类）
- `harness_agent_failures_total` — Agent 失败次数

**D - Duration（耗时）**：
- `harness_pipeline_duration_seconds` — Pipeline 总耗时（histogram）
- `harness_phase_duration_seconds` — 各阶段耗时（Develop/QA/Review）
- `harness_agent_duration_seconds` — Agent 执行耗时

#### 扩展指标

**质量指标**：
- `harness_qa_pass_rate` — QA 首次通过率
- `harness_review_pass_count` — Review 通过数（按视角分类）
- `harness_iterations_count` — 修复轮次（histogram）

**资源指标**：
- `harness_token_usage_total` — Token 使用量（input/output 分开）
- `harness_memory_usage_bytes` — 记忆匹配数量
- `harness_confidence_score` — 置信度评分（histogram）

### 3. Metrics 收集实现

#### 方案 A: 文件输出（推荐，初期）

**格式**：JSON Lines（每行一个 metric）

**文件**：`.harness/runtime/metrics.jsonl`

**示例**：
```jsonl
{"timestamp":"2026-06-18T15:30:00Z","metric":"harness_pipeline_starts_total","value":1,"labels":{"service":"moderation-service","task_type":"feature"}}
{"timestamp":"2026-06-18T15:30:15Z","metric":"harness_agent_calls_total","value":1,"labels":{"agent":"generator","iteration":1}}
{"timestamp":"2026-06-18T15:32:45Z","metric":"harness_agent_duration_seconds","value":150,"labels":{"agent":"generator"}}
{"timestamp":"2026-06-18T15:33:00Z","metric":"harness_agent_calls_total","value":1,"labels":{"agent":"qa"}}
{"timestamp":"2026-06-18T15:35:30Z","metric":"harness_qa_pass_rate","value":1,"labels":{"service":"moderation-service"}}
{"timestamp":"2026-06-18T15:38:00Z","metric":"harness_pipeline_duration_seconds","value":480,"labels":{"service":"moderation-service","status":"pass"}}
```

**实现**（在 `harness-pipeline.js` 中注入）：

```javascript
// ── Metrics 收集器 ──
const fs = require('fs')
const METRICS_FILE = '.harness/runtime/metrics.jsonl'

function emitMetric(name, value, labels = {}) {
  const line = JSON.stringify({
    timestamp: new Date().toISOString(),
    metric: name,
    value: value,
    labels: labels
  })
  fs.appendFileSync(METRICS_FILE, line + '\n')
}

// ── 在关键点注入 ──
emitMetric('harness_pipeline_starts_total', 1, {service: args.serviceName, task_type: TASK_TYPE})

const startTime = Date.now()
await agent(generatorPrompt(...), {...})
const duration = (Date.now() - startTime) / 1000
emitMetric('harness_agent_duration_seconds', duration, {agent: 'generator', iteration: iteration})

if (qaResult.verdict === 'PASS') {
  emitMetric('harness_qa_pass_rate', 1, {service: args.serviceName})
} else {
  emitMetric('harness_qa_pass_rate', 0, {service: args.serviceName})
}
```

#### 方案 B: Prometheus Exporter（进阶）

**工具**：`prom-client`（Node.js）

**端点**：`http://localhost:9090/metrics`

**实现**：
```javascript
const promClient = require('prom-client')

const pipelineStarts = new promClient.Counter({
  name: 'harness_pipeline_starts_total',
  help: 'Total pipeline starts',
  labelNames: ['service', 'task_type']
})

const pipelineDuration = new promClient.Histogram({
  name: 'harness_pipeline_duration_seconds',
  help: 'Pipeline execution duration',
  labelNames: ['service', 'status'],
  buckets: [10, 30, 60, 120, 300, 600, 1800]  // 10s, 30s, ..., 30min
})

// 使用
pipelineStarts.inc({service: args.serviceName, task_type: TASK_TYPE})
const end = pipelineDuration.startTimer({service: args.serviceName})
// ... 执行 pipeline
end({status: 'pass'})
```

**Prometheus 配置**（`prometheus.yml`）：
```yaml
scrape_configs:
  - job_name: 'harness'
    static_configs:
      - targets: ['localhost:9090']
```

### 4. Logging 设计

#### 日志级别

| Level | 用途 | 示例 |
|-------|------|------|
| ERROR | 失败、异常 | QA 失败、Agent 超时 |
| WARN | 警告、降级 | Workflow 不可用降级到 Agent 模式 |
| INFO | 关键事件 | Pipeline 启动、阶段完成、最终结果 |
| DEBUG | 详细跟踪 | Prompt 内容、Agent 返回值 |

#### 日志格式

**格式**：结构化 JSON（便于解析和查询）

**文件**：`.harness/runtime/pipeline-<timestamp>.log`

**示例**：
```json
{"level":"INFO","timestamp":"2026-06-18T15:30:00Z","phase":"start","service":"moderation-service","task_type":"feature","message":"Pipeline started"}
{"level":"INFO","timestamp":"2026-06-18T15:30:15Z","phase":"develop","iteration":1,"agent":"generator","message":"Generator agent started"}
{"level":"DEBUG","timestamp":"2026-06-18T15:32:44Z","phase":"develop","agent":"generator","message":"Generator completed","prompt_length":1200,"response_length":850}
{"level":"INFO","timestamp":"2026-06-18T15:32:45Z","phase":"develop","agent":"generator","duration":150,"message":"Generator agent completed"}
{"level":"INFO","timestamp":"2026-06-18T15:33:00Z","phase":"qa","agent":"qa","message":"QA agent started"}
{"level":"INFO","timestamp":"2026-06-18T15:35:30Z","phase":"qa","verdict":"PASS","summary":"15项检查全部通过","message":"QA completed"}
{"level":"INFO","timestamp":"2026-06-18T15:38:00Z","phase":"complete","status":"pass","iterations":1,"confidence":0.85,"duration":480,"message":"Pipeline completed"}
```

**实现**（Logger 封装）：

```javascript
// .harness/workflows/utils/logger.js
const fs = require('fs')
const path = require('path')

class Logger {
  constructor(logFile) {
    this.logFile = logFile
    this.context = {}  // 共享上下文（service, task_type 等）
  }
  
  setContext(ctx) {
    this.context = {...this.context, ...ctx}
  }
  
  _write(level, message, extra = {}) {
    const entry = {
      level,
      timestamp: new Date().toISOString(),
      ...this.context,
      ...extra,
      message
    }
    fs.appendFileSync(this.logFile, JSON.stringify(entry) + '\n')
  }
  
  info(message, extra) { this._write('INFO', message, extra) }
  warn(message, extra) { this._write('WARN', message, extra) }
  error(message, extra) { this._write('ERROR', message, extra) }
  debug(message, extra) { this._write('DEBUG', message, extra) }
}

// 使用
const logger = new Logger('.harness/runtime/pipeline.log')
logger.setContext({service: args.serviceName, task_type: TASK_TYPE})

logger.info('Pipeline started', {phase: 'start'})
// ... 执行
logger.info('Generator completed', {phase: 'develop', agent: 'generator', duration: 150})
```

### 5. 监控面板设计

#### Grafana Dashboard（可选）

**面板 1: Pipeline 执行概览**
- 图表类型：Time series
- Metrics：
  - `rate(harness_pipeline_starts_total[5m])` — 启动速率
  - `rate(harness_pipeline_failures_total[5m])` — 失败速率
  - `histogram_quantile(0.95, harness_pipeline_duration_seconds)` — P95 耗时

**面板 2: 阶段性能分析**
- 图表类型：Heatmap
- Metrics：`harness_phase_duration_seconds{phase=~"develop|qa|review"}`
- 效果：可视化哪个阶段最耗时

**面板 3: QA 通过率**
- 图表类型：Gauge
- Metrics：`avg_over_time(harness_qa_pass_rate[1h])` — 过去 1 小时平均通过率

**面板 4: Token 使用趋势**
- 图表类型：Time series (stacked)
- Metrics：
  - `harness_token_usage_total{type="input"}`
  - `harness_token_usage_total{type="output"}`

#### 简易本地面板（推荐，初期）

**工具**：生成静态 HTML 报告

**脚本**：`bash .harness/scripts/metrics-report.sh`

```bash
#!/usr/bin/env bash
# 从 metrics.jsonl 生成 HTML 报告

METRICS_FILE=".harness/runtime/metrics.jsonl"
REPORT_FILE=".harness/runtime/metrics-report.html"

# 统计
TOTAL_RUNS=$(grep "harness_pipeline_starts_total" "$METRICS_FILE" | wc -l)
FAILURES=$(grep "harness_pipeline_failures_total" "$METRICS_FILE" | wc -l)
AVG_DURATION=$(grep "harness_pipeline_duration_seconds" "$METRICS_FILE" | jq -s 'map(.value) | add / length')

# 生成 HTML
cat > "$REPORT_FILE" <<EOF
<html>
<head><title>Harness Metrics Report</title></head>
<body>
  <h1>Harness Pipeline Metrics</h1>
  <h2>Summary (Last 24h)</h2>
  <ul>
    <li>Total Runs: $TOTAL_RUNS</li>
    <li>Failures: $FAILURES</li>
    <li>Success Rate: $(echo "scale=2; ($TOTAL_RUNS - $FAILURES) * 100 / $TOTAL_RUNS" | bc)%</li>
    <li>Average Duration: ${AVG_DURATION}s</li>
  </ul>
  <!-- 更多图表... -->
</body>
</html>
EOF

echo "Report generated: $REPORT_FILE"
```

### 6. 告警规则

#### 告警场景

| 场景 | 条件 | 级别 | 动作 |
|------|------|------|------|
| Pipeline 失败率高 | 过去 1 小时失败率 >30% | P1 | 通知 + 暂停自动调度 |
| 性能退化 | P95 耗时比前一天增加 50% | P2 | 通知 |
| QA 通过率低 | 过去 6 小时通过率 <50% | P2 | 通知 |
| Token 使用异常 | 单次 Pipeline 超过 100k tokens | P3 | 记录 |

#### 告警实现（简易版）

**脚本**：`bash .harness/scripts/check-alerts.sh`

```bash
#!/usr/bin/env bash
# 检查告警条件

METRICS_FILE=".harness/runtime/metrics.jsonl"
CUTOFF=$(date -u -d '1 hour ago' +"%Y-%m-%dT%H:%M:%SZ")

# 统计过去 1 小时的失败率
RECENT_METRICS=$(jq -s "map(select(.timestamp > \"$CUTOFF\"))" "$METRICS_FILE")
TOTAL=$(echo "$RECENT_METRICS" | jq '[.[] | select(.metric == "harness_pipeline_starts_total")] | length')
FAILURES=$(echo "$RECENT_METRICS" | jq '[.[] | select(.metric == "harness_pipeline_failures_total")] | length')

if [ "$TOTAL" -gt 0 ]; then
  FAILURE_RATE=$(echo "scale=2; $FAILURES * 100 / $TOTAL" | bc)
  if (( $(echo "$FAILURE_RATE > 30" | bc -l) )); then
    echo "🚨 ALERT: Pipeline failure rate is ${FAILURE_RATE}% (threshold: 30%)"
    # 发送通知（邮件、Slack、钉钉等）
    bash .harness/scripts/notify.sh "Pipeline failure rate alert: ${FAILURE_RATE}%"
  fi
fi
```

**定时执行**（cron）：
```bash
# 每 15 分钟检查一次
*/15 * * * * bash /path/to/.harness/scripts/check-alerts.sh
```

## 实施步骤

### Phase 1: Metrics 收集（2 天）

**Task 1.1**: 实现 Metrics 收集器
- 文件输出模式（`metrics.jsonl`）
- 注入到 `harness-pipeline.js` 关键点

**Task 1.2**: 定义核心 Metrics
- RED 指标（Rate/Errors/Duration）
- 质量指标（QA 通过率、Review 通过数、迭代次数）
- 资源指标（Token 使用、置信度）

**Task 1.3**: 数据保留策略
- 每日归档（移动到 `.harness/runtime/archive/<date>/`）
- 保留 30 天详细数据
- 超过 30 天 → 删除或压缩

### Phase 2: Logging 系统（1.5 天）

**Task 2.1**: 实现 Logger 类
- 结构化 JSON 输出
- 日志级别（ERROR/WARN/INFO/DEBUG）
- 上下文注入（service/task_type）

**Task 2.2**: 集成到 Pipeline
- 替换所有 `log()` 调用为 `logger.info()`
- 关键错误路径增加 `logger.error()`
- Debug 模式输出 Prompt 内容（`HARNESS_DEBUG=true`）

**Task 2.3**: 日志查询工具
- `bash .harness/scripts/logs-query.sh --level ERROR --since "1 hour ago"`
- 输出：过滤后的日志条目（JSON 或人类可读格式）

### Phase 3: 监控面板（2 天）

**Task 3.1**: 本地 HTML 报告
- `metrics-report.sh` 生成静态 HTML
- 包含：Summary 统计、耗时趋势图（Chart.js）、失败原因分布

**Task 3.2**: 告警脚本
- `check-alerts.sh` 检查告警条件
- 集成通知渠道（Slack webhook / 钉钉机器人）

**Task 3.3**: （可选）Prometheus + Grafana
- 安装 Prometheus（Docker）
- 配置 scrape targets
- 导入 Grafana Dashboard 模板

### Phase 4: 集成和文档（1 天）

**Task 4.1**: Owner Agent 集成
- 在 Pipeline 启动前初始化 Logger 和 Metrics
- 在 Pipeline 完成后生成报告

**Task 4.2**: 文档
- 可观测性架构说明
- Metrics 定义和查询方法
- 告警规则和响应流程

**Task 4.3**: 创建 Memory
- `.harness/knowledge/memory/observability-system.md`

## 验收标准

### 功能验收

- [ ] Metrics 正确记录到 `metrics.jsonl`
- [ ] 日志正确输出到 `pipeline.log`
- [ ] HTML 报告生成成功
- [ ] 告警脚本能检测异常情况
- [ ] 数据归档自动运行

### 数据完整性验收

- [ ] 每次 Pipeline 运行至少产生 10 条 metrics
- [ ] 日志包含所有关键事件（启动/完成/失败）
- [ ] Metrics 和日志的 timestamp 一致

### 性能验收

- [ ] Metrics 收集开销 <1% Pipeline 总耗时
- [ ] 日志文件大小增长可控（<10MB/天）

## 风险和依赖

### 风险

**R1: 日志/Metrics 文件过大**
- **描述**：长期运行导致文件膨胀
- **缓解**：
  - 每日归档和压缩
  - 定期清理（保留 30 天）
  - 可配置的日志级别（生产环境只记录 INFO 以上）

**R2: 监控数据不准确**
- **描述**：时间戳不同步、计数错误
- **缓解**：
  - 使用统一的时间源（`Date.now()`）
  - 单元测试验证计数逻辑
  - 定期审计（人工抽查）

**R3: 告警疲劳**
- **描述**：告警规则过于敏感，频繁误报
- **缓解**：
  - 设置合理阈值（基于历史数据）
  - 告警聚合（1 小时内相同告警只发一次）
  - 告警升级（连续 3 次才升级到 P1）

### 依赖

**D1: 文件系统可靠性**
- Metrics/Logs 依赖文件写入
- 行动：定期备份 `.harness/runtime/` 到外部存储

**D2: 时间同步**
- 分布式场景需要 NTP 时间同步
- 行动：验证系统时间设置

## 效果预估

### 诊断效率提升

| 场景 | 改进前 | 改进后 | 提升 |
|------|-------|--------|------|
| 定位 Pipeline 慢的原因 | 30 分钟（逐个查日志） | 5 分钟（查看 metrics） | ↓ 83% |
| 分析失败原因 | 15 分钟（搜索日志） | 2 分钟（ERROR 级别过滤） | ↓ 87% |
| 回答"过去一周性能趋势" | 无法回答 | 查看报告即知 | - |

### 运维成本

| 项目 | 成本 | 说明 |
|------|------|------|
| 开发成本 | 6.5 人日 | 一次性投入 |
| 存储成本 | ~1GB/年 | Metrics + Logs（30 天保留） |
| 维护成本 | 0.5 人日/月 | 监控面板更新、告警调优 |

**ROI**：
- 节省故障排查时间：8 人日/年（假设 10 次故障 × 省 0.8 人日/次）
- 投资：6.5 人日 + 6 人日/年 = 12.5 人日/年
- **首年收益 -4.5 人日，第二年起纯收益 8 人日/年**

## 后续优化

1. **分布式追踪（Tracing）**：集成 OpenTelemetry，可视化跨服务调用链
2. **实时监控大屏**：WebSocket 推送实时 metrics 到浏览器面板
3. **机器学习异常检测**：基于历史数据训练模型，自动识别异常模式
4. **根因分析自动化**：失败时自动关联 metrics/logs，生成根因报告
5. **SLO 监控**：定义 SLO（如 95% Pipeline 在 20 分钟内完成），监控达成率
