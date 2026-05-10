<script setup lang="ts">
import { ref, reactive, onMounted, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getPendingCounts, getPendingItems, getApprovalDetail, reviewItem, batchReviewItems, getReviewedSubmissionRecords } from '@/api/masterdata'
import type { ApprovalPendingItem, PendingCounts, ApprovalDetail, SubmissionRecord } from '@common/types/masterdata'

const EntityType = {
  ResidentialArea: 'residential_area',
  AdministrativeDivision: 'administrative_division',
  Configuration: 'configuration',
  SensitiveWord: 'sensitive_word'
} as const

const SubmissionType = {
  Create: 1,
  Update: 2,
  Delete: 3
} as const

const loading = ref(false)
const tableData = ref<ApprovalPendingItem[]>([])
const total = ref(0)
const selectedRows = ref<ApprovalPendingItem[]>([])

const filters = reactive({
  entity_type: '' as string,
  submission_type: undefined as number | undefined,
  page: 1,
  page_size: 20
})

const pendingCounts = ref<PendingCounts>({
  residential_area: 0,
  administrative_division: 0,
  configuration: 0,
  sensitive_word: 0,
  total: 0
})

// Detail drawer
const drawerVisible = ref(false)
const detailLoading = ref(false)
const approvalDetail = ref<ApprovalDetail | null>(null)
const reviewNotes = ref('')
const reviewLoading = ref(false)

// Stat cards
const statCards = computed(() => [
  { key: 'residential_area', label: '住宅小区', color: '#409EFF' },
  { key: 'administrative_division', label: '行政区划', color: '#67C23A' },
  { key: 'configuration', label: '系统配置', color: '#E6A23C' },
  { key: 'sensitive_word', label: '敏感词', color: '#F56C6C' }
])

// Entity type label/tag maps
const entityTypeLabelMap: Record<string, string> = {
  [EntityType.ResidentialArea]: '住宅小区',
  [EntityType.AdministrativeDivision]: '行政区划',
  [EntityType.Configuration]: '系统配置',
  [EntityType.SensitiveWord]: '敏感词'
}

const entityTypeTagTypeMap: Record<string, string> = {
  [EntityType.ResidentialArea]: '',
  [EntityType.AdministrativeDivision]: 'success',
  [EntityType.Configuration]: 'warning',
  [EntityType.SensitiveWord]: 'danger'
}

const submissionTypeMap: Record<number, { label: string; tagType: string }> = {
  [SubmissionType.Create]: { label: '新增', tagType: 'success' },
  [SubmissionType.Update]: { label: '修改', tagType: 'warning' },
  [SubmissionType.Delete]: { label: '待删除', tagType: 'danger' }
}

const reviewResultMap: Record<number, { label: string; tagType: string }> = {
  0: { label: '待审核', tagType: 'warning' },
  1: { label: '已通过', tagType: 'success' },
  2: { label: '已拒绝', tagType: 'danger' },
  3: { label: '已撤回', tagType: 'info' }
}

// Active tab
const activeTab = ref('pending')

// ==================== Tab 2: 审核记录 ====================

const recordsLoading = ref(false)
const recordsTableData = ref<SubmissionRecord[]>([])
const recordsTotal = ref(0)
const recordsFilters = reactive({
  review_result: undefined as number | undefined,
  page: 1,
  page_size: 20
})

const loadRecords = async () => {
  recordsLoading.value = true
  try {
    const params: any = {
      page: recordsFilters.page,
      page_size: recordsFilters.page_size
    }
    if (recordsFilters.review_result !== undefined) {
      params.review_result = recordsFilters.review_result
    }
    const res = await getReviewedSubmissionRecords(params)
    if (res) {
      recordsTableData.value = (res as any).list || []
      recordsTotal.value = (res as any).total || 0
    }
  } catch (error: any) {
    ElMessage.error(error.message || '查询审核记录失败')
    recordsTableData.value = []
    recordsTotal.value = 0
  } finally {
    recordsLoading.value = false
  }
}

const handleRecordsResultChange = (val: number | undefined) => {
  recordsFilters.review_result = val
  recordsFilters.page = 1
  loadRecords()
}

const handleRecordsPageChange = (page: number) => {
  recordsFilters.page = page
  loadRecords()
}

const handleRecordsPageSizeChange = (pageSize: number) => {
  recordsFilters.page_size = pageSize
  recordsFilters.page = 1
  loadRecords()
}

watch(activeTab, (val) => {
  if (val === 'records') {
    recordsFilters.page = 1
    loadRecords()
  }
})

// Field name mappings: database column -> Chinese label
const fieldLabelMap: Record<string, Record<string, string>> = {
  [EntityType.ResidentialArea]: {
    id: 'ID', code: '小区编码', name: '小区名称', address: '地址',
    area: '面积(㎡)', population: '人口数', community_type: '小区类型',
    county_id: '所属区县', street_id: '所属街道/乡镇', community_div_id: '所属社区/村',
    submission_status: '提交状态', submitter_id: '提交人ID', submit_time: '提交时间',
    reviewer_id: '审核人ID', review_time: '审核时间', review_notes: '审核备注',
    created_time: '创建时间', updated_time: '更新时间'
  },
  [EntityType.AdministrativeDivision]: {
    id: 'ID', code: '区划代码', name: '区划名称', level: '级别',
    path: '路径', parent_id: '上级区划', sort_order: '排序', status: '状态',
    created_by: '创建人ID', created_time: '创建时间', updated_time: '更新时间',
    submission_status: '提交状态'
  },
  [EntityType.Configuration]: {
    id: 'ID', module: '模块', key: '配置键', value: '配置值',
    value_type: '值类型', description: '描述', is_public: '是否公开',
    approval_status: '审批状态', created_time: '创建时间', updated_time: '更新时间',
    submission_status: '提交状态'
  },
  [EntityType.SensitiveWord]: {
    id: 'ID', word: '敏感词', category: '分类', severity: '严重等级',
    action: '处理动作', status: '状态', created_time: '创建时间',
    updated_time: '更新时间', submission_status: '提交状态'
  }
}

const getFieldLabel = (entityType: string, fieldName: string): string => {
  return fieldLabelMap[entityType]?.[fieldName] || fieldName
}

// Comparison fields for update diff
const comparisonFields = computed(() => {
  if (!approvalDetail.value || approvalDetail.value.submission_type !== SubmissionType.Update) return []

  let currentData: Record<string, any> = {}
  let snapshotData: Record<string, any> = {}
  try {
    currentData = JSON.parse(approvalDetail.value.current_data)
  } catch { /* ignore */ }
  try {
    snapshotData = JSON.parse(approvalDetail.value.snapshot_data)
  } catch { /* ignore */ }

  const entityType = approvalDetail.value.entity_type
  const allKeys = Array.from(new Set([...Object.keys(currentData), ...Object.keys(snapshotData)]))
  return allKeys.map(key => ({
    field: getFieldLabel(entityType, key),
    oldValue: snapshotData[key] ?? '-',
    newValue: currentData[key] ?? '-',
    changed: JSON.stringify(snapshotData[key]) !== JSON.stringify(currentData[key])
  }))
})

// Current data fields for create type
const createDataFields = computed(() => {
  if (!approvalDetail.value || approvalDetail.value.submission_type !== SubmissionType.Create) return []
  const entityType = approvalDetail.value.entity_type
  try {
    return Object.entries(JSON.parse(approvalDetail.value.current_data)).map(([key, value]) => ({
      field: getFieldLabel(entityType, key),
      value: String(value ?? '')
    }))
  } catch {
    return []
  }
})

// Delete data fields
const deleteDataFields = computed(() => {
  if (!approvalDetail.value || approvalDetail.value.submission_type !== SubmissionType.Delete) return []
  const entityType = approvalDetail.value.entity_type
  try {
    return Object.entries(JSON.parse(approvalDetail.value.snapshot_data)).map(([key, value]) => ({
      field: getFieldLabel(entityType, key),
      value: String(value ?? '')
    }))
  } catch {
    return []
  }
})

const loadPendingCounts = async () => {
  try {
    const res = await getPendingCounts()
    if (res) {
      pendingCounts.value = res
    }
  } catch (error: any) {
    // silently fail for counts
  }
}

const loadPendingItems = async () => {
  loading.value = true
  try {
    const params: any = {
      page: filters.page,
      page_size: filters.page_size
    }
    if (filters.entity_type) {
      params.entity_type = filters.entity_type
    }
    if (filters.submission_type !== undefined) {
      params.submission_type = filters.submission_type
    }
    const res = await getPendingItems(params)
    if (res) {
      tableData.value = res.list || []
      total.value = res.total || 0
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载待审核列表失败')
  } finally {
    loading.value = false
  }
}

const handleStatCardClick = (key: string) => {
  filters.entity_type = filters.entity_type === key ? '' : key
  filters.page = 1
  loadPendingItems()
}

const handleEntityTypeFilter = (type: string) => {
  filters.entity_type = type
  filters.page = 1
  loadPendingItems()
}

const handleSubmissionTypeFilter = (type: number | undefined) => {
  filters.submission_type = type
  filters.page = 1
  loadPendingItems()
}

const handleSelectionChange = (rows: ApprovalPendingItem[]) => {
  selectedRows.value = rows
}

const handlePageChange = (page: number) => {
  filters.page = page
  loadPendingItems()
}

const handlePageSizeChange = (pageSize: number) => {
  filters.page_size = pageSize
  filters.page = 1
  loadPendingItems()
}

const openDetail = async (row: ApprovalPendingItem) => {
  drawerVisible.value = true
  detailLoading.value = true
  reviewNotes.value = ''
  approvalDetail.value = null
  try {
    const res = await getApprovalDetail(row.entity_type, row.id)
    if (res) {
      approvalDetail.value = res
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载审核详情失败')
  } finally {
    detailLoading.value = false
  }
}

const handleReview = async (action: 'approve' | 'reject') => {
  if (!approvalDetail.value) return
  if (action === 'reject' && !reviewNotes.value.trim()) {
    ElMessage.warning('拒绝时必须填写审核备注')
    return
  }
  reviewLoading.value = true
  try {
    await reviewItem(approvalDetail.value.entity_type, approvalDetail.value.id as any, {
      action,
      review_notes: reviewNotes.value
    })
    ElMessage.success(action === 'approve' ? '批准成功' : '拒绝成功')
    drawerVisible.value = false
    loadPendingItems()
    loadPendingCounts()
  } catch (error: any) {
    ElMessage.error(error.message || '审核操作失败')
  } finally {
    reviewLoading.value = false
  }
}

const handleBatchApprove = async () => {
  if (!selectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      `确定要批量批准选中的 ${selectedRows.value.length} 条记录吗？`,
      '批量批准',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
  } catch {
    return
  }

  const entityType = selectedRows.value[0].entity_type
  const ids = selectedRows.value.map(r => r.id)
  loading.value = true
  try {
    const res = await batchReviewItems({
      entity_type: entityType,
      ids,
      action: 'approve'
    })
    if (res) {
      ElMessage.success(`批量批准完成，成功 ${res.success_count} 条`)
    }
    loadPendingItems()
    loadPendingCounts()
  } catch (error: any) {
    ElMessage.error(error.message || '批量批准失败')
  } finally {
    loading.value = false
  }
}

const handleBatchReject = async () => {
  if (!selectedRows.value.length) return
  try {
    const { value } = await ElMessageBox.prompt(
      '请输入拒绝原因（必填）',
      '批量拒绝',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputValidator: (val: string) => {
          if (!val || !val.trim()) return '拒绝原因不能为空'
          return true
        }
      }
    )
    const entityType = selectedRows.value[0].entity_type
    const ids = selectedRows.value.map(r => r.id)
    loading.value = true
    try {
      const res = await batchReviewItems({
        entity_type: entityType,
        ids,
        action: 'reject',
        review_notes: value
      })
      if (res) {
        ElMessage.success(`批量拒绝完成，成功 ${res.success_count} 条`)
      }
      loadPendingItems()
      loadPendingCounts()
    } catch (error: any) {
      ElMessage.error(error.message || '批量拒绝失败')
    } finally {
      loading.value = false
    }
  } catch {
    // user cancelled
  }
}

onMounted(() => {
  loadPendingCounts()
  loadPendingItems()
})
</script>

<template>
  <div class="approval-center">
    <el-tabs v-model="activeTab" type="border-card">
      <!-- Tab 1: 待审核列表 -->
      <el-tab-pane label="待审核列表" name="pending">
        <!-- Stat Cards -->
        <el-row :gutter="16" class="stat-cards">
          <el-col :span="6" v-for="card in statCards" :key="card.key">
            <el-card
              shadow="hover"
              :class="['stat-card', { active: filters.entity_type === card.key }]"
              @click="handleStatCardClick(card.key)"
            >
              <div class="stat-card-content">
                <div class="stat-card-label">{{ card.label }}</div>
                <div class="stat-card-count" :style="{ color: card.color }">
                  {{ (pendingCounts as any)[card.key] || 0 }}
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>

        <!-- Filter Bar -->
        <el-card class="filter-card">
          <div class="filter-row">
            <span class="filter-label">数据类型：</span>
            <el-button-group>
              <el-button
                :type="filters.entity_type === '' ? 'primary' : 'default'"
                @click="handleEntityTypeFilter('')"
              >全部</el-button>
              <el-button
                v-for="(label, key) in entityTypeLabelMap"
                :key="key"
                :type="filters.entity_type === key ? 'primary' : 'default'"
                @click="handleEntityTypeFilter(key)"
              >{{ label }}</el-button>
            </el-button-group>
          </div>
          <div class="filter-row" style="margin-top: 12px">
            <span class="filter-label">操作类型：</span>
            <el-button-group>
              <el-button
                :type="filters.submission_type === undefined ? 'primary' : 'default'"
                @click="handleSubmissionTypeFilter(undefined)"
              >全部</el-button>
              <el-button
                v-for="(info, val) in submissionTypeMap"
                :key="val"
                :type="filters.submission_type === Number(val) ? 'primary' : 'default'"
                @click="handleSubmissionTypeFilter(Number(val))"
              >{{ info.label }}</el-button>
            </el-button-group>
          </div>
        </el-card>

        <!-- Batch Actions Bar -->
        <div v-if="selectedRows.length > 0" class="batch-actions-bar">
          <span class="batch-actions-info">已选 {{ selectedRows.length }} 项</span>
          <el-button type="success" :loading="loading" @click="handleBatchApprove">批量通过</el-button>
          <el-button type="danger" :loading="loading" @click="handleBatchReject">批量拒绝</el-button>
        </div>

        <!-- Table -->
        <el-card class="table-card">
          <template #header><div class="card-header">待审核列表</div></template>

          <el-table
            v-loading="loading"
            :data="tableData"
            stripe
            style="width: 100%"
            @selection-change="handleSelectionChange"
          >
            <el-table-column type="selection" width="50" />
            <el-table-column label="数据类型" width="120">
              <template #default="{ row }">
                <el-tag :type="(entityTypeTagTypeMap[row.entity_type] as any) || ''" size="small">
                  {{ entityTypeLabelMap[row.entity_type] || row.entity_type }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作类型" width="100">
              <template #default="{ row }">
                <el-tag
                  :type="(submissionTypeMap[row.submission_type]?.tagType as any) || 'info'"
                  size="small"
                >
                  {{ submissionTypeMap[row.submission_type]?.label || '未知' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
            <el-table-column prop="change_summary" label="变更摘要" min-width="200" show-overflow-tooltip />
            <el-table-column prop="submit_time" label="提交时间" width="180" />
            <el-table-column label="操作" width="100" fixed="right">
              <template #default="{ row }">
                <el-button type="primary" link size="small" @click="openDetail(row)">详情</el-button>
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
      </el-tab-pane>

      <!-- Tab 2: 审核记录 -->
      <el-tab-pane label="审核记录" name="records" lazy>
        <div class="filter-card">
          <div class="filter-row">
            <span class="filter-label">审核结果：</span>
            <el-radio-group v-model="recordsFilters.review_result" @change="handleRecordsResultChange">
              <el-radio-button :value="undefined">全部</el-radio-button>
              <el-radio-button :value="0">待审核</el-radio-button>
              <el-radio-button :value="1">已通过</el-radio-button>
              <el-radio-button :value="2">已拒绝</el-radio-button>
              <el-radio-button :value="3">已撤回</el-radio-button>
            </el-radio-group>
          </div>
        </div>

        <el-card class="table-card">
          <el-table
            v-loading="recordsLoading"
            :data="recordsTableData"
            stripe
            style="width: 100%"
          >
            <el-table-column label="实体类型" width="120">
              <template #default="{ row }">
                <el-tag :type="(entityTypeTagTypeMap[row.entity_type] as any) || ''" size="small">
                  {{ entityTypeLabelMap[row.entity_type] || row.entity_type }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="entity_name" label="实体名称" min-width="180" show-overflow-tooltip />
            <el-table-column prop="entity_code" label="实体代码" min-width="140" show-overflow-tooltip />
            <el-table-column label="操作类型" width="100">
              <template #default="{ row }">
                <el-tag :type="(submissionTypeMap[row.submission_type]?.tagType as any) || 'info'" size="small">
                  {{ submissionTypeMap[row.submission_type]?.label || '未知' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="submit_time" label="提交时间" width="180" />
            <el-table-column prop="submitter_id" label="提交人" width="100" />
            <el-table-column label="审核结果" width="100">
              <template #default="{ row }">
                <el-tag :type="(reviewResultMap[row.review_result]?.tagType as any) || 'info'" size="small">
                  {{ reviewResultMap[row.review_result]?.label || '未知' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="review_notes" label="审核备注" min-width="180" show-overflow-tooltip />
          </el-table>

          <el-pagination
            v-model:current-page="recordsFilters.page"
            v-model:page-size="recordsFilters.page_size"
            :page-sizes="[10, 20, 50, 100]"
            :total="recordsTotal"
            layout="total, sizes, prev, pager, next, jumper"
            @current-change="handleRecordsPageChange"
            @size-change="handleRecordsPageSizeChange"
            style="margin-top: 20px; justify-content: flex-end"
          />

          <el-empty v-if="!recordsLoading && recordsTableData.length === 0" description="暂无数据" />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- Detail Drawer -->
    <el-drawer
      v-model="drawerVisible"
      title="审核详情"
      size="600px"
      :destroy-on-close="true"
    >
      <div v-loading="detailLoading">
        <template v-if="approvalDetail">
          <!-- Submission type header -->
          <div class="detail-header">
            <el-tag
              :type="(entityTypeTagTypeMap[approvalDetail.entity_type] as any) || ''"
              size="large"
            >
              {{ entityTypeLabelMap[approvalDetail.entity_type] || approvalDetail.entity_type }}
            </el-tag>
            <el-tag
              :type="(submissionTypeMap[approvalDetail.submission_type]?.tagType as any) || 'info'"
              size="large"
              style="margin-left: 8px"
            >
              {{ submissionTypeMap[approvalDetail.submission_type]?.label || '未知' }}
            </el-tag>
          </div>

          <!-- Create type: display current_data -->
          <template v-if="approvalDetail.submission_type === SubmissionType.Create">
            <el-descriptions title="新增内容" :column="1" border style="margin-top: 16px">
              <el-descriptions-item
                v-for="item in createDataFields"
                :key="item.field"
                :label="item.field"
              >
                {{ item.value }}
              </el-descriptions-item>
            </el-descriptions>
          </template>

          <!-- Update type: comparison table -->
          <template v-else-if="approvalDetail.submission_type === SubmissionType.Update">
            <div class="detail-section-title">变更对比</div>
            <el-table :data="comparisonFields" border style="margin-top: 8px" size="small">
              <el-table-column prop="field" label="字段名" width="150" />
              <el-table-column label="原值" min-width="150">
                <template #default="{ row }">
                  <span :class="{ 'changed-field': row.changed }">{{ row.oldValue }}</span>
                </template>
              </el-table-column>
              <el-table-column label="新值" min-width="150">
                <template #default="{ row }">
                  <span :class="{ 'changed-field': row.changed }">{{ row.newValue }}</span>
                </template>
              </el-table-column>
            </el-table>
          </template>

          <!-- Delete type: display record + warning -->
          <template v-else-if="approvalDetail.submission_type === SubmissionType.Delete">
            <el-descriptions title="删除内容" :column="1" border style="margin-top: 16px">
              <el-descriptions-item
                v-for="item in deleteDataFields"
                :key="item.field"
                :label="item.field"
              >
                {{ item.value }}
              </el-descriptions-item>
            </el-descriptions>
            <el-alert
              title="将执行软删除"
              type="warning"
              description="批准后将标记该记录为已删除状态，不会物理删除数据。"
              show-icon
              :closable="false"
              style="margin-top: 16px"
            />
          </template>

          <!-- Meta info -->
          <el-descriptions :column="2" border style="margin-top: 24px">
            <el-descriptions-item label="提交人ID">{{ approvalDetail.submitter_id }}</el-descriptions-item>
            <el-descriptions-item label="提交时间">{{ approvalDetail.submit_time }}</el-descriptions-item>
            <el-descriptions-item v-if="approvalDetail.reviewer_id" label="审核人ID">
              {{ approvalDetail.reviewer_id }}
            </el-descriptions-item>
            <el-descriptions-item v-if="approvalDetail.review_time" label="审核时间">
              {{ approvalDetail.review_time }}
            </el-descriptions-item>
            <el-descriptions-item v-if="approvalDetail.review_notes" label="审核备注" :span="2">
              {{ approvalDetail.review_notes }}
            </el-descriptions-item>
          </el-descriptions>

          <!-- Review actions -->
          <div class="review-actions">
            <el-input
              v-model="reviewNotes"
              type="textarea"
              :rows="3"
              placeholder="请输入审核备注（拒绝时必填）"
              maxlength="500"
              show-word-limit
              style="margin-bottom: 16px"
            />
            <div class="review-buttons">
              <el-button type="success" :loading="reviewLoading" @click="handleReview('approve')">
                批准
              </el-button>
              <el-button type="danger" :loading="reviewLoading" @click="handleReview('reject')">
                拒绝
              </el-button>
            </div>
          </div>
        </template>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped lang="scss">
.approval-center {
  padding: 20px;

  .stat-cards {
    margin-bottom: 20px;

    .stat-card {
      cursor: pointer;
      transition: border-color 0.2s;

      &.active {
        border-color: #0091FF;
      }

      .stat-card-content {
        text-align: center;

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
  }

  .filter-card {
    margin-bottom: 20px;

    .filter-row {
      display: flex;
      align-items: center;

      .filter-label {
        font-size: 14px;
        color: #606266;
        margin-right: 12px;
        white-space: nowrap;
      }
    }
  }

  .batch-actions-bar {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 12px 16px;
    margin-bottom: 16px;
    background-color: #ecf5ff;
    border: 1px solid #d9ecff;
    border-radius: 4px;

    .batch-actions-info {
      font-size: 14px;
      color: #0091FF;
      font-weight: 500;
    }
  }

  .table-card .card-header {
    font-size: 18px;
    font-weight: 600;
  }

  .detail-header {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
  }

  .detail-section-title {
    font-size: 16px;
    font-weight: 600;
    color: #303133;
    margin-top: 16px;
  }

  .changed-field {
    color: #f56c6c;
    font-weight: 600;
  }

  .review-actions {
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid #ebeef5;

    .review-buttons {
      display: flex;
      justify-content: flex-end;
      gap: 12px;
    }
  }
}
</style>
