## Context

住宅小区模块后端 Go 代码（model、API handler、proto、RPC logic）和前端页面（Vue views、router、API service、菜单）已经全部基于 `md_residential_area` 表和新增字段实现完毕。但数据库层存在不一致：

- `masterdata_schema.sql` 仍定义旧表 `md_community`，缺少 `county_id`、`street_id`、`community_div_id`、`code` 字段
- `masterdata_seed.sql` 仍向 `md_community` 插入种子数据
- 部分文档仍引用旧表名

旧表 `md_community` 使用单一的 `division_id` 关联行政区划（level 5 社区），新表 `md_residential_area` 改为三级独立字段（`county_id`、`street_id`、`community_div_id`）+ `code` 编码字段，支持更灵活的区划筛选和唯一标识。

## Goals / Non-Goals

**Goals:**

- 将 SQL schema 中 `md_community` 表替换为 `md_residential_area`，包含所有新字段
- 更新种子数据以匹配新表结构
- 同步更新文档中的旧表名引用
- 确保建表脚本能直接运行，无需手动修改

**Non-Goals:**

- 不修改后端 Go 代码（已就绪）
- 不修改前端代码（已就绪）
- 不做数据迁移脚本（项目尚未有生产数据，直接用新 schema 初始化即可）
- 不修改 `md_administrative_division` 或其他无关表

## Decisions

### 1. 直接替换建表定义，不做迁移脚本

**选择**: 在 `masterdata_schema.sql` 中删除 `md_community` 建表语句，替换为 `md_residential_area` 建表语句。

**替代方案**: 保留 `md_community` + `ALTER TABLE` 迁移脚本。

**理由**: 项目处于开发阶段，无生产数据需要迁移。直接使用新 schema 更简洁，避免维护迁移脚本的复杂性。

### 2. 种子数据补充新字段值

**选择**: 将种子数据 INSERT 改为 `md_residential_area`，每条记录根据行政区划层级关系推导 `county_id`（level 3）、`street_id`（level 4），`community_div_id` 保持原 `division_id` 的值，`code` 使用行政区划编码后缀生成。

**理由**: 种子数据中的行政区划层级关系已知（从 seed data 的 parent_id 可推导），可直接填充。`code` 字段使用如 `RA-110108001001` 格式保持一致性。

### 3. 外键策略调整

**选择**: `md_residential_area` 对 `county_id`、`street_id`、`community_div_id` 均设置外键指向 `md_administrative_division(id)`。旧表的单字段外键 `fk_community_division` 替换为三个独立外键。

**理由**: 三级区划字段各自独立引用行政区划表，确保引用完整性。

## Risks / Trade-offs

- **[风险] 旧表 `md_community` 在已有开发环境中可能已存在** → 在 schema 脚本中增加 `DROP TABLE IF EXISTS md_community` 语句，确保干净切换
- **[风险] 种子数据的区划 ID 硬编码** → 种子数据仅用于开发环境，且与行政区划种子数据保持一致，风险可控
