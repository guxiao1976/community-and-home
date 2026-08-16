# 变更追溯索引

> 记录所有通过 OpenSpec 流程完成的需求/功能开发，按时间倒序排列

## 格式说明

```
## YYYY-MM-DD — <变更名>

**路径**: [直接Edit / Dev Agent / OpenSpec]
**状态**: [进行中 / 已完成 / 已归档]
**涉及服务**: service-a, service-b
**关联**: [PR #123] [Issue #456]

[一句话描述]

详见: [.harness/changes/<change-name>/](./<change-name>/)
```

## 2026-08-17 — 公告 XSS 净化 + 移动端完善（notice-xss-sanitize-and-frontend-fixes）

**路径**: L 级（spec-pipeline 阶段 0-2 需求分析+评审 → 因管线续跑 bug 改直派 2×harness-pipeline 编码）
**状态**: ✅ 已完成（community-hub 净化 19 PASS/0 FAIL、user-service 18 PASS/0 FAIL、web/mobile 6 PASS/0 FAIL）
**涉及服务**: community-hub-service, user-service, web/mobile

存储型 XSS 治本（bluemonday 白名单，content_posts 三写路径净化）+ 首页 watch 双加载守卫 + 登录 toast 合并 + profile 本人返回明文手机号 + 我的页重构（图标网格/新增身份/退出登录）+ rem 单位体系改造（rpx/px→rem，根字号16px 固定）+ 前端单位规范 §13 + unit_standard 检查规则 + 若干 Review 安全跟进（Kafka payload/IDOR/logout 清缓存/孤儿页清理）。

详见: [.harness/changes/notice-xss-sanitize-and-frontend-fixes/](./notice-xss-sanitize-and-frontend-fixes/)

---

## 2026-08-16 — 通用图文发布组件重构
## 2026-08-16 — 移动端首页信息架构改造（mobile-homepage-content-revamp）

**路径**: L 级 spec-driven（需求评审 4/4 + 架构设计 16 任务 + 编码双管线 + Owner 运维验证）
**状态**: ✅ 已完成（migration 落库 + EXPLAIN 索引 + since_days 透传 e2e + 全流程提交 8a09b86）
**涉及服务**: api-proto, community-hub-service, web/mobile

首页信息架构重排：通知 30 天窗口（后端 since_days 强制）+ 4 功能入口（便民联络做实/其余占位）+ 邻里互助占位 + 广告集中底部 + 通知列表/详情/附件预览改造 + 新建联络列表页。修复 community_contacts 缺表（migration 004）+ 新增窗口索引（005）。

详见: [.harness/changes/mobile-homepage-content-revamp/](./mobile-homepage-content-revamp/)

---

（content-post-generalization）

**路径**: L 级 spec-driven（设计评审 4 轮 + 实现 4 服务 + 运维验证）
**状态**: ✅ 已完成（Kafka 基建 + 三库迁移 + 端到端冒烟；运维验证修复 published_at 落库 / Kafka push 非阻塞 2 bug）
**涉及服务**: api-proto, community-hub-service, file-service, permission-service, moderation-service, docker-compose

把通知模块升级为通用图文发布组件（content_posts 通用化 + 内容级审核 + Kafka）。实现 + QA 由上一会话完成（4 服务提交 bddeadc/d684e7b/87f7d02/499330c），本会话续跑收尾：Kafka 基建（Task 0.4）、community-hub QA 闭环（TDD RED 证据补录）、Owner 运维验证（Task 6.1 三库迁移 + 6.2 端到端冒烟）。运维验证真实暴露并修复 2 个 bug（Insert 漏 published_at 列→跑马灯恒空；Kafka push 阻塞发布→DeadlineExceeded）。

详见: [.harness/changes/content-post-generalization/](./content-post-generalization/)

---

## 2026-08-16 — 通知多小区发布（notice-multicommunity-publish）

**路径**: L 级 spec-driven
**状态**: 📦 已归档（被 content-post-generalization 通用化方案取代）
**涉及服务**: community-hub-service, api-proto

原通知模块多小区发布 + 附件安全 + 权限设计（D1-D32）。B 方案（用户拍板 2026-08-16）将其升级为通用图文发布组件 content-post-generalization，本变更设计作为可复用参考（多小区发布/附件安全/权限/GetPublishPermission/GetMarqueeNotices 等）并入新变更，独立归档。

详见: [.harness/changes/notice-multicommunity-publish/](./notice-multicommunity-publish/)

---

## 2026-08-15 — rel_user_role 迁移 + 寻失路径修复（rel-user-role-migration-publish-fix）

**路径**: L 级 spec-pipeline（阶段 0-6 全流程）
**状态**: ✅ 已完成（需求评审 2 轮收敛 4/4，设计评审 2 视角，编码 3 服务 PASS）
**涉及服务**: permission-service, web/mobile

修复 rel_user_role 三列缺迁移（P0，从零建库 1054）+ 移动端寻失列表路径 404 + snake/camel 字段不匹配（P1）。需求评审第 1 轮 validity/clarity 抓到「REQ-P0-2 旧结构定义自相矛盾」，反馈注入后第 2 轮 4/4 收敛；执行评审 design-biz 抓到「前端 camelCase vs 后端 snake_case」CRITICAL，字段对齐后修复。本变更同时作为「检验完善后流水线」的真实 L 级样本，暴露 3 个流程 gap（评审报告落盘沙箱失效 / 执行评审无 CRITICAL 一票否决 / MySQL 迁移验证无法走 Go 管线）。

详见: [.harness/changes/rel-user-role-migration-publish-fix/](./rel-user-role-migration-publish-fix/)

---

## 2026-08-14 — 角色管理 bug 修复 + platforms 写链路（role-platforms-save）

**路径**: L 级 spec-pipeline（阶段 0-6 全流程）
**状态**: ✅ 已归档
**涉及服务**: api-proto, permission-service, web/pc

修复角色管理 500 panic（系统角色编辑）+ platforms（允许登录端）写链路补全 + 列表列宽重排。系统角色字段级编辑策略、60008 值域校验、base-check 类级审计、sort_order 落库修复。QA 16 PASS/0 FAIL，Review 3/3 PASS，已重启生效。

详见: [.harness/changes/role-platforms-save/](./role-platforms-save/)

---

## 2026-08-14 — spec-pipeline L 级 e2e 验证（spec-pipeline-e2e-l）

**路径**: spec-pipeline 全流程 e2e 测试
**状态**: ✅ 测试完成（流水线验证样本）
**涉及服务**: permission-service

L 级 spec-pipeline 全流程自动化 e2e 测试样本（role-list-sort 场景），检验阶段 0-6 衔接。

详见: [.harness/changes/spec-pipeline-e2e-l/](./spec-pipeline-e2e-l/)

---

## 2026-08-13 — web/pc 前端历史债务清理（web-pc-debt-cleanup）

**路径**: Dev Agent
**状态**: ✅ 已完成
**涉及服务**: web/pc

web/pc 122 条历史 TypeScript 错误清理（erasable syntax、类型对齐）。

详见: [.harness/changes/web-pc-debt-cleanup/](./web-pc-debt-cleanup/)

---

## 2026-08-13 — 管线重构规划（pipeline-refactor）

**路径**: 规划
**状态**: 📋 规划（部分落地）
**涉及服务**: global

消除硬编码服务名映射、prompt 模块化（templates + new 体系）的规划。服务名映射外提 registry 已于 2026-08-14 落地。

详见: [.harness/changes/pipeline-refactor/](./pipeline-refactor/)

---

## 2026-08-12 — 管线执行层修复（harness-pipeline-fix）

**路径**: Dev Agent
**状态**: ✅ 已完成
**涉及服务**: global

RED 证据分诊 + 变异测试（gomu）stub 接入管线，解决 QA 假 PASS 与测试有效性。

详见: [.harness/changes/harness-pipeline-fix/](./harness-pipeline-fix/)

---

## 2026-08-11 — 工作记录模块 v3（work-records-v3）

**路径**: Pipeline 测试
**状态**: ✅ 测试完成（流水线验证样本）
**涉及服务**: user-service

工作记录模块 v3，流水线完整修复后的验证样本。

详见: [.harness/changes/work-records-v3/](./work-records-v3/)

---

## 2026-08-11 — 工作记录模块 v2（work-records-v2）

**路径**: Pipeline 测试
**状态**: ✅ 测试完成（流水线验证样本）
**涉及服务**: user-service

工作记录模块 v2，流水线修复后开发验证样本。

详见: [.harness/changes/work-records-v2/](./work-records-v2/)

---

## 2026-08-11 — 工作记录 pipeline 测试（test-pipeline-work-records）

**路径**: Pipeline 测试
**状态**: ✅ 测试完成（流水线验证样本）
**涉及服务**: user-service

工作记录模块 pipeline 测试样本（含 p0-fix 报告、pipeline 评估）。

详见: [.harness/changes/test-pipeline-work-records/](./test-pipeline-work-records/)

---

## 2026-08-11 — clean-slate 测试（test-clean-slate）

**路径**: Pipeline 测试
**状态**: ✅ 测试完成（流水线验证样本）
**涉及服务**: global

clean-slate 实现 + 债务清理的流水线测试样本。

详见: [.harness/changes/test-clean-slate/](./test-clean-slate/)

---

## 2026-06-20 — RBAC 管理界面（rbac-management-ui）

**路径**: OpenSpec
**状态**: ✅ 已完成
**涉及服务**: web/pc, permission-service

RBAC 管理界面：角色/权限/用户-角色分配管理（含 specs 拆分与各阶段评审）。

详见: [.harness/changes/rbac-management-ui/](./rbac-management-ui/)

---

## 2026-06-17 — 内容审核全链路集成（moderation-integration）

**路径**: Dev Agent
**状态**: ✅ 已完成
**涉及服务**: moderation-service

内容审核全链路集成（请求/提案/回顾）。

详见: [.harness/changes/moderation-integration/](./moderation-integration/)

---

## 2026-08-13 — 访问控制与数据权限（access-control）

**路径**: OpenSpec → N×Workflow（7 服务）
**状态**: ✅ 已完成（7/7）
**涉及服务**: permission-service, auth-service, user-service, community-hub-service, master-data-service, web/pc, web/mobile
**关联**: api-proto c86dc84

端限制登录准入（sys_role.platforms + 50007）+ 当前小区应用状态（user_app_state + 10015）+ 板块发布配额（sys_section_quota + 80007）+ 成员约束（每户≤6 + 10014）+ 同屋互见（SameHouseInfo）。需求 1（角色分层）/3（数据权限统一模型）由前置 access-data-permission 交付，本变更只做需求 2/4/5/6。同时作为「检验开发流水线」的真实 L 级样本，暴露并沉淀 6+ 改进点：RED 证据分诊 + 变异测试 stub（harness-pipeline-fix）、web/pc 122 条历史 TS 错误清理（web-pc-debt-cleanup）。

详见: [.harness/changes/access-control/](./access-control/)

---

## 2026-08-12 — 数据权限核心（access-data-permission）

**路径**: OpenSpec → 3×Workflow + Owner 集成验收
**状态**: ✅ 全部完成（阶段①-⑥）
**涉及服务**: permission-service, master-data-service, user-service, community-hub-service, web/mobile
**关联**: proto commit 031f4e4 + c245c09

数据权限统一模型全量落地：scope 三态 + 能力分层(min_verf_level) + registered_user 基角色 + 祖先链解析 + 发布校验。Wave1=阶段①②(permission T1.1-T1.8 / master-data T2.1-T2.3)，Wave2=阶段③④⑤(user T3.1-T3.4 / community-hub T4.0-T4.8 / web-mobile T5.1)+ 阶段⑥ 集成验收(T6.1 九项矩阵全绿)。验收修复 2 个真实缺陷：permission 种子缺口(owner/tenant 发布+读权限/选举绑定) + master-data scope 缓存 O(N²) 加载(keyset 分页+预热,656k 行 27min→4s)。

详见: [.harness/changes/access-data-permission/](./access-data-permission/)

---

## 2026-06-18 — 审核服务管线配置化

**路径**: OpenSpec
**状态**: 已归档
**涉及服务**: moderation-service
**关联**: commit 84eadb2

实现内容审核的管线配置化功能，支持动态配置审核策略。

详见: [.harness/changes/moderation-pipeline-config/](./moderation-pipeline-config/)

---

## 2026-06-16 — AI 模型服务增强

**路径**: Dev Agent
**状态**: 已归档
**涉及服务**: ai-model-service
**关联**: 无

为 AI 模型服务添加连接测试和模板管理功能。

> ⚠️ 变更目录已清理（产物见 ai-model-service/CHANGELOG.md）
