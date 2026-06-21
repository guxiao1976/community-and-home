// ============================================================
// Skills Integration Extension for Harness Pipeline
// Phase 1: 前端技能集成 (frontend-design, webapp-testing, audit-website)
// Phase 2: 流程管理技能集成 (writing-plans, requesting-code-review)
// Phase 3: 文档生成技能集成 (pdf, xlsx)
// ============================================================

// 检测是否为前端项目
function isFrontendProject(serviceDir) {
  return (serviceDir || '').startsWith('web/')
}

// Phase 1.5: UI Design (前端开发后)
// 在 Generator 完成后，如果是前端 feature 任务，调用 frontend-design 生成 UI
async function runUIDesignPhase(serviceName, serviceDir, taskDesc, taskType) {
  if (!isFrontendProject(serviceDir) || taskType !== 'feature') {
    log('⏭️ UI Design 跳过 (非前端或非 feature 任务)')
    return null
  }

  phase('UI Design')
  log(`🎨 启动 UI 设计阶段...`)

  try {
    const designPrompt = `你是前端 UI 设计专家。使用 frontend-design 技能为以下功能设计现代化的用户界面：

## 任务描述
${taskDesc}

## 设计要求
1. 使用 Vue 3 + TypeScript 语法
2. 响应式设计，支持桌面和移动端
3. 遵循现代 UI/UX 最佳实践
4. 确保可访问性 (WCAG 2.1 AA)
5. 组件化设计，可复用

## 输出
- Vue 3 单文件组件 (.vue)
- 组件使用说明
- 设计决策说明

请调用 /frontend-design 技能完成设计。`

    const designResult = await agent(designPrompt, {
      phase: 'UI Design',
      label: `UI Designer: ${serviceName}`,
      model: 'opus',  // UI 设计需要更强的创意能力
    })

    if (designResult) {
      log(`✅ UI 设计完成`)
      return designResult
    } else {
      log(`⚠️ UI 设计被跳过`)
      return null
    }
  } catch (error) {
    log(`⚠️ UI 设计失败: ${error.message}，继续流水线`)
    return null
  }
}

// Phase 2.5: E2E Testing (QA 后，前端项目)
// 在单元测试通过后，运行端到端测试
async function runE2ETestingPhase(serviceName, serviceDir, qaResult) {
  if (!isFrontendProject(serviceDir)) {
    log('⏭️ E2E Testing 跳过 (非前端项目)')
    return null
  }

  phase('E2E Testing')
  log(`🧪 启动 E2E 测试阶段...`)

  try {
    const e2ePrompt = `你是前端测试专家。使用 webapp-testing 技能执行端到端测试。

## 测试目标
服务: ${serviceName}
目录: ${serviceDir}

## QA 单元测试结果
${qaResult.summary}

## 测试要求
1. 测试关键用户流程（登录、核心功能操作、数据提交）
2. 验证 UI 交互正确性
3. 检查错误处理和边界条件
4. 跨浏览器兼容性检查 (Chrome, Firefox)

## 输出
- 测试用例列表和执行结果
- 失败用例的截图和日志
- 测试覆盖率报告

请调用 /webapp-testing 技能完成测试。`

    const e2eResult = await agent(e2ePrompt, {
      phase: 'E2E Testing',
      label: `E2E Tester: ${serviceName}`,
    })

    if (e2eResult) {
      log(`✅ E2E 测试完成`)
      return e2eResult
    } else {
      log(`⚠️ E2E 测试被跳过`)
      return null
    }
  } catch (error) {
    log(`⚠️ E2E 测试失败: ${error.message}，继续流水线`)
    return null
  }
}

// Phase 3.5: Frontend Audit (Review 后，前端项目)
// 在代码审查通过后，执行全面的前端质量审计
async function runFrontendAuditPhase(serviceName, serviceDir) {
  if (!isFrontendProject(serviceDir)) {
    log('⏭️ Frontend Audit 跳过 (非前端项目)')
    return null
  }

  phase('Frontend Audit')
  log(`🔍 启动前端质量审计...`)

  try {
    const auditPrompt = `你是前端质量审计专家。使用 audit-website 技能对前端应用进行全面审计。

## 审计目标
服务: ${serviceName}
目录: ${serviceDir}

## 审计维度
1. **性能** (Lighthouse Performance Score)
   - 首次内容绘制 (FCP)
   - 最大内容绘制 (LCP)
   - 累计布局偏移 (CLS)
   - 首次输入延迟 (FID)

2. **可访问性** (WCAG 2.1 AA 标准)
   - 键盘导航
   - 屏幕阅读器兼容性
   - 颜色对比度
   - ARIA 标签

3. **SEO**
   - Meta 标签完整性
   - 语义化 HTML
   - 移动端友好性

4. **最佳实践**
   - HTTPS
   - 控制台错误
   - 现代 JavaScript API

## 输出
- Lighthouse 审计报告 (JSON + HTML)
- 问题清单和优化建议
- 优先级排序

请调用 /audit-website 技能完成审计。启动本地开发服务器并审计。`

    const auditResult = await agent(auditPrompt, {
      phase: 'Frontend Audit',
      label: `Quality Auditor: ${serviceName}`,
    })

    if (auditResult) {
      log(`✅ 前端审计完成`)
      return auditResult
    } else {
      log(`⚠️ 前端审计被跳过`)
      return null
    }
  } catch (error) {
    log(`⚠️ 前端审计失败: ${error.message}，继续流水线`)
    return null
  }
}

// 将技能结果整合到最终报告
function integrateSkillsResults(baseResult, uiDesign, e2eTesting, frontendAudit) {
  const skillsReport = {
    uiDesign: uiDesign ? 'completed' : 'skipped',
    e2eTesting: e2eTesting ? 'completed' : 'skipped',
    frontendAudit: frontendAudit ? 'completed' : 'skipped',
  }

  return {
    ...baseResult,
    skillsReport,
    enhancedSummary: `${baseResult.reviewSummary} | Skills: UI=${skillsReport.uiDesign}, E2E=${skillsReport.e2eTesting}, Audit=${skillsReport.frontendAudit}`,
  }
}

// ============================================================
// Phase 2: 流程管理技能集成
// ============================================================

// Phase 2.1: Task Planning (架构设计后)
// 使用 writing-plans 将设计文档分解为精细化的任务列表
async function runTaskPlanningPhase(serviceName, designDoc, tasksDoc) {
  phase('Task Planning')
  log(`📋 启动任务规划阶段...`)

  try {
    const planningPrompt = `你是项目管理专家。使用 writing-plans 技能将架构设计分解为精细化的开发任务。

## 输入材料

### 架构设计文档
${designDoc}

### 初步任务列表
${tasksDoc}

## 任务分解要求

1. **结构化分解**
   - 将每个任务分解为可独立完成的子任务
   - 识别任务间的依赖关系
   - 标注阻塞关系（blockedBy）

2. **优先级排序**
   - P0: 阻塞性任务（必须先完成）
   - P1: 核心功能任务
   - P2: 增强功能任务
   - P3: 优化和改进

3. **工作量估算**
   - 估算每个任务的开发时间
   - 标注技术难度（Easy/Medium/Hard）
   - 识别风险点

4. **并行度分析**
   - 识别可以并行开发的任务
   - 标注独立任务组
   - 优化关键路径

## 输出格式

为每个任务输出：
- 任务 ID
- 任务标题（简洁、可操作）
- 详细描述
- 优先级（P0-P3）
- 预估工时
- 技术难度
- 依赖任务（如有）
- 验收标准

请调用 /writing-plans 技能完成任务规划。`

    const planningResult = await agent(planningPrompt, {
      phase: 'Task Planning',
      label: `Task Planner: ${serviceName}`,
      model: 'opus',  // 需要强推理能力
    })

    if (planningResult) {
      log(`✅ 任务规划完成`)
      return planningResult
    } else {
      log(`⚠️ 任务规划被跳过`)
      return null
    }
  } catch (error) {
    log(`⚠️ 任务规划失败: ${error.message}，继续流水线`)
    return null
  }
}

// Phase 2.2: Code Review Request (Review 阶段)
// 使用 requesting-code-review 生成标准化的审查请求
async function runCodeReviewRequestPhase(serviceName, serviceDir, qaResult, reviewResults) {
  phase('Code Review Request')
  log(`📝 生成代码审查请求...`)

  try {
    const reviewRequestPrompt = `你是代码审查协调员。使用 requesting-code-review 技能为本次变更生成标准化的审查请求。

## 变更概览

**服务**: ${serviceName}
**目录**: ${serviceDir}

## QA 结果
${qaResult.summary}

## Review 结果
${reviewResults.map(r => `${r.label}: ${r.verdict} - ${r.summary}`).join('\n')}

## 审查请求要求

1. **变更摘要**
   - 本次变更的目的和范围
   - 影响的文件和模块
   - 关键技术决策

2. **审查重点**
   - 需要特别关注的代码段
   - 潜在风险点
   - 性能/安全性考虑

3. **测试覆盖**
   - 单元测试情况
   - E2E 测试结果（如有）
   - 手工测试清单

4. **审查清单**
   - 功能正确性
   - 代码质量
   - 安全性
   - 性能影响
   - 可维护性

5. **部署计划**
   - 部署步骤
   - 回滚方案
   - 监控指标

## 输出

生成 GitHub PR 描述格式的审查请求，包含：
- 标题（简洁，< 70 字符）
- 变更说明
- 审查清单
- 截图/示例（如适用）
- 相关 Issue/文档链接

请调用 /requesting-code-review 技能完成审查请求生成。`

    const reviewRequestResult = await agent(reviewRequestPrompt, {
      phase: 'Code Review Request',
      label: `Review Coordinator: ${serviceName}`,
    })

    if (reviewRequestResult) {
      log(`✅ 审查请求已生成`)
      return reviewRequestResult
    } else {
      log(`⚠️ 审查请求生成被跳过`)
      return null
    }
  } catch (error) {
    log(`⚠️ 审查请求生成失败: ${error.message}，继续流水线`)
    return null
  }
}

// 整合 Phase 2 结果
function integratePhase2Results(baseResult, taskPlanning, codeReviewRequest) {
  const phase2Report = {
    taskPlanning: taskPlanning ? 'completed' : 'skipped',
    codeReviewRequest: codeReviewRequest ? 'completed' : 'skipped',
  }

  return {
    ...baseResult,
    phase2Report,
  }
}

// ============================================================
// Phase 3: 文档生成技能集成
// ============================================================

// Phase 3.1: Test Report Generation (QA 阶段后)
// 使用 xlsx 生成数据化的测试报告
async function runTestReportGenerationPhase(serviceName, serviceDir, qaResult, e2eResult) {
  phase('Test Report')
  log(`📊 生成测试报告...`)

  try {
    const reportPrompt = `你是测试报告专家。使用 xlsx 技能生成数据化的测试报告。

## 服务信息
**服务**: ${serviceName}
**目录**: ${serviceDir}

## QA 测试结果
${qaResult.summary}

详细信息:
- 测试判定: ${qaResult.verdict}
- 构建状态: ${qaResult.buildStatus || 'N/A'}
- 测试覆盖率: ${qaResult.coverage || 'N/A'}
${e2eResult ? `\n## E2E 测试结果\n${e2eResult}` : ''}

## 报告要求

生成包含以下工作表的 Excel 测试报告：

### 工作表1: 测试概览
- 服务名称和版本
- 测试日期时间
- 测试判定（PASS/FAIL）
- 总体覆盖率
- 测试用例统计（总数、通过、失败、跳过）

### 工作表2: 单元测试明细
表格列：
- 测试包/模块
- 测试用例名称
- 测试结果（✓/✗）
- 执行时间
- 覆盖率
- 失败原因（如有）

### 工作表3: E2E 测试明细（如有）
表格列：
- 测试场景
- 测试步骤
- 预期结果
- 实际结果
- 测试结果（✓/✗）
- 截图路径（如有）

### 工作表4: 覆盖率分析
- 按包/模块统计覆盖率
- 未覆盖代码行数
- 覆盖率趋势图（如有历史数据）

### 工作表5: 问题清单（如有失败）
- 问题描述
- 影响范围
- 严重程度
- 责任人
- 状态

## 格式要求
- 使用专业的 Excel 格式
- 表头加粗、背景色
- PASS 用绿色，FAIL 用红色
- 数据对齐和格式化
- 添加数据验证和条件格式

请调用 /xlsx 技能生成测试报告。`

    const reportResult = await agent(reportPrompt, {
      phase: 'Test Report',
      label: `Report Generator: ${serviceName}`,
    })

    if (reportResult) {
      log(`✅ 测试报告已生成`)
      return reportResult
    } else {
      log(`⚠️ 测试报告生成被跳过`)
      return null
    }
  } catch (error) {
    log(`⚠️ 测试报告生成失败: ${error.message}，继续流水线`)
    return null
  }
}

// Phase 3.2: Delivery Documentation (流水线完成后)
// 使用 pdf 生成交付文档包
async function runDeliveryDocGenerationPhase(serviceName, serviceDir, allResults) {
  phase('Delivery Documentation')
  log(`📄 生成交付文档...`)

  try {
    const docPrompt = `你是技术文档专家。使用 pdf 技能生成专业的交付文档包。

## 服务信息
**服务**: ${serviceName}
**目录**: ${serviceDir}

## 流水线结果
${JSON.stringify(allResults, null, 2)}

## 交付文档要求

生成一份完整的 PDF 交付文档，包含以下章节：

### 封面
- 服务名称
- 版本号
- 交付日期
- 项目 Logo

### 目录
自动生成章节目录，带页码

### 第一章: 变更概览
- 本次交付的功能/修复
- 影响范围
- 关键技术决策
- 已知限制

### 第二章: 架构设计
- 系统架构图（如有）
- 模块划分
- 数据流程
- 接口定义

### 第三章: API 文档
- RESTful API 清单
- gRPC 接口清单
- 请求/响应示例
- 错误码说明

### 第四章: 数据库变更
- 表结构变更
- 索引变更
- 数据迁移脚本
- 回滚方案

### 第五章: 测试报告
- QA 测试结果
- 单元测试覆盖率
- E2E 测试场景
- 性能测试数据（如有）

### 第六章: 代码审查报告
- 安全架构审查结果
- 规范工程审查结果
- 设计业务审查结果
- 遗留问题和改进建议

### 第七章: 部署指南
- 部署前置条件
- 部署步骤详解
- 配置说明
- 验证步骤
- 回滚方案

### 第八章: 监控和维护
- 关键监控指标
- 告警规则
- 常见问题和解决方案
- 运维联系方式

### 附录
- 相关文档链接
- 变更历史
- 术语表

## 格式要求
- 专业的 PDF 排版
- 清晰的章节层次
- 代码块语法高亮
- 图表清晰可读
- 页眉页脚（服务名、页码）
- 书签导航

请调用 /pdf 技能生成交付文档。`

    const docResult = await agent(docPrompt, {
      phase: 'Delivery Documentation',
      label: `Doc Generator: ${serviceName}`,
      model: 'opus',  // 需要强文档生成能力
    })

    if (docResult) {
      log(`✅ 交付文档已生成`)
      return docResult
    } else {
      log(`⚠️ 交付文档生成被跳过`)
      return null
    }
  } catch (error) {
    log(`⚠️ 交付文档生成失败: ${error.message}，继续流水线`)
    return null
  }
}

// 整合 Phase 3 结果
function integratePhase3Results(baseResult, testReport, deliveryDoc) {
  const phase3Report = {
    testReport: testReport ? 'completed' : 'skipped',
    deliveryDoc: deliveryDoc ? 'completed' : 'skipped',
  }

  return {
    ...baseResult,
    phase3Report,
  }
}

// 导出技能集成函数
export {
  // Phase 1: 前端技能
  isFrontendProject,
  runUIDesignPhase,
  runE2ETestingPhase,
  runFrontendAuditPhase,
  integrateSkillsResults,

  // Phase 2: 流程管理技能
  runTaskPlanningPhase,
  runCodeReviewRequestPhase,
  integratePhase2Results,

  // Phase 3: 文档生成技能
  runTestReportGenerationPhase,
  runDeliveryDocGenerationPhase,
  integratePhase3Results,
}
