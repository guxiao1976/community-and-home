---
triggers: ["TDD", "RED", "GREEN", "测试未提交", "测试证据", "undefined", "编译失败", "QA FAIL", "证据摘录"]
type: process
severity: must-follow
service: all
status: active
created: 2026-08-12
updated: 2026-08-12
apply_count: 1
---
# QA 的 TDD RED 证据必须包含实际 FAIL 输出摘录

## 为什么会有这条经验

1. 管线 Generator 直接改主树、不提交（harness-pipeline 设计如此），RED 阶段的编译/测试失败只被口头描述进 CHANGELOG（"RED（编译失败：`Invalidate(ctx)` 签名变更 + `SetVersionReaderForTest` 未定义）"），**没有持久化任何实际错误输出**。
2. QA（master-data-service T2.2 整树缓存跨进程失效）时，要求 TDD 证据表 RED 列给出实际 FAIL 摘录（如 `undefined: SetVersionReaderForTest` / `too many arguments`），但全仓 grep 不到任何编译错误文本（reflog/stash/loop-runs/request 均无）。
3. 结果是：GREEN 全绿 + 机械门禁 16/16 PASS，但 RED 证据列 ❌ → 按规则判定 QA FAIL（TDD 证据不足）。
4. `.harness/changes/harness-pipeline-fix/request.md` 第 4 行已记录此问题（"TDD 证据缺失(RED→GREEN 无摘录,测试未 commit) → QA TDD FAIL"），修复目标 #4 为"TDD 证据强制捕获"——但 T2.2 仍未根治，证明该修复未闭环。

## 怎么做

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
