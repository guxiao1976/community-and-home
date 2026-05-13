# 小区数据获取 - 街道级别同步实现文档

## 需求概述

完善主数据管理的小区数据获取功能，将原有的按区县搜索改为按街道搜索，形成省-市-区县-街道四级选择。

## 实现内容

### 1. 前端改动

#### 1.1 增加街道级联选择器

**文件**: `web/pc/src/views/amap-sync/Index.vue`

**改动内容**:
- 增加第四级"街道"选择器
- 增加 `streetOptions` 和 `filters.streetId`
- 增加 `handleCountyChange` 方法加载街道列表
- 更新 `effectiveDivisionId` 和 `levelLabel` 计算逻辑

**UI 效果**:
```
省份: [选择框] → 城市: [选择框] → 区县: [选择框] → 街道: [选择框]
```

#### 1.2 更新进度显示

**改动内容**:
- 增加街道级别的进度条显示
- 显示当前处理的街道名称和进度
- 优化进度信息文案，显示"第X/Y个街道「街道名」"

**显示效果**:
```
区县进度: 1 / 3 (朝阳区)
街道进度: 5 / 15 (望京街道)
页面进度: 8 / 40
正在处理第1/3个区县「朝阳区」，第5/15个街道「望京街道」，第8/40页（共发现200个小区）
```

#### 1.3 更新类型定义

**文件**: `web/pc/src/api/masterdata.ts`

**改动内容**:
```typescript
export interface SyncProgress {
  // ... 原有字段
  total_streets: number
  current_street: number
  current_street_name?: string
}
```

### 2. 后端改动

#### 2.1 更新同步引擎逻辑

**文件**: `services/masterdata/api/internal/sync/sync_engine.go`

**核心改动**:

##### 2.1.1 更新进度结构
```go
type SyncProgress struct {
    // ... 原有字段
    TotalStreets      int32      `json:"total_streets"`
    CurrentStreetIdx  int32      `json:"current_street"`
    CurrentStreetName string     `json:"current_street_name"`
}
```

##### 2.1.2 修改同步流程

**原逻辑**: 
```
区县 → 直接搜索小区（最多1000条）
```

**新逻辑**:
```
区县 → 获取街道列表 → 遍历每个街道 → 按街道搜索小区
```

**关键代码**:
```go
func (e *SyncEngine) runSyncSingleCounty(ctx context.Context, p *SyncProgress, countyId int64) {
    // 1. 获取该区县下的所有街道（level=4）
    streets, err := e.divModel.FindChildrenWithFilter(ctx, countyId, ptrInt64(4), nil)
    if err != nil || len(streets) == 0 {
        logx.Infof("[AMap Sync] 区县没有街道数据，跳过")
        return
    }
    
    // 2. 遍历每个街道进行同步
    for i, street := range streets {
        e.runSyncSingleStreet(ctx, p, countyId, street.Id, county.Code)
        
        // 街道之间间隔 10-60 秒
        if i < len(streets)-1 {
            delay := time.Duration(10+rand.Intn(51)) * time.Second
            time.Sleep(delay)
        }
    }
}
```

##### 2.1.3 插入小区时设置 street_id

**改动**:
```go
area := &model.MdResidentialArea{
    CountyId:         sql.NullInt64{Int64: countyId, Valid: true},
    CityId:           sql.NullInt64{Int64: cityId, Valid: true},
    StreetId:         sql.NullInt64{Int64: streetId, Valid: true}, // 新增
    // ... 其他字段
}
```

##### 2.1.4 调整延迟时间

- **街道之间**: 10-60 秒（原为区县之间 5-60 秒）
- **页面之间**: 5-10 秒（保持不变）

#### 2.2 更新 API 定义

**文件**: `services/masterdata/api/masterdata.api`

**改动内容**:
```go
GetSyncProgressResp {
    // ... 原有字段
    TotalStreets      int32  `json:"total_streets"`
    CurrentStreet     int32  `json:"current_street"`
    CurrentStreetName string `json:"current_street_name,optional"`
}
```

## 数据流程

### 同步流程图

```
用户选择区县/街道
    ↓
触发同步 (POST /api/masterdata/amap-sync/sync)
    ↓
后端解析 division_id
    ↓
如果是区县 → 获取该区县下所有街道
如果是街道 → 直接使用该街道
    ↓
遍历每个街道:
    ↓
    调用高德API搜索小区 (types=120300, city=区县code)
    ↓
    翻页获取数据 (最多40页，每页25条)
    ↓
    插入数据库 (设置 county_id, city_id, street_id)
    ↓
    等待 5-10 秒
    ↓
等待 10-60 秒后处理下一个街道
    ↓
完成同步
```

### 数据库字段映射

| 字段 | 来源 | 说明 |
|------|------|------|
| county_id | 参数传入 | 区县ID (level=3) |
| city_id | county.parent_id | 城市ID (level=2) |
| street_id | 当前遍历的街道 | 街道ID (level=4) |
| code | 自动生成 | 区县code + 4位序号 |
| name | 高德API | 小区名称 |
| address | 高德API | 详细地址 |
| longitude | 高德API | 经度 |
| latitude | 高德API | 纬度 |
| data_source | 固定值 1 | 1=高德API |
| community_type | 固定值 1 | 1=住宅区 |
| submission_status | 固定值 2 | 2=已审核通过 |

## 前置条件

### 数据库要求

1. **必须有街道数据** (level=4)
   
   检查命令:
   ```sql
   SELECT COUNT(*) FROM md_administrative_division WHERE level = 4;
   ```

2. **表结构包含 street_id 字段**
   
   检查命令:
   ```sql
   SHOW COLUMNS FROM md_residential_area LIKE 'street_id';
   ```

### 如果没有街道数据

需要先导入街道数据，有两种方式：

#### 方式一：手动导入
```sql
-- 示例：为北京市朝阳区添加街道
INSERT INTO md_administrative_division (parent_id, level, name, code, path, sort_order, status, submission_status, created_by, created_time, updated_time)
VALUES 
(110105, 4, '望京街道', '110105001', '/1/11/110105/110105001/', 1, 1, 2, 0, NOW(), NOW()),
(110105, 4, '三里屯街道', '110105002', '/1/11/110105/110105002/', 2, 1, 2, 0, NOW(), NOW());
```

#### 方式二：调用高德API导入
```bash
# 调用高德行政区划API获取街道数据
curl "https://restapi.amap.com/v3/config/district?keywords=110105&subdistrict=1&extensions=base&key=YOUR_KEY"
```

## 测试步骤

### 1. 准备测试数据

```sql
-- 确认有街道数据
SELECT d.id, d.name, d.code, d.level, p.name as parent_name
FROM md_administrative_division d
LEFT JOIN md_administrative_division p ON d.parent_id = p.id
WHERE d.level = 4
LIMIT 10;
```

### 2. 前端测试

1. 访问 `http://localhost:3000/amap-sync`
2. 依次选择：省份 → 城市 → 区县 → 街道
3. 点击"开始同步"
4. 观察进度显示：
   - 应显示街道进度条
   - 应显示当前街道名称
   - 应显示页面进度

### 3. 后端日志验证

```bash
# 查看同步日志
tail -f services/masterdata/api/logs/masterdata-api.log | grep "AMap Sync"
```

**预期日志**:
```
[AMap Sync] 1/1 开始同步: 朝阳区 (code=110105)
[AMap Sync] 区县 朝阳区 共有 15 个街道
[AMap Sync] 1/15 开始同步街道: 望京街道 (code=110105001)
[AMap Sync] 等待 15s 后处理下一个街道...
[AMap Sync] 2/15 开始同步街道: 三里屯街道 (code=110105002)
```

### 4. 数据库验证

```sql
-- 验证 street_id 是否正确设置
SELECT 
    ra.id,
    ra.name,
    ra.code,
    c.name as county_name,
    s.name as street_name,
    ra.street_id
FROM md_residential_area ra
LEFT JOIN md_administrative_division c ON ra.county_id = c.id
LEFT JOIN md_administrative_division s ON ra.street_id = s.id
WHERE ra.data_source = 1
ORDER BY ra.created_time DESC
LIMIT 20;
```

**预期结果**:
- `street_id` 不为 NULL
- `street_name` 显示正确的街道名称

## 性能影响

### API 调用次数变化

假设一个区县有 10 个街道，每个街道平均 500 个小区：

**原方案**:
- 调用次数: 40 页 × 1 个区县 = 40 次
- 获取数量: 1000 个（丢失 4000 个）

**新方案**:
- 调用次数: 40 页 × 10 个街道 = 400 次
- 获取数量: 5000 个（完整覆盖）

### 时间成本

**原方案**:
- 单区县耗时: 40 页 × 7 秒 = 4.7 分钟

**新方案**:
- 单区县耗时: (40 页 × 7.5 秒 + 35 秒) × 10 街道 = 56 分钟

### 优化建议

如果时间成本过高，可以考虑：

1. **并发处理街道**（需要注意高德API限流）
2. **只同步指定街道**（用户选择具体街道）
3. **增量同步**（只同步新增/变更的小区）

## 注意事项

### 1. 高德API限制

- **单次请求**: 最多返回 25 条
- **最大页数**: 40 页
- **每日限额**: 根据账号等级不同

### 2. 数据去重

当前去重逻辑：
```go
existing, err := e.areaModel.FindByNameAndCountyId(ctx, poi.Name, countyId)
if err == nil && existing != nil {
    // 跳过重复数据
}
```

**注意**: 同名小区在不同街道会被去重，可能需要优化为：
```go
existing, err := e.areaModel.FindByNameAndStreetId(ctx, poi.Name, streetId)
```

### 3. 错误处理

- 如果某个街道没有数据，会跳过继续处理下一个
- 如果某个街道搜索失败，会记录错误但不中断整体流程
- 如果区县没有街道数据，会跳过该区县

### 4. 数据一致性

- 插入时设置 `submission_status = 2`（已审核通过），直接可用
- 如果需要审核流程，改为 `submission_status = 0`（待提交）

## 回滚方案

如果新方案有问题，可以快速回滚：

### 1. 前端回滚

```bash
git checkout HEAD~1 web/pc/src/views/amap-sync/Index.vue
git checkout HEAD~1 web/pc/src/api/masterdata.ts
```

### 2. 后端回滚

```bash
git checkout HEAD~1 services/masterdata/api/internal/sync/sync_engine.go
git checkout HEAD~1 services/masterdata/api/masterdata.api
cd services/masterdata/api && goctl api go -api masterdata.api -dir . -style go_zero
```

### 3. 重启服务

```bash
# 重启后端
cd services/masterdata/api && go run masterdata.go

# 重启前端
cd web/pc && npm run dev
```

## 后续优化方向

1. **关键词细分法**（参考 `docs/amap-data-fetch-solutions.md`）
   - 使用多个关键词分别搜索
   - 提高数据覆盖率到 80-90%

2. **网格切分法**（兜底方案）
   - 对超大街道使用地理网格切分
   - 保证 100% 覆盖

3. **增量同步**
   - 记录上次同步时间
   - 只同步新增/变更的小区

4. **并发优化**
   - 多个街道并发处理
   - 注意控制并发数避免触发限流

## 相关文档

- [高德API数据获取方案探索](./amap-data-fetch-solutions.md)
- [敏感词表结构分析](./analysis-sensitive-word-schema-mismatch.md)
