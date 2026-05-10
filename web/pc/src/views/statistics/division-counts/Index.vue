<template>
  <div class="division-counts-container">
    <el-card>
      <template #header>
        <span>社区/村数量统计</span>
      </template>

      <el-breadcrumb v-if="breadcrumb.length > 0" separator=">" class="breadcrumb">
        <el-breadcrumb-item
          v-for="(item, index) in breadcrumb"
          :key="item.id"
        >
          <a
            v-if="index < breadcrumb.length - 1"
            href="javascript:void(0)"
            @click="handleBreadcrumbClick(index)"
          >{{ item.name }}</a>
          <span v-else>{{ item.name }}</span>
        </el-breadcrumb-item>
      </el-breadcrumb>

      <div class="drill-down-panels">
        <div class="panel" v-for="(panel, idx) in panels" :key="idx">
          <div class="panel-title">
            {{ panel.title }}
            <span v-if="panel.parentName" class="parent-name">（{{ panel.parentName }}）</span>
          </div>
          <el-table
            v-if="panel.data.length > 0"
            :data="panel.data"
            highlight-current-row
            :row-class-name="({ row }) => row._selected ? 'is-selected' : ''"
            @row-click="(row) => panel.onRowClick(row)"
            size="small"
            :max-height="500"
          >
            <el-table-column prop="name" label="名称" min-width="100" />
            <el-table-column prop="community_count" label="小区" sortable width="80" align="right" />
            <el-table-column prop="village_count" label="村" sortable width="70" align="right" />
            <el-table-column prop="total_count" label="合计" sortable width="80" align="right" />
          </el-table>
          <el-empty v-else-if="!panel.loading" :description="panel.emptyText" :image-size="60" />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, markRaw } from 'vue'
import { getDivisionCounts } from '@/api/masterdata'
import type { DivisionCountItem } from '@common/types/masterdata'

const provinceData = ref<DivisionCountItem[]>([])
const cityData = ref<DivisionCountItem[]>([])
const districtData = ref<DivisionCountItem[]>([])
const streetData = ref<DivisionCountItem[]>([])

const selectedProvince = ref<DivisionCountItem | null>(null)
const selectedCity = ref<DivisionCountItem | null>(null)
const selectedDistrict = ref<DivisionCountItem | null>(null)

const breadcrumb = ref<{ id: number; name: string }[]>([])

const levelTitles = ['省', '市', '区县', '街道/乡镇']

async function loadData(parentId?: number): Promise<DivisionCountItem[]> {
  const res = await getDivisionCounts(parentId ? { parent_id: parentId } : undefined)
  return res?.list || []
}

async function loadProvinces() {
  provinceData.value = await loadData()
  // Default sort by total_count desc
  provinceData.value.sort((a, b) => b.total_count - a.total_count)
}

function selectProvince(row: DivisionCountItem) {
  selectedProvince.value = row
  selectedCity.value = null
  selectedDistrict.value = null
  cityData.value = []
  districtData.value = []
  streetData.value = []
  breadcrumb.value = [{ id: row.id, name: row.name }]
  provinceData.value.forEach(r => (r as any)._selected = r.id === row.id)
  loadCities(row.id)
}

async function loadCities(provinceId: number) {
  cityData.value = await loadData(provinceId)
  cityData.value.sort((a, b) => b.total_count - a.total_count)
}

function selectCity(row: DivisionCountItem) {
  selectedCity.value = row
  selectedDistrict.value = null
  districtData.value = []
  streetData.value = []
  breadcrumb.value = [
    { id: selectedProvince.value!.id, name: selectedProvince.value!.name },
    { id: row.id, name: row.name }
  ]
  cityData.value.forEach(r => (r as any)._selected = r.id === row.id)
  loadDistricts(row.id)
}

async function loadDistricts(cityId: number) {
  districtData.value = await loadData(cityId)
  districtData.value.sort((a, b) => b.total_count - a.total_count)
}

function selectDistrict(row: DivisionCountItem) {
  selectedDistrict.value = row
  streetData.value = []
  breadcrumb.value = [
    { id: selectedProvince.value!.id, name: selectedProvince.value!.name },
    { id: selectedCity.value!.id, name: selectedCity.value!.name },
    { id: row.id, name: row.name }
  ]
  districtData.value.forEach(r => (r as any)._selected = r.id === row.id)
  loadStreets(row.id)
}

async function loadStreets(districtId: number) {
  streetData.value = await loadData(districtId)
  streetData.value.sort((a, b) => b.total_count - a.total_count)
}

function handleBreadcrumbClick(index: number) {
  if (index === 0) {
    selectProvince(selectedProvince.value!)
  } else if (index === 1) {
    selectCity(selectedCity.value!)
  }
}

const panels = reactive([
  {
    title: levelTitles[0],
    parentName: '',
    data: provinceData,
    loading: false,
    emptyText: '暂无数据',
    onRowClick: selectProvince
  },
  {
    title: levelTitles[1],
    parentName: '',
    data: cityData,
    loading: false,
    emptyText: '请选择省',
    onRowClick: selectCity
  },
  {
    title: levelTitles[2],
    parentName: '',
    data: districtData,
    loading: false,
    emptyText: '请选择市',
    onRowClick: selectDistrict
  },
  {
    title: levelTitles[3],
    parentName: '',
    data: streetData,
    loading: false,
    emptyText: '请选择区县',
    onRowClick: (_row: DivisionCountItem) => {}
  }
])

onMounted(async () => {
  await loadProvinces()
})
</script>

<style scoped lang="scss">
.division-counts-container {
  padding: 20px;
}

.breadcrumb {
  margin-bottom: 16px;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;

  a {
    color: var(--el-color-primary);
    cursor: pointer;
  }
}

.drill-down-panels {
  display: flex;
  gap: 0;
}

.panel {
  flex: 1;
  min-width: 0;
  border: 1px solid var(--el-border-color);
  border-radius: 0;

  &:first-child {
    border-radius: 4px 0 0 4px;
  }

  &:last-child {
    border-radius: 0 4px 4px 0;
  }

  &:not(:last-child) {
    border-right: none;
  }
}

.panel-title {
  font-weight: bold;
  padding: 10px 12px;
  border-bottom: 1px solid var(--el-border-color);
  background: var(--el-fill-color-light);
  font-size: 14px;
}

.parent-name {
  color: var(--el-text-color-secondary);
  font-weight: normal;
  font-size: 12px;
}

:deep(.is-selected) {
  background-color: var(--el-color-primary-light-9) !important;
}
</style>
