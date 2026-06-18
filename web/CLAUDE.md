# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 角色定位

这是 **前端层**，包含 Vue 3 + TypeScript 管理后台（`web/pc/`）、共享库（`web/common/`）和移动端占位（`web/mobile/`）。

## 子实例索引

| 目录 | CLAUDE.md | 职责 |
|------|-----------|------|
| `pc/` | [pc/CLAUDE.md](pc/CLAUDE.md) | Vue 3 管理后台 — 详细文档 |
| `mobile/` | [mobile/CLAUDE.md](mobile/CLAUDE.md) | Uni-app (Vue 3) 移动端 — H5 + 小程序 |

## 关键规则

1. **禁止在前端定义业务逻辑** — 所有业务规则属于后端服务，前端只做展示和交互
2. **API 接口必须与 api-proto 一致** — 接口契约定义在 `../api-proto/`，前端类型在 `common/types/`
3. **禁止直接修改 `api-proto/`** — 需改接口时告知用户切换到全局 Claude
4. **只能调用本服务负责范围内的 API**，不直接访问其他服务数据库
5. **所有 ID 字段 TypeScript 类型必须为 `string`** — Snowflake 19 位 ID 超过 JS 安全整数范围，后端 JSON 序列化为字符串。`web/common/types/` 中所有 `id`/`userId`/`parentId` 等字段均使用 `string` 类型

## 目录总览

```
web/
  pc/             # Vue 3 管理后台 → 详见 pc/CLAUDE.md
  mobile/         # Uni-app (Vue 3) 移动端 → 详见 mobile/CLAUDE.md
  common/         # 前端共享库（types, constants, utils）
```
