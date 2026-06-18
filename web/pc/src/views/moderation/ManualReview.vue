<template>
  <div class="manual-review-page">
    <h3 style="margin-bottom: 16px;">人工审核</h3>

    <ReviewFilter
      v-model:source-type="sourceType"
      v-model:review-status="reviewStatus"
      @change="fetchList"
    />

    <div v-loading="loading">
      <ReviewTable
        :list="list"
        @detail="openDetail"
        @approve="onQuickApprove"
        @reject="onQuickReject"
      />
    </div>

    <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next"
        @change="fetchList"
      />
    </div>

    <ReviewDetailDrawer
      v-model="drawerVisible"
      :detail="currentDetail"
      @approve="onSubmitReview(1, $event)"
      @reject="onSubmitReview(2, $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import ReviewFilter from '@/components/moderation/ReviewFilter.vue';
import ReviewTable from '@/components/moderation/ReviewTable.vue';
import ReviewDetailDrawer from '@/components/moderation/ReviewDetailDrawer.vue';
import { listReview, getReviewDetail, submitReview } from '@/api/moderation';
import type { ReviewListItem, ReviewDetail } from '@common/types/moderation';

const sourceType = ref('');
const reviewStatus = ref(0);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const list = ref<ReviewListItem[]>([]);
const loading = ref(false);

const drawerVisible = ref(false);
const currentDetail = ref<ReviewDetail | null>(null);

async function fetchList() {
  loading.value = true;
  try {
    const res = await listReview({
      source_type: sourceType.value || undefined,
      review_status: reviewStatus.value,
      page: page.value,
      page_size: pageSize.value,
    });
    list.value = res.list;
    total.value = res.total;
  } catch (e) {
    console.error('fetchList failed:', e);
    ElMessage.error('加载审核列表失败');
  } finally {
    loading.value = false;
  }
}

async function openDetail(row: ReviewListItem) {
  try {
    const res = await getReviewDetail(row.id);
    // API wraps detail in { detail: {...} } after axios strips outer data layer
    currentDetail.value = (res as any).detail || res;
    drawerVisible.value = true;
  } catch (e) {
    console.error('getReviewDetail failed:', e);
    ElMessage.error('加载审核详情失败');
  }
}

async function onSubmitReview(status: number, notes: string) {
  if (!currentDetail.value) return;
  try {
    await submitReview({
      audit_log_id: currentDetail.value.id,
      review_status: status,
      review_notes: notes,
    });
    ElMessage.success(status === 1 ? '审核通过' : '已驳回');
    drawerVisible.value = false;
    fetchList();
  } catch (e) {
    console.error('submitReview failed:', e);
    ElMessage.error('提交审核失败');
  }
}

// Quick review from table row (without opening drawer)
async function onQuickApprove(row: ReviewListItem) {
  try {
    await ElMessageBox.confirm(
      `确认通过「${row.content_summary}」的审核？`,
      '快捷审核',
      { confirmButtonText: '确认通过', cancelButtonText: '取消', type: 'info' }
    );
    await submitReview({
      audit_log_id: row.id,
      review_status: 1,
      review_notes: '',
    });
    ElMessage.success('审核通过');
    fetchList();
  } catch (e: any) {
    if (e !== 'cancel') {
      console.error('quick approve failed:', e);
      ElMessage.error('操作失败');
    }
  }
}

async function onQuickReject(row: ReviewListItem) {
  try {
    await ElMessageBox.confirm(
      `确认驳回「${row.content_summary}」？`,
      '快捷审核',
      { confirmButtonText: '确认驳回', cancelButtonText: '取消', type: 'warning' }
    );
    await submitReview({
      audit_log_id: row.id,
      review_status: 2,
      review_notes: '',
    });
    ElMessage.success('已驳回');
    fetchList();
  } catch (e: any) {
    if (e !== 'cancel') {
      console.error('quick reject failed:', e);
      ElMessage.error('操作失败');
    }
  }
}

fetchList();
</script>
