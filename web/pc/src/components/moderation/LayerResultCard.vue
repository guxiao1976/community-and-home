<template>
  <div class="layer-card" :class="{ skipped: !result || !result.called }">
    <div class="layer-title">{{ title }}</div>
    <div v-if="result && result.called" class="layer-body">
      <div class="layer-row">
        <el-tag :type="result.passed ? 'success' : 'danger'" size="small">{{ result.passed ? '通过' : '未通过' }}</el-tag>
        <el-tag style="margin-left: 4px;" size="small">{{ result.risk_level || 'unknown' }}</el-tag>
      </div>
      <div class="layer-row" v-if="result.confidence !== undefined">置信度: {{ (result.confidence * 100).toFixed(0) }}%</div>
      <div class="layer-row" v-if="result.matched_words?.length">命中: {{ result.matched_words.join(', ') }}</div>
      <div class="layer-row" v-if="result.categories?.length">分类: {{ result.categories.join(', ') }}</div>
      <div class="layer-row" v-if="result.model_used">模型: {{ result.model_used }}</div>
      <div class="layer-row text-muted">耗时: {{ result.latency_ms }}ms</div>
      <el-button v-if="result.raw_response" size="small" text type="primary" @click="copyJson(result.raw_response)">复制原始JSON</el-button>
    </div>
    <div v-else class="layer-skipped">{{ result?.skipped_reason || '未调用' }}</div>
  </div>
</template>

<script setup lang="ts">
import { ElMessage } from 'element-plus';
import type { PipelineLayerResult } from '@common/types/moderation';

defineProps<{
  title: string;
  result: PipelineLayerResult | null;
  showMatchedWords?: boolean;
}>();

const copyJson = (data: string) => {
  navigator.clipboard.writeText(data).then(() => ElMessage.success('已复制到剪贴板'), () => ElMessage.error('复制失败'));
};
</script>

<style scoped>
.layer-card { border: 1px solid #ebeef5; border-radius: 4px; padding: 12px; min-height: 160px; }
.layer-card.skipped { background-color: #f5f7fa; }
.layer-title { font-weight: 600; margin-bottom: 8px; font-size: 14px; }
.layer-row { margin-bottom: 4px; font-size: 13px; }
.text-muted { color: #909399; }
.layer-skipped { color: #c0c4cc; font-size: 13px; margin-top: 24px; text-align: center; }
</style>
