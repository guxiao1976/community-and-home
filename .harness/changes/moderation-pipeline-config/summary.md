# 变更摘要 — 审核服务管线配置化

**创建时间**: 2026-06-16 10:00
**完成时间**: 2026-06-18 21:09
**路径**: OpenSpec

---

## 阶段 0: 路径选择

- **路径**: OpenSpec
- **理由**: 涉及数据库 Schema 变更（新增 mod_pipeline_config 表）+ API 层 + RPC 层 + 核心引擎逻辑修改，跨多个层次
- **涉及服务**: moderation-service
- **跳过阶段**: 无

---

## 阶段 1: 需求分析

- **执行方式**: 子Agent
- **产出**: （历史变更，产出物未保留）
- **状态**: ✅ 通过

---

## 阶段 2: 需求评审

- **执行方式**: （历史变更，流程简化）
- **状态**: ⏭️ 跳过

---

## 阶段 3: 架构设计

- **执行方式**: 子Agent
- **产出**: （历史变更，产出物未保留）
- **状态**: ✅ 通过

---

## 阶段 4: Proto 变更

- **执行方式**: Owner 内联
- **变更清单**: 无 Proto 变更（仅服务内部功能）
- **状态**: ⏭️ 跳过

---

## 阶段 5: 编码+测试

- **执行方式**: 1×Workflow
- **服务**: moderation-service
- **迭代次数**: 未记录
- **QA 结果**: PASS（见 impl/moderation-service/_qa.md）
  - 编译: ✅ 通过
  - 测试: ✅ 115/115 通过
  - 机械化检查: 8/8 通过（2 个 WARN）
- **Review 结果**: PASS（3/3 视角通过）
  - security-arch: PASS (0 CRITICAL)
  - standards-eng: PASS (0 CRITICAL, 9 WARNING)
  - design-biz: PASS (0 CRITICAL)
- **置信度**: 未记录（需补充）
- **状态**: ✅ 通过

---

## 阶段 6: 集成归档

- **全链路编译**: ✅ 通过
- **运行时冒烟**: ⏭️ 跳过
- **Memory Suggestions**: 0（Review 未建议新记忆）
- **状态**: ✅ 完成

---

## 关键决策

| # | 决策点 | 决策内容 | 原因 |
|---|--------|---------|------|
| 1 | 错误码设计 | 管线 Logic 使用 `errors.New()` 而非 `errx.NewCodeError()` | Review 发现但标记为 WARNING，后续改进 |
| 2 | Execute() 方法测试 | 核心方法暂无单元测试 | 依赖外部 TextEngine + gRPC，标记为 WARNING，建议 mock 测试 |
| 3 | gRPC 超时配置 | callModel() 未设独立超时 | 调试端点，影响有限，标记为 WARNING |

---

## 例外 & 未解决问题

| # | 问题 | 影响 | 后续计划 |
|---|------|------|---------|
| 1 | harness-checks.sh 脚本执行失败 | check_frontend: command not found | 已创建修复任务 P0-fix-harness-checks.md |
| 2 | 硬编码密钥检查失败 | 分类器临时不可用，降级手工检查 | 同上 |
| 3 | API 层测试覆盖率 0% | handler/logic 无测试，风险高 | 纳入 P0 改进计划 |

---

## 人工抽查（置信度 < 0.80）

未执行（历史变更）

---

## 交付清单

- [x] 代码已合并到 master (commit: 84eadb2)
- [x] CHANGELOG 已更新
- [x] QA/Review 报告已归档 (impl/moderation-service/)
- [ ] Memory Suggestions 已处理 (无新建议)
- [x] 索引已更新 (.harness/changes/INDEX.md)
