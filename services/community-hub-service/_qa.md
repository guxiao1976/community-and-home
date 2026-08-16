# QA Report — community-hub-service

**验证时间**: 2026-08-16 11:10 (CST)
**验证范围**: 工作树未提交改动 + 未跟踪文件（content-post-generalization Task 1.1-1.23：42 modified + 21 untracked，含 rename notices→content_posts、scope 包、kafkapush、contentcompat、API 代理）

## 机械化检查结果 (harness-checks.sh — FRESH run)

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | exit 0 (compilation succeeded) |
| 2 | go vet | ✅ | exit 0 (no issues) |
| 3 | go test | ⚠️ | 13P/0F/10N, ~119 测试函数 — WARN：NEW package `rpc`（package main，薄接线）无测试 |
| 4 | Proto int64 jstype | ✅ | diff 无 proto 变更 (skipped) |
| 5 | json:",string" | ✅ | AST 校验全过，所有 int64 ID 字段均带 `json:",string"` |
| 6 | 跨服务DB导入 | ✅ | no violations（68 个 Go 文件 diff 扫描） |
| 7 | 错误码格式 | ✅ | no magic numbers（全部命名常量或 0） |
| 8 | 硬编码密钥 | ✅ | no secrets detected |
| 9 | go_fmt | ✅ | 变更 Go 文件全部已格式化 |
| 10 | proto_ts_align | ❌ | MISMATCH: proto `FileInfo.confirmed` 未同步 TS `FileInfo`（web/common/types/file.d.ts）；proto `FileInfo.file_type` 未同步 TS `FileInfo` |
| 11 | graph_freshness | ✅ | graph up-to-date (synced 0h ago) |
| 12 | claude_structural_data | ✅ | 无结构数据重复 |
| 13 | api_stubs | ✅ | 无 TODO 桩 |
| 14 | response_wrap | ✅ | 无双重包裹风险 |
| 15 | bench_regression | ✅ | 无 benchmark（SKIP） |
| 16 | api_smoke | ⚠️ | 无法确定 API 端口 — SKIP |
| 17 | memory_index / design_consistency / git_hygiene | ⚠️/✅/⚠️ | git_hygiene WARN：gitlink 无 .gitmodules 条目 (api-proto) |
| 18 | mutation_testing / pipeline_evals | ✅ | 变异分数 ?%（未解析到分数）；管线 eval 全过 |

**汇总**: 17 PASS / 1 FAIL / 3 WARN，`exit_code=1`

**FAIL 归因（proto_ts_align）**: 本工作树 api-proto 子模块指针 3d3a8ad→006d4ae 引入 `FileInfo.file_type(11)/confirmed(12)`，但前端 `web/common/types/file.d.ts` 未同步。违规不在 community-hub-service 自身 Go 代码（file.d.ts 未在本变更中改动）。CHANGELOG 已预声明此为 file-service 任务范围的跨服务/前端同步项。**机械门禁角度仍是 FAIL（硬约束 #4「FAIL 不可提交」需 Owner 裁决或 file-service 任务同步 TS）**，非本服务源码缺陷。

## 编译检查
- [x] go build ./... — **exit 0**
- [x] go vet ./... — **exit 0**

## 单元测试
- [x] go test ./... -count=1 — **exit 0**（13 包含测试，119 个测试函数，0 fail；10 包无测试文件 = handler/types/main/配置薄层）

## 测试覆盖（go test -cover，非阻塞记录）
| 包 | 覆盖率 | 状态 |
|----|--------|------|
| rpc/internal/logic/scope | 94.3% | 优 |
| api/internal/util | 86.4% | 优 |
| internal/contentcompat | 82.4% | 优 |
| api/internal/logic/contact | 79.2% | 优 |
| api/internal/logic/lostfound | 75.0% | 优 |
| rpc/internal/logic/notice | 73.1% | 良 |
| rpc/internal/kafkapush | 68.8% | 良 |
| rpc/internal/logic/lostfound | 68.7% | 良 |
| model | 57.1% | 中 |
| api/internal/logic/notice | 56.8% | 中 |
| rpc/internal/logic/contact | 47.1% | 中 |
| api/internal/svc | 25.0% | 低（仅接线） |
| handler/main/types/config | 0% | 生成代码/薄层，常规 |

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

**分诊说明**：字段映射类（model struct 改名、SQL 加列、Tx 变体、proto 透出、types.go wire 兼容、servicecontext 接线）只要求测试绿 — 全部通过（content_post_test 13 / content_post_scope_test 4 / content_post_attachment_test 3 个 sqlmock 测试锁定 SQL；api_proxy_test 覆盖 types 透出）。以下「有逻辑函数」逐项核对。

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认 | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| scope.ExpandDivisionCommunities | 有逻辑 | ✅ TestExpandDivisionCommunities×3 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| scope.ResolveAdminDivision | 有逻辑 | ✅ TestResolveAdminDivision | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| scope.AssertCommunitiesScope | 有逻辑 | ✅ TestAssertCommunitiesScope | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| scope.PublishRolesFrom | 有逻辑 | ✅ TestPublishRolesFrom×6 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| scope.PublishRoleToString | 有逻辑(映射) | ✅ TestPublishRoleToString×6 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| scope.IsLevel2Grant | 有逻辑(谓词) | ⚠️ 仅间接（PublishRolesFrom/GetPublishPermission 覆盖，无直接用例） | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| CreateContentPost | 有逻辑 | ✅ createcontentpostlogic_test×6 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| UpdateContentPost（V5 presence 分流） | 有逻辑 | ✅ updatecontentpostlogic_test×14 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| DeleteContentPost | 有逻辑 | ✅ read_write_logic_test×4 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| ListContentPosts | 有逻辑 | ✅ read_write_logic_test×2 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| GetContentPost | 有逻辑 | ✅ read_write_logic_test×6 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| GetMarqueeNotices | 有逻辑 | ✅ read_write_logic_test×2 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| GetPublishPermission | 有逻辑 | ✅ read_write_logic_test×3 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| kafkapush.Producer.Push | 有逻辑 | ✅ producer_test×4 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| kafkapush.Rescanner.ScanOnce | 有逻辑 | ✅ rescanner_test×4 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| contentcompat.ResolveReadableCommunityForCompat | 有逻辑 | ✅ contentcompat_test×1 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| API CreateContentPost 代理（community_ids 解析） | 有逻辑 | ✅ api_proxy_test×2 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| API UpdateContentPost 代理（presence 转发） | 有逻辑 | ✅ TestUpdateContentPost_PresenceForwarding | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| API GetContentPost 代理（compat 回退） | 有逻辑 | ✅ api_proxy_test×4 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| API GetMarqueeNotices 代理 | 有逻辑 | ✅ api_proxy_test×1 | ❌ 无 FAIL 摘录 | ✅ | FAIL |
| model 字段映射/查询/写路径（IsReviewComplete/FindListByCommunity/FindOneReviewComplete/FindMarquee/UpdateIsPinned/UpdateStatusAndPublish/UpdateAttachmentCount/Withdraw/UpdateKafkaPushStatus/FindPendingPush/InsertBatch×2/FindCommunityIdsByPostId/DeleteByPostId×2/FindByPostId） | 字段映射 | ✅ content_post_test×13 + scope×4 + attachment×3 | —（不要求） | ✅ | PASS |

- **GREEN 全部确认**：119 个测试函数全绿（exit 0）。测试行为精确（080005/080006/080002 映射、attachment_count 重算、is_pinned 操作者分流、Kafka 推送 ack/pending、compat 回退均被断言）。
- **RED 全部缺失**：对全部 ~20 个「有逻辑函数」，工作树（测试注释/CHANGELOG/loop-runs/task 文件/`_tdd_evidence.md`）**均无任何具体 FAIL 输出摘录**（无 `undefined:` / `expected:... actual:...` / `Error:` 文本）。
- **结构性佐证（成立但不足）**：`git show HEAD:` 确认 division.go / producer.go / contentcompat.go / helper.go 新函数在 HEAD 均不存在（RED 真实），但按既定标准不替代真实摘录。

## 发现

| 级别 | 问题 | 建议 |
|------|------|------|
| ❌ FAIL（TDD 证据） | 全部「有逻辑函数」缺 RED 具体 FAIL 摘录；CHANGELOG 声称"RED 摘录留档于测试注释"与事实不符。系 `tdd-red-evidence-requires-fail-excerpt` 记忆所述失败类的**第 4 次复发**（此前 master-data T2.2 / restore / web-mobile T5.1 均判 QA FAIL）。机械门禁只查包级测试文件与 RED 列文本格式，未按函数维度门禁，故 17 PASS 无法拦截 | Generator 用 `git stash` 回退生产文件复现真实 RED 输出（含 `Error Trace`/`expected...actual...`/`undefined:` 行号），持久化到 CHANGELOG 或 `_tdd_evidence.md`；修复目标 #4（TDD 证据强制捕获）需闭环 |
| ⚠️ WARN（机械 FAIL 已声明） | proto_ts_align FAIL：FileInfo.confirmed/file_type 未同步 web/common/types/file.d.ts（本变更子模块指针引入，CHANGELOG 预声明属 file-service 范围） | file-service 任务同步 TS；硬约束 #4 需 Owner 确认可提交 |
| ⚠️ WARN | rpc 主包（package main）无测试 | 薄接线，可接受，已记录 |
| ⚠️ WARN | git_hygiene：api-proto gitlink 无 .modules 条目 | 按 Git治理规范补 .gitmodules |
| ⚠️ WARN | api_smoke 因无法确定端口 SKIP | 配置 smoke 测试端口 |

---
VERDICT: FAIL → PASS（2026-08-16 闭环）
---

**初判依据（11:10）**：
1. **TDD RED 证据不足（决定性）**：~20 个「有逻辑函数」RED 列全部 ❌（无具体 FAIL 摘录），按 `tdd-red-evidence-requires-fail-excerpt`（must-follow）及管线规则「只有有逻辑函数的 RED 缺失才判 QA FAIL」，判定 QA FAIL。
2. **机械化检查 1 项 FAIL**（proto_ts_align，exit_code=1）：虽为预声明的跨服务/前端同步项、非本服务源码缺陷，但机械门禁角度为 FAIL，需 Owner 裁决或 file-service 任务闭环。
3. 编译/静态/单测全部绿：go build exit 0 / go vet exit 0 / go test exit 0（13 包 119 测试函数），测试覆盖合理（scope 94%、contentcompat 82%、notice logic 73%）。**QA 对代码正确性无异议，FAIL 点纯粹是 TDD 过程证据 + 机械门禁。**

**闭环处理（11:53-12:00）**：
- **FAIL-1（TDD 证据）已补救**：新建 `services/community-hub-service/_tdd_evidence.md` —— 在父提交 `dca1225` 临时 worktree 叠加 13 个新测试文件 + HEAD go.mod/go.sum，`go test` 复现真实编译失败（`undefined: NewContentPostScopeModel` / `undefined: ResolveReadableCommunityForCompat` / `undefined: ContentReviewMessage` / `undefined: ExpandDivisionCommunities` / `undefined: model.ContentPostModel` / `undefined: communityv1.NoticeServiceClient` 等，均含行号）。覆盖 model / contentcompat / kafkapush / scope / notice logic / API 代理六组有逻辑函数。GREEN 复跑 13P/119F exit 0。
- **FAIL-2（proto_ts_align 机械门禁）已消除**：`web/common/types/file.d.ts` 已同步 `FileInfo.confirmed`/`file_type`；重跑 harness-checks → **19 PASS / 0 FAIL / 2 WARN（既有的 identity.ts 等跨服务 TS 滞后 + git_hygiene），exit 0**。

**终判复现命令（every-fresh-run 证据）**：
```
bash .harness/skills/qa/scripts/harness-checks.sh --service community-hub-service --json   # 19P/0F/2W, exit 0
cd services/community-hub-service && go build ./... && echo $?   # 0
cd services/community-hub-service && go vet ./... && echo $?     # 0
cd services/community-hub-service && go test ./... -count=1 && echo $?   # 0, 13P/119F
git worktree add --detach <tmp>/red-repro dca1225 && # 叠加测试文件后 go test → RED 各包 build failed（见 _tdd_evidence.md）
```

**保留 WARN（不阻塞）**：identity.ts / moderation.d.ts 既有 TS 滞后（跨服务欠账，非本变更引入）；api-proto gitlink 无 .gitmodules 条目（Git 治理欠账）。
