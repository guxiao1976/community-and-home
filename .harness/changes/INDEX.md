# 变更索引

> 最后更新：2026-06-21
>
> 变更模板见 [TEMPLATE.md](TEMPLATE.md) — 每个新需求复制此模板到 `.harness/changes/<name>/summary.md`。
>
> 下表每条记录是一个"迷你 summary"，链接到分布在各处的产物。

## 基础设施变更（Harness 自身建设）

### 🟢 Harness 体系搭建

| 属性 | 值 |
|------|-----|
| 日期 | 2026-06-09 ~ 2026-06-10 |
| 状态 | ✅ 已完成 |
| 摘要 | 四支柱 Harness 体系从零搭建 |

**产物索引**:
| 类型 | 路径 |
|------|------|
| Owner Agent | `.harness/agents/owner-agent.md` (~165 行，五模块对齐) |
| 规则体系 | `.harness/rules/` (工程结构, Proto管理, 项目编码) |
| 技能体系 | `.harness/skills/` (9 个 Skill + qa/ 目录含脚本) |
| 经验记忆 | `.harness/knowledge/memory/` (8 条，统一 frontmatter + type 分类) |
| 业务知识 | `.harness/knowledge/business-flows.md` (7 条端到端流程) |
| 知识索引 | `.harness/knowledge/INDEX.md` (架构/业务/数据/图谱) |
| 变更追溯 | `.harness/changes/INDEX.md` + `TEMPLATE.md` |
| 编排流水线 | `.harness/workflows/harness-pipeline.js` |
| 服务映射 | 6 服务 CLAUDE.md 统一引用 `.harness/` |
| 决策记录 | `.harness/knowledge/memory/harness-architecture-decisions.md` |
| Dry Run | [dry-run-2026-06-09.md](dry-run-2026-06-09.md) |

---

## 业务需求变更

### ✅ RBAC 管理界面

| 属性 | 值 |
|------|-----|
| 日期 | 2026-06-19 |
| 状态 | ✅ 已完成 |
| 范围 | permission-service, web/pc |

**产物索引**:
| 类型 | 路径 |
|------|------|
| Request | `.harness/changes/rbac-management-ui/request.md` |
| Proposal | `.harness/changes/rbac-management-ui/proposal.md` |
| Design | `.harness/changes/rbac-management-ui/design.md` |
| Specs | `.harness/changes/rbac-management-ui/specs/` (5个规格说明) |
| Tasks | `.harness/changes/rbac-management-ui/tasks.md` |
| Execution Summary | `.harness/changes/rbac-management-ui/EXECUTION_SUMMARY.md` |
| Implementation Summary | `.harness/changes/rbac-management-ui/IMPLEMENTATION_SUMMARY.md` |
| Delivery Report | `.harness/changes/rbac-management-ui/DELIVERY_REPORT.md` |
| CHANGELOG | `services/permission-service/CHANGELOG.md` |

**核心交付**:
- 后端：权限数据初始化脚本（4角色+50权限项）
- 后端：单元测试覆盖（Model层19个测试、Logic层5个测试）
- 后端：int64字段JSON序列化修复（7处修复点）
- 前端：用户角色分配页面（支持4种数据权限作用域）
- 前端：路由权限守卫（动态路由生成+权限过滤）
- 前端：角色管理、权限配置、v-permission指令

**关键特性**:
- 接口级权限控制（防止前端绕过）
- 数据权限管理（社区/楼栋/单元/房间四级作用域）
- 菜单权限控制（动态路由）
- 按钮/内容级权限控制（v-permission指令）

### 🟡 系统配置 Redis 化

| 属性 | 值 |
|------|-----|
| 日期 | 2026-06-09 |
| 状态 | 🟡 进行中 |
| 范围 | 全局（sysconfig fallback 机制） |

**产物索引**:
| 类型 | 路径 |
|------|------|
| Plan | `docs/superpowers/plans/2026-06-09-system-config-redis.md` |
| Spec | `docs/superpowers/specs/2026-06-09-system-config-redis-design.md` |
| CHANGELOG | *(进行中)* |
| QA | — |
| Review | — |

### ✅ 我的页面 & 加入小区流程

| 属性 | 值 |
|------|-----|
| 日期 | 2026-06-07 |
| 状态 | ✅ 已完成 |
| 范围 | user-service, api-proto, web/mobile |

**产物索引**:
| 类型 | 路径 |
|------|------|
| Plan | `docs/superpowers/plans/2026-06-07-my-page-and-join-flow.md` |
| Spec | `docs/superpowers/specs/2026-06-07-my-page-and-join-flow-design.md` |
| Proto CHANGELOG | `api-proto/CHANGELOG.md`（JoinCommunityRequest 增加地址字段） |
| QA | — |
| Review | — |

### ✅ 公告页重设计

| 属性 | 值 |
|------|-----|
| 日期 | 2026-06-06 |
| 状态 | ✅ 已完成 |
| 范围 | web/mobile |

**产物索引**:
| 类型 | 路径 |
|------|------|
| Plan | `docs/superpowers/plans/2026-06-06-notice-page-redesign.md` |
| Spec | `docs/superpowers/specs/2026-06-06-notice-page-redesign.md` |
| 前端 CHANGELOG | `web/mobile/CHANGELOG.md` |

### ✅ 移动端社区功能 + Uni-app 框架

| 属性 | 值 |
|------|-----|
| 日期 | 2026-06-06 |
| 状态 | ✅ 已完成 |
| 范围 | web/mobile |

**产物索引**:
| 类型 | 路径 |
|------|------|
| Plan | `docs/superpowers/plans/2026-06-06-mobile-community-feature.md` |
| Plan | `docs/superpowers/plans/2026-06-06-uni-app-mobile-framework.md` |
| 前端 CHANGELOG | `web/mobile/CHANGELOG.md`（SMS 登录 + RSA 加密） |

### ✅ AI 开发团队实施

| 属性 | 值 |
|------|-----|
| 日期 | 2026-06-05 |
| 状态 | ✅ 已完成 |
| 范围 | 全局基础设施 |

**产物索引**:
| 类型 | 路径 |
|------|------|
| Plan | `docs/superpowers/plans/2026-06-05-ai-dev-team-implementation.md` |
| 设计文档 | `docs/specs/ai-dev-team-design.md` |
| 多服务 CHANGELOG | W8 conf.MustLoad 迁移（全服务） |

---

## 快速修复（无完整 OpenSpec）

| 日期 | 变更 | 服务 | 备注 |
|------|------|------|------|
| 2026-06-21 | ai-model-service configx 统一 | ai-model-service | 改用 common/pkg/configx，消除本地副本 |
| 2026-06-19 | permission-service 单元测试补充 | permission-service | Model层19测试+Logic层5测试，TDD覆盖 |
| 2026-06-19 | permission-service int64序列化修复 | permission-service | 7处字段添加 json:",string" 标注 |
| 2026-06-18 | master-data DivisionCountItem 修正 | master-data-service | Id字段加,string标签，Level改int32 |
| 2026-06-16 | moderation-service 管线配置测试修复 | moderation-service | 修复 CRITICAL fallthrough + 7个WARNING |
| 2026-06-15 | master-data GetDivision 桩实现补全 | master-data-service | QA修复 |
| 2026-06-06 | community-hub-service 初始化 | community-hub-service | 新服务（通知/联络/寻失） |
| 2026-06-04 | C1/C2 架构债务修复 | moderation-service | HTTP→gRPC, 直读DB→gRPC |
| 2026-06-04 | C8 SMS 验证码绕过修复 | auth-service | 安全修复 |
| 2026-06-04 | W8 conf.MustLoad 迁移 | 多服务 | 全局公约对齐 |
| 2026-06-04 | common v2.1.0 错误码 6→5位 | common, user-service | Breaking change |

## 状态图例

| 标记 | 含义 |
|:---:|------|
| 🟡 进行中 | 有 plan/spec，尚未全部完成 |
| ✅ 已完成 | 已交付，CHANGELOG 已记录 |
| ⚠️ 缺 Review | 已完成但无评审记录 |
| ❌ 已废弃 | 方案取消或变更 |
