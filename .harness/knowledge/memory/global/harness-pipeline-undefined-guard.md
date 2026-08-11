---
triggers: ["undefined", "harness-pipeline", "args", "serviceName", "参数校验"]
type: pitfall
severity: should-follow
service: all
status: active
created: 2026-06-17
updated: 2026-08-09
apply_count: 0
---

# Harness Pipeline: 参数校验防止 undefined 字符串化

## 问题
当 `harness-pipeline.js` 被调用时，如果 `args.serviceName`、`args.serviceDir` 或 `args.task` 未传入（JavaScript `undefined`），在模板字符串中会被字符串化为字面量 `"undefined"`，导致构造出无效路径如 `/home/jiaoxh/.../services/undefined/`。

## 触发条件
- Harness pipeline 调用时，`args` 对象缺少 `serviceName`、`serviceDir`、`task` 任一字段
- 自动化流程中服务名解析失败，传入了 JS `undefined` 值

## 复现场景
```
QA of service "undefined" FAILED — the service directory .../undefined/ does not exist
Service directory: NOT FOUND
```

## 修复
在 `harness-pipeline.js` 入口处添加三层参数校验：
1. 检查 `args.serviceName` / `args.serviceDir` / `args.task` 存在且不为 `undefined`（字符串或值）
2. 从 `serviceDir` 中提取裸服务名（如 `services/moderation-service` → `moderation-service`），验证其在已知的有效服务列表中
3. 验证失败时抛出明确的错误消息，列出可用服务

## 相关文件
- `.harness/workflows/harness-pipeline.js` — 添加了入口参数校验
