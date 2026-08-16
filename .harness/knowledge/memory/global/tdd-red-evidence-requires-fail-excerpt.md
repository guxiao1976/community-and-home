---
triggers: ["TDD", "RED", "GREEN", "测试未提交", "测试证据", "undefined", "编译失败", "QA FAIL", "证据摘录"]
type: process
severity: must-follow
service: all
status: active
created: 2026-08-12
updated: 2026-08-16
apply_count: 3
---
# QA 的 TDD RED 证据必须包含实际 FAIL 输出摘录

## 为什么会有这条经验

1. 管线 Generator 直接改主树、不提交（harness-pipeline 设计如此），RED 阶段的编译/测试失败只被口头描述进 CHANGELOG（"RED（编译失败：`Invalidate(ctx)` 签名变更 + `SetVersionReaderForTest` 未定义）"），**没有持久化任何实际错误输出**。
2. QA（master-data-service T2.2 整树缓存跨进程失效）时，要求 TDD 证据表 RED 列给出实际 FAIL 摘录（如 `undefined: SetVersionReaderForTest` / `too many arguments`），但全仓 grep 不到任何编译错误文本（reflog/stash/loop-runs/request 均无）。
3. 结果是：GREEN 全绿 + 机械门禁 16/16 PASS，但 RED 证据列 ❌ → 按规则判定 QA FAIL（TDD 证据不足）。
4. `.harness/changes/harness-pipeline-fix/request.md` 第 4 行已记录此问题（"TDD 证据缺失(RED→GREEN 无摘录,测试未 commit) → QA TDD FAIL"），修复目标 #4 为"TDD 证据强制捕获"——但 T2.2 仍未根治，证明该修复未闭环。

## 怎么做

- **分诊（2026-08-13 定稿，详见 `.harness/changes/harness-pipeline-fix/design-tdd-evidence.md`）**：RED 证据只对「有逻辑函数」（分支/转换/计算/条件/校验）要求；「字段映射类」（struct 加字段、SQL 加列、proto 字段透出、seed、纯接线）无独立逻辑，**只需测试绿，不要求 RED 摘录**。QA/generator 均已按此分诊（qa.js / generator.js）。

- **Generator 侧**：RED 阶段必须把真实失败输出（`go build`/`go test` 的错误文本，含 `undefined: FuncName`、行号）写入任务记录 / CHANGELOG，不能只写"看到失败"。
- **测试与生产代码同 commit**：新增测试文件（如 `invalidate_test.go`）不能以未跟踪状态留在工作树，须与生产代码一起提交（request.md 修复目标 #2）。
- **QA 侧判定**：RED 列没有具体 FAIL 输出摘录（仅文字描述）→ ❌；只要存在 ❌ → QA FAIL（TDD 证据不足）。结构性证明（`git show HEAD:` 显示新符号不存在）可作辅助佐证，但不能替代真实摘录。
- **复现检查法**（只读，不改树）：`git show HEAD:<file>` 确认测试引用的新函数/新签名在 HEAD 中不存在，即可证明 RED 是真实编译期失败。

## 触发场景

- 任何"新函数/签名变更 + 测试"的管线任务，QA 检查 TDD 证据时。
- Generator 交付时 CHANGELOG 写"RED（编译失败）"但无实际错误文本。

## 复现场景（2026-08-12，master-data-service 多视角评审修复轮）

**现象**：本轮新增 3 个测试文件，其中 `restoreDeletedItemLogic_test.go`（删除恢复失效接线，5 用例）在 CHANGELOG 只写 "RED → GREEN"，**无实际 FAIL 输出摘录**；`routes_test.go` / `reviewItemLogic_test.go` 已按要求补摘录（`expected: 401, actual: 200/400/EOF` / `expected: "1", actual: ""`）。`_tdd_evidence.md` 仍只覆盖 T2.2 编译错误类 RED，未扩展到本轮行为型断言失败。

**根因**：
1. 行为型断言 RED（testify `require.True` / `assert.Equal` 失败）是**运行时断言失败文本**，不是编译期错误，Generator 未捕获持久化（编译失败有编译器输出天然可抄，断言失败需主动跑一次修复前测试并复制输出）。
2. `_tdd_evidence.md` 被当成"一次性产物"只覆盖首批编译失败，新增测试文件后未同步扩展。

**QA 判定**：restore 行 RED 无摘录 → ❌ → QA FAIL（与 T2.2 同一失败类，证明未根治）。

**补救**：行为型断言 RED 也必须在 RED 阶段实际跑一次测试并持久化失败文本（含 `Error Trace` / `Error:` / `Test:` 行）；`_tdd_evidence.md` 须覆盖**全部**新增测试文件，不能只覆盖编译错误类。

## 复现场景（2026-08-12，web/mobile 阶段⑤ T5.1 加入小区携带 ownership）

**现象**：`web/mobile/src/pages/join-community/join-form.ts`（validateJoinForm/joinFormToPayload/OWNERSHIP_OPTIONS）+ `joinCommunity` 签名改 5 参 + 3 个 spec 测试文件全部为工作树未提交/未跟踪状态。`user.spec.ts` 注释声称"RED 阶段捕获了真实 FAIL 摘录（request.post 仅收到 {community_id}，缺 building/unit/room/ownership），见 CHANGELOG"——**但 CHANGELOG 没有任何实际 FAIL 输出文本**（仅"新增 TDD 测试（17 cases）"一行）。全仓 grep（CHANGELOG/request.md/loop-runs/work-records/_tdd_evidence.md）均无 vitest 断言失败文本（如 `AssertionError` / `expected spy to have been called with`）。

**结构性佐证（成立但不够）**：`git show HEAD:web/mobile/src/pages/join-community/join-form.ts` → `fatal: path exists on disk, but not in 'HEAD'`；`git show HEAD:web/mobile/src/api/user.ts` 旧签名 `joinCommunity(communityId)` 仅传 `{community_id}`。证明 RED 是真实失败，但无摘录。

**根因**：与 T2.2/restore 用例同一失败类——Generator 把 RED 失败"口头描述进注释/CHANGELOG"（此处写进 `user.spec.ts` 顶部注释并指向不存在的 CHANGELOG 摘录），未实际跑一次修复前测试并持久化失败文本；`_tdd_evidence.md` 仍只存在于 master-data-service，mobile 未创建。修复目标 #4（"TDD 证据强制捕获"）**第 3 次未闭环**。

**QA 判定**：TDD 证据表 RED 列全部 ❌（无 FAIL 摘录）→ QA FAIL（TDD 证据不足）。机械门禁全 PASS（build/type-check/test 17/17 exit 0）+ 结构性 RED 成立均不改变该判定。

**补救**：mobile 也须按本记忆做法——RED 阶段把真实 vitest 断言失败输出（`AssertionError:` / `- Expected` / `+ Received` / `Test:` 行）写入 CHANGELOG 或新建 `_tdd_evidence.md`；TDD 证据表 RED 列必须贴摘录原文。

## 复现场景（2026-08-16，community-hub-service content-post-generalization Task 1.1-1.23）

**现象**：通用图文发布重构（1763 增/1421 删，42 modified + 21 untracked）。CHANGELOG 2026-08-16 条声称"含逻辑函数任务均先写失败测试（RED）再实现（GREEN）… RED 摘录留档于测试注释（080006/080005/080002 映射、attachment_count 重算、is_pinned 操作者路径）"——**但全仓工作树 grep 不到任何实际 FAIL 输出文本**：
- 新测试文件（division_test / userctx_test / createcontentpostlogic_test / updatecontentpostlogic_test / read_write_logic_test / producer_test / rescanner_test / contentcompat_test / api_proxy_test / content_post*_test）注释只描述行为与设计决策（如"scope 越权 → 080006"），**无 `Error:` / `expected:... actual:...` / `undefined:` 摘录**。
- 本服务无 `_tdd_evidence.md`（仅旧 change 归档 `.harness/changes/access-data-permission/impl/community-hub-service/` 有）。
- 全仓 grep 到的 RED 摘录全部来自**旧变更**（08-12 的 `expected: 80006, actual: 0` / `undefined: JWTUserID`、08-13 的 `expected: 80007, actual: 0`），与本轮无关。

**结构性佐证（成立但不够）**：`git show HEAD:` 确认 division.go / producer.go / contentcompat.go / helper.go 新函数在 HEAD 均不存在（RED 真实），但无摘录。

**根因**：与 T2.2/restore/mobile 同一失败类——Generator 把 RED 失败"口头描述进 CHANGELOG/测试注释"，未实际跑一次修复前测试并持久化失败文本。测试本身质量极高（119 测试函数全绿、行为断言精确、覆盖面广），但过程证据缺位。

**QA 判定**：TDD 证据表 RED 列全部 ❌（~20 个有逻辑函数）→ QA FAIL（TDD 证据不足）。机械门禁 17 PASS / 1 FAIL（proto_ts_align 为预声明的 file-service 范围同步项）/ 3 WARN 不改变该判定。

**补救**：同前——Generator 用 `git stash` 回退生产文件复现真实 FAIL（`go test`/`go build` 输出含 `Error Trace`/`Error:`/`expected...actual...` 行号），持久化到 CHANGELOG 或新建 `services/community-hub-service/_tdd_evidence.md`。**本次为第 4 次复发，修复目标 #4（TDD 证据强制捕获）仍未闭环——建议管线层面把「RED 摘录存在性」纳入机械门禁（如 tdd-evidence-validator 扩展为按函数维度校验 RED 列），否则 Generator 口头描述将持续漏检。**
