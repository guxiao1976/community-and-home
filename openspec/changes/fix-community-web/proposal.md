## Why

住宅小区模块的数据表已从 `md_community` 变更为 `md_residential_area`，并新增了 `county_id`、`street_id`、`community_div_id`、`code` 等行政区划关联字段。后端 Go 模型、API、Proto 已按新表结构调整完毕，但 SQL 建表脚本和种子数据仍引用旧表 `md_community`，部分文档也未同步更新，导致数据库层与代码层不一致，无法正常运行。

## What Changes

- **SQL Schema**: 将 `masterdata_schema.sql` 中的 `md_community` 表定义替换为 `md_residential_area`，包含新增的 `county_id`、`street_id`、`community_div_id`、`code` 字段
- **Seed Data**: 将 `masterdata_seed.sql` 中的 INSERT 语句从 `md_community` 改为 `md_residential_area`，并补充新字段的种子数据
- **文档同步**: 更新 `PROJECT_STRUCTURE.md`、`specs/001-identity-masterdata/data-model.md`、`specs/001-identity-masterdata/tasks.md`、`specs/001-identity-masterdata/contracts/masterdata-api.md` 中的旧表名引用
- **前端菜单确认**: 确认前端菜单入口为"住宅小区"，路由和视图组件正确对应

## Capabilities

### New Capabilities

(无新增能力)

### Modified Capabilities

- `residential-area`: 将数据库 schema 从 `md_community` 完全迁移到 `md_residential_area`，确保 SQL 脚本、种子数据、文档与已实现的后端代码和前端页面一致

## Impact

- **数据库**: `scripts/sql/masterdata_schema.sql`、`scripts/sql/masterdata_seed.sql` — 表结构定义和数据初始化
- **文档**: `PROJECT_STRUCTURE.md`、`specs/001-identity-masterdata/` 下的 data-model.md、tasks.md、contracts/masterdata-api.md
- **前端**: 菜单和路由已正确配置为"住宅小区"，无需改动
- **后端 Go 代码**: model、API handler、proto、RPC logic 已使用 `md_residential_area`，无需改动
