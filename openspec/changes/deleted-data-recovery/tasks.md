## 1. 后端 - 已删除记录查询 API

- [x] 1.1 在 masterdata API 定义文件中新增 `GET /api/masterdata/deleted-items` 和 `GET /api/masterdata/deleted-items/counts` 路由
- [x] 1.2 在 masterdata API 定义文件中新增 `POST /api/masterdata/deleted-items/:entity_type/:id/restore` 路由
- [x] 1.3 创建 deleted items handler（查询列表 handler + 查询统计 handler + 恢复 handler）
- [x] 1.4 创建 deleted items logic：查询已删除记录（UNION 多表查询，按 delete_time DESC 排序，支持分页）
- [x] 1.5 创建 deleted items logic：查询各类型已删除记录数量统计
- [x] 1.6 创建 deleted items logic：恢复记录（将 delete_time 置 NULL，同时重置 submission_status 为 2）
- [x] 1.7 为每种实体类型的 model 层新增 Restore 方法（复用现有 SoftDelete 的反向操作）

## 2. 后端 - API 生成与验证

- [x] 2.1 运行 goctl api 生成代码
- [x] 2.2 实现生成的 logic 代码并确保编译通过

## 3. 前端 - API 层

- [x] 3.1 在 `src/api/masterdata.ts` 新增 `getDeletedItems`、`getDeletedCounts`、`restoreDeletedItem` 三个 API 函数
- [x] 3.2 在 `web/common/types/masterdata.d.ts` 新增 `DeletedItem` 和 `DeletedCounts` 类型定义

## 4. 前端 - 删除数据恢复页面

- [x] 4.1 创建 `src/views/deleted-recovery/Index.vue` 页面组件，实现统计卡片区域（复用审核中心 stat-card 样式）
- [x] 4.2 实现已删除记录列表表格，展示 entity_type、名称、编码、删除时间等字段
- [x] 4.3 实现恢复按钮逻辑（确认对话框 + 调用恢复接口 + 刷新列表和统计）
- [x] 4.4 实现点击统计卡片筛选/取消筛选功能
- [x] 4.5 实现分页功能

## 5. 前端 - 路由配置

- [x] 5.1 在 `src/router/index.ts` 新增 `/masterdata/deleted-recovery` 路由，配置在审核中心路由之后
