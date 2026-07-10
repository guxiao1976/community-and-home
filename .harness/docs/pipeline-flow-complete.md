# Harness Pipeline 完整流程图

> 从需求到交付的端到端流程 · 工具调用链路详解 · 最后更新 2026-06-22

---

## 概述

本文档详细展示一个需求从接收到最终交付的完整流程，包括：
- 每个阶段调用的脚本/Agent/工具
- 调用条件和判断逻辑
- 文件读写和数据流转
- 决策点和回退路径

---

## 案例需求

**需求**: 在社区枢纽服务新增"紧急联络人"功能，涉及后端 API + 前端管理页面。

**涉及服务**:
- `services/community-hub-service` (后端 API + RPC)
- `web/pc` (管理后台前端)

**技术栈**:
- Proto 定义（需新增消息类型）
- Go 后端实现
- Vue 3 前端实现

---

## 完整流程概览

```
用户输入需求
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 0: 路径选择 (Owner Agent 内联执行)                        │
│ ├─ 读取: CLAUDE.md, owner-agent.md                            │
│ ├─ 判断: 跨服务? 涉及Proto? 需求明确?                          │
│ └─ 输出: request.md (路径=OpenSpec)                           │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 1: 需求分析 (子Agent: requirement-analyst)                │
│ ├─ 工具: Agent({subagent_type: "general-purpose"})            │
│ ├─ 读取: CLAUDE.md, design.md, graph-context.md              │
│ ├─ 输出: proposal.md + specs/emergency-contact-spec.md       │
│ └─ 门禁: 追溯表全✅ + Self-Review PASS                         │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 2: 需求评审 (3个子Agent并行)                              │
│ ├─ 工具: 3 × Agent({subagent_type: "general-purpose"})       │
│ │   ├─ coverage-reviewer (覆盖性)                             │
│ │   ├─ structure-reviewer (结构性)                            │
│ │   └─ clarity-reviewer (清晰度)                              │
│ ├─ 读取: proposal.md, specs/*.md, review.md                  │
│ ├─ 输出: review/spec_review_coverage_v1.md (×3)              │
│ └─ 门禁: 2/3 APPROVED                                         │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 3: 架构设计 (子Agent: architecture-designer)              │
│ ├─ 工具: Agent({subagent_type: "general-purpose"})            │
│ ├─ 读取: proposal.md, specs/*.md, api-proto/, design.md      │
│ ├─ 写入: design.md (更新), tasks.md                          │
│ └─ 门禁: 记忆注入完成 + 零占位符 + TDD步骤明确                 │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 4: Proto 变更 (Owner Agent 内联执行)                      │
│ ├─ 读取: tasks.md (提取Proto变更部分)                         │
│ ├─ 修改: api-proto/community/v1/emergency_contact.proto      │
│ ├─ 脚本: cd api-proto && make ci                             │
│ │   ├─ make lint (buf lint)                                  │
│ │   ├─ make breaking-check (buf breaking)                    │
│ │   └─ make generate (protoc + buf generate)                 │
│ └─ 门禁: lint PASS + breaking-check PASS                      │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 5: 编码+测试 (2个Workflow并行)                            │
│                                                               │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Workflow 1: community-hub-service                        │ │
│ │ ├─ 工具: Workflow({scriptPath: "harness-pipeline.js"}) │ │
│ │ ├─ 参数: {serviceName, serviceDir, task: "Task 1.1-1.5"}│ │
│ │ └─ 内部循环 ↓                                            │ │
│ └─────────────────────────────────────────────────────────┘ │
│      ↓                                                        │
│   ┌─────────────────────────────────────────────────────┐   │
│   │ Iteration 1:                                         │   │
│   │   Phase: Develop                                    │   │
│   │   ├─ Agent: Generator (TDD + Memory)               │   │
│   │   │   ├─ 读取: CLAUDE.md, design.md, tasks.md     │   │
│   │   │   ├─ 搜索记忆: memory-index-query.sh           │   │
│   │   │   ├─ 编写代码: model/*, logic/*, handler/*     │   │
│   │   │   ├─ 编写测试: *_test.go (RED→GREEN→REFACTOR) │   │
│   │   │   └─ 更新: CHANGELOG.md                        │   │
│   │   ↓                                                 │   │
│   │   Phase: QA                                        │   │
│   │   ├─ Agent: QA (Verification-Before-Completion)   │   │
│   │   │   ├─ 脚本: harness-checks.sh --service xxx --json│ │
│   │   │   │   ├─ check_go_build                       │   │
│   │   │   │   ├─ check_go_vet                         │   │
│   │   │   │   ├─ check_go_test (含0/0检测)            │   │
│   │   │   │   ├─ check_proto_jstype                   │   │
│   │   │   │   ├─ check_json_string                    │   │
│   │   │   │   ├─ check_cross_service_import           │   │
│   │   │   │   ├─ check_error_codes                    │   │
│   │   │   │   ├─ check_hardcoded_secrets              │   │
│   │   │   │   ├─ check_graph_freshness                │   │
│   │   │   │   ├─ check_claude_structural_data         │   │
│   │   │   │   ├─ check_proto_ts_align                 │   │
│   │   │   │   ├─ check_api_stubs                      │   │
│   │   │   │   ├─ check_response_wrap                  │   │
│   │   │   │   ├─ check_bench_regression               │   │
│   │   │   │   └─ check_api_smoke                      │   │
│   │   │   ├─ 运行: go build ./...                     │   │
│   │   │   ├─ 运行: go vet ./...                       │   │
│   │   │   ├─ 运行: go test ./... -count=1             │   │
│   │   │   ├─ 检查: TDD证据 (RED→GREEN摘录)            │   │
│   │   │   └─ 写入: _qa.md                             │   │
│   │   ↓                                                 │   │
│   │   判断: QA PASS?                                   │   │
│   │   ├─ YES → Phase: Review                          │   │
│   │   └─ NO → Phase: Debug                            │   │
│   └─────────────────────────────────────────────────────┘   │
│      ↓ (假设QA FAIL)                                       │
│   ┌─────────────────────────────────────────────────────┐   │
│   │   Phase: Debug                                      │   │
│   │   ├─ Agent: Debug (Systematic Debugging)           │   │
│   │   │   ├─ 读取: _qa.md (failures详情)              │   │
│   │   │   ├─ 复现问题: 运行失败的命令                  │   │
│   │   │   ├─ 分析根因: git diff + 依赖链追溯          │   │
│   │   │   └─ 输出: {rootCause, evidence, fixSuggestions}│ │
│   │   └─ 回到 Iteration 2: Develop (修复模式)         │   │
│   └─────────────────────────────────────────────────────┘   │
│      ↓                                                        │
│   ┌─────────────────────────────────────────────────────┐   │
│   │ Iteration 2:                                         │   │
│   │   Phase: Develop (修复)                             │   │
│   │   ├─ Agent: Generator (debt模式)                   │   │
│   │   │   ├─ 读取: Debug输出的fixSuggestions          │   │
│   │   │   ├─ 修复问题                                  │   │
│   │   │   └─ 补回归测试                                │   │
│   │   ↓                                                 │   │
│   │   Phase: QA (FRESH run)                            │   │
│   │   └─ 重新执行15项检查                               │   │
│   │   ↓                                                 │   │
│   │   判断: QA PASS? → YES                             │   │
│   └─────────────────────────────────────────────────────┘   │
│      ↓                                                        │
│   ┌─────────────────────────────────────────────────────┐   │
│   │   Phase: Review (3视角并行)                         │   │
│   │   ├─ Agent: security-arch-reviewer                 │   │
│   │   │   ├─ 读取: CLAUDE.md, design.md, _qa.md       │   │
│   │   │   ├─ 审查: 架构一致性、安全性、变更完整性      │   │
│   │   │   └─ 写入: _review_security-arch.md           │   │
│   │   ├─ Agent: standards-eng-reviewer                 │   │
│   │   │   ├─ 读取: 项目编码规范.md, MEMORY.md          │   │
│   │   │   ├─ 审查: 规范遵循、复用性、测试覆盖、记忆遵守│   │
│   │   │   │   ├─ M1: grep "// SEE: \[\[" (收集引用)   │   │
│   │   │   │   ├─ M2: 验证引用准确性                   │   │
│   │   │   │   ├─ M3: memory-index-query.sh (检查遗漏) │   │
│   │   │   │   └─ M4: 建议新记忆 (memorySuggestions)   │   │
│   │   │   └─ 写入: _review_standards-eng.md           │   │
│   │   └─ Agent: design-biz-reviewer                    │   │
│   │       ├─ 读取: design.md, business-flows.md       │   │
│   │       ├─ 审查: 设计一致性、代码质量、Migration安全 │   │
│   │       └─ 写入: _review_design-biz.md              │   │
│   │   ↓                                                 │   │
│   │   判断: 2/3 PASS?                                  │   │
│   │   ├─ YES → 返回 SUCCESS {confidence: 0.85}        │   │
│   │   └─ NO → Iteration 3 (修复CRITICAL)              │   │
│   └─────────────────────────────────────────────────────┘   │
│                                                               │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ Workflow 2: web/pc (并行运行)                           │ │
│ │ ├─ 工具: Workflow({scriptPath: "harness-pipeline.js"}) │ │
│ │ ├─ 参数: {serviceName: "PC前端", serviceDir: "web/pc"}  │ │
│ │ └─ 内部循环 (类似Workflow 1)                            │ │
│ │     ├─ Generator: 实现Vue组件                          │ │
│ │     ├─ QA: harness-checks-frontend.sh (6项检查)       │ │
│ │     │   ├─ check_type_check (vue-tsc)                 │ │
│ │     │   ├─ check_unit_test (vitest)                   │ │
│ │     │   ├─ check_build (vite build)                   │ │
│ │     │   ├─ check_hardcoded_secrets                    │ │
│ │     │   ├─ check_debug_artifacts (console.log)        │ │
│ │     │   └─ check_type_safety (as any统计)             │ │
│ │     └─ Review: 3视角 (前端焦点)                        │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                               │
│ 等待两个Workflow完成...                                      │
│ ↓                                                             │
│ Owner收集结果:                                                │
│ ├─ Workflow 1: {status: "SUCCESS", confidence: 0.85}        │
│ └─ Workflow 2: {status: "SUCCESS", confidence: 0.82}        │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ 阶段 6: 集成归档 (Owner Agent 内联执行)                       │
│ ├─ 门禁检查: harness-gate-check.sh --phase 5 --change xxx    │
│ │   ├─ 检查: 每个服务的_qa.md存在?                           │
│ │   ├─ 检查: 每个服务的_review*.md存在?                      │
│ │   ├─ 检查: QA PASS?                                        │
│ │   ├─ 检查: Review PASS?                                    │
│ │   └─ 检查: CHANGELOG更新?                                  │
│ ├─ 全链路编译: cd $ROOT && go build ./...                    │
│ ├─ 全链路静态分析: go vet ./...                              │
│ ├─ 归档QA/Review:                                            │
│ │   mkdir -p .harness/changes/<change>/impl/community-hub/  │
│ │   mv services/community-hub-service/_qa.md → impl/        │
│ │   mv services/community-hub-service/_review*.md → impl/   │
│ │   mkdir -p .harness/changes/<change>/impl/web-pc/         │
│ │   mv web/pc/_qa.md → impl/                                │
│ │   mv web/pc/_review*.md → impl/                           │
│ ├─ 运行时冒烟: harness-smoke.sh (非阻塞)                     │
│ │   ├─ L1: 端口监听检查                                      │
│ │   ├─ L2: gRPC连通性                                        │
│ │   └─ L3: 依赖链检查                                        │
│ ├─ 处理Memory建议:                                           │
│ │   for suggestion in memorySuggestions:                    │
│ │     if not exists(.harness/knowledge/memory/{slug}.md):   │
│ │       create draft memory file                            │
│ │       update MEMORY.md index                              │
│ ├─ 生成summary.md: 基于impl/*/摘要                           │
│ ├─ 门禁检查: harness-gate-check.sh --phase 6 --change xxx    │
│ │   ├─ 检查: impl/目录存在?                                  │
│ │   ├─ 检查: summary.md完整?                                 │
│ │   └─ 检查: 包含关键章节?                                   │
│ └─ 更新索引: 追加到.harness/changes/INDEX.md                 │
└───────────────────────────────────────────────────────────────┘
    ↓
┌───────────────────────────────────────────────────────────────┐
│ HITL 确认 (置信度自适应审查)                                  │
│ ├─ 读取置信度: Workflow1=0.85, Workflow2=0.82                │
│ ├─ 判断审查深度:                                              │
│ │   ├─ ≥0.80 → 摘要审查 (读QA+Review summary)               │
│ │   ├─ 0.50-0.79 → 抽查 (随机抽2个文件全文阅读)             │
│ │   └─ <0.50 → 全文审查 (建议人工确认)                      │
│ └─ 用户批准 → 进入交付                                        │
└───────────────────────────────────────────────────────────────┘
    ↓
完成 ✅
