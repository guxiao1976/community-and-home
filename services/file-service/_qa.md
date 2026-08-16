# QA Report — file-service

**验证时间**: 2026-08-16 11:30 (CST)
**验证范围**: 工作树未提交改动 + 未跟踪文件（通用图文发布附件安全重构 content-post-generalization Task 2.1-2.6）

## 机械化检查结果 (harness-checks.sh — FRESH run)

`bash .harness/skills/qa/scripts/harness-checks.sh --service file-service --json` → **18 PASS / 0 FAIL / 3 WARN，exit_code 0**

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅ | PASS — compilation succeeded (exit 0) |
| 2 | go vet | ✅ | PASS — no issues (exit 0) |
| 3 | go test | ⚠️ WARN | 3P/0F/15N, ~32 tests，全绿；WARN：NEW packages missing tests: api/internal/logic, api/internal/logic/file, rpc/internal/errx（本次仅新增错误码常量，见「发现」） |
| 4 | go_fmt | ✅ | PASS — 变更 Go 文件全部已格式化 |
| 5 | Proto int64 jstype | ✅ | PASS — diff 无 proto 变更（skipped） |
| 6 | json:",string" | ✅ | PASS — 所有 int64 ID 字段均有 json:",string"（AST 验证） |
| 7 | 跨服务DB导入 | ✅ | PASS — 无违规（diff 扫描 16 个 Go 文件） |
| 8 | 错误码格式 | ✅ | PASS — 无 magic numbers（均用命名常量或 0） |
| 9 | 硬编码密钥 | ✅ | PASS — 未发现密钥 |
| 10 | Knowledge graph freshness | ✅ | PASS — graph 已同步（0h 前） |
| 11 | CLAUDE.md structural data | ✅ | PASS — 无结构数据重复 |
| 12 | Proto→TS 对齐 | ⚠️ WARN | 与本次变更无关：identity.ts/moderation.d.ts 存在 TS 滞后 proto 字段（LoginSmsRequest.phone 等 5 项，属前端同步欠账；本次 diff 无 proto 变更） |
| 13 | API TODO stubs | ✅ | PASS — 无 TODO stub |
| 14 | Response single-wrap | ✅ | PASS — 无双重包装风险 |
| 15 | Benchmark regression | ✅ | PASS — 无 benchmark 函数（SKIP） |
| 16 | API smoke test | ✅ | PASS — diff 无新路由 |
| 17 | Memory Index Freshness | ✅ | PASS — 索引最新 |
| 18 | Design/code consistency | ✅ | PASS — model 列覆盖标准迁移源 |
| 18b | Git hygiene | ⚠️ WARN | 既有基础设施欠账：api-proto gitlink 无 .gitmodules 条目（与本次变更无关） |
| 19 | Mutation testing | ✅ | PASS — 未解析到分数（SKIP） |
| 20 | Pipeline evals | ✅ | PASS — 管线 eval 全部通过 |

## 编译检查
- [x] go build ./... — **exit 0**（FRESH run）

## 静态分析
- [x] go vet ./... — **exit 0，clean output**（FRESH run）

## 单元测试
- [x] go test ./... -count=1 — **exit 0**（FRESH run，禁用缓存）
- 含测试包：3（internal/guard、model、rpc/internal/logic/file）
- 测试函数：**32**，失败 0
- 注：api/*、rpc/internal/errx 等包为 [no test files]（非本次逻辑变更）

## 测试覆盖 (go test ./... -cover -count=1)

| 包 | 覆盖率 | 状态 |
|----|--------|------|
| internal/guard | 98.0% | ✅ 高覆盖（白名单/大小/override/sniff 全分支） |
| model | 20.0% | ⚠️ 仅新增 Insert/FindOne 两测试（存量方法未覆盖，非本次范围） |
| rpc/internal/logic/file | 25.8% | ⚠️ 本次新增的 GetUploadUrl L1 / ConfirmUpload L2 路径已覆盖；包内其他旧逻辑未覆盖 |

## TDD 证据检查（分诊：字段映射 vs 有逻辑函数）

> **RED 证据获取方式说明**：Generator 改动直接落在工作树、未提交，且我已在 .harness/tasks、loop-runs、docs/devlog 全库检索——**未找到 Generator 持久化的 RED FAIL 摘录**（CHANGELOG 声称「已按 TDD 留 RED 证据」，但仓库内无记录）。
> 为满足「RED 列必须有具体 error 文本」硬标准，QA 在 scratch git worktree（HEAD=bddeadc，预实现状态）+ 新测试文件上 **FRESH 复现 RED**：`git worktree add /tmp/file-service-red HEAD` → 放入新测试 → 移除实现 → `go test`，得到真实 FAIL 输出（下述 RED 均为复现摘录，非 `git show` 结构性推断）。复现后 worktree 已 `git worktree remove` 清理，主工作树未被修改。

| 新增/修改函数 | 类型 | 是否有测试 | RED 确认 | GREEN 确认 | 状态 |
|-------------|------|:---:|:---:|:---:|:---:|
| guard.ValidateFileName (whitelist.go:108) | 有逻辑 | ✅ | ✅ `whitelist_test.go:18:16: undefined: ValidateFileName` | ✅ 测试全绿 | PASS |
| guard.ValidateFileNameForEntityType (whitelist.go:113) | 有逻辑 | ✅ | ✅ `whitelist_test.go:113:14: undefined: ValidateFileNameForEntityType` | ✅ | PASS |
| guard.validateFileName (whitelist.go:117, unexported) | 有逻辑 | ✅ 经导出函数覆盖 | ✅ 同上包 build FAIL（undefined 集） | ✅ | PASS |
| guard.ValidateFileSize (whitelist.go:138) | 有逻辑 | ✅ | ✅ `whitelist_test.go:80:9: undefined: ValidateFileSize` | ✅ | PASS |
| guard.RegisterEntityTypeOverride (whitelist.go:72) | 有逻辑 | ✅ | ✅ `whitelist_test.go:101:2: undefined: RegisterEntityTypeOverride` | ✅ | PASS |
| guard.allowedExtensions (whitelist.go:81, unexported) | 有逻辑 | ✅ 经 override 测试覆盖 | ✅ 同上包 build FAIL | ✅ | PASS |
| guard.SniffType (magic.go:30) | 有逻辑 | ✅ | ✅ `magic_test.go:30:15: undefined: SniffType` | ✅ | PASS |
| guard.utf16LE (magic.go:58, unexported) | 有逻辑 | ✅ | ✅ `magic_test.go:39:49: undefined: utf16LE` | ✅ | PASS |
| ConfirmUploadLogic.ConfirmUpload (confirmuploadlogic.go:32) | 有逻辑 | ✅ | ✅ `confirmuploadlogic_test.go:74:14: l.confirmWithReader undefined (type *ConfirmUploadLogic has no field or method confirmWithReader)` | ✅ | PASS |
| ConfirmUploadLogic.confirmWithReader (confirmuploadlogic.go:47) | 有逻辑 | ✅ | ✅ 同上 `undefined: confirmWithReader` | ✅ | PASS |
| helper.verifySniffedContent (helper.go:34) | 有逻辑 | ✅ 经 confirmWithReader 测试间接覆盖 | ✅ 被 confirmWithReader RED 囊括（仅经其可达，实现前无直接符号） | ✅ | PASS |
| helper.readObjectPrefix (helper.go:49) | 有逻辑 | ✅ 同上 | ✅ 同上 | ✅ | PASS |
| helper.extMatch (helper.go:58) | 有逻辑 | ✅ TestConfirmUpload_DeclaredJpeg_SniffJpg_EquivalentAllowed | ✅ 同上 | ✅ | PASS |
| GetUploadUrlLogic.GetUploadUrl L1 守卫 (getuploadurllogic.go:32) | 有逻辑 | ✅ | ✅ 运行时断言 FAIL：`TestGetUploadUrl_L1RejectsBadExtension: Error: Should be true / Messages: exe → 070004，未触碰 MinIO`（另 3 项同断言失败） | ✅ | PASS |
| model.File.FileType/Confirmed 字段 (file.go) | 字段映射 | ✅ | —（不要求）复现：`filemodel_test.go:56:3: unknown field FileType in struct literal of type File` | ✅ | PASS |
| model.FileModel.Insert 新列接线 (filemodel.go) | 字段映射 | ✅ TestFileModel_Insert_IncludesFileTypeConfirmed | —（不要求） | ✅ | PASS |
| helper.toProtoFile FileType/Confirmed 透出 (helper.go) | 字段映射 | ✅ TestToProtoFile_FileTypeConfirmed | —（不要求） | ✅ | PASS |
| 错误码 70004/70005 四文件登记 | 常量登记 | ✅ 测试中引用 | —（不要求）复现：`undefined: ErrCodeUnsupportedFileType` | ✅ harness error_codes PASS | PASS |

**GREEN（FRESH）**: `go test ./... -count=1` → `internal/guard ok` / `model ok` / `rpc/internal/logic/file ok`，TEST_EXIT=0。

## 测试质量评估
- 新增/修改函数：18；有测试：18；缺失：0
- 断言质量：✅ 均用 `require`/`assert` + `errx.IsCode` 校验错误码（070004/070005），非「仅调用不报错」
- 边界覆盖：✅ 空值/零值（无扩展名、点文件）、大小边界（=10MB 放行、>10MB 拒绝）、错误路径（伪装扩展名：exe→png、msi→doc、xlsx→docx、通用 zip 拒绝）、override 弱化防护（禁止集不可放行、10MB 不可放宽、全局基线不受影响）、不落库断言（拒绝时 `inserted==nil`）
- 弱点：model / rpc 包覆盖率偏低（20%/25.8%），仅覆盖本次新增路径，属既有包无测试的历史现状

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|
| ⚠️ WARNING | 三个「新增包」无测试：api/internal/logic、api/internal/logic/file、rpc/internal/errx——但本次对它们的改动仅为错误码常量登记（无逻辑），且 error_codes 检查 PASS；不构成 FAIL | 可接受；后续错误码使用逻辑落地时补单测 |
| ⚠️ WARNING | Generator 未持久化 RED FAIL 摘录（CHANGELOG 声称已留，仓库实际无记录），QA 靠复现补证 | 管线应要求 Generator 将 RED 输出落盘到 task 执行记录 |
| ⚠️ WARNING | proto_ts_align / git_hygiene 两个 WARN 与本次变更无关（TS 同步欠账、api-proto gitlink 欠账） | 转前端/基础设施 backlog |
| ⚠️ WARNING | `ConfirmUpload.ConfirmUpload` 对 `RawMinio.GetObject` 错误与 `obj.Close()` 错误处理：GetObject 返回 err 时未 Close（此时无对象句柄），Close 错误被忽略——属 L2 校验流程既有取舍，不影响本 QA 判定 | 若需严谨可补 obj.Close 错误日志 |

---
VERDICT: **PASS**
---
