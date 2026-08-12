---
triggers: ["错误码", "门禁", "harness-checks", "errx", "NewBaseRespWithError", "魔数", "magic number", "裸数字", "080006", "80001"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# QA 错误码门禁盲区：responsex 包装的错误码裸字面量不被检测

## 为什么会有这条经验

community-hub access-data-permission 阶段④ 新增 `updatemoderationstatuslogic.go` 用 `responsex.NewBaseRespWithError(80001, ...)` / `(80004, ...)` 返回错误，数字字面量未走命名常量。

QA 的 `harness-checks.sh check_error_codes` 只 grep `errx.(NewCodeError|Wrap|Wrapf)(\s*\d+` 裸整数，`responsex.NewBaseRespWithError(80001, ...)` 完全绕过检测 → 机械化检查全绿但新代码引入魔数。

## 怎么做

1. 新增错误响应时优先用命名常量：同一服务内错误码用常量/枚举，禁止 `NewBaseRespWithError(<裸数字>, ...)`；若 RPC 层因依赖方向无法引用 api/types 常量，在 rpc 侧建等价命名常量（如 scope.CodePublishScopeDenied 的写法）
2. QA 门禁扩展（Owner/QA 侧）：`check_error_codes` 增加对 `responsex.NewBaseRespWithError(\s*\d+` 与 `NewBaseRespFromError(\s*\d+` 的检测
3. 新错误码处补 SEE 引用 `[[error-code-collision-and-namespace-alignment]]`

## 怎么验证

```bash
grep -rnP 'responsex\.NewBaseRespWithError\(\s*\d+' services/<name>/ --include='*.go'
# 期望：无裸数字（全命名常量/0）
```

## 关联经验

[[error-code-collision-and-namespace-alignment]]
