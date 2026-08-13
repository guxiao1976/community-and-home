# Change: web-pc-debt-cleanup — web/pc 前端历史债务清理

## 背景
access-control 变更 Stage 5 跑 web/pc 时，build 门禁被 **122 条历史 TS 错误**（29 文件）挡住（stash 对照确认全在 HEAD，本次变更 0 新增）。根因是 `npm run type-check`/`npm run lint`（vue-tsc --noEmit）读根 tsconfig.json（`files:[]`+project references）检查 0 文件恒 exit 0，把 122 条错误长期掩盖；只有 `npm run build`（vue-tsc -b）能暴露。

另有 permission.spec.ts 2 条单测失败（历史遗留）。

## 问题清单
| # | 问题 | 证据 |
|---|------|------|
| 1 | 122 条 TS 错误（29 文件） | `vue-tsc -b` 报错，division/grassroots/AdminUserList/ModelList/roles 等 |
| 2 | type-check 假通过 | 根 tsconfig `files:[]`+references，vue-tsc --noEmit 检查 0 文件 |
| 3 | permission.spec.ts 2 条单测失败 | 历史遗留 |

## 修复目标
1. 清 122 条 TS 错误（逐文件，可能涉及历史类型漂移、未导入类型、DefaultRow/PaginatedResponse 等）
2. 修 type-check 假通过（改 script 用 `-p tsconfig.app.json` 或 `-b`）
3. 修 permission.spec.ts 2 条失败

## 验收
- `npm run build`（vue-tsc -b）0 错误
- `npm run test:unit` 全绿
- type-check 脚本真正检查源码（非假通过）

## 依赖
- access-control 的 web/pc 部分（platforms 配置 + 50007 引导）已实现但被 build 门禁挡住，待本变更清债后一并提交。

## 阶段
- [x] 0 request
- [ ] 1 定向审计
- [ ] 2 修复
- [ ] 3 回归
