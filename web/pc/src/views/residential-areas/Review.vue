<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search, Check, Close } from '@element-plus/icons-vue'
import { getResidentialAreas, reviewResidentialArea } from '@/api/masterdata'
import type { ResidentialArea } from '@common/types/masterdata'
import type { FormInstance } from 'element-plus'

const loading = ref(false)
const tableData = ref<ResidentialArea[]>([])
const total = ref(0)

const filters = reactive({
  submission_status: 1,
  page: 1,
  page_size: 20
})

const reviewDialogVisible = ref(false)
const reviewForm = reactive({ id: 0, action: 'approve' as 'approve' | 'reject', review_notes: '' })
const reviewFormRef = ref<FormInstance>()

const communityTypeMap: Record<number, string> = { 1: '住宅小区', 2: '村庄', 3: '混合型' }

const loadResidentialAreas = async () => {
  loading.value = true
  try {
    const res = await getResidentialAreas({
      submission_status: filters.submission_status as any,
      page: filters.page,
      page_size: filters.page_size
    })
    if (res) {
      tableData.value = res.list || []
      total.value = res.total || 0
    }
  } catch (error: any) {
    ElMessage.error(error.message || '加载小区列表失败')
  } finally {
    loading.value = false
  }
}

const handleSearch = () => { filters.page = 1; loadResidentialAreas() }
const handlePageChange = (page: number) => { filters.page = page; loadResidentialAreas() }
const handlePageSizeChange = (pageSize: number) => { filters.page_size = pageSize; filters.page = 1; loadResidentialAreas() }

const openReviewDialog = (row: ResidentialArea, action: 'approve' | 'reject') => {
  reviewForm.id = row.id
  reviewForm.action = action
  reviewForm.review_notes = ''
  reviewDialogVisible.value = true
}

const submitReview = async () => {
  if (reviewForm.action === 'reject' && !reviewForm.review_notes.trim()) {
    ElMessage.warning('拒绝时必须填写审核备注')
    return
  }
  try {
    loading.value = true
    await reviewResidentialArea(reviewForm.id, {
      action: reviewForm.action,
      review_notes: reviewForm.review_notes
    })
    ElMessage.success(reviewForm.action === 'approve' ? '批准成功' : '拒绝成功')
    reviewDialogVisible.value = false
    loadResidentialAreas()
  } catch (error: any) {
    ElMessage.error(error.message || '审核失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => { loadResidentialAreas() })
</script>

<template>
  <div class="residential-areas-review">
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="提交状态">
          <el-select v-model="filters.submission_status" style="width: 150px">
            <el-option label="已提交" :value="1" />
            <el-option label="已批准" :value="2" />
            <el-option label="已拒绝" :value="3" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :icon="Search" @click="handleSearch">搜索</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="table-card">
      <template #header><div class="card-header">小区审核</div></template>

      <el-table v-loading="loading" :data="tableData" stripe style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="code" label="小区编码" width="140" />
        <el-table-column prop="name" label="小区名称" min-width="150" />
        <el-table-column prop="address" label="地址" min-width="200" />
        <el-table-column prop="area" label="面积(㎡)" width="120" />
        <el-table-column prop="population" label="人口数" width="100" />
        <el-table-column label="小区类型" width="120">
          <template #default="{ row }">{{ communityTypeMap[row.community_type] || '未知' }}</template>
        </el-table-column>
        <el-table-column prop="submit_time" label="提交时间" width="180" />
        <el-table-column prop="submitter_id" label="提交人ID" width="100" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <template v-if="row.submission_status === 1">
              <el-button type="success" :icon="Check" size="small" @click="openReviewDialog(row, 'approve')">批准</el-button>
              <el-button type="danger" :icon="Close" size="small" @click="openReviewDialog(row, 'reject')">拒绝</el-button>
            </template>
            <template v-else>
              <el-tag v-if="row.submission_status === 2" type="success">已批准</el-tag>
              <el-tag v-else-if="row.submission_status === 3" type="danger">已拒绝</el-tag>
            </template>
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

    <el-dialog v-model="reviewDialogVisible" :title="reviewForm.action === 'approve' ? '批准小区' : '拒绝小区'" width="500px">
      <el-form :model="reviewForm" label-width="100px">
        <el-form-item label="审核操作">
          <el-tag :type="reviewForm.action === 'approve' ? 'success' : 'danger'">
            {{ reviewForm.action === 'approve' ? '批准' : '拒绝' }}
          </el-tag>
        </el-form-item>
        <el-form-item label="审核备注" :required="reviewForm.action === 'reject'">
          <el-input
            v-model="reviewForm.review_notes"
            type="textarea" :rows="4"
            :placeholder="reviewForm.action === 'reject' ? '请输入拒绝原因（必填）' : '请输入审核备注（选填）'"
            maxlength="500" show-word-limit
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="reviewDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="loading" @click="submitReview">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped lang="scss">
.residential-areas-review {
  padding: 20px;
  .filter-card { margin-bottom: 20px; }
  .table-card .card-header { font-size: 18px; font-weight: 600; }
}
</style>
