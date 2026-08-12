# 交接文档:管线修复 + 残留收尾 → Wave 1 执行

> 生成时间: 2026-08-12
> 目的: 本会话完成「管线修复 + 历史残留收尾」,新会话在此基线上执行 access-control 变更① Wave 1。

## 1. 已完成并提交(新会话基线)

### 管线修复(6 项,验证通过)
1. generator 去 worktree 隔离 → 直接改主树(harness-pipeline-core.js)
2. TDD 检查范围 → 工作树 diff(harness-checks.sh)
3. generator 任务执行铁律(严格按 tasks.md、不 commit、RED→GREEN)(generator.js)
4. QA 审查范围 → 工作树 diff(qa.js)
5. failingLenses 作用域 bug(harness-pipeline-core.js)— 回归抓出
6. ast-checks 崩溃 bug(set -eu 下命令替换 exit 1)(harness-checks.sh)— 收尾抓出

### 小任务回归验证(已通过)
- master-data T2.1 ResolveScopeAncestors 单任务跑管线:实现落主树、QA PASS、1 轮收敛、无 worktree 泄漏
- 管线在「2/3 PASS 但个别视角异议」路径正确走 need_human 通知(验证第 5 修)

### 历史残留收尾(4 服务 + 根仓库,全部提交)
| 仓库 | 提交 | 说明 |
|------|------|------|
| 根仓库 | de3d9a2 起,至 91a3e9b | 管线第6修 + changes/ + identity.ts + 子模块指针 |
| common | 0a5216c | 统一响应 + 5位错误码 + BigInt ID + configx + 权限中间件 |
| master-data | 6cbc767 + 519e671 | ResolveScopeAncestors + JWT claim键统一 + 时间字段DB对齐;丢弃 RBAC/Outbox 半成品 |
| ai-model | 3c42741 + 9d111b4 | configx回退远程 + 丢弃apikey/template半成品;删 7.9GB venv |
| moderation | c0fffa0 | 规范对齐 + C3回调 + 时间字段迁移补齐 + json_string修复 |

门禁:master-data 16/16、moderation 16/16、common 全绿;9 服务构建全过。

## 2. 交接时需注意

1. **ResolveScopeAncestors(T2.1)已实现并提交**(master-data 6cbc767):
   - `rpc/internal/logic/scoperesolve/resolvescopeancestorslogic.go` + `_test.go`(8 用例)
   - `rpc/internal/svc/scopeancestorcache.go`(整树缓存 TTL30min)
   - 已在 `masterdataserver.go` 注册
   - Wave 1 重跑 T2.1 时应识别已有实现(或人工跳过,避免重复)

2. **master-data 时间字段 DB 契约已对齐**(6cbc767):db tag 用 `deleted_at`/`created_at`(Go 字段名仍 CreatedTime 不变),migration 006 已对真库执行。**不要回退这些改动。**

3. **JWT claim 键统一已保留**(6cbc767):`api/internal/util/userctx.go:ExtractUserID`(user_id 优先 + userId 兜底 + 兼容 json.Number/int64/float64)。auth-service 签发 user_id,消费方读 user_id。这是修复真实 bug 的资产,勿丢。

4. **common breaking 已提交**(0a5216c):
   - errx 错误码 3位→5位(400→99400 等)
   - responsex `Message→Msg`、model ID `,string`
   - 前端/客户端按旧数值匹配会断,前端契约需评估

5. **go.work 已移除 ai-model-service/pkg 模块**(根仓库 de3d9a2):ai-model 恢复 api/rpc 双模块,构建通过。

6. **.harness/backups/ 已 gitignore**(307M 残留备份,不提交,可安全删除或保留)。

## 3. 遗留待办(超收尾范围)

- ai-model 双模块结构 → harness-checks 需适配(registry services.json 中 ai-model-service module 为空)
- common `pkg/middleware/permission.go` userID==0 fail-open → Owner 收紧为 fail-closed
- master-data 2 处 WARN(api_smoke 未跑、svc loadSnapshot 覆盖低)

## 4. Wave 1 执行入口

```bash
# 按 CLAUDE.md 约束:先 dispatch Skill 做入口判定与工作量分级
# 变更文档: .harness/changes/access-data-permission/
# 阶段: ①数据权限核心(permission T1.1-T1.8 + master-data T2.1-T2.3)
# 任务类型: feature(taskType:feature, 强制 TDD)
# 注意: 子 Claude 禁止改 api-proto/(已有契约已提交 031f4e4)
```

## 5. 验证清单(新会话启动后确认)

- [ ] `git status` 干净(仅 my-ralph-project 的 .ralph 状态文件,无关)
- [ ] 根仓库 `git log --oneline -1` = 91a3e9b
- [ ] master-data `git log --oneline -1` = 519e671
- [ ] common `git log --oneline -1` = 0a5216c
- [ ] ai-model `git log --oneline -1` = 9d111b4
- [ ] moderation `git log --oneline -1` = c0fffa0
- [ ] `bash .harness/skills/qa/scripts/harness-checks.sh --service <name>` 4 服务全绿
