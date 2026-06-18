# QA Report — web/pc (v3 — 最终版)

**验证时间**: 2026-06-15 15:39
**验证范围**: AI 模型配置页 API 密钥持久化修复 + 前端 QA 流水线修复
**执行方式**: Dev Agent → TDD 测试 → 独立 Review → 前端 QA 脚本

## 机械化检查结果（harness-checks-frontend.sh）

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | type-check (vue-tsc) | ✅ | pc: type check passed |
| 2 | unit-test (vitest) | ✅ | pc: **49 tests passed** (41 existing + 8 new) |
| 3 | build (vue-tsc -b + vite) | ❌ | 140 pre-existing TS errors — not from this change |
| 4 | hardcoded-secrets | ✅ | no secrets detected in source |
| 5 | debug-artifacts | ✅ | no debug artifacts in production code |
| 6 | type-safety | ⚠️ | 63 `as any` usages (pre-existing, aspirational target ≤10) |

## 新增测试详情 (tests/unit/views/aimodel/ModelForm.spec.ts)

| # | 测试用例 | 验证行为 |
|---|---------|---------|
| 1 | 创建云端模型时，createApiKey 被调用且参数正确 | 🔑 核心：Key 持久化 + 参数正确 |
| 2 | 本地模型不调用 createApiKey | 边界：本地模型跳过 |
| 3 | Key 创建失败不阻断模型创建，记录错误 | 容错：Key 失败≠模型失败 |
| 4 | 空白密钥（含纯空格）不触发 | 边界：`.trim()` |
| 5 | 批量创建每个模型独立 Key | 批量：N→N |
| 6 | 表单验证失败不提交 | 门禁：validate=false |
| 7 | 空响应不创建密钥不崩溃 | 防御：res=null |
| 8 | 编辑模式不触发 Key 创建 | 隔离 |

## 独立 Review 发现处理

| 级别 | 发现 | 处理 |
|------|------|:--:|
| C1 | `model_type` 不持久化（跨层 bug） | ⚠️ 预先存在，已记录 |
| W2 | 批量密钥名称区分度低 | ✅ 已修复 `{provider}-{modelName}-默认` |
| I1 | 缺验证失败测试 | ✅ 已补充 |
| I5 | 缺 null 响应测试 | ✅ 已补充 |

## 流水线修复

| 文件 | 变更 |
|------|------|
| `.harness/skills/qa/scripts/harness-checks-frontend.sh` | **新增** — 前端 6 项机械化检查脚本 |
| `.harness/skills/qa/SKILL.md` | 更新 — 增加前端服务路由和检查项说明 |

## 关键发现 (预先存在，不在本次范围)

| 问题 | 影响 |
|------|------|
| `model_type` 丢失 | 新建模型的 cloud/local 类型不生效 |
| 140 TS 错误 (`vue-tsc -b`) | 无法构建生产包 |
| 63 `as any` 逃逸 | 类型安全性差 |

---
VERDICT: PASS (核心检查全部通过，FAIL 和 WARN 均为预先存在)
---
