<template>
  <div class="verification-list">
    <div class="page-header">
      <h2>业主认证审核</h2>
    </div>

    <el-card>
      <div class="filter-bar">
        <el-form :inline="true" :model="filters">
          <el-form-item label="审核状态">
            <el-select v-model="filters.status" placeholder="全部" clearable style="width: 150px">
              <el-option label="待审核" :value="0" />
              <el-option label="已通过" :value="1" />
              <el-option label="已拒绝" :value="2" />
            </el-select>
          </el-form-item>
          <el-form-item label="提交时间">
            <el-date-picker
              v-model="dateRange"
              type="daterange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              value-format="YYYY-MM-DD"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <el-table :data="tableData" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column label="申请人" width="150">
          <template #default="{ row }">
            <div>{{ row.realName }}</div>
            <div style="color: #909399; font-size: 12px">{{ desensitizePhone(row.user?.phone || '') }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="propertyUnit" label="房产单元" show-overflow-tooltip />
        <el-table-column label="身份证号" width="180">
          <template #default="{ row }">
            {{ desensitizeIdCard(row.idCard) }}
          </template>
        </el-table-column>
        <el-table-column label="证件照片" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.documentUrls?.length || 0 }} 张</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="审核状态" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.verificationStatus === 0" type="warning" size="small">待审核</el-tag>
            <el-tag v-else-if="row.verificationStatus === 1" type="success" size="small">已通过</el-tag>
            <el-tag v-else-if="row.verificationStatus === 2" type="danger" size="small">已拒绝</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="提交时间" width="180" />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleView(row)">查看详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadData"
          @current-change="loadData"
        />
      </div>
    </el-card>

    <!-- Detail Dialog -->
    <el-dialog
      v-model="detailVisible"
      title="认证详情"
      width="800px"
      @close="resetDetail"
    >
      <div v-if="currentVerification" class="verification-detail">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="申请人">{{ currentVerification.realName }}</el-descriptions-item>
          <el-descriptions-item label="手机号">{{ desensitizePhone(currentVerification.user?.phone || '') }}</el-descriptions-item>
          <el-descriptions-item label="身份证号">{{ desensitizeIdCard(currentVerification.idCard) }}</el-descriptions-item>
          <el-descriptions-item label="房产单元">{{ currentVerification.propertyUnit }}</el-descriptions-item>
          <el-descriptions-item label="提交时间" :span="2">{{ currentVerification.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="审核状态" :span="2">
            <el-tag v-if="currentVerification.verificationStatus === 0" type="warning">待审核</el-tag>
            <el-tag v-else-if="currentVerification.verificationStatus === 1" type="success">已通过</el-tag>
            <el-tag v-else-if="currentVerification.verificationStatus === 2" type="danger">已拒绝</el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="currentVerification.reviewerId" label="审核人" :span="2">
            {{ currentVerification.reviewer?.nickname }} ({{ currentVerification.reviewedAt }})
          </el-descriptions-item>
          <el-descriptions-item v-if="currentVerification.reviewNotes" label="审核备注" :span="2">
            {{ currentVerification.reviewNotes }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="document-section">
          <h4>证件照片</h4>
          <div class="document-images">
            <el-image
              v-for="(url, index) in currentVerification.documentUrls"
              :key="index"
              :src="url"
              :preview-src-list="currentVerification.documentUrls"
              :initial-index="index"
              fit="cover"
              class="document-image"
            />
          </div>
        </div>

        <div v-if="currentVerification.verificationStatus === 0" class="review-actions">
          <el-divider />
          <el-form :model="reviewForm" label-width="100px">
            <el-form-item label="审核备注">
              <el-input
                v-model="reviewForm.notes"
                type="textarea"
                :rows="3"
                placeholder="请输入审核备注（拒绝时必填）"
              />
            </el-form-item>
            <el-form-item>
              <el-button v-permission="'verification:approve'" type="success" :loading="submitting" @click="handleApprove">
                通过
              </el-button>
              <el-button v-permission="'verification:reject'" type="danger" :loading="submitting" @click="handleReject">
                拒绝
              </el-button>
            </el-form-item>
          </el-form>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import type { HomeownerVerification } from '@common/types/identity';
import * as identityApi from '@/api/identity';
import { desensitizePhone, desensitizeIdCard } from '@common/utils/desensitize';

const loading = ref(false);
const submitting = ref(false);
const detailVisible = ref(false);
const tableData = ref<HomeownerVerification[]>([]);
const currentVerification = ref<HomeownerVerification | null>(null);
const dateRange = ref<[string, string] | null>(null);

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
});

const filters = reactive({
  status: undefined as number | undefined
});

const reviewForm = reactive({
  notes: ''
});

onMounted(() => {
  loadData();
});

const loadData = async () => {
  loading.value = true;
  try {
    const params: any = {
      page: pagination.page,
      page_size: pagination.pageSize
    };

    if (filters.status !== undefined) {
      params.status = filters.status;
    }

    if (dateRange.value) {
      params.start_date = dateRange.value[0];
      params.end_date = dateRange.value[1];
    }

    const response = await identityApi.getVerifications(params);
    tableData.value = response?.list || [];
    pagination.total = response?.total || 0;
  } catch (error) {
    ElMessage.error('加载数据失败');
  } finally {
    loading.value = false;
  }
};

const handleSearch = () => {
  pagination.page = 1;
  loadData();
};

const handleReset = () => {
  filters.status = undefined;
  dateRange.value = null;
  pagination.page = 1;
  loadData();
};

const handleView = async (row: HomeownerVerification) => {
  try {
    const response = await identityApi.getVerificationById(row.id);
    currentVerification.value = response;
    detailVisible.value = true;
  } catch (error) {
    ElMessage.error('加载详情失败');
  }
};

const handleApprove = async () => {
  if (!currentVerification.value) return;

  try {
    await ElMessageBox.confirm('确定通过该认证申请吗？', '提示', {
      type: 'success'
    });

    submitting.value = true;
    await identityApi.reviewVerification(currentVerification.value.id, {
      status: 1,
      review_notes: reviewForm.notes
    });

    ElMessage.success('审核通过');
    detailVisible.value = false;
    loadData();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败');
    }
  } finally {
    submitting.value = false;
  }
};

const handleReject = async () => {
  if (!currentVerification.value) return;

  if (!reviewForm.notes.trim()) {
    ElMessage.warning('拒绝时必须填写审核备注');
    return;
  }

  try {
    await ElMessageBox.confirm('确定拒绝该认证申请吗？', '提示', {
      type: 'warning'
    });

    submitting.value = true;
    await identityApi.reviewVerification(currentVerification.value.id, {
      status: 2,
      review_notes: reviewForm.notes
    });

    ElMessage.success('已拒绝');
    detailVisible.value = false;
    loadData();
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败');
    }
  } finally {
    submitting.value = false;
  }
};

const resetDetail = () => {
  currentVerification.value = null;
  reviewForm.notes = '';
};
</script>

<style scoped lang="scss">
.verification-list {
  .page-header {
    margin-bottom: 20px;

    h2 {
      margin: 0;
      font-size: 20px;
      font-weight: 500;
    }
  }

  .filter-bar {
    margin-bottom: 20px;
  }

  .pagination {
    margin-top: 20px;
    display: flex;
    justify-content: flex-end;
  }
}

.verification-detail {
  .document-section {
    margin-top: 24px;

    h4 {
      margin: 0 0 16px 0;
      font-size: 16px;
      font-weight: 500;
    }

    .document-images {
      display: grid;
      grid-template-columns: repeat(3, 1fr);
      gap: 12px;

      .document-image {
        width: 100%;
        height: 200px;
        border-radius: 4px;
        cursor: pointer;
      }
    }
  }

  .review-actions {
    margin-top: 24px;
  }
}
</style>
