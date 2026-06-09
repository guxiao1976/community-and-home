# 多智能体协作开发规范

## 概述

本文档定义了主智能体、计划智能体、开发智能体、测试智能体之间的协作规范，确保开发流程高效、可追溯、高质量。

## 架构图

```
┌─────────────────────────────────────────────────────────────┐
│                        用户                                  │
│                          ↓                                   │
│                    需求文档路径                               │
└─────────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────────┐
│                    主智能体 (Main Agent)                      │
│  - 接收需求                                                   │
│  - 协调各智能体                                               │
│  - 管理任务状态                                               │
│  - 处理重试逻辑                                               │
└─────────────────────────────────────────────────────────────┘
         ↓                    ↓                    ↓
    ┌────────┐          ┌────────┐          ┌────────┐
    │ 计划   │          │ 开发   │          │ 测试   │
    │ 智能体 │          │ 智能体 │          │ 智能体 │
    └────────┘          └────────┘          └────────┘
         ↓                    ↓                    ↓
    任务计划            开发结果            测试结果
    (plan.md)      (dev-result.md)   (test-result.md)
```

## 工作流程

### 完整流程图

```
用户提供需求文档
    ↓
主智能体验证文档
    ↓
启动计划智能体 ──→ 生成任务计划 ──→ 返回计划路径
    ↓
主智能体读取任务列表
    ↓
┌─────────────────────────────────────────┐
│  任务执行循环（每个任务）                 │
│                                          │
│  1. 启动新的开发智能体                    │
│     ↓                                    │
│  2. 开发智能体完成开发 ──→ 返回结果路径   │
│     ↓                                    │
│  3. 启动新的测试智能体                    │
│     ↓                                    │
│  4. 测试智能体执行测试 ──→ 返回结果路径   │
│     ↓                                    │
│  5. 主智能体判断测试结果                  │
│     ├─ 通过 ──→ 更新状态，下一个任务      │
│     └─ 失败 ──→ 通知同一开发智能体修复    │
│                 ↓                        │
│              修复完成                     │
│                 ↓                        │
│              同一测试智能体重测            │
│                 ↓                        │
│              重复直到通过或达到重试上限    │
└─────────────────────────────────────────┘
    ↓
所有任务完成
    ↓
生成项目完成报告
    ↓
通知用户
```

## 通信协议

### 1. 主智能体 → 计划智能体

**请求：**
```json
{
  "action": "create_plan",
  "requirements_path": "/path/to/requirements.md",
  "project_root": "/home/jiaoxh/my-project/community-home"
}
```

**响应：**
```json
{
  "status": "success|error",
  "plan_path": "/path/to/plan.md",
  "total_tasks": 10,
  "estimated_hours": 40,
  "error": "错误信息（如果失败）"
}
```

### 2. 主智能体 → 开发智能体

**初始开发请求：**
```json
{
  "action": "develop",
  "task_id": "task-001",
  "task_title": "实现用户登录功能",
  "task_description": "详细描述...",
  "task_requirements": ["需求1", "需求2"],
  "acceptance_criteria": ["标准1", "标准2"],
  "related_files": ["/path/to/file.go"],
  "plan_path": "/path/to/plan.md",
  "estimated_hours": 3
}
```

**修复请求：**
```json
{
  "action": "fix",
  "task_id": "task-001",
  "test_result_path": "/path/to/test-result.md",
  "failure_summary": "简要描述失败原因",
  "retry_count": 1
}
```

**响应：**
```json
{
  "status": "success|error",
  "dev_result_path": "/path/to/dev-result.md",
  "changed_files": ["file1.go", "file2.go"],
  "new_files": 2,
  "modified_files": 1,
  "test_passed": true,
  "test_coverage": 85.7,
  "error": "错误信息（如果失败）"
}
```

### 3. 主智能体 → 测试智能体

**测试请求：**
```json
{
  "action": "test",
  "task_id": "task-001",
  "dev_result_path": "/path/to/dev-result.md",
  "is_retest": false,
  "previous_test_path": "/path/to/previous-test.md"
}
```

**响应：**
```json
{
  "status": "success|error",
  "test_result_path": "/path/to/test-result.md",
  "test_passed": true|false,
  "total_tests": 15,
  "passed_tests": 13,
  "failed_tests": 2,
  "pass_rate": 86.7,
  "critical_issues": 1,
  "summary": "简要总结",
  "error": "错误信息（如果失败）"
}
```

## 文件路径规范

### 目录结构

```
/home/jiaoxh/my-project/community-home/
└── doc/
    └── agents/
        ├── MAIN_AGENT.md           # 主智能体提示词
        ├── PLAN_AGENT.md           # 计划智能体提示词
        ├── DEV_AGENT.md            # 开发智能体提示词
        ├── TEST_AGENT.md           # 测试智能体提示词
        ├── WORKFLOW.md             # 本文档
        ├── task-status.json        # 任务状态文件
        ├── project-report.md       # 项目完成报告
        ├── plans/                  # 任务计划目录
        │   └── plan-YYYYMMDD-HHMMSS.md
        └── results/                # 结果文档目录
            ├── dev-result-task-001.md
            ├── test-result-task-001.md
            ├── test-result-task-001-v2.md  # 重测版本
            └── ...
```

### 文件命名规范

```yaml
计划文档:
  格式: plan-YYYYMMDD-HHMMSS.md
  示例: plan-20260527-100530.md
  位置: ./doc/agents/plans/

开发结果:
  格式: dev-result-{task_id}.md
  示例: dev-result-task-001.md
  位置: ./doc/agents/results/

测试结果:
  格式: test-result-{task_id}[-v{version}].md
  示例: test-result-task-001.md
        test-result-task-001-v2.md
  位置: ./doc/agents/results/

任务状态:
  文件名: task-status.json
  位置: ./doc/agents/

项目报告:
  文件名: project-report.md
  位置: ./doc/agents/
```

## 状态管理

### 任务状态文件格式

**文件：** `./doc/agents/task-status.json`

```json
{
  "project": "社区管理系统用户认证模块",
  "requirements_path": "/path/to/requirements.md",
  "plan_path": "/path/to/plan.md",
  "started_at": "2026-05-27T10:00:00Z",
  "updated_at": "2026-05-27T12:30:00Z",
  "status": "in_progress|completed|failed",
  "tasks": [
    {
      "id": "task-001",
      "title": "实现用户登录功能",
      "status": "pending|in_progress|testing|failed|completed",
      "priority": "high|medium|low",
      "estimated_hours": 3,
      "actual_hours": 2.5,
      "dev_agent_id": "agent-dev-001",
      "test_agent_id": "agent-test-001",
      "dev_result_path": "/path/to/dev-result-task-001.md",
      "test_result_path": "/path/to/test-result-task-001.md",
      "retry_count": 0,
      "max_retries": 3,
      "started_at": "2026-05-27T10:30:00Z",
      "completed_at": "2026-05-27T12:30:00Z",
      "issues": [
        {
          "description": "密码验证返回500",
          "severity": "high",
          "fixed": true
        }
      ]
    }
  ],
  "statistics": {
    "total_tasks": 10,
    "completed_tasks": 3,
    "failed_tasks": 0,
    "in_progress_tasks": 1,
    "pending_tasks": 6,
    "total_retries": 2,
    "average_retry_per_task": 0.2
  }
}
```

### 状态转换规则

```
pending (待处理)
  ↓ 启动开发智能体
in_progress (开发中)
  ↓ 开发完成，启动测试智能体
testing (测试中)
  ↓ 测试完成
  ├─ 测试通过 → completed (已完成)
  └─ 测试失败 → failed (失败)
      ↓ 通知开发智能体修复
      in_progress (修复中)
      ↓ 修复完成，重新测试
      testing (重测中)
      ↓ 重复直到通过或达到重试上限
      ├─ 通过 → completed
      └─ 达到上限 → failed (永久失败)
```

## 智能体管理规范

### 开发智能体

**创建规则：**
```yaml
时机: 每个新任务开始时
数量: 每个任务一个
隔离: worktree（独立工作空间）
模型: sonnet
生命周期: 任务完成后释放
```

**保持规则：**
```yaml
条件: 任务未完成
场景:
  - 开发中
  - 等待测试结果
  - 修复问题中
  - 等待重测结果
```

**释放规则：**
```yaml
条件: 任务完成或永久失败
操作:
  - 保存所有工作
  - 清理 worktree（如果没有变更）
  - 记录智能体 ID 到状态文件
```

### 测试智能体

**创建规则：**
```yaml
时机: 开发完成后
数量: 每个任务一个
隔离: 无（共享环境）
模型: sonnet
生命周期: 任务验收完成后释放
```

**保持规则：**
```yaml
条件: 任务未验收通过
场景:
  - 首次测试中
  - 等待修复
  - 重测中
```

**释放规则：**
```yaml
条件: 测试通过或永久失败
操作:
  - 保存测试结果
  - 记录智能体 ID 到状态文件
```

### 计划智能体

**创建规则：**
```yaml
时机: 项目开始时
数量: 每个项目一个
隔离: 无
模型: opus
生命周期: 计划完成后立即释放
```

## 错误处理规范

### 重试策略

```yaml
开发失败:
  最大重试次数: 3
  重试间隔: 无（立即重试）
  重试条件: 测试失败
  放弃条件: 达到最大重试次数

测试失败:
  最大重试次数: 无限制（跟随开发重试）
  重试间隔: 无（等待修复完成）
  重试条件: 开发修复完成
  放弃条件: 开发达到最大重试次数

计划失败:
  最大重试次数: 1
  重试间隔: 无
  重试条件: 需求文档问题
  放弃条件: 需求文档无法解析
```

### 错误升级

```yaml
开发智能体报告无法完成:
  1. 记录失败原因
  2. 询问用户：
     - 调整任务需求
     - 跳过该任务
     - 人工介入
  3. 根据用户选择继续

测试智能体报告无法测试:
  1. 检查是否环境问题
  2. 如果是环境问题，修复后重试
  3. 如果是测试用例问题，报告给用户

计划智能体报告无法规划:
  1. 检查需求文档是否存在
  2. 检查需求文档是否可读
  3. 报告给用户，请求澄清
```

## 质量标准

### 代码质量

```yaml
必须满足:
  - 通过 golangci-lint 检查
  - 单元测试通过率 100%
  - 测试覆盖率 > 75%
  - 无严重和高优先级问题

建议满足:
  - 测试覆盖率 > 80%
  - 代码复杂度合理
  - 性能达标
  - 无中优先级问题
```

### 文档质量

```yaml
开发结果文档:
  - 包含实现概述
  - 列出所有代码变更
  - 说明技术决策
  - 记录测试结果
  - 检查验收标准

测试结果文档:
  - 包含测试概要
  - 详细测试用例
  - 问题汇总
  - 修复建议
  - 测试结论
```

### 任务质量

```yaml
任务拆分:
  - 粒度合理（1-4小时）
  - 独立可测试
  - 依赖关系清晰
  - 验收标准明确
```

## 性能要求

### 时间限制

```yaml
计划智能体:
  超时: 10分钟
  预期: 5分钟内完成

开发智能体:
  超时: 30分钟
  预期: 按任务预估时间

测试智能体:
  超时: 10分钟
  预期: 5分钟内完成
```

### 并发控制

```yaml
开发智能体:
  最大并发: 1（串行执行任务）
  原因: 避免代码冲突

测试智能体:
  最大并发: 1（串行执行）
  原因: 共享测试环境

计划智能体:
  最大并发: 1
  原因: 每个项目一个计划
```

## 安全规范

### 敏感信息处理

```yaml
禁止在文档中记录:
  - 密码明文
  - API 密钥
  - 数据库连接字符串
  - 用户真实数据

允许记录:
  - 测试数据
  - 示例配置（脱敏）
  - 公开的 API 端点
```

### 权限控制

```yaml
开发智能体:
  可以: 读写代码、运行测试、查看日志
  不可以: 修改生产配置、访问生产数据

测试智能体:
  可以: 读代码、运行测试、查看日志
  不可以: 修改代码、修改配置

计划智能体:
  可以: 读需求文档、读项目文档
  不可以: 修改代码、修改配置
```

## 监控和日志

### 主智能体日志

```yaml
记录内容:
  - 任务开始/完成时间
  - 智能体启动/释放
  - 状态转换
  - 错误和重试
  - 用户交互

日志级别:
  - INFO: 正常流程
  - WARN: 重试、降级
  - ERROR: 失败、异常
```

### 子智能体日志

```yaml
记录内容:
  - 接收的任务
  - 执行的操作
  - 遇到的问题
  - 返回的结果

位置:
  - 开发结果文档
  - 测试结果文档
```

## 最佳实践

### 主智能体

1. **状态持久化** - 每次状态变更立即写入文件
2. **进度透明** - 实时向用户报告进度
3. **错误容忍** - 允许重试，但有上限
4. **智能体复用** - 同一任务的修复使用同一智能体
5. **路径传递** - 只传递文件路径，不传递大段内容

### 计划智能体

1. **任务独立** - 每个任务可独立开发和测试
2. **粒度合理** - 不要太大也不要太小
3. **依赖清晰** - 明确标注任务依赖关系
4. **标准明确** - 验收标准具体可测试
5. **风险识别** - 提前识别高风险任务

### 开发智能体

1. **测试先行** - 先写测试或同时写测试
2. **质量优先** - 代码质量比速度更重要
3. **文档完整** - 结果文档详细准确
4. **规范遵守** - 严格遵守项目规范
5. **问题定位** - 修复时先定位根本原因

### 测试智能体

1. **客观公正** - 基于事实，不偏袒
2. **详细记录** - 所有测试过程都要记录
3. **问题定位** - 不仅发现问题，还要定位原因
4. **建设性反馈** - 提供修复建议
5. **全面覆盖** - 功能、性能、安全都要测试

## 故障排查

### 常见问题

**问题 1：开发智能体无法访问文件**
```
原因: worktree 环境问题
解决: 检查 worktree 是否正确创建
```

**问题 2：测试智能体测试失败**
```
原因: 测试环境未准备好
解决: 确保数据库、Redis 等服务运行
```

**问题 3：任务状态文件损坏**
```
原因: 并发写入或异常中断
解决: 从备份恢复或重新开始
```

**问题 4：智能体超时**
```
原因: 任务过于复杂或网络问题
解决: 增加超时时间或拆分任务
```

## 版本历史

- v1.0 (2026-05-27) - 初始版本
