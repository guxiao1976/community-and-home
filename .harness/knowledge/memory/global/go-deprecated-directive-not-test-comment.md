---
triggers: ["Deprecated", "弃用标记", "go doc", "staticcheck", "SA1019", "golangci-lint", "自测注释", "占位注释", "流水线样例", "链路验证", "废弃指令"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-13
updated: 2026-08-13
---

# Go 的 `// Deprecated:` 是机器可读指令，禁止用作自测/占位注释

## 为什么会有这条经验

流水线正向自测时，给仍活跃使用的核心结构体 `UserBase` 加了一行 `// Deprecated: 流水线自测样例注释...`，被 Reviewer 捕获为缺陷：`Deprecated:` 是 Go 官方文档约定指令，不是自由文本。

1. `go doc` / pkg.go.dev 会把该符号标灰并显示「Deprecated」
2. `staticcheck`(SA1019) / golangci-lint 会对所有使用点报 `X is deprecated` 告警
3. IDE 会给所有引用处加删除线

把仍在生产使用的符号标成 Deprecated 会误导下游调用方（如 auth-service 依赖 user-service 的模型）以为该类型即将移除，并污染 lint/staticcheck 输出产生大量 SA1019 噪声。

## 正确做法

- 自测 / 占位 / 链路验证类注释用 `// NOTE:` 或普通注释，绝不使用 `// Deprecated:` 指令
- `Deprecated:` 仅用于确实计划移除/废弃的符号，且必须写明替代方案与废弃版本号
- 流水线链路验证样例完成后，清理临时注释，不要留在生产模型上

## 验证

- `grep -rn "Deprecated:" services/` 确认没有对活跃符号误用该指令
