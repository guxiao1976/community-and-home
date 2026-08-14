# 角色列表列宽重排（Role List Column Layout） Specification

## Purpose

修复角色管理列表表格因 ID 列固定宽度过宽（200px，实际 ID 仅 1-2 位数字）挤压角色名称/编码/描述列导致换行的布局问题；按决策 D7 重排各列宽度，让文本列自适应、操作按钮单行并排，保持表格可读性与既有功能不回退。

## Requirements

### Requirement: REQ-LAYOUT-1 — 列宽按 D7 方案整体重排

The `web/pc/src/views/roles/List.vue` table column widths SHALL be re-laid out as follows: ID column narrowed to approximately 70px; 操作 column narrowed from 380px to approximately 260px（编辑/权限配置/查看用户/删除 four link-buttons side by side）; 角色名称 / 角色编码 / 描述 columns SHALL use adaptive sizing（min-width, not a fixed narrow width）so content does not wrap; 系统角色 / 状态 / 允许登录端 / 创建时间 columns SHALL keep their fixed widths.

- **GIVEN** the role list renders rows whose IDs are 1-2 digits and whose name/code/description are typical Chinese strings
- **WHEN** the table renders at default viewport width
- **THEN** the ID column is narrow（≈70px）, name/code/description do not wrap, and the four action buttons sit on one line without overflow

- **GIVEN** a role whose description is very long
- **WHEN** the table renders
- **THEN** the description is truncated with an ellipsis（`show-overflow-tooltip`）and the row layout is not broken

### Requirement: REQ-LAYOUT-2 — 列宽调整不得回退既有功能

The layout change SHALL preserve all existing interactions of the table: 编辑（enabled for system roles per D1）、权限配置、查看用户、删除（disabled for system roles）buttons and the pagination controls SHALL continue to work unchanged.

- **GIVEN** the table has been re-laid out
- **WHEN** a user clicks 编辑 / 权限配置 / 查看用户 / 删除 on a row
- **THEN** each action behaves as before（system-role delete stays disabled）and the create/edit dialog opens correctly

- **GIVEN** the table has been re-laid out and pagination exists
- **WHEN** the user changes page size or page number
- **THEN** `loadRoles` re-fetches data and the new widths render consistently for the next page
