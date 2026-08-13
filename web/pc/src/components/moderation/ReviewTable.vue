<template>
  <el-table :data="list" border stripe style="width: 100%">
    <el-table-column prop="id" label="ID" width="80" />
    <el-table-column label="板块" width="100">
      <template #default="{ row }">
        {{ sourceTypeLabel(row.source_type) }}
      </template>
    </el-table-column>
    <el-table-column prop="content_summary" label="内容摘要" show-overflow-tooltip />
    <el-table-column label="风险等级" width="90">
      <template #default="{ row }">
        <el-tag :type="riskTagType(row.risk_level)" size="small">
          {{ riskLabel(row.risk_level) }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column label="审核状态" width="100">
      <template #default="{ row }">
        <el-tag :type="reviewStatusTagType(row.review_status)" size="small">
          {{ reviewStatusLabel(row.review_status) }}
        </el-tag>
      </template>
    </el-table-column>
    <el-table-column prop="created_time" label="提交时间" width="170" />
    <el-table-column label="操作" width="240" fixed="right">
      <template #default="{ row }">
        <div style="display: flex; gap: 4px; flex-wrap: nowrap;">
          <el-button type="primary" size="small" @click="$emit('detail', row as ReviewListItem)">详情</el-button>
          <template v-if="row.review_status === 0">
            <el-button type="success" size="small" @click="$emit('approve', row as ReviewListItem)">通过</el-button>
            <el-button type="danger" size="small" @click="$emit('reject', row as ReviewListItem)">不通过</el-button>
          </template>
        </div>
      </template>
    </el-table-column>
  </el-table>
</template>

<script setup lang="ts">
/* eslint-disable */
import type { ReviewListItem } from '@common/types/moderation';
import { SOURCE_TYPE_LABELS, MODERATION_STATUS_MAP } from '@common/constants/moderation';

defineProps<{ list: ReviewListItem[] }>();
defineEmits<{
  (e: 'detail', row: ReviewListItem): void;
  (e: 'approve', row: ReviewListItem): void;
  (e: 'reject', row: ReviewListItem): void;
}>();

function sourceTypeLabel(t: string) { return SOURCE_TYPE_LABELS[t] || t; }
function riskLabel(r: string) {
  const m: Record<string, string> = { high: '高', medium: '中', low: '低' };
  return m[r] || r;
}
type TagType = 'primary' | 'success' | 'warning' | 'info' | 'danger';

function riskTagType(r: string): TagType {
  const m: Record<string, TagType> = { high: 'danger', medium: 'warning', low: 'info' };
  return m[r] || 'info';
}
function reviewStatusTagType(s: number): TagType {
  return (MODERATION_STATUS_MAP[s]?.type || 'info') as TagType;
}
function reviewStatusLabel(s: number) {
  return MODERATION_STATUS_MAP[s]?.label || '未知';
}
</script>
