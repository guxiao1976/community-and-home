<template>
  <div class="residential-query">
    <div class="filter-section">
      <el-form :inline="true" :model="filters">
        <el-form-item label="省份">
          <el-select v-model="filters.province_id" placeholder="选择省份" clearable style="width: 150px" @change="handleProvinceChange">
            <el-option v-for="item in provinceOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="城市">
          <el-select v-model="filters.city_id" placeholder="选择城市" clearable :disabled="!filters.province_id" :loading="divisionLoading" style="width: 150px" @change="handleCityChange">
            <el-option v-for="item in cityOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="区县">
          <el-select v-model="filters.county_id" placeholder="选择区县" clearable :disabled="!filters.city_id" :loading="divisionLoading" style="width: 130px" @change="handleDistrictChange">
            <el-option v-for="item in districtOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="街道/乡镇">
          <el-select v-model="filters.street_id" placeholder="选择街道" clearable :disabled="!filters.county_id" :loading="divisionLoading" style="width: 140px" @change="handleStreetChange">
            <el-option v-for="item in streetOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="社区/村">
          <el-select v-model="filters.community_div_id" placeholder="选择社区" clearable :disabled="!filters.street_id" :loading="divisionLoading" style="width: 130px" @change="handleCommunityChange">
            <el-option v-for="item in communityOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="小区名称">
          <el-input v-model="filters.keyword" placeholder="搜索小区名称" clearable style="width: 180px" @keyup.enter="handleSearch" />
        </el-form-item>

        <el-form-item label="小区类型">
          <el-select v-model="filters.community_type" placeholder="请选择类型" clearable style="width: 150px">
            <el-option label="住宅小区" :value="1" />
            <el-option label="村庄" :value="2" />
            <el-option label="混合型" :value="3" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" :disabled="!filters.city_id" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <el-table v-loading="loading" :data="tableData" stripe border style="width: 100%">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="小区名称" min-width="150" />
      <el-table-column prop="code" label="小区编码" width="140" />
      <el-table-column label="城市名称（ID）" min-width="140">
        <template #default="{ row }">{{ formatDivName(row.city_name, row.city_id) }}</template>
      </el-table-column>
      <el-table-column label="县区名称（ID）" min-width="140">
        <template #default="{ row }">{{ formatDivName(row.county_name, row.county_id) }}</template>
      </el-table-column>
      <el-table-column label="街道/乡镇名称（ID）" min-width="160">
        <template #default="{ row }">{{ formatDivName(row.street_name, row.street_id) }}</template>
      </el-table-column>
      <el-table-column label="社区名称（ID）" min-width="160">
        <template #default="{ row }">{{ formatDivName(row.community_name, row.community_div_id) }}</template>
      </el-table-column>
      <el-table-column prop="address" label="地址" min-width="200" show-overflow-tooltip />
      <el-table-column label="小区类型" width="100">
        <template #default="{ row }">{{ communityTypeMap[row.community_type] || '-' }}</template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button type="primary" link @click="handleView(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-wrapper" v-if="total > pagination.pageSize">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="total"
        :page-sizes="[20, 50]"
        layout="total, sizes, prev, pager, next"
        @size-change="handleSearch"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Search } from '@element-plus/icons-vue'
import { getAdministrativeDivisions, queryResidentialAreas } from '@/api/masterdata'
import type { QueryResidentialAreaItem } from '@/api/masterdata'
import type { AdministrativeDivision } from '@common/types/masterdata'

const router = useRouter()
const loading = ref(false)
const tableData = ref<QueryResidentialAreaItem[]>([])
const total = ref(0)
const divisionLoading = ref(false)

const communityTypeMap: Record<number, string> = {
  1: '住宅小区',
  2: '村庄',
  3: '混合型'
}

const filters = reactive({
  province_id: undefined as number | undefined,
  city_id: undefined as number | undefined,
  county_id: undefined as number | undefined,
  street_id: undefined as number | undefined,
  community_div_id: undefined as number | undefined,
  community_type: undefined as number | undefined,
  keyword: ''
})

const pagination = reactive({ page: 1, pageSize: 20 })

const provinceOptions = ref<AdministrativeDivision[]>([])
const cityOptions = ref<AdministrativeDivision[]>([])
const districtOptions = ref<AdministrativeDivision[]>([])
const streetOptions = ref<AdministrativeDivision[]>([])
const communityOptions = ref<AdministrativeDivision[]>([])

const formatDivName = (name: string, id: number | null) => {
  if (!name) return '-'
  if (id) return `${name}（${id}）`
  return name
}

const loadDivisions = async (parentId?: number, level?: number): Promise<AdministrativeDivision[]> => {
  try {
    const params: any = { page: 1, page_size: 1000 }
    if (parentId) params.parent_id = parentId
    if (level) params.level = level
    const res = await getAdministrativeDivisions(params)
    return (res?.list || []).map((d: any) => ({ id: d.id, name: d.name }))
  } catch {
    return []
  }
}

const handleProvinceChange = async () => {
  filters.city_id = undefined
  filters.county_id = undefined
  filters.street_id = undefined
  filters.community_div_id = undefined
  cityOptions.value = []
  districtOptions.value = []
  streetOptions.value = []
  communityOptions.value = []
  if (!filters.province_id) return
  divisionLoading.value = true
  cityOptions.value = await loadDivisions(filters.province_id, 2)
  divisionLoading.value = false
}

const handleCityChange = async () => {
  filters.county_id = undefined
  filters.street_id = undefined
  filters.community_div_id = undefined
  districtOptions.value = []
  streetOptions.value = []
  communityOptions.value = []
  if (!filters.city_id) return
  divisionLoading.value = true
  districtOptions.value = await loadDivisions(filters.city_id, 3)
  divisionLoading.value = false
}

const handleDistrictChange = async () => {
  filters.street_id = undefined
  filters.community_div_id = undefined
  streetOptions.value = []
  communityOptions.value = []
  if (!filters.county_id) return
  divisionLoading.value = true
  streetOptions.value = await loadDivisions(filters.county_id, 4)
  divisionLoading.value = false
}

const handleStreetChange = async () => {
  filters.community_div_id = undefined
  communityOptions.value = []
  if (!filters.street_id) return
  divisionLoading.value = true
  communityOptions.value = await loadDivisions(filters.street_id, 5)
  divisionLoading.value = false
}

const handleCommunityChange = () => {}

const doQuery = async () => {
  loading.value = true
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    }
    if (filters.community_div_id) params.community_div_id = filters.community_div_id
    else if (filters.street_id) params.street_id = filters.street_id
    else if (filters.county_id) params.county_id = filters.county_id
    else if (filters.city_id) params.city_id = filters.city_id
    if (filters.community_type) params.community_type = filters.community_type
    if (filters.keyword) params.keyword = filters.keyword

    const res = await queryResidentialAreas(params)
    tableData.value = res?.list || []
    total.value = res?.total || 0
  } catch {
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  pagination.page = 1
  doQuery()
}

const handlePageChange = () => {
  doQuery()
}

const handleReset = () => {
  filters.province_id = undefined
  filters.city_id = undefined
  filters.county_id = undefined
  filters.street_id = undefined
  filters.community_div_id = undefined
  filters.community_type = undefined
  filters.keyword = ''
  pagination.page = 1
  cityOptions.value = []
  districtOptions.value = []
  streetOptions.value = []
  communityOptions.value = []
  tableData.value = []
  total.value = 0
}

const handleView = (row: QueryResidentialAreaItem) => {
  router.push(`/masterdata/residential-areas/${row.id}`)
}

onMounted(async () => {
  provinceOptions.value = await loadDivisions(undefined, 1)
})
</script>

<style scoped lang="scss">
.residential-query {
  .filter-section {
    margin-bottom: 16px;
  }

  .pagination-wrapper {
    margin-top: 16px;
    display: flex;
    justify-content: flex-end;
  }
}
</style>
