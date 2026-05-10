<template>
  <div class="amap-sync">
    <div class="page-header">
      <h2>高德地图同步</h2>
      <p class="page-desc">通过高德地图 API 批量同步住宅小区数据，支持按省/市/区县同步</p>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-form :inline="true" label-width="60px">
          <el-form-item label="省份">
            <el-select v-model="filters.provinceId" placeholder="请选择" clearable style="width: 150px" @change="handleProvinceChange">
              <el-option v-for="item in provinceOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="城市">
            <el-select v-model="filters.cityId" placeholder="请选择" clearable style="width: 150px" @change="handleCityChange">
              <el-option v-for="item in cityOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="区县">
            <el-select v-model="filters.countyId" placeholder="请选择" clearable style="width: 150px">
              <el-option v-for="item in countyOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
          </el-form-item>
        </el-form>
        <div class="action-bar">
          <el-button type="primary" :disabled="!effectiveDivisionId || syncing" @click="handleSync">
            开始同步
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- Progress Panel -->
    <el-card v-if="syncing || progress" class="progress-card" style="margin-top: 16px">
      <template #header>
        <span>同步进度</span>
      </template>
      <div v-if="progress" class="progress-content">
        <!-- County-level progress (multi-county only) -->
        <template v-if="progress.total_counties > 1">
          <div style="margin-bottom: 8px; color: #606266; font-size: 14px">
            区县进度：{{ progress.current_county }} / {{ progress.total_counties }}
            <span v-if="progress.current_county_name">（{{ progress.current_county_name }}）</span>
          </div>
          <el-progress
            :percentage="Math.round(progress.current_county / progress.total_counties * 100)"
            :stroke-width="12"
            style="margin-bottom: 16px"
          />
        </template>

        <!-- Page-level progress -->
        <el-progress
          :percentage="progress.total_pages > 0 ? Math.round(progress.current_page / progress.total_pages * 100) : 0"
          :status="progressStatus"
          :stroke-width="18"
          style="margin-bottom: 16px"
        />

        <div class="progress-info">
          <p v-if="progress.status === 'running'">
            <template v-if="progress.total_counties > 1">
              正在处理第 {{ progress.current_county }}/{{ progress.total_counties }} 个区县
              <span v-if="progress.current_county_name">「{{ progress.current_county_name }}」</span>，
              第 {{ progress.current_page }}/{{ progress.total_pages }} 页
              （共发现 <strong>{{ progress.total_found }}</strong> 个小区）
            </template>
            <template v-else>
              正在处理第 <strong>{{ progress.current_page }}</strong>/{{ progress.total_pages }} 页
              （共发现 <strong>{{ progress.total_found }}</strong> 个小区）
            </template>
          </p>
          <p v-else-if="progress.status === 'completed'">
            同步完成（共 {{ progress.total_counties }} 个区县）
          </p>
          <p v-else-if="progress.status === 'failed'">
            同步失败：{{ progress.error_message }}
          </p>
        </div>

        <el-descriptions :column="4" border size="small" style="margin-top: 12px">
          <el-descriptions-item label="区县数">{{ progress.total_counties }}</el-descriptions-item>
          <el-descriptions-item label="已同步">{{ progress.total_synced }}</el-descriptions-item>
          <el-descriptions-item label="已跳过（重复）">{{ progress.total_skipped }}</el-descriptions-item>
          <el-descriptions-item label="失败">{{ progress.total_failed }}</el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAdministrativeDivisions, syncResidentialAreas, getSyncProgress } from '@/api/masterdata'
import type { SyncProgress } from '@/api/masterdata'
import { logger } from '@/utils/logger'

interface DivisionOption {
  id: number
  name: string
}

const provinceOptions = ref<DivisionOption[]>([])
const cityOptions = ref<DivisionOption[]>([])
const countyOptions = ref<DivisionOption[]>([])
const syncing = ref(false)
const progress = ref<SyncProgress | null>(null)
let pollTimer: ReturnType<typeof setInterval> | null = null

const filters = reactive({
  provinceId: undefined as number | undefined,
  cityId: undefined as number | undefined,
  countyId: undefined as number | undefined
})

const effectiveDivisionId = computed(() => {
  return filters.countyId || filters.cityId || filters.provinceId
})

const levelLabel = computed(() => {
  if (filters.countyId) return '区县'
  if (filters.cityId) return '城市'
  if (filters.provinceId) return '省份'
  return ''
})

const progressStatus = computed(() => {
  if (!progress.value) return undefined
  if (progress.value.status === 'completed') return 'success'
  if (progress.value.status === 'failed') return 'exception'
  return undefined
})

const loadDivisions = async (parentId?: number, level?: number): Promise<DivisionOption[]> => {
  try {
    const params: any = { page_size: 1000 }
    if (parentId) params.parent_id = parentId
    if (level) params.level = level
    const res = await getAdministrativeDivisions(params)
    return (res?.list || []).map((d: any) => ({ id: d.id, name: d.name }))
  } catch {
    return []
  }
}

const handleProvinceChange = async () => {
  filters.cityId = undefined
  filters.countyId = undefined
  cityOptions.value = []
  countyOptions.value = []
  if (!filters.provinceId) return
  cityOptions.value = await loadDivisions(filters.provinceId, 2)
}

const handleCityChange = async () => {
  filters.countyId = undefined
  countyOptions.value = []
  if (!filters.cityId) return
  countyOptions.value = await loadDivisions(filters.cityId, 3)
}

const handleSync = async () => {
  const divId = effectiveDivisionId.value
  if (!divId) return

  try {
    await ElMessageBox.confirm(
      `确定要同步该${levelLabel.value}下所有区县的小区数据吗？此操作将从高德地图获取数据并写入数据库。`,
      '确认同步',
      { confirmButtonText: '开始同步', cancelButtonText: '取消', type: 'warning' }
    )
    syncing.value = true
    progress.value = null
    try {
      const res = await syncResidentialAreas({ division_id: divId })
      startPolling(res.task_id)
    } catch (error: any) {
      syncing.value = false
      ElMessage.error(error.message || '启动同步失败')
    }
  } catch (error: any) {
    if (error !== 'cancel') {
      syncing.value = false
      ElMessage.error(error.message || '启动同步失败')
    }
  }
}

const startPolling = (taskId: string) => {
  pollTimer = setInterval(async () => {
    try {
      const p = await getSyncProgress(taskId)
      progress.value = p
      if (p.status === 'completed' || p.status === 'failed') {
        stopPolling()
        syncing.value = false
        if (p.status === 'completed') {
          const countyInfo = p.total_counties > 1 ? `共 ${p.total_counties} 个区县，` : ''
          ElMessage.success(`同步完成！${countyInfo}共发现 ${p.total_found} 个小区，同步 ${p.total_synced} 个，跳过 ${p.total_skipped} 个`)
        } else {
          ElMessage.error(`同步失败: ${p.error_message || '未知错误'}`)
        }
      }
    } catch (error: any) {
      console.error('Poll error:', error)
    }
  }, 2000)
}

const stopPolling = () => {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  logger.componentMounted('AMap Sync')
  provinceOptions.value = await loadDivisions(undefined, 1)
})

onUnmounted(() => {
  stopPolling()
})
</script>

<style scoped lang="scss">
@import '@/styles/variables.scss';

.amap-sync {
  .page-header {
    margin-bottom: 20px;

    h2 {
      margin: 0 0 4px 0;
      font-size: 20px;
      font-weight: 500;
    }

    .page-desc {
      margin: 0;
      color: $text-secondary;
      font-size: 14px;
    }
  }

  .filter-bar {
    .action-bar {
      margin-top: 12px;
      padding-left: 68px;
    }
  }

  .progress-card {
    .progress-content {
      .progress-info {
        color: $text-regular;
        font-size: 14px;

        p {
          margin: 0;
        }
      }
    }
  }
}
</style>
