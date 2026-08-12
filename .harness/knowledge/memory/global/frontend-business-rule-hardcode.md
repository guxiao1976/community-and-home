---
triggers: ["前端", "校验", "业务规则", "区间", "building", "unit", "room", "ownership", "硬编码", "proto", "漂移", "表单校验"]
service: all
severity: should-follow
type: pitfall
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 前端硬编码业务规则区间易与 proto/后端漂移

## 为什么会有这条经验

前端表单校验若硬编码业务规则区间（如楼号 1-150 / 单元 1-5 / 房号 100-999 / 权属枚举 1/2），这些值通常来自 proto 注释。当后端调整范围时，前端会静默拒绝合法输入或放行非法值，形成「前端比后端更严/更松」的漂移，且 QA/测试难察觉。

## 怎么做

1. 前端校验仅作 UX 提示，保持纯展示层校验，权威校验始终在后端
2. 若区间需多次复用，应从后端配置/常量接口下发，避免在 TS 中二次定义
3. 新增此类校验时，须同步核对对应 proto 字段注释（如 api/user/v1/user.proto），并在 CHANGELOG 标注所对齐的 proto 字段与枚举

## 怎么验证

- grep 前端校验代码中的魔法区间数字，确认与对应 proto 注释一致
- 后端调整范围后前端无需改动即可生效

## 关联经验

[[api-required-field-marked-optional]]
