<template>
  <div class="stats-panel">
    <div v-if="lazy && !loaded" class="lazy-bar">
      <el-button type="primary" :loading="loading" @click="handleQuery">查询</el-button>
    </div>

    <template v-if="!lazy || loaded">
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
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import type { DivisionCountItem } from '@common/types/masterdata'

interface Props {
  dataLoader: (parentId?: number) => Promise<DivisionCountItem[]>
  lazy?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  lazy: false
})

const provinceData = ref<DivisionCountItem[]>([])
const cityData = ref<DivisionCountItem[]>([])
const districtData = ref<DivisionCountItem[]>([])
const streetData = ref<DivisionCountItem[]>([])

const selectedProvince = ref<DivisionCountItem | null>(null)
const selectedCity = ref<DivisionCountItem | null>(null)
const selectedDistrict = ref<DivisionCountItem | null>(null)

const breadcrumb = ref<{ id: number; name: string }[]>([])
const loading = ref(false)
const loaded = ref(false)

async function handleQuery() {
  loading.value = true
  try {
    await loadProvinces()
    loaded.value = true
  } finally {
    loading.value = false
  }
}

const levelTitles = ['省', '市', '区县', '街道/乡镇']

function sortByTotal(data: DivisionCountItem[]) {
  data.sort((a, b) => b.total_count - a.total_count)
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

async function loadProvinces() {
  provinceData.value = await props.dataLoader()
  sortByTotal(provinceData.value)
}

async function loadCities(provinceId: number) {
  cityData.value = await props.dataLoader(provinceId)
  sortByTotal(cityData.value)
}

async function loadDistricts(cityId: number) {
  districtData.value = await props.dataLoader(cityId)
  sortByTotal(districtData.value)
}

async function loadStreets(districtId: number) {
  streetData.value = await props.dataLoader(districtId)
  sortByTotal(streetData.value)
}

function handleBreadcrumbClick(index: number) {
  if (index === 0 && selectedProvince.value) {
    selectProvince(selectedProvince.value)
  } else if (index === 1 && selectedCity.value) {
    selectCity(selectedCity.value)
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
  if (!props.lazy) {
    await loadProvinces()
  }
})
</script>

<style scoped lang="scss">
.stats-panel {
  .lazy-bar {
    display: flex;
    justify-content: center;
    padding: 40px 0;
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
}
</style>
