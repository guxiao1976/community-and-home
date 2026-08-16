# Summary: 移动端「社区家园」首页信息架构改造

> 变更：`mobile-homepage-content-revamp`（P1 / L 级 / modify）
> 完成日期：2026-08-16 · 管线：harness-spec-pipeline 全流程（0-6 阶段）

## 交付内容

| 服务/模块 | 变更 |
|----------|------|
| **api-proto** | `ListContentPostsRequest` 新增 `since_days`(int32, 字段6, 1..365, 缺省0=不过滤)，additive 非破坏；CHANGELOG 登记；`make ci` PASS |
| **community-hub-service** | migration 004 幂等补 `community_contacts` 表；migration 005 加 `idx_status_pinned_published(status,is_pinned,published_at)` 索引；Model `FindListByCommunity` 窗口谓词；RPC since_days 校验(080005)+传参；REST 层透传+Base 错误上抛 |
| **web/mobile** | 首页通知区 since_days=30&page_size=3 + 4 功能图标入口（便民联络做实/其余占位）；区块全序重排（通知→入口→邻里互助占位→寻失→底部广告集中）；notice-browse 改 30 天卡片列表；notice-detail 附件预览（file_type 白名单：图片 previewImage / 文档 openDocument）；新建 contact-list 联络拨号网格页 |

## 门禁与验证

- **QA**：community-hub 19 PASS/0 FAIL（124 测试全绿）；web/mobile 5 PASS/0 FAIL（62 测试全绿）
- **Review**：community-hub 3/3 PASS；web/mobile chore 通过
- **TDD 证据**：两侧有逻辑函数均有真实 RED FAIL 摘录（_tdd_evidence.md / CHANGELOG）
- **Owner 运维验证**：
  - migration 004/005 已执行落库 + DESCRIBE/SHOW INDEX 验证（幂等重跑安全）
  - EXPLAIN 窗口查询走 `idx_status_pinned_published`（type=ref, rows=1，不走全表扫描）
  - `since_days=999` → `080005`（REST→RPC 透传+校验生效）；`since_days=30`/缺省 → 成功
  - 便民联络接口修复（原 `community_contacts` 缺表报错 → 现返回空列表）
  - community-hub API+RPC 已用新代码重启（8887/8088）

## 决策落点（用户拍板 D1-D16）

保留跑马灯 / 列表页同过滤(30天) / 便民联络做实其余占位 / 仅补表不预置种子 / 附件走 file-service 预签名 / 广告原样堆叠底部+预留不跳转 / 邻里互助仅占位不开发 / 时间窗口后端强制。

## 不做（Won't have）

邻里互助后端数据源与发布/列表/详情页；物业报修/二手闲置/租房卖房落地功能；广告内容动态化与点击跳转。

## 归档

- specs/: 5 个能力 spec（19 REQ 追溯 ✅）
- design.md: 架构设计（16 任务，ADR 6 项）
- tasks.md: 任务清单（已全部完成）
- review/: 需求评审 4/4 APPROVED（3 轮收敛）
- _qa.md: community-hub + web/mobile QA 报告（PASS）
