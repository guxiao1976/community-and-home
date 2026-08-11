---
triggers: ["VITE_APP_TITLE", "系统名称", "环境变量", "AppHeader", "Login", "title"]
type: reference
severity: info
service: web
status: active
created: 2026-06-17
updated: 2026-08-09
apply_count: 0
---
系统名称已从硬编码改为通过 `VITE_APP_TITLE` 环境变量配置。涉及 4 个位置：

| 文件 | 位置 |
|------|------|
| `web/pc/src/components/layout/AppHeader.vue` | 顶部导航栏标题 |
| `web/pc/src/views/auth/Login.vue` | 登录页标题 |
| `web/pc/src/views/dashboard/Index.vue` | 欢迎语 |
| `web/pc/src/router/guards.ts` | 浏览器标签页标题 |

所有位置统一使用 `import.meta.env.VITE_APP_TITLE || '默认值'`。配置文件：`web/pc/.env`。修改后需重启 dev server 生效。
