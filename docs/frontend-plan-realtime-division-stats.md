# 前端改造方案：小区数据统计 - 双 Tab 结构

## 概述

将当前「小区数据统计」页面改造为双 Tab 结构，Tab1 为「昨日数据」（现有功能），Tab2 为「实时数据」（新增后端 API 已就绪）。

## 后端 API 说明

### 现有接口（昨日数据 - 保持不变）
```
GET /api/masterdata/statistics/division-counts
参数: parent_id (可选, number)
响应: { code: 0, message: "success", data: { list: DivisionCountItem[] } }
```

### 新增接口（实时数据 - 已实现）
```
GET /api/masterdata/statistics/division-counts/realtime
参数: parent_id (可选, number)
响应: { code: 0, message: "success", data: { list: DivisionCountItem[] } }
```

两个接口的请求参数和响应格式完全一致，`DivisionCountItem` 结构：
```typescript
interface DivisionCountItem {
  id: number
  name: string
  level: number        // 1=省, 2=市, 3=区县, 4=街道
  community_count: number
  village_count: number
  total_count: number
}
```

## 改造内容

### 1. 新增 API 函数
**文件**: `web/pc/src/api/masterdata.ts`

```typescript
// 在现有 getDivisionCounts 下方新增
export const getDivisionCountsRealtime = (params?: { parent_id?: number }) => {
  return request.get<{ list: DivisionCountItem[] }>('/api/masterdata/statistics/division-counts/realtime', { params })
}
```

### 2. 抽取统计面板为独立组件
**新文件**: `web/pc/src/views/statistics/division-counts/StatsPanel.vue`

将当前 `Index.vue` 中的 4 面板下钻逻辑抽取为可复用组件：

**Props**:
```typescript
interface Props {
  dataLoader: (parentId?: number) => Promise<DivisionCountItem[]>
}
```

**Emits**: 无（面板内部自管理下钻状态）

**抽取内容**:
- 4 个 ElTable 面板（省、市、区县、街道）
- 面包屑导航逻辑
- 下钻交互逻辑（点击行 → 加载子级 → 重置后续面板）
- 空状态提示
- 选中行高亮

### 3. 改造 Index.vue
**文件**: `web/pc/src/views/statistics/division-counts/Index.vue`

```vue
<template>
  <div>
    <el-tabs v-model="activeTab">
      <el-tab-pane label="昨日数据" name="yesterday">
        <StatsPanel :data-loader="loadYesterdayData" />
      </el-tab-pane>
      <el-tab-pane label="实时数据" name="realtime">
        <StatsPanel :data-loader="loadRealtimeData" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { getDivisionCounts, getDivisionCountsRealtime } from '@/api/masterdata'
import StatsPanel from './StatsPanel.vue'

const activeTab = ref('yesterday')

function loadYesterdayData(parentId?: number) {
  return getDivisionCounts(parentId ? { parent_id: parentId } : undefined)
    .then(res => res?.list || [])
}

function loadRealtimeData(parentId?: number) {
  return getDivisionCountsRealtime(parentId ? { parent_id: parentId } : undefined)
    .then(res => res?.list || [])
}
</script>
```

### 4. 关键设计要点

- **组件复用**: `StatsPanel` 通过 `dataLoader` prop 注入不同的数据加载函数，Tab1 和 Tab2 共用同一套展示组件
- **独立状态**: 切换 Tab 时，两个面板各自维护独立的下钻状态（面包屑、选中等）
- **列头排序**: 两个 Tab 的排序逻辑保持一致（前端本地排序）
- **无分页**: 统计数据量有限（省~34条、市~330条），无需分页

## 涉及文件清单

| 文件 | 操作 |
|------|------|
| `web/pc/src/api/masterdata.ts` | 新增 `getDivisionCountsRealtime` 函数 |
| `web/pc/src/views/statistics/division-counts/StatsPanel.vue` | 新建，抽取面板组件 |
| `web/pc/src/views/statistics/division-counts/Index.vue` | 改造，增加 ElTabs 包裹 |

## 不涉及的文件

- 路由配置 `router/index.ts` — 路由路径不变
- 侧边栏 `AppSidebar.vue` — 菜单入口不变
- API 接口定义 `masterdata.api` — 后端已就绪
