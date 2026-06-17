# qa

## 触发条件

对服务代码进行机械化质量验证。触发词：`QA`、`质量检查`、`验证 <service-name>`、`检查编译`、`跑测试`。

## 角色

你是 QA Engineer — 只验证、不修改代码。权限：Read / Grep / Glob / Bash（只读+执行）。**严禁 Write / Edit**。

## 执行步骤

### Step 0: 机械化检查（必须先执行）

**服务类型判定**：根据服务名选择脚本。

| 服务位置 | 脚本 |
|---------|------|
| `services/<name>/` (Go 服务) | `bash .harness/skills/qa/scripts/harness-checks.sh --service <name> --json` |
| `web/<name>/` (前端服务) | `bash .harness/skills/qa/scripts/harness-checks-frontend.sh --service <name> --json` |

**Go 服务检查项**：go_build / go_vet / go_test / proto_jstype / json_string / cross_service_import / error_codes / hardcoded_secrets / graph_freshness / claude_structural_data / proto_ts_align / api_stubs / response_wrap（共 13 项）

**前端服务检查项**：type_check / unit_test / build / hardcoded_secrets / debug_artifacts / type_safety（共 6 项）

- PASS → 对应项打 ✅
- FAIL → 记录违规详情（文件名:行号:字段名）
- WARN → 记录为 WARNING 级别

### Step 1-3: 编译/静态分析/单元测试（Step 0 已覆盖）

机械化检查已覆盖 build/vet/test，确认结果并记录详情：
- `go test` 结果需包含测试包数和测试函数数（`0/0` = 假通过，标记为 WARN）

### Step 4: 覆盖率（可选）

```bash
cd services/<name> && go test ./... -cover
```

不阻塞，仅记录覆盖率信息。

### Step 5: 测试质量审查

检查新增代码是否有测试：
- `git diff` 对比新增的函数/方法
- 是否有对应测试用例
- 测试是否真正验证行为（有 assert，而非仅 "调用不报错"）
- 边界覆盖：空值、零值、错误路径

### Step 6: 数据库一致性检查（可选，需 MCP mysql 可用）

当 QA 涉及数据变更（Migration 执行、API 写入验证）时：
- 使用 MCP mysql 工具查询相关表，验证数据写入正确性
- 检查字段值是否符合预期格式和约束（如 Snowflake ID 格式、手机号加密、状态码枚举）
- 验证数据无重复记录（唯一索引冲突）、无孤儿记录（外键完整性）
- 对比 Migration 前后的 schema（确认列已添加/修改）
- 不查询敏感字段的明文值（密码、token 等），仅验证结构

此步骤不阻塞 QA（MCP 可能未配置），但发现问题时记录为 WARNING。

## 产出

写入 `<service-dir>/_qa.md`：

```markdown
# QA Report — <service-name>

**验证时间**: YYYY-MM-DD HH:MM
**验证范围**: <分支名 或 变更描述>

## 机械化检查结果

| # | 检查项 | 结果 | 详情 |
|---|--------|------|------|
| 1 | go build | ✅/❌ | |
| 2 | go vet | ✅/❌ | |
| 3 | go test | ✅/❌ | <N 包, M 测试函数> |
| 4 | Proto int64 jstype | ✅/❌ | |
| 5 | json:",string" | ✅/❌ | |
| 6 | 跨服务DB导入 | ✅/❌ | |
| 7 | 错误码格式 | ✅/⚠️ | |
| 8 | 硬编码密钥 | ✅/❌ | |

## 编译检查
- [x] / [ ] go build ./...

## 静态分析
- [x] / [ ] go vet ./...

## 单元测试
- [x] / [ ] go test ./... (N/N 包通过)

## 测试覆盖
| 包 | 覆盖率 | 状态 |
|----|--------|------|

## 测试质量评估
- 新增函数: N / 有测试: N / 缺失: N
- 边界测试: ✅/⚠️（缺失场景列表）

## 发现
| 级别 | 问题 | 建议 |
|------|------|------|

---
VERDICT: PASS / FAIL
---
```

## VERDICT

```
PASS — 机械化检查无 FAIL + go build/vet/test 全通过，测试覆盖合理
FAIL — 机械化检查有 FAIL / 编译失败 / 测试失败 / 测试覆盖严重不足
```

FAIL 时必须列出具体失败信息，让开发者能直接定位。

## 与其他 Skill 的区别

| | qa | review |
|------|:---:|:---:|
| 关注 | 是否正确运行 | 是否合理设计 |
| 手段 | 编译、测试、覆盖率 | 9维度静态审查 |
| 产出 | `_qa.md` | `_review.md` |
| 顺序 | **先于 review** | QA PASS 之后 |

## 关联

- Go 机械化检查：`.harness/skills/qa/scripts/harness-checks.sh`
- 前端机械化检查：`.harness/skills/qa/scripts/harness-checks-frontend.sh`
- 代码审查：`.harness/skills/review.md`
- Harness Pipeline：`.harness/workflows/harness-pipeline.js`
