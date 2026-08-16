# Summary — notice-xss-sanitize-and-frontend-fixes

> 2026-08-16 归档。跨服务 L 级（community-hub-service + web/mobile）。

## 变更内容
1. **存储型 XSS 净化（community-hub-service，治本）**：bluemonday 白名单净化器 `internal/sanitize`，在 content_posts 三写路径（create/update/submit）净化正文。白名单 p/br/strong/em/u/s/ul/ol/li/a/h1-h6/blockquote/pre/code；img/iframe/script/object/embed 整体剔除；a 仅 href(http/https/mailto)+rel（nofollow+noopener noreferrer）、target 剔除。净化顺序：080005 非空校验原始正文先行、净化后为空接受空串。**Review 跟进**：submit 分支 Kafka 推送 payload 修正为净化后值（防事件重播未净化 HTML）+ 三条写路径净化命中日志。
2. **首页 watch 双重加载（web/mobile）**：`membershipsResolved` 首载守卫（onMounted `await loadMemberships()` 之后置位，load 内部覆写触发的 watch 被忽略，用户手动切换正常触发），单次权威加载。
3. **登录 toast 覆盖（web/mobile）**：profile 拉取失败改为单条合并提示「登录成功，但资料加载失败」（icon:none），不再被成功 toast 覆盖。

## 执行方式（与 spec-pipeline 的偏差）
- spec-pipeline 阶段 0-2（工具选择→需求澄清→需求分析→需求评审 3/4）**已完成**，产出 3 份 spec + 14 条决策 + 2 轮评审 16 条反馈全部解决 + Owner 2 条少数派裁决（守卫标志置位时点 / a 标签 target 剔除）。
- **阶段 2 续跑触发 spec-pipeline 自身 bug**（`approve.includes is not a function`：该 checkpoint 的决策要求传选项字符串而非索引，与阶段 1 不一致）→ **中止剩余仪式**，改为按已评审 spec 直派两条 harness-pipeline 编码（等价于阶段 5），QA/Review 各自闭环。
- 编码结果：community-hub（19 PASS/0 FAIL，131+ 测试，sanitize 覆盖率 100%）+ web/mobile（109 tests 全绿）均通过。

## 门禁
- community-hub：`go build/vet/test`（14 包）全绿；`harness-checks.sh --service community-hub-service` 19 PASS / 0 FAIL / 2 WARN（既有 gitlink/proto_ts_align 存量）。
- web/mobile：`npm run type-check/build:h5` 全绿；`npx vitest run` 109 tests；`harness-checks-frontend.sh --service mobile` 5 PASS / 0 FAIL / 2 WARN（既有 as-any/api_field_align 存量）。

## 待办/残余
- 存量已入库恶意 content **不回填**（用户决策 D5）：仅新写入净化，历史注入内容在重新编辑前仍可触发——已记录残余风险。
- web/mobile `@types/node ^24` 与 `typescript ^4.9` 的 typesVersions 兼容风险（review should-follow）：node 类型进入 app tsconfig 作用域，建议后续改 test 专用 tsconfig。
