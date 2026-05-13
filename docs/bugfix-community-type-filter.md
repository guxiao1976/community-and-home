# Bug修复：主数据查询页面 community_type 过滤问题

## 问题描述

在主数据-数据查询页面中，选择**山东省东营市**（city_id=283993）并指定**小区类型=住宅小区**（community_type=1）后点击搜索，返回空结果。但数据库中实际有900多条符合条件的记录：

```sql
SELECT * FROM `md_residential_area` WHERE city_id=283993 AND community_type=1
-- 返回 900+ 条记录
```

## 根本原因

### 1. 内存过滤导致分页失效

在 `getResidentialAreasLogic.go` 和 `queryResidentialAreasLogic.go` 中，`community_type` 过滤是在**内存中**进行的，而不是在数据库查询层：

```go
// 错误的实现（第108-121行）
if communityType != nil {
    filtered := make([]*model.MdResidentialArea, 0)
    for _, a := range areas {
        if int32(a.CommunityType) == *communityType {
            filtered = append(filtered, a)
        }
    }
    areas = filtered
    // ...
}
```

**问题流程：**
1. 用户选择 city_id=283993, community_type=1, page=1, page_size=20
2. 后端将 city_id 转换为所有下级 county_ids（东营市的所有区县）
3. 调用 `FindByCountyIds` 查询，**没有 community_type 条件**，返回前20条记录
4. 这20条记录可能都是 `community_type=2`（村庄）或 `community_type=3`（混合型）
5. 内存过滤后变成空数组，返回给前端

### 2. Count 统计不准确

`CountByCommunityType` 方法只统计全局的 community_type 数量，**没有考虑地域过滤条件**（city_id/county_id等），导致 total 数量也不正确。

## 修复方案

### 核心思路

将 `community_type` 过滤条件**下推到数据库查询层**，而不是在内存中过滤。

### 修改内容

#### 1. 更新 Model 接口和实现

**文件：** `services/masterdata/model/mdResidentialAreaModel.go`

- 在所有 `Find*` 方法中增加 `communityType *int32` 参数
- 在 `Count` 方法中增加 `communityType *int32` 参数
- 在 SQL 查询中添加 `community_type = ?` 条件

**修改的方法：**
- `FindByCountyId`
- `FindByCountyIds`
- `FindByStreetId`
- `FindByCommunityDivId`
- `SearchByName`
- `Count`

**示例：**
```go
func (m *customMdResidentialAreaModel) FindByCountyIds(ctx context.Context, countyIds []int64, submissionStatus *int32, communityType *int32, page, pageSize int64, excludeStatuses ...int32) ([]*MdResidentialArea, error) {
    // ...
    appendSubmissionFilter(&conditions, &args, submissionStatus, excludeStatuses...)
    if communityType != nil {
        conditions = append(conditions, "community_type = ?")
        args = append(args, *communityType)
    }
    // ...
}
```

#### 2. 更新 Logic 层调用

**文件：**
- `services/masterdata/api/internal/logic/residentialarea/getResidentialAreasLogic.go`
- `services/masterdata/api/internal/logic/dataquery/queryResidentialAreasLogic.go`
- `services/masterdata/api/internal/logic/division/deleteDivisionLogic.go`
- `services/masterdata/api/internal/logic/division/getDivisionChildCountLogic.go`
- `services/masterdata/rpc/internal/logic/getresidentialareasbydivisionlogic.go`

**主要改动：**
1. 移除内存过滤逻辑（删除第108-121行的代码）
2. 在所有 `Find*` 和 `Count` 调用中传入 `communityType` 参数

**修改前：**
```go
areas, err = l.svcCtx.MdResidentialAreaModel.FindByCountyIds(l.ctx, countyIds, submissionStatus, page, pageSize, excludeArg...)
total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, nil, nil, nil, submissionStatus, countyIds, nil, excludeArg...)

// 内存过滤
if communityType != nil {
    filtered := make([]*model.MdResidentialArea, 0)
    for _, a := range areas {
        if int32(a.CommunityType) == *communityType {
            filtered = append(filtered, a)
        }
    }
    areas = filtered
    total, err = l.svcCtx.MdResidentialAreaModel.CountByCommunityType(l.ctx, *communityType, excludeArg...)
}
```

**修改后：**
```go
areas, err = l.svcCtx.MdResidentialAreaModel.FindByCountyIds(l.ctx, countyIds, submissionStatus, communityType, page, pageSize, excludeArg...)
total, err = l.svcCtx.MdResidentialAreaModel.Count(l.ctx, nil, nil, nil, submissionStatus, countyIds, nil, communityType, excludeArg...)
```

## 验证方法

### 1. 编译检查
```bash
cd services/masterdata/api
go build
```

### 2. 功能测试

**测试用例：**
```bash
# 查询东营市的住宅小区
curl -X GET "http://localhost:8888/api/masterdata/query/residential-areas?city_id=283993&community_type=1&page=1&page_size=20" \
  -H "Authorization: Bearer <token>"
```

**预期结果：**
- 返回符合条件的住宅小区列表（最多20条）
- total 字段显示正确的总数（应该是900+）
- 分页功能正常工作

### 3. SQL 验证

修复后，后端执行的 SQL 应该类似：
```sql
SELECT * FROM md_residential_area 
WHERE county_id IN (283994, 283995, ...) 
  AND community_type = 1 
  AND delete_time IS NULL 
  AND submission_status != 4
ORDER BY id DESC 
LIMIT 20 OFFSET 0
```

## 影响范围

### 修改的文件
1. `services/masterdata/model/mdResidentialAreaModel.go` - Model层接口和实现
2. `services/masterdata/api/internal/logic/residentialarea/getResidentialAreasLogic.go` - 小区管理查询
3. `services/masterdata/api/internal/logic/dataquery/queryResidentialAreasLogic.go` - 主数据查询页面
4. `services/masterdata/api/internal/logic/division/deleteDivisionLogic.go` - 行政区划删除检查
5. `services/masterdata/api/internal/logic/division/getDivisionChildCountLogic.go` - 行政区划子节点统计
6. `services/masterdata/rpc/internal/logic/getresidentialareasbydivisionlogic.go` - RPC服务

### 受影响的功能
- ✅ 主数据查询页面（修复了bug）
- ✅ 小区管理列表页面（保持兼容）
- ✅ 行政区划管理（保持兼容）
- ✅ RPC服务调用（保持兼容）

## 总结

此次修复通过将 `community_type` 过滤条件从内存层下推到数据库查询层，解决了分页失效的问题。修改后：

1. **性能提升**：数据库直接过滤，减少内存操作和数据传输
2. **逻辑正确**：分页和过滤条件正确组合，不会出现"有数据但查不到"的问题
3. **代码简化**：移除了内存过滤逻辑，代码更清晰
4. **向后兼容**：所有调用点都已更新，不影响现有功能

## 日期
2026-05-12
