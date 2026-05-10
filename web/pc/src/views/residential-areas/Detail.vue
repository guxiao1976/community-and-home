<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { getResidentialAreaById } from '@/api/masterdata'
import type { ResidentialArea } from '@common/types/masterdata'

const router = useRouter()
const route = useRoute()
const areaId = Number(route.params.id)

const loading = ref(false)
const area = ref<ResidentialArea | null>(null)

const communityTypeMap: Record<number, string> = { 1: '住宅小区', 2: '村庄', 3: '混合型' }
type StatusType = 'info' | 'warning' | 'success' | 'danger'
const submissionStatusMap: Record<number, { label: string; type: StatusType }> = {
  0: { label: '草稿', type: 'info' },
  1: { label: '已提交', type: 'warning' },
  2: { label: '已批准', type: 'success' },
  3: { label: '已拒绝', type: 'danger' }
}

const loadArea = async () => {
  loading.value = true
  try {
    const res: any = await getResidentialAreaById(areaId)
    area.value = res.residential_area || res
  } catch (error: any) {
    ElMessage.error(error.message || '加载小区信息失败')
    router.back()
  } finally {
    loading.value = false
  }
}

const handleBack = () => router.back()

const handleEdit = () => {
  if (area.value && (area.value.submission_status === 0 || area.value.submission_status === 3)) {
    router.push(`/masterdata/residential-areas/${areaId}/edit`)
  } else {
    ElMessage.warning('已批准的小区不能编辑')
  }
}

onMounted(() => { loadArea() })
</script>

<template>
  <div class="residential-area-detail">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <el-button :icon="ArrowLeft" @click="handleBack">返回</el-button>
          <span>小区详情</span>
          <el-button
            v-if="area && (area.submission_status === 0 || area.submission_status === 3)"
            type="primary"
            @click="handleEdit"
          >编辑</el-button>
        </div>
      </template>

      <div v-if="area" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="小区ID">{{ area.id }}</el-descriptions-item>
          <el-descriptions-item label="小区编码">{{ area.code || '-' }}</el-descriptions-item>
          <el-descriptions-item label="小区名称">{{ area.name }}</el-descriptions-item>
          <el-descriptions-item label="小区类型">{{ communityTypeMap[area.community_type] || '未知' }}</el-descriptions-item>
          <el-descriptions-item label="小区地址" :span="2">{{ area.address }}</el-descriptions-item>
          <el-descriptions-item label="提交状态">
            <el-tag :type="submissionStatusMap[area.submission_status]?.type || 'info'">
              {{ submissionStatusMap[area.submission_status]?.label || '未知' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ area.created_time }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ area.updated_time }}</el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">提交信息</el-divider>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="提交人ID">{{ area.submitter_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ area.submit_time || '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核人ID">{{ area.reviewer_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核时间">{{ area.review_time || '-' }}</el-descriptions-item>
          <el-descriptions-item label="审核备注" :span="2">
            <div v-if="area.review_notes" class="review-notes">{{ area.review_notes }}</div>
            <span v-else>-</span>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>
  </div>
</template>

<style scoped lang="scss">
.residential-area-detail {
  padding: 20px;
  .card-header {
    display: flex; justify-content: space-between; align-items: center; gap: 20px;
    span { flex: 1; text-align: center; font-size: 18px; font-weight: 600; }
  }
  .review-notes {
    padding: 10px; background-color: #f5f7fa; border-radius: 4px;
    white-space: pre-wrap; word-break: break-word;
  }
}
</style>
