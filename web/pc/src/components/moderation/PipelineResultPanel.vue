<template>
  <div v-if="result" class="pipeline-result">
    <el-card shadow="never">
      <template #header>
        <div class="result-header">
          <span>执行结果</span>
          <el-tag :type="verdictTagType">{{ verdictLabel }}</el-tag>
        </div>
      </template>

      <el-row :gutter="16">
        <el-col :span="8"><LayerResultCard title="AC 引擎" :result="result.ac_result" /></el-col>
        <el-col :span="8"><LayerResultCard title="小模型" :result="result.small_model_result" /></el-col>
        <el-col :span="8"><LayerResultCard title="大模型" :result="result.large_model_result" /></el-col>
      </el-row>

      <el-table :data="overviewRows" border style="margin-top: 16px;">
        <el-table-column prop="layer" label="审核层" width="100">
          <template #default="{ row }"><el-tag size="small">{{ row.label }}</el-tag></template>
        </el-table-column>
        <el-table-column prop="called" label="调用情况" width="100">
          <template #default="{ row }">
            <el-tag :type="row.called ? 'success' : 'info'" size="small">{{ row.called ? '已调用' : '未调用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="passed" label="结果" width="100">
          <template #default="{ row }">
            <template v-if="row.called"><el-tag :type="row.passed ? 'success' : 'danger'" size="small">{{ row.passed ? '通过' : '未通过' }}</el-tag></template>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="confidence" label="置信度" width="100">
          <template #default="{ row }">
            <span v-if="row.called && row.confidence !== undefined">{{ (row.confidence * 100).toFixed(0) }}%</span>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="latency" label="耗时" width="100">
          <template #default="{ row }">
            <span v-if="row.called">{{ row.latency }}ms</span>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
        <el-table-column prop="detail" label="详情" min-width="200">
          <template #default="{ row }">
            <span v-if="row.detail" style="color: #909399; font-size: 13px;">{{ row.detail }}</span>
            <span v-else style="color: #c0c4cc;">—</span>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import LayerResultCard from './LayerResultCard.vue';
import type { PipelineTestResponse, PipelineLayerResult } from '@common/types/moderation';

const props = defineProps<{ result: PipelineTestResponse | null }>();

const verdictLabel = computed(() => {
  const m: Record<string, string> = { pass: '通过', reject: '拒绝', need_review: '需人工审核' };
  return m[props.result?.final_verdict || ''] || props.result?.final_verdict || '';
});
const verdictTagType = computed(() => {
  const m: Record<string, string> = { pass: 'success', reject: 'danger', need_review: 'warning' };
  return m[props.result?.final_verdict || ''] || 'info';
});

const buildRow = (key: string, label: string, lr: PipelineLayerResult | null) => {
  if (!lr) return { layer: key, label, called: false, passed: false, detail: '' };
  return {
    layer: key, label,
    called: lr.called, passed: lr.passed,
    confidence: lr.confidence, latency: lr.latency_ms,
    detail: lr.called ? (lr.reason || lr.matched_words?.join(', ') || (lr.passed ? '' : '未通过')) : (lr.skipped_reason || ''),
  };
};
const overviewRows = computed(() => {
  if (!props.result) return [];
  return [buildRow('ac', 'AC引擎', props.result.ac_result), buildRow('small', '小模型', props.result.small_model_result), buildRow('large', '大模型', props.result.large_model_result)];
});
</script>

<style scoped>
.pipeline-result { margin-top: 16px; }
.result-header { display: flex; justify-content: space-between; align-items: center; }
</style>
