<template>
  <div class="moderation-result">
    <el-alert
      :type="result.finalResult ? 'success' : 'error'"
      :title="result.finalResult ? '审核通过' : '审核未通过'"
      :closable="false"
      show-icon
    />

    <div class="result-meta">
      <span>请求ID: {{ result.requestId }}</span>
      <span>处理时间: {{ result.processingTime }}ms</span>
    </div>

    <el-collapse v-model="activeNames" class="result-layers">
      <!-- Traditional Technology Layer -->
      <el-collapse-item
        v-if="result.traditional"
        name="traditional"
        title="传统技术检测"
      >
        <div class="layer-content">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="检测结果">
              <el-tag :type="result.traditional.passed ? 'success' : 'danger'">
                {{ result.traditional.passed ? '通过' : '未通过' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item v-if="result.traditional.score" label="评分">
              {{ result.traditional.score }}
            </el-descriptions-item>
            <el-descriptions-item v-if="result.traditional.keywords" label="关键词">
              <el-tag
                v-for="keyword in result.traditional.keywords"
                :key="keyword"
                type="warning"
                size="small"
                style="margin-right: 8px"
              >
                {{ keyword }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item v-if="result.traditional.reason" label="原因">
              {{ result.traditional.reason }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-collapse-item>

      <!-- Small Model Layer -->
      <el-collapse-item
        v-if="result.smallModel"
        name="smallModel"
        title="小模型检测"
      >
        <div class="layer-content">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="检测结果">
              <el-tag :type="result.smallModel.passed ? 'success' : 'danger'">
                {{ result.smallModel.passed ? '通过' : '未通过' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="置信度">
              <el-progress
                :percentage="Math.round(result.smallModel.confidence * 100)"
                :color="getConfidenceColor(result.smallModel.confidence)"
              />
            </el-descriptions-item>
            <el-descriptions-item v-if="result.smallModel.categories" label="分类">
              <el-tag
                v-for="category in result.smallModel.categories"
                :key="category"
                type="info"
                size="small"
                style="margin-right: 8px"
              >
                {{ category }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item v-if="result.smallModel.reason" label="原因">
              {{ result.smallModel.reason }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-collapse-item>

      <!-- Large Model Layer -->
      <el-collapse-item
        v-if="result.largeModel"
        name="largeModel"
        title="大模型检测"
      >
        <div class="layer-content">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="检测结果">
              <el-tag :type="result.largeModel.passed ? 'success' : 'danger'">
                {{ result.largeModel.passed ? '通过' : '未通过' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="置信度">
              <el-progress
                :percentage="Math.round(result.largeModel.confidence * 100)"
                :color="getConfidenceColor(result.largeModel.confidence)"
              />
            </el-descriptions-item>
            <el-descriptions-item v-if="result.largeModel.categories" label="分类">
              <el-tag
                v-for="category in result.largeModel.categories"
                :key="category"
                type="info"
                size="small"
                style="margin-right: 8px"
              >
                {{ category }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item v-if="result.largeModel.analysis" label="分析">
              {{ result.largeModel.analysis }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </el-collapse-item>
    </el-collapse>

    <!-- Raw JSON View -->
    <el-collapse v-model="showRaw" class="raw-data">
      <el-collapse-item name="raw" title="原始数据 (JSON)">
        <pre>{{ JSON.stringify(result, null, 2) }}</pre>
      </el-collapse-item>
    </el-collapse>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import type { TextModerationResponse, ImageModerationResponse } from '@common/types/moderation';

interface Props {
  result: TextModerationResponse | ImageModerationResponse;
}

defineProps<Props>();

const activeNames = ref(['traditional', 'smallModel', 'largeModel']);
const showRaw = ref<string[]>([]);

const getConfidenceColor = (confidence: number) => {
  if (confidence >= 0.8) return '#67c23a';
  if (confidence >= 0.5) return '#e6a23c';
  return '#f56c6c';
};
</script>

<style scoped>
.moderation-result {
  margin-top: 20px;
}

.result-meta {
  display: flex;
  justify-content: space-between;
  margin: 16px 0;
  padding: 12px;
  background-color: #f5f7fa;
  border-radius: 4px;
  font-size: 14px;
  color: #606266;
}

.result-layers {
  margin-bottom: 16px;
}

.layer-content {
  padding: 12px;
}

.raw-data pre {
  background-color: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  overflow-x: auto;
  font-size: 12px;
  line-height: 1.5;
}
</style>
