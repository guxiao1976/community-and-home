<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getDeletedCounts, getDeletedItems, restoreDeletedItem } from '@/api/masterdata'
import type { DeletedItem, DeletedCounts } from '@common/types/masterdata'

const loading = ref(false)
const tableData = ref<DeletedItem[]>([])
const total = ref(0)

const filters = reactive({
  entity_type: '' as string,
  page: 1,
  page_size: 20
})

const deletedCounts = ref<DeletedCounts>({
  residential_area: 0,
  administrative_division: 0,
  configuration: 0,
  sensitive_word: 0,
  total: 0
})

const restoringId = ref<number | null>(null)

const statCards = computed(() => [
  { key: 'residential_area', label: '住宅小区', color: '#409EFF' },
  { key: 'administrative_division', label: '行政区划', color: '#67C23A' },
  { key: 'configuration', label: '系统配置', color: '#E6A23C' },
  { key: 'sensitive_word', label: '敏感词', color: '#F56C6C' }
])

const entityTypeLabelMap: Record<string, string> = {
  residential_area: '住宅小区',
  administrative_division: '行政区划',
  configuration: '系统配置',
  sensitive_word: '敏感词'
}

const entityTypeTagTypeMap: Record<string, string> = {
  residential_area: '',
  administrative_division: 'success',
  configuration: 'warning',
  sensitive_word: 'danger'
}

const loadDeletedCounts = async () => {
  try {
    const res = await getDeletedCounts()
    if (res) {
      deletedCounts.value = res
    }
  } catch (error: any) {
    // silently fail for counts
  }
}

const loadDeletedItems = async () => {
  loading.value = true
  try {
    const params: any = {
      page: filters.page,
      page_size: filters.page_size
    }
    if (filters.entity_type) {
      params.entity_type = filters.entity_type
    }
    const res = await getDeletedItems(params)
    if (res) {
      tableData.value = res.list || []
      total.value = res.total || 0
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载已删除数据失败')
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handleStatCardClick = (key: string) => {
  filters.entity_type = filters.entity_type === key ? '' : key
  filters.page = 1
  loadDeletedItems()
}

const handleRestore = async (row: DeletedItem) => {
  try {
    await ElMessageBox.confirm(
      `确定要恢复「${row.name}」吗？恢复后数据将重新出现在正常列表中。`,
      '确认恢复',
      { confirmButtonText: '恢复', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }

  restoringId.value = row.id
  try {
    await restoreDeletedItem(row.entity_type, row.id)
    ElMessage.success('恢复成功')
    loadDeletedItems()
    loadDeletedCounts()
  } catch (error: any) {
    ElMessage.error(error.message || '恢复失败')
  } finally {
    restoringId.value = null
  }
}

const handlePageChange = (page: number) => {
  filters.page = page
  loadDeletedItems()
}

const handlePageSizeChange = (pageSize: number) => {
  filters.page_size = pageSize
  filters.page = 1
  loadDeletedItems()
}

onMounted(() => {
  loadDeletedCounts()
  loadDeletedItems()
})
</script>

<template>
  <div class="deleted-recovery">
    <el-card class="stat-cards" shadow="never">
      <el-row :gutter="16">
        <el-col :span="6" v-for="card in statCards" :key="card.key">
          <div
            :class="['stat-card', { active: filters.entity_type === card.key }]"
            @click="handleStatCardClick(card.key)"
          >
            <div class="stat-card-label">{{ card.label }}</div>
            <div class="stat-card-count" :style="{ color: card.color }">
              {{ (deletedCounts as any)[card.key] || 0 }}
            </div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <el-card class="table-card">
      <template #header>
        <div class="card-header">
          <span>已删除数据列表</span>
          <el-button v-if="filters.entity_type" type="info" link @click="handleStatCardClick(filters.entity_type)">
            清除筛选
          </el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" stripe style="width: 100%">
        <el-table-column label="数据类型" width="120">
          <template #default="{ row }">
            <el-tag :type="(entityTypeTagTypeMap[row.entity_type] as any) || ''" size="small">
              {{ entityTypeLabelMap[row.entity_type] || row.entity_type }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="code" label="编码/分类" min-width="140" show-overflow-tooltip />
        <el-table-column prop="delete_time" label="删除时间" width="180" />
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              link
              size="small"
              :loading="restoringId === row.id"
              @click="handleRestore(row)"
            >
              恢复
            </el-button>
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

      <el-empty v-if="!loading && tableData.length === 0" description="暂无已删除数据" />
    </el-card>
  </div>
</template>

<style scoped lang="scss">
.deleted-recovery {
  padding: 20px;

  .stat-cards {
    margin-bottom: 20px;

    .stat-card {
      cursor: pointer;
      padding: 16px;
      border-radius: 4px;
      transition: all 0.2s;
      text-align: center;

      &:hover {
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      }

      &.active {
        border-color: #0091FF;
        box-shadow: 0 2px 8px rgba(0, 145, 255, 0.2);
      }

      .stat-card-label {
        font-size: 14px;
        color: #909399;
        margin-bottom: 8px;
      }

      .stat-card-count {
        font-size: 28px;
        font-weight: 600;
      }
    }
  }

  .table-card {
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 18px;
      font-weight: 600;
    }
  }
}
</style>
