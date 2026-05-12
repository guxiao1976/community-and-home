# 高德API小区数据全量获取方案探索

## 当前实现分析

### 现有代码逻辑 (sync_engine.go)

```go
// 当前实现：按区县搜索
func (e *SyncEngine) searchResidentialAreas(countyCode string, page int) (*amapTextSearchResp, error) {
    reqUrl := fmt.Sprintf(
        "https://restapi.amap.com/v3/place/text?keywords=&types=120300&city=%s&citylimit=true&offset=25&page=%d&key=%s",
        countyCode, page, e.amapKey,
    )
    // ...
}

// 翻页逻辑
totalPages := totalCount / 25
if totalCount%25 > 0 {
    totalPages++
}
if totalPages > 100 {  // ⚠️ 硬编码限制为100页
    totalPages = 100
}
```

### 当前问题

1. **硬编码限制**: `totalPages` 最大为 100 页，但实际高德API只支持到 40 页
2. **单一搜索策略**: 只使用 `types=120300`（住宅区类型），按区县搜索
3. **数据丢失**: 对于小区数量超过 1000 个的区县，会丢失数据
4. **无去重机制**: 不同搜索策略可能返回重复数据

---

## 方案一：行政层级下钻法 ⭐ 推荐

### 原理

将区县（level=3）进一步细分为街道/乡镇（level=4），以街道为单位搜索小区。

### 实施步骤

#### 1. 数据库准备

检查 `md_administrative_division` 表是否已有 level=4 的街道数据：

```sql
SELECT COUNT(*) FROM md_administrative_division WHERE level = 4;
```

如果没有，需要先导入街道数据。

#### 2. 修改搜索逻辑

```go
// 新增：按街道搜索
func (e *SyncEngine) searchByTownship(countyCode, townshipName string, page int) (*amapTextSearchResp, error) {
    // 高德API不直接支持 township 参数，需要使用 district 参数组合
    // 方案A: 使用街道的 adcode（如果有）
    // 方案B: 使用 keywords 组合："小区 街道名"
    
    reqUrl := fmt.Sprintf(
        "https://restapi.amap.com/v3/place/text?keywords=%s&types=120300&city=%s&citylimit=true&offset=25&page=%d&key=%s",
        townshipName, countyCode, page, e.amapKey,
    )
    // ...
}

// 修改同步流程
func (e *SyncEngine) runSyncSingleCounty(ctx context.Context, p *SyncProgress, countyId int64) {
    // 1. 获取该区县下的所有街道
    townships, err := e.divModel.FindChildrenWithFilter(ctx, countyId, ptrInt64(4), nil)
    if err != nil || len(townships) == 0 {
        // 如果没有街道数据，回退到原有逻辑
        e.runSyncByCountyDirect(ctx, p, countyId)
        return
    }
    
    // 2. 遍历每个街道进行搜索
    for _, township := range townships {
        e.runSyncByTownship(ctx, p, countyId, township)
        time.Sleep(time.Duration(3+rand.Intn(3)) * time.Second)
    }
}
```

### 优点

- ✅ 逻辑简单，改动量小
- ✅ 利用现有行政区划数据
- ✅ 大部分区县能解决问题（街道级别小区数量通常 < 1000）

### 缺点

- ❌ 依赖街道数据的完整性
- ❌ 超大街道（如望京街道）仍可能超过 1000 条
- ❌ 高德API不直接支持街道参数，需要通过 keywords 组合，可能不精确

### 可行性评估

**可行性: ⭐⭐⭐⭐ (4/5)**

- 需要先确认是否有 level=4 街道数据
- 如果没有，需要先调用高德行政区划API导入街道数据
- 适合作为第一阶段优化方案

---

## 方案二：地理网格切分法 ⭐⭐⭐ 最彻底

### 原理

将区县的地理边界切分成小网格（如 0.01° × 0.01°，约 1km × 1km），对每个网格进行搜索。

### 实施步骤

#### 1. 获取区县边界

调用高德行政区划API获取边界多边形：

```go
// 新增：获取区县边界
func (e *SyncEngine) getCountyBoundary(countyCode string) ([][]float64, error) {
    reqUrl := fmt.Sprintf(
        "https://restapi.amap.com/v3/config/district?keywords=%s&subdistrict=0&extensions=all&key=%s",
        countyCode, e.amapKey,
    )
    
    // 返回格式: [[lng1,lat1], [lng2,lat2], ...]
    // 解析 polyline 字段
}
```

#### 2. 网格切分算法

```go
type Grid struct {
    MinLng float64
    MinLat float64
    MaxLng float64
    MaxLat float64
}

// 将边界切分成网格
func (e *SyncEngine) splitIntoGrids(boundary [][]float64, gridSize float64) []Grid {
    // 1. 计算边界的最小外接矩形
    minLng, minLat := boundary[0][0], boundary[0][1]
    maxLng, maxLat := minLng, minLat
    
    for _, point := range boundary {
        if point[0] < minLng { minLng = point[0] }
        if point[0] > maxLng { maxLng = point[0] }
        if point[1] < minLat { minLat = point[1] }
        if point[1] > maxLat { maxLat = point[1] }
    }
    
    // 2. 按 gridSize 切分
    var grids []Grid
    for lng := minLng; lng < maxLng; lng += gridSize {
        for lat := minLat; lat < maxLat; lat += gridSize {
            grids = append(grids, Grid{
                MinLng: lng,
                MinLat: lat,
                MaxLng: lng + gridSize,
                MaxLat: lat + gridSize,
            })
        }
    }
    
    return grids
}
```

#### 3. 按网格搜索

```go
// 按网格搜索
func (e *SyncEngine) searchByGrid(grid Grid, page int) (*amapTextSearchResp, error) {
    // 使用 polygon 参数
    polygon := fmt.Sprintf("%f,%f;%f,%f", 
        grid.MinLng, grid.MinLat, 
        grid.MaxLng, grid.MaxLat,
    )
    
    reqUrl := fmt.Sprintf(
        "https://restapi.amap.com/v3/place/polygon?keywords=&types=120300&polygon=%s&offset=25&page=%d&key=%s",
        polygon, page, e.amapKey,
    )
    // ...
}

// 修改同步流程
func (e *SyncEngine) runSyncByGrids(ctx context.Context, p *SyncProgress, countyId int64) {
    county, _ := e.divModel.FindOne(ctx, countyId)
    
    // 1. 获取边界
    boundary, err := e.getCountyBoundary(county.Code)
    if err != nil {
        logx.Errorf("get boundary failed: %v", err)
        return
    }
    
    // 2. 切分网格（0.01° ≈ 1.1km）
    grids := e.splitIntoGrids(boundary, 0.01)
    
    // 3. 遍历每个网格
    seen := make(map[string]bool) // 去重：name+location
    
    for _, grid := range grids {
        for page := 1; page <= 40; page++ {
            resp, err := e.searchByGrid(grid, page)
            if err != nil {
                break
            }
            
            if len(resp.POIs) == 0 {
                break
            }
            
            for _, poi := range resp.POIs {
                key := poi.Name + "|" + poi.Location
                if seen[key] {
                    continue // 跨网格重复
                }
                seen[key] = true
                
                // 插入数据库
                e.insertResidentialArea(ctx, poi, countyId)
            }
            
            if len(resp.POIs) < 25 {
                break // 最后一页
            }
        }
        
        time.Sleep(time.Duration(2+rand.Intn(2)) * time.Second)
    }
}
```

### 优点

- ✅ **100% 全量覆盖**，不会漏数据
- ✅ 每个网格小区数量必然 < 1000
- ✅ 不依赖行政区划数据

### 缺点

- ❌ 实现复杂度高（需要几何计算）
- ❌ API 调用次数大幅增加（网格数量 × 页数）
- ❌ 网格边界会有重复数据，需要去重
- ❌ 高德API的 `polygon` 参数可能有格式要求，需要测试

### 可行性评估

**可行性: ⭐⭐⭐ (3/5)**

- 需要测试高德 `polygon` 参数的支持情况
- 需要实现几何切分算法
- 需要实现去重逻辑（基于 name + location）
- API 调用成本高，适合作为兜底方案

---

## 方案三：关键词细分法 ⭐⭐⭐⭐⭐ 最简单

### 原理

利用小区命名规律，使用不同的关键词后缀分别搜索，然后合并去重。

### 实施步骤

#### 1. 定义关键词列表

```go
var residentialKeywords = []string{
    "小区",
    "花园",
    "苑",
    "公寓",
    "村",
    "家园",
    "居",
    "园",
    "庭",
    "府",
    "城",
    "里",
    "坊",
    "阁",
    "轩",
    "墅",
    "邸",
    "郡",
    "湾",
    "岸",
}
```

#### 2. 修改搜索逻辑

```go
// 按关键词搜索
func (e *SyncEngine) searchByKeyword(countyCode, keyword string, page int) (*amapTextSearchResp, error) {
    reqUrl := fmt.Sprintf(
        "https://restapi.amap.com/v3/place/text?keywords=%s&types=120300&city=%s&citylimit=true&offset=25&page=%d&key=%s",
        keyword, countyCode, page, e.amapKey,
    )
    // ...
}

// 修改同步流程
func (e *SyncEngine) runSyncByKeywords(ctx context.Context, p *SyncProgress, countyId int64) {
    county, _ := e.divModel.FindOne(ctx, countyId)
    
    seen := make(map[string]bool) // 去重：name
    
    for _, keyword := range residentialKeywords {
        logx.Infof("[AMap Sync] 搜索关键词: %s", keyword)
        
        // 1. 获取第一页，确定总数
        firstResp, err := e.searchByKeyword(county.Code, keyword, 1)
        if err != nil {
            continue
        }
        
        totalCount, _ := strconv.Atoi(firstResp.Count)
        totalPages := (totalCount + 24) / 25
        if totalPages > 40 {
            totalPages = 40 // 高德限制
        }
        
        // 2. 遍历所有页
        for page := 1; page <= totalPages; page++ {
            var resp *amapTextSearchResp
            if page == 1 {
                resp = firstResp
            } else {
                resp, err = e.searchByKeyword(county.Code, keyword, page)
                if err != nil {
                    break
                }
            }
            
            for _, poi := range resp.POIs {
                if seen[poi.Name] {
                    p.mu.Lock()
                    p.TotalSkipped++
                    p.mu.Unlock()
                    continue
                }
                seen[poi.Name] = true
                
                // 检查数据库是否已存在
                existing, _ := e.areaModel.FindByNameAndCountyId(ctx, poi.Name, countyId)
                if existing != nil {
                    p.mu.Lock()
                    p.TotalSkipped++
                    p.mu.Unlock()
                    continue
                }
                
                // 插入数据库
                e.insertResidentialArea(ctx, poi, countyId)
                
                p.mu.Lock()
                p.TotalSynced++
                p.mu.Unlock()
            }
            
            time.Sleep(time.Duration(3+rand.Intn(3)) * time.Second)
        }
        
        // 关键词之间延迟
        time.Sleep(time.Duration(5+rand.Intn(5)) * time.Second)
    }
}
```

### 优点

- ✅ **实现极简**，改动量最小
- ✅ 不需要额外的地理数据或几何计算
- ✅ 能覆盖大部分常见命名模式
- ✅ 去重逻辑简单（基于小区名称）

### 缺点

- ❌ 无法保证 100% 覆盖（可能有特殊命名）
- ❌ 关键词选择需要经验和测试
- ❌ 不同关键词可能返回相同小区（需要去重）
- ❌ API 调用次数增加（关键词数量 × 页数）

### 可行性评估

**可行性: ⭐⭐⭐⭐⭐ (5/5)**

- 实现简单，立即可用
- 适合作为第一阶段快速优化
- 可以与方案一组合使用

---

## 方案四：混合策略（推荐）

### 策略组合

```
1. 优先使用"关键词细分法"（方案三）
   - 快速实现，覆盖大部分场景
   
2. 对于数据量特别大的区县，启用"街道下钻法"（方案一）
   - 判断条件：某个关键词返回 count > 800
   
3. 对于超大街道，启用"网格切分法"（方案二）
   - 判断条件：街道级别搜索仍然 count > 800
```

### 实施代码框架

```go
func (e *SyncEngine) runSyncSingleCountyAdaptive(ctx context.Context, p *SyncProgress, countyId int64) {
    county, _ := e.divModel.FindOne(ctx, countyId)
    
    // 策略1: 先尝试关键词细分
    totalFound := e.runSyncByKeywords(ctx, p, countyId)
    
    // 策略2: 如果发现某个关键词数据量过大，切换到街道下钻
    if totalFound > 5000 { // 经验阈值
        logx.Infof("[AMap Sync] 区县 %s 数据量大，启用街道下钻", county.Name)
        
        townships, err := e.divModel.FindChildrenWithFilter(ctx, countyId, ptrInt64(4), nil)
        if err == nil && len(townships) > 0 {
            for _, township := range townships {
                e.runSyncByTownshipWithKeywords(ctx, p, countyId, township)
            }
            return
        }
    }
    
    // 策略3: 如果街道数据不可用，使用网格切分（兜底）
    if totalFound > 10000 {
        logx.Infof("[AMap Sync] 区县 %s 数据量极大，启用网格切分", county.Name)
        e.runSyncByGrids(ctx, p, countyId)
    }
}
```

---

## 实施建议

### 阶段一：快速优化（1-2天）

**实施方案三：关键词细分法**

1. 修改 `searchResidentialAreas` 函数，增加 `keywords` 参数
2. 修改 `runSyncSingleCounty`，遍历关键词列表
3. 增加内存去重逻辑（基于小区名称）
4. 测试 2-3 个区县，验证效果

**预期效果**：
- 数据覆盖率从 40% 提升到 80-90%
- 代码改动量 < 100 行

### 阶段二：完善覆盖（3-5天）

**实施方案一：街道下钻法**

1. 确认数据库中是否有 level=4 街道数据
2. 如果没有，先调用高德API导入街道数据
3. 修改同步逻辑，支持街道级别搜索
4. 与关键词细分法组合使用

**预期效果**：
- 数据覆盖率提升到 95%+
- 解决大部分区县的数据丢失问题

### 阶段三：兜底方案（5-7天）

**实施方案二：网格切分法**

1. 实现边界获取和网格切分算法
2. 测试高德 `polygon` 参数
3. 实现基于坐标的去重逻辑
4. 仅对超大区县启用

**预期效果**：
- 数据覆盖率达到 99%+
- 完全解决数据丢失问题

---

## 技术细节

### 去重策略

```go
// 方案三（关键词）：基于名称去重
type DeduplicatorByName struct {
    seen map[string]bool
}

func (d *DeduplicatorByName) IsDuplicate(poi amapPOI) bool {
    if d.seen[poi.Name] {
        return true
    }
    d.seen[poi.Name] = true
    return false
}

// 方案二（网格）：基于名称+坐标去重
type DeduplicatorByLocation struct {
    seen map[string]bool
}

func (d *DeduplicatorByLocation) IsDuplicate(poi amapPOI) bool {
    key := poi.Name + "|" + poi.Location
    if d.seen[key] {
        return true
    }
    d.seen[key] = true
    return false
}
```

### API 限流策略

```go
// 当前实现已有随机延迟，建议优化：
func (e *SyncEngine) rateLimitedRequest(reqFunc func() (*amapTextSearchResp, error)) (*amapTextSearchResp, error) {
    // 1. 请求前延迟
    delay := time.Duration(2+rand.Intn(3)) * time.Second
    time.Sleep(delay)
    
    // 2. 执行请求
    resp, err := reqFunc()
    
    // 3. 如果遇到限流错误，指数退避
    if err != nil && strings.Contains(err.Error(), "DAILY_QUERY_OVER_LIMIT") {
        logx.Errorf("[AMap Sync] 达到每日限额，暂停同步")
        return nil, err
    }
    
    return resp, err
}
```

### 进度追踪优化

```go
// 增加更详细的进度信息
type SyncProgress struct {
    // ... 现有字段
    
    // 新增字段
    CurrentStrategy  string `json:"current_strategy"`  // "keyword" / "township" / "grid"
    CurrentKeyword   string `json:"current_keyword"`   // 当前关键词
    TotalKeywords    int32  `json:"total_keywords"`    // 总关键词数
    CurrentKeywordIdx int32 `json:"current_keyword_idx"` // 当前关键词索引
}
```

---

## 成本估算

### API 调用次数对比

假设一个区县有 2000 个小区：

| 方案 | 调用次数 | 说明 |
|------|---------|------|
| 当前方案 | 40 次 | 只能拿到 1000 个，丢失 1000 个 |
| 方案一（街道下钻） | 40 × 街道数 | 假设 10 个街道 = 400 次 |
| 方案二（网格切分） | 40 × 网格数 | 假设 100 个网格 = 4000 次 |
| 方案三（关键词细分） | 40 × 关键词数 | 假设 20 个关键词 = 800 次 |

### 时间成本

假设每次请求延迟 3 秒：

| 方案 | 单区县耗时 | 全国 3000 区县耗时 |
|------|-----------|------------------|
| 当前方案 | 2 分钟 | 100 小时 |
| 方案一 | 20 分钟 | 1000 小时 |
| 方案二 | 3.3 小时 | 10000 小时 |
| 方案三 | 40 分钟 | 2000 小时 |

---

## 总结

### 推荐实施路径

**第一优先级：方案三（关键词细分法）**
- 实现简单，立即可用
- 覆盖率 80-90%
- 成本可控

**第二优先级：方案一（街道下钻法）**
- 与方案三组合使用
- 覆盖率提升到 95%+
- 需要街道数据支持

**兜底方案：方案二（网格切分法）**
- 仅对超大区县使用
- 保证 100% 覆盖
- 成本较高，按需启用

### 下一步行动

1. ✅ 立即实施方案三（关键词细分法）
2. 📋 测试 2-3 个典型区县（大、中、小）
3. 📋 评估覆盖率提升效果
4. 📋 根据效果决定是否实施方案一或方案二
