<template>
  <el-drawer v-model="visible" title="审核详情" size="500px" @close="$emit('close')">
    <template v-if="detail">
      <el-descriptions :column="1" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="板块">{{ SOURCE_TYPE_LABELS[detail.source_type] || detail.source_type }}</el-descriptions-item>
        <el-descriptions-item label="内容摘要">{{ detail.content_summary }}</el-descriptions-item>
        <el-descriptions-item label="风险等级">
          <el-tag :type="detail.risk_level === 'high' ? 'danger' : detail.risk_level === 'medium' ? 'warning' : 'info'" size="small">{{ detail.risk_level }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="机器审核">{{ detail.pass ? '通过' : '不通过' }}</el-descriptions-item>
        <el-descriptions-item label="审核理由">{{ detail.reason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="审核层">{{ detail.check_layer || '-' }}</el-descriptions-item>
        <el-descriptions-item label="命中详情">
          <pre v-if="matchedItems.length" style="max-height: 200px; overflow: auto; font-size: 12px;">{{ JSON.stringify(matchedItems, null, 2) }}</pre>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="提交时间">{{ detail.created_time }}</el-descriptions-item>
      </el-descriptions>

      <div v-if="detail.review_status === 0" style="margin-top: 20px;">
        <el-input
          v-model="reviewNotes"
          type="textarea"
          :rows="3"
          placeholder="审核备注（可选）"
        />
        <div style="margin-top: 12px; display: flex; gap: 12px; justify-content: flex-end;">
          <el-button type="danger" @click="onReject" :loading="loading">不通过</el-button>
          <el-button type="success" @click="onApprove" :loading="loading">通过</el-button>
        </div>
      </div>
      <div v-else style="margin-top: 20px; color: #909399;">
        该记录已审核完成
        <div v-if="detail.review_notes">审核备注：{{ detail.review_notes }}</div>
      </div>
    </template>
  </el-drawer>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';
import type { ReviewDetail } from '@common/types/moderation';
import { SOURCE_TYPE_LABELS } from '@common/types/moderation';

const props = defineProps<{
  modelValue: boolean;
  detail: ReviewDetail | null;
}>();
const emit = defineEmits<{
  (e: 'update:modelValue', v: boolean): void;
  (e: 'close'): void;
  (e: 'approve', notes: string): void;
  (e: 'reject', notes: string): void;
}>();

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v),
});

const reviewNotes = ref('');
const loading = ref(false);

const matchedItems = computed(() => {
  if (!props.detail?.matched_items) return [];
  try { return JSON.parse(props.detail.matched_items); } catch { return []; }
});

function onApprove() { loading.value = true; emit('approve', reviewNotes.value); }
function onReject() { loading.value = true; emit('reject', reviewNotes.value); }
</script>
