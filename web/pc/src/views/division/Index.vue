<template>
  <div class="division-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>行政区划管理</span>
        </div>
      </template>

      <el-tabs v-model="activeTab" type="border-card">
        <!-- Tab 1: 查询编辑 -->
        <el-tab-pane label="查询编辑" name="edit">
          <el-table
            ref="editTableRef"
            v-loading="editTableLoading"
            :key="editTableKey"
            :data="divisionList"
            row-key="id"
            :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
            :default-expand-all="false"
            lazy
            :load="loadChildren"
          >
            <el-table-column prop="code" label="区划代码" min-width="140" />
            <el-table-column prop="name" label="区划名称" min-width="200" />
            <el-table-column prop="level" label="级别" width="100">
              <template #default="{ row }">
                {{ getLevelLabel(row.level) }}
              </template>
            </el-table-column>
            <el-table-column label="操作类型" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.submission_status === 0 && row.submission_type" :type="submissionTypeMap[row.submission_type]?.type || 'info'" size="small">
                  {{ submissionTypeMap[row.submission_type]?.label || '' }}
                </el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="提交状态" width="100">
              <template #default="{ row }">
                <el-tag :type="submissionStatusMap[row.submission_status]?.type || 'info'" size="small">
                  {{ submissionStatusMap[row.submission_status]?.label || '待提交' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="280" fixed="right">
              <template #default="{ row }">
                <el-button v-if="canEdit(row)" link type="primary" size="small" @click="handleEdit(row)">编辑</el-button>
                <el-button v-if="canAddChild(row)" link type="primary" size="small" @click="handleCreateChild(row)">添加下级</el-button>
                <el-button v-if="canDelete(row)" link type="danger" size="small" @click="handleDelete(row)">删除</el-button>
                <el-button v-if="canCancelDelete(row)" link type="warning" size="small" @click="handleCancelDelete(row)">取消删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-empty v-if="!editTableLoading && divisionList.length === 0" description="暂无数据" />
        </el-tab-pane>

        <!-- Tab 2: 提交管理 -->
        <el-tab-pane label="提交管理" name="submit" lazy>
          <div class="filter-bar">
            <div class="filter-row">
              <span class="filter-label">提交状态：</span>
              <el-radio-group v-model="submitStatusFilter" @change="handleSubmitStatusChange">
                <el-radio-button :value="0">待提交</el-radio-button>
                <el-radio-button :value="1">已提交</el-radio-button>
              </el-radio-group>
              <el-button
                v-if="submitSelectedRows.length > 0"
                type="warning"
                :loading="batchSubmitLoading"
                @click="handleBatchSubmit"
              >
                批量提交 ({{ submitSelectedRows.length }})
              </el-button>
            </div>
          </div>

          <el-table
            v-loading="submitTableLoading"
            :data="submitTableData"
            row-key="id"
            style="margin-top: 16px"
            @selection-change="handleSubmitSelectionChange"
          >
            <el-table-column type="selection" width="50" :selectable="canSubmitSelect" />
            <el-table-column prop="code" label="区划代码" min-width="120" />
            <el-table-column prop="name" label="区划名称" min-width="180" />
            <el-table-column label="级别" width="100">
              <template #default="{ row }">
                {{ getLevelLabel(row.level) }}
              </template>
            </el-table-column>
            <el-table-column label="操作类型" width="100">
              <template #default="{ row }">
                <el-tag v-if="row.submission_type" :type="submissionTypeMap[row.submission_type]?.type || 'info'" size="small">
                  {{ submissionTypeMap[row.submission_type]?.label || '' }}
                </el-tag>
                <span v-else>-</span>
              </template>
            </el-table-column>
            <el-table-column label="提交状态" width="100">
              <template #default="{ row }">
                <el-tag :type="submissionStatusMap[row.submission_status]?.type || 'info'" size="small">
                  {{ submissionStatusMap[row.submission_status]?.label || '待提交' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button v-if="canSubmit(row)" link type="warning" size="small" @click="handleSubmitSingle(row)">提交</el-button>
                <el-button v-if="canWithdraw(row)" link type="primary" size="small" @click="handleWithdraw(row)">撤回</el-button>
                <el-button v-if="canPhysicalDelete(row)" link type="danger" size="small" @click="handlePhysicalDelete(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-if="submitTotal > 0"
            v-model:current-page="submitPage"
            v-model:page-size="submitPageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="submitTotal"
            layout="total, sizes, prev, pager, next, jumper"
            style="margin-top: 16px; justify-content: flex-end"
            @current-change="handleSubmitPageChange"
            @size-change="handleSubmitPageSizeChange"
          />
          <el-empty v-if="!submitTableLoading && submitTableData.length === 0" description="暂无数据" />
        </el-tab-pane>

        <!-- Tab 3: 提交记录 -->
        <el-tab-pane label="提交记录" name="records" lazy>
          <div class="filter-bar">
            <div class="filter-row">
              <span class="filter-label">审核结果：</span>
              <el-radio-group v-model="recordsReviewResultFilter" @change="handleRecordsReviewResultChange">
                <el-radio-button :value="undefined">全部</el-radio-button>
                <el-radio-button :value="0">待审核</el-radio-button>
                <el-radio-button :value="1">已通过</el-radio-button>
                <el-radio-button :value="2">已拒绝</el-radio-button>
                <el-radio-button :value="3">已撤回</el-radio-button>
              </el-radio-group>
            </div>
          </div>

          <el-table
            v-loading="recordsTableLoading"
            :data="recordsTableData"
            style="margin-top: 16px"
          >
            <el-table-column prop="entity_name" label="实体名称" min-width="200" show-overflow-tooltip />
            <el-table-column prop="entity_code" label="实体代码" min-width="140" show-overflow-tooltip />
            <el-table-column label="操作类型" width="100">
              <template #default="{ row }">
                <el-tag :type="(submissionTypeMap[row.submission_type]?.type as any) || 'info'" size="small">
                  {{ submissionTypeMap[row.submission_type]?.label || '未知' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="submit_time" label="提交时间" width="180" />
            <el-table-column label="审核结果" width="100">
              <template #default="{ row }">
                <el-tag :type="(reviewResultMap[row.review_result]?.type as any) || 'info'" size="small">
                  {{ reviewResultMap[row.review_result]?.label || '未知' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="review_notes" label="审核备注" min-width="180" show-overflow-tooltip />
          </el-table>

          <el-pagination
            v-if="recordsTotal > 0"
            v-model:current-page="recordsPage"
            v-model:page-size="recordsPageSize"
            :page-sizes="[10, 20, 50, 100]"
            :total="recordsTotal"
            layout="total, sizes, prev, pager, next, jumper"
            style="margin-top: 16px; justify-content: flex-end"
            @current-change="handleRecordsPageChange"
            @size-change="handleRecordsPageSizeChange"
          />
          <el-empty v-if="!recordsTableLoading && recordsTableData.length === 0" description="暂无数据" />
        </el-tab-pane>
      </el-tabs>
    </el-card>

    <!-- 新增/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="formData"
        :rules="formRules"
        label-width="100px"
      >
        <el-form-item label="区划代码" prop="code">
          <el-input
            v-model="formData.code"
            :disabled="isEdit"
            placeholder="请输入区划代码"
            maxlength="12"
          />
        </el-form-item>
        <el-form-item label="区划名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入区划名称" maxlength="50" />
        </el-form-item>
        <el-form-item v-if="parentDivision" label="上级区划">
          <el-input :value="parentDivision.name" disabled />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useDivisionStore } from '@/stores/division'
import { getAdministrativeDivisions, submitDivision, batchSubmitDivisions, requestDeleteDivision, cancelDeleteDivision, withdrawDivision, getMySubmissionRecords, getDivisionChildCount } from '@/api/masterdata'
import type { SubmissionRecord, AdministrativeDivision } from '@common/types/masterdata'

type StatusType = 'info' | 'warning' | 'success' | 'danger'

const submissionStatusMap: Record<number, { label: string; type: StatusType }> = {
  0: { label: '待提交', type: 'info' },
  1: { label: '已提交', type: 'warning' },
  2: { label: '已批准', type: 'success' },
  3: { label: '已拒绝', type: 'danger' }
}

const submissionTypeMap: Record<number, { label: string; type: StatusType }> = {
  1: { label: '新增', type: 'primary' },
  2: { label: '修改', type: 'warning' },
  3: { label: '删除', type: 'danger' }
}

const divisionStore = useDivisionStore()
const activeTab = ref('edit')

// ==================== Permission Helpers ====================

const canEdit = (row: any) => {
  const s = row.submission_status
  const t = row.submission_type
  if (s === 0 && t !== 3) return true
  if (s === 3) return true
  return false
}

const canAddChild = (row: any) => row.submission_status === 2 && row.level < 3

const canDelete = (row: any) => {
  if (row.submission_status === 0 && row.submission_type === 1) return true
  if (row.submission_status === 2) return true
  return false
}

const canCancelDelete = (row: any) => row.submission_status === 0 && row.submission_type === 3

const canSubmit = (row: any) => row.submission_status === 0
const canSubmitSelect = (row: any) => row.submission_status === 0
const canWithdraw = (row: any) => row.submission_status === 1
const canPhysicalDelete = (row: any) => row.submission_status === 0 && row.submission_type === 1

// ==================== Tab 3: 提交记录 ====================

type ReviewResultType = 'info' | 'warning' | 'success' | 'danger'

const reviewResultMap: Record<number, { label: string; type: ReviewResultType }> = {
  0: { label: '待审核', type: 'warning' },
  1: { label: '已通过', type: 'success' },
  2: { label: '已拒绝', type: 'danger' },
  3: { label: '已撤回', type: 'info' }
}

const recordsTableLoading = ref(false)
const recordsTableData = ref<SubmissionRecord[]>([])
const recordsTotal = ref(0)
const recordsPage = ref(1)
const recordsPageSize = ref(20)
const recordsReviewResultFilter = ref<number | undefined>(undefined)

const handleRecordsReviewResultChange = () => {
  recordsPage.value = 1
  handleRecordsSearch()
}

const handleRecordsSearch = async () => {
  recordsTableLoading.value = true
  try {
    const params: any = {
      entity_type: 'administrative_division',
      page: recordsPage.value,
      page_size: recordsPageSize.value
    }
    if (recordsReviewResultFilter.value !== undefined) {
      params.review_result = recordsReviewResultFilter.value
    }
    const res = await getMySubmissionRecords(params)
    if (res) {
      const all = (res as any).list || []
      recordsTableData.value = all.filter((r: any) => (r.entity_code || '').length <= 6)
      recordsTotal.value = recordsTableData.value.length
    }
  } catch (error: any) {
    ElMessage.error(error.message || '查询提交记录失败')
    recordsTableData.value = []
    recordsTotal.value = 0
  } finally {
    recordsTableLoading.value = false
  }
}

const handleRecordsPageChange = (page: number) => { recordsPage.value = page; handleRecordsSearch() }
const handleRecordsPageSizeChange = (pageSize: number) => { recordsPageSize.value = pageSize; recordsPage.value = 1; handleRecordsSearch() }

const getLevelLabel = (level: number): string => {
  const labels: Record<number, string> = { 1: '省级', 2: '市级', 3: '区县级' }
  return labels[level] || '未知'
}

// ==================== Tab 1: 查询编辑 ====================

const divisionList = ref<AdministrativeDivision[]>([])
const editTableLoading = ref(false)
const editTableKey = ref(0)
const expandedRowIds = ref<Set<number>>(new Set())
const editTableRef = ref<any>(null)

const refreshEditTable = async () => {
  const savedExpanded = new Set(expandedRowIds.value)
  editTableKey.value++
  expandedRowIds.value = savedExpanded
  await nextTick()
  await loadRootDivisions()
  await nextTick()
  restoreExpandedRows()
}

const loadRootDivisions = async () => {
  editTableLoading.value = true
  try {
    const response = await getAdministrativeDivisions({ level: 1, page_size: 1000 })
    const list = (response as any)?.list || []
    divisionList.value = list.map((item: any) => ({ ...item, hasChildren: true }))
  } catch (error: any) {
    ElMessage.error(error.message || '查询失败')
    divisionList.value = []
  } finally {
    editTableLoading.value = false
  }
}

const loadChildren = async (row: any, _treeNode: any, resolve: any) => {
  try {
    const response = await getAdministrativeDivisions({ parent_id: row.id, page_size: 1000 })
    const children = (response as any)?.list || []
    row.children = children.map((item: any) => ({ ...item, hasChildren: item.level < 3 }))
    expandedRowIds.value.add(row.id)
    resolve(row.children)
  } catch { resolve([]) }
}

const restoreExpandedRows = async () => {
  if (!editTableRef.value) return
  for (const id of expandedRowIds.value) {
    const row = findRowById(divisionList.value, id)
    if (row) {
      try {
        const response = await getAdministrativeDivisions({ parent_id: row.id, page_size: 1000 })
        const children = (response as any)?.list || []
        row.children = children.map((item: any) => ({ ...item, hasChildren: item.level < 3 }))
        await nextTick()
        editTableRef.value.toggleRowExpansion({ id }, true)
      } catch { /* ignore */ }
    }
  }
}

const findRowById = (rows: any[], id: number): any => {
  for (const row of rows) {
    if (row.id === id) return row
    if (row.children) {
      const found = findRowById(row.children, id)
      if (found) return found
    }
  }
  return null
}

// ==================== Tab 2: 提交管理 ====================

const submitStatusFilter = ref<number>(0)
const submitTableData = ref<AdministrativeDivision[]>([])
const submitTableLoading = ref(false)
const submitTotal = ref(0)
const submitPage = ref(1)
const submitPageSize = ref(20)
const submitSelectedRows = ref<AdministrativeDivision[]>([])
const batchSubmitLoading = ref(false)

const handleSubmitStatusChange = () => { submitPage.value = 1; handleSubmitSearch() }

const handleSubmitSearch = async () => {
  submitTableLoading.value = true
  try {
    const params: any = { page: submitPage.value, page_size: submitPageSize.value }
    if (submitStatusFilter.value !== undefined) params.submission_status = submitStatusFilter.value
    const response = await getAdministrativeDivisions(params)
    const list = (response as any)?.list || []
    submitTableData.value = list
    submitTotal.value = (response as any)?.total || 0
  } catch (error: any) {
    ElMessage.error(error.message || '查询失败')
    submitTableData.value = []
    submitTotal.value = 0
  } finally {
    submitTableLoading.value = false
  }
}

const handleSubmitPageChange = (page: number) => { submitPage.value = page; handleSubmitSearch() }
const handleSubmitPageSizeChange = (pageSize: number) => { submitPageSize.value = pageSize; submitPage.value = 1; handleSubmitSearch() }

const handleSubmitSelectionChange = (rows: AdministrativeDivision[]) => { submitSelectedRows.value = rows }

const handleSubmitSingle = async (row: AdministrativeDivision) => {
  try {
    await ElMessageBox.confirm('确定要提交该区划进行审核吗？', '提示', { type: 'warning' })
    await submitDivision(row.id)
    ElMessage.success('提交成功')
    handleSubmitSearch()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '提交失败')
  }
}

const handleWithdraw = async (row: AdministrativeDivision) => {
  try {
    await ElMessageBox.confirm('确定要撤回该提交吗？撤回后可继续编辑。', '撤回确认', { type: 'warning' })
    await withdrawDivision(row.id)
    ElMessage.success('撤回成功')
    handleSubmitSearch()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '撤回失败')
  }
}

const handlePhysicalDelete = async (row: AdministrativeDivision) => {
  if (!(await checkBeforeDelete(row))) return
  try {
    await ElMessageBox.confirm(`确定要删除"${row.name}"吗？删除后将无法恢复。`, '删除确认', { type: 'warning' })
    await divisionStore.deleteDivision(row.id)
    ElMessage.success('删除成功')
    handleSubmitSearch()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '删除失败')
  }
}

const handleBatchSubmit = async () => {
  if (!submitSelectedRows.value.length) return
  try {
    await ElMessageBox.confirm(
      `确定要批量提交选中的 ${submitSelectedRows.value.length} 条记录吗？`,
      '批量提交',
      { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' }
    )
    batchSubmitLoading.value = true
    const ids = submitSelectedRows.value.map(r => r.id)
    await batchSubmitDivisions(ids)
    ElMessage.success(`批量提交成功，共 ${ids.length} 条`)
    handleSubmitSearch()
  } catch (error: any) {
    if (error !== 'cancel') ElMessage.error(error.message || '批量提交失败')
  } finally {
    batchSubmitLoading.value = false
  }
}

watch(activeTab, (val) => {
  if (val === 'submit') { submitPage.value = 1; handleSubmitSearch() }
  if (val === 'records') { recordsPage.value = 1; handleRecordsSearch() }
})

// ==================== 对话框（查询编辑 Tab 共用） ====================

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const parentDivision = ref<AdministrativeDivision | null>(null)

const dialogTitle = computed(() => {
  if (parentDivision.value) {
    return `添加下级 - ${parentDivision.value.name}`
  }
  return isEdit.value ? '编辑区划' : '新增区划'
})

interface FormData {
  id?: number; code: string; name: string; level: number
  parent_id?: number | null; sort_order: number
}

const formData = ref<FormData>({ code: '', name: '', level: 1, sort_order: 0 })
const siblingCodes = ref<string[]>([])

const validateCode = (_rule: any, value: string, callback: (error?: Error) => void) => {
  if (!value || isEdit.value) return callback()
  if (parentDivision.value) {
    const parentCode = parentDivision.value.code
    if (!value.startsWith(parentCode)) return callback(new Error(`代码必须以父级代码"${parentCode}"开头`))
    if (value.length === parentCode.length) return callback(new Error('请输入父级代码后的自编码部分'))
  }
  if (siblingCodes.value.length > 0) {
    const expectedLen = siblingCodes.value[0].length
    if (value.length !== expectedLen) return callback(new Error(`代码长度应为${expectedLen}位，与同级已有数据保持一致`))
  }
  callback()
}

const formRules: FormRules<FormData> = {
  code: [
    { required: true, message: '请输入区划代码', trigger: 'blur' },
    { pattern: /^\d{2,12}$/, message: '区划代码必须是2-12位数字', trigger: 'blur' }
  ],
  name: [
    { required: true, message: '请输入区划名称', trigger: 'blur' },
    { min: 2, max: 50, message: '区划名称长度为2-50个字符', trigger: 'blur' }
  ]
}

const handleCreateChild = async (row: AdministrativeDivision) => {
  isEdit.value = false
  parentDivision.value = row
  try {
    const res = await getAdministrativeDivisions({ parent_id: row.id, page_size: 1000 })
    siblingCodes.value = ((res as any)?.list || []).map((item: any) => item.code)
  } catch {
    siblingCodes.value = []
  }
  formData.value = { code: row.code, name: '', level: row.level + 1, parent_id: row.id, sort_order: 0 }
  dialogVisible.value = true
}

const handleEdit = (row: AdministrativeDivision) => {
  isEdit.value = true
  parentDivision.value = null
  formData.value = { id: row.id, code: row.code, name: row.name, level: row.level, parent_id: row.parent_id, sort_order: row.sort_order }
  dialogVisible.value = true
}

const checkBeforeDelete = async (row: AdministrativeDivision): Promise<boolean> => {
  try {
    const res = await getDivisionChildCount(row.id)
    if (res?.has_data) {
      const details: string[] = []
      if (res.has_child_divisions) details.push('下级区划')
      if (res.has_residential_areas) details.push('关联小区/村')
      await ElMessageBox.alert(`该节点下存在${details.join('和')}数据，请先删除下级数据后再操作。`, '无法删除', { type: 'warning' })
      return false
    }
    return true
  } catch {
    return true
  }
}

const handleDelete = async (row: AdministrativeDivision) => {
  if (!(await checkBeforeDelete(row))) return
  if (row.submission_status === 0 && row.submission_type === 1) {
    try {
      await ElMessageBox.confirm(`确定要删除"${row.name}"吗？删除后将无法恢复。`, '删除确认',
        { confirmButtonText: '确定', cancelButtonText: '取消', type: 'warning' })
      await divisionStore.deleteDivision(row.id)
      ElMessage.success('删除成功')
      refreshEditTable()
    } catch (error: any) { if (error !== 'cancel') ElMessage.error('删除失败') }
  } else if (row.submission_status === 2) {
    try {
      await ElMessageBox.confirm('确定要申请删除该区划吗？删除需审批通过后生效。', '发起删除', { type: 'warning' })
      await requestDeleteDivision(row.id)
      ElMessage.success('已发起删除申请')
      refreshEditTable()
    } catch (error: any) { if (error !== 'cancel') ElMessage.error(error.message || '操作失败') }
  }
}

const handleCancelDelete = async (row: AdministrativeDivision) => {
  try {
    await ElMessageBox.confirm('确定要取消删除申请吗？取消后数据将恢复为已批准状态。', '取消删除', { type: 'warning' })
    await cancelDeleteDivision(row.id)
    ElMessage.success('已取消删除，数据恢复为已批准')
    refreshEditTable()
  } catch (error: any) { if (error !== 'cancel') ElMessage.error(error.message || '操作失败') }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    submitting.value = true
    if (isEdit.value && formData.value.id) {
      await divisionStore.updateDivision(formData.value.id, { name: formData.value.name, sort_order: formData.value.sort_order })
      ElMessage.success('更新成功')
      dialogVisible.value = false
      refreshEditTable()
    } else {
      await divisionStore.createDivision({
        code: formData.value.code, name: formData.value.name, level: formData.value.level,
        parent_id: formData.value.parent_id, sort_order: formData.value.sort_order
      })
      ElMessage.success('创建成功')
      dialogVisible.value = false
      refreshEditTable()
    }
  } catch (error: any) {
    if (error !== false) ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
  parentDivision.value = null
  siblingCodes.value = []
}

onMounted(() => {
  loadRootDivisions()
})
</script>

<style scoped lang="scss">
.division-container {
  padding: 20px;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .filter-bar {
    .filter-row {
      display: flex;
      align-items: center;
      gap: 12px;

      .filter-label {
        font-size: 14px;
        color: #606266;
        white-space: nowrap;
      }
    }
  }
}
</style>
