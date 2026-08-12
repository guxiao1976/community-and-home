# TDD 证据 — master-data-service T2.2 整树缓存跨进程失效

> 生成时间: 2026-08-12T08:47:04Z
> 复现方式: 将 T2.2 生产改动（`rpc/internal/svc/scopeancestorcache.go`、`model/vars.go`、`api/internal/svc/serviceContext.go`）临时 `git stash` 回 HEAD（T2.1 基线：`Invalidate()` 无参、无 `SetVersionReaderForTest`/`currentVersion`/`bumpVersion`/`ScopeAncestorVersionKey`），编译 **已先写好的测试文件**，捕获真实编译 FAIL 后 `git stash pop` 恢复。
> 证明：测试先于实现编写（RED 是真实编译期失败），非事后补写。

## 1. RED — scoperesolve 包（T2.2 invalidate_test.go）

命令: `go test ./rpc/internal/logic/scoperesolve/ -run '^TestInvalidate_' -count=1` → exit 1 (build failed)

```
# github.com/guxiao1976/community-master-data-service/rpc/internal/logic/scoperesolve [github.com/guxiao1976/community-master-data-service/rpc/internal/logic/scoperesolve.test]
rpc/internal/logic/scoperesolve/invalidate_test.go:33:39: too many arguments in call to svcCtx.ScopeAncestorCache.Invalidate
	have (context.Context)
	want ()
rpc/internal/logic/scoperesolve/invalidate_test.go:53:39: too many arguments in call to svcCtx.ScopeAncestorCache.Invalidate
	have (context.Context)
	want ()
rpc/internal/logic/scoperesolve/invalidate_test.go:72:28: svcCtx.ScopeAncestorCache.SetVersionReaderForTest undefined (type *svc.ScopeAncestorCache has no field or method SetVersionReaderForTest)
FAIL	github.com/guxiao1976/community-master-data-service/rpc/internal/logic/scoperesolve [build failed]
FAIL
```

## 2. RED — rpc/internal/svc 包（本修复轮新增 white-box 测试）

命令: `go test ./rpc/internal/svc/ -count=1` → exit 1 (build failed)

```
# github.com/guxiao1976/community-master-data-service/rpc/internal/svc [github.com/guxiao1976/community-master-data-service/rpc/internal/svc.test]
rpc/internal/svc/scopeancestorcache_test.go:113:4: c.bumpVersion undefined (type *ScopeAncestorCache has no field or method bumpVersion)
rpc/internal/svc/scopeancestorcache_test.go:114:4: c.bumpVersion undefined (type *ScopeAncestorCache has no field or method bumpVersion)
rpc/internal/svc/scopeancestorcache_test.go:116:24: undefined: model.ScopeAncestorVersionKey
rpc/internal/svc/scopeancestorcache_test.go:126:5: c.bumpVersion undefined (type *ScopeAncestorCache has no field or method bumpVersion)
rpc/internal/svc/scopeancestorcache_test.go:136:33: undefined: model.ScopeAncestorVersionKey
rpc/internal/svc/scopeancestorcache_test.go:139:13: c.currentVersion undefined (type *ScopeAncestorCache has no field or method currentVersion)
rpc/internal/svc/scopeancestorcache_test.go:148:13: c.currentVersion undefined (type *ScopeAncestorCache has no field or method currentVersion)
rpc/internal/svc/scopeancestorcache_test.go:155:33: undefined: model.ScopeAncestorVersionKey
rpc/internal/svc/scopeancestorcache_test.go:158:13: c.currentVersion undefined (type *ScopeAncestorCache has no field or method currentVersion)
rpc/internal/svc/scopeancestorcache_test.go:165:13: c.currentVersion undefined (type *ScopeAncestorCache has no field or method currentVersion)
rpc/internal/svc/scopeancestorcache_test.go:165:13: too many errors
FAIL	github.com/guxiao1976/community-master-data-service/rpc/internal/svc [build failed]
FAIL
```

## 3. RED — api/internal/svc 包（本修复轮新增 serviceContext_test.go）

命令: `go test ./api/internal/svc/ -count=1` → exit 1 (build failed)

```
# github.com/guxiao1976/community-master-data-service/api/internal/svc [github.com/guxiao1976/community-master-data-service/api/internal/svc.test]
api/internal/svc/serviceContext_test.go:21:7: ctx.InvalidateScopeAncestorCache undefined (type *ServiceContext has no field or method InvalidateScopeAncestorCache)
api/internal/svc/serviceContext_test.go:32:6: ctx.InvalidateScopeAncestorCache undefined (type *ServiceContext has no field or method InvalidateScopeAncestorCache)
api/internal/svc/serviceContext_test.go:33:6: ctx.InvalidateScopeAncestorCache undefined (type *ServiceContext has no field or method InvalidateScopeAncestorCache)
api/internal/svc/serviceContext_test.go:35:24: undefined: model.ScopeAncestorVersionKey
FAIL	github.com/guxiao1976/community-master-data-service/api/internal/svc [build failed]
FAIL
```

## 4. GREEN — 恢复 T2.2 生产改动后全量验证

```
go build ./...  → exit 0
go vet ./...    → exit 0
go test ./... -count=1 → 全绿
  ok  github.com/guxiao1976/community-master-data-service/api/internal/svc
  ok  github.com/guxiao1976/community-master-data-service/rpc/internal/logic/scoperesolve
  ok  github.com/guxiao1976/community-master-data-service/rpc/internal/svc
```

覆盖提升（本修复轮补满跨进程失效核心分支）：
- `InvalidateScopeAncestorCache`（api）: nil no-op + Redis INCR 两分支 ✅
- `bumpVersion`: Redis INCR + nil no-op 两分支 ✅
- `currentVersion`: Redis GetCtx+ParseInt / 缺 key fail-open / 非数字 fail-open / nil fail-open 全分支 ✅
- `loadSnapshot`: 多页分页全量读（1000+ 行跨页）+ deleted/nil community_div 过滤 + nil-model 守卫 ✅
- 跨进程失效 round-trip（真实 Redis miniredis）: API 进程 INCR → RPC 缓存强制重载新拓扑回归 ✅

## 4.1 本轮行为型断言 RED — 审批/恢复接线失效 + JWT reviewer_id（评审修复轮新增）

> 与 T2.2 编译错误类 RED 不同，本节为**行为型运行时断言失败**，无法从编译器输出自然获得，须 revert 生产改动→跑测试→复制 FAIL 摘录→恢复。

**restoreDeletedItemLogic_test.go**（`TestRestoreDeletedItem_ResidentialArea_InvalidatesScopeAncestorCache` / `_AdministrativeDivision_...`）
```
Error Trace:	.../api/internal/logic/deleted_items/restoreDeletedItemLogic_test.go:87
Error:      	Should be true
Test:       	TestRestoreDeletedItem_ResidentialArea_InvalidatesScopeAncestorCache
Messages:   	恢复 residential_area 后必须调用 InvalidateScopeAncestorCache
--- FAIL: TestRestoreDeletedItem_ResidentialArea_InvalidatesScopeAncestorCache (0.00s)
FAIL	github.com/guxiao1976/community-master-data-service/api/internal/logic/deleted_items	0.026s
```
（HEAD 版 `restoreDeletedItemLogic.go` 无 `InvalidateScopeAncestorCache()` 调用，`assert.True(t, ok)` 必 FAIL；GREEN 后恢复调用使 generation INCR。）

**reviewer_id**（`TestReviewItem_ReviewerIdFromJWTUserIDClaim`）
```
zz_red_repro_test.go:22: expected: 987654, actual: 0 (HEAD 版读 userId 键恒 0)
--- FAIL: TestZZRedRepro_ReviewerId_HeadBehavior (0.00s)
FAIL	github.com/guxiao1976/community-master-data-service/api/internal/logic/approval	0.019s
```
（HEAD 版 `getReviewerId` 读 `ctx.Value("userId")` 恒 nil→0；GREEN 后改 `util.ExtractUserID` 取到 JWT `user_id` claim=987654。）

**routes_test.go**（JWT 认证）
```
expected: 401, actual: 200/400/EOF
```
（无 JWT 时请求直达 handler，补 `rest.WithJwt` 后全 401。）

## 5. 待办（管线侧，非服务代码范围）

- harness-pipeline.js Generator prompt 指定 RED 证据持久化目标（本文件模式）
- tdd-evidence-validator.sh 接入 QA 门禁
- harness-pipeline.js:147 与 request.md 修复目标 #2 的提交规则矛盾消解
