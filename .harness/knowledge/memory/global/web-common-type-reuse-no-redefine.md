---
triggers: ["web/common", "复用", "重复定义", "CommunityMembership", "joinCommunity", "类型", "枚举", "enums.ts", "OWNERSHIP_OPTIONS", "权属", "共享层", "@common"]
service: all
severity: should-follow
type: guideline
status: active
created: 2026-08-12
updated: 2026-08-12
---

# 前端禁止重复定义 web/common 已有类型，新业务枚举应上移共享层

## 为什么会有这条经验

web/mobile/CLAUDE.md 规则#4「禁止在移动端目录中重复定义」在实践中被违反：`src/api/user.ts` 本地重定义了 `web/common/types/identity.ts` 已有的 `CommunityMembership`（且字段风格相反 snake_case vs camelCase，正是 api_field_align 检查根因之一）；新业务枚举（如 CommunityOwnership OWNED=1/RENTED=2 + 中文标签）被定义在页面目录而非 `web/common/constants/enums.ts` 既有范式。

## 怎么做

1. 复用已有共享类型：`import { CommunityMembership } from '@common/types/identity'`，wire snake_case 需在 store/映射层转换，不复制类型
2. 新业务枚举：先在 `web/common/types/*` 定义枚举、在 `web/common/constants/enums.ts` 导出 label 映射，各端经 @common 引用
3. 枚举魔数（1/2）不得在 validate/转换/模板多处硬编码，应引用唯一常量源

## 怎么验证

- grep 变更文件确认无 `interface X` 与 web/common 同名重复
- 业务枚举无 `const XXX_OPTIONS` 定义在页面/组件目录

## 触发场景

- 新增/修改 API 相关函数或业务枚举时
- 发现某接口字段同时存在于 web/common 与端内定义时

## 关联经验

[[frontend-business-rule-hardcode]]
