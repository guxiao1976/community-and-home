<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search, Plus, Edit, Delete, View, Upload } from '@element-plus/icons-vue'
import { getResidentialAreas, deleteResidentialArea, submitResidentialArea, batchSubmitResidentialAreas, getAdministrativeDivisions } from '@/api/masterdata'
import type { ResidentialArea, AdministrativeDivision } from '@common/types/masterdata'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

// State
const loading = ref(false)
const tableData = ref<ResidentialArea[]>([])
const total = ref(0)
const selectedRows = ref<ResidentialArea[]>([])
const statusFilter = ref<number | undefined>(undefined)

// Filters
const filters = reactive({
  province_id: undefined as number | undefined,
  city_id: undefined as number | undefined,
  county_id: undefined as number | undefined,
  street_id: undefined as number | undefined,
  community_div_id: undefined as number | undefined,
  submission_status: undefined as number | undefined,
  community_type: undefined as number | undefined,
  keyword: '',
  page: 1,
  page_size: 20
})

// Division dropdown data (five-level cascading)
const provinceOptions = ref<AdministrativeDivision[]>([])
const cityOptions = ref<AdministrativeDivision[]>([])
const districtOptions = ref<AdministrativeDivision[]>([])
const streetOptions = ref<AdministrativeDivision[]>([])
const communityOptions = ref<AdministrativeDivision[]>([])
const communityOptionsLoaded = ref(false)
const divisionLoading = ref(false)

// Enums
const communityTypeMap: Record<number, string> = {
  1: '住宅小区',
  2: '村庄',
  3: '混合型'
}

type StatusType = 'info' | 'warning' | 'success' | 'danger'

const submissionStatusMap: Record<number, { label: string; type: StatusType }> = {
  0: { label: '待提交', type: 'info' },
  1: { label: '已提交', type: 'warning' },
  2: { label: '已批准', type: 'success' },
  3: { label: '已拒绝', type: 'danger' },
  4: { label: '待删除', type: 'danger' }
}

// Load provinces on page mount
const loadProvinces = async () => {
  try {
    const res = await getAdministrativeDivisions({ level: 1, page: 1, page_size: 100 })
    provinceOptions.value = res.list || []
  } catch (error) {
    console.error('Failed to load provinces:', error)
  }
}

// Province change -> load cities
const handleProvinceChange = async (provinceId: number) => {
  filters.city_id = undefined
  filters.county_id = undefined
  filters.street_id = undefined
  filters.community_div_id = undefined
  cityOptions.value = []
  districtOptions.value = []
  streetOptions.value = []
  communityOptions.value = []
  if (!provinceId) return
  divisionLoading.value = true
  try {
    const res = await getAdministrativeDivisions({ parent_id: provinceId, level: 2, page: 1, page_size: 100 })
    cityOptions.value = res.list || []
  } catch (error) {
    console.error('Failed to load cities:', error)
  } finally {
    divisionLoading.value = false
  }
}

// City change -> load districts
const handleCityChange = async (cityId: number) => {
  filters.county_id = undefined
  filters.street_id = undefined
  filters.community_div_id = undefined
  districtOptions.value = []
  streetOptions.value = []
  communityOptions.value = []
  if (!cityId) return
  divisionLoading.value = true
  try {
    const res = await getAdministrativeDivisions({ parent_id: cityId, level: 3, page: 1, page_size: 500 })
    districtOptions.value = res.list || []
  } catch (error) {
    console.error('Failed to load districts:', error)
  } finally {
    divisionLoading.value = false
  }
}

// District change -> load streets
const handleDistrictChange = async (districtId: number) => {
  filters.street_id = undefined
  filters.community_div_id = undefined
  filters.county_id = districtId || undefined
  streetOptions.value = []
  communityOptions.value = []
  if (!districtId) return
  divisionLoading.value = true
  try {
    const res = await getAdministrativeDivisions({ parent_id: districtId, level: 4, page: 1, page_size: 500 })
    streetOptions.value = res.list || []
  } catch (error) {
    console.error('Failed to load streets:', error)
  } finally {
    divisionLoading.value = false
  }
}

// Street change -> load communities
const handleStreetChange = async (streetId: number) => {
  filters.community_div_id = undefined
  communityOptions.value = []
  communityOptionsLoaded.value = false
  if (!streetId) return
  divisionLoading.value = true
  try {
    const res = await getAdministrativeDivisions({ parent_id: streetId, level: 5, page: 1, page_size: 500 })
    communityOptions.value = res.list || []
    communityOptionsLoaded.value = true
  } catch (error) {
    console.error('Failed to load communities:', error)
    communityOptionsLoaded.value = true
  } finally {
    divisionLoading.value = false
  }
}

// Community change -> set community_div_id
const handleCommunityChange = (communityId: number) => {
  filters.community_div_id = communityId || undefined
}

// Load residential areas
const loadResidentialAreas = async () => {
  loading.value = true
  try {
    const params: any = {
      page: filters.page,
      page_size: filters.page_size
    }

    if (filters.community_div_id) params.community_div_id = filters.community_div_id
    else if (filters.street_id) params.street_id = filters.street_id
    else if (filters.county_id) params.county_id = filters.county_id
    else if (filters.city_id) params.city_id = filters.city_id
    if (filters.submission_status !== undefined) params.submission_status = filters.submission_status
    if (filters.community_type) params.community_type = filters.community_type
    if (filters.keyword) params.keyword = filters.keyword

    const res = await getResidentialAreas(params)
    tableData.value = res.list || []
    total.value = res.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || '加载小区列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  filters.page = 1
  loadResidentialAreas()
}

const handleStatusFilterChange = (val: number | undefined) => {
  filters.submission_status = val
  filters.page = 1
  selectedRows.value = []
  loadResidentialAreas()
}

const handleReset = () => {
  filters.province_id = undefined
  filters.city_id = undefined
  filters.county_id = undefined
  filters.street_id = undefined
  filters.community_div_id = undefined
  filters.submission_status = undefined
  filters.community_type = undefined
  filters.keyword = ''
  filters.page = 1
  statusFilter.value = undefined
  cityOptions.value = []
  districtOptions.value = []
  streetOptions.value = []
  communityOptions.value = []
  sessionStorage.removeItem(FILTER_KEY)
  loadResidentialAreas()
}

const handlePageChange = (page: number) => {
  filters.page = page
  loadResidentialAreas()
}

const handlePageSizeChange = (pageSize: number) => {
  filters.page_size = pageSize
  filters.page = 1
  loadResidentialAreas()
}

const handleCreate = () => {
  if (!filters.county_id) {
    ElMessage.warning('请选择到区县后再新建小区')
    return
  }
  if (!filters.street_id) {
    ElMessage.warning('请选择街道/乡镇后再新建小区')
    return
  }
  if (!communityOptionsLoaded.value) {
    ElMessage.warning('社区列表加载中，请稍后重试')
    return
  }
  if (communityOptions.value.length > 0 && !filters.community_div_id) {
    ElMessage.warning('该街道下有社区，请选择社区后再新建小区')
    return
  }

  const query: Record<string, string | number> = {
    county_id: filters.county_id,
    county_name: districtOptions.value.find(d => d.id === filters.county_id)?.name || '',
    street_id: filters.street_id,
    street_name: streetOptions.value.find(s => s.id === filters.street_id)?.name || ''
  }
  if (filters.community_div_id) {
    query.community_div_id = filters.community_div_id
    query.community_name = communityOptions.value.find(c => c.id === filters.community_div_id)?.name || ''
  }

  saveFilterState()
  router.push({ path: '/masterdata/residential-areas/create', query })
}

const handleEdit = (row: ResidentialArea) => {
  saveFilterState()
  router.push(`/masterdata/residential-areas/${row.id}/edit`)
}

const handleView = (row: ResidentialArea) => {
  saveFilterState()
  router.push(`/masterdata/residential-areas/${row.id}`)
}

const handleSubmit = async (row: ResidentialArea) => {
  try {
    await ElMessageBox.confirm('确定要提交该小区进行审核吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    loading.value = true
    await submitResidentialArea(row.id)
    ElMessage.success('提交成功')
    loadResidentialAreas()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '提交失败')
    }
  } finally {
    loading.value = false
  }
}

const handleDelete = async (row: ResidentialArea) => {
  try {
    await ElMessageBox.confirm(`确定要删除小区"${row.name}"吗？删除将提交审批，审核通过后才会真正删除。`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    loading.value = true
    await deleteResidentialArea(row.id)
    ElMessage.success('删除成功')
    loadResidentialAreas()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '删除失败')
    }
  } finally {
    loading.value = false
  }
}

const FILTER_KEY = 'residential-areas-filters'

const saveFilterState = () => {
  sessionStorage.setItem(FILTER_KEY, JSON.stringify({
    filters: {
      province_id: filters.province_id,
      city_id: filters.city_id,
      county_id: filters.county_id,
      street_id: filters.street_id,
      community_div_id: filters.community_div_id,
      submission_status: filters.submission_status,
      community_type: filters.community_type,
      keyword: filters.keyword,
      page: filters.page,
      page_size: filters.page_size
    },
    options: {
      cityOptions: cityOptions.value,
      districtOptions: districtOptions.value,
      streetOptions: streetOptions.value,
      communityOptions: communityOptions.value
    }
  }))
}

const restoreFilterState = () => {
  try {
    const saved = sessionStorage.getItem(FILTER_KEY)
    if (!saved) return false
    const state = JSON.parse(saved)
    Object.assign(filters, state.filters || {})
    statusFilter.value = filters.submission_status
    if (state.options?.cityOptions?.length) cityOptions.value = state.options.cityOptions
    if (state.options?.districtOptions?.length) districtOptions.value = state.options.districtOptions
    if (state.options?.streetOptions?.length) streetOptions.value = state.options.streetOptions
    if (state.options?.communityOptions?.length) {
      communityOptions.value = state.options.communityOptions
      communityOptionsLoaded.value = true
    }
    return true
  } catch { return false }
}

const canEdit = (row: ResidentialArea) => row.submission_status !== 1 && row.submission_status !== 4
const canSubmit = (row: ResidentialArea) => row.submission_status === 0 || row.submission_status === 3 || row.submission_status === 4
const canDelete = (row: ResidentialArea) => row.submission_status !== 4
const canSelect = (row: ResidentialArea) => row.submission_status === 0 || row.submission_status === 3 || row.submission_status === 4

onMounted(async () => {
  await loadProvinces()
  const restored = restoreFilterState()
  if (restored) {
    loadResidentialAreas()
  } else {
    loadResidentialAreas()
  }
})

const handleSelectionChange = (rows: ResidentialArea[]) => {
  selectedRows.value = rows
}

const handleBatchSubmit = async () => {
  if (selectedRows.value.length === 0) return
  try {
    await ElMessageBox.confirm(`确定要批量提交 ${selectedRows.value.length} 条小区吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    loading.value = true
    const ids = selectedRows.value.map(r => r.id)
    await batchSubmitResidentialAreas(ids)
    ElMessage.success('批量提交成功')
    loadResidentialAreas()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '批量提交失败')
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="residential-areas-list">
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="省份">
          <el-select
            v-model="filters.province_id"
            placeholder="选择省份"
            clearable
            style="width: 150px"
            @change="handleProvinceChange"
          >
            <el-option v-for="item in provinceOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="城市">
          <el-select
            v-model="filters.city_id"
            placeholder="选择城市"
            clearable
            :disabled="!filters.province_id"
            :loading="divisionLoading"
            style="width: 150px"
            @change="handleCityChange"
          >
            <el-option v-for="item in cityOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="区县">
          <el-select
            v-model="filters.county_id"
            placeholder="选择区县"
            clearable
            :disabled="!filters.city_id"
            :loading="divisionLoading"
            style="width: 130px"
            @change="handleDistrictChange"
          >
            <el-option v-for="item in districtOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="街道/乡镇">
          <el-select
            v-model="filters.street_id"
            placeholder="选择街道"
            clearable
            :disabled="!filters.county_id"
            :loading="divisionLoading"
            style="width: 140px"
            @change="handleStreetChange"
          >
            <el-option v-for="item in streetOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="社区/村">
          <el-select
            v-model="filters.community_div_id"
            placeholder="选择社区"
            clearable
            :disabled="!filters.street_id"
            :loading="divisionLoading"
            style="width: 130px"
            @change="handleCommunityChange"
          >
            <el-option v-for="item in communityOptions" :key="item.id" :label="item.name" :value="item.id" />
          </el-select>
        </el-form-item>

        <el-form-item label="小区名称">
          <el-input
            v-model="filters.keyword"
            placeholder="搜索小区名称"
            clearable
            style="width: 180px"
            @keyup.enter="handleSearch"
          />
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
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>住宅小区列表</span>
          <el-button type="primary" :icon="Plus" @click="handleCreate">新建小区</el-button>
        </div>
      </template>

      <div class="status-bar">
        <el-radio-group v-model="statusFilter" @change="handleStatusFilterChange">
          <el-radio-button :value="undefined">全部</el-radio-button>
          <el-radio-button :value="0">待提交</el-radio-button>
          <el-radio-button :value="1">已提交</el-radio-button>
          <el-radio-button :value="2">已批准</el-radio-button>
          <el-radio-button :value="3">已拒绝</el-radio-button>
          <el-radio-button :value="4">待删除</el-radio-button>
        </el-radio-group>
        <el-button
          v-if="statusFilter === 0 || statusFilter === 4"
          type="warning"
          :disabled="selectedRows.length === 0"
          @click="handleBatchSubmit"
        >
          批量提交 ({{ selectedRows.length }})
        </el-button>
      </div>

      <el-table v-loading="loading" :data="tableData" stripe style="width: 100%" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="50" :selectable="canSelect" />
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="小区编码" width="140" />
        <el-table-column prop="name" label="小区名称" min-width="150" />
        <el-table-column prop="address" label="地址" min-width="200" />
        <el-table-column label="城市ID" width="90">
          <template #default="{ row }">{{ row.city_id ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="区县ID" width="90">
          <template #default="{ row }">{{ row.county_id ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="街道ID" width="90">
          <template #default="{ row }">{{ row.street_id ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="社区ID" width="90">
          <template #default="{ row }">{{ row.community_div_id ?? '-' }}</template>
        </el-table-column>
        <el-table-column label="小区类型" width="120">
          <template #default="{ row }">{{ communityTypeMap[row.community_type] || '未知' }}</template>
        </el-table-column>
        <el-table-column label="提交状态" width="120">
          <template #default="{ row }">
            <el-tag :type="submissionStatusMap[row.submission_status]?.type || 'info'">
              {{ submissionStatusMap[row.submission_status]?.label || '未知' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="280" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" :icon="View" size="small" link @click="handleView(row)">查看</el-button>
            <el-button v-if="canEdit(row)" type="primary" :icon="Edit" size="small" link @click="handleEdit(row)">编辑</el-button>
            <el-button v-if="canSubmit(row)" type="warning" :icon="Upload" size="small" link @click="handleSubmit(row)">提交审核</el-button>
            <el-button v-if="canDelete(row)" type="danger" :icon="Delete" size="small" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="filters.page"
        v-model:page-size="filters.page_size"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @current-change="handlePageChange"
        @size-change="handlePageSizeChange"
        style="margin-top: 20px; justify-content: flex-end"
      />
    </el-card>
  </div>
</template>

<style scoped lang="scss">
.residential-areas-list {
  padding: 20px;

  .filter-card { margin-bottom: 20px; }
  .table-card .status-bar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
  }
  .table-card .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
</style>
