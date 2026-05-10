## ADDED Requirements

### Requirement: Query deleted records by entity type
系统 SHALL 提供 API 接口，允许按实体类型查询已软删除的主数据记录列表。已删除记录定义为 `delete_time IS NOT NULL` 的记录。支持分页查询。

#### Scenario: 查询所有类型的已删除记录
- **WHEN** 调用 `GET /api/masterdata/deleted-items` 不带 entity_type 参数
- **THEN** 返回所有实体类型的已删除记录分页列表，每条记录包含 entity_type、entity_id、entity_name、delete_time 等字段

#### Scenario: 按实体类型筛选已删除记录
- **WHEN** 调用 `GET /api/masterdata/deleted-items?entity_type=residential_area`
- **THEN** 仅返回住宅小区类型的已删除记录

#### Scenario: 获取各类型已删除记录数量统计
- **WHEN** 调用 `GET /api/masterdata/deleted-items/counts`
- **THEN** 返回各实体类型（residential_area、administrative_division、configuration、sensitive_word）的已删除记录数量及总数

### Requirement: Restore deleted record
系统 SHALL 提供恢复接口，将已软删除的主数据记录恢复为正常状态。恢复操作将 `delete_time` 置为 NULL，使记录重新出现在正常查询结果中。

#### Scenario: 成功恢复一条已删除记录
- **WHEN** 调用 `POST /api/masterdata/deleted-items/{entity_type}/{id}/restore` 且该记录存在且 delete_time 不为空
- **THEN** 将该记录的 delete_time 置为 NULL，返回成功响应

#### Scenario: 恢复不存在的记录
- **WHEN** 调用恢复接口但记录 ID 不存在
- **THEN** 返回 404 错误

#### Scenario: 恢复未被删除的记录
- **WHEN** 调用恢复接口但记录的 delete_time 为 NULL
- **THEN** 返回 400 错误，提示该记录未被删除

### Requirement: Deleted data recovery page UI
系统 SHALL 在主数据管理菜单下新增"删除数据恢复"页面。页面上半部分展示统计卡片（与审核中心风格一致），显示各主数据表名称及对应已删除记录数量。点击卡片后，下方展示该类型已删除的数据列表。

#### Scenario: 页面初始加载
- **WHEN** 用户访问"删除数据恢复"页面
- **THEN** 页面上方展示 4 张统计卡片（住宅小区、行政区划、系统配置、敏感词），每张卡片显示对应类型的已删除记录数量

#### Scenario: 点击统计卡片筛选
- **WHEN** 用户点击"住宅小区"统计卡片
- **THEN** 下方列表仅展示住宅小区类型的已删除记录，卡片高亮显示选中状态

#### Scenario: 再次点击取消筛选
- **WHEN** 用户再次点击已选中的统计卡片
- **THEN** 取消筛选，下方列表展示所有类型的已删除记录

#### Scenario: 恢复操作
- **WHEN** 用户在列表中点击某条记录的"恢复"按钮
- **THEN** 弹出确认对话框，用户确认后调用恢复接口，成功后刷新列表和统计数量

#### Scenario: 恢复后记录从列表消失
- **WHEN** 恢复操作成功后
- **THEN** 该记录从已删除列表中移除，对应统计卡片的数量减 1
