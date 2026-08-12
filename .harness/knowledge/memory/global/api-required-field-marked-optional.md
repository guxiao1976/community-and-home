---
triggers: ["optional", "goctl", "必填", "参数校验", "边界", "10040", "ownership", "JoinCommunity", "types.go", "json tag"]
service: all
severity: should-follow
type: guideline
status: active
created: 2026-08-12
updated: 2026-08-12
---

# API 层必填字段禁止误标 `,optional`

## 为什么会有这条经验

user-service 阶段③：JoinCommunity 在 RPC 层强制 ownership 必填（UNSPECIFIED→10040），但 API 层 types.go 中字段标注 `json:"ownership,optional"`，与代码注释「必填」矛盾，客户端缺省时在 API 边界不拦截、透传 0 才得到 10040。

## 怎么做

1. 一个字段「后端必填」时，API 层（go-zero types.go）不得标 `,optional` —— 移除 optional 让 httpx 在边界直接 400 拦截，错误更早更明确
2. 若因向后兼容确实需要 optional（如下游客户端未升级），必须在字段注释中写明过渡策略与到期移除时间，避免「注释说必填、tag 说可选」的静默不一致
3. 提交前 grep 变更中的 `,optional` 字段，逐一确认与 RPC/Proto 语义一致

## 怎么验证

- grep 变更 diff 中的 `,optional` 字段，与 RPC/Proto 必填语义核对
- 客户端缺省必填字段应在 API 边界即返回 400 而非透传后返回业务码

## 关联经验

[[goctl-logic-stubs]]
